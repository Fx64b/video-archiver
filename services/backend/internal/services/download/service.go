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
	JobRepository domain.JobRepository
	DownloadPath  string
	Concurrency   int
	MaxQuality    int
}

type Service struct {
	config        *Config
	jobs          domain.JobRepository
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
		config: config,
		jobs:   config.JobRepository,
		queue:  make(chan domain.Job, 100),
		hub:    hub,
		ctx:    ctx,
		cancel: cancel,
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

func (s *Service) downloadPlaylistOrChannel(ctx context.Context, job domain.Job, metadataModel domain.Metadata, outputPath string) error {
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
		totalItems = m.ItemCount
	case *domain.ChannelMetadata:
		totalItems = m.PlaylistCount
	}

	// Build the command with progress template
	downloadCmd := exec.CommandContext(ctx, "yt-dlp",
		"-N", fmt.Sprintf("%d", s.config.Concurrency),
		"--format", fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best", s.config.MaxQuality),
		"--merge-output-format", "mp4",
		"--newline", // Important for progress parsing
		"--progress-template", fmt.Sprintf("[%d][%%(info.playlist_index)s][%%(info.id)s][%%(info.title).50s]prog:[%%(progress.downloaded_bytes)s/%%(progress.total_bytes)s][%%(progress._percent_str)s][%%(progress.speed)s][%%(progress.eta)s]", totalItems),
		"--retries", "3",
		"--continue",
		"--ignore-errors",
		"--add-metadata",
		"--write-info-json",               // Write metadata for each video
		"--download-archive", archiveFile, // Track downloaded videos
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
	if err := downloadCmd.Wait(); err != nil {
		return fmt.Errorf("download command failed: %w", err)
	}
	// Process the archive file to get the downloaded video IDs
	downloadedIDs, err := s.processArchiveFile(archiveFile)
	if err != nil {
		log.WithError(err).Warn("Failed to process archive file")
	}

	// Determine the membership type based on metadata type
	var membershipType string
	switch metadataModel.(type) {
	case *domain.PlaylistMetadata:
		membershipType = "playlist"
	case *domain.ChannelMetadata:
		membershipType = "channel"
	default:
		membershipType = "unknown"
	}

	// Scan for downloaded video metadata files in the output directory
	baseDir := filepath.Dir(outputPath)

	// For each downloaded video, create a virtual job and link it to the playlist/channel
	for extractor, ids := range downloadedIDs {
		for _, id := range ids {
			// First, try to find the exact metadata file
			pattern := filepath.Join(baseDir, "*", fmt.Sprintf("%s-%s.info.json", extractor, id))
			matches, err := filepath.Glob(pattern)

			if err != nil || len(matches) == 0 {
				// Try a less specific pattern if the exact one doesn't work
				pattern = filepath.Join(baseDir, "*", "*.info.json")
				matches, err = filepath.Glob(pattern)

				if err != nil || len(matches) == 0 {
					log.Warnf("Could not find metadata file for video %s-%s", extractor, id)
					continue
				}
			}

			// Process each found metadata file
			for _, metadataFilePath := range matches {
				videoData, err := os.ReadFile(metadataFilePath)
				if err != nil {
					log.WithError(err).Warnf("Failed to read video metadata file: %s", metadataFilePath)
					continue
				}

				// Check if this is the video we're looking for
				var videoInfo map[string]interface{}
				if err := json.Unmarshal(videoData, &videoInfo); err != nil {
					log.WithError(err).Warn("Failed to parse video metadata JSON")
					continue
				}

				videoID, ok := videoInfo["id"].(string)
				if !ok || videoID != id {
					// This is not the video we're looking for
					continue
				}

				// Create a virtual job for this video
				videoJobID := fmt.Sprintf("%s-%s", extractor, id)

				// Check if a job already exists for this video
				existingJob, err := s.jobs.GetByID(videoJobID)
				if err == nil && existingJob != nil {
					// Job exists, create membership relationship
					if err := s.linkVideoToParent(videoJobID, job.ID, membershipType); err != nil {
						log.WithError(err).Warnf("Failed to link video %s to %s %s", videoJobID, membershipType, job.ID)
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

				// Create the job and store its metadata
				if err := s.jobs.Create(&videoJob); err != nil {
					log.WithError(err).Warnf("Failed to create virtual job for video %s", videoJobID)
					continue
				}

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

				// Link the video to the playlist/channel
				if err := s.linkVideoToParent(videoJobID, job.ID, membershipType); err != nil {
					log.WithError(err).Warnf("Failed to link video %s to %s %s", videoJobID, membershipType, job.ID)
				}
			}
		}
	}

	return nil
}

func (s *Service) linkVideoToParent(videoJobID, parentJobID, membershipType string) error {
	// Ideally, this would call a repository method
	// But for now, we'll implement a basic version in the service

	// TODO: This should be implemented in the repository
	log.Infof("Linking video %s to %s %s", videoJobID, membershipType, parentJobID)

	// Placeholder for actual database operation
	// This would be implemented in the repository layer
	return nil
}

func (s *Service) processArchiveFile(archiveFile string) (map[string][]string, error) {
	// Read the archive file to get the IDs of downloaded videos
	data, err := os.ReadFile(archiveFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	result := make(map[string][]string)

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		extractor := parts[0]
		id := parts[1]

		result[extractor] = append(result[extractor], id)
	}

	return result, nil
}

func (s *Service) downloadVideo(ctx context.Context, job domain.Job, outputPath string) error {
	downloadCmd := exec.CommandContext(ctx, "yt-dlp",
		"-N", fmt.Sprintf("%d", s.config.Concurrency),
		"--format", fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best", s.config.MaxQuality),
		"--merge-output-format", "mp4",
		"--newline",
		"--progress-template", "[NA][NA][%(info.id)s][%(info.title).50s]prog:[%(progress.downloaded_bytes)s/%(progress.total_bytes)s][%(progress._percent_str)s][%(progress.speed)s][%(progress.eta)s]",
		"--retries", "3",
		"--continue",
		"--ignore-errors",
		"--add-metadata",
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
	return downloadCmd.Wait()
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

	// Phase 2: Enhance metadata for playlists and channels with detailed info
	if isPlaylistOrChannel(extractedMetadata) {
		err = s.enhanceMetadata(ctx, job, extractedMetadata)
		if err != nil {
			log.WithError(err).Warn("Failed to enhance metadata, continuing with basic metadata")
		}
	}

	if err := s.jobs.StoreMetadata(job.ID, extractedMetadata); err != nil {
		return nil, fmt.Errorf("failed to store metadata: %w", err)
	}

	metadataUpdate := domain.MetadataUpdate{
		JobID:    job.ID,
		Metadata: extractedMetadata,
	}
	s.hub.broadcast <- metadataUpdate

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
