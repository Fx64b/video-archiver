package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	URL     string `json:"url"`
	Quality *int   `json:"quality,omitempty"`
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
}

func NewHandler(downloadService *download.Service, downloadPath string, settingsRepository domain.SettingsRepository) *Handler {
	return &Handler{
		downloadService:    downloadService,
		downloadPath:       downloadPath,
		settingsRepository: settingsRepository,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Use(corsMiddleware)

	r.Post("/download", h.HandleDownload)
	r.Delete("/download/{id}", h.HandleCancelDownload)
	r.Get("/recent", h.HandleRecent)
	r.Get("/job/{id}", h.HandleGetJob)
	r.Get("/job/{id}/parents", h.HandleGetJobParents)
	r.Get("/job/{id}/videos", h.HandleGetJobVideos)
	r.Get("/statistics", h.HandleGetStatistics)
	r.Get("/downloads/{type}", h.HandleGetDownloads)
	r.Get("/video/{jobID}", h.HandleServeVideo)
	r.Get("/settings", h.HandleGetSettings)
	r.Put("/settings", h.HandleUpdateSettings)
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

	// Set up ping/pong handlers for connection health
	const (
		pongWait   = 60 * time.Second
		pingPeriod = (pongWait * 9) / 10 // Send pings at 90% of pong deadline
	)

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker in a goroutine
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.WithError(err).Debug("Failed to send ping")
					return
				}
			case <-done:
				return
			}
		}
	}()
	defer close(done)

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

	if req.Quality != nil {
		log.Infof("Received download request for URL: %s with custom quality: %dp", req.URL, *req.Quality)
	} else {
		log.Infof("Received download request for URL: %s", req.URL)
	}

	job := domain.Job{
		ID:            uuid.New().String(),
		URL:           req.URL,
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

	items, totalCount, err := h.downloadService.GetRepository().GetMetadataByType(contentType, page, limit, sortBy, order)
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

	// Locate the file on disk. yt-dlp's filename sanitization (and the
	// container extension) cannot be reliably reconstructed from the
	// title, so this scans the download directory and matches by title.
	videoPath, err := tools.ResolveVideoFile(h.downloadPath, metadata)
	if err != nil {
		log.WithField("title", metadata.Title).Warn("Video file not found")
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		log.WithField("path", videoPath).Warn("Video file not found")
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	log.WithField("path", videoPath).Debug("Serving video file")

	// ServeFile handles range requests and content type detection.
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, videoPath)
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
