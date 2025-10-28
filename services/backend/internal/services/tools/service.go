package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

type Config struct {
	ToolsRepository    domain.ToolsRepository
	JobRepository      domain.JobRepository
	SettingsRepository domain.SettingsRepository
	DownloadPath       string
	ProcessedPath      string
	Concurrency        int
}

type Service struct {
	config         *Config
	toolsRepo      domain.ToolsRepository
	jobRepo        domain.JobRepository
	settingsRepo   domain.SettingsRepository
	queue          chan *domain.ToolsJob
	wg             sync.WaitGroup
	hub            *WebSocketHub
	ctx            context.Context
	cancel         context.CancelFunc
	ffmpeg         *FFmpegWrapper
	activeJobs     sync.Map // map[string]*domain.ToolsJob
	downloadPath   string
	processedPath  string
}

func NewService(config *Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	hub := NewWebSocketHub()

	if config.ProcessedPath == "" {
		config.ProcessedPath = "./data/processed"
	}

	// Ensure processed directory exists
	if err := os.MkdirAll(config.ProcessedPath, 0755); err != nil {
		log.WithError(err).Warn("Failed to create processed directory")
	}

	return &Service{
		config:        config,
		toolsRepo:     config.ToolsRepository,
		jobRepo:       config.JobRepository,
		settingsRepo:  config.SettingsRepository,
		queue:         make(chan *domain.ToolsJob, 100),
		hub:           hub,
		ctx:           ctx,
		cancel:        cancel,
		ffmpeg:        NewFFmpegWrapper(),
		downloadPath:  config.DownloadPath,
		processedPath: config.ProcessedPath,
	}
}

func (s *Service) Start() error {
	go s.hub.Run()

	// Start worker pool
	concurrency := s.config.Concurrency
	if concurrency < 1 {
		concurrency = 2 // Default: 2 concurrent tools jobs
	}

	for i := 0; i < concurrency; i++ {
		s.wg.Add(1)
		go s.processJobs()
	}

	log.WithField("concurrency", concurrency).Info("Tools service started")
	return nil
}

func (s *Service) Stop() {
	log.Info("Stopping tools service...")
	s.cancel()
	close(s.queue)
	s.wg.Wait()
	log.Info("Tools service stopped")
}

func (s *Service) GetHub() *WebSocketHub {
	return s.hub
}

func (s *Service) GetRepository() domain.ToolsRepository {
	return s.toolsRepo
}

func (s *Service) Submit(job *domain.ToolsJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	job.Status = domain.ToolsJobStatusPending
	job.Progress = 0
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	if err := s.toolsRepo.Create(job); err != nil {
		return fmt.Errorf("create tools job: %w", err)
	}

	s.queue <- job
	log.WithField("job_id", job.ID).WithField("operation", job.OperationType).Info("Tools job submitted")

	return nil
}

func (s *Service) GetJobByID(id string) (*domain.ToolsJob, error) {
	return s.toolsRepo.GetByID(id)
}

func (s *Service) CancelJob(id string) error {
	job, err := s.toolsRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found")
	}

	if job.Status == domain.ToolsJobStatusComplete || job.Status == domain.ToolsJobStatusFailed {
		return fmt.Errorf("cannot cancel completed or failed job")
	}

	job.Status = domain.ToolsJobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	if err := s.toolsRepo.Update(job); err != nil {
		return fmt.Errorf("update job: %w", err)
	}

	s.sendProgress(job.ID, job.Status, 0, "Job cancelled", 0, 0)

	log.WithField("job_id", id).Info("Tools job cancelled")
	return nil
}

func (s *Service) processJobs() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case job, ok := <-s.queue:
			if !ok {
				return
			}

			s.activeJobs.Store(job.ID, job)
			s.processJob(job)
			s.activeJobs.Delete(job.ID)
		}
	}
}

func (s *Service) processJob(job *domain.ToolsJob) {
	log.WithField("job_id", job.ID).WithField("operation", job.OperationType).Info("Processing tools job")

	// Update status to processing
	job.Status = domain.ToolsJobStatusProcessing
	job.Progress = 0
	if err := s.toolsRepo.Update(job); err != nil {
		log.WithError(err).Error("Failed to update job status")
		return
	}

	s.sendProgress(job.ID, job.Status, 0, "Starting...", 0, 0)

	// Expand input files if needed (playlist/channel to videos)
	inputFiles, err := s.expandInputFiles(job)
	if err != nil {
		s.failJob(job, fmt.Errorf("expand input files: %w", err))
		return
	}

	// Get actual file paths from job IDs
	filePaths, err := s.resolveFilePaths(inputFiles)
	if err != nil {
		s.failJob(job, fmt.Errorf("resolve file paths: %w", err))
		return
	}

	// Generate output file path
	outputPath, err := s.generateOutputPath(job)
	if err != nil {
		s.failJob(job, fmt.Errorf("generate output path: %w", err))
		return
	}

	// Execute operation
	var execErr error
	switch job.OperationType {
	case domain.OpTypeTrim:
		execErr = s.executeTrim(job, filePaths[0], outputPath)
	case domain.OpTypeConcat:
		execErr = s.executeConcat(job, filePaths, outputPath)
	case domain.OpTypeConvert:
		execErr = s.executeConvert(job, filePaths[0], outputPath)
	case domain.OpTypeExtractAudio:
		execErr = s.executeExtractAudio(job, filePaths[0], outputPath)
	case domain.OpTypeAdjustQuality:
		execErr = s.executeAdjustQuality(job, filePaths[0], outputPath)
	case domain.OpTypeRotate:
		execErr = s.executeRotate(job, filePaths[0], outputPath)
	case domain.OpTypeWorkflow:
		execErr = s.executeWorkflow(job, filePaths)
	default:
		execErr = fmt.Errorf("unsupported operation type: %s", job.OperationType)
	}

	if execErr != nil {
		s.failJob(job, execErr)
		return
	}

	// Mark as complete
	job.Status = domain.ToolsJobStatusComplete
	job.Progress = 100
	job.OutputFile = outputPath
	now := time.Now()
	job.CompletedAt = &now

	// Get actual file size
	if stat, err := os.Stat(outputPath); err == nil {
		job.ActualSize = stat.Size()
	}

	if err := s.toolsRepo.Update(job); err != nil {
		log.WithError(err).Error("Failed to update completed job")
		return
	}

	s.sendProgress(job.ID, job.Status, 100, "Complete", 0, 0)

	log.WithField("job_id", job.ID).WithField("output", outputPath).Info("Tools job completed")
}

func (s *Service) failJob(job *domain.ToolsJob, err error) {
	log.WithError(err).WithField("job_id", job.ID).Error("Tools job failed")

	job.Status = domain.ToolsJobStatusFailed
	job.ErrorMessage = err.Error()
	now := time.Now()
	job.CompletedAt = &now

	if updateErr := s.toolsRepo.Update(job); updateErr != nil {
		log.WithError(updateErr).Error("Failed to update failed job")
	}

	s.sendProgress(job.ID, job.Status, job.Progress, "Failed", 0, 0)
}

func (s *Service) sendProgress(jobID string, status domain.ToolsJobStatus, progress float64, step string, elapsed, remaining int) {
	update := domain.ToolsProgressUpdate{
		JobID:         jobID,
		Status:        status,
		Progress:      progress,
		CurrentStep:   step,
		TimeElapsed:   elapsed,
		TimeRemaining: remaining,
	}

	s.hub.Broadcast(update)
}

func (s *Service) expandInputFiles(job *domain.ToolsJob) ([]string, error) {
	switch job.InputType {
	case "videos", "":
		// Already individual video job IDs
		return job.InputFiles, nil

	case "playlist":
		// Expand playlist to video IDs
		if len(job.InputFiles) != 1 {
			return nil, fmt.Errorf("playlist input type requires exactly one playlist ID")
		}

		playlistJobID := job.InputFiles[0]
		videos, err := s.jobRepo.GetVideosForParent(playlistJobID)
		if err != nil {
			return nil, fmt.Errorf("get videos for playlist: %w", err)
		}

		videoIDs := make([]string, len(videos))
		for i, v := range videos {
			videoIDs[i] = v.Job.ID
		}

		log.WithField("playlist_id", playlistJobID).WithField("video_count", len(videoIDs)).Info("Expanded playlist to videos")
		return videoIDs, nil

	case "channel":
		// Expand channel to video IDs
		if len(job.InputFiles) != 1 {
			return nil, fmt.Errorf("channel input type requires exactly one channel ID")
		}

		channelJobID := job.InputFiles[0]

		// Get all videos for the channel
		// For now, get via parent relationship - in production would query by channel_id in metadata
		videos, err := s.jobRepo.GetVideosForParent(channelJobID)
		if err != nil {
			return nil, fmt.Errorf("get videos for channel: %w", err)
		}

		videoIDs := make([]string, len(videos))
		for i, v := range videos {
			videoIDs[i] = v.Job.ID
		}

		log.WithField("channel_id", channelJobID).WithField("video_count", len(videoIDs)).Info("Expanded channel to videos")
		return videoIDs, nil

	default:
		return nil, fmt.Errorf("invalid input type: %s", job.InputType)
	}
}

func (s *Service) resolveFilePaths(jobIDs []string) ([]string, error) {
	var paths []string

	for _, jobID := range jobIDs {
		jobWithMetadata, err := s.jobRepo.GetJobWithMetadata(jobID)
		if err != nil {
			return nil, fmt.Errorf("get job %s: %w", jobID, err)
		}

		if jobWithMetadata == nil || jobWithMetadata.Job == nil {
			return nil, fmt.Errorf("job %s not found", jobID)
		}

		// Get video metadata
		videoMeta, ok := jobWithMetadata.Metadata.(*domain.VideoMetadata)
		if !ok {
			return nil, fmt.Errorf("job %s is not a video", jobID)
		}

		// Construct file path
		channelDir := videoMeta.Channel
		if channelDir == "" {
			channelDir = videoMeta.Uploader
		}
		if channelDir == "" {
			channelDir = "Unknown"
		}

		title := videoMeta.Title
		if title == "" {
			title = "Unknown"
		}

		// Clean filename
		title = cleanFilename(title)

		filePath := filepath.Join(s.downloadPath, channelDir, title+".mp4")

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("video file not found: %s", filePath)
		}

		paths = append(paths, filePath)
	}

	return paths, nil
}

func (s *Service) generateOutputPath(job *domain.ToolsJob) (string, error) {
	// Generate unique filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s", job.OperationType, timestamp, job.ID[:8])

	// Add appropriate extension based on operation
	var ext string
	switch job.OperationType {
	case domain.OpTypeExtractAudio:
		// Get format from parameters
		if format, ok := job.Parameters["output_format"].(string); ok {
			ext = "." + format
		} else {
			ext = ".mp3"
		}
	case domain.OpTypeConvert:
		// Get format from parameters
		if format, ok := job.Parameters["output_format"].(string); ok {
			ext = "." + format
		} else {
			ext = ".mp4"
		}
	default:
		ext = ".mp4"
	}

	filename += ext
	return filepath.Join(s.processedPath, filename), nil
}

func cleanFilename(s string) string {
	// Replace invalid filename characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(s)
}

// Helper to parse parameters into specific types
func parseParameters[T any](params map[string]any) (*T, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal parameters: %w", err)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal parameters: %w", err)
	}

	return &result, nil
}
