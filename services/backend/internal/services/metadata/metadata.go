package metadata

import (
	"encoding/json"
	"os"
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

	if typeStr, ok := rawData["_type"].(string); ok && typeStr == "playlist" {
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
