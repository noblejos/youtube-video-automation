package voice

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Service handles voice generation
type Service struct {
	repo     *Repository
	provider Provider
}

// NewService creates a new voice service
func NewService(repo *Repository, provider Provider) *Service {
	return &Service{
		repo:     repo,
		provider: provider,
	}
}

// Generate generates audio for a scene
func (s *Service) Generate(ctx context.Context, projectID, sceneID uuid.UUID, text, voiceID, engine string) (*models.AudioFile, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithSceneID(ctx, sceneID.String())
	ctx = logger.WithStage(ctx, "voice_generation")
	ctx = logger.WithProvider(ctx, s.provider.Name())

	logger.Info(ctx, "generating voice audio")

	// Check if audio already exists (idempotency)
	existing, err := s.repo.GetBySceneID(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "audio already exists, returning existing")
		return existing, nil
	}

	// Determine text type
	textType := "text"
	if len(text) > 7 && text[:7] == "<speak>" {
		textType = "ssml"
	}

	// Generate audio
	req := contracts.VoiceRequest{
		ProjectID: projectID,
		SceneID:   sceneID,
		Provider:  s.provider.Name(),
		VoiceID:   voiceID,
		Engine:    engine,
		TextType:  textType,
		Text:      text,
	}

	result, err := s.provider.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate audio: %w", err)
	}

	// Create audio file record
	audio := &models.AudioFile{
		ID:         uuid.New(),
		ProjectID:  projectID,
		SceneID:    sceneID,
		Provider:   s.provider.Name(),
		VoiceID:    voiceID,
		Engine:     engine,
		StorageKey: result.StorageKey,
		DurationMs: &result.DurationMs,
	}

	if result.SpeechMarksKey != "" {
		audio.SpeechMarksKey = &result.SpeechMarksKey
	}

	if err := s.repo.Create(ctx, audio); err != nil {
		return nil, fmt.Errorf("failed to save audio file: %w", err)
	}

	logger.Info(ctx, "voice audio generated successfully",
		"storage_key", result.StorageKey,
		"duration_ms", result.DurationMs,
	)

	return audio, nil
}

// GetBySceneID retrieves the audio file for a scene
func (s *Service) GetBySceneID(ctx context.Context, sceneID uuid.UUID) (*models.AudioFile, error) {
	return s.repo.GetBySceneID(ctx, sceneID)
}

// GetByProjectID retrieves all audio files for a project
func (s *Service) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.AudioFile, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}
