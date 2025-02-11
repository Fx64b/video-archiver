package download

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
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

	outputPath := fmt.Sprintf("%s/%%(uploader)s/%%(playlist|title)s/%%(title)s.%%(ext)s",
		s.config.DownloadPath)

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-N", fmt.Sprintf("%d", s.config.Concurrency),
		"--format", fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best", s.config.MaxQuality),
		"--merge-output-format", "mp4",
		"--retries", "3",
		"--continue",
		"--ignore-errors",
		"--add-metadata",
		"--write-info-json",
		"--write-playlist-metafiles",
		"--output", outputPath,
		job.URL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}

	go s.trackProgress(stdout, job.ID)
	go s.trackProgress(stderr, job.ID)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if metadataPath, ok := s.getMetadataPath(job.ID); ok {
		if err := s.processMetadata(ctx, job, metadataPath); err != nil {
			log.WithError(err).Error("Failed to process metadata")
		}
		s.metadataPaths.Delete(job.ID)
	}

	job.Status = domain.JobStatusComplete
	job.Progress = 100.0

	return s.jobs.Update(&job)
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
