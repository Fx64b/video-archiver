package tools

import (
	"os"
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

func TestCandidateDirs(t *testing.T) {
	meta := &domain.VideoMetadata{Uploader: "Some Uploader", Channel: "Some Channel"}
	dirs := candidateDirs("/downloads", meta)
	if dirs[0] != filepath.Join("/downloads", "Some Uploader") {
		t.Errorf("first dir = %q, want uploader dir", dirs[0])
	}
	if len(dirs) != 2 || dirs[1] != filepath.Join("/downloads", "Some Channel") {
		t.Errorf("expected channel dir as fallback, got %v", dirs)
	}

	// Uploader == Channel deduplicates to a single dir.
	meta = &domain.VideoMetadata{Uploader: "Fireship", Channel: "Fireship"}
	if dirs := candidateDirs("/downloads", meta); len(dirs) != 1 {
		t.Errorf("expected deduplicated single dir, got %v", dirs)
	}

	// Empty uploader/channel falls back to Unknown, still within base.
	meta = &domain.VideoMetadata{}
	dirs = candidateDirs("/downloads", meta)
	if len(dirs) != 1 || !strings.HasPrefix(dirs[0], filepath.Clean("/downloads")) {
		t.Errorf("unexpected fallback dirs: %v", dirs)
	}
}

func TestNormalizeForMatch(t *testing.T) {
	tests := map[string]string{
		"Google's AI endgame… at I/O 2026": "googlesaiendgameatio2026",
		"10 weird OSS projects...":         "10weirdossprojects",
		"  Mixed CASE 99 ":                 "mixedcase99",
		"":                                 "",
	}
	for in, want := range tests {
		if got := normalizeForMatch(in); got != want {
			t.Errorf("normalizeForMatch(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestResolveVideoFileSanitizationMismatch reproduces the reported failure: the
// title contains characters ('/', '…', a curly quote) and the file on disk has a
// different extension than the metadata hint. Exact reconstruction fails but the
// normalized directory scan must still find the file.
func TestResolveVideoFileSanitizationMismatch(t *testing.T) {
	downloads := t.TempDir()
	dir := filepath.Join(downloads, "Fireship")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On-disk: curly apostrophe, ellipsis, '/'→'_', and a .webm extension.
	onDisk := "Google’s AI endgame is here… everything you missed at I_O 2026.webm"
	target := filepath.Join(dir, onDisk)
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	meta := &domain.VideoMetadata{
		Uploader:  "Fireship",
		Title:     "Google’s AI endgame is here… everything you missed at I/O 2026",
		Extension: "mp4", // wrong hint on purpose
	}
	got, err := ResolveVideoFile(downloads, meta)
	if err != nil {
		t.Fatalf("ResolveVideoFile: %v", err)
	}
	if got != target {
		t.Errorf("ResolveVideoFile = %q, want %q", got, target)
	}
}

func TestResolveVideoFileExactAndTrailingDots(t *testing.T) {
	downloads := t.TempDir()
	dir := filepath.Join(downloads, "Fireship")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Trailing dots are stripped by sanitizeFilename but kept by yt-dlp on disk.
	onDisk := filepath.Join(dir, "10 weird OSS projects you need right now....mp4")
	if err := os.WriteFile(onDisk, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := &domain.VideoMetadata{Uploader: "Fireship", Title: "10 weird OSS projects you need right now...", Extension: "mp4"}
	got, err := ResolveVideoFile(downloads, meta)
	if err != nil || got != onDisk {
		t.Fatalf("ResolveVideoFile = %q, %v; want %q", got, err, onDisk)
	}
}

func TestResolveVideoFileNotFound(t *testing.T) {
	downloads := t.TempDir()
	meta := &domain.VideoMetadata{Uploader: "Nobody", Title: "Nope", Extension: "mp4"}
	if _, err := ResolveVideoFile(downloads, meta); err == nil {
		t.Error("expected error when no file exists")
	}
}

func TestResolveVideoFileTraversalSafe(t *testing.T) {
	downloads := t.TempDir()
	meta := &domain.VideoMetadata{Uploader: "../../etc", Title: "../passwd", Extension: "mp4"}
	// Should not find anything outside the base and must not error-escape.
	if _, err := ResolveVideoFile(downloads, meta); err == nil {
		t.Error("expected not-found error for traversal-style names")
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
