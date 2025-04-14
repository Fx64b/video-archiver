package statistics

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"video-archiver/internal/domain"

	log "github.com/sirupsen/logrus"
)

func GetStatistics(repo domain.JobRepository, downloadPath string) (domain.Statistics, error) {
	stats := domain.Statistics{
		LastUpdate: time.Now(),
	}

	jobs, err := repo.GetJobs()
	if err != nil {
		return stats, fmt.Errorf("failed to get jobs: %w", err)
	}
	stats.TotalJobs = len(jobs)

	videoCount, err := repo.CountVideos()
	if err != nil {
		log.WithError(err).Warn("Failed to count videos")
	} else {
		stats.TotalVideos = videoCount
	}

	playlistCount, err := repo.CountPlaylists()
	if err != nil {
		log.WithError(err).Warn("Failed to count playlists")
	} else {
		stats.TotalPlaylists = playlistCount
	}

	channelCount, err := repo.CountChannels()
	if err != nil {
		log.WithError(err).Warn("Failed to count channels")
	} else {
		stats.TotalChannels = channelCount
	}

	// Calculate storage usage if download path is provided
	if downloadPath != "" {
		totalSize, err := calculateDirSize(downloadPath)
		if err != nil {
			log.WithError(err).Warn("Failed to calculate total storage")
		} else {
			stats.TotalStorage = totalSize

			// We're assigning all storage to videos for now
			// A more sophisticated approach would categorize files by type
			stats.VideosStorage = totalSize
		}
	}

	return stats, nil
}

func calculateDirSize(path string) (int, error) {
	var size int

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return 0, fmt.Errorf("directory does not exist: %s", path)
	}

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += int(info.Size())
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate directory size: %w", err)
	}

	return size, nil
}
