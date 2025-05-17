package download

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"video-archiver/internal/domain"
)

// progressPattern is the regex to parse the structured output from yt-dlp
// Format: [playlist_count][playlist_index][id][title]prog:[downloaded_bytes/total_bytes][percent][speed][eta]
var progressPattern = regexp.MustCompile(`\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]prog:\[([^\]]*)/([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]`)

// metadataPattern captures metadata file paths
var metadataPattern = regexp.MustCompile(`Writing (video|playlist|channel) metadata as JSON to: (.+\.info\.json)`)

// trackProgress parses and tracks download progress from yt-dlp output
func (s *Service) trackProgress(pipe io.Reader, jobID string, jobType string) {
	scanner := bufio.NewScanner(pipe)

	// Track progress state
	var currentItem, totalItems int
	var overallProgress float64
	var currentPhase string = domain.DownloadPhaseVideo // Default to video
	var videoProgress, audioProgress float64
	var videoCompleted, audioCompleted bool
	var downloadingAudio bool

	// Configure buffer size for large lines (some progress lines can be quite long)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Log line for debugging
		if os.Getenv("DEBUG") == "true" {
			log.Debug(line)
		}

		// Check for download phase transitions
		if strings.Contains(line, "Downloading video") ||
			strings.Contains(line, "Destination:") && strings.Contains(line, ".mp4") && !strings.Contains(line, ".f251") {
			currentPhase = domain.DownloadPhaseVideo
			downloadingAudio = false
		} else if strings.Contains(line, "Downloading audio") ||
			strings.Contains(line, "Destination:") && (strings.Contains(line, ".webm") || strings.Contains(line, ".f251")) {
			currentPhase = domain.DownloadPhaseAudio
			downloadingAudio = true
			// When switching to audio phase, video is complete
			if !videoCompleted {
				videoCompleted = true
				videoProgress = 100.0
			}
		} else if strings.Contains(line, "Merging formats") {
			currentPhase = domain.DownloadPhaseMerging
			videoCompleted = true
			audioCompleted = true
		}

		// Check for metadata file paths
		if metaMatch := metadataPattern.FindStringSubmatch(line); metaMatch != nil {
			metadataType := metaMatch[1]
			metadataPath := metaMatch[2]

			// Set the current phase to metadata
			currentPhase = domain.DownloadPhaseMetadata

			if metadataType == "playlist" || metadataType == "channel" {
				s.processPlaylistMetadata(jobID, metadataPath)
			} else {
				s.setMetadataPath(jobID, metadataPath)
			}
			continue
		}

		// Check for progress information
		if match := progressPattern.FindStringSubmatch(line); match != nil {
			// Extract values from matching groups
			playlistCount := match[1]
			playlistIndex := match[2]
			videoID := match[3]
			videoTitle := match[4]
			percent := match[7]

			// Debug logging for raw matched values
			if os.Getenv("DEBUG") == "true" {
				log.WithFields(log.Fields{
					"playlistCount": playlistCount,
					"playlistIndex": playlistIndex,
					"videoID":       videoID,
					"videoTitle":    videoTitle,
					"raw_percent":   percent,
					"phase":         currentPhase,
				}).Debug("Matched progress values")
			}

			// Process playlist info if available
			isPlaylist := playlistCount != "NA" && playlistCount != ""

			if isPlaylist {
				newTotalItems, _ := strconv.Atoi(playlistCount)
				newCurrentItem, _ := strconv.Atoi(playlistIndex)

				if newTotalItems > 0 {
					totalItems = newTotalItems
				}

				if newCurrentItem > 0 {
					// Only update current item if it's a new item
					if newCurrentItem != currentItem {
						currentItem = newCurrentItem
						videoCompleted = false
						audioCompleted = false
						videoProgress = 0
						audioProgress = 0
					}
				}
			} else {
				// Single video case
				if totalItems == 0 {
					totalItems = 1
				}
				if currentItem == 0 {
					currentItem = 1
				}
			}

			// Parse percentage to calculate progress
			percentValue := 0.0
			if percent != "NA" && percent != "" {
				// Clean up the percentage string and handle possible formatting issues
				percentStr := strings.TrimSpace(percent)
				percentStr = strings.TrimSuffix(percentStr, "%")

				// Try to parse the percentage
				var err error
				percentValue, err = strconv.ParseFloat(percentStr, 64)
				if err != nil {
					log.WithError(err).WithField("raw_percent", percent).
						Warn("Failed to parse percentage value, trying alternative method")

					// Try again with stronger cleaning - remove everything except digits and decimal point
					reg := regexp.MustCompile(`[^0-9.]`)
					cleanPercentStr := reg.ReplaceAllString(percentStr, "")
					percentValue, _ = strconv.ParseFloat(cleanPercentStr, 64)
				}
			}

			// Update progress based on current phase
			if downloadingAudio {
				audioProgress = percentValue
			} else {
				videoProgress = percentValue
			}

			// Calculate overall progress
			if isPlaylist {
				// For playlists: (completed items + current item progress)
				overallProgress = (float64(currentItem-1) / float64(totalItems)) * 100

				// Current item progress is a combination of video and audio progress
				if !videoCompleted && !audioCompleted {
					itemProgress := 0.0
					if downloadingAudio {
						// Video is done, audio is in progress
						itemProgress = 90.0 + (audioProgress * 0.1) // Audio is worth 10% of progress
					} else {
						// Video is in progress
						itemProgress = videoProgress * 0.9 // Video is worth 90% of progress
					}
					overallProgress += (itemProgress / float64(totalItems))
				}
			} else {
				// For single videos: combine video and audio progress
				if downloadingAudio {
					// Video is done, audio is in progress
					overallProgress = 90.0 + (audioProgress * 0.1) // Audio is worth 10% of progress
				} else {
					// Video is in progress
					overallProgress = videoProgress * 0.9 // Video is worth 90% of progress
				}
			}

			// Handle completed phases
			if percentValue >= 100 {
				if downloadingAudio {
					audioCompleted = true
				} else {
					videoCompleted = true
				}

				// If both phases are complete, consider the item done
				if videoCompleted && audioCompleted && isPlaylist && currentItem < totalItems {
					currentItem++
					videoCompleted = false
					audioCompleted = false
					videoProgress = 0
					audioProgress = 0
				}
			}

			// Special case for already downloaded videos
			if strings.Contains(line, "already downloaded") || strings.Contains(line, "has already been recorded in archive") {
				if isPlaylist {
					videoCompleted = true
					audioCompleted = true
					// If already done, just count it as complete and move to next
					overallProgress = (float64(currentItem) / float64(totalItems)) * 100
					if currentItem < totalItems {
						currentItem++
					}
				} else {
					// Single video already downloaded
					overallProgress = 100
					videoCompleted = true
					audioCompleted = true
				}
				percentValue = 100
			}

			// If all items completed, set progress to 100%
			if isPlaylist && currentItem > totalItems {
				currentItem = totalItems
				overallProgress = 100
			}

			// For merging phase, set progress to 99% to indicate almost done
			if currentPhase == domain.DownloadPhaseMerging {
				overallProgress = 99
				percentValue = 99
			}

			// Create progress update
			update := domain.ProgressUpdate{
				JobID:                jobID,
				JobType:              jobType,
				CurrentItem:          currentItem,
				TotalItems:           totalItems,
				Progress:             overallProgress,
				CurrentVideoProgress: percentValue,
				DownloadPhase:        currentPhase,
			}

			// Send progress update
			s.hub.broadcast <- update

			// Update job progress in database
			if err := s.updateJobProgress(jobID, overallProgress); err != nil {
				log.WithError(err).Error("Failed to update job progress")
			}

			// Debug logging
			if os.Getenv("DEBUG") == "true" {
				log.WithFields(log.Fields{
					"jobID":                jobID,
					"currentItem":          currentItem,
					"totalItems":           totalItems,
					"progress":             overallProgress,
					"currentVideoProgress": percentValue,
					"downloadPhase":        currentPhase,
					"videoProgress":        videoProgress,
					"audioProgress":        audioProgress,
					"videoCompleted":       videoCompleted,
					"audioCompleted":       audioCompleted,
					"videoTitle":           videoTitle,
					"videoID":              videoID,
				}).Debug("Progress update")
			}
		}
	}

	// Final progress update to ensure 100% is reported
	finalUpdate := domain.ProgressUpdate{
		JobID:                jobID,
		JobType:              jobType,
		CurrentItem:          totalItems,
		TotalItems:           totalItems,
		Progress:             100,
		CurrentVideoProgress: 100,
		DownloadPhase:        domain.DownloadPhaseMerging,
	}
	s.hub.broadcast <- finalUpdate

	if err := s.updateJobProgress(jobID, 100); err != nil {
		log.WithError(err).Error("Failed to update final job progress")
	}

	if err := scanner.Err(); err != nil {
		log.WithError(err).Error("Error reading progress output")
	}
}

// The other helper methods remain the same
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
