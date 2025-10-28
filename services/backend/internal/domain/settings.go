package domain

import "time"

type Settings struct {
	ID                    int       `json:"id"`
	Theme                 string    `json:"theme"`
	DownloadQuality       int       `json:"download_quality"`
	ConcurrentDownloads   int       `json:"concurrent_downloads"`
	ToolsDefaultFormat    string    `json:"tools_default_format"`
	ToolsDefaultQuality   string    `json:"tools_default_quality"`
	ToolsPreserveOriginal bool      `json:"tools_preserve_original"`
	ToolsOutputPath       string    `json:"tools_output_path"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

//tygo:ignore
type SettingsRepository interface {
	Get() (*Settings, error)
	Update(settings *Settings) error
}
