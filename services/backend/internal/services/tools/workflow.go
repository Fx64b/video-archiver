package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// executeWorkflow runs an ordered chain of operations, feeding each step's
// output into the next. It returns the path of the final output. Intermediate
// outputs are removed unless KeepIntermediateFiles is set.
func (s *Service) executeWorkflow(ctx context.Context, job *domain.ToolsJob, inputs []resolvedInput) (string, error) {
	params, err := parseParameters[domain.WorkflowParameters](job.Parameters)
	if err != nil {
		return "", fmt.Errorf("parse workflow parameters: %w", err)
	}
	if len(params.Steps) == 0 {
		return "", fmt.Errorf("workflow requires at least one step")
	}

	var intermediates []string
	defer func() {
		if params.KeepIntermediateFiles {
			return
		}
		for _, f := range intermediates {
			if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
				log.WithError(err).WithField("file", f).Warn("Failed to remove intermediate file")
			}
		}
	}()

	stepWeight := 100.0 / float64(len(params.Steps))
	var currentOutput string

	for i, step := range params.Steps {
		stepNum := i + 1

		stepInputs := inputs
		if i > 0 {
			stepInputs = []resolvedInput{{path: currentOutput}}
		}

		output := s.workflowStepOutput(job.ID, stepNum, step)

		base := float64(i) * stepWeight
		progress := s.workflowStepProgress(job, stepNum, len(params.Steps), base, stepWeight)

		err := s.runOperation(ctx, job, step.Operation, step.Parameters, stepInputs, output, progress)
		if err != nil {
			if ctx.Err() != nil {
				return "", err
			}
			if params.StopOnError {
				return "", fmt.Errorf("workflow step %d (%s) failed: %w", stepNum, step.Operation, err)
			}
			log.WithError(err).WithField("step", stepNum).Warn("Workflow step failed; continuing")
			continue
		}

		if currentOutput != "" {
			intermediates = append(intermediates, currentOutput)
		}
		currentOutput = output
	}

	if currentOutput == "" {
		return "", fmt.Errorf("workflow produced no output")
	}
	return currentOutput, nil
}

func (s *Service) workflowStepOutput(jobID string, stepNum int, step domain.WorkflowStep) string {
	if step.OutputName != "" {
		return filepath.Join(s.processedPath, sanitizeFilename(step.OutputName))
	}
	shortID := jobID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	ext := outputExtension(step.Operation, step.Parameters)
	name := fmt.Sprintf("workflow_%s_step%d_%s.%s", shortID, stepNum, time.Now().Format("150405"), ext)
	return filepath.Join(s.processedPath, name)
}

// workflowStepProgress scales a step's 0-100 progress into the job's overall
// progress band [base, base+weight].
func (s *Service) workflowStepProgress(job *domain.ToolsJob, stepNum, totalSteps int, base, weight float64) ProgressFunc {
	step := fmt.Sprintf("Step %d/%d", stepNum, totalSteps)
	var lastDBWrite time.Time
	return func(percent float64, elapsed time.Duration) {
		overall := base + percent*weight/100.0
		job.Progress = overall
		if time.Since(lastDBWrite) > time.Second || percent >= 100 {
			if err := s.toolsRepo.Update(job); err != nil {
				log.WithError(err).Warn("Failed to persist workflow progress")
			}
			lastDBWrite = time.Now()
		}
		s.broadcast(job, overall, step, int(elapsed.Seconds()), 0)
	}
}
