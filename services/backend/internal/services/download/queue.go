package download

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"video-archiver/internal/services/metadata"
	"video-archiver/internal/storage"
	jobs "video-archiver/internal/utils"
	"video-archiver/models"
)

var DownloadQueue = make(chan models.DownloadJob, 10)
var clients = make(map[*websocket.Conn]bool)
var wsMutex sync.Mutex
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func StartQueueWorker() {
	go func() {
		for job := range DownloadQueue {
			processJob(job)
		}
	}()
}

func processJob(job models.DownloadJob) {
	log.Infof("Starting download for job ID: %s, URL: %s", job.ID, job.URL)

	downloadPath := os.Getenv("DOWNLOAD_PATH")
	if downloadPath == "" {
		downloadPath = "./data/downloads"
	}

	var isPlaylist bool = !jobs.IsVideo(job.URL)

	// job needs to be added before processing
	err := storage.AddJob(job.ID, job.URL, isPlaylist)
	if err != nil {
		log.Errorf("failed to add job to queue with job ID %s: %v", job.ID, err)
	}

	outputPath := fmt.Sprintf("%s/%%(uploader)s/%%(playlist|title)s/%%(title)s.%%(ext)s", downloadPath)

	cmd := exec.Command("yt-dlp",
		"-N", "8",
		"--format", "bestvideo[height<=1080]+bestaudio/best",
		"--merge-output-format", "mp4",
		"--retries", "3",
		"--continue",
		"--ignore-errors",
		"--add-metadata",
		"--write-info-json",
		"--write-playlist-metafiles",
		"--output", outputPath,
		job.URL,
	)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to start yt-dlp for job ID %s: %v", job.ID, err)
	}

	go trackProgress(stdout, job.ID)
	go trackProgress(stderr, job.ID)

	if err := cmd.Wait(); err != nil {
		log.Errorf("Failed to download job ID %s: %v", job.ID, err)
		err := storage.UpdateJobProgress(job.ID, 100.0, "error")
		if err != nil {
			log.Errorf("failed to finish errored job progress for job ID %s: %v", job.ID, err)
		}
	} else {
		err := storage.UpdateJobProgress(job.ID, 100.0, "completed")
		if err != nil {
			log.Errorf("failed to finish successful job progress for job ID %s: %v", job.ID, err)
		}

		if err := metadata.ExtractAndStoreMetadata(job.ID, downloadPath, isPlaylist); err != nil {
			log.Errorf("Failed to extract metadata for job ID %s: %v", job.ID, err)
		}

		log.Infof("Successfully downloaded job ID: %s", job.ID)
	}
}

func trackProgress(pipe io.Reader, jobID string) {
	scanner := bufio.NewScanner(pipe)
	itemRegex := regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)

	var totalItems, currentItem int
	var progress float64

	for scanner.Scan() {
		line := scanner.Text()

		if match := itemRegex.FindStringSubmatch(line); match != nil {
			currentItem = parseToInt(match[1])
			totalItems = parseToInt(match[2])

			progress = (float64(currentItem) / float64(totalItems)) * 100
			progress = float64(int(progress*100)) / 100

			err := storage.UpdateJobProgress(jobID, progress, "in_progress")
			if err != nil {
				log.Errorf("Failed to update job progress for job ID %s: %v", jobID, err)
			}

			sendProgressUpdate(jobID, currentItem, totalItems, progress)
			fmt.Printf("Job ID %s: Downloading item %d of %d - %.2f%%\n", jobID, currentItem, totalItems, progress)
		}
	}
}

func parseToInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	wsMutex.Lock()
	clients[ws] = true
	wsMutex.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			wsMutex.Lock()
			delete(clients, ws)
			wsMutex.Unlock()
			break
		}
	}
}

func sendProgressUpdate(jobID string, currentItem, totalItems int, progress float64) {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	message := fmt.Sprintf(`{"jobID":"%s","currentItem":%d,"totalItems":%d,"progress":%.2f}`, jobID, currentItem, totalItems, progress)

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Errorf("Failed to send WebSocket message: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func StartWebSocketServer() {
	http.HandleFunc("/ws", WebSocketHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
