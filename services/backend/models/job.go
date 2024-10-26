package models

import "time"

type DownloadJob struct {
	ID        string
	URL       string
	TIMESTAMP time.Time
}
