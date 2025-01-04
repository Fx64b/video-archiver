package download

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"regexp"
	"strconv"
)

type ProgressUpdate struct {
	JobID       string  `json:"jobID"`
	CurrentItem int     `json:"currentItem"`
	TotalItems  int     `json:"totalItems"`
	Progress    float64 `json:"progress"`
}

func (s *Service) trackProgress(pipe io.Reader, jobID string) {
	scanner := bufio.NewScanner(pipe)

	const maxCapacity = 1024 * 1024 // 1MB buffer due to some problems with yt-dlp logs
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	itemRegex := regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)
	progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%`)

	var totalItems, currentItem int
	var progress float64

	for scanner.Scan() {
		line := scanner.Text()
		log.Debug(line)

		if match := itemRegex.FindStringSubmatch(line); match != nil {
			currentItem = s.parseToInt(match[1])
			totalItems = s.parseToInt(match[2])

			progress = (float64(currentItem-1) / float64(totalItems)) * 100

			update := ProgressUpdate{
				JobID:       jobID,
				CurrentItem: currentItem,
				TotalItems:  totalItems,
				Progress:    progress,
			}

			if err := s.updateJobProgress(jobID, progress); err != nil {
				log.WithError(err).Error("Failed to update job progress")
			}

			// Broadcast progress via WebSocket hub
			s.hub.broadcast <- update
		} else if match := progressRegex.FindStringSubmatch(line); match != nil {
			if currentProgress, err := strconv.ParseFloat(match[1], 64); err == nil {
				if totalItems > 0 {
					// If we're downloading multiple items, add the progress for the current item
					itemProgress := currentProgress / float64(totalItems)
					baseProgress := (float64(currentItem-1) / float64(totalItems)) * 100
					progress = baseProgress + itemProgress
				} else {
					progress = currentProgress
				}

				update := ProgressUpdate{
					JobID:       jobID,
					CurrentItem: currentItem,
					TotalItems:  totalItems,
					Progress:    progress,
				}

				// Update job in database
				if err := s.updateJobProgress(jobID, progress); err != nil {
					log.WithError(err).Error("Failed to update job progress")
				}

				// Broadcast progress via WebSocket hub
				s.hub.broadcast <- update
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.WithError(err).Error("Error reading progress output")
	}
}

func (s *Service) parseToInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}

func (s *Service) updateJobProgress(jobID string, progress float64) error {
	job, err := s.jobs.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	job.Progress = progress
	return s.jobs.Update(job)
}
