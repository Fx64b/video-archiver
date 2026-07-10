package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (r *CollectionRepository) Create(collection *domain.Collection) error {
	_, err := r.db.Exec(`
        INSERT INTO collections (id, name, description, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)`,
		collection.ID, collection.Name, collection.Description,
		collection.CreatedAt, collection.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) Update(collection *domain.Collection) error {
	collection.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
        UPDATE collections
        SET name = ?, description = ?, updated_at = ?
        WHERE id = ?`,
		collection.Name, collection.Description, collection.UpdatedAt, collection.ID)
	if err != nil {
		return fmt.Errorf("update collection: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("collection not found")
	}
	return nil
}

// Delete removes a collection and its memberships. Member videos themselves
// are untouched — a collection only references downloads, it does not own them.
func (r *CollectionRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete collection: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM collection_videos WHERE collection_id = ?`, id); err != nil {
		return fmt.Errorf("delete collection memberships: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM collections WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	return tx.Commit()
}

// collectionSelect returns collections enriched with their member count and
// the thumbnail of their first member, so listings need no extra queries.
const collectionSelect = `
    SELECT c.id, c.name, c.description, c.created_at, c.updated_at,
           (SELECT COUNT(*) FROM collection_videos cv WHERE cv.collection_id = c.id) AS video_count,
           COALESCE((
               SELECT json_extract(v.metadata_json, '$.thumbnail')
               FROM collection_videos cv
               JOIN videos v ON v.job_id = cv.video_job_id
               WHERE cv.collection_id = c.id
               ORDER BY cv.position ASC, cv.rowid ASC
               LIMIT 1
           ), '') AS thumbnail
    FROM collections c`

func scanCollection(row interface{ Scan(...any) error }) (*domain.Collection, error) {
	c := &domain.Collection{}
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt,
		&c.VideoCount, &c.Thumbnail)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CollectionRepository) GetByID(id string) (*domain.Collection, error) {
	c, err := scanCollection(r.db.QueryRow(collectionSelect+` WHERE c.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get collection by id: %w", err)
	}
	return c, nil
}

func (r *CollectionRepository) List() ([]*domain.Collection, error) {
	rows, err := r.db.Query(collectionSelect + ` ORDER BY c.name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	collections := []*domain.Collection{}
	for rows.Next() {
		c, err := scanCollection(rows)
		if err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		collections = append(collections, c)
	}
	return collections, rows.Err()
}

// AddVideos appends videos to the end of a collection. IDs that are already
// members are skipped (the membership keeps its position); IDs that do not
// reference a downloaded video are rejected so a collection can never contain
// playlists, channels or dangling jobs.
func (r *CollectionRepository) AddVideos(collectionID string, videoJobIDs []string) error {
	collection, err := r.GetByID(collectionID)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin add videos: %w", err)
	}
	defer tx.Rollback()

	var position int
	if err := tx.QueryRow(`
        SELECT COALESCE(MAX(position), 0) FROM collection_videos
        WHERE collection_id = ?`, collectionID).Scan(&position); err != nil {
		return fmt.Errorf("next collection position: %w", err)
	}

	for _, videoJobID := range videoJobIDs {
		var exists int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM videos WHERE job_id = ?`, videoJobID).Scan(&exists); err != nil {
			return fmt.Errorf("check video %s: %w", videoJobID, err)
		}
		if exists == 0 {
			return fmt.Errorf("job %s is not a downloaded video", videoJobID)
		}

		position++
		res, err := tx.Exec(`
            INSERT OR IGNORE INTO collection_videos (collection_id, video_job_id, position)
            VALUES (?, ?, ?)`, collectionID, videoJobID, position)
		if err != nil {
			return fmt.Errorf("add video %s to collection: %w", videoJobID, err)
		}
		// An already-present video consumes no position slot.
		if n, err := res.RowsAffected(); err == nil && n == 0 {
			position--
		}
	}

	if _, err := tx.Exec(`UPDATE collections SET updated_at = ? WHERE id = ?`,
		time.Now(), collectionID); err != nil {
		return fmt.Errorf("touch collection: %w", err)
	}
	return tx.Commit()
}

func (r *CollectionRepository) RemoveVideo(collectionID string, videoJobID string) error {
	_, err := r.db.Exec(`
        DELETE FROM collection_videos
        WHERE collection_id = ? AND video_job_id = ?`, collectionID, videoJobID)
	if err != nil {
		return fmt.Errorf("remove video from collection: %w", err)
	}
	if _, err := r.db.Exec(`UPDATE collections SET updated_at = ? WHERE id = ?`,
		time.Now(), collectionID); err != nil {
		return fmt.Errorf("touch collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) GetVideos(collectionID string) ([]*domain.JobWithMetadata, error) {
	rows, err := r.db.Query(`
        SELECT j.job_id, j.url, j.status, j.progress, j.warnings, j.file_path,
               j.created_at, j.updated_at, v.metadata_json
        FROM collection_videos cv
        JOIN jobs j ON j.job_id = cv.video_job_id
        JOIN videos v ON v.job_id = cv.video_job_id
        WHERE cv.collection_id = ?
        ORDER BY cv.position ASC, cv.rowid ASC`, collectionID)
	if err != nil {
		return nil, fmt.Errorf("get collection videos: %w", err)
	}
	defer rows.Close()

	result := []*domain.JobWithMetadata{}
	for rows.Next() {
		job := &domain.Job{}
		var metadataJSON string
		var warningsJSON, filePath sql.NullString

		err := rows.Scan(&job.ID, &job.URL, &job.Status, &job.Progress,
			&warningsJSON, &filePath, &job.CreatedAt, &job.UpdatedAt, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("scan collection video: %w", err)
		}

		if warningsJSON.Valid && warningsJSON.String != "" {
			if err := json.Unmarshal([]byte(warningsJSON.String), &job.Warnings); err != nil {
				log.WithError(err).Warn("Failed to unmarshal warnings")
				job.Warnings = []string{}
			}
		}
		job.FilePath = filePath.String

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
	return result, rows.Err()
}

func (r *CollectionRepository) ListForVideo(videoJobID string) ([]string, error) {
	rows, err := r.db.Query(`
        SELECT collection_id FROM collection_videos
        WHERE video_job_id = ?`, videoJobID)
	if err != nil {
		return nil, fmt.Errorf("list collections for video: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan collection id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
