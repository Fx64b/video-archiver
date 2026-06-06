package tools

import (
	"fmt"

	"video-archiver/internal/domain"
)

// validateJob performs synchronous, ffmpeg-free validation of a submitted job
// so the API can reject malformed requests immediately.
func validateJob(job *domain.ToolsJob) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	if len(job.InputFiles) == 0 {
		return fmt.Errorf("input_files is required")
	}

	switch job.InputType {
	case domain.InputTypeVideos, domain.InputTypePlaylist, domain.InputTypeChannel, "":
	default:
		return fmt.Errorf("invalid input_type: %q", job.InputType)
	}

	// playlist/channel inputs reference a single parent that is expanded later.
	if job.InputType == domain.InputTypePlaylist || job.InputType == domain.InputTypeChannel {
		if len(job.InputFiles) != 1 {
			return fmt.Errorf("%s input requires exactly one parent ID", job.InputType)
		}
	}

	if job.OperationType == domain.OpTypeConcat &&
		(job.InputType == domain.InputTypeVideos || job.InputType == "") &&
		len(job.InputFiles) < 2 {
		return fmt.Errorf("concat requires at least two videos")
	}

	if job.OperationType == domain.OpTypeWorkflow {
		return validateWorkflowParams(job.Parameters)
	}
	return validateOperationParams(job.OperationType, job.Parameters)
}

// validateOperationParams validates the parameters of a single operation.
func validateOperationParams(op domain.ToolsOperationType, params map[string]any) error {
	switch op {
	case domain.OpTypeTrim:
		p, err := parseParameters[domain.TrimParameters](params)
		if err != nil {
			return err
		}
		return validateTrim(p)
	case domain.OpTypeConcat:
		p, err := parseParameters[domain.ConcatParameters](params)
		if err != nil {
			return err
		}
		return validateConcat(p)
	case domain.OpTypeConvert:
		p, err := parseParameters[domain.ConvertParameters](params)
		if err != nil {
			return err
		}
		return validateConvert(p)
	case domain.OpTypeExtractAudio:
		p, err := parseParameters[domain.ExtractAudioParameters](params)
		if err != nil {
			return err
		}
		return validateExtractAudio(p)
	case domain.OpTypeAdjustQuality:
		p, err := parseParameters[domain.AdjustQualityParameters](params)
		if err != nil {
			return err
		}
		return validateAdjustQuality(p)
	case domain.OpTypeRotate:
		p, err := parseParameters[domain.RotateParameters](params)
		if err != nil {
			return err
		}
		return validateRotate(p)
	case domain.OpTypeWorkflow:
		return validateWorkflowParams(params)
	default:
		return fmt.Errorf("unsupported operation type: %q", op)
	}
}

func validateWorkflowParams(params map[string]any) error {
	p, err := parseParameters[domain.WorkflowParameters](params)
	if err != nil {
		return err
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("workflow requires at least one step")
	}
	for i, step := range p.Steps {
		if step.Operation == domain.OpTypeWorkflow {
			return fmt.Errorf("workflow step %d cannot itself be a workflow", i+1)
		}
		if err := validateOperationParams(step.Operation, step.Parameters); err != nil {
			return fmt.Errorf("workflow step %d (%s): %w", i+1, step.Operation, err)
		}
	}
	return nil
}
