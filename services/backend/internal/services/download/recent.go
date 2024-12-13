package download

import (
	log "github.com/sirupsen/logrus"
	"video-archiver/internal/storage"
	"video-archiver/models"
)

func GetRecentJobs() []models.JobData {
	var recentJobs, _ = storage.GetRecentJobs()

	if len(recentJobs) == 0 {
		log.Println("No recent downloads found.")
		return nil
	}

	return recentJobs
}
