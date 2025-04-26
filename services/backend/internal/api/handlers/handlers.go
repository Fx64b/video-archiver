package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/download"
	"video-archiver/internal/util/statistics"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type Response struct {
	Message interface{} `json:"message"`
}

type Handler struct {
	downloadService *download.Service
	downloadPath    string
}

func NewHandler(downloadService *download.Service, downloadPath string) *Handler {
	return &Handler{
		downloadService: downloadService,
		downloadPath:    downloadPath,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Use(corsMiddleware)

	r.Post("/download", h.HandleDownload)
	r.Get("/recent", h.HandleRecent)
	r.Get("/job/:id", h.HandleGetJob)
	r.Get("/statistics", h.HandleGetStatistics)
	r.Get("/downloads/{type}", h.HandleGetDownloads)
}

func (h *Handler) RegisterWSRoutes(r *chi.Mux) {
	r.Use(corsMiddleware)
	r.Get("/ws", h.HandleWebSocket)
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Debug("WebSocket connection attempt received")

	upgrader := download.GetUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}
	log.Debug("WebSocket connection successfully upgraded")

	hub := h.downloadService.GetHub()
	hub.Register(conn)
	log.Debug("WebSocket connection registered with hub")

	defer func() {
		log.Debug("Cleaning up WebSocket connection")
		hub.Unregister(conn)
		conn.Close()
	}()

	conn.SetReadDeadline(time.Time{}) // No timeout TODO: currently works, but implement timeout or ping/ping later

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.WithError(err).Error("WebSocket unexpected close error")
			} else {
				log.WithError(err).Debug("WebSocket connection closed")
			}
			return
		}
	}
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Infof("Received download request for URL: %s", req.URL)

	job := domain.Job{
		ID:        uuid.New().String(),
		URL:       req.URL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.downloadService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: "Video added to download queue"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleRecent(w http.ResponseWriter, r *http.Request) {
	recentJobs, err := h.downloadService.GetRecentWithMetadata(5)
	if err != nil {
		log.WithError(err).Error("Failed to get recent jobs with metadata")
		http.Error(w, "Failed to get recent jobs with metadata", http.StatusInternalServerError)
		return
	}

	if len(recentJobs) == 0 {
		http.Error(w, "No recent downloads found", http.StatusNotFound)
		return
	}

	resp := Response{Message: recentJobs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	jobWithMetadata, err := h.downloadService.GetJobWithMetadata(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get job with metadata")
		http.Error(w, "Failed to get job with metadata", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: jobWithMetadata}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := statistics.GetStatistics(h.downloadService.GetRepository(), h.downloadPath)
	if err != nil {
		log.WithError(err).Error("Failed to get statistics")
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: stats}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGetDownloads(w http.ResponseWriter, r *http.Request) {
	contentType := chi.URLParam(r, "type")

	// Validate content type
	if contentType != "videos" && contentType != "playlists" && contentType != "channels" {
		http.Error(w, "Invalid content type. Must be 'videos', 'playlists', or 'channels'", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	jobRepo := h.downloadService.GetRepository()
	items, totalCount, err := jobRepo.GetMetadataByType(contentType, page, limit, sortBy, order)
	if err != nil {
		log.WithError(err).Errorf("Failed to get %s", contentType)
		http.Error(w, fmt.Sprintf("Failed to get %s", contentType), http.StatusInternalServerError)
		return
	}

	if len(items) == 0 && page == 1 {
		// Only return 404 for the first page with no results
		log.WithError(err).Errorf("No %s found", contentType)
		http.Error(w, fmt.Sprintf("No %s found", contentType), http.StatusNotFound)
		return
	}

	resp := struct {
		Items      []*domain.JobWithMetadata `json:"items"`
		TotalCount int                       `json:"total_count"`
		Page       int                       `json:"page"`
		Limit      int                       `json:"limit"`
		TotalPages int                       `json:"total_pages"`
	}{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: (totalCount + limit - 1) / limit, // Ceiling division
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Response{Message: resp})
	if err != nil {
		return
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
