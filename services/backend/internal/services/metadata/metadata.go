package metadata

import (
	"encoding/json"
	"os"
	"video-archiver/internal/domain"
)

type VideoMetadata struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Thumbnail        string   `json:"thumbnail"`
	Duration         int      `json:"duration"`
	ViewCount        int      `json:"view_count"`
	Channel          string   `json:"channel"`
	ChannelID        string   `json:"channel_id"`
	ChannelURL       string   `json:"channel_url"`
	ChannelFollowers int      `json:"channel_follower_count"`
	Categories       []string `json:"categories"`
	Tags             []string `json:"tags"`
	UploadDate       string   `json:"upload_date"`
	FileSize         int64    `json:"filesize_approx"`
}

func ExtractMetadata(path string) (*domain.VideoMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata domain.VideoMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}
