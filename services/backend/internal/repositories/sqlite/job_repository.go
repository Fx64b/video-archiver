package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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
		return r.storePlaylistMetadata(jobID, m)
	case *domain.ChannelMetadata:
		return r.storeChannelMetadata(jobID, m)
	default:
		return fmt.Errorf("unsupported metadata type: %T", metadata)
	}
}

func (r *JobRepository) storeVideoMetadata(jobID string, metadata *domain.VideoMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = r.db.Exec(`
        INSERT INTO videos (job_id, title, metadata_json)
        VALUES (?, ?, ?)`,
		jobID, metadata.Title, string(metadataJSON))

	return err
}

func (r *JobRepository) storePlaylistMetadata(jobID string, metadata *domain.PlaylistMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = r.db.Exec(`
        INSERT INTO playlists (job_id, title, metadata_json)
        VALUES ( ?, ?, ?)`,
		jobID, metadata.Title, string(metadataJSON))

	return err
}

func (r *JobRepository) storeChannelMetadata(jobID string, metadata *domain.ChannelMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = r.db.Exec(`
        INSERT INTO channels (job_id, name, metadata_json)
        VALUES (?, ?, ?)`,
		jobID, metadata.Channel, string(metadataJSON))

	return err
}

func (r *JobRepository) GetJobWithMetadata(jobID string) (*domain.JobWithMetadata, error) {
	job, err := r.GetByID(jobID)
	if err != nil {
		return nil, fmt.Errorf("get job by id: %w", err)
	}

	metadata, err := r.getMetadataForJob(jobID)
	if err != nil {
		log.WithError(err).Warnf("Could not retrieve metadata for job %s", jobID)
		// Continue without metadata
	}

	return &domain.JobWithMetadata{
		Job:      job,
		Metadata: metadata,
	}, nil
}

func (r *JobRepository) GetRecentWithMetadata(limit int) ([]*domain.JobWithMetadata, error) {
	jobs, err := r.GetRecent(limit)
	if err != nil {
		return nil, fmt.Errorf("get recent jobs: %w", err)
	}

	result := make([]*domain.JobWithMetadata, 0, len(jobs))
	for _, job := range jobs {
		metadata, err := r.getMetadataForJob(job.ID)
		if err != nil {
			log.WithError(err).Warnf("Could not retrieve metadata for job %s", job.ID)
			// Still add the job without metadata
			result = append(result, &domain.JobWithMetadata{
				Job: job,
			})
			continue
		}

		result = append(result, &domain.JobWithMetadata{
			Job:      job,
			Metadata: metadata,
		})
	}

	return result, nil
}

func (r *JobRepository) GetAllJobsWithMetadata() ([]*domain.JobWithMetadata, error) {
	jobs, err := r.GetJobs()
	if err != nil {
		return nil, fmt.Errorf("get jobs: %w", err)
	}

	result := make([]*domain.JobWithMetadata, 0, len(jobs))
	for _, job := range jobs {
		metadata, err := r.getMetadataForJob(job.ID)
		if err != nil {
			log.WithError(err).Warnf("Failed to get metadata for job %s", job.ID)
			// Continue without metadata
			result = append(result, &domain.JobWithMetadata{
				Job: job,
			})
			continue
		}

		result = append(result, &domain.JobWithMetadata{
			Job:      job,
			Metadata: metadata,
		})
	}

	return result, nil
}

func (r *JobRepository) GetJobs() ([]*domain.Job, error) {
	rows, err := r.db.Query(`
		SELECT job_id, url, status, progress, created_at, updated_at 
		FROM jobs`)
	if err != nil {
		return nil, fmt.Errorf("get jobs: %w", err)
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

func (r *JobRepository) CountVideos() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM videos").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count videos: %w", err)
	}
	return count, nil
}

func (r *JobRepository) CountPlaylists() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM playlists").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count playlists: %w", err)
	}
	return count, nil
}

func (r *JobRepository) CountChannels() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count channels: %w", err)
	}
	return count, nil
}

func (r *JobRepository) getMetadataForJob(jobID string) (domain.Metadata, error) {
	var metadataJSON string
	var metadataType string

	err := r.db.QueryRow(`
        SELECT 'video' as type, metadata_json FROM videos WHERE job_id = ?
        UNION
        SELECT 'playlist' as type, metadata_json FROM playlists WHERE job_id = ?
        UNION
        SELECT 'channel' as type, metadata_json FROM channels WHERE job_id = ?
        LIMIT 1
    `, jobID, jobID, jobID).Scan(&metadataType, &metadataJSON)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query metadata: %w", err)
	}

	switch metadataType {
	case "video":
		var metadata domain.VideoMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal video metadata: %w", err)
		}
		return &metadata, nil
	case "playlist":
		var metadata domain.PlaylistMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal playlist metadata: %w", err)
		}
		return &metadata, nil
	case "channel":
		var metadata domain.ChannelMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal channel metadata: %w", err)
		}
		return &metadata, nil
	default:
		return nil, fmt.Errorf("unknown metadata type: %s", metadataType)
	}
}

func (r *JobRepository) GetMetadataByType(contentType string, page int, limit int, sortBy string, order string) ([]*domain.JobWithMetadata, int, error) {
	// Validate and sanitize parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 { // 100 for now. TODO: Later this can be a config
		limit = 20
	}
	offset := (page - 1) * limit

	// Validate sort parameters
	validSortFields := map[string]bool{
		"created_at": true, // Default sort field
		"updated_at": true,
		"title":      true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at" // Default if invalid
	}

	if order != "asc" && order != "desc" {
		order = "desc" // Default to descending order
	}

	// Map content type to table name
	var tableName string
	var sortField string

	switch contentType {
	case "videos":
		tableName = "videos"
		if sortBy == "title" {
			sortField = "videos.title"
		} else {
			sortField = "jobs." + sortBy
		}
	case "playlists":
		tableName = "playlists"
		if sortBy == "title" {
			sortField = "playlists.title"
		} else {
			sortField = "jobs." + sortBy
		}
	case "channels":
		tableName = "channels"
		if sortBy == "title" {
			sortField = "channels.name"
		} else {
			sortField = "jobs." + sortBy
		}
	default:
		return nil, 0, fmt.Errorf("invalid content type: %s", contentType)
	}

	// Get total count first
	var totalCount int
	countQuery := `
        SELECT COUNT(*) 
        FROM ` + tableName + `
        JOIN jobs ON ` + tableName + `.job_id = jobs.job_id
    `

	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", contentType, err)
	}

	// Query for paginated items
	query := fmt.Sprintf(`
        SELECT jobs.job_id, jobs.url, jobs.status, jobs.progress, jobs.created_at, jobs.updated_at, %s.metadata_json
        FROM %s
        JOIN jobs ON %s.job_id = jobs.job_id
        ORDER BY %s %s
        LIMIT ? OFFSET ?
    `, tableName, tableName, tableName, sortField, order)

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s: %w", contentType, err)
	}
	defer rows.Close()

	var result []*domain.JobWithMetadata
	for rows.Next() {
		job := &domain.Job{}
		var metadataJSON string

		err := rows.Scan(
			&job.ID, &job.URL, &job.Status, &job.Progress,
			&job.CreatedAt, &job.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan job row: %w", err)
		}

		// Unmarshal metadata based on content type
		var metadata domain.Metadata
		switch contentType {
		case "videos":
			var videoMetadata domain.VideoMetadata
			if err := json.Unmarshal([]byte(metadataJSON), &videoMetadata); err != nil {
				log.WithError(err).Warnf("Could not unmarshal video metadata for job %s", job.ID)
				continue
			}
			metadata = &videoMetadata
		case "playlists":
			var playlistMetadata domain.PlaylistMetadata
			if err := json.Unmarshal([]byte(metadataJSON), &playlistMetadata); err != nil {
				log.WithError(err).Warnf("Could not unmarshal playlist metadata for job %s", job.ID)
				continue
			}
			metadata = &playlistMetadata
		case "channels":
			var channelMetadata domain.ChannelMetadata
			if err := json.Unmarshal([]byte(metadataJSON), &channelMetadata); err != nil {
				log.WithError(err).Warnf("Could not unmarshal channel metadata for job %s", job.ID)
				continue
			}
			metadata = &channelMetadata
		}

		result = append(result, &domain.JobWithMetadata{
			Job:      job,
			Metadata: metadata,
		})
	}

	return result, totalCount, nil
}
