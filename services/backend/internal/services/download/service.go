package download

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"sync"
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
	job.Status = domain.JobStatusInProgress
	if err := s.jobs.Update(&job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	basePath := fmt.Sprintf("%s/%%(uploader)s/%%(title)s", s.config.DownloadPath)

	// TODO: this part is incorrect because the metadata path is only known after metadata has been extracted
	// Instead you should read the metdata path from the console output with a regex like "[info] * .info.json"
	metadataPath := basePath + ".%(ext)s"

	videoPath := basePath + ".%(ext)s"

	if err := s.extractMetadata(ctx, job, metadataPath); err != nil {
		log.WithError(err).Error("Failed to extract metadata")
	}

	if err := s.downloadVideo(ctx, job, videoPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	job.Status = domain.JobStatusComplete
	job.Progress = 100.0

	return s.jobs.Update(&job)
}

func (s *Service) downloadVideo(ctx context.Context, job domain.Job, outputPath string) error {
	downloadCmd := exec.CommandContext(ctx, "yt-dlp",
		"-N", fmt.Sprintf("%d", s.config.Concurrency),
		"--format", fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best", s.config.MaxQuality),
		"--merge-output-format", "mp4",
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

func (s *Service) extractMetadata(ctx context.Context, job domain.Job, outputPath string) error {
	//pathRegex := regexp.MustCompile(`\[info\]\s+.+\.(f\d+)\.info.json`)

	metadataCmd := exec.CommandContext(ctx, "yt-dlp",
		"--skip-download",
		"--write-info-json",
		"--no-progress",
		"--output", outputPath,
		job.URL,
	)

	var metadataPath bytes.Buffer
	metadataCmd.Stdout = &metadataPath

	if err := metadataCmd.Run(); err != nil {
		return fmt.Errorf("metadata extraction failed: %w", err)
	}

	// TODO: pretty sure this does not work
	// Get the actual path from yt-dlp output
	actualPath := strings.TrimSpace(metadataPath.String())

	// Extract and process metadata
	extractedMetadata, err := metadata.ExtractMetadata(actualPath)
	if err != nil {
		return fmt.Errorf("failed to extract metadata from file: %w", err)
	}

	stdout, err := metadataCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := metadataCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Notify about metadata extraction start
	update := domain.ProgressUpdate{
		JobID:    job.ID,
		JobType:  string(domain.JobTypeMetadata),
		Progress: 0,
	}
	s.hub.broadcast <- update

	if err := metadataCmd.Start(); err != nil {
		return fmt.Errorf("failed to start metadata extraction: %w", err)
	}

	// Track progress
	go s.trackProgress(stdout, job.ID, string(domain.JobTypeMetadata))
	go s.trackProgress(stderr, job.ID, string(domain.JobTypeMetadata))

	// Wait for command completion
	if err := metadataCmd.Wait(); err != nil {
		return fmt.Errorf("metadata extraction failed: %w", err)
	}

	// The metadata file will be at outputPath + ".info.json"
	actualMetadataPath := outputPath + ".info.json"

	// Extract and process metadata
	extractedMetadata, err = metadata.ExtractMetadata(actualMetadataPath)
	if err != nil {
		return fmt.Errorf("failed to extract metadata from file: %w", err)
	}

	// Store metadata in database
	if err := s.jobs.StoreMetadata(job.ID, extractedMetadata); err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	// Send metadata update to frontend
	metadataUpdate := domain.MetadataUpdate{
		JobID:    job.ID,
		Metadata: extractedMetadata,
	}
	s.hub.broadcast <- metadataUpdate

	// Send completion update
	update.Progress = 100
	s.hub.broadcast <- update

	return nil
}

func (s *Service) storeMetadata(ctx context.Context, jobID string, metadata *domain.VideoMetadata) error {
	return s.jobs.StoreMetadata(jobID, metadata)
}

func (s *Service) processMetadata(ctx context.Context, job domain.Job, metadataPath string) error {
	extractedMetadata, err := metadata.ExtractMetadata(metadataPath)
	if err != nil {
		return err
	}

	update := domain.MetadataUpdate{
		JobID:    job.ID,
		Metadata: extractedMetadata,
	}
	s.hub.broadcast <- update

	return s.storeMetadata(ctx, job.ID, extractedMetadata)
}
