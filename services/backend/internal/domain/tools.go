package domain

import "time"

type ToolsJobStatus string

const (
	ToolsJobStatusPending    ToolsJobStatus = "pending"
	ToolsJobStatusProcessing ToolsJobStatus = "processing"
	ToolsJobStatusComplete   ToolsJobStatus = "complete"
	ToolsJobStatusFailed     ToolsJobStatus = "failed"
	ToolsJobStatusCancelled  ToolsJobStatus = "cancelled"
)

type ToolsOperationType string

const (
	OpTypeTrim         ToolsOperationType = "trim"
	OpTypeConcat       ToolsOperationType = "concat"
	OpTypeConvert      ToolsOperationType = "convert"
	OpTypeExtractAudio ToolsOperationType = "extract_audio"
	OpTypeAdjustQuality ToolsOperationType = "adjust_quality"
	OpTypeRotate       ToolsOperationType = "rotate"
	OpTypeWorkflow     ToolsOperationType = "workflow"
)

type ToolsJob struct {
	ID            string             `json:"id"`
	OperationType ToolsOperationType `json:"operation_type"`
	Status        ToolsJobStatus     `json:"status"`
	Progress      float64            `json:"progress"`        // 0-100
	InputFiles    []string           `json:"input_files"`     // Job IDs (videos, playlists, or channels)
	InputType     string             `json:"input_type"`      // "videos", "playlist", "channel"
	OutputFile    string             `json:"output_file"`     // Generated file path
	Parameters    map[string]any     `json:"parameters"`      // Operation-specific params
	ErrorMessage  string             `json:"error_message,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	EstimatedSize int64              `json:"estimated_size,omitempty"` // Bytes
	ActualSize    int64              `json:"actual_size,omitempty"`    // Bytes
}

// Operation-specific parameter structures
type TrimParameters struct {
	StartTime string `json:"start_time"` // HH:MM:SS or seconds
	EndTime   string `json:"end_time"`   // HH:MM:SS or seconds
	ReEncode  bool   `json:"re_encode"`  // Force re-encode or stream copy
}

type ConcatParameters struct {
	OutputFormat string   `json:"output_format"` // mp4, mkv, etc.
	ReEncode     bool     `json:"re_encode"`     // Force re-encode if codecs differ
	FileOrder    []string `json:"file_order"`    // Explicit ordering of input files
	SortBy       string   `json:"sort_by"`       // For channel: upload_date, title, duration
	Order        string   `json:"order"`         // asc or desc
}

type ConvertParameters struct {
	OutputFormat    string `json:"output_format"`    // mp4, webm, mkv, avi, mov
	VideoCodec      string `json:"video_codec"`      // h264, h265, vp9, etc.
	AudioCodec      string `json:"audio_codec"`      // aac, mp3, opus, etc.
	Bitrate         string `json:"bitrate"`          // e.g., "2M", "5M"
	PreserveQuality bool   `json:"preserve_quality"` // Use original quality settings
}

type ExtractAudioParameters struct {
	OutputFormat string `json:"output_format"` // mp3, aac, flac, wav, ogg
	Bitrate      string `json:"bitrate"`       // e.g., "128k", "320k"
	SampleRate   int    `json:"sample_rate"`   // e.g., 44100, 48000
}

type AdjustQualityParameters struct {
	Resolution string `json:"resolution"` // 480p, 720p, 1080p, etc.
	Bitrate    string `json:"bitrate"`    // Target bitrate
	CRF        int    `json:"crf"`        // Constant Rate Factor (0-51)
	TwoPass    bool   `json:"two_pass"`   // Use two-pass encoding
}

type RotateParameters struct {
	Rotation int  `json:"rotation"` // 90, 180, 270 degrees
	FlipH    bool `json:"flip_h"`   // Flip horizontal
	FlipV    bool `json:"flip_v"`   // Flip vertical
}

type WorkflowParameters struct {
	Steps                 []WorkflowStep `json:"steps"`                   // Ordered list of operations
	KeepIntermediateFiles bool           `json:"keep_intermediate_files"` // Save intermediate outputs
	StopOnError           bool           `json:"stop_on_error"`           // Stop workflow if step fails
}

type WorkflowStep struct {
	Operation  ToolsOperationType `json:"operation"`            // Operation type for this step
	Parameters map[string]any     `json:"parameters"`           // Parameters for this operation
	OutputName string             `json:"output_name,omitempty"` // Custom name for output
}

// Progress update for WebSocket
type ToolsProgressUpdate struct {
	JobID         string         `json:"jobID"`
	Status        ToolsJobStatus `json:"status"`
	Progress      float64        `json:"progress"`
	CurrentStep   string         `json:"current_step"`    // e.g., "Analyzing", "Encoding", "Finalizing"
	TimeElapsed   int            `json:"time_elapsed"`    // Seconds
	TimeRemaining int            `json:"time_remaining"`  // Estimated seconds
	Error         string         `json:"error,omitempty"`
}

//tygo:ignore
type ToolsRepository interface {
	Create(job *ToolsJob) error
	Update(job *ToolsJob) error
	GetByID(id string) (*ToolsJob, error)
	GetAll() ([]*ToolsJob, error)
	GetByStatus(status ToolsJobStatus) ([]*ToolsJob, error)
	Delete(id string) error
	List(page int, limit int, status string, operationType string) ([]*ToolsJob, int, error)
}
