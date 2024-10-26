package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"video-archiver/internal/queue"
	"video-archiver/models"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

type Response struct {
	Message string `json:"message"`
}

/*var youtubeURLRegex = regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.be)/(watch\?v=|embed/|v/|.+\?v=)?([^&=%\?]{11})`)

func isValidYouTubeURL(url string) bool {
	return youtubeURLRegex.Match([]byte(url))
}*/

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	/*	if !isValidYouTubeURL(req.URL) {
		http.Error(w, "Invalid YouTube URL", http.StatusBadRequest)
		return
	}*/

	log.Infof("Received download request for URL: %s", req.URL)
	// TODO: Call yt-dlp command here and handle response

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
