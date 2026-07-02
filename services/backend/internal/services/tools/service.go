package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// Broadcaster pushes messages to connected WebSocket clients. The download
// service's hub satisfies this interface, letting tools progress travel over
// the same /ws connection the frontend already uses.
type Broadcaster interface {
	Broadcast(update interface{})
}

type Config struct {
	ToolsRepository domain.ToolsRepository
	JobRepository   domain.JobRepository
	Broadcaster     Broadcaster
	FFmpeg          *FFmpeg
	DownloadPath    string
	ProcessedPath   string
	Concurrency     int
}

type Service struct {
	toolsRepo     domain.ToolsRepository
	jobRepo       domain.JobRepository
	broadcaster   Broadcaster
	ffmpeg        *FFmpeg
	downloadPath  string
	processedPath string
	concurrency   int

	queue      chan *domain.ToolsJob
	activeJobs sync.Map // job ID -> context.CancelFunc
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

type resolvedInput struct {
	jobID string
	path  string
}

func NewService(config *Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	if config.ProcessedPath == "" {
		config.ProcessedPath = "./data/processed"
	}
	if config.FFmpeg == nil {
		config.FFmpeg = NewFFmpeg()
	}
	concurrency := config.Concurrency
	if concurrency < 1 {
		concurrency = 2
	}

	if err := os.MkdirAll(config.ProcessedPath, 0o755); err != nil {
		log.WithError(err).Warn("Failed to create processed directory")
	}

	return &Service{
		toolsRepo:     config.ToolsRepository,
		jobRepo:       config.JobRepository,
		broadcaster:   config.Broadcaster,
		ffmpeg:        config.FFmpeg,
		downloadPath:  config.DownloadPath,
		processedPath: config.ProcessedPath,
		concurrency:   concurrency,
		queue:         make(chan *domain.ToolsJob, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (s *Service) Start() error {
	for i := 0; i < s.concurrency; i++ {
		s.wg.Add(1)
		go s.worker()
	}
	log.WithField("concurrency", s.concurrency).Info("Tools service started")
	return nil
}

func (s *Service) Stop() {
	log.Info("Stopping tools service...")
	s.cancel()
	close(s.queue)
	s.wg.Wait()
	log.Info("Tools service stopped")
}

func (s *Service) Repository() domain.ToolsRepository { return s.toolsRepo }

func (s *Service) GetJobByID(id string) (*domain.ToolsJob, error) {
	return s.toolsRepo.GetByID(id)
}

// ResolveOutputFile returns the validated absolute path of a completed job's
// output file, guarding against any path that escapes the processed directory.
func (s *Service) ResolveOutputFile(job *domain.ToolsJob) (string, error) {
	if job.OutputFile == "" {
		return "", fmt.Errorf("job has no output file")
	}
	base := filepath.Clean(s.processedPath)
	path := filepath.Clean(job.OutputFile)
	if err := ensureWithin(base, path); err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("output file not found")
	}
	return path, nil
}

// Submit validates and enqueues a job. Validation happens here so the API can
// reject bad requests synchronously instead of failing asynchronously.
func (s *Service) Submit(job *domain.ToolsJob) error {
	if err := validateJob(job); err != nil {
		return err
	}

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.InputType == "" {
		job.InputType = domain.InputTypeVideos
	}
	now := time.Now()
	job.Status = domain.ToolsJobStatusPending
	job.Progress = 0
	job.CreatedAt = now
	job.UpdatedAt = now

	if err := s.toolsRepo.Create(job); err != nil {
		return fmt.Errorf("create tools job: %w", err)
	}

	select {
	case s.queue <- job:
	default:
		return fmt.Errorf("tools queue is full")
	}

	log.WithFields(log.Fields{"job_id": job.ID, "operation": job.OperationType}).Info("Tools job submitted")
	return nil
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
		return fmt.Errorf("cannot cancel a %s job", job.Status)
	}

	if cancel, ok := s.activeJobs.Load(id); ok {
		if fn, ok := cancel.(context.CancelFunc); ok {
			fn()
		}
	}

	job.Status = domain.ToolsJobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now
	if err := s.toolsRepo.Update(job); err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	s.broadcast(job, 0, "Cancelled", 0, 0)
	log.WithField("job_id", id).Info("Tools job cancelled")
	return nil
}

// DeleteJob removes a finished job's record and its output file from the
// processed directory. Running or queued jobs must be cancelled instead.
func (s *Service) DeleteJob(id string) error {
	job, err := s.toolsRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}
	if job.Status == domain.ToolsJobStatusPending || job.Status == domain.ToolsJobStatusProcessing {
		return fmt.Errorf("cannot delete a %s job, cancel it first", job.Status)
	}

	if path, err := s.ResolveOutputFile(job); err == nil {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.WithError(err).WithField("path", path).Warn("Failed to delete tools output file")
		}
	}

	if err := s.toolsRepo.Delete(id); err != nil {
		return fmt.Errorf("delete job record: %w", err)
	}
	log.WithField("job_id", id).Info("Tools job deleted")
	return nil
}

func (s *Service) worker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case job, ok := <-s.queue:
			if !ok {
				return
			}
			jobCtx, cancel := context.WithCancel(s.ctx)
			s.activeJobs.Store(job.ID, cancel)
			s.processJob(jobCtx, job)
			s.activeJobs.Delete(job.ID)
			cancel()
		}
	}
}

func (s *Service) processJob(ctx context.Context, job *domain.ToolsJob) {
	log.WithFields(log.Fields{"job_id": job.ID, "operation": job.OperationType}).Info("Processing tools job")

	job.Status = domain.ToolsJobStatusProcessing
	job.Progress = 0
	if err := s.toolsRepo.Update(job); err != nil {
		log.WithError(err).Error("Failed to update job status")
		return
	}
	s.broadcast(job, 0, "Starting", 0, 0)

	inputs, err := s.resolveInputs(job)
	if err != nil {
		s.failJob(job, err)
		return
	}

	var outputPath string
	if job.OperationType == domain.OpTypeWorkflow {
		outputPath, err = s.executeWorkflow(ctx, job, inputs)
	} else {
		outputPath = generateOutputPath(s.processedPath, job.OperationType, job.ID, job.Parameters)
		err = s.runOperation(ctx, job, job.OperationType, job.Parameters, inputs, outputPath, s.progressCallback(job, stepLabel(job.OperationType)))
	}

	if err != nil {
		// A cancellation is not a failure.
		if ctx.Err() != nil {
			log.WithField("job_id", job.ID).Info("Tools job cancelled mid-flight")
			return
		}
		s.failJob(job, err)
		return
	}

	job.Status = domain.ToolsJobStatusComplete
	job.Progress = 100
	job.OutputFile = outputPath
	now := time.Now()
	job.CompletedAt = &now
	if stat, statErr := os.Stat(outputPath); statErr == nil {
		job.ActualSize = stat.Size()
	}
	if err := s.toolsRepo.Update(job); err != nil {
		log.WithError(err).Error("Failed to update completed job")
		return
	}
	s.broadcast(job, 100, "Complete", 0, 0)
	log.WithFields(log.Fields{"job_id": job.ID, "output": outputPath}).Info("Tools job completed")
}

// resolveInputs expands playlist/channel jobs to their videos and resolves each
// video job ID to an on-disk file path.
func (s *Service) resolveInputs(job *domain.ToolsJob) ([]resolvedInput, error) {
	jobIDs, err := s.expandInputFiles(job)
	if err != nil {
		return nil, err
	}
	if len(jobIDs) == 0 {
		return nil, fmt.Errorf("no input videos found")
	}

	resolved := make([]resolvedInput, 0, len(jobIDs))
	for _, id := range jobIDs {
		path, err := s.resolveVideoPath(id)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, resolvedInput{jobID: id, path: path})
	}
	return resolved, nil
}

func (s *Service) expandInputFiles(job *domain.ToolsJob) ([]string, error) {
	switch job.InputType {
	case domain.InputTypeVideos, "":
		return job.InputFiles, nil
	case domain.InputTypePlaylist, domain.InputTypeChannel:
		if len(job.InputFiles) != 1 {
			return nil, fmt.Errorf("%s input requires exactly one parent ID", job.InputType)
		}
		videos, err := s.jobRepo.GetVideosForParent(job.InputFiles[0])
		if err != nil {
			return nil, fmt.Errorf("get videos for %s: %w", job.InputType, err)
		}
		ids := make([]string, 0, len(videos))
		for _, v := range videos {
			if v != nil && v.Job != nil {
				ids = append(ids, v.Job.ID)
			}
		}
		return ids, nil
	default:
		return nil, fmt.Errorf("invalid input type: %q", job.InputType)
	}
}

// resolveVideoPath maps a video job ID to the file on disk.
func (s *Service) resolveVideoPath(jobID string) (string, error) {
	jwm, err := s.jobRepo.GetJobWithMetadata(jobID)
	if err != nil {
		return "", fmt.Errorf("get job %s: %w", jobID, err)
	}
	if jwm == nil || jwm.Job == nil {
		return "", fmt.Errorf("job %s not found", jobID)
	}
	meta, ok := jwm.Metadata.(*domain.VideoMetadata)
	if !ok || meta == nil {
		return "", fmt.Errorf("job %s is not a video", jobID)
	}

	path, err := ResolveVideoFileWithHint(s.downloadPath, jwm.Job.FilePath, meta)
	if err != nil {
		return "", fmt.Errorf("job %s: %w", jobID, err)
	}
	if path != jwm.Job.FilePath {
		if err := s.jobRepo.SetFilePath(jobID, path); err != nil {
			log.WithError(err).WithField("jobID", jobID).Warn("Failed to persist resolved file path")
		}
	}
	return path, nil
}

// runOperation builds and executes ffmpeg for a single (non-workflow) operation.
func (s *Service) runOperation(ctx context.Context, job *domain.ToolsJob, op domain.ToolsOperationType, params map[string]any, inputs []resolvedInput, output string, progress ProgressFunc) error {
	opArgs, totalDuration, cleanup, err := s.prepareOperation(op, params, inputs)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return err
	}
	return s.ffmpeg.Run(ctx, opArgs, output, totalDuration, progress)
}

// prepareOperation produces the ffmpeg arguments and the expected total
// duration (for progress) for an operation. The returned cleanup must be called
// after ffmpeg finishes (it removes any temporary concat list file).
func (s *Service) prepareOperation(op domain.ToolsOperationType, params map[string]any, inputs []resolvedInput) (args []string, totalDuration float64, cleanup func(), err error) {
	if len(inputs) == 0 {
		return nil, 0, nil, fmt.Errorf("operation requires at least one input")
	}
	primary := inputs[0].path

	switch op {
	case domain.OpTypeTrim:
		p, perr := parseParameters[domain.TrimParameters](params)
		if perr != nil {
			return nil, 0, nil, perr
		}
		args, err = buildTrimArgs(primary, p)
		if err == nil {
			start, _ := parseTimecode(p.StartTime)
			end, _ := parseTimecode(p.EndTime)
			totalDuration = end - start
		}

	case domain.OpTypeConcat:
		p, perr := parseParameters[domain.ConcatParameters](params)
		if perr != nil {
			return nil, 0, nil, perr
		}
		ordered := orderConcatInputs(inputs, p.FileOrder)
		var listFile string
		listFile, cleanup, err = s.writeConcatList(ordered)
		if err != nil {
			return nil, 0, cleanup, err
		}
		args, err = buildConcatArgs(listFile, p)
		totalDuration = s.sumDurations(ordered)

	case domain.OpTypeConvert:
		p, perr := parseParameters[domain.ConvertParameters](params)
		if perr != nil {
			return nil, 0, nil, perr
		}
		args, err = buildConvertArgs(primary, p)
		totalDuration = s.probeDuration(primary)

	case domain.OpTypeExtractAudio:
		p, perr := parseParameters[domain.ExtractAudioParameters](params)
		if perr != nil {
			return nil, 0, nil, perr
		}
		args, err = buildExtractAudioArgs(primary, p)
		totalDuration = s.probeDuration(primary)

	case domain.OpTypeAdjustQuality:
		p, perr := parseParameters[domain.AdjustQualityParameters](params)
		if perr != nil {
			return nil, 0, nil, perr
		}
		args, err = buildAdjustQualityArgs(primary, p)
		totalDuration = s.probeDuration(primary)

	case domain.OpTypeRotate:
		p, perr := parseParameters[domain.RotateParameters](params)
		if perr != nil {
			return nil, 0, nil, perr
		}
		args, err = buildRotateArgs(primary, p)
		totalDuration = s.probeDuration(primary)

	default:
		return nil, 0, nil, fmt.Errorf("unsupported operation type: %q", op)
	}

	return args, totalDuration, cleanup, err
}

func (s *Service) probeDuration(path string) float64 {
	info, err := s.ffmpeg.Probe(path)
	if err != nil {
		log.WithError(err).WithField("path", path).Debug("Failed to probe duration")
		return 0
	}
	return info.Duration
}

func (s *Service) sumDurations(inputs []resolvedInput) float64 {
	var total float64
	for _, in := range inputs {
		total += s.probeDuration(in.path)
	}
	return total
}

// orderConcatInputs reorders resolved inputs by an explicit list of job IDs.
// Inputs missing from fileOrder keep their original relative order at the end.
func orderConcatInputs(inputs []resolvedInput, fileOrder []string) []resolvedInput {
	if len(fileOrder) == 0 {
		return inputs
	}
	byID := make(map[string]resolvedInput, len(inputs))
	for _, in := range inputs {
		byID[in.jobID] = in
	}
	ordered := make([]resolvedInput, 0, len(inputs))
	used := make(map[string]bool, len(inputs))
	for _, id := range fileOrder {
		if in, ok := byID[id]; ok && !used[id] {
			ordered = append(ordered, in)
			used[id] = true
		}
	}
	for _, in := range inputs {
		if !used[in.jobID] {
			ordered = append(ordered, in)
		}
	}
	return ordered
}

// writeConcatList writes a temporary ffmpeg concat demuxer list file.
func (s *Service) writeConcatList(inputs []resolvedInput) (string, func(), error) {
	listFile := filepath.Join(os.TempDir(), fmt.Sprintf("concat_%s.txt", uuid.New().String()))
	cleanup := func() { _ = os.Remove(listFile) }

	var b strings.Builder
	for _, in := range inputs {
		abs, err := filepath.Abs(in.path)
		if err != nil {
			return "", cleanup, fmt.Errorf("absolute path: %w", err)
		}
		// Escape single quotes for the concat demuxer syntax.
		escaped := strings.ReplaceAll(abs, "'", `'\''`)
		fmt.Fprintf(&b, "file '%s'\n", escaped)
	}
	if err := os.WriteFile(listFile, []byte(b.String()), 0o644); err != nil {
		return "", cleanup, fmt.Errorf("write concat list: %w", err)
	}
	return listFile, cleanup, nil
}

func (s *Service) failJob(job *domain.ToolsJob, err error) {
	log.WithError(err).WithField("job_id", job.ID).Error("Tools job failed")
	job.Status = domain.ToolsJobStatusFailed
	job.ErrorMessage = err.Error()
	now := time.Now()
	job.CompletedAt = &now
	if updateErr := s.toolsRepo.Update(job); updateErr != nil {
		log.WithError(updateErr).Error("Failed to persist failed job")
	}
	s.broadcastError(job, err.Error())
}

// progressCallback returns a ProgressFunc that throttles DB writes to once per
// second while broadcasting every update over the WebSocket.
func (s *Service) progressCallback(job *domain.ToolsJob, step string) ProgressFunc {
	var lastDBWrite time.Time
	return func(percent float64, elapsed time.Duration) {
		job.Progress = percent
		if time.Since(lastDBWrite) > time.Second || percent >= 100 {
			if err := s.toolsRepo.Update(job); err != nil {
				log.WithError(err).Warn("Failed to persist job progress")
			}
			lastDBWrite = time.Now()
		}

		remaining := 0
		if percent > 0 && percent < 100 {
			estimatedTotal := time.Duration(float64(elapsed) / (percent / 100))
			remaining = int((estimatedTotal - elapsed).Seconds())
		}
		s.broadcast(job, percent, step, int(elapsed.Seconds()), remaining)
	}
}

func (s *Service) broadcast(job *domain.ToolsJob, percent float64, step string, elapsed, remaining int) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.Broadcast(domain.ToolsProgressUpdate{
		Type:          "tools-progress",
		JobID:         job.ID,
		Status:        job.Status,
		Progress:      percent,
		CurrentStep:   step,
		TimeElapsed:   elapsed,
		TimeRemaining: remaining,
	})
}

func (s *Service) broadcastError(job *domain.ToolsJob, message string) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.Broadcast(domain.ToolsProgressUpdate{
		Type:        "tools-progress",
		JobID:       job.ID,
		Status:      job.Status,
		Progress:    job.Progress,
		CurrentStep: "Failed",
		Error:       message,
	})
}

func stepLabel(op domain.ToolsOperationType) string {
	switch op {
	case domain.OpTypeTrim:
		return "Trimming"
	case domain.OpTypeConcat:
		return "Concatenating"
	case domain.OpTypeConvert:
		return "Converting"
	case domain.OpTypeExtractAudio:
		return "Extracting audio"
	case domain.OpTypeAdjustQuality:
		return "Adjusting quality"
	case domain.OpTypeRotate:
		return "Rotating"
	default:
		return "Processing"
	}
}
