package models

import "time"

type DownloadJob struct {
	ID        string
	URL       string
	TIMESTAMP time.Time
}

type JobData struct {
	ID         string
	JobID      string
	URL        string
	IsPlaylist bool
	STATUS     string
	PROGRESS   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
