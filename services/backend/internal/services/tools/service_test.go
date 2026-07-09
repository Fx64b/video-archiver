package tools

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

// memToolsRepo is a minimal in-memory ToolsRepository for tests.
type memToolsRepo struct {
	mu   sync.Mutex
	jobs map[string]*domain.ToolsJob
}

func newMemToolsRepo() *memToolsRepo {
	return &memToolsRepo{jobs: make(map[string]*domain.ToolsJob)}
}

func (r *memToolsRepo) Create(job *domain.ToolsJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *job
	r.jobs[job.ID] = &cp
	return nil
}

func (r *memToolsRepo) Update(job *domain.ToolsJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *job
	r.jobs[job.ID] = &cp
	return nil
}

func (r *memToolsRepo) GetByID(id string) (*domain.ToolsJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil, nil
	}
	cp := *job
	return &cp, nil
}

func (r *memToolsRepo) GetAll() ([]*domain.ToolsJob, error) { return nil, nil }
func (r *memToolsRepo) GetByStatus(domain.ToolsJobStatus) ([]*domain.ToolsJob, error) {
	return nil, nil
}
func (r *memToolsRepo) FindLatestConvertForInput(jobID string) (*domain.ToolsJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var latest *domain.ToolsJob
	for _, job := range r.jobs {
		if job.OperationType != domain.OpTypeConvert {
			continue
		}
		if len(job.InputFiles) != 1 || job.InputFiles[0] != jobID {
			continue
		}
		if latest == nil || job.CreatedAt.After(latest.CreatedAt) {
			cp := *job
			latest = &cp
		}
	}
	return latest, nil
}
func (r *memToolsRepo) Delete(string) error { return nil }
func (r *memToolsRepo) List(int, int, string, string) ([]*domain.ToolsJob, int, error) {
	return nil, 0, nil
}

type captureBroadcaster struct {
	mu      sync.Mutex
	updates []domain.ToolsProgressUpdate
}

func (c *captureBroadcaster) Broadcast(u interface{}) {
	if up, ok := u.(domain.ToolsProgressUpdate); ok {
		c.mu.Lock()
		c.updates = append(c.updates, up)
		c.mu.Unlock()
	}
}

func newTestService(t *testing.T, jobRepo domain.JobRepository) (*Service, *memToolsRepo, *captureBroadcaster) {
	t.Helper()
	toolsRepo := newMemToolsRepo()
	bc := &captureBroadcaster{}
	svc := NewService(&Config{
		ToolsRepository: toolsRepo,
		JobRepository:   jobRepo,
		Broadcaster:     bc,
		DownloadPath:    t.TempDir(),
		ProcessedPath:   t.TempDir(),
		Concurrency:     1,
	})
	return svc, toolsRepo, bc
}

func TestValidateJob(t *testing.T) {
	tests := []struct {
		name    string
		job     *domain.ToolsJob
		wantErr bool
	}{
		{
			name: "valid trim",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeTrim, InputFiles: []string{"v1"}, InputType: domain.InputTypeVideos,
				Parameters: map[string]any{"start_time": "0", "end_time": "10"},
			},
		},
		{
			name:    "no inputs",
			job:     &domain.ToolsJob{OperationType: domain.OpTypeTrim, Parameters: map[string]any{}},
			wantErr: true,
		},
		{
			name: "concat needs two videos",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeConcat, InputFiles: []string{"v1"}, InputType: domain.InputTypeVideos,
				Parameters: map[string]any{"output_format": "mp4"},
			},
			wantErr: true,
		},
		{
			name: "concat playlist single parent ok",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeConcat, InputFiles: []string{"p1"}, InputType: domain.InputTypePlaylist,
				Parameters: map[string]any{"output_format": "mp4"},
			},
		},
		{
			name: "playlist requires single parent",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeTrim, InputFiles: []string{"p1", "p2"}, InputType: domain.InputTypePlaylist,
				Parameters: map[string]any{"start_time": "0", "end_time": "10"},
			},
			wantErr: true,
		},
		{
			name: "invalid trim params",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeTrim, InputFiles: []string{"v1"},
				Parameters: map[string]any{"start_time": "10", "end_time": "5"},
			},
			wantErr: true,
		},
		{
			name: "workflow needs steps",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeWorkflow, InputFiles: []string{"v1"},
				Parameters: map[string]any{"steps": []any{}},
			},
			wantErr: true,
		},
		{
			name: "workflow valid",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeWorkflow, InputFiles: []string{"v1"},
				Parameters: map[string]any{"steps": []any{
					map[string]any{"operation": "rotate", "parameters": map[string]any{"rotation": 90}},
				}},
			},
		},
		{
			name: "workflow invalid step",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeWorkflow, InputFiles: []string{"v1"},
				Parameters: map[string]any{"steps": []any{
					map[string]any{"operation": "rotate", "parameters": map[string]any{}},
				}},
			},
			wantErr: true,
		},
		{
			name: "invalid input type",
			job: &domain.ToolsJob{
				OperationType: domain.OpTypeTrim, InputFiles: []string{"v1"}, InputType: "bogus",
				Parameters: map[string]any{"start_time": "0", "end_time": "10"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJob(tt.job)
			if tt.wantErr && err == nil {
				t.Errorf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSubmitValidatesAndEnqueues(t *testing.T) {
	svc, repo, _ := newTestService(t, testutil.NewMockJobRepository())

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeTrim,
		InputFiles:    []string{"v1"},
		Parameters:    map[string]any{"start_time": "0", "end_time": "10"},
	}
	if err := svc.Submit(job); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	if job.ID == "" {
		t.Error("expected job ID to be assigned")
	}
	if job.Status != domain.ToolsJobStatusPending {
		t.Errorf("expected pending status, got %s", job.Status)
	}
	if job.InputType != domain.InputTypeVideos {
		t.Errorf("expected default input type videos, got %s", job.InputType)
	}
	if stored, _ := repo.GetByID(job.ID); stored == nil {
		t.Error("expected job to be persisted")
	}

	bad := &domain.ToolsJob{OperationType: domain.OpTypeTrim, InputFiles: []string{"v1"},
		Parameters: map[string]any{"start_time": "10", "end_time": "5"}}
	if err := svc.Submit(bad); err == nil {
		t.Error("expected Submit to reject invalid job")
	}
}

func TestCancelJob(t *testing.T) {
	svc, repo, bc := newTestService(t, testutil.NewMockJobRepository())

	job := &domain.ToolsJob{ID: "job1", OperationType: domain.OpTypeTrim, Status: domain.ToolsJobStatusPending}
	_ = repo.Create(job)

	if err := svc.CancelJob("job1"); err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}
	stored, _ := repo.GetByID("job1")
	if stored.Status != domain.ToolsJobStatusCancelled {
		t.Errorf("expected cancelled, got %s", stored.Status)
	}
	if len(bc.updates) == 0 {
		t.Error("expected a broadcast on cancel")
	}

	// Cannot cancel completed jobs.
	done := &domain.ToolsJob{ID: "job2", Status: domain.ToolsJobStatusComplete}
	_ = repo.Create(done)
	if err := svc.CancelJob("job2"); err == nil {
		t.Error("expected error cancelling completed job")
	}

	if err := svc.CancelJob("missing"); err == nil {
		t.Error("expected error cancelling missing job")
	}
}

func TestExpandInputFiles(t *testing.T) {
	jobRepo := testutil.NewMockJobRepository()
	svc, _, _ := newTestService(t, jobRepo)

	// videos pass through unchanged.
	ids, err := svc.expandInputFiles(&domain.ToolsJob{InputType: domain.InputTypeVideos, InputFiles: []string{"a", "b"}})
	if err != nil || len(ids) != 2 {
		t.Fatalf("videos expand = %v, %v", ids, err)
	}

	// playlist expands to its member videos.
	parent := testutil.CreateTestJob("playlist1", "url")
	_ = jobRepo.Create(parent)
	for _, vid := range []string{"v1", "v2"} {
		j := testutil.CreateTestJob(vid, "url")
		_ = jobRepo.Create(j)
		_ = jobRepo.AddVideoToParent(vid, "playlist1", "playlist")
	}
	ids, err = svc.expandInputFiles(&domain.ToolsJob{InputType: domain.InputTypePlaylist, InputFiles: []string{"playlist1"}})
	if err != nil || len(ids) != 2 {
		t.Fatalf("playlist expand = %v, %v", ids, err)
	}
}

func TestResolveVideoPath(t *testing.T) {
	jobRepo := testutil.NewMockJobRepository()
	svc, _, _ := newTestService(t, jobRepo)

	meta := testutil.CreateTestVideoMetadata()
	job := testutil.CreateTestJob("v1", "url")
	_ = jobRepo.Create(job)
	_ = jobRepo.StoreMetadata("v1", meta)

	// Place the file where the resolver expects it (uploader/title.ext).
	target := filepath.Join(svc.downloadPath, sanitizeFilename(meta.Uploader), sanitizeFilename(meta.Title)+"."+meta.Extension)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := svc.resolveVideoPath("v1")
	if err != nil {
		t.Fatalf("resolveVideoPath: %v", err)
	}
	if got != target {
		t.Errorf("resolveVideoPath = %q, want %q", got, target)
	}

	// Missing file errors.
	missing := testutil.CreateTestJob("v2", "url")
	_ = jobRepo.Create(missing)
	_ = jobRepo.StoreMetadata("v2", &domain.VideoMetadata{Uploader: "Nobody", Title: "Nope", Extension: "mp4"})
	if _, err := svc.resolveVideoPath("v2"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestResolveOutputFile(t *testing.T) {
	svc, _, _ := newTestService(t, testutil.NewMockJobRepository())

	// Valid output inside the processed dir.
	good := filepath.Join(svc.processedPath, "out.mp4")
	if err := os.WriteFile(good, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := svc.ResolveOutputFile(&domain.ToolsJob{OutputFile: good}); err != nil || got != good {
		t.Errorf("ResolveOutputFile(good) = %q, %v", got, err)
	}

	// Empty output file.
	if _, err := svc.ResolveOutputFile(&domain.ToolsJob{}); err == nil {
		t.Error("expected error for empty output file")
	}

	// Path escaping the processed directory is rejected.
	if _, err := svc.ResolveOutputFile(&domain.ToolsJob{OutputFile: filepath.Join(svc.processedPath, "..", "escape.mp4")}); err == nil {
		t.Error("expected traversal to be rejected")
	}

	// Missing file.
	if _, err := svc.ResolveOutputFile(&domain.ToolsJob{OutputFile: filepath.Join(svc.processedPath, "missing.mp4")}); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestOrderConcatInputs(t *testing.T) {
	inputs := []resolvedInput{{jobID: "a", path: "a.mp4"}, {jobID: "b", path: "b.mp4"}, {jobID: "c", path: "c.mp4"}}

	got := orderConcatInputs(inputs, []string{"c", "a"})
	wantOrder := []string{"c", "a", "b"} // unspecified keep original order at the end
	for i, id := range wantOrder {
		if got[i].jobID != id {
			t.Errorf("position %d = %q, want %q (full: %+v)", i, got[i].jobID, id, got)
		}
	}

	// No ordering keeps the input order.
	got = orderConcatInputs(inputs, nil)
	if got[0].jobID != "a" || got[2].jobID != "c" {
		t.Errorf("unexpected default order: %+v", got)
	}
}

func TestWriteConcatList(t *testing.T) {
	svc, _, _ := newTestService(t, testutil.NewMockJobRepository())

	inputs := []resolvedInput{{jobID: "a", path: "/videos/a's movie.mp4"}}
	listFile, cleanup, err := svc.writeConcatList(inputs)
	if err != nil {
		t.Fatalf("writeConcatList: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(listFile)
	if err != nil {
		t.Fatalf("read list: %v", err)
	}
	content := string(data)
	if want := `file '/videos/a'\''s movie.mp4'`; want+"\n" != content {
		t.Errorf("concat list = %q, want %q", content, want+"\n")
	}

	cleanup()
	if _, err := os.Stat(listFile); !os.IsNotExist(err) {
		t.Error("expected list file to be removed by cleanup")
	}
}
