package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Repository handles job persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new job repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new job
func (r *Repository) Create(ctx context.Context, job *models.Job) error {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	query := `
		INSERT INTO jobs (
			id, project_id, job_type, status, attempt_count, max_attempts,
			payload, result, error_message, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.pool.Exec(ctx, query,
		job.ID,
		job.ProjectID,
		job.JobType,
		job.Status,
		job.AttemptCount,
		job.MaxAttempts,
		job.Payload,
		job.Result,
		job.ErrorMessage,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// Update updates a job
func (r *Repository) Update(ctx context.Context, job *models.Job) error {
	job.UpdatedAt = time.Now()

	query := `
		UPDATE jobs SET
			status = $2, attempt_count = $3, result = $4, error_message = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		job.ID,
		job.Status,
		job.AttemptCount,
		job.Result,
		job.ErrorMessage,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	query := `
		SELECT id, project_id, job_type, status, attempt_count, max_attempts,
			payload, result, error_message, created_at, updated_at
		FROM jobs WHERE id = $1
	`

	var job models.Job
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.ProjectID,
		&job.JobType,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.Payload,
		&job.Result,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("job not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// GetByProjectID retrieves all jobs for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Job, error) {
	query := `
		SELECT id, project_id, job_type, status, attempt_count, max_attempts,
			payload, result, error_message, created_at, updated_at
		FROM jobs WHERE project_id = $1
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.JobType,
			&job.Status,
			&job.AttemptCount,
			&job.MaxAttempts,
			&job.Payload,
			&job.Result,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// GetByProjectIDAndType retrieves jobs by project and type
func (r *Repository) GetByProjectIDAndType(ctx context.Context, projectID uuid.UUID, jobType string) ([]*models.Job, error) {
	query := `
		SELECT id, project_id, job_type, status, attempt_count, max_attempts,
			payload, result, error_message, created_at, updated_at
		FROM jobs WHERE project_id = $1 AND job_type = $2
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query, projectID, jobType)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.JobType,
			&job.Status,
			&job.AttemptCount,
			&job.MaxAttempts,
			&job.Payload,
			&job.Result,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// CountByStatus counts jobs by status for a project
func (r *Repository) CountByStatus(ctx context.Context, projectID uuid.UUID, jobType string) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) FROM jobs
		WHERE project_id = $1 AND job_type = $2
		GROUP BY status
	`

	rows, err := r.pool.Query(ctx, query, projectID, jobType)
	if err != nil {
		return nil, fmt.Errorf("failed to count jobs: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		counts[status] = count
	}

	return counts, nil
}

// GetFailedJobs retrieves failed jobs for a project
func (r *Repository) GetFailedJobs(ctx context.Context, projectID uuid.UUID) ([]*models.Job, error) {
	query := `
		SELECT id, project_id, job_type, status, attempt_count, max_attempts,
			payload, result, error_message, created_at, updated_at
		FROM jobs WHERE project_id = $1 AND status = $2
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, query, projectID, models.JobStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.JobType,
			&job.Status,
			&job.AttemptCount,
			&job.MaxAttempts,
			&job.Payload,
			&job.Result,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}
