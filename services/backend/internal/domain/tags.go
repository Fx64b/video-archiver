package domain

import "strings"

// TagSource records how a tag got attached to a job.
const (
	TagSourceUser = "user"
	TagSourceAuto = "auto"
)

// Tag is a label attached to a downloaded video, playlist or channel. Count is
// only populated when listing the tag catalog; Source is only populated when a
// tag is returned in the context of a specific job.
type Tag struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source,omitempty"`
	Count  int    `json:"count,omitempty"`
}

// maxAutoTags caps how many tags auto-tagging attaches per item so noisy
// metadata (yt-dlp keyword lists can have dozens of entries) stays manageable.
const maxAutoTags = 8

// AutoTagsFor derives tags from extracted metadata: the content's categories
// and channel name, topped up with the first few uploader-supplied keywords.
// The result is deterministic so re-applying it on metadata refresh is
// idempotent.
func AutoTagsFor(metadata Metadata) []string {
	var candidates []string
	switch m := metadata.(type) {
	case *VideoMetadata:
		candidates = append(candidates, m.Categories...)
		candidates = append(candidates, m.Channel)
		candidates = append(candidates, m.Tags...)
	case *PlaylistMetadata:
		candidates = append(candidates, m.Channel)
	case *ChannelMetadata:
		candidates = append(candidates, m.Channel)
	}

	tags := make([]string, 0, maxAutoTags)
	seen := map[string]bool{}
	for _, c := range candidates {
		name := NormalizeTagName(c)
		if name == "" || seen[strings.ToLower(name)] {
			continue
		}
		seen[strings.ToLower(name)] = true
		tags = append(tags, name)
		if len(tags) >= maxAutoTags {
			break
		}
	}
	return tags
}

// NormalizeTagName trims and length-caps a tag name. Tags are matched
// case-insensitively by the database, so casing is preserved for display.
func NormalizeTagName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	if runes := []rune(name); len(runes) > 50 {
		name = strings.TrimSpace(string(runes[:50]))
	}
	return name
}
