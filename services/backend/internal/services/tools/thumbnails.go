package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// thumbsDirName is the directory under the processed path holding generated
// poster images. Hidden so it never shows up next to the actual outputs.
const thumbsDirName = ".thumbs"

// thumbnailPathFor returns the on-disk poster path for a job ID, validated to
// stay within the processed directory (job IDs come from the database, but the
// guard keeps a hostile ID from ever becoming a traversal).
func (s *Service) thumbnailPathFor(jobID string) (string, error) {
	base := filepath.Clean(s.processedPath)
	path := filepath.Clean(filepath.Join(base, thumbsDirName, jobID+".jpg"))
	if err := ensureWithin(base, path); err != nil {
		return "", err
	}
	return path, nil
}

// ThumbnailPath returns the poster image for a completed video job, generating
// it from the output file if it does not exist yet. That lazy generation also
// backfills jobs completed before thumbnails existed. Audio-only outputs have
// no poster and return an error.
func (s *Service) ThumbnailPath(job *domain.ToolsJob) (string, error) {
	thumb, err := s.thumbnailPathFor(job.ID)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(thumb); err == nil && !info.IsDir() {
		return thumb, nil
	}

	if job.MediaKind == domain.MediaKindAudio {
		return "", fmt.Errorf("audio output has no thumbnail")
	}

	source, err := s.ResolveOutputFile(job)
	if err != nil {
		return "", err
	}

	duration := job.Duration
	// Legacy rows have no probed metadata; probe now so an audio file never
	// gets a frame extraction attempt.
	if job.MediaKind == "" {
		info, err := s.ffmpeg.Probe(source)
		if err != nil {
			return "", fmt.Errorf("probe output: %w", err)
		}
		if !info.HasVideo {
			return "", fmt.Errorf("audio output has no thumbnail")
		}
		duration = info.Duration
	}

	if err := os.MkdirAll(filepath.Dir(thumb), 0o755); err != nil {
		return "", fmt.Errorf("create thumbnails directory: %w", err)
	}

	// Grab a frame early in the file, but never past short outputs. Generate
	// to a temp name and rename so concurrent requests never see a partial
	// file; a racing double-generation is idempotent.
	seek := duration * 0.1
	if seek > 10 {
		seek = 10
	}
	if duration < 1 {
		seek = 0
	}
	tmp := thumb + ".tmp"
	if err := s.ffmpeg.ExtractPoster(context.Background(), source, tmp, seek); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, thumb); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("finalize thumbnail: %w", err)
	}
	return thumb, nil
}

// removeThumbnail deletes a job's poster image, if any. Used when the job
// itself is deleted.
func (s *Service) removeThumbnail(jobID string) {
	path, err := s.thumbnailPathFor(jobID)
	if err != nil {
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.WithError(err).WithField("path", path).Warn("Failed to delete tools thumbnail")
	}
}
