package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	queue "video-archiver/internal/services/download"
	"video-archiver/models"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type Response struct {
	Message string `json:"message"`
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Infof("Received download request for URL: %s", req.URL)

	job := models.DownloadJob{
		ID:        uuid.New().String(),
		URL:       req.URL,
		TIMESTAMP: time.Now(),
	}

	queue.DownloadQueue <- job

	resp := Response{Message: "Video added to download queue"}
	json.NewEncoder(w).Encode(resp)
}

func Handler(r *chi.Mux) {
	r.Post("/download", DownloadHandler)
	// TODO: Other routes (e.g., /progress, /categorize, /stream) to be added here
}
