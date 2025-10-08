package download

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/metadata"
)

type Config struct {
	JobRepository      domain.JobRepository
	SettingsRepository domain.SettingsRepository
	DownloadPath       string
	Concurrency        int
	MaxQuality         int
}

type Service struct {
	config        *Config
	jobs          domain.JobRepository
	settings      domain.SettingsRepository
	queue         chan domain.Job
	wg            sync.WaitGroup
	hub           *WebSocketHub
	ctx           context.Context
	cancel        context.CancelFunc
	metadataPaths sync.Map
}

func NewService(config *Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	hub := NewWebSocketHub()

	return &Service{
		config:   config,
		jobs:     config.JobRepository,
		settings: config.SettingsRepository,
		queue:    make(chan domain.Job, 100),
		hub:      hub,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Service) Start() error {
	go s.hub.Run()

	for i := 0; i < s.config.Concurrency; i++ {
		s.wg.Add(1)
		go s.processJobs()
	}

	return nil
}

func (s *Service) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *Service) GetHub() *WebSocketHub {
	return s.hub
}

func (s *Service) Submit(job domain.Job) error {
	job.Status = domain.JobStatusPending
	job.Progress = 0

	if err := s.jobs.Create(&job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	s.queue <- job
	return nil
}

func (s *Service) GetRepository() domain.JobRepository {
	return s.jobs
}

func (s *Service) GetRecent(limit int) ([]*domain.Job, error) {
	return s.jobs.GetRecent(limit)
}

func (s *Service) processJobs() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case job := <-s.queue:
			if err := s.processJob(s.ctx, job); err != nil {
				log.WithError(err).
					WithField("jobID", job.ID).
					Error("Failed to process job")

				job.Status = domain.JobStatusError
				if err := s.jobs.Update(&job); err != nil {
					log.WithError(err).Error("Failed to update job status")
				}

				// Broadcast error status via WebSocket
				errorUpdate := domain.ProgressUpdate{
					JobID:    job.ID,
					JobType:  string(domain.JobTypeVideo),
					Status:   domain.JobStatusError,
					Progress: job.Progress,
				}
				s.hub.broadcast <- errorUpdate
			}
		}
	}
}

func (s *Service) setMetadataPath(jobID string, path string) {
	s.metadataPaths.Store(jobID, path)
}

func (s *Service) getMetadataPath(jobID string) (string, bool) {
	path, ok := s.metadataPaths.Load(jobID)
	if !ok {
		return "", false
	}
	return path.(string), true
}

func (s *Service) processJob(ctx context.Context, job domain.Job) error {
	job.Status = domain.JobStatusPending
	job.Progress = 0
	_ = s.jobs.Create(&job)

	job.Status = domain.JobStatusInProgress
	if err := s.jobs.Update(&job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	basePath := fmt.Sprintf("%s/%%(uploader)s/%%(title)s", s.config.DownloadPath)

	log.Infof("Download Path: %s", basePath)

	// TODO: start metadata extraction in a separate goroutine and then do the downloading
	// Extract metadata first
	extractedMetadata, err := s.extractMetadata(ctx, job, basePath)
	if err != nil {
		log.WithError(err).Error("Failed to extract metadata")
	}

	isPlaylist := false
	isChannel := false

	if extractedMetadata != nil {
		switch extractedMetadata.(type) {
		case *domain.PlaylistMetadata:
			isPlaylist = true
		case *domain.ChannelMetadata:
			isChannel = true
		}
	}

	if isPlaylist || isChannel {
		// For playlists and channels, we may want to modify the download command
		// to better track the individual videos
		err = s.downloadPlaylistOrChannel(ctx, job, extractedMetadata, basePath)
	} else {
		// For single videos, use the existing download method
		err = s.downloadVideo(ctx, job, basePath)
	}

	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	job.Status = domain.JobStatusComplete
	job.Progress = 100.0

	return s.jobs.Update(&job)
}

func (s *Service) getSettings() (int, int) {
	// Get settings from database
	settings, err := s.settings.Get()
	if err != nil {
		log.WithError(err).Warn("Failed to get settings, using defaults")
		log.Debugf("Using default settings - Concurrency: %d, Quality: %dp", s.config.Concurrency, s.config.MaxQuality)
		return s.config.Concurrency, s.config.MaxQuality
	}
	log.Debugf("Retrieved settings from database - Concurrency: %d, Quality: %dp", settings.ConcurrentDownloads, settings.DownloadQuality)
	return settings.ConcurrentDownloads, settings.DownloadQuality
}

func (s *Service) getQualityForJob(job domain.Job) int {
	// If job has custom quality, use it
	if job.CustomQuality != nil {
		log.Debugf("[Job %s] Using custom quality: %dp", job.ID, *job.CustomQuality)
		return *job.CustomQuality
	}

	// Otherwise, get from settings
	_, quality := s.getSettings()
	return quality
}

func (s *Service) downloadPlaylistOrChannel(ctx context.Context, job domain.Job, metadataModel domain.Metadata, outputPath string) error {
	// Prepare the URL for playlist/channel downloads
	downloadURL := job.URL

	// For channels, ensure we're using the proper URL format for downloading all videos
	if channelMeta, ok := metadataModel.(*domain.ChannelMetadata); ok {
		// If URL is a channel page, convert it to download all videos
		if strings.Contains(downloadURL, "/channel/") || strings.Contains(downloadURL, "/@") {
			// yt-dlp can handle channel URLs directly, but we might need to specify videos explicitly
			if !strings.Contains(downloadURL, "/videos") {
				log.Infof("Channel URL detected, using: %s", downloadURL)
			}
		}
		log.Infof("Processing channel: %s with %d videos", channelMeta.Channel, channelMeta.VideoCount)
	}

	if playlistMeta, ok := metadataModel.(*domain.PlaylistMetadata); ok {
		log.Infof("Processing playlist: %s with %d items", playlistMeta.Title, playlistMeta.ItemCount)
	}
	// Create a temporary directory for archiving downloaded video IDs
	tempDir, err := os.MkdirTemp("", "archive-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	archiveFile := filepath.Join(tempDir, "archive.txt")

	// Get item count for playlists/channels for more accurate progress tracking
	totalItems := 0

	switch m := metadataModel.(type) {
	case *domain.PlaylistMetadata:
		// Use the actual number of items in the playlist
		if len(m.Items) > 0 {
			totalItems = len(m.Items)
		} else {
			totalItems = m.ItemCount
		}
	case *domain.ChannelMetadata:
		// Use the video count for channels
		if m.VideoCount > 0 {
			totalItems = m.VideoCount
		} else {
			totalItems = m.PlaylistCount
		}
	}

	// If we still don't have a valid total items count, log a warning and set a reasonable default
	if totalItems <= 0 {
		log.Warnf("No valid total items count found for playlist/channel %s, metadata might be incomplete", job.ID)
		totalItems = 1 // Set to 1 to prevent division by zero in progress calculations
	}

	log.Infof("Starting download of %d items from %s", totalItems, downloadURL)

	// Get current settings
	concurrency, _ := s.getSettings()
	maxQuality := s.getQualityForJob(job)

	if job.CustomQuality != nil {
		log.Infof("[Job %s] Starting playlist/channel download with custom quality: %dp, concurrency: %d", job.ID, maxQuality, concurrency)
	} else {
		log.Infof("[Job %s] Starting playlist/channel download with quality: %dp, concurrency: %d", job.ID, maxQuality, concurrency)
	}

	// Build the command with enhanced progress template
	// Template format: [totalItems][playlist_index][video_id][title][format_id][format_note][vcodec][acodec]prog:[bytes/total][percent][speed][eta]
	// This provides enough info to distinguish video/audio streams and track progress accurately
	cmdArgs := []string{
		"-N", fmt.Sprintf("%d", concurrency),
		"--format", fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best", maxQuality),
		"--merge-output-format", "mp4",
		"--newline", // Important for progress parsing
		"--progress-template", fmt.Sprintf(
			"[%d][%%(info.playlist_index)s][%%(info.id)s][%%(info.title).50s][%%(info.format_id)s][%%(info.format_note)s][%%(info.vcodec)s][%%(info.acodec)s]prog:[%%(progress.downloaded_bytes)s/%%(progress.total_bytes)s][%%(progress._percent_str)s][%%(progress.speed)s][%%(progress.eta)s]",
			totalItems,
		),
		"--retries", "3",             // Retry up to 3 times per fragment
		"--fragment-retries", "5",    // Retry fragments up to 5 times
		"--file-access-retries", "2", // Retry file access operations
		"--continue",
		"--ignore-errors",
		"--add-metadata",
		"--write-info-json",               // Write metadata for each video
		"--download-archive", archiveFile, // Track downloaded videos
		"--output", outputPath,
		"--yes-playlist", // Ensure playlist processing is enabled
	}
	
	// Add channel-specific arguments if this is a channel
	if _, isChannel := metadataModel.(*domain.ChannelMetadata); isChannel {
		// For channels, we may want to limit the number of videos or specify sorting
		// Add any channel-specific parameters here if needed
		log.Info("Adding channel-specific download parameters")
	}
	
	cmdArgs = append(cmdArgs, downloadURL)
	downloadCmd := exec.CommandContext(ctx, "yt-dlp", cmdArgs...)

	stdout, err := downloadCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := downloadCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := downloadCmd.Start(); err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Track progress in separate goroutines
	go s.trackProgress(stdout, job.ID, string(domain.JobTypeVideo))
	go s.trackProgress(stderr, job.ID, string(domain.JobTypeVideo))

	// Wait for command to complete
	if err := downloadCmd.Wait(); err != nil {
		return fmt.Errorf("download command failed: %w", err)
	}
	// Process the archive file to get the downloaded video IDs
	log.Debugf("Processing archive file: %s", archiveFile)
	downloadedIDs, err := s.processArchiveFile(archiveFile)
	if err != nil {
		log.WithError(err).Warn("Failed to process archive file")
	} else {
		log.Debugf("Found %d extractors with videos in archive file", len(downloadedIDs))
		for extractor, ids := range downloadedIDs {
			log.Debugf("Extractor %s has %d videos: %v", extractor, len(ids), ids)
		}
	}

	// TODO: Determine the membership type based on metadata type

	// Scan for downloaded video metadata files in the output directory
	// Note: outputPath contains yt-dlp template variables like %(uploader)s, so we need to use the actual download path
	baseDir := s.config.DownloadPath

	// Check if we have any downloaded videos to process
	if len(downloadedIDs) == 0 {
		log.Warn("No videos found in archive file - playlist/channel download may have been metadata-only or videos were skipped")
		return nil
	}

	// For each downloaded video, create a virtual job and link it to the playlist/channel
	for extractor, ids := range downloadedIDs {
		log.Debugf("Processing %d videos from extractor %s", len(ids), extractor)
		for _, id := range ids {
			// Search for metadata file by walking the directory tree
			var metadataFilePath string

			// Walk the download directory to find .info.json files
			err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip files with errors
				}

				// Skip if not a .info.json file
				if info.IsDir() || !strings.HasSuffix(path, ".info.json") {
					return nil
				}

				// Check if filename matches our patterns
				filename := filepath.Base(path)

				// Pattern 1: extractor-id.info.json (e.g., youtube-ZzI9JE0i6Lc.info.json)
				if filename == fmt.Sprintf("%s-%s.info.json", extractor, id) {
					metadataFilePath = path
					return filepath.SkipAll // Found it, stop searching
				}

				// Pattern 2: just id.info.json (e.g., ZzI9JE0i6Lc.info.json)
				if filename == fmt.Sprintf("%s.info.json", id) {
					metadataFilePath = path
					return filepath.SkipAll // Found it, stop searching
				}

				// Pattern 3: Check file content for video ID
				data, err := os.ReadFile(path)
				if err != nil {
					return nil // Skip files we can't read
				}

				var videoInfo map[string]interface{}
				if err := json.Unmarshal(data, &videoInfo); err != nil {
					return nil // Skip files that aren't valid JSON
				}

				if videoID, ok := videoInfo["id"].(string); ok && videoID == id {
					metadataFilePath = path
					return filepath.SkipAll // Found it, stop searching
				}

				return nil
			})

			if err != nil {
				log.WithError(err).Warnf("Failed to search for metadata files for video %s", id)
				continue
			}

			if metadataFilePath == "" {
				log.Warnf("Could not find metadata file for video %s from extractor %s", id, extractor)
				continue
			}

			log.Debugf("Found metadata file: %s for video %s", metadataFilePath, id)

			// Process the found metadata file
			{
				log.Debugf("Processing metadata file: %s for video %s", metadataFilePath, id)

				// Create a virtual job for this video using just the video ID
				videoJobID := id
				log.Debugf("Creating virtual job for video %s (extractor: %s)", videoJobID, extractor)

				// Check if a job already exists for this video
				existingJob, err := s.jobs.GetByID(videoJobID)
				if err == nil && existingJob != nil {
					log.Debugf("Job already exists for video %s, linking to parent", videoJobID)
					// Job exists, create membership relationship
					membershipType := "unknown"
					switch metadataModel.(type) {
					case *domain.PlaylistMetadata:
						membershipType = "playlist"
					case *domain.ChannelMetadata:
						membershipType = "channel"
					}
					
					if err := s.jobs.AddVideoToParent(videoJobID, job.ID, membershipType); err != nil {
						log.WithError(err).Warnf("Failed to link existing video %s to %s %s", videoJobID, membershipType, job.ID)
					} else {
						log.Debugf("Successfully linked existing video %s to %s %s", videoJobID, membershipType, job.ID)
					}
					continue
				}

				// Create a new virtual job for this video
				videoJob := domain.Job{
					ID:        videoJobID,
					URL:       fmt.Sprintf("https://%s.com/watch?v=%s", extractor, id),
					Status:    domain.JobStatusComplete,
					Progress:  100.0,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				log.Debugf("Creating new virtual job with ID: %s, URL: %s", videoJob.ID, videoJob.URL)

				// Create the job and store its metadata
				if err := s.jobs.Create(&videoJob); err != nil {
					log.WithError(err).Warnf("Failed to create virtual job for video %s", videoJobID)
					continue
				}
				log.Debugf("Successfully created virtual job for video %s", videoJobID)

				// Extract and store the video metadata
				videoMetadata, err := metadata.ExtractMetadata(metadataFilePath)

				if err != nil {
					log.WithError(err).Warnf("Failed to extract metadata for video %s", videoJobID)
					continue
				}

				if err := s.jobs.StoreMetadata(videoJobID, videoMetadata); err != nil {
					log.WithError(err).Warnf("Failed to store metadata for video %s", videoJobID)
					continue
				}
				log.Debugf("Successfully stored metadata for video %s", videoJobID)

				// Link the video to the playlist/channel
				membershipType := "unknown"
				switch metadataModel.(type) {
				case *domain.PlaylistMetadata:
					membershipType = "playlist"
				case *domain.ChannelMetadata:
					membershipType = "channel"
				}

				if err := s.jobs.AddVideoToParent(videoJobID, job.ID, membershipType); err != nil {
					log.WithError(err).Warnf("Failed to link video %s to %s %s", videoJobID, membershipType, job.ID)
				} else {
					log.Debugf("Successfully linked video %s to %s %s", videoJobID, membershipType, job.ID)
				}
			}
		}
	}

	return nil
}

func (s *Service) processArchiveFile(archiveFile string) (map[string][]string, error) {
	// Read the archive file to get the IDs of downloaded videos
	log.Debugf("Reading archive file: %s", archiveFile)
	data, err := os.ReadFile(archiveFile)
	if err != nil {
		log.WithError(err).Errorf("Failed to read archive file: %s", archiveFile)
		return nil, err
	}

	log.Debugf("Archive file content (%d bytes): %s", len(data), string(data))
	lines := strings.Split(string(data), "\n")
	result := make(map[string][]string)

	log.Debugf("Processing %d lines from archive file", len(lines))

	for i, line := range lines {
		if line == "" {
			continue
		}

		log.Debugf("Archive line %d: %s", i, line)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			log.Warnf("Invalid archive line format: %s", line)
			continue
		}

		extractor := parts[0]
		id := parts[1]

		log.Debugf("Parsed extractor: %s, id: %s", extractor, id)

		result[extractor] = append(result[extractor], id)
	}

	log.Debugf("Processed archive file, found %d extractors", len(result))
	return result, nil
}

func (s *Service) downloadVideo(ctx context.Context, job domain.Job, outputPath string) error {
	// Get current settings
	concurrency, _ := s.getSettings()
	maxQuality := s.getQualityForJob(job)

	if job.CustomQuality != nil {
		log.Infof("[Job %s] Starting video download with custom quality: %dp, concurrency: %d", job.ID, maxQuality, concurrency)
	} else {
		log.Infof("[Job %s] Starting video download with quality: %dp, concurrency: %d", job.ID, maxQuality, concurrency)
	}

	downloadCmd := exec.CommandContext(ctx, "yt-dlp",
		"-N", fmt.Sprintf("%d", concurrency),
		"--format", fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best", maxQuality),
		"--merge-output-format", "mp4",
		"--newline",
		// Enhanced progress template with format info to distinguish video/audio streams
		"--progress-template", "[NA][NA][%(info.id)s][%(info.title).50s][%(info.format_id)s][%(info.format_note)s][%(info.vcodec)s][%(info.acodec)s]prog:[%(progress.downloaded_bytes)s/%(progress.total_bytes)s][%(progress._percent_str)s][%(progress.speed)s][%(progress.eta)s]",
		"--retries", "3",            // Retry up to 3 times per fragment
		"--fragment-retries", "5",   // Retry fragments up to 5 times
		"--file-access-retries", "2", // Retry file access operations
		"--continue",
		"--ignore-errors",
		"--add-metadata",
		"--write-info-json", // Write metadata with actual downloaded format info
		"--output", outputPath,
		job.URL,
	)

	stdout, err := downloadCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := downloadCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := downloadCmd.Start(); err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	// Track progress in separate goroutines
	go s.trackProgress(stdout, job.ID, string(domain.JobTypeVideo))
	go s.trackProgress(stderr, job.ID, string(domain.JobTypeVideo))

	// Wait for command to complete
	err = downloadCmd.Wait()
	if err != nil {
		log.WithError(err).WithField("jobID", job.ID).Error("yt-dlp command failed")
		return err
	}

	log.WithField("jobID", job.ID).Info("yt-dlp download completed successfully")

	// After download, update metadata with actual downloaded resolution
	if err := s.updateDownloadedMetadata(job.ID, maxQuality); err != nil {
		log.WithError(err).Warn("Failed to update downloaded metadata, continuing...")
	}

	return nil
}

func (s *Service) extractMetadata(ctx context.Context, job domain.Job, outputPath string) (domain.Metadata, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	metadataCmd := exec.CommandContext(ctx, "yt-dlp",
		"--skip-download",
		"--write-info-json",
		"--no-progress",
		"--flat-playlist",
		"--output", outputPath,
		job.URL,
	)

	stdoutPipe, err := metadataCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := metadataCmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	stdoutWriter := io.MultiWriter(&stdoutBuf, os.Stdout)
	stderrWriter := io.MultiWriter(&stderrBuf, os.Stderr)

	go io.Copy(stdoutWriter, stdoutPipe)
	go io.Copy(stderrWriter, stderrPipe)

	update := domain.ProgressUpdate{
		JobID:    job.ID,
		JobType:  string(domain.JobTypeMetadata),
		Progress: 0,
	}
	s.hub.broadcast <- update

	if err := metadataCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start metadata extraction: %w", err)
	}

	if err := metadataCmd.Wait(); err != nil {
		return nil, fmt.Errorf("metadata extraction failed: %w", err)
	}

	// New regex pattern to match both video and playlist/channel info JSON paths
	infoJsonRegex := regexp.MustCompile(`Writing (?:video|playlist) metadata as JSON to: (.+\.info\.json)`)

	var metadataPath string
	for _, output := range []string{stdoutBuf.String(), stderrBuf.String()} {
		if matches := infoJsonRegex.FindStringSubmatch(output); len(matches) > 1 {
			metadataPath = matches[1]
			break
		}
	}

	if metadataPath == "" {
		return nil, fmt.Errorf("could not find metadata file path in command output")
	}

	extractedMetadata, err := metadata.ExtractMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata from file: %w", err)
	}

	// Send immediate basic metadata update to give user instant feedback
	if err := s.jobs.StoreMetadata(job.ID, extractedMetadata); err != nil {
		log.WithError(err).Warn("Failed to store basic metadata, continuing...")
	} else {
		// Send the basic metadata update immediately
		basicMetadataUpdate := domain.MetadataUpdate{
			JobID:    job.ID,
			Metadata: extractedMetadata,
		}
		s.hub.broadcast <- basicMetadataUpdate
		log.Debug("Sent immediate basic metadata update to UI")
	}

	// Phase 2: Enhance metadata for playlists and channels with detailed info
	if isPlaylistOrChannel(extractedMetadata) {
		log.Debug("Enhancing metadata for playlist/channel with detailed information")
		err = s.enhanceMetadata(ctx, job, extractedMetadata)
		if err != nil {
			log.WithError(err).Warn("Failed to enhance metadata, continuing with basic metadata")

			// Log the basic metadata to help debug
			switch m := extractedMetadata.(type) {
			case *domain.PlaylistMetadata:
				log.Debugf("Basic playlist metadata: %d items", m.ItemCount)
			case *domain.ChannelMetadata:
				log.Debugf("Basic channel metadata: %d playlists, %d videos", m.PlaylistCount, m.VideoCount)
			}
		} else {
			// Log enhanced metadata info
			switch m := extractedMetadata.(type) {
			case *domain.PlaylistMetadata:
				log.Debugf("Enhanced playlist metadata: %d items in Items slice", len(m.Items))
			case *domain.ChannelMetadata:
				log.Debugf("Enhanced channel metadata: %d videos, %d recent videos", m.VideoCount, len(m.RecentVideos))
			}

			// Store and broadcast the enhanced metadata
			if err := s.jobs.StoreMetadata(job.ID, extractedMetadata); err != nil {
				log.WithError(err).Warn("Failed to store enhanced metadata")
			} else {
				enhancedMetadataUpdate := domain.MetadataUpdate{
					JobID:    job.ID,
					Metadata: extractedMetadata,
				}
				s.hub.broadcast <- enhancedMetadataUpdate
				log.Debug("Sent enhanced metadata update to UI")
			}
		}
	}

	update.Progress = 1
	s.hub.broadcast <- update

	return extractedMetadata, nil
}

func (s *Service) enhanceMetadata(ctx context.Context, job domain.Job, basicMetadata domain.Metadata) error {
	// Create a temporary output path for the detailed extraction
	tempDir, err := os.MkdirTemp("", "detailed-metadata-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Run yt-dlp with --dump-single-json to get comprehensive info
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--skip-download",
		"--dump-single-json",
		"--no-flat-playlist", // Get full information about playlist items
		"--write-playlist-metafiles",
		job.URL,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("detailed metadata extraction failed: %w", err)
	}

	// Parse the JSON output
	var detailedData map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &detailedData); err != nil {
		return fmt.Errorf("failed to parse detailed metadata: %w", err)
	}

	// Enhance the basic metadata with detailed information
	switch m := basicMetadata.(type) {
	case *domain.PlaylistMetadata:
		return s.enhancePlaylistMetadata(m, detailedData)
	case *domain.ChannelMetadata:
		return s.enhanceChannelMetadata(m, detailedData)
	default:
		return fmt.Errorf("unknown metadata type: %T", basicMetadata)
	}
}

func (s *Service) enhancePlaylistMetadata(playlist *domain.PlaylistMetadata, detailedData map[string]interface{}) error {
	// Extract the entries/videos from the detailed JSON
	entries, ok := detailedData["entries"].([]interface{})
	if !ok {
		return fmt.Errorf("no entries found in detailed playlist data")
	}

	// Create enhanced playlist items
	playlist.Items = make([]domain.PlaylistItem, 0, len(entries))

	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		// Create a full PlaylistItem with more metadata
		item := domain.PlaylistItem{
			ID:             getString(entryMap, "id"),
			Title:          getString(entryMap, "title"),
			Description:    getString(entryMap, "description"),
			Thumbnail:      extractThumbnailURL(entryMap),
			Duration:       getInt(entryMap, "duration"),
			DurationString: formatDuration(getInt(entryMap, "duration")),
			UploadDate:     getString(entryMap, "upload_date"),
			ViewCount:      getInt(entryMap, "view_count"),
			LikeCount:      getInt(entryMap, "like_count"),
			Channel:        getString(entryMap, "channel"),
			ChannelID:      getString(entryMap, "channel_id"),
			ChannelURL:     getString(entryMap, "channel_url"),
			Width:          getInt(entryMap, "width"),
			Height:         getInt(entryMap, "height"),
			Resolution:     getString(entryMap, "resolution"),
			FileSize:       getInt64(entryMap, "filesize_approx"),
			Format:         getString(entryMap, "format"),
			Extension:      getString(entryMap, "ext"),
		}

		// Extract tags if present
		if tagsArray, ok := entryMap["tags"].([]interface{}); ok {
			tags := make([]string, 0, len(tagsArray))
			for _, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			item.Tags = tags
		}

		playlist.Items = append(playlist.Items, item)
	}

	// Update playlist metadata with more accurate counts
	playlist.ItemCount = len(playlist.Items)

	// Sum up view counts from items if playlist view count is missing
	if playlist.ViewCount == 0 && len(playlist.Items) > 0 {
		totalViews := 0
		for _, item := range playlist.Items {
			totalViews += item.ViewCount
		}
		playlist.ViewCount = totalViews
	}

	return nil
}

func (s *Service) enhanceChannelMetadata(channel *domain.ChannelMetadata, detailedData map[string]interface{}) error {
	// Extract the entries/videos from the detailed JSON
	entries, ok := detailedData["entries"].([]interface{})
	if !ok {
		return fmt.Errorf("no entries found in detailed channel data")
	}

	// Update channel video count
	channel.VideoCount = len(entries)

	// Add most recent videos to channel metadata (up to 10)
	recentVideos := make([]domain.PlaylistItem, 0, min(10, len(entries)))
	totalViews := 0
	totalStorage := int64(0)

	for i, entry := range entries {
		if i >= 10 {
			break // Only process the first 10 videos for recent videos list
		}

		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract view count for total channel views calculation
		viewCount := getInt(entryMap, "view_count")
		totalViews += viewCount

		// Extract file size for total storage calculation
		fileSize := getInt64(entryMap, "filesize_approx")
		totalStorage += fileSize

		// Create PlaylistItem for recent videos
		item := domain.PlaylistItem{
			ID:             getString(entryMap, "id"),
			Title:          getString(entryMap, "title"),
			Description:    getString(entryMap, "description"),
			Thumbnail:      extractThumbnailURL(entryMap),
			Duration:       getInt(entryMap, "duration"),
			DurationString: formatDuration(getInt(entryMap, "duration")),
			UploadDate:     getString(entryMap, "upload_date"),
			ViewCount:      viewCount,
			LikeCount:      getInt(entryMap, "like_count"),
			Channel:        getString(entryMap, "channel"),
			ChannelID:      getString(entryMap, "channel_id"),
			ChannelURL:     getString(entryMap, "channel_url"),
		}

		recentVideos = append(recentVideos, item)
	}

	// Update channel metadata with calculated statistics
	channel.TotalViews = totalViews
	channel.TotalStorage = totalStorage
	channel.RecentVideos = recentVideos

	return nil
}

func (s *Service) GetJobWithMetadata(jobID string) (*domain.JobWithMetadata, error) {
	return s.jobs.GetJobWithMetadata(jobID)
}

func (s *Service) GetRecentWithMetadata(limit int) ([]*domain.JobWithMetadata, error) {
	return s.jobs.GetRecentWithMetadata(limit)
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func isPlaylistOrChannel(metadata domain.Metadata) bool {
	switch metadata.(type) {
	case *domain.PlaylistMetadata, *domain.ChannelMetadata:
		return true
	default:
		return false
	}
}

func getInt(data map[string]interface{}, key string) int {
	switch v := data[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func getInt64(data map[string]interface{}, key string) int64 {
	switch v := data[key].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func extractThumbnailURL(data map[string]interface{}) string {
	thumbnails, ok := data["thumbnails"].([]interface{})
	if !ok || len(thumbnails) == 0 {
		return ""
	}

	// Get the best quality thumbnail (usually the last one)
	thumbnail, ok := thumbnails[len(thumbnails)-1].(map[string]interface{})
	if !ok {
		return ""
	}

	if url, ok := thumbnail["url"].(string); ok {
		return url
	}
	return ""
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return ""
	}

	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// updateDownloadedMetadata updates the video metadata with the actual downloaded resolution
func (s *Service) updateDownloadedMetadata(jobID string, maxQuality int) error {
	// Get the existing metadata
	jobWithMetadata, err := s.jobs.GetJobWithMetadata(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job metadata: %w", err)
	}

	videoMetadata, ok := jobWithMetadata.Metadata.(*domain.VideoMetadata)
	if !ok {
		return fmt.Errorf("metadata is not VideoMetadata type")
	}

	// Cap the resolution based on what was actually downloaded
	// The downloaded resolution is the minimum of the original resolution and the quality setting
	originalHeight := videoMetadata.Height
	downloadedHeight := min(originalHeight, maxQuality)

	if downloadedHeight != originalHeight {
		// Calculate new width maintaining aspect ratio
		aspectRatio := float64(videoMetadata.Width) / float64(videoMetadata.Height)
		downloadedWidth := int(float64(downloadedHeight) * aspectRatio)

		// Update the metadata with actual downloaded resolution
		videoMetadata.Height = downloadedHeight
		videoMetadata.Width = downloadedWidth
		videoMetadata.Resolution = fmt.Sprintf("%dx%d", downloadedWidth, downloadedHeight)

		log.Infof("[Job %s] Updated metadata: original %dx%d -> downloaded %dx%d (quality limit: %dp)",
			jobID, videoMetadata.Width, originalHeight, downloadedWidth, downloadedHeight, maxQuality)

		// Store the updated metadata
		if err := s.jobs.StoreMetadata(jobID, videoMetadata); err != nil {
			return fmt.Errorf("failed to store updated metadata: %w", err)
		}
	} else {
		log.Debugf("[Job %s] Video resolution %dp is within quality limit %dp, no update needed",
			jobID, originalHeight, maxQuality)
	}

	return nil
}
