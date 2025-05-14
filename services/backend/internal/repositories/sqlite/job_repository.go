package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
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

	// Check if the record already exists
	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM videos WHERE job_id = ?", jobID).Scan(&count)
	if err != nil {
		return fmt.Errorf("check existing video: %w", err)
	}

	if count > 0 {
		// Update existing record
		_, err = r.db.Exec(`
            UPDATE videos 
            SET title = ?, metadata_json = ? 
            WHERE job_id = ?`,
			metadata.Title, string(metadataJSON), jobID)
		if err != nil {
			return fmt.Errorf("update video metadata: %w", err)
		}
	} else {
		// Insert new record
		_, err = r.db.Exec(`
            INSERT INTO videos (job_id, title, metadata_json)
            VALUES (?, ?, ?)`,
			jobID, metadata.Title, string(metadataJSON))
		if err != nil {
			return fmt.Errorf("insert video metadata: %w", err)
		}
	}

	return nil
}

func (r *JobRepository) storePlaylistMetadata(jobID string, metadata *domain.PlaylistMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	// Check if the record already exists
	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM playlists WHERE job_id = ?", jobID).Scan(&count)
	if err != nil {
		return fmt.Errorf("check existing playlist: %w", err)
	}

	if count > 0 {
		// Update existing record only if items are not empty
		if len(metadata.Items) == 0 {
			log.Info("No items in playlist metadata, skipping update")
			return nil
		}

		_, err = r.db.Exec(`
            UPDATE playlists 
            SET title = ?, metadata_json = ? 
            WHERE job_id = ?`,
			metadata.Title, string(metadataJSON), jobID)
		if err != nil {
			return fmt.Errorf("update playlist metadata: %w", err)
		}
	} else {
		// Insert new record
		_, err = r.db.Exec(`
            INSERT INTO playlists (job_id, title, metadata_json)
            VALUES (?, ?, ?)`,
			jobID, metadata.Title, string(metadataJSON))
		if err != nil {
			return fmt.Errorf("insert playlist metadata: %w", err)
		}
	}

	return nil
}

func (r *JobRepository) storeChannelMetadata(jobID string, metadata *domain.ChannelMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	// Check if the record already exists
	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM channels WHERE job_id = ?", jobID).Scan(&count)
	if err != nil {
		return fmt.Errorf("check existing channel: %w", err)
	}

	if count > 0 {
		// Update existing record
		_, err = r.db.Exec(`
            UPDATE channels 
            SET name = ?, metadata_json = ? 
            WHERE job_id = ?`,
			metadata.Channel, string(metadataJSON), jobID)
		if err != nil {
			return fmt.Errorf("update channel metadata: %w", err)
		}
	} else {
		// Insert new record
		_, err = r.db.Exec(`
            INSERT INTO channels (job_id, name, metadata_json)
            VALUES (?, ?, ?)`,
			jobID, metadata.Channel, string(metadataJSON))
		if err != nil {
			return fmt.Errorf("insert channel metadata: %w", err)
		}
	}

	return nil
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
	log.WithFields(log.Fields{
		"contentType": contentType,
		"page":        page,
		"limit":       limit,
		"sortBy":      sortBy,
		"order":       order,
	}).Debug("Getting metadata by type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Validate content type using a strict whitelist
	var tableName string
	switch contentType {
	case "videos":
		tableName = "videos"
	case "playlists":
		tableName = "playlists"
	case "channels":
		tableName = "channels"
	default:
		return nil, 0, fmt.Errorf("invalid content type: %s", contentType)
	}

	// Validate order with a strict whitelist
	var orderDirection string
	switch strings.ToLower(order) {
	case "asc":
		orderDirection = "ASC"
	default:
		orderDirection = "DESC"
	}

	// Create a whitelist mapping of allowed sort fields to their actual SQL counterparts
	var sortFieldMapping = map[string]map[string]string{
		"videos": {
			"created_at": "jobs.created_at",
			"updated_at": "jobs.updated_at",
			"title":      "videos.title",
		},
		"playlists": {
			"created_at": "jobs.created_at",
			"updated_at": "jobs.updated_at",
			"title":      "playlists.title",
		},
		"channels": {
			"created_at": "jobs.created_at",
			"updated_at": "jobs.updated_at",
			"title":      "channels.name",
		},
	}

	// Validate sort field - only allow predefined values
	validSortFields, exists := sortFieldMapping[contentType]
	if !exists {
		return nil, 0, fmt.Errorf("invalid content type for sort: %s", contentType)
	}

	sortField, valid := validSortFields[sortBy]
	if !valid {
		// Default to created_at if invalid
		sortField = "jobs.created_at"
	}

	// Get total count first using specific validated table name
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM " + tableName + " JOIN jobs ON " + tableName + ".job_id = jobs.job_id"

	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", contentType, err)
	}

	// Build the query using only validated table names and sort fields
	query := `
        SELECT jobs.job_id, jobs.url, jobs.status, jobs.progress, jobs.created_at, jobs.updated_at, ` +
		tableName + `.metadata_json 
        FROM ` + tableName + ` 
        JOIN jobs ON ` + tableName + `.job_id = jobs.job_id 
        ORDER BY ` + sortField + ` ` + orderDirection + ` 
        LIMIT ? OFFSET ?`

	log.Debugf("Executing download query with limit=%d offset=%d", limit, offset)
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s: %w", contentType, err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.WithError(err).Error("Failed to close rows")
		}
	}(rows)

	// The rest of the function to process rows remains unchanged
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

	log.WithFields(log.Fields{
		"contentType": contentType,
		"resultCount": len(result),
		"totalCount":  totalCount,
	}).Debug("Fetched metadata by type")

	return result, totalCount, nil
}

func (r *JobRepository) AddVideoToParent(videoJobID, parentJobID, membershipType string) error {
	// Check if the relationship already exists
	var count int
	err := r.db.QueryRow(`
        SELECT COUNT(*) FROM video_memberships 
        WHERE video_job_id = ? AND parent_job_id = ?`,
		videoJobID, parentJobID).Scan(&count)

	if err != nil {
		return fmt.Errorf("check existing membership: %w", err)
	}

	if count > 0 {
		// Relationship already exists
		return nil
	}

	// Insert the new relationship
	_, err = r.db.Exec(`
        INSERT INTO video_memberships (video_job_id, parent_job_id, membership_type)
        VALUES (?, ?, ?)`,
		videoJobID, parentJobID, membershipType)

	if err != nil {
		return fmt.Errorf("insert video membership: %w", err)
	}

	return nil
}
func (r *JobRepository) GetParentsForVideo(videoJobID string) ([]*domain.JobWithMetadata, error) {
	rows, err := r.db.Query(`
        WITH parent_types AS (
            SELECT parent_job_id, membership_type
            FROM video_memberships
            WHERE video_job_id = ?
        )
        SELECT j.job_id, j.url, j.status, j.progress, j.created_at, j.updated_at, 
               pt.membership_type,
               CASE 
                   WHEN pt.membership_type = 'playlist' THEN p.metadata_json
                   WHEN pt.membership_type = 'channel' THEN c.metadata_json
                   ELSE NULL
               END as metadata_json
        FROM jobs j
        JOIN parent_types pt ON j.job_id = pt.parent_job_id
        LEFT JOIN playlists p ON j.job_id = p.job_id AND pt.membership_type = 'playlist'
        LEFT JOIN channels c ON j.job_id = c.job_id AND pt.membership_type = 'channel'`,
		videoJobID)

	if err != nil {
		return nil, fmt.Errorf("query parents for video: %w", err)
	}
	defer rows.Close()

	var result []*domain.JobWithMetadata

	for rows.Next() {
		job := &domain.Job{}
		var membershipType string
		var metadataJSON sql.NullString

		err := rows.Scan(
			&job.ID, &job.URL, &job.Status, &job.Progress,
			&job.CreatedAt, &job.UpdatedAt, &membershipType, &metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("scan job row: %w", err)
		}

		jobWithMetadata := &domain.JobWithMetadata{
			Job: job,
		}

		if metadataJSON.Valid {
			var metadata domain.Metadata

			switch membershipType {
			case "playlist":
				var playlistMetadata domain.PlaylistMetadata
				if err := json.Unmarshal([]byte(metadataJSON.String), &playlistMetadata); err == nil {
					metadata = &playlistMetadata
				}
			case "channel":
				var channelMetadata domain.ChannelMetadata
				if err := json.Unmarshal([]byte(metadataJSON.String), &channelMetadata); err == nil {
					metadata = &channelMetadata
				}
			}

			jobWithMetadata.Metadata = metadata
		}

		result = append(result, jobWithMetadata)
	}

	return result, nil
}

func (r *JobRepository) GetVideosForParent(parentJobID string) ([]*domain.JobWithMetadata, error) {
	rows, err := r.db.Query(`
        SELECT j.job_id, j.url, j.status, j.progress, j.created_at, j.updated_at, 
               v.metadata_json
        FROM jobs j
        JOIN video_memberships vm ON j.job_id = vm.video_job_id
        JOIN videos v ON j.job_id = v.job_id
        WHERE vm.parent_job_id = ?`,
		parentJobID)

	if err != nil {
		return nil, err
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
			return nil, err
		}

		var videoMetadata domain.VideoMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &videoMetadata); err != nil {
			log.WithError(err).Warn("Failed to unmarshal video metadata")
			continue
		}

		result = append(result, &domain.JobWithMetadata{
			Job:      job,
			Metadata: &videoMetadata,
		})
	}

	return result, nil
}
