package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Generator defines the interface for script generation
type Generator interface {
	Generate(ctx context.Context, topic, channelStyle string, targetDuration int) (*contracts.ScriptResponse, error)
}

// Service handles script generation and management
type Service struct {
	repo      *Repository
	generator Generator
}

// NewService creates a new script service
func NewService(repo *Repository, generator Generator) *Service {
	return &Service{
		repo:      repo,
		generator: generator,
	}
}

// Generate generates a script for a project
func (s *Service) Generate(ctx context.Context, projectID uuid.UUID, topic, channelStyle string, targetDuration int) (*models.Script, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "script_generation")

	logger.Info(ctx, "generating script", "topic", topic, "style", channelStyle)

	// Check if script already exists (idempotency)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "script already exists, returning existing")
		return existing, nil
	}

	// Generate script
	response, err := s.generator.Generate(ctx, topic, channelStyle, targetDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate script: %w", err)
	}

	// Marshal raw response
	rawResponse, _ := json.Marshal(response)

	// Create script record
	script := &models.Script{
		ID:               uuid.New(),
		ProjectID:        projectID,
		Hook:             response.Hook,
		SetupText:        response.SetupText,
		BuildText:        response.BuildText,
		TurningPointText: response.TurningPointText,
		CollapseText:     response.CollapseText,
		ConclusionText:   response.ConclusionText,
		FullScript:       response.FullScript,
		RawModelResponse: rawResponse,
	}

	if err := s.repo.Create(ctx, script); err != nil {
		return nil, fmt.Errorf("failed to save script: %w", err)
	}

	logger.Info(ctx, "script generated successfully")
	return script, nil
}

// GetByProjectID retrieves the script for a project
func (s *Service) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Script, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}

// SaveUserScript saves a user-provided script directly (skips AI generation)
func (s *Service) SaveUserScript(ctx context.Context, projectID uuid.UUID, scriptText, title string) (*models.Script, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "script_save")

	logger.Info(ctx, "saving user-provided script")

	// Check if script already exists (idempotency)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "script already exists, returning existing")
		return existing, nil
	}

	// For user-provided scripts, we store the full text
	// The scene generator will parse it into scenes
	script := &models.Script{
		ID:         uuid.New(),
		ProjectID:  projectID,
		FullScript: scriptText,
		// Leave other fields empty - they're for AI-generated structured scripts
		Hook:             "",
		SetupText:        "",
		BuildText:        "",
		TurningPointText: "",
		CollapseText:     "",
		ConclusionText:   "",
	}

	if err := s.repo.Create(ctx, script); err != nil {
		return nil, fmt.Errorf("failed to save script: %w", err)
	}

	logger.Info(ctx, "user script saved successfully", "length", len(scriptText))
	return script, nil
}

// MockGenerator is a mock script generator for testing
type MockGenerator struct{}

// NewMockGenerator creates a new mock script generator
func NewMockGenerator() Generator {
	return &MockGenerator{}
}

// Generate generates a mock script
func (g *MockGenerator) Generate(ctx context.Context, topic, channelStyle string, targetDuration int) (*contracts.ScriptResponse, error) {
	// Generate a structured mock script based on the topic
	hook := fmt.Sprintf("What if I told you that %s changed the course of history?", strings.ToLower(topic))
	setup := fmt.Sprintf("In the ancient world, %s emerged as one of the most significant forces of its time.", topic)
	build := fmt.Sprintf("As years passed, %s grew in power and influence, reshaping everything around it.", topic)
	turningPoint := "But then, everything changed. A single event would alter the course of history forever."
	collapse := "The decline was swift and devastating. What once seemed invincible crumbled into dust."
	conclusion := fmt.Sprintf("And that is the remarkable story of %s. A reminder that nothing lasts forever.", topic)

	fullScript := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s",
		hook, setup, build, turningPoint, collapse, conclusion)

	return &contracts.ScriptResponse{
		Title:            fmt.Sprintf("The Rise and Fall of %s", topic),
		Hook:             hook,
		SetupText:        setup,
		BuildText:        build,
		TurningPointText: turningPoint,
		CollapseText:     collapse,
		ConclusionText:   conclusion,
		FullScript:       fullScript,
	}, nil
}
