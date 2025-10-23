package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/download"
	"video-archiver/internal/util/statistics"
)

type DownloadRequest struct {
	URL     string `json:"url"`
	Quality *int   `json:"quality,omitempty"`
}

type Response struct {
	Message interface{} `json:"message"`
}

type Handler struct {
	downloadService    *download.Service
	downloadPath       string
	settingsRepository domain.SettingsRepository
	allowedOrigins     []string
}

func NewHandler(downloadService *download.Service, downloadPath string, settingsRepository domain.SettingsRepository, allowedOrigins []string) *Handler {
	return &Handler{
		downloadService:    downloadService,
		downloadPath:       downloadPath,
		settingsRepository: settingsRepository,
		allowedOrigins:     allowedOrigins,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Use(h.corsMiddleware)

	r.Post("/download", h.HandleDownload)
	r.Get("/recent", h.HandleRecent)
	r.Get("/job/{id}", h.HandleGetJob)
	r.Get("/job/{id}/parents", h.HandleGetJobParents)
	r.Get("/statistics", h.HandleGetStatistics)
	r.Get("/downloads/{type}", h.HandleGetDownloads)
	r.Get("/video/{jobID}", h.HandleServeVideo)
	r.Get("/settings", h.HandleGetSettings)
	r.Put("/settings", h.HandleUpdateSettings)
}

func (h *Handler) RegisterWSRoutes(r *chi.Mux) {
	r.Use(h.corsMiddleware)
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

	// Validate custom quality if provided
	if req.Quality != nil {
		validQualities := []int{360, 480, 720, 1080, 1440, 2160}
		qualityValid := false
		for _, q := range validQualities {
			if *req.Quality == q {
				qualityValid = true
				break
			}
		}
		if !qualityValid {
			http.Error(w, "Invalid quality. Must be 360, 480, 720, 1080, 1440, or 2160", http.StatusBadRequest)
			return
		}
	}

	if req.Quality != nil {
		log.Infof("Received download request for URL: %s with custom quality: %dp", req.URL, *req.Quality)
	} else {
		log.Infof("Received download request for URL: %s", req.URL)
	}

	job := domain.Job{
		ID:              uuid.New().String(),
		URL:             req.URL,
		CustomQuality:   req.Quality,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
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
		// Return empty array with 200 status instead of 404
		resp := Response{Message: []interface{}{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := Response{Message: recentJobs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	log.WithField("jobID", jobID).Info("Received request for job")

	if jobID == "" {
		log.Warn("Missing job ID in request")
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

	resp := Response{Message: jobWithMetadata}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGetJobParents(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")
	log.WithField("jobID", jobID).Info("Received request for job parents")

	if jobID == "" {
		log.Warn("Missing job ID in request")
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	jobRepo := h.downloadService.GetRepository()
	parents, err := jobRepo.GetParentsForVideo(jobID)
	if err != nil {
		log.WithError(err).Error("Failed to get parents for video")
		http.Error(w, "Failed to get parents for video", http.StatusInternalServerError)
		return
	}

	// Return empty array if no parents found (not an error)
	if parents == nil {
		parents = []*domain.JobWithMetadata{}
	}

	resp := Response{Message: parents}
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
		log.Infof("No %s found", contentType)

		resp := struct {
			Items      []*domain.JobWithMetadata `json:"items"`
			TotalCount int                       `json:"total_count"`
			Page       int                       `json:"page"`
			Limit      int                       `json:"limit"`
			TotalPages int                       `json:"total_pages"`
		}{
			Items:      []*domain.JobWithMetadata{},
			TotalCount: 0,
			Page:       page,
			Limit:      limit,
			TotalPages: 0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Message: resp})
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

	// Try to find the video file based on the metadata
	var videoPath string

	switch metadata := jobWithMetadata.Metadata.(type) {
	case *domain.VideoMetadata:
		// For videos, construct the expected file path
		channelDir := metadata.Channel
		if channelDir == "" {
			channelDir = metadata.Uploader
		}
		if channelDir == "" {
			channelDir = "Unknown"
		}

		title := metadata.Title
		if title == "" {
			title = "Unknown"
		}

		// Clean the filename to prevent path traversal
		title = sanitizeFilename(title)
		channelDir = sanitizeFilename(channelDir)

		videoPath = filepath.Join(h.downloadPath, channelDir, title+".mp4")

	default:
		http.Error(w, "Unsupported content type for video playback", http.StatusBadRequest)
		return
	}

	// Validate the path to prevent directory traversal attacks
	if err := validateVideoPath(h.downloadPath, videoPath); err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"requested_path": videoPath,
			"base_path":     h.downloadPath,
			"remote_ip":     r.RemoteAddr,
		}).Warn("Path traversal attempt blocked")
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if the file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		log.WithField("path", videoPath).Warn("Video file not found")
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}

	log.WithField("path", videoPath).Info("Serving video file")

	// Set appropriate headers for video streaming
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "no-cache")

	// Serve the file
	http.ServeFile(w, r, videoPath)
}

// sanitizeFilename removes dangerous characters from filenames
func sanitizeFilename(filename string) string {
	// Replace path separators and other dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")
	return filename
}

// validateVideoPath ensures the requested path is within the allowed download directory
func validateVideoPath(basePath, videoPath string) error {
	// Clean and get absolute paths
	cleanBasePath, err := filepath.Abs(filepath.Clean(basePath))
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	cleanVideoPath, err := filepath.Abs(filepath.Clean(videoPath))
	if err != nil {
		return fmt.Errorf("failed to resolve video path: %w", err)
	}

	// Check if the video path is within the base path
	if !strings.HasPrefix(cleanVideoPath, cleanBasePath) {
		return fmt.Errorf("path traversal detected: video path %s is outside base path %s",
			cleanVideoPath, cleanBasePath)
	}

	return nil
}


func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsRepository.Get()
	if err != nil {
		log.WithError(err).Error("Failed to get settings")
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	resp := Response{Message: settings}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

	// Validate theme
	if req.Theme != "light" && req.Theme != "dark" && req.Theme != "system" {
		http.Error(w, "Invalid theme. Must be 'light', 'dark', or 'system'", http.StatusBadRequest)
		return
	}

	// Validate download quality
	validQualities := []int{360, 480, 720, 1080, 1440, 2160}
	qualityValid := false
	for _, q := range validQualities {
		if req.DownloadQuality == q {
			qualityValid = true
			break
		}
	}
	if !qualityValid {
		http.Error(w, "Invalid download quality. Must be 360, 480, 720, 1080, 1440, or 2160", http.StatusBadRequest)
		return
	}

	// Validate concurrent downloads
	if req.ConcurrentDownloads < 1 || req.ConcurrentDownloads > 10 {
		http.Error(w, "Invalid concurrent downloads. Must be between 1 and 10", http.StatusBadRequest)
		return
	}

	// Get current settings
	settings, err := h.settingsRepository.Get()
	if err != nil {
		log.WithError(err).Error("Failed to get current settings")
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	// Update settings
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

	resp := Response{Message: settings}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// corsMiddleware implements secure CORS handling with origin validation
func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if the origin is in the allowed list
		allowed := false
		for _, allowedOrigin := range h.allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// Only set CORS headers if origin is allowed
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin != "" {
			// Log blocked origin attempts for security monitoring
			log.WithFields(log.Fields{
				"origin":     origin,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remote_ip":  r.RemoteAddr,
			}).Warn("CORS request from unauthorized origin blocked")
		}

		if r.Method == "OPTIONS" {
			if allowed {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
