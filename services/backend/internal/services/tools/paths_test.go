package tools

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"video-archiver/internal/domain"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"normal", "normal"},
		{"a/b:c*d?", "a_b_c_d_"},
		{`bad\name`, "bad_name"},
		{"  spaced  ", "spaced"},
		{"...", "Unknown"},
		{"", "Unknown"},
		{"a<b>c|d\"e", "a_b_c_d_e"},
	}
	for _, tt := range tests {
		if got := sanitizeFilename(tt.in); got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestOrderedExtensions(t *testing.T) {
	got := orderedExtensions("mkv")
	if got[0] != "mkv" {
		t.Errorf("preferred extension should be first: %v", got)
	}
	// No duplicates.
	seen := map[string]bool{}
	for _, e := range got {
		if seen[e] {
			t.Errorf("duplicate extension %q in %v", e, got)
		}
		seen[e] = true
	}

	got = orderedExtensions("")
	if !reflect.DeepEqual(got, videoFileExtensions) {
		t.Errorf("empty preference = %v, want %v", got, videoFileExtensions)
	}

	got = orderedExtensions(".MP4")
	if got[0] != "mp4" {
		t.Errorf("expected normalized mp4 first, got %v", got)
	}
}

func TestCandidateInputPaths(t *testing.T) {
	meta := &domain.VideoMetadata{Uploader: "Some Uploader", Title: "My Video", Extension: "mkv"}
	paths, err := candidateInputPaths("/downloads", meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/downloads", "Some Uploader", "My Video.mkv")
	if paths[0] != want {
		t.Errorf("first candidate = %q, want %q", paths[0], want)
	}

	// Falls back to channel when uploader is empty.
	meta = &domain.VideoMetadata{Channel: "Chan", Title: "T", Extension: "mp4"}
	paths, _ = candidateInputPaths("/downloads", meta)
	if !strings.Contains(paths[0], filepath.Join("Chan", "T.mp4")) {
		t.Errorf("expected channel fallback in %q", paths[0])
	}
}

func TestCandidateInputPathsTraversal(t *testing.T) {
	meta := &domain.VideoMetadata{Uploader: "../../etc", Title: "../passwd", Extension: "mp4"}
	// Sanitization neutralizes separators, so this must stay within the base.
	paths, err := candidateInputPaths("/downloads", meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range paths {
		if !strings.HasPrefix(p, filepath.Clean("/downloads")) {
			t.Errorf("path escaped base: %q", p)
		}
	}
}

func TestEnsureWithin(t *testing.T) {
	if err := ensureWithin("/base", "/base/sub/file.mp4"); err != nil {
		t.Errorf("expected within: %v", err)
	}
	if err := ensureWithin("/base", "/other/file.mp4"); err == nil {
		t.Error("expected escape error")
	}
}

func TestOutputExtension(t *testing.T) {
	tests := []struct {
		op     domain.ToolsOperationType
		params map[string]any
		want   string
	}{
		{domain.OpTypeExtractAudio, map[string]any{"output_format": "flac"}, "flac"},
		{domain.OpTypeExtractAudio, map[string]any{}, "mp3"},
		{domain.OpTypeConvert, map[string]any{"output_format": "webm"}, "webm"},
		{domain.OpTypeConcat, map[string]any{}, "mp4"},
		{domain.OpTypeTrim, map[string]any{}, "mp4"},
		{domain.OpTypeRotate, nil, "mp4"},
	}
	for _, tt := range tests {
		if got := outputExtension(tt.op, tt.params); got != tt.want {
			t.Errorf("outputExtension(%s) = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestGenerateOutputPath(t *testing.T) {
	path := generateOutputPath("/processed", domain.OpTypeConvert, "abcdef1234567890", map[string]any{"output_format": "webm"})
	if filepath.Dir(path) != "/processed" {
		t.Errorf("unexpected dir: %q", path)
	}
	if !strings.HasSuffix(path, ".webm") {
		t.Errorf("expected .webm suffix: %q", path)
	}
	if !strings.Contains(filepath.Base(path), "abcdef12") {
		t.Errorf("expected short job id in name: %q", path)
	}
}
