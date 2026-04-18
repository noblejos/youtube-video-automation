package scenes

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

// Generator defines the interface for scene generation
type Generator interface {
	Generate(ctx context.Context, script *models.Script, aspectRatio, styleProfile string, targetDuration int) (*contracts.ScenesResponse, error)
}

// Service handles scene generation and management
type Service struct {
	repo      *Repository
	generator Generator
}

// NewService creates a new scene service
func NewService(repo *Repository, generator Generator) *Service {
	return &Service{
		repo:      repo,
		generator: generator,
	}
}

// Generate generates scenes for a project from a script
func (s *Service) Generate(ctx context.Context, projectID uuid.UUID, script *models.Script, aspectRatio, styleProfile string, targetDuration int) ([]*models.Scene, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "scene_generation")

	logger.Info(ctx, "generating scenes")

	// Check if scenes already exist (idempotency)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		logger.Info(ctx, "scenes already exist, returning existing", "count", len(existing))
		return existing, nil
	}

	// Generate scenes
	response, err := s.generator.Generate(ctx, script, aspectRatio, styleProfile, targetDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate scenes: %w", err)
	}

	// Convert to models
	var sceneModels []*models.Scene
	for _, sceneResp := range response.Scenes {
		keywords, _ := json.Marshal(sceneResp.Keywords)

		scene := &models.Scene{
			ID:            uuid.New(),
			ProjectID:     projectID,
			SceneNumber:   sceneResp.SceneNumber,
			Status:        models.SceneStatusPending,
			StoryRole:     sceneResp.StoryRole,
			EnergyLevel:   sceneResp.EnergyLevel,
			NarrationText: sceneResp.NarrationText,
			DurationSec:   sceneResp.DurationSec,
			StartTimeSec:  sceneResp.StartTimeSec,
			Mood:          sceneResp.Mood,
			Keywords:      keywords,
			AssetStrategy: sceneResp.AssetStrategy,
		}

		// Handle visual config (prefer new structure, fallback to legacy)
		if sceneResp.Visual != nil {
			scene.VisualPrompt = sceneResp.Visual.Prompt
			scene.NegativePrompt = sceneResp.Visual.NegativePrompt
			scene.CameraMotion = sceneResp.Visual.CameraMotion
			scene.TransitionIn = sceneResp.Visual.TransitionIn
			scene.TransitionOut = sceneResp.Visual.TransitionOut
		} else if sceneResp.VisualPrompt != "" {
			scene.VisualPrompt = sceneResp.VisualPrompt
		}

		if sceneResp.SSMLText != "" {
			scene.SSMLText = &sceneResp.SSMLText
		}

		// Store audio config as JSON
		if sceneResp.Audio != nil {
			audioConfig, _ := json.Marshal(sceneResp.Audio)
			scene.AudioConfig = audioConfig
		}

		sceneModels = append(sceneModels, scene)
	}

	// Save scenes in batch
	if err := s.repo.CreateBatch(ctx, sceneModels); err != nil {
		return nil, fmt.Errorf("failed to save scenes: %w", err)
	}

	logger.Info(ctx, "scenes generated successfully", "count", len(sceneModels))
	return sceneModels, nil
}

// GetByProjectID retrieves all scenes for a project
func (s *Service) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Scene, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}

// GetByID retrieves a scene by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Scene, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateStatus updates a scene's status
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// MockGenerator is a mock scene generator for testing
type MockGenerator struct{}

// NewMockGenerator creates a new mock scene generator
func NewMockGenerator() Generator {
	return &MockGenerator{}
}

// Generate generates mock scenes from a script
func (g *MockGenerator) Generate(ctx context.Context, script *models.Script, aspectRatio, styleProfile string, targetDuration int) (*contracts.ScenesResponse, error) {
	// Split script into sections and create scenes
	sections := []struct {
		text     string
		mood     string
		duration float64
	}{
		{script.Hook, "dramatic", 10},
		{script.SetupText, "informative", 18},
		{script.BuildText, "building", 25},
		{script.TurningPointText, "tense", 20},
		{script.CollapseText, "somber", 22},
		{script.ConclusionText, "reflective", 15},
	}

	// Adjust durations to fit target
	totalDuration := 0.0
	for _, s := range sections {
		totalDuration += s.duration
	}
	scale := float64(targetDuration) / totalDuration

	var scenesList []contracts.SceneResponse
	for i, section := range sections {
		if section.text == "" {
			continue
		}

		duration := section.duration * scale
		keywords := extractKeywords(section.text)
		visualPrompt := generateVisualPrompt(section.text, section.mood, styleProfile, aspectRatio)
		ssml := generateSSML(section.text)

		scenesList = append(scenesList, contracts.SceneResponse{
			SceneNumber:   i + 1,
			NarrationText: section.text,
			SSMLText:      ssml,
			DurationSec:   duration,
			Mood:          section.mood,
			Keywords:      keywords,
			VisualPrompt:  visualPrompt,
			AssetStrategy: "ai_or_archive",
		})
	}

	return &contracts.ScenesResponse{
		AspectRatio:  aspectRatio,
		StyleProfile: styleProfile,
		Scenes:       scenesList,
	}, nil
}

func extractKeywords(text string) []string {
	// Simple keyword extraction - in production, use NLP
	words := strings.Fields(strings.ToLower(text))
	keywords := make(map[string]bool)

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"as": true, "is": true, "was": true, "are": true, "were": true,
		"been": true, "be": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"that": true, "which": true, "who": true, "whom": true, "this": true,
		"these": true, "those": true, "it": true, "its": true, "what": true,
		"if": true, "then": true, "than": true, "so": true, "i": true,
		"you": true, "he": true, "she": true, "we": true, "they": true,
		"into": true, "once": true, "seemed": true, "everything": true,
	}

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:'\"")
		if len(word) > 3 && !stopWords[word] {
			keywords[word] = true
		}
	}

	var result []string
	for k := range keywords {
		result = append(result, k)
		if len(result) >= 5 {
			break
		}
	}

	if len(result) == 0 {
		result = []string{"history", "civilization", "ancient"}
	}

	return result
}

func generateVisualPrompt(text, mood, styleProfile, aspectRatio string) string {
	keywords := extractKeywords(text)
	keywordStr := strings.Join(keywords, ", ")

	moodDescriptors := map[string]string{
		"dramatic":    "dramatic lighting, intense atmosphere",
		"informative": "educational documentary style, clear composition",
		"building":    "rising tension, dynamic perspective",
		"tense":       "high contrast, suspenseful mood",
		"somber":      "muted colors, melancholic atmosphere",
		"reflective":  "soft lighting, contemplative mood",
	}

	moodDesc := moodDescriptors[mood]
	if moodDesc == "" {
		moodDesc = "cinematic documentary style"
	}

	return fmt.Sprintf("historical documentary style, %s, depicting %s, %s, %s aspect ratio, no text, no modern elements",
		moodDesc, keywordStr, styleProfile, aspectRatio)
}

func generateSSML(text string) string {
	// Add breaks after sentences
	ssml := text
	ssml = strings.ReplaceAll(ssml, ". ", ".<break time=\"400ms\"/> ")
	ssml = strings.ReplaceAll(ssml, "? ", "?<break time=\"500ms\"/> ")
	ssml = strings.ReplaceAll(ssml, "! ", "!<break time=\"400ms\"/> ")

	return fmt.Sprintf("<speak>%s</speak>", ssml)
}
