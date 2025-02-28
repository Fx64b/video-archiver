package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"video-archiver/internal/domain"
)

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *domain.Job) error {
	_, err := r.db.Exec(`
        INSERT INTO jobs (job_id, url, status, progress, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		job.ID, job.URL, job.Status, job.Progress, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (r *JobRepository) Update(job *domain.Job) error {
	job.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
        UPDATE jobs 
        SET status = ?, progress = ?, updated_at = ? 
        WHERE job_id = ?`,
		job.Status, job.Progress, job.UpdatedAt, job.ID)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	return nil
}

func (r *JobRepository) GetByID(id string) (*domain.Job, error) {
	job := &domain.Job{}
	err := r.db.QueryRow(`
        SELECT job_id, url, status, progress, created_at, updated_at 
        FROM jobs 
        WHERE job_id = ?`, id).
		Scan(&job.ID, &job.URL, &job.Status, &job.Progress, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get job by id: %w", err)
	}
	return job, nil
}

func (r *JobRepository) GetRecent(limit int) ([]*domain.Job, error) {
	rows, err := r.db.Query(`
        SELECT job_id, url, status, progress, created_at, updated_at 
        FROM jobs 
        ORDER BY updated_at DESC 
        LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.Job
	for rows.Next() {
		job := &domain.Job{}
		err := rows.Scan(&job.ID, &job.URL, &job.Status, &job.Progress,
			&job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan job row: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *JobRepository) StoreMetadata(jobID string, metadata domain.Metadata) error {
	switch m := metadata.(type) {
	case *domain.VideoMetadata:
		return r.storeVideoMetadata(jobID, m)
	case *domain.PlaylistMetadata:
		return r.storePlaylistMetadata(m)
	default:
		return fmt.Errorf("unsupported metadata type: %T", metadata)
	}
}

func (r *JobRepository) storeVideoMetadata(jobID string, metadata *domain.VideoMetadata) error {
	_, err := r.db.Exec(`
        INSERT OR IGNORE INTO channels (id, name, url, follower_count)
        VALUES (?, ?, ?, ?)`,
		metadata.ChannelID, metadata.Channel, metadata.ChannelURL, metadata.ChannelFollowers)
	if err != nil {
		return fmt.Errorf("store channel: %w", err)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = r.db.Exec(`
        INSERT INTO videos (job_id, title, channel_id, metadata_json)
        VALUES (?, ?, ?, ?)`,
		jobID, metadata.Title, metadata.ChannelID, string(metadataJSON))

	return err
}

func (r *JobRepository) storePlaylistMetadata(metadata *domain.PlaylistMetadata) error {
	_, err := r.db.Exec(`
        INSERT OR IGNORE INTO channels (id, name, url, follower_count)
        VALUES (?, ?, ?, ?)`,
		metadata.ChannelID, metadata.Channel, metadata.ChannelURL, metadata.ChannelFollowers)
	if err != nil {
		return fmt.Errorf("store channel: %w", err)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = r.db.Exec(`
        INSERT INTO playlists (id, title, description, metadata_json)
        VALUES (?, ?, ?, ?)`,
		metadata.ID, metadata.Title, metadata.Description, string(metadataJSON))

	return err
}
