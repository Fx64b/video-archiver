package domain

import "time"

type Statistics struct {
	TotalJobs      int                `json:"total_jobs"`
	TotalVideos    int                `json:"total_videos"`
	TotalPlaylists int                `json:"total_playlists"`
	TotalChannels  int                `json:"total_channels"`
	TotalStorage   int                `json:"total_storage"`
	TopVideos      []VideoStorageInfo `json:"top_videos"`
	OtherStorage   int                `json:"other_storage"`
	LastUpdate     time.Time          `json:"last_update"`
}

type VideoStorageInfo struct {
	Title   string `json:"title"`
	Size    int    `json:"size"`
	Channel string `json:"channel"`
}
