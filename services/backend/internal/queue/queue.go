package queue

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"video-archiver/models"
)

var DownloadQueue = make(chan models.DownloadJob, 10)

func StartQueueWorker() {
	go func() {
		for job := range DownloadQueue {
			processJob(job)
		}
	}()
}

func processJob(job models.DownloadJob) {
	logrus.Infof("Starting download for job ID: %s, URL: %s", job.ID, job.URL)

	downloadPath := os.Getenv("DOWNLOAD_PATH")
	if downloadPath == "" {
		downloadPath = "./downloads" // Default to relative path in the container
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
		"--output", outputPath,
		job.URL,
	)

	err := cmd.Run()

	if err != nil {
		logrus.Errorf("Failed to download job ID %s: %v", job.ID, err)
	} else {
		logrus.Infof("Successfully downloaded job ID: %s", job.ID)
	}
}
