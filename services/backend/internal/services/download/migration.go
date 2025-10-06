package download

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
	"video-archiver/internal/domain"
	"video-archiver/internal/services/metadata"
)

// MigratePlaylistVideos creates individual video jobs for all videos in existing playlists
func (s *Service) MigratePlaylistVideos() error {
	log.Info("Starting migration of playlist videos to individual jobs")
	
	// Get all playlists and channels from the database
	allJobs, err := s.jobs.GetAllJobsWithMetadata()
	if err != nil {
		return fmt.Errorf("failed to get all jobs: %w", err)
	}

	migrationCount := 0
	
	for _, jobWithMetadata := range allJobs {
		if jobWithMetadata.Job == nil || jobWithMetadata.Metadata == nil {
			continue
		}

		// Check if this is a playlist or channel
		switch metadata := jobWithMetadata.Metadata.(type) {
		case *domain.PlaylistMetadata:
			count, err := s.migratePlaylistVideos(jobWithMetadata.Job.ID, metadata)
			if err != nil {
				log.WithError(err).Warnf("Failed to migrate playlist %s", jobWithMetadata.Job.ID)
				continue
			}
			migrationCount += count
			log.Infof("Migrated %d videos from playlist %s", count, metadata.Title)

		case *domain.ChannelMetadata:
			count, err := s.migrateChannelVideos(jobWithMetadata.Job.ID, metadata)
			if err != nil {
				log.WithError(err).Warnf("Failed to migrate channel %s", jobWithMetadata.Job.ID)
				continue
			}
			migrationCount += count
			log.Infof("Migrated %d videos from channel %s", count, metadata.Channel)
		}
	}

	log.Infof("Migration completed. Created %d individual video jobs", migrationCount)
	return nil
}

func (s *Service) migratePlaylistVideos(playlistJobID string, playlistMetadata *domain.PlaylistMetadata) (int, error) {
	if playlistMetadata.Items == nil || len(playlistMetadata.Items) == 0 {
		log.Warnf("Playlist %s has no items to migrate", playlistJobID)
		return 0, nil
	}

	count := 0
	baseDir := s.config.DownloadPath

	for _, item := range playlistMetadata.Items {
		if item.ID == "" {
			continue
		}

		// Check if job already exists
		existingJob, err := s.jobs.GetByID(item.ID)
		if err == nil && existingJob != nil {
			continue // Job already exists
		}

		// Try to find the metadata file for this video
		metadataPattern := filepath.Join(baseDir, "*", fmt.Sprintf("*%s*.info.json", item.ID))
		matches, err := filepath.Glob(metadataPattern)
		if err != nil || len(matches) == 0 {
			// Try alternative pattern
			metadataPattern = filepath.Join(baseDir, "*", "*.info.json")
			allMatches, err := filepath.Glob(metadataPattern)
			if err != nil {
				continue
			}

			// Search through all metadata files for this video ID
			var found bool
			for _, match := range allMatches {
				data, err := os.ReadFile(match)
				if err != nil {
					continue
				}

				var videoInfo map[string]interface{}
				if err := json.Unmarshal(data, &videoInfo); err != nil {
					continue
				}

				if videoID, ok := videoInfo["id"].(string); ok && videoID == item.ID {
					matches = []string{match}
					found = true
					break
				}
			}
			
			if !found {
				log.Warnf("Could not find metadata file for video %s", item.ID)
				continue
			}
		}

		// Create the video job using the first (best) match
		if len(matches) > 0 {
			metadataFilePath := matches[0]
			
			// Extract extractor from video URL or assume YouTube
			extractor := "youtube"
			if item.ChannelURL != "" && strings.Contains(item.ChannelURL, "youtube.com") {
				extractor = "youtube"
			}

			// Create video job
			videoJob := domain.Job{
				ID:        item.ID,
				URL:       fmt.Sprintf("https://%s.com/watch?v=%s", extractor, item.ID),
				Status:    domain.JobStatusComplete,
				Progress:  100.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create the job
			if err := s.jobs.Create(&videoJob); err != nil {
				log.WithError(err).Warnf("Failed to create migrated job for video %s", item.ID)
				continue
			}

			// Extract and store metadata
			videoMetadata, err := metadata.ExtractMetadata(metadataFilePath)
			if err != nil {
				log.WithError(err).Warnf("Failed to extract metadata for video %s during migration", item.ID)
				continue
			}

			if err := s.jobs.StoreMetadata(item.ID, videoMetadata); err != nil {
				log.WithError(err).Warnf("Failed to store metadata for video %s during migration", item.ID)
				continue
			}

			// Link the video to the playlist/channel
			membershipType := "playlist"
			if err := s.jobs.AddVideoToParent(item.ID, playlistJobID, membershipType); err != nil {
				log.WithError(err).Warnf("Failed to link video %s to playlist %s during migration", item.ID, playlistJobID)
			} else {
				log.Infof("Successfully linked video %s to playlist %s during migration", item.ID, playlistJobID)
			}

			count++
			log.Debugf("Successfully migrated video %s (%s)", item.ID, item.Title)
		}
	}

	return count, nil
}

func (s *Service) migrateChannelVideos(channelJobID string, channelMetadata *domain.ChannelMetadata) (int, error) {
	if channelMetadata.RecentVideos == nil || len(channelMetadata.RecentVideos) == 0 {
		log.Warnf("Channel %s has no recent videos to migrate", channelJobID)
		return 0, nil
	}

	count := 0
	baseDir := s.config.DownloadPath

	for _, item := range channelMetadata.RecentVideos {
		if item.ID == "" {
			continue
		}

		// Check if job already exists
		existingJob, err := s.jobs.GetByID(item.ID)
		if err == nil && existingJob != nil {
			continue // Job already exists
		}

		// Try to find the metadata file for this video
		metadataPattern := filepath.Join(baseDir, "*", fmt.Sprintf("*%s*.info.json", item.ID))
		matches, err := filepath.Glob(metadataPattern)
		if err != nil || len(matches) == 0 {
			// Try alternative pattern
			metadataPattern = filepath.Join(baseDir, "*", "*.info.json")
			allMatches, err := filepath.Glob(metadataPattern)
			if err != nil {
				continue
			}

			// Search through all metadata files for this video ID
			var found bool
			for _, match := range allMatches {
				data, err := os.ReadFile(match)
				if err != nil {
					continue
				}

				var videoInfo map[string]interface{}
				if err := json.Unmarshal(data, &videoInfo); err != nil {
					continue
				}

				if videoID, ok := videoInfo["id"].(string); ok && videoID == item.ID {
					matches = []string{match}
					found = true
					break
				}
			}
			
			if !found {
				log.Warnf("Could not find metadata file for video %s", item.ID)
				continue
			}
		}

		// Create the video job using the first (best) match
		if len(matches) > 0 {
			metadataFilePath := matches[0]
			
			// Extract extractor from channel URL or assume YouTube
			extractor := "youtube"
			if channelMetadata.URL != "" && strings.Contains(channelMetadata.URL, "youtube.com") {
				extractor = "youtube"
			}

			// Create video job
			videoJob := domain.Job{
				ID:        item.ID,
				URL:       fmt.Sprintf("https://%s.com/watch?v=%s", extractor, item.ID),
				Status:    domain.JobStatusComplete,
				Progress:  100.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create the job
			if err := s.jobs.Create(&videoJob); err != nil {
				log.WithError(err).Warnf("Failed to create migrated job for video %s", item.ID)
				continue
			}

			// Extract and store metadata
			videoMetadata, err := metadata.ExtractMetadata(metadataFilePath)
			if err != nil {
				log.WithError(err).Warnf("Failed to extract metadata for video %s during migration", item.ID)
				continue
			}

			if err := s.jobs.StoreMetadata(item.ID, videoMetadata); err != nil {
				log.WithError(err).Warnf("Failed to store metadata for video %s during migration", item.ID)
				continue
			}

			// Link the video to the channel
			membershipType := "channel"
			if err := s.jobs.AddVideoToParent(item.ID, channelJobID, membershipType); err != nil {
				log.WithError(err).Warnf("Failed to link video %s to channel %s during migration", item.ID, channelJobID)
			} else {
				log.Infof("Successfully linked video %s to channel %s during migration", item.ID, channelJobID)
			}

			count++
			log.Debugf("Successfully migrated video %s (%s)", item.ID, item.Title)
		}
	}

	return count, nil
}