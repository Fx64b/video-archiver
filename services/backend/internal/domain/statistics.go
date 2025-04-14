package domain

import "time"

type Statistics struct {
	TotalJobs       int       `json:"total_jobs"`
	TotalVideos     int       `json:"total_videos"`
	TotalPlaylists  int       `json:"total_playlists"`
	TotalChannels   int       `json:"total_channels"`
	TotalStorage    int       `json:"total_storage"`
	VideosStorage   int       `json:"videos_storage"`
	PlaylistStorage int       `json:"playlist_storage"`
	ChannelStorage  int       `json:"channel_storage"`
	LastUpdate      time.Time `json:"last_update"`
}
