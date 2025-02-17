package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/download"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type Response struct {
	Message interface{} `json:"message"`
}

type Handler struct {
	downloadService *download.Service
}

func NewHandler(downloadService *download.Service) *Handler {
	return &Handler{
		downloadService: downloadService,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Use(corsMiddleware)

	// Routes at root level
	r.Post("/download", h.HandleDownload)
	r.Get("/recent", h.HandleRecent)
}

func (h *Handler) RegisterWSRoutes(r *chi.Mux) {
	r.Use(corsMiddleware)
	r.Get("/ws", h.HandleWebSocket)
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Info("WebSocket connection attempt received")

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
		log.Info("Cleaning up WebSocket connection")
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
				log.WithError(err).Info("WebSocket connection closed")
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
	recentJobs, err := h.downloadService.GetRecent(5)
	if err != nil {
		log.WithError(err).Error("Failed to get recent jobs")
		http.Error(w, "Failed to get recent jobs", http.StatusInternalServerError)
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
