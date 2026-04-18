package scenes

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Repository handles scene persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new scene repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new scene
func (r *Repository) Create(ctx context.Context, scene *models.Scene) error {
	if scene.ID == uuid.Nil {
		scene.ID = uuid.New()
	}
	scene.CreatedAt = time.Now()
	scene.UpdatedAt = time.Now()

	query := `
		INSERT INTO scenes (
			id, project_id, scene_number, status, narration_text, ssml_text,
			duration_sec, mood, keywords, visual_prompt, asset_strategy,
			transition_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.pool.Exec(ctx, query,
		scene.ID,
		scene.ProjectID,
		scene.SceneNumber,
		scene.Status,
		scene.NarrationText,
		scene.SSMLText,
		scene.DurationSec,
		scene.Mood,
		scene.Keywords,
		scene.VisualPrompt,
		scene.AssetStrategy,
		scene.TransitionType,
		scene.CreatedAt,
		scene.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create scene: %w", err)
	}

	return nil
}

// CreateBatch creates multiple scenes in a transaction
func (r *Repository) CreateBatch(ctx context.Context, scenes []*models.Scene) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO scenes (
			id, project_id, scene_number, status, narration_text, ssml_text,
			duration_sec, mood, keywords, visual_prompt, asset_strategy,
			transition_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	for _, scene := range scenes {
		if scene.ID == uuid.Nil {
			scene.ID = uuid.New()
		}
		scene.CreatedAt = time.Now()
		scene.UpdatedAt = time.Now()

		_, err := tx.Exec(ctx, query,
			scene.ID,
			scene.ProjectID,
			scene.SceneNumber,
			scene.Status,
			scene.NarrationText,
			scene.SSMLText,
			scene.DurationSec,
			scene.Mood,
			scene.Keywords,
			scene.VisualPrompt,
			scene.AssetStrategy,
			scene.TransitionType,
			scene.CreatedAt,
			scene.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create scene %d: %w", scene.SceneNumber, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a scene by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Scene, error) {
	query := `
		SELECT id, project_id, scene_number, status, narration_text, ssml_text,
			duration_sec, mood, keywords, visual_prompt, asset_strategy,
			transition_type, created_at, updated_at
		FROM scenes WHERE id = $1
	`

	var scene models.Scene
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&scene.ID,
		&scene.ProjectID,
		&scene.SceneNumber,
		&scene.Status,
		&scene.NarrationText,
		&scene.SSMLText,
		&scene.DurationSec,
		&scene.Mood,
		&scene.Keywords,
		&scene.VisualPrompt,
		&scene.AssetStrategy,
		&scene.TransitionType,
		&scene.CreatedAt,
		&scene.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("scene not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get scene: %w", err)
	}

	return &scene, nil
}

// GetByProjectID retrieves all scenes for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Scene, error) {
	query := `
		SELECT id, project_id, scene_number, status, narration_text, ssml_text,
			duration_sec, mood, keywords, visual_prompt, asset_strategy,
			transition_type, created_at, updated_at
		FROM scenes WHERE project_id = $1
		ORDER BY scene_number
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scenes: %w", err)
	}
	defer rows.Close()

	var scenes []*models.Scene
	for rows.Next() {
		var scene models.Scene
		err := rows.Scan(
			&scene.ID,
			&scene.ProjectID,
			&scene.SceneNumber,
			&scene.Status,
			&scene.NarrationText,
			&scene.SSMLText,
			&scene.DurationSec,
			&scene.Mood,
			&scene.Keywords,
			&scene.VisualPrompt,
			&scene.AssetStrategy,
			&scene.TransitionType,
			&scene.CreatedAt,
			&scene.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scene: %w", err)
		}
		scenes = append(scenes, &scene)
	}

	return scenes, nil
}

// UpdateStatus updates the scene status
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE scenes SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update scene status: %w", err)
	}

	return nil
}

// DeleteByProjectID deletes all scenes for a project
func (r *Repository) DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM scenes WHERE project_id = $1`

	_, err := r.pool.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete scenes: %w", err)
	}

	return nil
}

// CountByStatus counts scenes by status for a project
func (r *Repository) CountByStatus(ctx context.Context, projectID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) FROM scenes
		WHERE project_id = $1
		GROUP BY status
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to count scenes: %w", err)
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
