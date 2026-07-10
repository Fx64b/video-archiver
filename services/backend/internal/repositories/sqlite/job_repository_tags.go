package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"video-archiver/internal/domain"
)

// ListTags returns the tag catalog with how many jobs carry each tag, most
// used first.
func (r *JobRepository) ListTags() ([]domain.Tag, error) {
	rows, err := r.db.Query(`
        SELECT t.id, t.name, COUNT(jt.job_id) as usage_count
        FROM tags t
        LEFT JOIN job_tags jt ON jt.tag_id = t.id
        GROUP BY t.id, t.name
        ORDER BY usage_count DESC, t.name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	tags := []domain.Tag{}
	for rows.Next() {
		var tag domain.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Count); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *JobRepository) GetTagsForJob(jobID string) ([]domain.Tag, error) {
	rows, err := r.db.Query(`
        SELECT t.id, t.name, jt.source
        FROM job_tags jt
        JOIN tags t ON t.id = jt.tag_id
        WHERE jt.job_id = ?
        ORDER BY jt.source ASC, t.name COLLATE NOCASE ASC`, jobID)
	if err != nil {
		return nil, fmt.Errorf("get tags for job: %w", err)
	}
	defer rows.Close()

	tags := []domain.Tag{}
	for rows.Next() {
		var tag domain.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Source); err != nil {
			return nil, fmt.Errorf("scan job tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// AddTagsToJob attaches the named tags to a job, creating missing tags on the
// fly. Names are matched case-insensitively; already-attached tags are left
// untouched (a user tag is not downgraded to auto). Returns the job's full tag
// list afterwards.
func (r *JobRepository) AddTagsToJob(jobID string, names []string, source string) ([]domain.Tag, error) {
	if source != domain.TagSourceAuto {
		source = domain.TagSourceUser
	}

	for _, name := range names {
		name = domain.NormalizeTagName(name)
		if name == "" {
			continue
		}
		tagID, err := r.ensureTag(name)
		if err != nil {
			return nil, err
		}
		if _, err := r.db.Exec(`
            INSERT OR IGNORE INTO job_tags (job_id, tag_id, source)
            VALUES (?, ?, ?)`, jobID, tagID, source); err != nil {
			return nil, fmt.Errorf("attach tag %q: %w", name, err)
		}
	}
	return r.GetTagsForJob(jobID)
}

// ensureTag returns the ID of the named tag, creating it if needed.
func (r *JobRepository) ensureTag(name string) (int64, error) {
	var id int64
	err := r.db.QueryRow(`SELECT id FROM tags WHERE name = ? COLLATE NOCASE`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("look up tag %q: %w", name, err)
	}

	res, err := r.db.Exec(`INSERT INTO tags (name) VALUES (?)`, name)
	if err != nil {
		return 0, fmt.Errorf("create tag %q: %w", name, err)
	}
	return res.LastInsertId()
}

// RemoveTagFromJob detaches a tag from a job and removes the tag entirely once
// nothing references it, so the catalog does not fill with orphans.
func (r *JobRepository) RemoveTagFromJob(jobID string, tagID int64) error {
	if _, err := r.db.Exec(`DELETE FROM job_tags WHERE job_id = ? AND tag_id = ?`, jobID, tagID); err != nil {
		return fmt.Errorf("remove tag from job: %w", err)
	}
	if _, err := r.db.Exec(`
        DELETE FROM tags WHERE id = ?
        AND NOT EXISTS (SELECT 1 FROM job_tags WHERE tag_id = ?)`, tagID, tagID); err != nil {
		return fmt.Errorf("prune orphan tag: %w", err)
	}
	return nil
}

// applyAutoTags derives and attaches automatic tags for freshly stored
// metadata. Failures are logged rather than propagated: tagging must never
// fail a download.
func (r *JobRepository) applyAutoTags(jobID string, metadata domain.Metadata) {
	names := domain.AutoTagsFor(metadata)
	if len(names) == 0 {
		return
	}
	if _, err := r.AddTagsToJob(jobID, names, domain.TagSourceAuto); err != nil {
		log.WithError(err).WithField("jobID", jobID).Warn("Failed to apply auto tags")
	}
}

// BackfillAutoTags applies auto-tagging to every item that was downloaded
// before tagging existed. It is idempotent and safe to run at every startup.
func (r *JobRepository) BackfillAutoTags() error {
	jobs, err := r.GetAllJobsWithMetadata()
	if err != nil {
		return fmt.Errorf("load jobs for tag backfill: %w", err)
	}
	for _, jwm := range jobs {
		if jwm == nil || jwm.Job == nil || jwm.Metadata == nil {
			continue
		}
		r.applyAutoTags(jwm.Job.ID, jwm.Metadata)
	}
	return nil
}

// DeleteJob removes a job and everything referencing it: metadata records,
// playlist/channel memberships and tag assignments. Orphaned tags are pruned
// afterwards. Files on disk are the caller's responsibility.
func (r *JobRepository) DeleteJob(jobID string) error {
	statements := []string{
		`DELETE FROM videos WHERE job_id = ?`,
		`DELETE FROM playlists WHERE job_id = ?`,
		`DELETE FROM channels WHERE job_id = ?`,
		`DELETE FROM video_memberships WHERE video_job_id = ? OR parent_job_id = ?`,
		`DELETE FROM collection_videos WHERE video_job_id = ?`,
		`DELETE FROM job_tags WHERE job_id = ?`,
		`DELETE FROM jobs WHERE job_id = ?`,
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range statements {
		args := []any{jobID}
		if strings.Contains(stmt, "OR parent_job_id") {
			args = append(args, jobID)
		}
		if _, err := tx.Exec(stmt, args...); err != nil {
			return fmt.Errorf("delete job %s: %w", jobID, err)
		}
	}
	if _, err := tx.Exec(`
        DELETE FROM tags WHERE NOT EXISTS (
            SELECT 1 FROM job_tags WHERE job_tags.tag_id = tags.id
        )`); err != nil {
		return fmt.Errorf("prune orphan tags: %w", err)
	}

	return tx.Commit()
}
