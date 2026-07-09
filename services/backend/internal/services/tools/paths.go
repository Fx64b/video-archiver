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

// audioFileExtensions are the extensions an audio-only download may have.
var audioFileExtensions = []string{"mp3", "m4a", "opus", "aac", "flac", "wav", "ogg"}

// mediaFileExtensions is every extension a downloaded media file may have,
// video containers first since they are the common case.
var mediaFileExtensions = append(append([]string{}, videoFileExtensions...), audioFileExtensions...)

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
// followed by the remaining known media extensions, without duplicates.
func orderedExtensions(preferred string) []string {
	preferred = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(preferred)), ".")
	exts := make([]string, 0, len(mediaFileExtensions)+1)
	seen := map[string]bool{}
	if preferred != "" {
		exts = append(exts, preferred)
		seen[preferred] = true
	}
	for _, e := range mediaFileExtensions {
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
		if !contains(mediaFileExtensions, ext) {
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

// ResolveVideoFile locates the on-disk file for a downloaded video without a
// stored path hint. Prefer ResolveVideoFileWithHint when the job has a
// recorded file_path.
func ResolveVideoFile(downloadPath string, meta *domain.VideoMetadata) (string, error) {
	return ResolveVideoFileWithHint(downloadPath, "", meta)
}

// ResolveVideoFileWithHint locates the on-disk file for a downloaded video.
// storedPath is the path recorded at download time and wins when it still
// exists. Otherwise the file is found under
// <downloadPath>/<uploader>/<title>.<ext>: first by exact reconstruction, then
// by normalized title matching inside the reconstructed directories, and
// finally by a normalized scan of the uploader directories themselves —
// yt-dlp's own sanitization maps characters like '/' and ':' to lookalikes
// ('⧸', '：') that reconstruction can't reproduce, so directory names must be
// matched the same way titles are.
func ResolveVideoFileWithHint(downloadPath string, storedPath string, meta *domain.VideoMetadata) (string, error) {
	base := filepath.Clean(downloadPath)

	if storedPath != "" {
		p := filepath.Clean(storedPath)
		if ensureWithin(base, p) == nil {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p, nil
			}
		}
		// Stale or invalid hint (file moved/deleted): fall through to search.
	}

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

	// Last resort: the uploader directory itself may be named with yt-dlp's
	// sanitization. Match directory names by normalization, then look for the
	// title inside. Downloads are always exactly one level deep, so a single
	// directory scan is sufficient.
	for _, dir := range normalizedCandidateDirs(base, meta, dirs) {
		if p, ok := findVideoInDir(dir, meta.Title); ok {
			return p, nil
		}
	}

	return "", fmt.Errorf("video file for %q not found on disk", meta.Title)
}

// normalizedCandidateDirs scans base for directories whose normalized name
// matches the video's uploader or channel, skipping any directory already in
// tried.
func normalizedCandidateDirs(base string, meta *domain.VideoMetadata, tried []string) []string {
	targets := map[string]bool{}
	for _, name := range []string{meta.Uploader, meta.Channel} {
		if n := normalizeForMatch(name); n != "" {
			targets[n] = true
		}
	}
	if len(targets) == 0 {
		return nil
	}

	alreadyTried := map[string]bool{}
	for _, dir := range tried {
		alreadyTried[dir] = true
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, e := range entries {
		if !e.IsDir() || !targets[normalizeForMatch(e.Name())] {
			continue
		}
		dir := filepath.Join(base, e.Name())
		if !alreadyTried[dir] {
			dirs = append(dirs, dir)
		}
	}
	return dirs
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
// A user-supplied "output_name" parameter takes precedence over the generated
// name; the job's short ID is appended if that name is already taken.
func generateOutputPath(processedPath string, op domain.ToolsOperationType, jobID string, params map[string]any) string {
	timestamp := time.Now().Format("20060102_150405")
	shortID := jobID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	ext := outputExtension(op, params)

	if custom, ok := params["output_name"].(string); ok {
		if name := sanitizeOutputName(custom, ext); name != "" {
			path := filepath.Join(processedPath, name+"."+ext)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return path
			}
			return filepath.Join(processedPath, name+"_"+shortID+"."+ext)
		}
	}

	name := fmt.Sprintf("%s_%s_%s.%s", op, timestamp, shortID, ext)
	return filepath.Join(processedPath, name)
}

// outputNameExtensions are extensions stripped from a user-supplied output name
// so "clip.mp4" does not turn into "clip.mp4.mp4".
var outputNameExtensions = []string{"mp4", "mkv", "webm", "avi", "mov", "mp3", "aac", "flac", "wav", "ogg"}

// sanitizeOutputName turns a user-supplied output name into a safe file stem:
// directory components and filesystem-unsafe characters are removed, a known
// media extension (or the target extension) is stripped, and the length capped.
func sanitizeOutputName(name, targetExt string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = sanitizeFilename(name)
	if name == "Unknown" {
		return ""
	}

	lower := strings.ToLower(name)
	for _, ext := range append([]string{targetExt}, outputNameExtensions...) {
		if suffix := "." + ext; strings.HasSuffix(lower, suffix) {
			name = name[:len(name)-len(suffix)]
			break
		}
	}

	name = strings.Trim(name, " .")
	if len(name) > 100 {
		name = strings.Trim(name[:100], " .")
	}
	return name
}
