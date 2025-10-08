package domain

import "time"

type Settings struct {
	ID                  int       `json:"id"`
	Theme               string    `json:"theme"`
	DownloadQuality     int       `json:"download_quality"`
	ConcurrentDownloads int       `json:"concurrent_downloads"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

//tygo:ignore
type SettingsRepository interface {
	Get() (*Settings, error)
	Update(settings *Settings) error
}
