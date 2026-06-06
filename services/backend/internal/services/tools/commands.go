package tools

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"video-archiver/internal/domain"
)

// This file contains the pure, side-effect-free logic for the tools service:
// parameter parsing, validation, and ffmpeg argument construction. Keeping it
// free of process execution and filesystem access makes it fully unit testable
// without ffmpeg installed.

// Supported output formats per operation. Used for validation and for the
// operations listing exposed to the frontend.
var (
	videoContainerFormats = []string{"mp4", "mkv", "webm", "avi", "mov"}
	concatFormats         = []string{"mp4", "mkv", "webm"}
	audioFormats          = []string{"mp3", "aac", "flac", "wav", "ogg"}
)

// parseParameters decodes a job's generic parameter map into a typed struct.
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

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// parseTimecode converts a timestamp into seconds. It accepts plain seconds
// ("12", "12.5"), "MM:SS", "HH:MM:SS" and optional fractional seconds.
func parseTimecode(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty timecode")
	}

	parts := strings.Split(s, ":")
	if len(parts) > 3 {
		return 0, fmt.Errorf("invalid timecode %q", s)
	}

	var total float64
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid timecode %q", s)
		}
		if v < 0 {
			return 0, fmt.Errorf("invalid timecode %q", s)
		}
		total = total*60 + v
	}
	return total, nil
}

// audioCodecForFormat returns the ffmpeg encoder for an audio container format.
func audioCodecForFormat(format string) (string, error) {
	switch format {
	case "mp3":
		return "libmp3lame", nil
	case "aac":
		return "aac", nil
	case "flac":
		return "flac", nil
	case "wav":
		return "pcm_s16le", nil
	case "ogg":
		return "libvorbis", nil
	default:
		return "", fmt.Errorf("unsupported audio format: %q", format)
	}
}

// rotationFilters builds the ffmpeg -vf filter chain for a rotate/flip op.
func rotationFilters(rotation int, flipH, flipV bool) ([]string, error) {
	var filters []string
	switch rotation {
	case 0:
		// no rotation
	case 90:
		filters = append(filters, "transpose=1")
	case 180:
		filters = append(filters, "transpose=2,transpose=2")
	case 270:
		filters = append(filters, "transpose=2")
	default:
		return nil, fmt.Errorf("unsupported rotation: %d (must be 0, 90, 180 or 270)", rotation)
	}

	if flipH {
		filters = append(filters, "hflip")
	}
	if flipV {
		filters = append(filters, "vflip")
	}
	return filters, nil
}

// --- Validation ---------------------------------------------------------

func validateTrim(p *domain.TrimParameters) error {
	start, err := parseTimecode(p.StartTime)
	if err != nil {
		return fmt.Errorf("start_time: %w", err)
	}
	end, err := parseTimecode(p.EndTime)
	if err != nil {
		return fmt.Errorf("end_time: %w", err)
	}
	if end <= start {
		return fmt.Errorf("end_time must be greater than start_time")
	}
	return nil
}

func validateConcat(p *domain.ConcatParameters) error {
	if p.OutputFormat != "" && !contains(concatFormats, p.OutputFormat) {
		return fmt.Errorf("unsupported output_format %q for concat", p.OutputFormat)
	}
	return nil
}

func validateConvert(p *domain.ConvertParameters) error {
	if p.OutputFormat == "" {
		return fmt.Errorf("output_format is required")
	}
	if !contains(videoContainerFormats, p.OutputFormat) {
		return fmt.Errorf("unsupported output_format %q for convert", p.OutputFormat)
	}
	return nil
}

func validateExtractAudio(p *domain.ExtractAudioParameters) error {
	if p.OutputFormat == "" {
		return fmt.Errorf("output_format is required")
	}
	if _, err := audioCodecForFormat(p.OutputFormat); err != nil {
		return err
	}
	if p.SampleRate < 0 {
		return fmt.Errorf("sample_rate must not be negative")
	}
	return nil
}

func validateAdjustQuality(p *domain.AdjustQualityParameters) error {
	if p.Resolution != "" {
		if _, err := parseResolutionHeight(p.Resolution); err != nil {
			return err
		}
	}
	if p.CRF < 0 || p.CRF > 51 {
		return fmt.Errorf("crf must be between 0 and 51")
	}
	if p.TwoPass && p.Bitrate == "" {
		return fmt.Errorf("two_pass requires a target bitrate")
	}
	if p.Resolution == "" && p.Bitrate == "" && p.CRF == 0 {
		return fmt.Errorf("at least one of resolution, bitrate or crf must be set")
	}
	return nil
}

func validateRotate(p *domain.RotateParameters) error {
	filters, err := rotationFilters(p.Rotation, p.FlipH, p.FlipV)
	if err != nil {
		return err
	}
	if len(filters) == 0 {
		return fmt.Errorf("rotate requires a rotation or a flip")
	}
	return nil
}

// parseResolutionHeight extracts the target height from a resolution label such
// as "1080p" or "720".
func parseResolutionHeight(resolution string) (int, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(resolution), "p")
	height, err := strconv.Atoi(trimmed)
	if err != nil || height <= 0 {
		return 0, fmt.Errorf("invalid resolution %q", resolution)
	}
	return height, nil
}

// --- Argument builders --------------------------------------------------
//
// Each builder returns the operation-specific portion of the ffmpeg argument
// vector: everything that must appear before the universal progress flags and
// the output path. The FFmpeg runner appends "-progress pipe:1 -nostats -y
// <output>". Builders never touch the filesystem or process state.

func buildTrimArgs(input string, p *domain.TrimParameters) ([]string, error) {
	if err := validateTrim(p); err != nil {
		return nil, err
	}

	args := []string{"-i", input, "-ss", p.StartTime, "-to", p.EndTime}
	if p.ReEncode {
		args = append(args, "-c:v", "libx264", "-crf", "18", "-c:a", "aac", "-b:a", "192k")
	} else {
		// Stream copy: fast, lossless, cuts at nearest keyframe.
		args = append(args, "-c", "copy")
	}
	return args, nil
}

func buildConcatArgs(listFile string, p *domain.ConcatParameters) ([]string, error) {
	if err := validateConcat(p); err != nil {
		return nil, err
	}

	args := []string{"-f", "concat", "-safe", "0", "-i", listFile}
	if p.ReEncode {
		args = append(args, "-c:v", "libx264", "-crf", "18", "-c:a", "aac", "-b:a", "192k")
	} else {
		args = append(args, "-c", "copy")
	}
	return args, nil
}

func buildConvertArgs(input string, p *domain.ConvertParameters) ([]string, error) {
	if err := validateConvert(p); err != nil {
		return nil, err
	}

	args := []string{"-i", input}

	videoCodec := p.VideoCodec
	if videoCodec == "" {
		videoCodec = "libx264"
	}
	args = append(args, "-c:v", videoCodec)

	audioCodec := p.AudioCodec
	if audioCodec == "" {
		audioCodec = "aac"
	}
	args = append(args, "-c:a", audioCodec)

	if p.Bitrate != "" {
		args = append(args, "-b:v", p.Bitrate)
	}
	return args, nil
}

func buildExtractAudioArgs(input string, p *domain.ExtractAudioParameters) ([]string, error) {
	if err := validateExtractAudio(p); err != nil {
		return nil, err
	}

	codec, _ := audioCodecForFormat(p.OutputFormat)
	args := []string{"-i", input, "-vn", "-acodec", codec}

	if p.Bitrate != "" {
		args = append(args, "-b:a", p.Bitrate)
	}
	if p.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(p.SampleRate))
	}
	return args, nil
}

func buildAdjustQualityArgs(input string, p *domain.AdjustQualityParameters) ([]string, error) {
	if err := validateAdjustQuality(p); err != nil {
		return nil, err
	}

	args := []string{"-i", input}
	if p.Resolution != "" {
		height, _ := parseResolutionHeight(p.Resolution)
		// scale=-2:<height> keeps the aspect ratio and an even width.
		args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", height))
	}

	args = append(args, "-c:v", "libx264")
	if p.CRF > 0 {
		args = append(args, "-crf", strconv.Itoa(p.CRF))
	}
	if p.Bitrate != "" {
		args = append(args, "-b:v", p.Bitrate)
	}
	args = append(args, "-c:a", "aac", "-b:a", "192k")
	return args, nil
}

func buildRotateArgs(input string, p *domain.RotateParameters) ([]string, error) {
	filters, err := rotationFilters(p.Rotation, p.FlipH, p.FlipV)
	if err != nil {
		return nil, err
	}
	if len(filters) == 0 {
		return nil, fmt.Errorf("rotate requires a rotation or a flip")
	}

	return []string{
		"-i", input,
		"-vf", strings.Join(filters, ","),
		// Rotation re-encodes video; copy audio untouched.
		"-c:a", "copy",
	}, nil
}
