package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

func TestThumbnailPathForRejectsTraversal(t *testing.T) {
	svc, _, _ := newTestService(t, testutil.NewMockJobRepository())

	if _, err := svc.thumbnailPathFor("../../etc/passwd"); err == nil {
		t.Error("expected traversal job ID to be rejected")
	}

	got, err := svc.thumbnailPathFor("job1")
	if err != nil {
		t.Fatalf("thumbnailPathFor: %v", err)
	}
	want := filepath.Join(filepath.Clean(svc.processedPath), thumbsDirName, "job1.jpg")
	if got != want {
		t.Errorf("thumbnailPathFor = %q, want %q", got, want)
	}
}

func TestThumbnailPathAudioJob(t *testing.T) {
	svc, _, _ := newTestService(t, testutil.NewMockJobRepository())

	job := &domain.ToolsJob{ID: "a1", MediaKind: domain.MediaKindAudio, OutputFile: filepath.Join(svc.processedPath, "a.mp3")}
	if _, err := svc.ThumbnailPath(job); err == nil {
		t.Error("expected error for audio-only output")
	}
}

func TestThumbnailPathReturnsExistingWithoutGeneration(t *testing.T) {
	svc, _, _ := newTestService(t, testutil.NewMockJobRepository())

	thumb, err := svc.thumbnailPathFor("v1")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(thumb), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumb, []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}

	// No output file exists, so this only succeeds if the pre-existing
	// thumbnail short-circuits generation.
	job := &domain.ToolsJob{ID: "v1", MediaKind: domain.MediaKindVideo}
	got, err := svc.ThumbnailPath(job)
	if err != nil {
		t.Fatalf("ThumbnailPath: %v", err)
	}
	if got != thumb {
		t.Errorf("ThumbnailPath = %q, want %q", got, thumb)
	}
}

// fakeProbeFFmpeg returns an FFmpeg whose ffprobe is a shell script printing
// the given JSON, so enrichment can be tested without a real ffprobe binary.
func fakeProbeFFmpeg(t *testing.T, probeJSON string, fail bool) *FFmpeg {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "ffprobe")
	body := "#!/bin/sh\n"
	if fail {
		body += "exit 1\n"
	} else {
		body += fmt.Sprintf("cat <<'EOF'\n%s\nEOF\n", probeJSON)
	}
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return &FFmpeg{ffmpegPath: "ffmpeg", ffprobePath: script}
}

func TestEnrichOutputMetadata(t *testing.T) {
	tests := []struct {
		name      string
		probeJSON string
		fail      bool
		wantKind  string
	}{
		{
			name: "video output",
			probeJSON: `{"format": {"duration": "12.5"}, "streams": [
				{"codec_type": "video", "codec_name": "h264", "width": 1280, "height": 720},
				{"codec_type": "audio", "codec_name": "aac"}]}`,
			wantKind: domain.MediaKindVideo,
		},
		{
			name: "audio output",
			probeJSON: `{"format": {"duration": "30"}, "streams": [
				{"codec_type": "audio", "codec_name": "mp3"}]}`,
			wantKind: domain.MediaKindAudio,
		},
		{
			name:     "probe failure leaves job untouched",
			fail:     true,
			wantKind: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _ := newTestService(t, testutil.NewMockJobRepository())
			svc.ffmpeg = fakeProbeFFmpeg(t, tt.probeJSON, tt.fail)

			job := &domain.ToolsJob{ID: "j1"}
			svc.enrichOutputMetadata(job, "/tmp/whatever.out")

			if job.MediaKind != tt.wantKind {
				t.Errorf("MediaKind = %q, want %q", job.MediaKind, tt.wantKind)
			}
			switch tt.wantKind {
			case domain.MediaKindVideo:
				if job.Duration != 12.5 || job.Width != 1280 || job.Height != 720 || job.VideoCodec != "h264" || job.AudioCodec != "aac" {
					t.Errorf("unexpected video enrichment: %+v", job)
				}
			case domain.MediaKindAudio:
				if job.Duration != 30 || job.AudioCodec != "mp3" || job.Width != 0 || job.Height != 0 {
					t.Errorf("unexpected audio enrichment: %+v", job)
				}
			default:
				if job.Duration != 0 || job.VideoCodec != "" || job.AudioCodec != "" {
					t.Errorf("expected zero values after probe failure: %+v", job)
				}
			}
		})
	}
}

func TestDeleteJobRemovesThumbnail(t *testing.T) {
	svc, repo, _ := newTestService(t, testutil.NewMockJobRepository())

	out := filepath.Join(svc.processedPath, "out.mp4")
	if err := os.WriteFile(out, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	thumb, err := svc.thumbnailPathFor("j1")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(thumb), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumb, []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = repo.Create(&domain.ToolsJob{ID: "j1", Status: domain.ToolsJobStatusComplete, OutputFile: out})

	if err := svc.DeleteJob("j1"); err != nil {
		t.Fatalf("DeleteJob: %v", err)
	}
	for _, path := range []string{out, thumb} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected %s to be deleted", strings.TrimPrefix(path, svc.processedPath))
		}
	}
}
