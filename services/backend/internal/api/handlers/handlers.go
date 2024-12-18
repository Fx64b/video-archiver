package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"video-archiver/internal/services/download"
	"video-archiver/models"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type Response struct {
	Message interface{} `json:"message"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "*") // or your specific domains
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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

	download.DownloadQueue <- job

	resp := Response{Message: "Video added to download queue"}
	json.NewEncoder(w).Encode(resp)
}

func RecentHandler(w http.ResponseWriter, r *http.Request) {
	recentJobs := download.GetRecentJobs()

	if recentJobs == nil {
		http.Error(w, "No recent downloads found", http.StatusNotFound)
		return
	}

	resp := Response{recentJobs}
	json.NewEncoder(w).Encode(resp)
}

func Handler(r *chi.Mux) {
	r.Use(corsMiddleware)

	r.Post("/download", DownloadHandler)
	r.Get("/recent", RecentHandler)
	// TODO: Other routes (e.g., /categorize, /stream) to be added here
}
