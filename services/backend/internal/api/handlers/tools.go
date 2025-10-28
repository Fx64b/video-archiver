package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	return &ToolsHandler{
		toolsService: toolsService,
	}
}

func (h *ToolsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tools", func(r chi.Router) {
		r.Get("/operations", h.HandleListOperations)
		r.Post("/trim", h.HandleTrim)
		r.Post("/concat", h.HandleConcat)
		r.Post("/convert", h.HandleConvert)
		r.Post("/extract-audio", h.HandleExtractAudio)
		r.Post("/adjust-quality", h.HandleAdjustQuality)
		r.Post("/rotate", h.HandleRotate)
		r.Post("/workflow", h.HandleWorkflow)
		r.Get("/jobs", h.HandleListJobs)
		r.Get("/jobs/{id}", h.HandleGetJob)
		r.Delete("/jobs/{id}", h.HandleCancelJob)
	})
}

// Request structures
type ToolsJobRequest struct {
	InputFile  string         `json:"input_file,omitempty"`  // For single file operations
	InputFiles []string       `json:"input_files,omitempty"` // For multi-file operations
	InputType  string         `json:"input_type,omitempty"`  // videos, playlist, channel
	Parameters map[string]any `json:"parameters"`
}

// HandleListOperations returns available tools operations
func (h *ToolsHandler) HandleListOperations(w http.ResponseWriter, r *http.Request) {
	operations := []map[string]interface{}{
		{
			"type":              "trim",
			"name":              "Video Trimmer",
			"description":       "Cut videos by specifying start and end times",
			"supported_formats": []string{"mp4", "mkv", "webm", "avi", "mov"},
		},
		{
			"type":              "concat",
			"name":              "Video Concatenator",
			"description":       "Merge multiple videos into one file",
			"supported_formats": []string{"mp4", "mkv", "webm"},
		},
		{
			"type":              "convert",
			"name":              "Format Converter",
			"description":       "Convert video between formats",
			"supported_formats": []string{"mp4", "webm", "mkv", "avi", "mov"},
		},
		{
			"type":              "extract_audio",
			"name":              "Audio Extractor",
			"description":       "Extract audio track from video",
			"supported_formats": []string{"mp3", "aac", "flac", "wav", "ogg"},
		},
		{
			"type":              "adjust_quality",
			"name":              "Quality Adjuster",
			"description":       "Change video resolution and quality",
			"supported_formats": []string{"mp4"},
		},
		{
			"type":              "rotate",
			"name":              "Video Rotator",
			"description":       "Rotate or flip video orientation",
			"supported_formats": []string{"mp4", "mkv", "webm"},
		},
		{
			"type":              "workflow",
			"name":              "Workflow",
			"description":       "Chain multiple operations together",
			"supported_formats": []string{"mp4", "mkv", "webm"},
		},
	}

	resp := Response{Message: operations}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleTrim creates a trim job
func (h *ToolsHandler) HandleTrim(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.InputFile == "" {
		http.Error(w, "input_file is required", http.StatusBadRequest)
		return
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeTrim,
		InputFiles:    []string{req.InputFile},
		InputType:     "videos",
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit trim job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleConcat creates a concat job
func (h *ToolsHandler) HandleConcat(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(req.InputFiles) == 0 {
		http.Error(w, "input_files is required", http.StatusBadRequest)
		return
	}

	inputType := req.InputType
	if inputType == "" {
		inputType = "videos"
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeConcat,
		InputFiles:    req.InputFiles,
		InputType:     inputType,
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit concat job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleConvert creates a convert job
func (h *ToolsHandler) HandleConvert(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.InputFile == "" {
		http.Error(w, "input_file is required", http.StatusBadRequest)
		return
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeConvert,
		InputFiles:    []string{req.InputFile},
		InputType:     "videos",
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit convert job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleExtractAudio creates an audio extraction job
func (h *ToolsHandler) HandleExtractAudio(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.InputFile == "" {
		http.Error(w, "input_file is required", http.StatusBadRequest)
		return
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeExtractAudio,
		InputFiles:    []string{req.InputFile},
		InputType:     "videos",
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit extract audio job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleAdjustQuality creates a quality adjustment job
func (h *ToolsHandler) HandleAdjustQuality(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.InputFile == "" {
		http.Error(w, "input_file is required", http.StatusBadRequest)
		return
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeAdjustQuality,
		InputFiles:    []string{req.InputFile},
		InputType:     "videos",
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit adjust quality job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleRotate creates a rotate job
func (h *ToolsHandler) HandleRotate(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.InputFile == "" {
		http.Error(w, "input_file is required", http.StatusBadRequest)
		return
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeRotate,
		InputFiles:    []string{req.InputFile},
		InputType:     "videos",
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit rotate job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleWorkflow creates a workflow job
func (h *ToolsHandler) HandleWorkflow(w http.ResponseWriter, r *http.Request) {
	var req ToolsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(req.InputFiles) == 0 {
		http.Error(w, "input_files is required", http.StatusBadRequest)
		return
	}

	inputType := req.InputType
	if inputType == "" {
		inputType = "videos"
	}

	job := &domain.ToolsJob{
		OperationType: domain.OpTypeWorkflow,
		InputFiles:    req.InputFiles,
		InputType:     inputType,
		Parameters:    req.Parameters,
	}

	if err := h.toolsService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit workflow job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleListJobs lists tools jobs with pagination
func (h *ToolsHandler) HandleListJobs(w http.ResponseWriter, r *http.Request) {
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

	status := r.URL.Query().Get("status")
	operationType := r.URL.Query().Get("operation_type")

	jobs, totalCount, err := h.toolsService.GetRepository().List(page, limit, status, operationType)
	if err != nil {
		log.WithError(err).Error("Failed to list tools jobs")
		http.Error(w, "Failed to list jobs", http.StatusInternalServerError)
		return
	}

	if jobs == nil {
		jobs = []*domain.ToolsJob{}
	}

	totalPages := (totalCount + limit - 1) / limit

	resp := struct {
		Items      []*domain.ToolsJob `json:"items"`
		TotalCount int                `json:"total_count"`
		Page       int                `json:"page"`
		Limit      int                `json:"limit"`
		TotalPages int                `json:"total_pages"`
	}{
		Items:      jobs,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: resp})
}

// HandleGetJob gets a specific job by ID
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

	resp := Response{Message: job}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleCancelJob cancels a job
func (h *ToolsHandler) HandleCancelJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	if err := h.toolsService.CancelJob(jobID); err != nil {
		log.WithError(err).Error("Failed to cancel tools job")
		http.Error(w, fmt.Sprintf("Failed to cancel job: %v", err), http.StatusInternalServerError)
		return
	}

	resp := Response{Message: "Job cancelled successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
