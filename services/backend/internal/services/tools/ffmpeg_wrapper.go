package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type FFmpegWrapper struct {
	ffmpegPath  string
	ffprobePath string
}

type VideoInfo struct {
	Duration  float64
	Width     int
	Height    int
	Bitrate   int64
	Codec     string
	AudioCodec string
	FileSize  int64
}

type ProgressCallback func(progress float64, timeElapsed time.Duration)

func NewFFmpegWrapper() *FFmpegWrapper {
	return &FFmpegWrapper{
		ffmpegPath:  "/usr/bin/ffmpeg",
		ffprobePath: "/usr/bin/ffprobe",
	}
}

// GetMetadata retrieves video metadata using ffprobe
func (f *FFmpegWrapper) GetMetadata(inputPath string) (*VideoInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	}

	cmd := exec.Command(f.ffprobePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(output))
	}

	// Parse the JSON output
	// For now, return basic info - can be expanded later
	info := &VideoInfo{}

	// Simple duration extraction using regex (can be improved with proper JSON parsing)
	durationRegex := regexp.MustCompile(`"duration":\s*"([^"]+)"`)
	if matches := durationRegex.FindStringSubmatch(string(output)); len(matches) > 1 {
		if duration, err := strconv.ParseFloat(matches[1], 64); err == nil {
			info.Duration = duration
		}
	}

	return info, nil
}

// Trim cuts a video by specifying start and end timestamps
func (f *FFmpegWrapper) Trim(ctx context.Context, input, output, start, end string, reEncode bool, callback ProgressCallback) error {
	args := []string{
		"-i", input,
		"-ss", start,
		"-to", end,
	}

	if !reEncode {
		// Stream copy (fast, no quality loss)
		args = append(args, "-c", "copy")
	} else {
		// Re-encode with high quality
		args = append(args, "-c:v", "libx264", "-crf", "18", "-c:a", "aac", "-b:a", "192k")
	}

	args = append(args, "-progress", "pipe:1", "-y", output)
	return f.executeWithProgress(ctx, args, callback)
}

// Concat merges multiple videos into a single file
func (f *FFmpegWrapper) Concat(ctx context.Context, inputs []string, output string, reEncode bool, callback ProgressCallback) error {
	// Create temporary file list for concat demuxer
	listFile := filepath.Join(os.TempDir(), fmt.Sprintf("concat_list_%d.txt", time.Now().UnixNano()))
	defer os.Remove(listFile)

	file, err := os.Create(listFile)
	if err != nil {
		return fmt.Errorf("create concat list: %w", err)
	}

	for _, input := range inputs {
		// Make paths absolute
		absPath, err := filepath.Abs(input)
		if err != nil {
			file.Close()
			return fmt.Errorf("get absolute path: %w", err)
		}
		fmt.Fprintf(file, "file '%s'\n", absPath)
	}
	file.Close()

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
	}

	if !reEncode {
		args = append(args, "-c", "copy")
	} else {
		args = append(args, "-c:v", "libx264", "-crf", "18", "-c:a", "aac", "-b:a", "192k")
	}

	args = append(args, "-progress", "pipe:1", "-y", output)
	return f.executeWithProgress(ctx, args, callback)
}

// Convert changes video format/codec
func (f *FFmpegWrapper) Convert(ctx context.Context, input, output string, videoCodec, audioCodec, bitrate string, callback ProgressCallback) error {
	args := []string{
		"-i", input,
	}

	if videoCodec != "" {
		args = append(args, "-c:v", videoCodec)
	}

	if audioCodec != "" {
		args = append(args, "-c:a", audioCodec)
	}

	if bitrate != "" {
		args = append(args, "-b:v", bitrate)
	}

	args = append(args, "-progress", "pipe:1", "-y", output)
	return f.executeWithProgress(ctx, args, callback)
}

// ExtractAudio extracts audio track from video
func (f *FFmpegWrapper) ExtractAudio(ctx context.Context, input, output, format, bitrate string, sampleRate int, callback ProgressCallback) error {
	args := []string{
		"-i", input,
		"-vn", // No video
	}

	// Set audio codec based on format
	switch format {
	case "mp3":
		args = append(args, "-acodec", "libmp3lame")
	case "aac":
		args = append(args, "-acodec", "aac")
	case "flac":
		args = append(args, "-acodec", "flac")
	case "wav":
		args = append(args, "-acodec", "pcm_s16le")
	case "ogg":
		args = append(args, "-acodec", "libvorbis")
	default:
		return fmt.Errorf("unsupported audio format: %s", format)
	}

	if bitrate != "" {
		args = append(args, "-b:a", bitrate)
	}

	if sampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", sampleRate))
	}

	args = append(args, "-progress", "pipe:1", "-y", output)
	return f.executeWithProgress(ctx, args, callback)
}

// AdjustQuality changes video resolution and/or bitrate
func (f *FFmpegWrapper) AdjustQuality(ctx context.Context, input, output, resolution, bitrate string, crf int, twoPass bool, callback ProgressCallback) error {
	args := []string{
		"-i", input,
	}

	// Parse resolution (e.g., "1080p" -> height 1080, maintain aspect ratio)
	if resolution != "" {
		resolutionNum := strings.TrimSuffix(resolution, "p")
		args = append(args, "-vf", fmt.Sprintf("scale=-2:%s", resolutionNum))
	}

	args = append(args, "-c:v", "libx264")

	if crf > 0 {
		args = append(args, "-crf", fmt.Sprintf("%d", crf))
	}

	if bitrate != "" {
		args = append(args, "-b:v", bitrate)
	}

	args = append(args, "-c:a", "aac", "-b:a", "192k")

	if twoPass {
		// Two-pass encoding (first pass)
		passArgs := append(args, "-pass", "1", "-f", "null", "/dev/null")
		if err := f.executeWithProgress(ctx, passArgs, func(p float64, t time.Duration) {
			// First pass is 50% of total progress
			if callback != nil {
				callback(p/2, t)
			}
		}); err != nil {
			return fmt.Errorf("two-pass first pass failed: %w", err)
		}

		// Second pass
		args = append(args, "-pass", "2")
	}

	args = append(args, "-progress", "pipe:1", "-y", output)
	return f.executeWithProgress(ctx, args, func(p float64, t time.Duration) {
		if twoPass && callback != nil {
			// Second pass is the remaining 50%
			callback(50+p/2, t)
		} else if callback != nil {
			callback(p, t)
		}
	})
}

// Rotate rotates or flips video
func (f *FFmpegWrapper) Rotate(ctx context.Context, input, output string, rotation int, flipH, flipV bool, callback ProgressCallback) error {
	args := []string{
		"-i", input,
	}

	var filters []string

	// Rotation
	switch rotation {
	case 90:
		filters = append(filters, "transpose=1")
	case 180:
		filters = append(filters, "transpose=1,transpose=1")
	case 270:
		filters = append(filters, "transpose=2")
	}

	// Flips
	if flipH {
		filters = append(filters, "hflip")
	}
	if flipV {
		filters = append(filters, "vflip")
	}

	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args, "-c:a", "copy", "-progress", "pipe:1", "-y", output)
	return f.executeWithProgress(ctx, args, callback)
}

// executeWithProgress runs ffmpeg command and tracks progress
func (f *FFmpegWrapper) executeWithProgress(ctx context.Context, args []string, callback ProgressCallback) error {
	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	log.WithField("args", args).Debug("Executing ffmpeg command")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	startTime := time.Now()
	var duration float64
	progressRegex := regexp.MustCompile(`out_time_us=(\d+)`)
	durationRegex := regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+)\.(\d+)`)

	// Read stderr for duration
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := durationRegex.FindStringSubmatch(line); len(matches) > 4 {
				hours, _ := strconv.Atoi(matches[1])
				minutes, _ := strconv.Atoi(matches[2])
				seconds, _ := strconv.Atoi(matches[3])
				duration = float64(hours*3600 + minutes*60 + seconds)
			}
			log.Debug(line)
		}
	}()

	// Read stdout for progress
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
			microseconds, _ := strconv.ParseInt(matches[1], 10, 64)
			currentTime := float64(microseconds) / 1000000.0

			var progress float64
			if duration > 0 {
				progress = (currentTime / duration) * 100.0
				if progress > 100 {
					progress = 100
				}
			}

			if callback != nil {
				callback(progress, time.Since(startTime))
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	// Final progress callback
	if callback != nil {
		callback(100, time.Since(startTime))
	}

	return nil
}
