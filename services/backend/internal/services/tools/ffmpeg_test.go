package tools

import "testing"

func TestParseOutTimeMicros(t *testing.T) {
	tests := []struct {
		line   string
		want   int64
		wantOK bool
	}{
		{"out_time_us=1500000", 1500000, true},
		{"out_time_us=0", 0, true},
		{"out_time_us=N/A", 0, false},
		{"frame=10", 0, false},
		{"out_time_ms=1500", 0, false},
		{"out_time_us=", 0, false},
		{"out_time_us=-5", 0, false},
	}
	for _, tt := range tests {
		got, ok := parseOutTimeMicros(tt.line)
		if ok != tt.wantOK || (ok && got != tt.want) {
			t.Errorf("parseOutTimeMicros(%q) = %d, %v; want %d, %v", tt.line, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestComputeProgress(t *testing.T) {
	tests := []struct {
		micros   int64
		duration float64
		want     float64
	}{
		{5_000_000, 10, 50},
		{10_000_000, 10, 100},
		{20_000_000, 10, 100}, // clamped
		{0, 10, 0},
		{5_000_000, 0, 0}, // unknown duration
	}
	for _, tt := range tests {
		if got := computeProgress(tt.micros, tt.duration); got != tt.want {
			t.Errorf("computeProgress(%d, %v) = %v, want %v", tt.micros, tt.duration, got, tt.want)
		}
	}
}

func TestParseProbeOutput(t *testing.T) {
	data := []byte(`{
		"format": {"duration": "123.45", "bit_rate": "5000000"},
		"streams": [
			{"codec_type": "video", "codec_name": "h264", "width": 1920, "height": 1080},
			{"codec_type": "audio", "codec_name": "aac"}
		]
	}`)

	info, err := parseProbeOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Duration != 123.45 {
		t.Errorf("duration = %v, want 123.45", info.Duration)
	}
	if info.Bitrate != 5000000 {
		t.Errorf("bitrate = %v, want 5000000", info.Bitrate)
	}
	if !info.HasVideo || info.VideoCodec != "h264" || info.Width != 1920 || info.Height != 1080 {
		t.Errorf("unexpected video info: %+v", info)
	}
	if !info.HasAudio || info.AudioCodec != "aac" {
		t.Errorf("unexpected audio info: %+v", info)
	}
}

func TestParseProbeOutputSkipsAttachedPic(t *testing.T) {
	// An mp3 with embedded cover art reports a video stream flagged as
	// attached_pic; it must still classify as audio-only.
	data := []byte(`{
		"format": {"duration": "60.0"},
		"streams": [
			{"codec_type": "video", "codec_name": "mjpeg", "width": 600, "height": 600, "disposition": {"attached_pic": 1}},
			{"codec_type": "audio", "codec_name": "mp3"}
		]
	}`)

	info, err := parseProbeOutput(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.HasVideo {
		t.Errorf("expected attached_pic stream to be ignored: %+v", info)
	}
	if !info.HasAudio || info.AudioCodec != "mp3" {
		t.Errorf("unexpected audio info: %+v", info)
	}
}

func TestParseProbeOutputInvalid(t *testing.T) {
	if _, err := parseProbeOutput([]byte("not json")); err == nil {
		t.Error("expected error for invalid json")
	}
}
