package download

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintedFilepath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single line", "/data/downloads/Uploader/Title.mp4\n", "/data/downloads/Uploader/Title.mp4"},
		{"last non-empty wins", "/data/a.mp4\n\n/data/b.mp4\n\n", "/data/b.mp4"},
		{"whitespace only", "   \n\t\n", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := printedFilepath(strings.NewReader(tt.input)); got != tt.want {
				t.Errorf("printedFilepath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrintedFilepathsByID(t *testing.T) {
	input := "abc123\t/data/downloads/Channel/Video One.mp4\n" +
		"\n" +
		"no-tab-line\n" +
		"def456\t/data/downloads/AC⧸DC/Thunderstruck.mp4\n" +
		"\t/data/missing-id.mp4\n" +
		"ghi789\t\n"

	got := printedFilepathsByID(strings.NewReader(input))

	want := map[string]string{
		"abc123": "/data/downloads/Channel/Video One.mp4",
		"def456": "/data/downloads/AC⧸DC/Thunderstruck.mp4",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(got), len(want), got)
	}
	for id, path := range want {
		if got[id] != path {
			t.Errorf("got[%q] = %q, want %q", id, got[id], path)
		}
	}
}

func TestValidMediaPath(t *testing.T) {
	base := t.TempDir()
	inside := filepath.Join(base, "Uploader", "video.mp4")
	if err := os.MkdirAll(filepath.Dir(inside), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !validMediaPath(base, inside) {
		t.Errorf("expected %q to be valid inside %q", inside, base)
	}
	if validMediaPath(base, filepath.Join(base, "Uploader", "missing.mp4")) {
		t.Error("missing file should not be valid")
	}
	if validMediaPath(base, filepath.Join(base, "..", "escape.mp4")) {
		t.Error("path escaping base should not be valid")
	}
	if validMediaPath(base, filepath.Join(base, "Uploader")) {
		t.Error("directory should not be valid")
	}
}

func TestDownloadFormatArgs(t *testing.T) {
	args := downloadFormatArgs(1080)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-S res:1080,vcodec:h264,acodec:m4a") {
		t.Errorf("expected format sort for browser-safe codecs, got %q", joined)
	}
	if !strings.Contains(joined, "-f bv*+ba/b") {
		t.Errorf("expected generic best-video+best-audio selector, got %q", joined)
	}
	if !strings.Contains(joined, "--merge-output-format mp4") {
		t.Errorf("expected mp4 merge container, got %q", joined)
	}
}
