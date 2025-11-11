package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"video-archiver/internal/domain"
)

type ToolsRepository struct {
	db *sql.DB
}

func NewToolsRepository(db *sql.DB) *ToolsRepository {
	return &ToolsRepository{db: db}
}

func (r *ToolsRepository) Create(job *domain.ToolsJob) error {
	inputFilesJSON, err := json.Marshal(job.InputFiles)
	if err != nil {
		return fmt.Errorf("marshal input files: %w", err)
	}

	paramsJSON, err := json.Marshal(job.Parameters)
	if err != nil {
		return fmt.Errorf("marshal parameters: %w", err)
	}

	_, err = r.db.Exec(`
        INSERT INTO tools_jobs (id, operation_type, status, progress, input_files, input_type,
                                 output_file, parameters, error_message, created_at, updated_at,
                                 completed_at, estimated_size, actual_size)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.OperationType, job.Status, job.Progress, string(inputFilesJSON), job.InputType,
		job.OutputFile, string(paramsJSON), job.ErrorMessage, job.CreatedAt, job.UpdatedAt,
		job.CompletedAt, job.EstimatedSize, job.ActualSize)
	if err != nil {
		return fmt.Errorf("create tools job: %w", err)
	}
	return nil
}

func (r *ToolsRepository) Update(job *domain.ToolsJob) error {
	inputFilesJSON, err := json.Marshal(job.InputFiles)
	if err != nil {
		return fmt.Errorf("marshal input files: %w", err)
	}

	paramsJSON, err := json.Marshal(job.Parameters)
	if err != nil {
		return fmt.Errorf("marshal parameters: %w", err)
	}

	job.UpdatedAt = time.Now()

	_, err = r.db.Exec(`
        UPDATE tools_jobs
        SET operation_type = ?, status = ?, progress = ?, input_files = ?, input_type = ?,
            output_file = ?, parameters = ?, error_message = ?, updated_at = ?,
            completed_at = ?, estimated_size = ?, actual_size = ?
        WHERE id = ?`,
		job.OperationType, job.Status, job.Progress, string(inputFilesJSON), job.InputType,
		job.OutputFile, string(paramsJSON), job.ErrorMessage, job.UpdatedAt,
		job.CompletedAt, job.EstimatedSize, job.ActualSize, job.ID)
	if err != nil {
		return fmt.Errorf("update tools job: %w", err)
	}
	return nil
}

func (r *ToolsRepository) GetByID(id string) (*domain.ToolsJob, error) {
	job := &domain.ToolsJob{}
	var inputFilesJSON, paramsJSON string

	err := r.db.QueryRow(`
        SELECT id, operation_type, status, progress, input_files, input_type,
               output_file, parameters, error_message, created_at, updated_at,
               completed_at, estimated_size, actual_size
        FROM tools_jobs
        WHERE id = ?`, id).
		Scan(&job.ID, &job.OperationType, &job.Status, &job.Progress, &inputFilesJSON, &job.InputType,
			&job.OutputFile, &paramsJSON, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt,
			&job.CompletedAt, &job.EstimatedSize, &job.ActualSize)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tools job by id: %w", err)
	}

	if err := json.Unmarshal([]byte(inputFilesJSON), &job.InputFiles); err != nil {
		return nil, fmt.Errorf("unmarshal input files: %w", err)
	}

	if err := json.Unmarshal([]byte(paramsJSON), &job.Parameters); err != nil {
		return nil, fmt.Errorf("unmarshal parameters: %w", err)
	}

	return job, nil
}

func (r *ToolsRepository) GetAll() ([]*domain.ToolsJob, error) {
	rows, err := r.db.Query(`
        SELECT id, operation_type, status, progress, input_files, input_type,
               output_file, parameters, error_message, created_at, updated_at,
               completed_at, estimated_size, actual_size
        FROM tools_jobs
        ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query all tools jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

func (r *ToolsRepository) GetByStatus(status domain.ToolsJobStatus) ([]*domain.ToolsJob, error) {
	rows, err := r.db.Query(`
        SELECT id, operation_type, status, progress, input_files, input_type,
               output_file, parameters, error_message, created_at, updated_at,
               completed_at, estimated_size, actual_size
        FROM tools_jobs
        WHERE status = ?
        ORDER BY created_at DESC`, status)
	if err != nil {
		return nil, fmt.Errorf("query tools jobs by status: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

func (r *ToolsRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM tools_jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete tools job: %w", err)
	}
	return nil
}

func (r *ToolsRepository) List(page int, limit int, status string, operationType string) ([]*domain.ToolsJob, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	if operationType != "" {
		where += " AND operation_type = ?"
		args = append(args, operationType)
	}

	// Get total count
	var totalCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tools_jobs %s", where)
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("count tools jobs: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
        SELECT id, operation_type, status, progress, input_files, input_type,
               output_file, parameters, error_message, created_at, updated_at,
               completed_at, estimated_size, actual_size
        FROM tools_jobs
        %s
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?`, where)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query tools jobs list: %w", err)
	}
	defer rows.Close()

	jobs, err := r.scanJobs(rows)
	if err != nil {
		return nil, 0, err
	}

	return jobs, totalCount, nil
}

func (r *ToolsRepository) scanJobs(rows *sql.Rows) ([]*domain.ToolsJob, error) {
	var jobs []*domain.ToolsJob

	for rows.Next() {
		job := &domain.ToolsJob{}
		var inputFilesJSON, paramsJSON string

		err := rows.Scan(&job.ID, &job.OperationType, &job.Status, &job.Progress, &inputFilesJSON, &job.InputType,
			&job.OutputFile, &paramsJSON, &job.ErrorMessage, &job.CreatedAt, &job.UpdatedAt,
			&job.CompletedAt, &job.EstimatedSize, &job.ActualSize)
		if err != nil {
			return nil, fmt.Errorf("scan tools job: %w", err)
		}

		if err := json.Unmarshal([]byte(inputFilesJSON), &job.InputFiles); err != nil {
			return nil, fmt.Errorf("unmarshal input files: %w", err)
		}

		if err := json.Unmarshal([]byte(paramsJSON), &job.Parameters); err != nil {
			return nil, fmt.Errorf("unmarshal parameters: %w", err)
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tools jobs: %w", err)
	}

	return jobs, nil
}
