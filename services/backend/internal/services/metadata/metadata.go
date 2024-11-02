package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"video-archiver/internal/storage"
	"video-archiver/models"
)

func ExtractAndStoreMetadata(jobID, downloadPath string, isPlaylist bool) error {
	var folderPath string
	if isPlaylist {
		folderPath = filepath.Join(downloadPath, jobID)
	} else {
		folderPath = downloadPath
	}

	playlistInfoPath := ""
	if isPlaylist {
		err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), ".info.json") && isPlaylistFile(path) {
				playlistInfoPath = path
				return nil
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory %s: %v", folderPath, err)
		}
	}

	if playlistInfoPath != "" {
		if err := parseAndStorePlaylistMetadata(playlistInfoPath, jobID); err != nil {
			return fmt.Errorf("failed to parse playlist metadata: %v", err)
		}
	}

	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".info.json") && path != playlistInfoPath {
			return parseAndStoreVideoMetadata(path, jobID)
		}
		return nil
	})
}

func isPlaylistFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return false
	}

	_, hasID := data["id"]
	_, hasTitle := data["title"]
	return hasID && hasTitle && data["id"] != nil && data["title"] != nil
}

func parseAndStoreVideoMetadata(filePath, jobID string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var videoData models.VideoMetadata
	if err := json.NewDecoder(file).Decode(&videoData); err != nil {
		return err
	}

	video := models.Video{
		JobID:            jobID,
		Title:            videoData.Title,
		Uploader:         videoData.Uploader,
		FilePath:         filePath,
		LastDownloadedAt: time.Now(),
		Length:           videoData.Duration,
		Size:             videoData.Filesize,
		Quality:          videoData.Format,
	}
	if err := storage.SaveVideoMetadata(video); err != nil {
		return err
	}

	return os.Remove(filePath)
}

func parseAndStorePlaylistMetadata(filePath, jobID string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var playlistData models.PlaylistMetadata
	if err := json.Unmarshal(file, &playlistData); err != nil {
		return err
	}

	// Store playlist data
	playlist := models.Playlist{
		ID:          playlistData.ID,
		Title:       playlistData.Title,
		Description: playlistData.Description,
	}
	return storage.SavePlaylistMetadata(playlist)
}
