package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/download"
	"video-archiver/internal/services/tools"
	"video-archiver/internal/util/statistics"
)

type DownloadRequest struct {
	URL       string `json:"url"`
	Quality   *int   `json:"quality,omitempty"`
	MediaType string `json:"media_type,omitempty"` // "video" (default) or "audio"
}

type Response struct {
	Message interface{} `json:"message"`
}

// paginatedJobs is the envelope returned by list endpoints.
type paginatedJobs struct {
	Items      []*domain.JobWithMetadata `json:"items"`
	TotalCount int                       `json:"total_count"`
	Page       int                       `json:"page"`
	Limit      int                       `json:"limit"`
	TotalPages int                       `json:"total_pages"`
}

// validQualities lists the download resolutions accepted by the API.
var validQualities = []int{360, 480, 720, 1080, 1440, 2160}

func isValidQuality(quality int) bool {
	for _, q := range validQualities {
		if quality == q {
			return true
		}
	}
	return false
}

type Handler struct {
	downloadService    *download.Service
	downloadPath       string
	settingsRepository domain.SettingsRepository
	toolsService       *tools.Service
	toolsRepository    domain.ToolsRepository
	ffmpeg             *tools.FFmpeg
}

func NewHandler(downloadService *download.Service, downloadPath string, settingsRepository domain.SettingsRepository,
	toolsService *tools.Service, toolsRepository domain.ToolsRepository, ffmpeg *tools.FFmpeg) *Handler {
	return &Handler{
		downloadService:    downloadService,
		downloadPath:       downloadPath,
		settingsRepository: settingsRepository,
		toolsService:       toolsService,
		toolsRepository:    toolsRepository,
		ffmpeg:             ffmpeg,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Use(corsMiddleware)

	r.Post("/download", h.HandleDownload)
	r.Delete("/download/{id}", h.HandleCancelDownload)
	r.Get("/recent", h.HandleRecent)
	r.Get("/job/{id}", h.HandleGetJob)
	r.Delete("/job/{id}", h.HandleDeleteJob)
	r.Get("/job/{id}/parents", h.HandleGetJobParents)
	r.Get("/job/{id}/videos", h.HandleGetJobVideos)
	r.Get("/job/{id}/tags", h.HandleGetJobTags)
	r.Post("/job/{id}/tags", h.HandleAddJobTags)
	r.Delete("/job/{id}/tags/{tagID}", h.HandleRemoveJobTag)
	r.Get("/tags", h.HandleListTags)
	r.Get("/statistics", h.HandleGetStatistics)
	r.Get("/downloads/{type}", h.HandleGetDownloads)
	r.Get("/video/{jobID}", h.HandleServeVideo)
	r.Get("/video/{jobID}/playback-info", h.HandlePlaybackInfo)
	r.Post("/video/{jobID}/transcode", h.HandleRequestTranscode)
	r.Get("/settings", h.HandleGetSettings)
	r.Put("/settings", h.HandleUpdateSettings)
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

	// Connection health: pings are sent by the hub's per-client writer (the
	// connection's single writer); this side only refreshes the read deadline
	// as pongs arrive.
	const pongWait = 60 * time.Second

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Read loop - waits for messages or pong responses
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

	if req.Quality != nil && !isValidQuality(*req.Quality) {
		http.Error(w, "Invalid quality. Must be 360, 480, 720, 1080, 1440, or 2160", http.StatusBadRequest)
		return
	}

	mediaType := domain.MediaTypeVideo
	switch req.MediaType {
	case "", string(domain.MediaTypeVideo):
	case string(domain.MediaTypeAudio):
		mediaType = domain.MediaTypeAudio
	default:
		http.Error(w, "Invalid media_type. Must be 'video' or 'audio'", http.StatusBadRequest)
		return
	}

	if req.Quality != nil {
		log.Infof("Received %s download request for URL: %s with custom quality: %dp", mediaType, req.URL, *req.Quality)
	} else {
		log.Infof("Received %s download request for URL: %s", mediaType, req.URL)
	}

	job := domain.Job{
		ID:            uuid.New().String(),
		URL:           req.URL,
		MediaType:     mediaType,
		CustomQuality: req.Quality,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.downloadService.Submit(job); err != nil {
		log.WithError(err).Error("Failed to submit job")
		http.Error(w, "Failed to submit job", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, Response{Message: "Video added to download queue"})
}

func (h *Handler) HandleCancelDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	log.Infof("Received cancel request for job: %s", id)

	if err := h.downloadService.CancelJob(id); err != nil {
		log.WithError(err).Error("Failed to cancel job")
		http.Error(w, fmt.Sprintf("Failed to cancel job: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, Response{Message: "Download cancelled successfully"})
}

func (h *Handler) HandleRecent(w http.ResponseWriter, r *http.Request) {
	recentJobs, err := h.downloadService.GetRecentWithMetadata(5)
	if err != nil {
		log.WithError(err).Error("Failed to get recent jobs with metadata")
		http.Error(w, "Failed to get recent jobs with metadata", http.StatusInternalServerError)
		return
	}

	if recentJobs == nil {
		recentJobs = []*domain.JobWithMetadata{}
	}

	writeJSON(w, http.StatusOK, Response{Message: recentJobs})
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

	if jobWithMetadata == nil {
		log.WithField("jobID", jobID).Warn("Job not found")
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, Response{Message: jobWithMetadata})
}

// HandleDeleteJob removes a download from the library, including the media
// file on disk for videos. Running jobs are cancelled before deletion.
func (h *Handler) HandleDeleteJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	log.Infof("Received delete request for job: %s", jobID)

	if err := h.downloadService.DeleteJob(jobID); err != nil {
		log.WithError(err).Error("Failed to delete job")
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete job", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, Response{Message: "Download deleted successfully"})
}

func (h *Handler) HandleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.downloadService.GetRepository().ListTags()
	if err != nil {
		log.WithError(err).Error("Failed to list tags")
		http.Error(w, "Failed to list tags", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: tags})
}

func (h *Handler) HandleGetJobTags(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	tags, err := h.downloadService.GetRepository().GetTagsForJob(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get tags for job")
		http.Error(w, "Failed to get tags", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: tags})
}

type AddTagsRequest struct {
	Tags []string `json:"tags"`
}

func (h *Handler) HandleAddJobTags(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	var req AddTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if len(req.Tags) == 0 {
		http.Error(w, "At least one tag is required", http.StatusBadRequest)
		return
	}

	// Make sure the job exists before tagging it.
	if _, err := h.downloadService.GetRepository().GetByID(jobID); err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	tags, err := h.downloadService.GetRepository().AddTagsToJob(jobID, req.Tags, domain.TagSourceUser)
	if err != nil {
		log.WithError(err).Error("Failed to add tags to job")
		http.Error(w, "Failed to add tags", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: tags})
}

func (h *Handler) HandleRemoveJobTag(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	tagIDParam := chi.URLParam(r, "tagID")
	if jobID == "" || tagIDParam == "" {
		http.Error(w, "Missing job ID or tag ID", http.StatusBadRequest)
		return
	}

	tagID, err := strconv.ParseInt(tagIDParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	if err := h.downloadService.GetRepository().RemoveTagFromJob(jobID, tagID); err != nil {
		log.WithError(err).Error("Failed to remove tag from job")
		http.Error(w, "Failed to remove tag", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, Response{Message: "Tag removed successfully"})
}

func (h *Handler) HandleGetJobParents(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	parents, err := h.downloadService.GetRepository().GetParentsForVideo(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get parents for video")
		http.Error(w, "Failed to get parents for video", http.StatusInternalServerError)
		return
	}

	if parents == nil {
		parents = []*domain.JobWithMetadata{}
	}

	writeJSON(w, http.StatusOK, Response{Message: parents})
}

func (h *Handler) HandleGetJobVideos(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	videos, err := h.downloadService.GetRepository().GetVideosForParent(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get videos for parent")
		http.Error(w, "Failed to get videos for parent", http.StatusInternalServerError)
		return
	}

	if videos == nil {
		videos = []*domain.JobWithMetadata{}
	}

	writeJSON(w, http.StatusOK, Response{Message: videos})
}

func (h *Handler) HandleGetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := statistics.GetStatistics(h.downloadService.GetRepository(), h.downloadPath)
	if err != nil {
		log.WithError(err).Error("Failed to get statistics")
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, Response{Message: stats})
}

func (h *Handler) HandleGetDownloads(w http.ResponseWriter, r *http.Request) {
	contentType := chi.URLParam(r, "type")
	if contentType != "videos" && contentType != "playlists" && contentType != "channels" {
		http.Error(w, "Invalid content type. Must be 'videos', 'playlists', or 'channels'", http.StatusBadRequest)
		return
	}

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)
	if limit > 100 {
		limit = 20
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	items, totalCount, err := h.downloadService.GetRepository().GetMetadataByType(contentType, domain.MetadataQuery{
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Order:  order,
		Search: r.URL.Query().Get("search"),
		Tag:    r.URL.Query().Get("tag"),
	})
	if err != nil {
		log.WithError(err).Errorf("Failed to get %s", contentType)
		http.Error(w, fmt.Sprintf("Failed to get %s", contentType), http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []*domain.JobWithMetadata{}
	}

	writeJSON(w, http.StatusOK, Response{Message: paginatedJobs{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: (totalCount + limit - 1) / limit, // Ceiling division
	}})
}

func (h *Handler) HandleServeVideo(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	// Get job metadata to find the video file
	jobWithMetadata, err := h.downloadService.GetJobWithMetadata(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get job with metadata")
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	if jobWithMetadata == nil || jobWithMetadata.Job == nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	metadata, ok := jobWithMetadata.Metadata.(*domain.VideoMetadata)
	if !ok {
		http.Error(w, "Unsupported content type for video playback", http.StatusBadRequest)
		return
	}

	videoPath, err := h.locateVideoFile(jobWithMetadata.Job, metadata)
	if err != nil {
		log.WithField("title", metadata.Title).Warn("Video file not found")
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	log.WithField("path", videoPath).Debug("Serving video file")

	// ServeFile handles range requests and content type detection.
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, videoPath)
}

// locateVideoFile finds a video job's media file, preferring the path recorded
// at download time. When the file is only found by searching the download
// directory (legacy rows, moved files), the result is persisted so the next
// lookup is direct.
func (h *Handler) locateVideoFile(job *domain.Job, metadata *domain.VideoMetadata) (string, error) {
	path, err := tools.ResolveVideoFileWithHint(h.downloadPath, job.FilePath, metadata)
	if err != nil {
		return "", err
	}
	if path != job.FilePath {
		if err := h.downloadService.GetRepository().SetFilePath(job.ID, path); err != nil {
			log.WithError(err).WithField("jobID", job.ID).Warn("Failed to persist resolved file path")
		}
	}
	return path, nil
}

func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsRepository.Get()
	if err != nil {
		log.WithError(err).Error("Failed to get settings")
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, Response{Message: settings})
}

type UpdateSettingsRequest struct {
	Theme               string `json:"theme"`
	DownloadQuality     int    `json:"download_quality"`
	ConcurrentDownloads int    `json:"concurrent_downloads"`
	// Tools preferences are optional (pointers) so requests that omit them —
	// like the current settings page — don't clobber stored values.
	ToolsDefaultFormat    *string `json:"tools_default_format,omitempty"`
	ToolsDefaultQuality   *string `json:"tools_default_quality,omitempty"`
	ToolsPreserveOriginal *bool   `json:"tools_preserve_original,omitempty"`
	ToolsOutputPath       *string `json:"tools_output_path,omitempty"`
}

func (h *Handler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Theme != "light" && req.Theme != "dark" && req.Theme != "system" {
		http.Error(w, "Invalid theme. Must be 'light', 'dark', or 'system'", http.StatusBadRequest)
		return
	}

	if !isValidQuality(req.DownloadQuality) {
		http.Error(w, "Invalid download quality. Must be 360, 480, 720, 1080, 1440, or 2160", http.StatusBadRequest)
		return
	}

	if req.ConcurrentDownloads < 1 || req.ConcurrentDownloads > 10 {
		http.Error(w, "Invalid concurrent downloads. Must be between 1 and 10", http.StatusBadRequest)
		return
	}

	settings, err := h.settingsRepository.Get()
	if err != nil {
		log.WithError(err).Error("Failed to get current settings")
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	settings.Theme = req.Theme
	settings.DownloadQuality = req.DownloadQuality
	settings.ConcurrentDownloads = req.ConcurrentDownloads
	if req.ToolsDefaultFormat != nil {
		settings.ToolsDefaultFormat = *req.ToolsDefaultFormat
	}
	if req.ToolsDefaultQuality != nil {
		settings.ToolsDefaultQuality = *req.ToolsDefaultQuality
	}
	if req.ToolsPreserveOriginal != nil {
		settings.ToolsPreserveOriginal = *req.ToolsPreserveOriginal
	}
	if req.ToolsOutputPath != nil {
		settings.ToolsOutputPath = *req.ToolsOutputPath
	}

	if err := h.settingsRepository.Update(settings); err != nil {
		log.WithError(err).Error("Failed to update settings")
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	log.Infof("Settings updated successfully - Theme: %s, Quality: %dp, Concurrent Downloads: %d",
		settings.Theme, settings.DownloadQuality, settings.ConcurrentDownloads)

	writeJSON(w, http.StatusOK, Response{Message: settings})
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
