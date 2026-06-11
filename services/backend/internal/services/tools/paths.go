package tools

import (
	"fmt"
	"os"
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

// candidateDirs returns the directories a downloaded video may live in. Downloads
// are written under the uploader (yt-dlp template %(uploader)s); the channel name
// is tried as a fallback since some metadata records it separately.
func candidateDirs(downloadPath string, meta *domain.VideoMetadata) []string {
	base := filepath.Clean(downloadPath)
	var dirs []string
	seen := map[string]bool{}
	for _, name := range []string{meta.Uploader, meta.Channel} {
		if name == "" {
			continue
		}
		dir := filepath.Clean(filepath.Join(base, sanitizeFilename(name)))
		if ensureWithin(base, dir) != nil || seen[dir] {
			continue
		}
		dirs = append(dirs, dir)
		seen[dir] = true
	}
	if len(dirs) == 0 {
		dirs = append(dirs, filepath.Join(base, "Unknown"))
	}
	return dirs
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

// normalizeForMatch reduces a name to lowercase alphanumerics so titles can be
// compared regardless of how punctuation/unicode was sanitized on disk. This is
// what makes file lookup robust against yt-dlp's filename rules (e.g. '/', '…',
// curly quotes, trailing dots).
func normalizeForMatch(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// findVideoInDir scans dir for a video file whose name matches title after
// normalization. It prefers an exact normalized match, then falls back to a
// prefix match (tolerating any trailing format suffix yt-dlp may leave).
func findVideoInDir(dir, title string) (string, bool) {
	target := normalizeForMatch(title)
	if target == "" {
		return "", false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	var prefixMatch string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
		if !contains(videoFileExtensions, ext) {
			continue
		}
		stem := normalizeForMatch(strings.TrimSuffix(name, filepath.Ext(name)))
		if stem == target {
			return filepath.Join(dir, name), true
		}
		if prefixMatch == "" && strings.HasPrefix(stem, target) {
			prefixMatch = filepath.Join(dir, name)
		}
	}
	if prefixMatch != "" {
		return prefixMatch, true
	}
	return "", false
}

// ResolveVideoFile locates the on-disk file for a downloaded video. Downloads are
// written by yt-dlp as <downloadPath>/<uploader>/<title>.<ext>, but yt-dlp's
// filename sanitization does not match a naive reconstruction for titles with
// characters like '/', '…' or curly quotes, and the container extension is not
// guaranteed. ResolveVideoFile tries the reconstructed paths first, then falls
// back to scanning the directory and matching the title with normalization.
func ResolveVideoFile(downloadPath string, meta *domain.VideoMetadata) (string, error) {
	base := filepath.Clean(downloadPath)
	dirs := candidateDirs(downloadPath, meta)
	title := sanitizeFilename(meta.Title)
	exts := orderedExtensions(meta.Extension)

	// Fast path: exact reconstructed file names.
	for _, dir := range dirs {
		for _, ext := range exts {
			p := filepath.Clean(filepath.Join(dir, title+"."+ext))
			if ensureWithin(base, p) != nil {
				continue
			}
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p, nil
			}
		}
	}

	// Robust path: scan the directory and match the title by normalization.
	for _, dir := range dirs {
		if p, ok := findVideoInDir(dir, meta.Title); ok {
			return p, nil
		}
	}

	return "", fmt.Errorf("video file for %q not found on disk", meta.Title)
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
