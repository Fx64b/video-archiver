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
	playlistItemPattern = regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)

	// Channel patterns
	channelPattern = regexp.MustCompile(`\[youtube:tab\] @([^:]+):`)

	// Phase detection patterns
	metadataPattern = regexp.MustCompile(`\[youtube\] [^:]+: Downloading (webpage|tv.*API JSON|tv client config)`)
	videoStreamPattern = regexp.MustCompile(`\[download\] Destination: .+\.f\d+\.mp4`)
	audioStreamPattern = regexp.MustCompile(`\[download\] Destination: .+\.f\d+\.webm`)
	mergePattern = regexp.MustCompile(`\[Merger\] Merging formats into`)

	// Already downloaded pattern
	alreadyDownloadedPattern = regexp.MustCompile(`has already been downloaded|already been recorded in archive`)
)

// ProgressState represents the current state of a download job
type ProgressState struct {
	JobID           string
	JobType         string
	ContentType     string // "video", "playlist", "channel"
	Phase           string
	CurrentItem     int
	TotalItems      int
	ItemsCompleted  int
	CurrentProgress float64 // 0-100 for current item
	OverallProgress float64 // 0-100 for entire job
	LastUpdate      time.Time
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
		},
		service:        service,
		updateThrottle: 250 * time.Millisecond, // 0.25 seconds max
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

	// 1. Content type detection (only set once)
	if pt.state.ContentType == "video" {
		pt.detectContentType(line)
	}

	// 2. Playlist item tracking
	pt.updatePlaylistProgress(line)

	// 3. Phase detection and progress simulation
	previousPhase := pt.state.Phase
	pt.detectPhase(line)

	// 4. Simulate progress within current phase (only if not transitioning)
	if previousPhase == pt.state.Phase {
		pt.simulateProgressInCurrentPhase()
	}

	// 5. Handle special cases
	pt.handleSpecialCases(line)

	// 6. Send throttled update (this handles the 0.25s throttling)
	pt.sendThrottledUpdate()
}

// simulateProgressInCurrentPhase adds gradual progress within the current phase
func (pt *ProgressTracker) simulateProgressInCurrentPhase() {
	switch pt.state.Phase {
	case domain.DownloadPhaseVideo:
		// Slowly increment video progress from 5% to 75%
		if pt.state.CurrentProgress < 75 {
			increment := 0.5 + (time.Since(pt.state.LastUpdate).Seconds() * 0.1)
			pt.state.CurrentProgress += increment
			if pt.state.CurrentProgress > 75 {
				pt.state.CurrentProgress = 75
			}
		}
	case domain.DownloadPhaseAudio:
		// Slowly increment audio progress from 80% to 95%
		if pt.state.CurrentProgress < 95 {
			increment := 0.3 + (time.Since(pt.state.LastUpdate).Seconds() * 0.05)
			pt.state.CurrentProgress += increment
			if pt.state.CurrentProgress > 95 {
				pt.state.CurrentProgress = 95
			}
		}
	case domain.DownloadPhaseMetadata:
		// Slowly increment metadata progress from 1% to 4%
		if pt.state.CurrentProgress < 4 {
			pt.state.CurrentProgress += 0.1
			if pt.state.CurrentProgress > 4 {
				pt.state.CurrentProgress = 4
			}
		}
	}
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

		log.WithFields(log.Fields{
			"currentItem": pt.state.CurrentItem,
			"totalItems":  pt.state.TotalItems,
		}).Debug("Playlist item progress updated")
	}
}

// detectPhase determines the current download phase and simulates progress
func (pt *ProgressTracker) detectPhase(line string) {
	previousPhase := pt.state.Phase

	if metadataPattern.MatchString(line) {
		if pt.state.Phase != domain.DownloadPhaseMetadata {
			pt.state.Phase = domain.DownloadPhaseMetadata
			pt.state.CurrentProgress = 1 // Start with 1% for metadata
		}
	} else if videoStreamPattern.MatchString(line) {
		// Only transition to video phase if we're not already in video or later phases
		if pt.state.Phase == domain.DownloadPhaseMetadata {
			pt.state.Phase = domain.DownloadPhaseVideo
			pt.state.CurrentProgress = 5 // Start video at 5%
		}
	} else if audioStreamPattern.MatchString(line) {
		// Only transition to audio phase if we're coming from video phase
		if pt.state.Phase == domain.DownloadPhaseVideo {
			pt.state.Phase = domain.DownloadPhaseAudio
			pt.state.CurrentProgress = 80 // Start audio at 80% for visual
		}
	} else if mergePattern.MatchString(line) {
		pt.state.Phase = domain.DownloadPhaseMerging
		pt.state.CurrentProgress = 95 // Merging phase
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


// handleSpecialCases deals with edge cases like already downloaded files
func (pt *ProgressTracker) handleSpecialCases(line string) {
	if alreadyDownloadedPattern.MatchString(line) {
		pt.state.CurrentProgress = 100
		pt.state.Phase = domain.DownloadPhaseComplete

		// Mark this item as completed
		pt.completeCurrentItem()

		log.WithField("line", line).Debug("Already downloaded detected")
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
	pt.state.OverallProgress = 100
	pt.state.Phase = domain.DownloadPhaseComplete

	pt.broadcastUpdate()
	pt.updateDatabase()

	log.WithField("jobID", pt.state.JobID).Debug("Final progress update sent")
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
		}).Debug("Progress update broadcast")
	}
}

// updateDatabase updates the job progress in the database
func (pt *ProgressTracker) updateDatabase() {
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

