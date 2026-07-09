package handlers

import "testing"

func TestBrowserSafeCodecs(t *testing.T) {
	tests := []struct {
		name       string
		videoCodec string
		audioCodec string
		hasAudio   bool
		want       bool
	}{
		{"h264+aac", "h264", "aac", true, true},
		{"h264+mp3", "h264", "mp3", true, true},
		{"h264 no audio track", "h264", "", false, true},
		{"h264+opus", "h264", "opus", true, false},
		{"vp9+aac", "vp9", "aac", true, false},
		{"av1+opus", "av1", "opus", true, false},
		{"empty probe", "", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := browserSafeCodecs(tt.videoCodec, tt.audioCodec, tt.hasAudio); got != tt.want {
				t.Errorf("browserSafeCodecs(%q, %q, %v) = %v, want %v",
					tt.videoCodec, tt.audioCodec, tt.hasAudio, got, tt.want)
			}
		})
	}
}
