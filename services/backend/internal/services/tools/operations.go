package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// createThrottledCallback creates a progress callback that throttles database updates
// to at most once per second, while still sending WebSocket updates on every call
func (s *Service) createThrottledCallback(job *domain.ToolsJob, stepDescription string) func(float64, time.Duration) {
	var lastUpdate time.Time

	return func(progress float64, elapsed time.Duration) {
		job.Progress = progress

		// Only update database if >1 second since last update or at 100%
		if time.Since(lastUpdate) > time.Second || progress >= 100 {
			if err := s.toolsRepo.Update(job); err != nil {
				log.WithError(err).Warn("Failed to update job progress")
			}
			lastUpdate = time.Now()
		}

		// Calculate time estimates
		estimatedTotal := time.Duration(0)
		if progress > 0 {
			estimatedTotal = time.Duration(float64(elapsed) / (progress / 100))
		}
		remaining := estimatedTotal - elapsed

		// Send WebSocket update on every callback (lightweight)
		s.sendProgress(job.ID, job.Status, progress, stepDescription,
			int(elapsed.Seconds()), int(remaining.Seconds()))
	}
}

// executeTrim handles video trimming operation
func (s *Service) executeTrim(job *domain.ToolsJob, inputPath, outputPath string) error {
	params, err := parseParameters[domain.TrimParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse trim parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("input", inputPath).WithField("output", outputPath).
		WithField("start", params.StartTime).WithField("end", params.EndTime).Info("Executing trim operation")

	startTime := time.Now()
	callback := s.createThrottledCallback(job, "Trimming video")

	err = s.ffmpeg.Trim(s.ctx, inputPath, outputPath, params.StartTime, params.EndTime, params.ReEncode, callback)
	if err != nil {
		return fmt.Errorf("ffmpeg trim: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Trim operation completed")
	return nil
}

// executeConcat handles video concatenation operation
func (s *Service) executeConcat(job *domain.ToolsJob, inputPaths []string, outputPath string) error {
	params, err := parseParameters[domain.ConcatParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse concat parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("input_count", len(inputPaths)).WithField("output", outputPath).
		Info("Executing concat operation")

	// If file_order is specified, reorder inputs
	if len(params.FileOrder) > 0 {
		// Reorder inputPaths based on FileOrder (job IDs)
		// This is a simplified version - in production would need proper mapping
		log.WithField("file_order", params.FileOrder).Debug("Using specified file order")
	}

	startTime := time.Now()
	callback := s.createThrottledCallback(job, fmt.Sprintf("Concatenating %d videos", len(inputPaths)))

	err = s.ffmpeg.Concat(s.ctx, inputPaths, outputPath, params.ReEncode, callback)
	if err != nil {
		return fmt.Errorf("ffmpeg concat: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Concat operation completed")
	return nil
}

// executeConvert handles format conversion operation
func (s *Service) executeConvert(job *domain.ToolsJob, inputPath, outputPath string) error {
	params, err := parseParameters[domain.ConvertParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse convert parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("input", inputPath).WithField("output", outputPath).
		WithField("format", params.OutputFormat).Info("Executing convert operation")

	startTime := time.Now()
	callback := s.createThrottledCallback(job, fmt.Sprintf("Converting to %s", params.OutputFormat))

	err = s.ffmpeg.Convert(s.ctx, inputPath, outputPath, params.VideoCodec, params.AudioCodec, params.Bitrate, callback)
	if err != nil {
		return fmt.Errorf("ffmpeg convert: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Convert operation completed")
	return nil
}

// executeExtractAudio handles audio extraction operation
func (s *Service) executeExtractAudio(job *domain.ToolsJob, inputPath, outputPath string) error {
	params, err := parseParameters[domain.ExtractAudioParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse extract audio parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("input", inputPath).WithField("output", outputPath).
		WithField("format", params.OutputFormat).WithField("bitrate", params.Bitrate).Info("Executing extract audio operation")

	startTime := time.Now()
	callback := s.createThrottledCallback(job, fmt.Sprintf("Extracting audio to %s", params.OutputFormat))

	err = s.ffmpeg.ExtractAudio(s.ctx, inputPath, outputPath, params.OutputFormat, params.Bitrate, params.SampleRate, callback)
	if err != nil {
		return fmt.Errorf("ffmpeg extract audio: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Extract audio operation completed")
	return nil
}

// executeAdjustQuality handles quality/resolution adjustment operation
func (s *Service) executeAdjustQuality(job *domain.ToolsJob, inputPath, outputPath string) error {
	params, err := parseParameters[domain.AdjustQualityParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse adjust quality parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("input", inputPath).WithField("output", outputPath).
		WithField("resolution", params.Resolution).WithField("crf", params.CRF).Info("Executing adjust quality operation")

	startTime := time.Now()
	// Custom throttled callback for adjust quality with dynamic step descriptions
	var lastUpdate time.Time
	callback := func(progress float64, elapsed time.Duration) {
		job.Progress = progress

		// Throttle database updates to at most once per second
		if time.Since(lastUpdate) > time.Second || progress >= 100 {
			if err := s.toolsRepo.Update(job); err != nil {
				log.WithError(err).Warn("Failed to update job progress")
			}
			lastUpdate = time.Now()
		}

		estimatedTotal := time.Duration(0)
		if progress > 0 {
			estimatedTotal = time.Duration(float64(elapsed) / (progress / 100))
		}
		remaining := estimatedTotal - elapsed

		step := "Adjusting quality"
		if params.TwoPass {
			if progress < 50 {
				step = "Pass 1/2: Analyzing"
			} else {
				step = "Pass 2/2: Encoding"
			}
		}

		s.sendProgress(job.ID, job.Status, progress, step, int(elapsed.Seconds()), int(remaining.Seconds()))
	}

	err = s.ffmpeg.AdjustQuality(s.ctx, inputPath, outputPath, params.Resolution, params.Bitrate, params.CRF, params.TwoPass, callback)
	if err != nil {
		return fmt.Errorf("ffmpeg adjust quality: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Adjust quality operation completed")
	return nil
}

// executeRotate handles video rotation/flip operation
func (s *Service) executeRotate(job *domain.ToolsJob, inputPath, outputPath string) error {
	params, err := parseParameters[domain.RotateParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse rotate parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("input", inputPath).WithField("output", outputPath).
		WithField("rotation", params.Rotation).WithField("flip_h", params.FlipH).WithField("flip_v", params.FlipV).
		Info("Executing rotate operation")

	startTime := time.Now()
	callback := s.createThrottledCallback(job, "Rotating video")

	err = s.ffmpeg.Rotate(s.ctx, inputPath, outputPath, params.Rotation, params.FlipH, params.FlipV, callback)
	if err != nil {
		return fmt.Errorf("ffmpeg rotate: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Rotate operation completed")
	return nil
}

// executeWorkflow handles multi-step workflow execution
func (s *Service) executeWorkflow(job *domain.ToolsJob, inputPaths []string) error {
	params, err := parseParameters[domain.WorkflowParameters](job.Parameters)
	if err != nil {
		return fmt.Errorf("parse workflow parameters: %w", err)
	}

	log.WithField("job_id", job.ID).WithField("step_count", len(params.Steps)).Info("Executing workflow")

	var currentOutput string
	var intermediateFiles []string // Track all intermediate files for cleanup
	startTime := time.Now()

	// Defer cleanup of intermediate files if requested
	defer func() {
		if !params.KeepIntermediateFiles && len(intermediateFiles) > 0 {
			for _, file := range intermediateFiles {
				if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
					log.WithError(err).WithField("file", file).Warn("Failed to cleanup intermediate file")
				} else {
					log.WithField("file", file).Debug("Cleaned up intermediate file")
				}
			}
		}
	}()

	for i, step := range params.Steps {
		stepNum := i + 1
		log.WithField("job_id", job.ID).WithField("step", stepNum).WithField("operation", step.Operation).
			Info("Executing workflow step")

		// Calculate step progress (each step gets equal weight)
		stepProgressWeight := 100.0 / float64(len(params.Steps))
		baseProgress := float64(i) * stepProgressWeight

		// Determine input for this step
		var stepInput string
		var stepInputs []string

		if i == 0 {
			// First step uses original inputs
			if len(inputPaths) == 1 {
				stepInput = inputPaths[0]
			}
			stepInputs = inputPaths
		} else {
			// Subsequent steps use output from previous step
			stepInput = currentOutput
			stepInputs = []string{currentOutput}
		}

		// Generate output path for this step
		var stepOutput string
		if step.OutputName != "" {
			stepOutput = filepath.Join(s.processedPath, step.OutputName)
		} else {
			timestamp := time.Now().Format("20060102_150405")
			stepOutput = filepath.Join(s.processedPath, fmt.Sprintf("workflow_%s_step%d_%s", job.ID[:8], stepNum, timestamp))

			// Add appropriate extension
			if step.Operation == domain.OpTypeExtractAudio {
				if format, ok := step.Parameters["output_format"].(string); ok {
					stepOutput += "." + format
				} else {
					stepOutput += ".mp3"
				}
			} else {
				stepOutput += ".mp4"
			}
		}

		// Create a temporary job for this step (for progress tracking)
		stepJob := &domain.ToolsJob{
			ID:            job.ID,
			OperationType: step.Operation,
			Status:        domain.ToolsJobStatusProcessing,
			Parameters:    step.Parameters,
		}

		// Execute the step
		var stepErr error
		stepCallback := func(progress float64, elapsed time.Duration) {
			// Scale progress to this step's portion
			overallProgress := baseProgress + (progress * stepProgressWeight / 100.0)
			job.Progress = overallProgress

			if err := s.toolsRepo.Update(job); err != nil {
				log.WithError(err).Warn("Failed to update job progress")
			}

			s.sendProgress(job.ID, job.Status, overallProgress,
				fmt.Sprintf("Step %d/%d: %s", stepNum, len(params.Steps), step.Operation),
				int(time.Since(startTime).Seconds()), 0)
		}

		switch step.Operation {
		case domain.OpTypeTrim:
			trimParams, _ := parseParameters[domain.TrimParameters](step.Parameters)
			stepErr = s.ffmpeg.Trim(s.ctx, stepInput, stepOutput, trimParams.StartTime, trimParams.EndTime, trimParams.ReEncode, stepCallback)

		case domain.OpTypeConcat:
			concatParams, _ := parseParameters[domain.ConcatParameters](step.Parameters)
			stepErr = s.ffmpeg.Concat(s.ctx, stepInputs, stepOutput, concatParams.ReEncode, stepCallback)

		case domain.OpTypeConvert:
			convertParams, _ := parseParameters[domain.ConvertParameters](step.Parameters)
			stepErr = s.ffmpeg.Convert(s.ctx, stepInput, stepOutput, convertParams.VideoCodec, convertParams.AudioCodec, convertParams.Bitrate, stepCallback)

		case domain.OpTypeExtractAudio:
			audioParams, _ := parseParameters[domain.ExtractAudioParameters](step.Parameters)
			stepErr = s.ffmpeg.ExtractAudio(s.ctx, stepInput, stepOutput, audioParams.OutputFormat, audioParams.Bitrate, audioParams.SampleRate, stepCallback)

		case domain.OpTypeAdjustQuality:
			qualityParams, _ := parseParameters[domain.AdjustQualityParameters](step.Parameters)
			stepErr = s.ffmpeg.AdjustQuality(s.ctx, stepInput, stepOutput, qualityParams.Resolution, qualityParams.Bitrate, qualityParams.CRF, qualityParams.TwoPass, stepCallback)

		case domain.OpTypeRotate:
			rotateParams, _ := parseParameters[domain.RotateParameters](step.Parameters)
			stepErr = s.ffmpeg.Rotate(s.ctx, stepInput, stepOutput, rotateParams.Rotation, rotateParams.FlipH, rotateParams.FlipV, stepCallback)

		default:
			stepErr = fmt.Errorf("unsupported workflow step operation: %s", step.Operation)
		}

		if stepErr != nil {
			if params.StopOnError {
				return fmt.Errorf("workflow step %d failed: %w", stepNum, stepErr)
			}
			log.WithError(stepErr).WithField("step", stepNum).Warn("Workflow step failed, continuing")
		}

		// Update current output for next step
		currentOutput = stepOutput

		// Track intermediate files (all except the final output)
		// Skip the first step's output if we're not keeping intermediates, as it will be cleaned later
		if i < len(params.Steps)-1 {
			intermediateFiles = append(intermediateFiles, stepOutput)
		}
	}

	// Set the final output file
	job.OutputFile = currentOutput

	log.WithField("job_id", job.ID).WithField("duration", time.Since(startTime)).Info("Workflow completed")
	return nil
}
