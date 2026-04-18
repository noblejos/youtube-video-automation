package scripts

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Repository handles script persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new script repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new script
func (r *Repository) Create(ctx context.Context, script *models.Script) error {
	if script.ID == uuid.Nil {
		script.ID = uuid.New()
	}
	script.CreatedAt = time.Now()
	script.UpdatedAt = time.Now()

	query := `
		INSERT INTO scripts (
			id, project_id, hook, setup_text, build_text, turning_point_text,
			collapse_text, conclusion_text, full_script, raw_model_response,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.pool.Exec(ctx, query,
		script.ID,
		script.ProjectID,
		script.Hook,
		script.SetupText,
		script.BuildText,
		script.TurningPointText,
		script.CollapseText,
		script.ConclusionText,
		script.FullScript,
		script.RawModelResponse,
		script.CreatedAt,
		script.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create script: %w", err)
	}

	return nil
}

// GetByProjectID retrieves the script for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Script, error) {
	query := `
		SELECT id, project_id, hook, setup_text, build_text, turning_point_text,
			collapse_text, conclusion_text, full_script, raw_model_response,
			created_at, updated_at
		FROM scripts WHERE project_id = $1
	`

	var script models.Script
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&script.ID,
		&script.ProjectID,
		&script.Hook,
		&script.SetupText,
		&script.BuildText,
		&script.TurningPointText,
		&script.CollapseText,
		&script.ConclusionText,
		&script.FullScript,
		&script.RawModelResponse,
		&script.CreatedAt,
		&script.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get script: %w", err)
	}

	return &script, nil
}

// Update updates an existing script
func (r *Repository) Update(ctx context.Context, script *models.Script) error {
	script.UpdatedAt = time.Now()

	query := `
		UPDATE scripts SET
			hook = $2, setup_text = $3, build_text = $4, turning_point_text = $5,
			collapse_text = $6, conclusion_text = $7, full_script = $8,
			raw_model_response = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		script.ID,
		script.Hook,
		script.SetupText,
		script.BuildText,
		script.TurningPointText,
		script.CollapseText,
		script.ConclusionText,
		script.FullScript,
		script.RawModelResponse,
		script.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update script: %w", err)
	}

	return nil
}

// Delete deletes a script
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM scripts WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete script: %w", err)
	}

	return nil
}
