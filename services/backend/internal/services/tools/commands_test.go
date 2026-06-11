package tools

import (
	"reflect"
	"testing"

	"video-archiver/internal/domain"
)

func TestParseTimecode(t *testing.T) {
	tests := []struct {
		in      string
		want    float64
		wantErr bool
	}{
		{"00:00:10", 10, false},
		{"1:30", 90, false},
		{"01:02:03", 3723, false},
		{"12.5", 12.5, false},
		{"90", 90, false},
		{"00:01:30.5", 90.5, false},
		{"", 0, true},
		{"aa", 0, true},
		{"1:2:3:4", 0, true},
		{"-5", 0, true},
	}
	for _, tt := range tests {
		got, err := parseTimecode(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseTimecode(%q) expected error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTimecode(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseTimecode(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestAudioCodecForFormat(t *testing.T) {
	want := map[string]string{
		"mp3":  "libmp3lame",
		"aac":  "aac",
		"flac": "flac",
		"wav":  "pcm_s16le",
		"ogg":  "libvorbis",
	}
	for format, codec := range want {
		got, err := audioCodecForFormat(format)
		if err != nil || got != codec {
			t.Errorf("audioCodecForFormat(%q) = %q, %v; want %q", format, got, err, codec)
		}
	}
	if _, err := audioCodecForFormat("xyz"); err == nil {
		t.Error("audioCodecForFormat(xyz) expected error")
	}
}

func TestRotationFilters(t *testing.T) {
	tests := []struct {
		rotation     int
		flipH, flipV bool
		want         []string
		wantErr      bool
	}{
		{90, false, false, []string{"transpose=1"}, false},
		{270, false, false, []string{"transpose=2"}, false},
		{180, false, false, []string{"transpose=2,transpose=2"}, false},
		{0, true, false, []string{"hflip"}, false},
		{0, false, true, []string{"vflip"}, false},
		{90, true, true, []string{"transpose=1", "hflip", "vflip"}, false},
		{0, false, false, nil, false},
		{45, false, false, nil, true},
	}
	for _, tt := range tests {
		got, err := rotationFilters(tt.rotation, tt.flipH, tt.flipV)
		if tt.wantErr {
			if err == nil {
				t.Errorf("rotationFilters(%d) expected error", tt.rotation)
			}
			continue
		}
		if err != nil {
			t.Errorf("rotationFilters(%d) unexpected error: %v", tt.rotation, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("rotationFilters(%d,%v,%v) = %v, want %v", tt.rotation, tt.flipH, tt.flipV, got, tt.want)
		}
	}
}

func TestParseResolutionHeight(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"1080p", 1080, false},
		{"720", 720, false},
		{"480p", 480, false},
		{"abc", 0, true},
		{"0p", 0, true},
		{"-2p", 0, true},
	}
	for _, tt := range tests {
		got, err := parseResolutionHeight(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseResolutionHeight(%q) expected error", tt.in)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("parseResolutionHeight(%q) = %d, %v; want %d", tt.in, got, err, tt.want)
		}
	}
}

func TestBuildTrimArgs(t *testing.T) {
	streamCopy, err := buildTrimArgs("in.mp4", &domain.TrimParameters{StartTime: "00:00:05", EndTime: "00:00:10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-i", "in.mp4", "-ss", "00:00:05", "-to", "00:00:10", "-c", "copy"}
	if !reflect.DeepEqual(streamCopy, want) {
		t.Errorf("trim stream copy args = %v, want %v", streamCopy, want)
	}

	reencode, err := buildTrimArgs("in.mp4", &domain.TrimParameters{StartTime: "0", EndTime: "10", ReEncode: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsArg(reencode, "libx264") {
		t.Errorf("trim re-encode should use libx264: %v", reencode)
	}

	if _, err := buildTrimArgs("in.mp4", &domain.TrimParameters{StartTime: "10", EndTime: "5"}); err == nil {
		t.Error("expected error when end <= start")
	}
}

func TestBuildConcatArgs(t *testing.T) {
	args, err := buildConcatArgs("/tmp/list.txt", &domain.ConcatParameters{OutputFormat: "mp4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-f", "concat", "-safe", "0", "-i", "/tmp/list.txt", "-c", "copy"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("concat args = %v, want %v", args, want)
	}

	if _, err := buildConcatArgs("/tmp/list.txt", &domain.ConcatParameters{OutputFormat: "flac"}); err == nil {
		t.Error("expected error for unsupported concat format")
	}
}

func TestBuildConvertArgs(t *testing.T) {
	defaults, err := buildConvertArgs("in.mkv", &domain.ConvertParameters{OutputFormat: "mp4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsArg(defaults, "libx264") || !containsArg(defaults, "aac") {
		t.Errorf("convert defaults should include libx264/aac: %v", defaults)
	}

	custom, err := buildConvertArgs("in.mkv", &domain.ConvertParameters{
		OutputFormat: "webm", VideoCodec: "libvpx-vp9", AudioCodec: "libopus", Bitrate: "2M",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsArg(custom, "libvpx-vp9") || !containsArg(custom, "libopus") || !containsArg(custom, "2M") {
		t.Errorf("convert custom args missing values: %v", custom)
	}

	if _, err := buildConvertArgs("in.mkv", &domain.ConvertParameters{}); err == nil {
		t.Error("expected error when output_format missing")
	}
}

func TestBuildExtractAudioArgs(t *testing.T) {
	args, err := buildExtractAudioArgs("in.mp4", &domain.ExtractAudioParameters{
		OutputFormat: "mp3", Bitrate: "320k", SampleRate: 44100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-i", "in.mp4", "-vn", "-acodec", "libmp3lame", "-b:a", "320k", "-ar", "44100"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("extract audio args = %v, want %v", args, want)
	}

	if _, err := buildExtractAudioArgs("in.mp4", &domain.ExtractAudioParameters{OutputFormat: "xyz"}); err == nil {
		t.Error("expected error for unsupported audio format")
	}
}

func TestBuildAdjustQualityArgs(t *testing.T) {
	args, err := buildAdjustQualityArgs("in.mp4", &domain.AdjustQualityParameters{Resolution: "720p", CRF: 23})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsArg(args, "scale=-2:720") {
		t.Errorf("expected scale filter in %v", args)
	}
	if !containsArg(args, "23") {
		t.Errorf("expected crf 23 in %v", args)
	}

	if _, err := buildAdjustQualityArgs("in.mp4", &domain.AdjustQualityParameters{TwoPass: true}); err == nil {
		t.Error("expected error for two_pass without bitrate")
	}
	if _, err := buildAdjustQualityArgs("in.mp4", &domain.AdjustQualityParameters{}); err == nil {
		t.Error("expected error when nothing to adjust")
	}
	if _, err := buildAdjustQualityArgs("in.mp4", &domain.AdjustQualityParameters{CRF: 99}); err == nil {
		t.Error("expected error for out-of-range crf")
	}
}

func TestBuildRotateArgs(t *testing.T) {
	args, err := buildRotateArgs("in.mp4", &domain.RotateParameters{Rotation: 90})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-i", "in.mp4", "-vf", "transpose=1", "-c:a", "copy"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("rotate args = %v, want %v", args, want)
	}

	if _, err := buildRotateArgs("in.mp4", &domain.RotateParameters{}); err == nil {
		t.Error("expected error when no rotation or flip")
	}
}

func containsArg(args []string, target string) bool {
	for _, a := range args {
		if a == target {
			return true
		}
	}
	return false
}
