package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// FFmpeg wraps the ffmpeg/ffprobe binaries. Binary locations can be overridden
// with the FFMPEG_PATH / FFPROBE_PATH environment variables, otherwise they are
// resolved from PATH.
type FFmpeg struct {
	ffmpegPath  string
	ffprobePath string
}

// ProgressFunc receives a 0-100 percentage and the elapsed wall-clock time.
type ProgressFunc func(percent float64, elapsed time.Duration)

// MediaInfo holds the subset of ffprobe output the tools service needs.
type MediaInfo struct {
	Duration   float64
	Width      int
	Height     int
	Bitrate    int64
	VideoCodec string
	AudioCodec string
	HasVideo   bool
	HasAudio   bool
}

func NewFFmpeg() *FFmpeg {
	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	ffprobePath := os.Getenv("FFPROBE_PATH")
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	return &FFmpeg{ffmpegPath: ffmpegPath, ffprobePath: ffprobePath}
}

// ffprobeOutput mirrors the relevant parts of `ffprobe -print_format json`.
type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType   string `json:"codec_type"`
		CodecName   string `json:"codec_name"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Disposition struct {
			AttachedPic int `json:"attached_pic"`
		} `json:"disposition"`
	} `json:"streams"`
}

// Probe returns metadata about a media file using ffprobe.
func (f *FFmpeg) Probe(input string) (*MediaInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		input,
	}

	out, err := exec.Command(f.ffprobePath, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe %s: %w", input, err)
	}
	return parseProbeOutput(out)
}

// parseProbeOutput converts raw ffprobe JSON into a MediaInfo. Kept separate so
// it can be unit tested without invoking ffprobe.
func parseProbeOutput(data []byte) (*MediaInfo, error) {
	var raw ffprobeOutput
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	info := &MediaInfo{}
	if raw.Format.Duration != "" {
		if d, err := strconv.ParseFloat(raw.Format.Duration, 64); err == nil {
			info.Duration = d
		}
	}
	if raw.Format.BitRate != "" {
		if b, err := strconv.ParseInt(raw.Format.BitRate, 10, 64); err == nil {
			info.Bitrate = b
		}
	}

	for _, s := range raw.Streams {
		switch s.CodecType {
		case "video":
			// Embedded cover art (e.g. in an mp3) shows up as a video stream
			// with the attached_pic disposition; it must not make an audio
			// file count as video.
			if s.Disposition.AttachedPic == 1 {
				continue
			}
			if !info.HasVideo {
				info.HasVideo = true
				info.VideoCodec = s.CodecName
				info.Width = s.Width
				info.Height = s.Height
			}
		case "audio":
			if !info.HasAudio {
				info.HasAudio = true
				info.AudioCodec = s.CodecName
			}
		}
	}
	return info, nil
}

// ExtractPoster writes a single scaled JPEG frame from input to output. seek is
// the offset in seconds to grab the frame from (placed before -i so ffmpeg can
// seek without decoding the whole file).
func (f *FFmpeg) ExtractPoster(ctx context.Context, input, output string, seek float64) error {
	if seek < 0 {
		seek = 0
	}
	args := []string{
		"-hide_banner", "-nostdin",
		"-ss", fmt.Sprintf("%.2f", seek),
		"-i", input,
		"-frames:v", "1",
		"-vf", "scale=480:-2",
		"-q:v", "4",
		// The caller writes to a temp name without an image extension, so
		// the format cannot be inferred from the filename.
		"-f", "image2",
		"-y", output,
	}
	out, err := exec.CommandContext(ctx, f.ffmpegPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("extract poster from %s: %w: %s", input, err, tailOf(string(out)))
	}
	return nil
}

// tailOf returns the last few lines of ffmpeg output for error messages.
func tailOf(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > maxTailLines {
		lines = lines[len(lines)-maxTailLines:]
	}
	return strings.Join(lines, "\n")
}

// Run executes ffmpeg with the supplied operation args, appending the universal
// progress and output flags. totalDuration (seconds) is used to turn ffmpeg's
// reported timestamp into a percentage; pass 0 if unknown. The callback is
// invoked as progress is reported and once more at 100% on success.
func (f *FFmpeg) Run(ctx context.Context, opArgs []string, output string, totalDuration float64, cb ProgressFunc) error {
	args := append([]string{"-hide_banner", "-nostdin"}, opArgs...)
	args = append(args, "-progress", "pipe:1", "-nostats", "-y", output)

	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stderr pipe: %w", err)
	}

	log.WithField("args", args).Debug("Executing ffmpeg command")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	start := time.Now()

	// Capture the tail of stderr for diagnostics on failure.
	var stderrTail tailBuffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrTail.add(line)
			log.Debug(line)
		}
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if micros, ok := parseOutTimeMicros(scanner.Text()); ok && cb != nil {
			cb(computeProgress(micros, totalDuration), time.Since(start))
		}
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("ffmpeg cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("ffmpeg failed: %w: %s", err, stderrTail.string())
	}

	if cb != nil {
		cb(100, time.Since(start))
	}
	return nil
}

// parseOutTimeMicros extracts the microsecond timestamp from an ffmpeg
// `-progress` line such as "out_time_us=1234567". It returns false for any
// other line (including "out_time_us=N/A" emitted before encoding starts).
func parseOutTimeMicros(line string) (int64, bool) {
	const prefix = "out_time_us="
	if !strings.HasPrefix(line, prefix) {
		return 0, false
	}
	value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	micros, err := strconv.ParseInt(value, 10, 64)
	if err != nil || micros < 0 {
		return 0, false
	}
	return micros, true
}

// computeProgress turns an out_time in microseconds into a clamped 0-100
// percentage given a total duration in seconds.
func computeProgress(outTimeMicros int64, totalDuration float64) float64 {
	if totalDuration <= 0 {
		return 0
	}
	percent := (float64(outTimeMicros) / 1_000_000.0 / totalDuration) * 100.0
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

// tailBuffer keeps the last maxTailLines lines of output.
type tailBuffer struct {
	lines []string
}

const maxTailLines = 20

func (t *tailBuffer) add(line string) {
	t.lines = append(t.lines, line)
	if len(t.lines) > maxTailLines {
		t.lines = t.lines[len(t.lines)-maxTailLines:]
	}
}

func (t *tailBuffer) string() string {
	return strings.Join(t.lines, "\n")
}
