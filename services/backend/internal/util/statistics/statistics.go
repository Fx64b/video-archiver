package statistics

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"video-archiver/internal/domain"

	log "github.com/sirupsen/logrus"
)

func GetStatistics(repo domain.JobRepository, downloadPath string) (domain.Statistics, error) {
	stats := domain.Statistics{
		LastUpdate: time.Now(),
		TopVideos:  make([]domain.VideoStorageInfo, 0, 10),
	}

	// Get counts
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

	// Calculate storage usage with top videos
	if downloadPath != "" {
		topVideos, otherStorage, totalStorage, err := findTopVideosBySize(downloadPath, 10)
		if err != nil {
			log.WithError(err).Warn("Failed to calculate video storage")
			totalSize, _ := calculateDirSize(downloadPath)
			stats.TotalStorage = totalSize
			stats.OtherStorage = totalSize
		} else {
			stats.TopVideos = topVideos
			stats.OtherStorage = otherStorage
			stats.TotalStorage = totalStorage
		}
	}

	return stats, nil
}

// findTopVideosBySize returns the top N videos by size and calculates other metrics
func findTopVideosBySize(downloadPath string, limit int) ([]domain.VideoStorageInfo, int, int, error) {
	var videos []domain.VideoStorageInfo
	var totalSize int

	err := filepath.Walk(downloadPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			return nil // Skip directories
		}

		size := int(info.Size())
		totalSize += size // Count everything toward total size

		// Filter for video files
		ext := strings.ToLower(filepath.Ext(path))
		isVideo := ext == ".mp4" || ext == ".webm" || ext == ".mkv" ||
			ext == ".avi" || ext == ".mov"

		// Skip temporary/metadata files and non-video files
		if !isVideo || strings.HasSuffix(path, ".part") ||
			strings.HasSuffix(path, ".info.json") ||
			strings.HasSuffix(path, ".ytdl") {
			return nil
		}

		// Extract information
		channel := filepath.Base(filepath.Dir(path))
		fileName := filepath.Base(path)
		title := strings.TrimSuffix(fileName, ext)

		videos = append(videos, domain.VideoStorageInfo{
			Title:   title,
			Size:    size,
			Channel: channel,
		})

		return nil
	})

	if err != nil {
		return nil, 0, 0, fmt.Errorf("error walking directory: %w", err)
	}

	// Sort videos by size (largest first)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Size > videos[j].Size
	})

	// Take top N videos
	topVideos := videos
	if len(videos) > limit {
		topVideos = videos[:limit]
	}

	// Calculate size of top videos
	topSize := 0
	for _, v := range topVideos {
		topSize += v.Size
	}

	// Other storage = total - top videos
	otherStorage := totalSize - topSize

	return topVideos, otherStorage, totalSize, nil
}

// PathClassifier holds information to classify file paths by content type
type PathClassifier struct {
	videoPatterns    map[string]bool
	playlistPatterns map[string]bool
	channelPatterns  map[string]bool
	folderToType     map[string]string // Maps folder paths to content types
	mu               sync.RWMutex
}

// calculateDirSize returns the total size of all files in a directory (recursive)
func calculateDirSize(path string) (int, error) {
	var size int

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return 0, fmt.Errorf("directory does not exist: %s", path)
	}

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
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
