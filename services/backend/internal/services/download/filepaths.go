package download

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// yt-dlp is asked to record where each finished file ended up via
// --print-to-file with the "after_move" hook. A separate file is used because
// stdout already carries the --progress-template stream that trackProgress
// parses, and after_move fires only once the merged file is at its final
// location (never an intermediate .f137.mp4 fragment).

// printedFilepath returns the last non-empty line of a file written with
// --print-to-file "after_move:%(filepath)s" — the final media file location.
func printedFilepath(r io.Reader) string {
	var last string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			last = line
		}
	}
	return last
}

// printedFilepathsByID parses lines of "<video id>\t<filepath>" written with
// --print-to-file "after_move:%(id)s\t%(filepath)s" during playlist/channel
// downloads. Videos that failed simply don't appear. Malformed lines are
// skipped.
func printedFilepathsByID(r io.Reader) map[string]string {
	paths := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		id, path, found := strings.Cut(line, "\t")
		id = strings.TrimSpace(id)
		path = strings.TrimSpace(path)
		if !found || id == "" || path == "" {
			continue
		}
		paths[id] = path
	}
	return paths
}

// createPrintFile creates the temp file yt-dlp writes finished-file paths to.
// Returns an empty path (and a no-op cleanup) if creation fails — path capture
// is best-effort and must never block a download.
func createPrintFile(jobID string) (string, func()) {
	f, err := os.CreateTemp("", "ytdlp-filepath-*")
	if err != nil {
		log.WithError(err).WithField("jobID", jobID).
			Warn("Failed to create print file; downloaded file path will not be recorded")
		return "", func() {}
	}
	name := f.Name()
	f.Close()
	return name, func() { os.Remove(name) }
}

// downloadFormatArgs selects what yt-dlp downloads. -S is *sorting*, not
// filtering: res prefers the largest resolution not above maxQuality (instead
// of failing like a hard height filter when only larger formats exist), and
// vcodec:h264/acodec:m4a prefer avc1+aac — codecs every browser can decode —
// falling back gracefully when unavailable. Merging into mp4 only remuxes, so
// preferring mp4-native codecs here is what keeps the file playable in the
// web player.
func downloadFormatArgs(maxQuality int) []string {
	return []string{
		"-f", "bv*+ba/b",
		"-S", "res:" + strconv.Itoa(maxQuality) + ",vcodec:h264,acodec:m4a",
		"--merge-output-format", "mp4",
	}
}

// validMediaPath reports whether path is an existing regular file inside base,
// guarding against yt-dlp output that escaped the download directory.
func validMediaPath(base, path string) bool {
	rel, err := filepath.Rel(filepath.Clean(base), filepath.Clean(path))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// recordFilePath persists the captured media file location on a job. Capture
// is best-effort: a missing or invalid path is logged, never a job failure —
// resolution falls back to scanning the download directory.
func (s *Service) recordFilePath(jobID, path string) {
	if path == "" {
		return
	}
	if !validMediaPath(s.config.DownloadPath, path) {
		log.WithField("jobID", jobID).WithField("path", path).
			Warn("Ignoring printed file path outside the download directory")
		return
	}
	if err := s.jobs.SetFilePath(jobID, path); err != nil {
		log.WithError(err).WithField("jobID", jobID).Warn("Failed to store file path")
		return
	}
	log.WithField("jobID", jobID).WithField("path", path).Debug("Recorded downloaded file path")
}
