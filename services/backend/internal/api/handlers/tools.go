package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/tools"
)

type ToolsHandler struct {
	toolsService *tools.Service
}

func NewToolsHandler(toolsService *tools.Service) *ToolsHandler {
	return &ToolsHandler{toolsService: toolsService}
}

func (h *ToolsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tools", func(r chi.Router) {
		r.Get("/operations", h.HandleListOperations)
		r.Post("/{operation}", h.HandleSubmit)
		r.Get("/jobs", h.HandleListJobs)
		r.Get("/jobs/{id}", h.HandleGetJob)
		r.Get("/jobs/{id}/output", h.HandleServeOutput)
		r.Get("/jobs/{id}/thumbnail", h.HandleServeThumbnail)
		r.Delete("/jobs/{id}", h.HandleCancelJob)
	})
}

// ToolsJobRequest is the body accepted by every operation endpoint. A single
// file may be supplied via input_file or input_files.
type ToolsJobRequest struct {
	InputFile  string         `json:"input_file,omitempty"`
	InputFiles []string       `json:"input_files,omitempty"`
	InputType  string         `json:"input_type,omitempty"`
	Parameters map[string]any `json:"parameters"`
}

// supportedOperations is the single source of truth for the operations the API
// accepts and advertises.
var supportedOperations = []struct {
	Type             domain.ToolsOperationType
	Name             string
	Description      string
	SupportedFormats []string
	MinInputs        int
}{
	{domain.OpTypeTrim, "Trim Video", "Cut videos to a start/end time range", []string{"mp4", "mkv", "webm", "avi", "mov"}, 1},
	{domain.OpTypeConcat, "Concatenate Videos", "Merge multiple videos into one file", []string{"mp4", "mkv", "webm"}, 2},
	{domain.OpTypeConvert, "Convert Format", "Convert videos between container formats", []string{"mp4", "webm", "mkv", "avi", "mov"}, 1},
	{domain.OpTypeExtractAudio, "Extract Audio", "Extract the audio track from a video", []string{"mp3", "aac", "flac", "wav", "ogg"}, 1},
	{domain.OpTypeAdjustQuality, "Adjust Quality", "Change resolution, bitrate or CRF", []string{"mp4"}, 1},
	{domain.OpTypeRotate, "Rotate Video", "Rotate or flip a video", []string{"mp4", "mkv", "webm"}, 1},
	{domain.OpTypeWorkflow, "Workflow", "Chain multiple operations together", []string{"mp4", "mkv", "webm"}, 1},
}

func operationByType(op string) (domain.ToolsOperationType, bool) {
	for _, o := range supportedOperations {
		if string(o.Type) == op {
			return o.Type, true
		}
	}
	return "", false
}

func (h *ToolsHandler) HandleListOperations(w http.ResponseWriter, r *http.Request) {
	operations := make([]map[string]any, 0, len(supportedOperations))
	for _, o := range supportedOperations {
		operations = append(operations, map[string]any{
			"type":              o.Type,
			"name":              o.Name,
			"description":       o.Description,
			"supported_formats": o.SupportedFormats,
			"min_inputs":        o.MinInputs,
		})
	}
	writeJSON(w, http.StatusOK, Response{Message: operations})
}

// HandleSubmit creates a job for the operation named in the URL. It replaces the
// seven near-identical per-operation handlers of the previous implementation.
func (h *ToolsHandler) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	op, ok := operationByType(chi.URLParam(r, "operation"))
	if !ok {
		http.Error(w, "Unknown operation", http.StatusNotFound)
		return
	}

	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	inputFiles := req.InputFiles
	if len(inputFiles) == 0 && req.InputFile != "" {
		inputFiles = []string{req.InputFile}
	}

	inputType := domain.ToolsInputType(req.InputType)
	if inputType == "" {
		inputType = domain.InputTypeVideos
	}

	job := &domain.ToolsJob{
		OperationType: op,
		InputFiles:    inputFiles,
		InputType:     inputType,
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		// Validation failures are client errors; everything else is a 500.
		log.WithError(err).WithField("operation", op).Warn("Failed to submit tools job")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusAccepted, Response{Message: job})
}

func (h *ToolsHandler) HandleListJobs(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)
	if limit < 1 || limit > 100 {
		limit = 20
	}
	status := r.URL.Query().Get("status")
	operationType := r.URL.Query().Get("operation_type")

	jobs, totalCount, err := h.toolsService.Repository().List(page, limit, status, operationType)
	if err != nil {
		log.WithError(err).Error("Failed to list tools jobs")
		http.Error(w, "Failed to list jobs", http.StatusInternalServerError)
		return
	}
	if jobs == nil {
		jobs = []*domain.ToolsJob{}
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	writeJSON(w, http.StatusOK, Response{Message: struct {
		Items      []*domain.ToolsJob `json:"items"`
		TotalCount int                `json:"total_count"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
		TotalPages int                `json:"total_pages"`
	}{jobs, totalCount, page, limit, totalPages}})
}

func (h *ToolsHandler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	job, err := h.toolsService.GetJobByID(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get tools job")
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: job})
}

// HandleServeOutput streams the produced file of a completed job as a download.
func (h *ToolsHandler) HandleServeOutput(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	job, err := h.toolsService.GetJobByID(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get tools job")
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	if job.Status != domain.ToolsJobStatusComplete {
		http.Error(w, "Job output is not available yet", http.StatusConflict)
		return
	}

	path, err := h.toolsService.ResolveOutputFile(job)
	if err != nil {
		log.WithError(err).WithField("job_id", jobID).Warn("Tools output file not found")
		http.Error(w, "Output file not found", http.StatusNotFound)
		return
	}

	// inline=1 lets the frontend preview the file (video/audio playback in
	// the browser) instead of forcing a download.
	disposition := "attachment"
	if r.URL.Query().Get("inline") == "1" {
		disposition = "inline"
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", disposition, filepath.Base(path)))
	http.ServeFile(w, r, path)
}

// HandleServeThumbnail serves the poster image of a completed video job,
// generating it on first request when it doesn't exist yet (which also covers
// jobs completed before thumbnails were introduced). Audio outputs have no
// poster and return 404.
func (h *ToolsHandler) HandleServeThumbnail(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	job, err := h.toolsService.GetJobByID(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get tools job")
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	if job.Status != domain.ToolsJobStatusComplete {
		http.Error(w, "Job output is not available yet", http.StatusConflict)
		return
	}

	path, err := h.toolsService.ThumbnailPath(job)
	if err != nil {
		log.WithError(err).WithField("job_id", jobID).Debug("No thumbnail for tools job")
		http.Error(w, "Thumbnail not available", http.StatusNotFound)
		return
	}

	// A job's thumbnail never changes once generated, and it is deleted
	// together with the job, so long-lived caching is safe.
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, path)
}

// HandleCancelJob cancels a queued/running job, or deletes a finished one
// (record and output file) so processed files don't accumulate forever.
func (h *ToolsHandler) HandleCancelJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	job, err := h.toolsService.GetJobByID(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get tools job")
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.Status == domain.ToolsJobStatusPending || job.Status == domain.ToolsJobStatusProcessing {
		if err := h.toolsService.CancelJob(jobID); err != nil {
			log.WithError(err).Error("Failed to cancel tools job")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, Response{Message: "Job cancelled successfully"})
		return
	}

	if err := h.toolsService.DeleteJob(jobID); err != nil {
		log.WithError(err).Error("Failed to delete tools job")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: "Job deleted successfully"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}

func parseIntQuery(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}
