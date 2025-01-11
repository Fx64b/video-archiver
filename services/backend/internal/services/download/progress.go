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
	"video-archiver/internal/domain"
)

func (s *Service) trackProgress(pipe io.Reader, jobID string) {
	reader := bufio.NewReader(pipe)
	var line bytes.Buffer

	itemRegex := regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)
	progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)% of.* \s+\d+\.?\d*\w+`)
	destinationRegex := regexp.MustCompile(`\[download\] Destination: .+\.(f\d+)\.(mp4|webm)`)
	metadataRegex := regexp.MustCompile(`\[info\] Writing video metadata as JSON`)
	alreadyDownloadedRegex := regexp.MustCompile(`\[download\].*has already been downloaded`)
	tabRegex := regexp.MustCompile(`\[youtube(?::tab)?\]`)

	var totalItems, currentItem int
	var overallProgress float64
	isPlaylist := false
	jobType := string(domain.JobTypeVideo)

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

			if metadataRegex.MatchString(currentLine) || tabRegex.MatchString(currentLine) {
				jobType = string(domain.JobTypeMetadata)
				update := domain.ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          currentItem,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: 0,
				}
				s.hub.broadcast <- update
				continue
			}

			if match := destinationRegex.FindStringSubmatch(currentLine); match != nil {
				format := match[1]
				// Typically f251 is audio, higher numbers are video
				if format == "f251" {
					jobType = string(domain.JobTypeAudio)
				} else {
					jobType = string(domain.JobTypeVideo)
				}
				continue
			}

			if match := itemRegex.FindStringSubmatch(currentLine); match != nil {
				isPlaylist = true
				currentItem = s.parseToInt(match[1])
				totalItems = s.parseToInt(match[2])

				if jobType != string(domain.JobTypeAudio) {
					overallProgress = (float64(currentItem-1) / float64(totalItems)) * 100
				}

				update := domain.ProgressUpdate{
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

			// Check if job has already been downloaded
			if match := alreadyDownloadedRegex.FindStringSubmatch(currentLine); match != nil {
				overallProgress = 100
				update := domain.ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          0,
					TotalItems:           0,
					Progress:             101,
					CurrentVideoProgress: 100,
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
					overallProgress = (float64(currentItem-1) / float64(totalItems)) * 100
					overallProgress += currentProgress / float64(totalItems)
				}

				update := domain.ProgressUpdate{
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

		if totalItems == currentItem {
			overallProgress = 100
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
