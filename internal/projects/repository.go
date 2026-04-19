package projects

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Repository handles project persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new project repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new project
func (r *Repository) Create(ctx context.Context, project *models.Project) error {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	query := `
		INSERT INTO projects (
			id, external_id, topic, title, channel_style, target_duration_sec,
			aspect_ratio, voice_id, voice_engine, status, review_required, current_step, error_message,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.pool.Exec(ctx, query,
		project.ID,
		project.ExternalID,
		project.Topic,
		project.Title,
		project.ChannelStyle,
		project.TargetDurationSec,
		project.AspectRatio,
		project.VoiceID,
		project.VoiceEngine,
		project.Status,
		project.ReviewRequired,
		project.CurrentStep,
		project.ErrorMessage,
		project.CreatedAt,
		project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetByID retrieves a project by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, external_id, topic, title, channel_style, target_duration_sec,
			aspect_ratio, voice_id, voice_engine, status, review_required, current_step, error_message,
			created_at, updated_at
		FROM projects WHERE id = $1
	`

	var project models.Project
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.ExternalID,
		&project.Topic,
		&project.Title,
		&project.ChannelStyle,
		&project.TargetDurationSec,
		&project.AspectRatio,
		&project.VoiceID,
		&project.VoiceEngine,
		&project.Status,
		&project.ReviewRequired,
		&project.CurrentStep,
		&project.ErrorMessage,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetByExternalID retrieves a project by external ID
func (r *Repository) GetByExternalID(ctx context.Context, externalID string) (*models.Project, error) {
	query := `
		SELECT id, external_id, topic, title, channel_style, target_duration_sec,
			aspect_ratio, voice_id, voice_engine, status, review_required, current_step, error_message,
			created_at, updated_at
		FROM projects WHERE external_id = $1
	`

	var project models.Project
	err := r.pool.QueryRow(ctx, query, externalID).Scan(
		&project.ID,
		&project.ExternalID,
		&project.Topic,
		&project.Title,
		&project.ChannelStyle,
		&project.TargetDurationSec,
		&project.AspectRatio,
		&project.VoiceID,
		&project.VoiceEngine,
		&project.Status,
		&project.ReviewRequired,
		&project.CurrentStep,
		&project.ErrorMessage,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", externalID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// UpdateStatus updates the project status
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, currentStep *string) error {
	query := `
		UPDATE projects SET status = $2, current_step = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, status, currentStep, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}

// UpdateTitle updates the project title
func (r *Repository) UpdateTitle(ctx context.Context, id uuid.UUID, title string) error {
	query := `UPDATE projects SET title = $2, updated_at = $3 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id, title, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update project title: %w", err)
	}

	return nil
}

// SetError sets an error message on the project
func (r *Repository) SetError(ctx context.Context, id uuid.UUID, errorMsg string) error {
	query := `
		UPDATE projects SET status = $2, error_message = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, models.ProjectStatusFailed, errorMsg, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set project error: %w", err)
	}

	return nil
}

// List lists all projects with optional status filter
func (r *Repository) List(ctx context.Context, status string, limit, offset int) ([]*models.Project, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, external_id, topic, title, channel_style, target_duration_sec,
				aspect_ratio, voice_id, voice_engine, status, review_required, current_step, error_message,
				created_at, updated_at
			FROM projects WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, external_id, topic, title, channel_style, target_duration_sec,
				aspect_ratio, voice_id, voice_engine, status, review_required, current_step, error_message,
				created_at, updated_at
			FROM projects
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID,
			&project.ExternalID,
			&project.Topic,
			&project.Title,
			&project.ChannelStyle,
			&project.TargetDurationSec,
			&project.AspectRatio,
			&project.VoiceID,
			&project.VoiceEngine,
			&project.Status,
			&project.ReviewRequired,
			&project.CurrentStep,
			&project.ErrorMessage,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, &project)
	}

	return projects, nil
}
