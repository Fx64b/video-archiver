package tools

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"video-archiver/internal/domain"
)

// videoFileExtensions are the container extensions a downloaded video may have.
// yt-dlp picks the extension based on the muxed output, so the stored metadata
// extension is a hint rather than a guarantee.
var videoFileExtensions = []string{"mp4", "mkv", "webm", "avi", "mov"}

// sanitizeFilename mirrors the way the download service names files closely
// enough to locate them, replacing characters that are illegal on common
// filesystems. It also strips leading/trailing dots and spaces.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"\n", " ",
		"\r", " ",
	)
	cleaned := strings.TrimSpace(replacer.Replace(name))
	cleaned = strings.Trim(cleaned, ".")
	if cleaned == "" {
		cleaned = "Unknown"
	}
	return cleaned
}

// candidateInputPaths returns the possible on-disk locations for a downloaded
// video, most likely first. Downloads are stored as
// <downloadPath>/<uploader>/<title>.<ext>. The metadata extension is tried
// first, followed by the other known video extensions.
func candidateInputPaths(downloadPath string, meta *domain.VideoMetadata) ([]string, error) {
	uploader := meta.Uploader
	if uploader == "" {
		uploader = meta.Channel
	}
	uploader = sanitizeFilename(uploader)
	title := sanitizeFilename(meta.Title)

	dir := filepath.Join(downloadPath, uploader)

	exts := orderedExtensions(meta.Extension)
	paths := make([]string, 0, len(exts))
	base := filepath.Clean(downloadPath)
	for _, ext := range exts {
		p := filepath.Clean(filepath.Join(dir, title+"."+ext))
		if err := ensureWithin(base, p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

// orderedExtensions returns the metadata extension first (if known and valid)
// followed by the remaining known video extensions, without duplicates.
func orderedExtensions(preferred string) []string {
	preferred = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(preferred)), ".")
	exts := make([]string, 0, len(videoFileExtensions)+1)
	seen := map[string]bool{}
	if preferred != "" {
		exts = append(exts, preferred)
		seen[preferred] = true
	}
	for _, e := range videoFileExtensions {
		if !seen[e] {
			exts = append(exts, e)
			seen[e] = true
		}
	}
	return exts
}

// ensureWithin returns an error if target escapes base (path traversal guard).
func ensureWithin(base, target string) error {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid path: %q escapes %q", target, base)
	}
	return nil
}

// outputExtension determines the extension of the produced file for a job.
func outputExtension(op domain.ToolsOperationType, params map[string]any) string {
	switch op {
	case domain.OpTypeExtractAudio:
		if format, ok := params["output_format"].(string); ok && format != "" {
			return format
		}
		return "mp3"
	case domain.OpTypeConvert, domain.OpTypeConcat:
		if format, ok := params["output_format"].(string); ok && format != "" {
			return format
		}
		return "mp4"
	default:
		return "mp4"
	}
}

// generateOutputPath builds a unique output path inside processedPath for a job.
func generateOutputPath(processedPath string, op domain.ToolsOperationType, jobID string, params map[string]any) string {
	timestamp := time.Now().Format("20060102_150405")
	shortID := jobID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	name := fmt.Sprintf("%s_%s_%s.%s", op, timestamp, shortID, outputExtension(op, params))
	return filepath.Join(processedPath, name)
}
