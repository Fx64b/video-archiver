package download

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"video-archiver/internal/domain"
)

func (s *Service) trackProgress(pipe io.Reader, jobID string, jobType string) {
	reader := bufio.NewReader(pipe)
	var line bytes.Buffer

	itemRegex := regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)
	progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)% of.* \s+\d+\.?\d*\w+`)
	destinationRegex := regexp.MustCompile(`\[download\] Destination: .+\.(f\d+)\.(mp4|webm)`)
	alreadyDownloadedRegex := regexp.MustCompile(`\[download\].*has already been downloaded`)
	videoCompleteRegex := regexp.MustCompile(`\[download\] .*: .* has already been recorded in archive`)
	mergerCompleteRegex := regexp.MustCompile(`\[Merger\] Merging formats into`)

	playlistStartRegex := regexp.MustCompile(`\[download\] Downloading playlist:`)
	playlistEndRegex := regexp.MustCompile(`\[download\] Finished downloading playlist:`)
	playlistMetadataRegex := regexp.MustCompile(`\[info\] Writing playlist metadata as JSON to: (.+\.info\.json)`)
	playlistProgressRegex := regexp.MustCompile(`\[youtube:tab\] Playlist .+: Downloading (\d+) items of (\d+)`)

	var totalItems, currentItem int
	var overallProgress float64

	// Flags
	isPlaylist := false
	isProcessingAudio := false
	currentVideoComplete := false
	mergerComplete := false
	alreadyDownloaded := false

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

			if match := playlistMetadataRegex.FindStringSubmatch(currentLine); match != nil {
				metadataPath := strings.TrimSpace(match[1])
				s.processPlaylistMetadata(jobID, metadataPath)
				continue
			}

			if match := playlistProgressRegex.FindStringSubmatch(currentLine); match != nil {
				totalItems := s.parseToInt(match[2])
				update := domain.ProgressUpdate{
					JobID:      jobID,
					JobType:    string(domain.JobTypeMetadata),
					TotalItems: totalItems,
					Progress:   0,
				}
				s.hub.broadcast <- update
				continue
			}

			if mergerCompleteRegex.MatchString(currentLine) {
				mergerComplete = true
				if !isPlaylist {
					overallProgress = 100
					update := domain.ProgressUpdate{
						JobID:                jobID,
						JobType:              jobType,
						CurrentItem:          currentItem,
						TotalItems:           totalItems,
						Progress:             overallProgress,
						CurrentVideoProgress: 100,
					}
					s.hub.broadcast <- update
					if err := s.updateJobProgress(jobID, overallProgress); err != nil {
						log.WithError(err).Error("Failed to update job progress")
					}
				}
				continue
			}

			// Handle video completion detection
			if videoCompleteRegex.MatchString(currentLine) {
				if isPlaylist && !currentVideoComplete {
					currentItem++
					currentVideoComplete = true
					overallProgress = (float64(currentItem) / float64(totalItems)) * 100

					update := domain.ProgressUpdate{
						JobID:                jobID,
						JobType:              jobType,
						CurrentItem:          currentItem,
						TotalItems:           totalItems,
						Progress:             overallProgress,
						CurrentVideoProgress: 100,
					}
					s.hub.broadcast <- update
					if err := s.updateJobProgress(jobID, overallProgress); err != nil {
						log.WithError(err).Error("Failed to update job progress")
					}
				}
				continue
			}

			if playlistStartRegex.MatchString(currentLine) {
				isPlaylist = true
				currentVideoComplete = false
				continue
			}

			currentVideoProgress := 100.0

			if alreadyDownloaded {
				currentVideoProgress = 101
			}

			if playlistEndRegex.MatchString(currentLine) {
				overallProgress = 100
				update := domain.ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          totalItems,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: currentVideoProgress,
				}
				s.hub.broadcast <- update
				if err := s.updateJobProgress(jobID, overallProgress); err != nil {
					log.WithError(err).Error("Failed to update job progress")
				}
				continue
			}

			if match := alreadyDownloadedRegex.FindStringSubmatch(currentLine); match != nil {
				if !isProcessingAudio {
					if isPlaylist {
						currentItem++
						currentVideoComplete = true
						overallProgress = (float64(currentItem) / float64(totalItems)) * 100
					} else {
						overallProgress = 100
					}
				}

				alreadyDownloaded = true

				update := domain.ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          currentItem,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: 101,
				}

				if !isProcessingAudio {
					if err := s.updateJobProgress(jobID, overallProgress); err != nil {
						log.WithError(err).Error("Failed to update job progress")
					}
				}

				s.hub.broadcast <- update
				continue
			}

			if match := destinationRegex.FindStringSubmatch(currentLine); match != nil {
				format := match[1]
				isProcessingAudio = format == "f251"
				currentVideoComplete = false
				jobType = string(domain.JobTypeAudio)
				if !isProcessingAudio {
					jobType = string(domain.JobTypeVideo)
				}
				continue
			}

			if match := itemRegex.FindStringSubmatch(currentLine); match != nil {
				if !isProcessingAudio && !currentVideoComplete {
					currentItem = s.parseToInt(match[1])
					totalItems = s.parseToInt(match[2])
					overallProgress = ((float64(currentItem) - 1) / float64(totalItems)) * 100
				}

				update := domain.ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          currentItem,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: 0,
				}

				if !isProcessingAudio {
					if err := s.updateJobProgress(jobID, overallProgress); err != nil {
						log.WithError(err).Error("Failed to update job progress")
					}
				}

				s.hub.broadcast <- update
				continue
			}

			if match := progressRegex.FindStringSubmatch(currentLine); match != nil {
				currentProgress, err := strconv.ParseFloat(match[1], 64)
				if err != nil {
					continue
				}

				if !isProcessingAudio && !currentVideoComplete {
					if !isPlaylist {
						currentItem = 1
						totalItems = 1
						// For single videos, cap progress at 99% until merger completes
						if !mergerComplete {
							overallProgress = math.Min(currentProgress, 99)
						}
					} else {
						overallProgress = ((float64(currentItem) - 1) / float64(totalItems)) * 100
						overallProgress += currentProgress / float64(totalItems)
					}

					if currentProgress >= 100 {
						currentVideoComplete = true
						if isPlaylist {
							currentItem++
						}
					}
				}

				update := domain.ProgressUpdate{
					JobID:                jobID,
					JobType:              jobType,
					CurrentItem:          currentItem,
					TotalItems:           totalItems,
					Progress:             overallProgress,
					CurrentVideoProgress: currentProgress,
				}

				if !isProcessingAudio {
					if err := s.updateJobProgress(jobID, overallProgress); err != nil {
						log.WithError(err).Error("Failed to update job progress")
					}
				}

				s.hub.broadcast <- update
			}

			// TODO: improve
			if strings.Contains(currentLine, "Writing video metadata as JSON to:") && !isPlaylist {
				metadataPath := strings.TrimPrefix(currentLine, "[info] Writing video metadata as JSON to: ")
				s.setMetadataPath(jobID, strings.TrimSpace(metadataPath))
			}
		} else {
			line.WriteByte(char)
		}
	}
}

func (s *Service) processPlaylistMetadata(jobID string, metadataPath string) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		log.WithError(err).Error("Failed to read playlist metadata file")
		return
	}

	var metadata domain.PlaylistMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		log.WithError(err).Error("Failed to parse playlist metadata")
		return
	}

	update := domain.MetadataUpdate{
		JobID:    jobID,
		Metadata: &metadata,
	}
	s.hub.broadcast <- update

	if err := s.jobs.StoreMetadata(jobID, &metadata); err != nil {
		log.WithError(err).Error("Failed to store playlist metadata")
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
