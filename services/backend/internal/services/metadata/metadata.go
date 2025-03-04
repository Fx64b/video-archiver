package metadata

import (
	"encoding/json"
	"os"
	"strings"
	"video-archiver/internal/domain"
)

func ExtractMetadata(path string) (domain.Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// First try to unmarshal as a generic map to check the _type field
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	typeStr, _ := rawData["_type"].(string)

	_, hasChannelID := rawData["channel_id"]
	title, _ := rawData["title"].(string)
	isChannel := typeStr == "playlist" && hasChannelID && strings.HasSuffix(title, " - Videos")

	if isChannel {
		var metadata domain.ChannelMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, err
		}
		if strings.HasSuffix(metadata.Channel, " - Videos") {
			metadata.Channel = strings.TrimSuffix(metadata.Channel, " - Videos")
		}
		metadata.Type = "channel"
		return &metadata, nil
	} else if typeStr == "playlist" {
		var metadata domain.PlaylistMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, err
		}
		return &metadata, nil
	} else {
		var metadata domain.VideoMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, err
		}
		return &metadata, nil
	}
}
