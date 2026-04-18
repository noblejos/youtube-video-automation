package images

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Repository handles asset persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new asset repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new asset record
func (r *Repository) Create(ctx context.Context, asset *models.Asset) error {
	if asset.ID == uuid.Nil {
		asset.ID = uuid.New()
	}
	asset.CreatedAt = time.Now()

	query := `
		INSERT INTO assets (
			id, project_id, scene_id, asset_type, provider, storage_key,
			mime_type, source_url, prompt_used, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.pool.Exec(ctx, query,
		asset.ID,
		asset.ProjectID,
		asset.SceneID,
		asset.AssetType,
		asset.Provider,
		asset.StorageKey,
		asset.MimeType,
		asset.SourceURL,
		asset.PromptUsed,
		asset.Metadata,
		asset.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	return nil
}

// GetBySceneID retrieves the primary asset for a scene
func (r *Repository) GetBySceneID(ctx context.Context, sceneID uuid.UUID, assetType string) (*models.Asset, error) {
	query := `
		SELECT id, project_id, scene_id, asset_type, provider, storage_key,
			mime_type, source_url, prompt_used, metadata, created_at
		FROM assets WHERE scene_id = $1 AND asset_type = $2
		ORDER BY created_at DESC LIMIT 1
	`

	var asset models.Asset
	err := r.pool.QueryRow(ctx, query, sceneID, assetType).Scan(
		&asset.ID,
		&asset.ProjectID,
		&asset.SceneID,
		&asset.AssetType,
		&asset.Provider,
		&asset.StorageKey,
		&asset.MimeType,
		&asset.SourceURL,
		&asset.PromptUsed,
		&asset.Metadata,
		&asset.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return &asset, nil
}

// GetByProjectID retrieves all assets for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Asset, error) {
	query := `
		SELECT a.id, a.project_id, a.scene_id, a.asset_type, a.provider, a.storage_key,
			a.mime_type, a.source_url, a.prompt_used, a.metadata, a.created_at
		FROM assets a
		LEFT JOIN scenes s ON a.scene_id = s.id
		WHERE a.project_id = $1
		ORDER BY s.scene_number NULLS LAST, a.created_at
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.ProjectID,
			&asset.SceneID,
			&asset.AssetType,
			&asset.Provider,
			&asset.StorageKey,
			&asset.MimeType,
			&asset.SourceURL,
			&asset.PromptUsed,
			&asset.Metadata,
			&asset.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// GetImagesByProjectID retrieves all image assets for a project
func (r *Repository) GetImagesByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Asset, error) {
	query := `
		SELECT a.id, a.project_id, a.scene_id, a.asset_type, a.provider, a.storage_key,
			a.mime_type, a.source_url, a.prompt_used, a.metadata, a.created_at
		FROM assets a
		LEFT JOIN scenes s ON a.scene_id = s.id
		WHERE a.project_id = $1 AND a.asset_type = $2
		ORDER BY s.scene_number NULLS LAST
	`

	rows, err := r.pool.Query(ctx, query, projectID, models.AssetTypeImage)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.ProjectID,
			&asset.SceneID,
			&asset.AssetType,
			&asset.Provider,
			&asset.StorageKey,
			&asset.MimeType,
			&asset.SourceURL,
			&asset.PromptUsed,
			&asset.Metadata,
			&asset.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// DeleteBySceneID deletes assets for a scene
func (r *Repository) DeleteBySceneID(ctx context.Context, sceneID uuid.UUID) error {
	query := `DELETE FROM assets WHERE scene_id = $1`

	_, err := r.pool.Exec(ctx, query, sceneID)
	if err != nil {
		return fmt.Errorf("failed to delete assets: %w", err)
	}

	return nil
}

// CountImagesByProjectID counts image assets for a project
func (r *Repository) CountImagesByProjectID(ctx context.Context, projectID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM assets WHERE project_id = $1 AND asset_type = $2`

	var count int
	err := r.pool.QueryRow(ctx, query, projectID, models.AssetTypeImage).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count assets: %w", err)
	}

	return count, nil
}
