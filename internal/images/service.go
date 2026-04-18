package images

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Service handles image generation
type Service struct {
	repo     *Repository
	provider Provider
}

// NewService creates a new image service
func NewService(repo *Repository, provider Provider) *Service {
	return &Service{
		repo:     repo,
		provider: provider,
	}
}

// Generate generates an image for a scene
func (s *Service) Generate(ctx context.Context, projectID, sceneID uuid.UUID, prompt, aspectRatio, styleProfile string) (*models.Asset, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithSceneID(ctx, sceneID.String())
	ctx = logger.WithStage(ctx, "image_generation")
	ctx = logger.WithProvider(ctx, s.provider.Name())

	logger.Info(ctx, "generating image", "prompt_length", len(prompt))

	// Check if image already exists (idempotency)
	existing, err := s.repo.GetBySceneID(ctx, sceneID, models.AssetTypeImage)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "image already exists, returning existing")
		return existing, nil
	}

	// Generate image
	req := contracts.ImageRequest{
		ProjectID:    projectID,
		SceneID:      sceneID,
		Provider:     s.provider.Name(),
		Prompt:       prompt,
		AspectRatio:  aspectRatio,
		StyleProfile: styleProfile,
	}

	result, err := s.provider.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	// Create asset record
	metadata, _ := json.Marshal(result.Metadata)
	asset := &models.Asset{
		ID:         uuid.New(),
		ProjectID:  projectID,
		SceneID:    &sceneID,
		AssetType:  models.AssetTypeImage,
		Provider:   result.Provider,
		StorageKey: result.StorageKey,
		MimeType:   "image/png",
		PromptUsed: &prompt,
		Metadata:   metadata,
	}

	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to save asset: %w", err)
	}

	logger.Info(ctx, "image generated successfully", "storage_key", result.StorageKey)

	return asset, nil
}

// GetBySceneID retrieves the image asset for a scene
func (s *Service) GetBySceneID(ctx context.Context, sceneID uuid.UUID) (*models.Asset, error) {
	return s.repo.GetBySceneID(ctx, sceneID, models.AssetTypeImage)
}

// GetByProjectID retrieves all image assets for a project
func (s *Service) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Asset, error) {
	return s.repo.GetImagesByProjectID(ctx, projectID)
}
