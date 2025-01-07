package download

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type ProgressUpdate struct {
	JobID                string  `json:"jobID"`
	JobType              string  `json:"jobType"`
	CurrentItem          int     `json:"currentItem"`
	TotalItems           int     `json:"totalItems"`
	Progress             float64 `json:"progress"`
	CurrentVideoProgress float64 `json:"currentVideoProgress"`
}

func (s *Service) trackProgress(pipe io.Reader, jobID string) {
	reader := bufio.NewReader(pipe)
	var line bytes.Buffer

	itemRegex := regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)
	progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)% of\s+\d+\.?\d*\w+`)
	destinationRegex := regexp.MustCompile(`\[download\] Destination: .+\.(f\d+)\.(mp4|webm)`)

	var totalItems, currentItem int
	var overallProgress float64
	isPlaylist := false
	jobType := "video"

	for {
		char, err := reader.ReadByte()
		if err != nil {
			if err != io.EOF {
				log.WithError(err).Error("Error reading progress output")
			}
			break
		}

		if char == '\r' || char == '\n' {
			currentLine := line.String()
			line.Reset()

			log.Debug(currentLine)
			if len(strings.TrimSpace(currentLine)) == 0 {
				continue
			}

			// Check for destination to determine if video or audio
			if match := destinationRegex.FindStringSubmatch(currentLine); match != nil {
				format := match[1]
				// Typically f251 is audio, higher numbers are video
				if format == "f251" {
					jobType = "audio"
				} else {
					jobType = "video"
				}
				continue
			}

			if match := itemRegex.FindStringSubmatch(currentLine); match != nil {
				isPlaylist = true
				currentItem = s.parseToInt(match[1])
				totalItems = s.parseToInt(match[2])
				overallProgress = (float64(currentItem-1) / float64(totalItems)) * 100

				update := ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          currentItem,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: 0,
				}

				if err := s.updateJobProgress(jobID, overallProgress); err != nil {
					log.WithError(err).Error("Failed to update job progress")
				}

				s.hub.broadcast <- update
				continue
			}

			if match := progressRegex.FindStringSubmatch(currentLine); match != nil {
				currentProgress, err := strconv.ParseFloat(match[1], 64)
				if err != nil {
					continue
				}

				if !isPlaylist {
					currentItem = 1
					totalItems = 1
					overallProgress = currentProgress
				} else {
					overallProgress = (float64(currentItem) / float64(totalItems)) * 100
				}

				update := ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          currentItem,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: currentProgress,
				}

				if err := s.updateJobProgress(jobID, overallProgress); err != nil {
					log.WithError(err).Error("Failed to update job progress")
				}

				s.hub.broadcast <- update
			}
		} else {
			line.WriteByte(char)
		}
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
