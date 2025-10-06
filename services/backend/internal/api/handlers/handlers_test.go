package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"video-archiver/internal/services/download"
	"video-archiver/internal/testutil"

	"github.com/go-chi/chi"
)

func setupTestHandler(t *testing.T) (*Handler, *testutil.MockJobRepository) {
	t.Helper()

	mockRepo := testutil.NewMockJobRepository()
	config := &download.Config{
		JobRepository: mockRepo,
		DownloadPath:  "/tmp/test",
		Concurrency:   2,
		MaxQuality:    1080,
	}
	service := download.NewService(config)

	handler := NewHandler(service, "/tmp/test")
	return handler, mockRepo
}

func TestHandleDownload(t *testing.T) {
	handler, _ := setupTestHandler(t)

	tests := []struct {
		name           string
		requestBody    DownloadRequest
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "valid download request",
			requestBody: DownloadRequest{
				URL: "https://youtube.com/watch?v=test",
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name: "valid playlist request",
			requestBody: DownloadRequest{
				URL: "https://youtube.com/playlist?list=test",
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.HandleDownload(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.checkResponse {
				var resp Response
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Message == nil {
					t.Error("Expected non-nil message in response")
				}
			}
		})
	}
}

func TestHandleDownload_InvalidRequest(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.HandleDownload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetJob(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	// Create test job
	job := testutil.CreateTestJob("test-job-id", "https://youtube.com/watch?v=test")
	mockRepo.Create(job)
	mockRepo.StoreMetadata(job.ID, testutil.CreateTestVideoMetadata())

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
	}{
		{
			name:           "existing job",
			jobID:          "test-job-id",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing job",
			jobID:          "non-existent",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/job/"+tt.jobID, nil)
			w := httptest.NewRecorder()

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.jobID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			handler.HandleGetJob(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp Response
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
			}
		})
	}
}

func TestHandleGetJobParents(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	// Create parent and video jobs
	videoJob := testutil.CreateTestJob("video-id", "https://youtube.com/watch?v=test")
	mockRepo.Create(videoJob)
	mockRepo.StoreMetadata(videoJob.ID, testutil.CreateTestVideoMetadata())

	playlistJob := testutil.CreateTestJob("playlist-id", "https://youtube.com/playlist?list=test")
	mockRepo.Create(playlistJob)
	mockRepo.StoreMetadata(playlistJob.ID, testutil.CreateTestPlaylistMetadata())

	mockRepo.AddVideoToParent("video-id", "playlist-id", "playlist")

	req := httptest.NewRequest(http.MethodGet, "/job/video-id/parents", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "video-id")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	handler.HandleGetJobParents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestHandleGetStatistics(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	// Create test data
	videoJob := testutil.CreateTestJob("video-1", "https://youtube.com/watch?v=test1")
	mockRepo.Create(videoJob)
	mockRepo.StoreMetadata(videoJob.ID, testutil.CreateTestVideoMetadata())

	playlistJob := testutil.CreateTestJob("playlist-1", "https://youtube.com/playlist?list=test")
	mockRepo.Create(playlistJob)
	mockRepo.StoreMetadata(playlistJob.ID, testutil.CreateTestPlaylistMetadata())

	req := httptest.NewRequest(http.MethodGet, "/statistics", nil)
	w := httptest.NewRecorder()

	handler.HandleGetStatistics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestHandleGetDownloads(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	// Create test data
	for i := 1; i <= 5; i++ {
		job := testutil.CreateTestJob("video-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
		mockRepo.Create(job)
		mockRepo.StoreMetadata(job.ID, testutil.CreateTestVideoMetadata())
		time.Sleep(1 * time.Millisecond)
	}

	tests := []struct {
		name           string
		contentType    string
		queryParams    string
		expectedStatus int
		checkCount     bool
		minCount       int
	}{
		{
			name:           "get videos default",
			contentType:    "videos",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkCount:     true,
			minCount:       5,
		},
		{
			name:           "get videos page 1",
			contentType:    "videos",
			queryParams:    "?page=1&limit=3",
			expectedStatus: http.StatusOK,
			checkCount:     true,
			minCount:       3,
		},
		{
			name:           "get playlists",
			contentType:    "playlists",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkCount:     false,
		},
		{
			name:           "invalid content type",
			contentType:    "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			checkCount:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/downloads/"+tt.contentType+tt.queryParams, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.contentType)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			handler.HandleGetDownloads(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK && tt.checkCount {
				var resp Response
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
			}
		})
	}
}

func TestHandleRecent(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	tests := []struct {
		name           string
		setupJobs      int
		expectedStatus int
	}{
		{
			name:           "with jobs",
			setupJobs:      3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no jobs",
			setupJobs:      0,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean mock repo for each test
			mockRepo = testutil.NewMockJobRepository()
			config := &download.Config{
				JobRepository: mockRepo,
				DownloadPath:  "/tmp/test",
				Concurrency:   2,
				MaxQuality:    1080,
			}
			service := download.NewService(config)
			handler = NewHandler(service, "/tmp/test")

			// Setup test data
			for i := 0; i < tt.setupJobs; i++ {
				job := testutil.CreateTestJob("job-"+string(rune('0'+i)), "https://youtube.com/watch?v=test"+string(rune('0'+i)))
				mockRepo.Create(job)
				mockRepo.StoreMetadata(job.ID, testutil.CreateTestVideoMetadata())
			}

			req := httptest.NewRequest(http.MethodGet, "/recent", nil)
			w := httptest.NewRecorder()

			handler.HandleRecent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestCorsMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := corsMiddleware(nextHandler)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "OPTIONS request",
			method:         http.MethodOptions,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			handlerToTest.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.checkHeaders {
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Error("CORS header Access-Control-Allow-Origin not set correctly")
				}
				if w.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("CORS header Access-Control-Allow-Methods not set")
				}
			}
		})
	}
}

func TestResponse_Structure(t *testing.T) {
	// Test that Response can be marshaled/unmarshaled correctly
	resp := Response{
		Message: map[string]string{"status": "success"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal Response: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Response: %v", err)
	}
}

func TestDownloadRequest_Structure(t *testing.T) {
	// Test that DownloadRequest can be marshaled/unmarshaled correctly
	req := DownloadRequest{
		URL: "https://youtube.com/watch?v=test",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal DownloadRequest: %v", err)
	}

	var decoded DownloadRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DownloadRequest: %v", err)
	}

	if decoded.URL != req.URL {
		t.Errorf("URL = %v, want %v", decoded.URL, req.URL)
	}
}
