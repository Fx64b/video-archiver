package domain

import "time"

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusComplete   JobStatus = "complete"
	JobStatusError      JobStatus = "error"
)

type Job struct {
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	Status        JobStatus `json:"status"`
	Progress      float64   `json:"progress"`
	CustomQuality *int      `json:"custom_quality,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

//tygo:ignore
type JobRepository interface {
	Create(job *Job) error
	Update(job *Job) error
	GetByID(id string) (*Job, error)
	GetRecent(limit int) ([]*Job, error)
	StoreMetadata(jobID string, metadata Metadata) error
	GetJobWithMetadata(jobID string) (*JobWithMetadata, error)
	GetRecentWithMetadata(limit int) ([]*JobWithMetadata, error)
	GetAllJobsWithMetadata() ([]*JobWithMetadata, error)
	GetJobs() (jobs []*Job, err error)
	CountVideos() (int, error)
	CountPlaylists() (int, error)
	CountChannels() (int, error)
	GetMetadataByType(contentType string, page int, limit int, sortBy string, order string) ([]*JobWithMetadata, int, error)
	AddVideoToParent(videoJobID, parentJobID, membershipType string) error
	GetVideosForParent(parentJobID string) ([]*JobWithMetadata, error)
	GetParentsForVideo(videoJobID string) ([]*JobWithMetadata, error)
}

type JobType string

const (
	JobTypeVideo    JobType = "video"
	JobTypeAudio    JobType = "audio"
	JobTypeMetadata JobType = "metadata"
)

// deprecated, remove in the future
type JobWithMetadata struct {
	Job      *Job     `json:"job"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ProgressUpdate struct {
	JobID                string    `json:"jobID"`
	JobType              string    `json:"jobType"`
	Status               JobStatus `json:"status,omitempty"`
	CurrentItem          int       `json:"currentItem"`
	TotalItems           int       `json:"totalItems"`
	Progress             float64   `json:"progress"`
	CurrentVideoProgress float64   `json:"currentVideoProgress"`
	DownloadPhase        string    `json:"downloadPhase"`
	IsRetrying           bool      `json:"isRetrying,omitempty"`
	RetryCount           int       `json:"retryCount,omitempty"`
	MaxRetries           int       `json:"maxRetries,omitempty"`
	RetryError           string    `json:"retryError,omitempty"`
}

const (
	DownloadPhaseMetadata = "metadata"
	DownloadPhaseVideo    = "video"
	DownloadPhaseAudio    = "audio"
	DownloadPhaseMerging  = "merging"
	DownloadPhaseComplete = "complete"
)

type VideoMetadata struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Thumbnail         string   `json:"thumbnail"`
	Duration          int      `json:"duration"`
	DurationString    string   `json:"duration_string"`
	ViewCount         int      `json:"view_count"`
	LikeCount         int      `json:"like_count"`
	CommentCount      int      `json:"comment_count"`
	Channel           string   `json:"channel"`
	ChannelID         string   `json:"channel_id"`
	ChannelURL        string   `json:"channel_url"`
	ChannelFollowers  int      `json:"channel_follower_count"`
	ChannelIsVerified bool     `json:"channel_is_verified"`
	Uploader          string   `json:"uploader"`
	UploaderID        string   `json:"uploader_id"`
	UploaderURL       string   `json:"uploader_url"`
	Tags              []string `json:"tags"`
	Categories        []string `json:"categories"`
	UploadDate        string   `json:"upload_date"`
	FileSize          int64    `json:"filesize_approx"`
	Format            string   `json:"format"`
	Extension         string   `json:"ext"`
	Language          string   `json:"language"`
	Width             int      `json:"width"`
	Height            int      `json:"height"`
	Resolution        string   `json:"resolution"`
	FPS               float64  `json:"fps"`
	DynamicRange      string   `json:"dynamic_range"`
	VideoCodec        string   `json:"vcodec"`
	AspectRatio       float64  `json:"aspect_ratio"`
	AudioCodec        string   `json:"acodec"`
	AudioChannels     int      `json:"audio_channels"`
	WasLive           bool     `json:"was_live"`
	WebpageURLDomain  string   `json:"webpage_url_domain"`
	Extractor         string   `json:"extractor"`
	FullTitle         string   `json:"fulltitle"`
	Type              string   `json:"_type"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
	ID     string `json:"id"`
}

type PlaylistItem struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Thumbnail      string   `json:"thumbnail,omitempty"`
	Duration       int      `json:"duration,omitempty"`
	DurationString string   `json:"duration_string,omitempty"`
	UploadDate     string   `json:"upload_date,omitempty"`
	ViewCount      int      `json:"view_count,omitempty"`
	LikeCount      int      `json:"like_count,omitempty"`
	VideoFile      string   `json:"video_file,omitempty"`
	Channel        string   `json:"channel,omitempty"`
	ChannelID      string   `json:"channel_id,omitempty"`
	ChannelURL     string   `json:"channel_url,omitempty"`
	Width          int      `json:"width,omitempty"`
	Height         int      `json:"height,omitempty"`
	Resolution     string   `json:"resolution,omitempty"`
	FileSize       int64    `json:"filesize_approx,omitempty"`
	Format         string   `json:"format,omitempty"`
	Extension      string   `json:"ext,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type PlaylistMetadata struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Thumbnails       []Thumbnail    `json:"thumbnails"`
	UploaderID       string         `json:"uploader_id"`
	UploaderURL      string         `json:"uploader_url"`
	ChannelID        string         `json:"channel_id"`
	Channel          string         `json:"channel"`
	ChannelURL       string         `json:"channel_url"`
	ChannelFollowers int            `json:"channel_follower_count"`
	ItemCount        int            `json:"playlist_count"`
	ViewCount        int            `json:"view_count,omitempty"`
	Items            []PlaylistItem `json:"items,omitempty"`
	Type             string         `json:"_type"`
}

type ChannelMetadata struct {
	ID               string         `json:"id"`
	Channel          string         `json:"channel"`
	URL              string         `json:"channel_url"`
	Description      string         `json:"description"`
	Thumbnails       []Thumbnail    `json:"thumbnails"`
	ChannelFollowers int            `json:"channel_follower_count"`
	PlaylistCount    int            `json:"playlist_count"`
	Type             string         `json:"_type"`
	VideoCount       int            `json:"video_count,omitempty"`
	TotalStorage     int64          `json:"total_storage,omitempty"`
	TotalViews       int            `json:"total_views,omitempty"`
	RecentVideos     []PlaylistItem `json:"recent_videos,omitempty"`
}

type MetadataUpdate struct {
	JobID    string   `json:"jobID"`
	Metadata Metadata `json:"metadata"`
}

type Metadata interface {
	GetType() string
}

func (p *PlaylistMetadata) GetType() string {
	return "playlist"
}

func (v *VideoMetadata) GetType() string {
	return "video"
}

func (c *ChannelMetadata) GetType() string {
	return "channel"
}
