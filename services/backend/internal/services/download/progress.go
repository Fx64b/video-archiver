package download

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"video-archiver/internal/domain"

	log "github.com/sirupsen/logrus"
)

// Content type detection patterns
var (
	// Single video patterns
	singleVideoPattern = regexp.MustCompile(`\[youtube\] Extracting URL: https://www\.youtube\.com/watch\?v=`)

	// Playlist patterns
	playlistStartPattern = regexp.MustCompile(`\[download\] Downloading playlist: (.+)`)
	playlistItemPattern  = regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)

	// Channel patterns
	channelPattern = regexp.MustCompile(`\[youtube:tab\] @([^:]+):`)

	// Phase detection patterns
	metadataPattern       = regexp.MustCompile(`\[youtube\] [^:]+: Downloading (webpage|tv.*API JSON|tv client config)`)
	streamDownloadPattern = regexp.MustCompile(`\[download\] Destination: (.+) \[([A-Za-z0-9_-]+)\]\.f(\d+)\.(mp4|webm|m4a)`)
	mergePattern          = regexp.MustCompile(`\[Merger\] Merging formats into`)

	// Progress parsing pattern for yt-dlp template output
	// Format: [totalItems][playlist_index][video_id][title][format_id][format_note][vcodec][acodec]prog:[bytes/total][percent][speed][eta]
	progressTemplatePattern = regexp.MustCompile(`\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]\[([^\]]*)\]prog:\[([^\]]*)/([^\]]*)\]\[\s*([0-9.]+)%\]`)

	// Already downloaded pattern
	alreadyDownloadedPattern = regexp.MustCompile(`has already been downloaded|already been recorded in archive`)

	// Retry and error patterns
	retryPattern         = regexp.MustCompile(`\[download\] Got error: (.+)\. Retrying fragment \d+ \((\d+)/(\d+)\)`)
	fragmentSkipPattern  = regexp.MustCompile(`\[download\] fragment not found; Skipping fragment \d+`)
	httpErrorPattern     = regexp.MustCompile(`HTTP Error (\d+): (.+)`)
	fragmentRetryPattern = regexp.MustCompile(`Retrying fragment (\d+) \((\d+)/(\d+)\)`)
	errorPattern         = regexp.MustCompile(`^ERROR:`)
	fileEmptyPattern     = regexp.MustCompile(`The downloaded file is empty`)
)

const (
	maxRetryDuration = 60 * time.Second // Maximum time to spend retrying before giving up
)

// ProgressState represents the current state of a download job
type ProgressState struct {
	JobID             string
	JobType           string
	ContentType       string // "video", "playlist", "channel"
	Phase             string
	CurrentItem       int
	TotalItems        int
	ItemsCompleted    int
	CurrentProgress   float64 // 0-100 for current item
	OverallProgress   float64 // 0-100 for entire job
	LastUpdate        time.Time
	VideoStreams      map[string]int     // Track stream count per video ID (1st = video, 2nd = audio)
	RawProgress       map[string]float64 // Track raw yt-dlp progress per video ID
	CurrentStreamType map[string]string  // Track current stream type per video ID ("video" or "audio")
	VideoProgress     map[string]float64 // Track final video progress per video ID
	RetryCount        int                // Track consecutive retries
	MaxRetries        int                // Maximum retries from yt-dlp
	LastRetryTime     time.Time          // Track when last retry occurred
	RetryError        string             // Current retry error message
	IsStuck           bool               // Flag if download is stuck retrying
	HasError          bool               // Flag if download encountered an error
	Warnings          []string           // Collected warnings/errors during download
}

// ProgressTracker handles robust progress tracking
type ProgressTracker struct {
	state           *ProgressState
	service         *Service
	updateThrottle  time.Duration
	lastBroadcast   time.Time
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(service *Service, jobID, jobType string) *ProgressTracker {
	return &ProgressTracker{
		state: &ProgressState{
			JobID:          jobID,
			JobType:        jobType,
			ContentType:    "video", // default
			Phase:          domain.DownloadPhaseMetadata,
			CurrentItem:    1,
			TotalItems:     1,
			ItemsCompleted: 0,
			LastUpdate:     time.Now(),
			VideoStreams:   make(map[string]int),
			RawProgress:    make(map[string]float64),
			CurrentStreamType: make(map[string]string),
			VideoProgress: make(map[string]float64),
			Warnings:      make([]string, 0),
		},
		service:        service,
		updateThrottle: 100 * time.Millisecond, // 0.1 seconds max for more responsive updates
	}
}

// trackProgress is the main entry point for progress tracking
func (s *Service) trackProgress(pipe io.Reader, jobID string, jobType string) {
	tracker := NewProgressTracker(s, jobID, jobType)

	scanner := bufio.NewScanner(pipe)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		tracker.processLine(line)
	}

	// Send final completion update
	tracker.sendFinalUpdate()

	if err := scanner.Err(); err != nil {
		if !strings.Contains(err.Error(), "file already closed") && !strings.Contains(err.Error(), "broken pipe") {
			log.WithError(err).Error("Error reading progress output")
		}
	}
}

// processLine analyzes each line and updates state accordingly
func (pt *ProgressTracker) processLine(line string) {
	// Update last activity
	pt.state.LastUpdate = time.Now()

	// Debug logging
	if os.Getenv("DEBUG") == "true" {
		log.WithField("line", line).Debug("Processing line")
	}

	// Check if this line matches our patterns (debug logging only)
	if progressTemplatePattern.MatchString(line) {
		log.WithField("line", line).Debug("Line matches progress template pattern")
	} else if metadataPattern.MatchString(line) {
		log.WithField("line", line).Debug("Line matches metadata pattern")
	} else if streamDownloadPattern.MatchString(line) {
		log.WithField("line", line).Debug("Line matches stream download pattern")
	} else if mergePattern.MatchString(line) {
		log.WithField("line", line).Debug("Line matches merge pattern")
	}

	// 1. Content type detection (only set once)
	if pt.state.ContentType == "video" {
		pt.detectContentType(line)
	}

	// 2. Playlist item tracking
	pt.updatePlaylistProgress(line)

	// 3. Phase detection and progress simulation
	pt.detectPhase(line)

	// 4. No need to simulate progress - we get real progress from yt-dlp template

	// 5. Handle special cases
	pt.handleSpecialCases(line)

	// 6. Send throttled update (this handles the 0.25s throttling)
	pt.sendThrottledUpdate()
}


// detectContentType determines if this is a video, playlist, or channel download
func (pt *ProgressTracker) detectContentType(line string) {
	if playlistStartPattern.MatchString(line) {
		if channelPattern.MatchString(line) || strings.Contains(line, " - Videos") {
			pt.state.ContentType = "channel"
		} else {
			pt.state.ContentType = "playlist"
		}
		log.WithField("contentType", pt.state.ContentType).Debug("Content type detected")
	}
}

// updatePlaylistProgress tracks progress through playlist/channel items
func (pt *ProgressTracker) updatePlaylistProgress(line string) {
	if match := playlistItemPattern.FindStringSubmatch(line); match != nil {
		if current, err := strconv.Atoi(match[1]); err == nil {
			pt.state.CurrentItem = current
		}
		if total, err := strconv.Atoi(match[2]); err == nil {
			pt.state.TotalItems = total
		}

		// Reset current item progress when starting new item
		pt.state.CurrentProgress = 0
		pt.state.Phase = domain.DownloadPhaseMetadata
		// Clear stream tracking for new item
		pt.state.VideoStreams = make(map[string]int)
		pt.state.RawProgress = make(map[string]float64)
		pt.state.CurrentStreamType = make(map[string]string)
		pt.state.VideoProgress = make(map[string]float64)

		log.WithFields(log.Fields{
			"currentItem": pt.state.CurrentItem,
			"totalItems":  pt.state.TotalItems,
		}).Debug("Playlist item progress updated")
	}
}

// detectPhase determines the current download phase using alternating video/audio logic
func (pt *ProgressTracker) detectPhase(line string) {
	previousPhase := pt.state.Phase

	if metadataPattern.MatchString(line) {
		// Only set metadata phase if we're not already past it
		if pt.state.Phase == "" || pt.state.Phase == domain.DownloadPhaseMetadata {
			pt.state.Phase = domain.DownloadPhaseMetadata
			pt.state.CurrentProgress = 1 // Start with 1% for metadata
		}
	} else if progressMatch := progressTemplatePattern.FindStringSubmatch(line); progressMatch != nil {
		// Parse actual progress data from yt-dlp template
		// Format: [totalItems][playlist_index][video_id][title][format_id][format_note][vcodec][acodec]prog:[bytes/total][percent][speed][eta]
		videoID := progressMatch[3]     // video_id
		formatID := progressMatch[5]    // format_id
		formatNote := progressMatch[6]  // format_note
		vcodec := progressMatch[7]      // vcodec
		acodec := progressMatch[8]      // acodec
		percentStr := progressMatch[11] // percent

		if videoID != "NA" && videoID != "" {
			// Initialize maps if needed
			if pt.state.VideoStreams == nil {
				pt.state.VideoStreams = make(map[string]int)
			}
			if pt.state.RawProgress == nil {
				pt.state.RawProgress = make(map[string]float64)
			}
			if pt.state.CurrentStreamType == nil {
				pt.state.CurrentStreamType = make(map[string]string)
			}
			if pt.state.VideoProgress == nil {
				pt.state.VideoProgress = make(map[string]float64)
			}

			// Determine stream type based on codec information
			// Video stream: has video codec (vcodec != "none") and no audio codec (acodec == "none")
			// Audio stream: has audio codec (acodec != "none") and no video codec (vcodec == "none")
			var streamType string
			if vcodec != "none" && acodec == "none" {
				streamType = "video"
			} else if acodec != "none" && vcodec == "none" {
				streamType = "audio"
			} else {
				// Fallback: combined format or uncertain - treat as video
				streamType = "video"
			}

			// Parse the actual progress percentage
			if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
				// Check if stream type changed (e.g., from video to audio)
				previousStreamType := pt.state.CurrentStreamType[videoID]
				if previousStreamType != streamType && previousStreamType != "" {
					log.WithFields(log.Fields{
						"videoID":            videoID,
						"previousStreamType": previousStreamType,
						"newStreamType":      streamType,
						"formatID":           formatID,
						"formatNote":         formatNote,
						"vcodec":             vcodec,
						"acodec":             acodec,
					}).Debug("Stream type transition detected")
				}

				// Update current stream type
				pt.state.CurrentStreamType[videoID] = streamType

				// Initialize if this is the first time we see this video
				if pt.state.VideoStreams[videoID] == 0 {
					pt.state.VideoStreams[videoID] = 1
					log.WithFields(log.Fields{
						"videoID":    videoID,
						"streamType": streamType,
						"formatID":   formatID,
						"formatNote": formatNote,
						"vcodec":     vcodec,
						"acodec":     acodec,
					}).Debug("Video tracking initialized")
				}

				// Update raw progress tracking
				pt.state.RawProgress[videoID] = percent

				// Determine phase and progress based on current stream type
				if streamType == "audio" {
					pt.state.Phase = domain.DownloadPhaseAudio
					// Use saved video progress as base (default to 80 if not set)
					videoBaseProgress := pt.state.VideoProgress[videoID]
					if videoBaseProgress == 0 {
						videoBaseProgress = 80 // Default if video progress wasn't captured
					}
					pt.state.CurrentProgress = videoBaseProgress + (percent * 0.20) // Map audio to continue from video progress
				} else {
					pt.state.Phase = domain.DownloadPhaseVideo
					mappedVideoProgress := percent * 0.80 // Map video to 0-80%
					pt.state.CurrentProgress = mappedVideoProgress
					// Save video progress for when audio starts
					pt.state.VideoProgress[videoID] = mappedVideoProgress
				}

				// Debug logging for all progress updates
				log.WithFields(log.Fields{
					"videoID":        videoID,
					"formatID":       formatID,
					"formatNote":     formatNote,
					"vcodec":         vcodec,
					"acodec":         acodec,
					"rawPercent":     percent,
					"mappedProgress": pt.state.CurrentProgress,
					"streamType":     streamType,
					"phase":          pt.state.Phase,
					"videoProgress":  pt.state.VideoProgress[videoID],
				}).Debug("Progress update")
			}
		}
	} else if streamMatch := streamDownloadPattern.FindStringSubmatch(line); streamMatch != nil {
		// Handle stream destination announcements to determine stream type
		videoID := streamMatch[2]
		formatCode := streamMatch[3] // Extract format code (e.g., "401", "251")
		extension := streamMatch[4]   // Extract extension (e.g., "mp4", "webm", "m4a")

		// Initialize maps if needed
		if pt.state.CurrentStreamType == nil {
			pt.state.CurrentStreamType = make(map[string]string)
		}

		// Determine stream type based on format code and extension
		var streamType string
		if formatCodeNum, err := strconv.Atoi(formatCode); err == nil {
			// Video streams typically have format codes 300+ and use mp4/webm
			// Audio streams typically have format codes 200-299 and use webm/m4a
			if formatCodeNum >= 300 || extension == "mp4" {
				streamType = "video"
			} else if formatCodeNum < 300 || extension == "webm" || extension == "m4a" {
				streamType = "audio"
			} else {
				streamType = "video" // Default fallback
			}
		} else {
			// Fallback based on extension only
			if extension == "mp4" {
				streamType = "video"
			} else {
				streamType = "audio"
			}
		}

		// Check if this is a stream type change
		previousStreamType := pt.state.CurrentStreamType[videoID]
		pt.state.CurrentStreamType[videoID] = streamType

		// If transitioning from video to audio, update phase immediately
		if previousStreamType == "video" && streamType == "audio" {
			pt.state.Phase = domain.DownloadPhaseAudio
			log.WithFields(log.Fields{
				"videoID":            videoID,
				"previousStreamType": previousStreamType,
				"newStreamType":      streamType,
				"transition":         "video_to_audio",
			}).Debug("Stream type transition detected - switching to audio phase")
		} else if streamType == "video" {
			pt.state.Phase = domain.DownloadPhaseVideo
		}

		log.WithFields(log.Fields{
			"videoID":    videoID,
			"formatCode": formatCode,
			"extension":  extension,
			"streamType": streamType,
			"phase":      pt.state.Phase,
			"line":       line,
		}).Debug("Stream destination detected - stream type determined")

	} else if mergePattern.MatchString(line) {
		pt.state.Phase = domain.DownloadPhaseMerging
		pt.state.CurrentProgress = 95 // Merging phase
		log.WithField("line", line).Debug("Merge phase detected")
	}

	// Check for item completion patterns
	if strings.Contains(line, "Deleting original file") || strings.Contains(line, "Finished downloading") {
		pt.completeCurrentItem()
	}

	// Phase transition logging
	if previousPhase != pt.state.Phase {
		log.WithFields(log.Fields{
			"previousPhase": previousPhase,
			"newPhase":     pt.state.Phase,
			"line":         line,
		}).Debug("Phase transition detected")
	}
}


// handleSpecialCases deals with edge cases like already downloaded files and retries
func (pt *ProgressTracker) handleSpecialCases(line string) {
	// Check for error messages
	if errorPattern.MatchString(line) || fileEmptyPattern.MatchString(line) {
		pt.state.HasError = true

		// Extract the error message (remove "ERROR: " prefix if present)
		warningMsg := line
		if strings.HasPrefix(line, "ERROR: ") {
			warningMsg = strings.TrimPrefix(line, "ERROR: ")
		}

		// Add to warnings list if not already present
		isDuplicate := false
		for _, w := range pt.state.Warnings {
			if w == warningMsg {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			pt.state.Warnings = append(pt.state.Warnings, warningMsg)
		}

		log.WithFields(log.Fields{
			"jobID": pt.state.JobID,
			"error": line,
		}).Error("Download error detected")
		return
	}

	if alreadyDownloadedPattern.MatchString(line) {
		pt.state.CurrentProgress = 100
		pt.state.Phase = domain.DownloadPhaseComplete

		// Mark this item as completed
		pt.completeCurrentItem()

		log.WithField("line", line).Debug("Already downloaded detected")
		return
	}

	// Handle retry patterns
	if retryMatch := retryPattern.FindStringSubmatch(line); retryMatch != nil {
		errorMsg := retryMatch[1]
		currentRetryStr := retryMatch[2]
		maxRetriesStr := retryMatch[3]

		// Parse retry counts
		currentRetry, _ := strconv.Atoi(currentRetryStr)
		maxRetries, _ := strconv.Atoi(maxRetriesStr)

		// Extract fragment number if present
		var fragmentNum string
		if fragmentMatch := fragmentRetryPattern.FindStringSubmatch(line); fragmentMatch != nil {
			fragmentNum = fragmentMatch[1]
		}

		// Detect HTTP errors
		var httpError string
		if httpMatch := httpErrorPattern.FindStringSubmatch(errorMsg); httpMatch != nil {
			httpError = fmt.Sprintf("HTTP %s: %s", httpMatch[1], httpMatch[2])
		} else {
			httpError = errorMsg
		}

		// Update retry state
		pt.state.RetryCount = currentRetry
		pt.state.MaxRetries = maxRetries
		pt.state.RetryError = httpError

		// Check if we just started retrying
		if pt.state.LastRetryTime.IsZero() {
			pt.state.LastRetryTime = time.Now()
		}

		// Check if we've been retrying for too long
		retryDuration := time.Since(pt.state.LastRetryTime)
		if retryDuration > maxRetryDuration {
			pt.state.IsStuck = true
			log.WithFields(log.Fields{
				"jobID":         pt.state.JobID,
				"retryDuration": retryDuration,
				"retryCount":    pt.state.RetryCount,
				"error":         httpError,
			}).Error("Download stuck - exceeded maximum retry duration")
			return
		}

		if fragmentNum != "" {
			log.WithFields(log.Fields{
				"jobID":      pt.state.JobID,
				"fragment":   fragmentNum,
				"retry":      currentRetry,
				"maxRetries": maxRetries,
				"error":      httpError,
				"duration":   retryDuration.Round(time.Second),
			}).Warn("Download retry in progress")
		} else {
			log.WithFields(log.Fields{
				"jobID":      pt.state.JobID,
				"retry":      currentRetry,
				"maxRetries": maxRetries,
				"error":      httpError,
				"duration":   retryDuration.Round(time.Second),
			}).Warn("Download retry in progress")
		}
		return
	}

	// Handle fragment skip (usually follows retries)
	if fragmentSkipPattern.MatchString(line) {
		log.WithField("jobID", pt.state.JobID).Debug("Fragment skipped after retries")
		// Reset retry tracking since we're moving on
		pt.state.RetryCount = 0
		pt.state.MaxRetries = 0
		pt.state.RetryError = ""
		pt.state.LastRetryTime = time.Time{}
		return
	}

	// Reset retry tracking on successful progress
	if pt.state.RetryCount > 0 && !retryPattern.MatchString(line) && !fragmentSkipPattern.MatchString(line) {
		if pt.state.RetryCount > 5 {
			log.WithFields(log.Fields{
				"jobID":      pt.state.JobID,
				"retryCount": pt.state.RetryCount,
			}).Info("Download recovered after retries")
		}
		pt.state.RetryCount = 0
		pt.state.MaxRetries = 0
		pt.state.RetryError = ""
		pt.state.LastRetryTime = time.Time{}
	}
}

// completeCurrentItem marks the current item as completed
func (pt *ProgressTracker) completeCurrentItem() {
	// Only increment completed count if this item isn't already counted
	if pt.state.ItemsCompleted < pt.state.CurrentItem {
		pt.state.ItemsCompleted = pt.state.CurrentItem
	}
	pt.state.CurrentProgress = 100
	pt.state.Phase = domain.DownloadPhaseComplete

	// If this was the last item, mark overall job as complete
	if pt.state.CurrentItem >= pt.state.TotalItems {
		pt.state.OverallProgress = 100
	}

	log.WithFields(log.Fields{
		"currentItem":    pt.state.CurrentItem,
		"itemsCompleted": pt.state.ItemsCompleted,
		"totalItems":     pt.state.TotalItems,
	}).Debug("Item completed")
}

// calculateProgress computes the overall job progress
func (pt *ProgressTracker) calculateProgress() {
	if pt.state.TotalItems == 1 {
		// Single video: use current progress directly (already mapped to visual representation)
		pt.state.OverallProgress = pt.state.CurrentProgress
	} else {
		// Playlist/channel: (completed items * 100 + current item progress) / total items
		completedProgress := float64(pt.state.ItemsCompleted) * 100
		currentItemProgress := pt.state.CurrentProgress

		pt.state.OverallProgress = (completedProgress + currentItemProgress) / float64(pt.state.TotalItems)
	}

	// Ensure progress stays within bounds
	if pt.state.OverallProgress > 100 {
		pt.state.OverallProgress = 100
	}
	if pt.state.OverallProgress < 0 {
		pt.state.OverallProgress = 0
	}
}

// sendThrottledUpdate sends progress updates with throttling
func (pt *ProgressTracker) sendThrottledUpdate() {
	now := time.Now()

	// Only send update if enough time has passed
	if now.Sub(pt.lastBroadcast) < pt.updateThrottle {
		return
	}

	pt.calculateProgress()
	pt.broadcastUpdate()
	pt.updateDatabase()

	pt.lastBroadcast = now
}

// sendFinalUpdate sends the final completion update
func (pt *ProgressTracker) sendFinalUpdate() {
	// Don't mark as complete if there was an error
	if !pt.state.HasError {
		pt.state.OverallProgress = 100
		pt.state.Phase = domain.DownloadPhaseComplete

		pt.broadcastUpdate()
		pt.updateDatabase()

		log.WithField("jobID", pt.state.JobID).Debug("Final progress update sent")
	} else {
		log.WithField("jobID", pt.state.JobID).Warn("Skipping final update due to error")
	}
}

// broadcastUpdate sends the progress update via WebSocket
func (pt *ProgressTracker) broadcastUpdate() {
	update := domain.ProgressUpdate{
		JobID:                pt.state.JobID,
		JobType:              pt.state.JobType,
		CurrentItem:          pt.state.CurrentItem,
		TotalItems:           pt.state.TotalItems,
		Progress:             pt.state.OverallProgress,
		CurrentVideoProgress: pt.state.CurrentProgress,
		DownloadPhase:        pt.state.Phase,
		IsRetrying:           pt.state.RetryCount > 0,
		RetryCount:           pt.state.RetryCount,
		MaxRetries:           pt.state.MaxRetries,
		RetryError:           pt.state.RetryError,
		Warnings:             pt.state.Warnings,
	}

	pt.service.hub.broadcast <- update

	if os.Getenv("DEBUG") == "true" {
		log.WithFields(log.Fields{
			"jobID":              update.JobID,
			"currentItem":        update.CurrentItem,
			"totalItems":         update.TotalItems,
			"overallProgress":    update.Progress,
			"currentProgress":    update.CurrentVideoProgress,
			"phase":              update.DownloadPhase,
			"contentType":        pt.state.ContentType,
			"isRetrying":         update.IsRetrying,
			"retryCount":         update.RetryCount,
			"warningCount":       len(update.Warnings),
		}).Debug("Progress update broadcast")
	}
}

// updateDatabase updates the job progress in the database
func (pt *ProgressTracker) updateDatabase() {
	// If download is stuck, mark job as failed
	if pt.state.IsStuck {
		job, err := pt.service.jobs.GetByID(pt.state.JobID)
		if err != nil {
			log.WithError(err).Error("Failed to get job for stuck download")
			return
		}
		job.Status = domain.JobStatusError
		job.Warnings = pt.state.Warnings
		if err := pt.service.jobs.Update(job); err != nil {
			log.WithError(err).Error("Failed to mark stuck job as failed")
		}
		return
	}

	if err := pt.updateJobProgress(pt.state.JobID, pt.state.OverallProgress); err != nil {
		log.WithError(err).Error("Failed to update job progress in database")
	}
}

// updateJobProgress updates job progress in database
func (pt *ProgressTracker) updateJobProgress(jobID string, progress float64) error {
	job, err := pt.service.jobs.GetByID(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	job.Progress = progress
	job.Warnings = pt.state.Warnings
	return pt.service.jobs.Update(job)
}

// Legacy functions for metadata handling (unchanged)
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

