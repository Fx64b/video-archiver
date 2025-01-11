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
}

type JobType string

const (
	JobTypeVideo    JobType = "video"
	JobTypeAudio    JobType = "audio"
	JobTypeMetadata JobType = "metadata"
)

type ProgressUpdate struct {
	JobID                string  `json:"jobID"`
	JobType              string  `json:"jobType"`
	CurrentItem          int     `json:"currentItem"`
	TotalItems           int     `json:"totalItems"`
	Progress             float64 `json:"progress"`
	CurrentVideoProgress float64 `json:"currentVideoProgress"`
}
