package queue

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
	log.Infof("Starting download for job ID: %s, URL: %s", job.ID, job.URL)

	downloadPath := os.Getenv("DOWNLOAD_PATH")
	if downloadPath == "" {
		downloadPath = "./downloads"
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

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to start yt-dlp for job ID %s: %v", job.ID, err)
		return
	}

	// Track progress in a separate goroutine
	go trackProgress(stdout, job.ID)
	go trackProgress(stderr, job.ID)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		log.Errorf("Failed to download job ID %s: %v", job.ID, err)
	} else {
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

			//log.Infof("Job ID %s: Downloading item %d of %d - %.2f%%", jobID, currentItem, totalItems, progress)
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

func parseToFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}
