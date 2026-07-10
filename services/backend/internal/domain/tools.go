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
	OpTypeTrim          ToolsOperationType = "trim"
	OpTypeConcat        ToolsOperationType = "concat"
	OpTypeConvert       ToolsOperationType = "convert"
	OpTypeExtractAudio  ToolsOperationType = "extract_audio"
	OpTypeAdjustQuality ToolsOperationType = "adjust_quality"
	OpTypeRotate        ToolsOperationType = "rotate"
	OpTypeWorkflow      ToolsOperationType = "workflow"
)

// ToolsInputType describes how InputFiles should be interpreted.
type ToolsInputType string

const (
	InputTypeVideos     ToolsInputType = "videos"
	InputTypePlaylist   ToolsInputType = "playlist"
	InputTypeChannel    ToolsInputType = "channel"
	InputTypeCollection ToolsInputType = "collection"
)

// ToolsJob represents a single media-processing job operating on already
// downloaded videos. InputFiles holds download job IDs (videos, or a single
// playlist/channel/collection that is expanded into its videos).
type ToolsJob struct {
	ID            string             `json:"id"`
	OperationType ToolsOperationType `json:"operation_type"`
	Status        ToolsJobStatus     `json:"status"`
	Progress      float64            `json:"progress"` // 0-100
	InputFiles    []string           `json:"input_files"`
	InputType     ToolsInputType     `json:"input_type"`
	OutputFile    string             `json:"output_file"`
	Parameters    map[string]any     `json:"parameters"`
	ErrorMessage  string             `json:"error_message,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	EstimatedSize int64              `json:"estimated_size,omitempty"`
	ActualSize    int64              `json:"actual_size,omitempty"`
	// Probed from the produced file on completion; zero values on legacy rows
	// or when ffprobe failed.
	MediaKind  string  `json:"media_kind,omitempty"` // MediaKindVideo or MediaKindAudio
	Duration   float64 `json:"duration,omitempty"`   // seconds
	Width      int     `json:"width,omitempty"`
	Height     int     `json:"height,omitempty"`
	VideoCodec string  `json:"video_codec,omitempty"`
	AudioCodec string  `json:"audio_codec,omitempty"`
}

// Media kind of a produced output file, probed after the job completes.
const (
	MediaKindVideo = "video"
	MediaKindAudio = "audio"
)

// Operation-specific parameter structures. These mirror the JSON the frontend
// submits under the "parameters" key for each operation.

type TrimParameters struct {
	StartTime string `json:"start_time"` // HH:MM:SS(.ms) or seconds
	EndTime   string `json:"end_time"`   // HH:MM:SS(.ms) or seconds
	ReEncode  bool   `json:"re_encode"`  // Re-encode for frame-accurate cuts
}

type ConcatParameters struct {
	OutputFormat string   `json:"output_format"` // mp4, mkv, webm
	ReEncode     bool     `json:"re_encode"`     // Re-encode if codecs differ
	FileOrder    []string `json:"file_order"`    // Explicit ordering by job ID
}

type ConvertParameters struct {
	OutputFormat string `json:"output_format"` // mp4, webm, mkv, avi, mov
	VideoCodec   string `json:"video_codec"`   // libx264, libx265, libvpx-vp9, copy
	AudioCodec   string `json:"audio_codec"`   // aac, libmp3lame, libopus, copy
	Bitrate      string `json:"bitrate"`       // e.g. "2M", "5M"
}

type ExtractAudioParameters struct {
	OutputFormat string `json:"output_format"` // mp3, aac, flac, wav, ogg
	Bitrate      string `json:"bitrate"`       // e.g. "128k", "320k"
	SampleRate   int    `json:"sample_rate"`   // e.g. 44100, 48000
}

type AdjustQualityParameters struct {
	Resolution string `json:"resolution"` // 480p, 720p, 1080p, ...
	Bitrate    string `json:"bitrate"`    // optional target bitrate
	CRF        int    `json:"crf"`        // Constant Rate Factor (0-51)
	TwoPass    bool   `json:"two_pass"`   // Two-pass encoding (requires bitrate)
}

type RotateParameters struct {
	Rotation int  `json:"rotation"` // 0, 90, 180, 270
	FlipH    bool `json:"flip_h"`   // Horizontal flip
	FlipV    bool `json:"flip_v"`   // Vertical flip
}

type WorkflowParameters struct {
	Steps                 []WorkflowStep `json:"steps"`
	KeepIntermediateFiles bool           `json:"keep_intermediate_files"`
	StopOnError           bool           `json:"stop_on_error"`
}

type WorkflowStep struct {
	Operation  ToolsOperationType `json:"operation"`
	Parameters map[string]any     `json:"parameters"`
	OutputName string             `json:"output_name,omitempty"`
}

// ToolsProgressUpdate is broadcast over the WebSocket as a job runs. The Type
// field lets the frontend route the message without fragile field sniffing.
type ToolsProgressUpdate struct {
	Type          string         `json:"type"` // always "tools-progress"
	JobID         string         `json:"jobID"`
	Status        ToolsJobStatus `json:"status"`
	Progress      float64        `json:"progress"`
	CurrentStep   string         `json:"current_step"`
	TimeElapsed   int            `json:"time_elapsed"`   // seconds
	TimeRemaining int            `json:"time_remaining"` // estimated seconds
	Error         string         `json:"error,omitempty"`
}

//tygo:ignore
type ToolsRepository interface {
	Create(job *ToolsJob) error
	Update(job *ToolsJob) error
	GetByID(id string) (*ToolsJob, error)
	GetAll() ([]*ToolsJob, error)
	GetByStatus(status ToolsJobStatus) ([]*ToolsJob, error)
	FindLatestConvertForInput(jobID string) (*ToolsJob, error)
	Delete(id string) error
	List(page int, limit int, status string, operationType string) ([]*ToolsJob, int, error)
}

// PlaybackInfo describes whether a downloaded video's codecs can be decoded by
// browsers, and any transcode job that produces a compatible version.
type PlaybackInfo struct {
	Container   string             `json:"container"`
	VideoCodec  string             `json:"video_codec"`
	AudioCodec  string             `json:"audio_codec"`
	BrowserSafe bool               `json:"browser_safe"`
	Transcode   *PlaybackTranscode `json:"transcode,omitempty"`
}

// PlaybackTranscode is the state of the convert job backing a browser-safe
// version of a video.
type PlaybackTranscode struct {
	JobID    string         `json:"job_id"`
	Status   ToolsJobStatus `json:"status"`
	Progress float64        `json:"progress"`
}
