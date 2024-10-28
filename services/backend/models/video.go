package models

import "time"

// Video represents metadata for a single video.
type Video struct {
	ID               int       `json:"id"`                 // Primary key in database
	JobID            string    `json:"job_id"`             // ID associated with the download job
	Title            string    `json:"title"`              // Title of the video
	Uploader         string    `json:"uploader"`           // Uploader or channel name
	FilePath         string    `json:"file_path"`          // Path where the video file is stored
	LastDownloadedAt time.Time `json:"last_downloaded_at"` // Timestamp of when the video was last downloaded
	Length           float64   `json:"length"`             // Duration of the video in seconds
	Size             int64     `json:"size"`               // File size in bytes
	Quality          string    `json:"quality"`            // Video quality
}

// Playlist represents metadata for a playlist or channel.
type Playlist struct {
	ID          string `json:"id"`          // Playlist or channel ID
	Title       string `json:"title"`       // Title of the playlist or channel
	Description string `json:"description"` // Description of the playlist or channel
}
