package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/go-chi/chi"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/tools"
	"video-archiver/internal/testutil"
)

// inMemoryToolsRepo is a tiny ToolsRepository for handler tests.
type inMemoryToolsRepo struct {
	mu   sync.Mutex
	jobs map[string]*domain.ToolsJob
}

func newInMemoryToolsRepo() *inMemoryToolsRepo {
	return &inMemoryToolsRepo{jobs: make(map[string]*domain.ToolsJob)}
}

func (r *inMemoryToolsRepo) Create(job *domain.ToolsJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *job
	r.jobs[job.ID] = &cp
	return nil
}
func (r *inMemoryToolsRepo) Update(job *domain.ToolsJob) error { return r.Create(job) }
func (r *inMemoryToolsRepo) GetByID(id string) (*domain.ToolsJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if job, ok := r.jobs[id]; ok {
		cp := *job
		return &cp, nil
	}
	return nil, nil
}
func (r *inMemoryToolsRepo) GetAll() ([]*domain.ToolsJob, error) { return nil, nil }
func (r *inMemoryToolsRepo) GetByStatus(domain.ToolsJobStatus) ([]*domain.ToolsJob, error) {
	return nil, nil
}
func (r *inMemoryToolsRepo) Delete(string) error { return nil }
func (r *inMemoryToolsRepo) List(page, limit int, status, operationType string) ([]*domain.ToolsJob, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.ToolsJob
	for _, j := range r.jobs {
		if status != "" && string(j.Status) != status {
			continue
		}
		if operationType != "" && string(j.OperationType) != operationType {
			continue
		}
		cp := *j
		out = append(out, &cp)
	}
	return out, len(out), nil
}

func newToolsTestServer(t *testing.T) (*chi.Mux, *tools.Service) {
	t.Helper()
	svc := tools.NewService(&tools.Config{
		ToolsRepository: newInMemoryToolsRepo(),
		JobRepository:   testutil.NewMockJobRepository(),
		DownloadPath:    t.TempDir(),
		ProcessedPath:   t.TempDir(),
		Concurrency:     1,
	})
	h := NewToolsHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r, svc
}

func TestHandleListOperations(t *testing.T) {
	r, _ := newToolsTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/tools/operations", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		Message []map[string]any `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Message) != 7 {
		t.Errorf("expected 7 operations, got %d", len(resp.Message))
	}
}

func TestHandleSubmitValid(t *testing.T) {
	r, _ := newToolsTestServer(t)

	body := `{"input_files":["v1"],"parameters":{"start_time":"0","end_time":"10"}}`
	req := httptest.NewRequest(http.MethodPost, "/tools/trim", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (body: %s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Message domain.ToolsJob `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Message.ID == "" {
		t.Error("expected job ID in response")
	}
	if resp.Message.OperationType != domain.OpTypeTrim {
		t.Errorf("operation = %s, want trim", resp.Message.OperationType)
	}
}

func TestHandleSubmitUnknownOperation(t *testing.T) {
	r, _ := newToolsTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/tools/bogus", bytes.NewBufferString(`{"input_files":["v1"]}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandleSubmitInvalidParams(t *testing.T) {
	r, _ := newToolsTestServer(t)

	// end before start
	body := `{"input_files":["v1"],"parameters":{"start_time":"10","end_time":"5"}}`
	req := httptest.NewRequest(http.MethodPost, "/tools/trim", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestHandleSubmitConcatRequiresTwo(t *testing.T) {
	r, _ := newToolsTestServer(t)

	body := `{"input_files":["v1"],"parameters":{"output_format":"mp4"}}`
	req := httptest.NewRequest(http.MethodPost, "/tools/concat", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleGetAndCancelJob(t *testing.T) {
	r, svc := newToolsTestServer(t)

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeRotate,
		InputFiles:    []string{"v1"},
		Parameters:    map[string]any{"rotation": float64(90)},
	}
	if err := svc.Submit(job); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// GET the job
	req := httptest.NewRequest(http.MethodGet, "/tools/jobs/"+job.ID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", rec.Code)
	}

	// Cancel the job
	req = httptest.NewRequest(http.MethodDelete, "/tools/jobs/"+job.ID, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, want 200", rec.Code)
	}

	// Unknown job 404
	req = httptest.NewRequest(http.MethodGet, "/tools/jobs/missing", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing job status = %d, want 404", rec.Code)
	}
}

func TestHandleServeOutput(t *testing.T) {
	toolsRepo := newInMemoryToolsRepo()
	processedDir := t.TempDir()
	svc := tools.NewService(&tools.Config{
		ToolsRepository: toolsRepo,
		JobRepository:   testutil.NewMockJobRepository(),
		DownloadPath:    t.TempDir(),
		ProcessedPath:   processedDir,
		Concurrency:     1,
	})
	h := NewToolsHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// A completed job with a real output file is downloadable.
	outputPath := filepath.Join(processedDir, "concat_out.mp4")
	if err := os.WriteFile(outputPath, []byte("video-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	done := &domain.ToolsJob{
		ID:            "done1",
		OperationType: domain.OpTypeConcat,
		Status:        domain.ToolsJobStatusComplete,
		OutputFile:    outputPath,
	}
	_ = toolsRepo.Create(done)

	req := httptest.NewRequest(http.MethodGet, "/tools/jobs/done1/output", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "video-bytes" {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}
	if cd := rec.Header().Get("Content-Disposition"); cd == "" {
		t.Error("expected Content-Disposition header")
	}

	// A pending job has no output yet.
	pending := &domain.ToolsJob{ID: "pending1", Status: domain.ToolsJobStatusProcessing}
	_ = toolsRepo.Create(pending)
	req = httptest.NewRequest(http.MethodGet, "/tools/jobs/pending1/output", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("pending output status = %d, want 409", rec.Code)
	}

	// Completed job whose file is missing on disk.
	missing := &domain.ToolsJob{ID: "missing1", Status: domain.ToolsJobStatusComplete, OutputFile: filepath.Join(processedDir, "gone.mp4")}
	_ = toolsRepo.Create(missing)
	req = httptest.NewRequest(http.MethodGet, "/tools/jobs/missing1/output", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing file status = %d, want 404", rec.Code)
	}

	// Unknown job.
	req = httptest.NewRequest(http.MethodGet, "/tools/jobs/nope/output", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown job status = %d, want 404", rec.Code)
	}
}

func TestHandleListJobs(t *testing.T) {
	r, svc := newToolsTestServer(t)

	for i := 0; i < 3; i++ {
		job := &domain.ToolsJob{
			OperationType: domain.OpTypeRotate,
			InputFiles:    []string{"v1"},
			Parameters:    map[string]any{"rotation": float64(90)},
		}
		if err := svc.Submit(job); err != nil {
			t.Fatalf("submit: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/tools/jobs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		Message struct {
			Items      []*domain.ToolsJob `json:"items"`
			TotalCount int                `json:"total_count"`
		} `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Message.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", resp.Message.TotalCount)
	}
}
