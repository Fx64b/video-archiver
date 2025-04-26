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
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    JobStatus `json:"status"`
	Progress  float64   `json:"progress"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
}

type JobType string

const (
	JobTypeVideo    JobType = "video"
	JobTypeAudio    JobType = "audio"
	JobTypeMetadata JobType = "metadata"
)

type JobWithMetadata struct {
	Job      *Job     `json:"job"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ProgressUpdate struct {
	JobID                string  `json:"jobID"`
	JobType              string  `json:"jobType"`
	CurrentItem          int     `json:"currentItem"`
	TotalItems           int     `json:"totalItems"`
	Progress             float64 `json:"progress"`
	CurrentVideoProgress float64 `json:"currentVideoProgress"`
}

type VideoMetadata struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Thumbnail        string   `json:"thumbnail"`
	Duration         int      `json:"duration"`
	ViewCount        int      `json:"view_count"`
	Channel          string   `json:"channel"`
	ChannelID        string   `json:"channel_id"`
	ChannelURL       string   `json:"channel_url"`
	ChannelFollowers int      `json:"channel_follower_count"`
	Tags             []string `json:"tags"`
	Categories       []string `json:"categories"`
	UploadDate       string   `json:"upload_date"`
	FileSize         int64    `json:"filesize_approx"`
	Type             string   `json:"_type"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
	ID     string `json:"id"`
}

type PlaylistMetadata struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Thumbnails       []Thumbnail `json:"thumbnails"`
	UploaderID       string      `json:"uploader_id"`
	UploaderURL      string      `json:"uploader_url"`
	ChannelID        string      `json:"channel_id"`
	Channel          string      `json:"channel"`
	ChannelURL       string      `json:"channel_url"`
	ChannelFollowers int         `json:"channel_follower_count"`
	ItemCount        int         `json:"playlist_count"`
	Type             string      `json:"_type"`
}

type ChannelMetadata struct {
	ID               string      `json:"id"`
	Channel          string      `json:"channel"`
	URL              string      `json:"channel_url"`
	Description      string      `json:"description"`
	Thumbnails       []Thumbnail `json:"thumbnails"`
	ChannelFollowers int         `json:"channel_follower_count"`
	PlaylistCount    int         `json:"playlist_count"`
	Type             string      `json:"_type"`
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
