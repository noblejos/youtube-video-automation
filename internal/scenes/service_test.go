package scenes

import (
	"context"
	"testing"

	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/models"
)

func TestMockGenerator_Generate(t *testing.T) {
	generator := NewMockGenerator()
	ctx := context.Background()

	script := &models.Script{
		Hook:             "Test hook text for the video.",
		SetupText:        "Test setup text explaining the context.",
		BuildText:        "Test build text developing the story.",
		TurningPointText: "Test turning point text with drama.",
		CollapseText:     "Test collapse text showing the fall.",
		ConclusionText:   "Test conclusion wrapping up the story.",
	}

	result, err := generator.Generate(ctx, script, "9:16", "dramatic_history_shorts", 120)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check scene count
	if len(result.Scenes) < 6 || len(result.Scenes) > 8 {
		t.Errorf("expected 6-8 scenes, got %d", len(result.Scenes))
	}

	// Check each scene has required fields
	for i, scene := range result.Scenes {
		if scene.SceneNumber != i+1 {
			t.Errorf("scene %d has wrong scene_number: %d", i, scene.SceneNumber)
		}
		if scene.NarrationText == "" {
			t.Errorf("scene %d has empty narration_text", i+1)
		}
		if scene.DurationSec <= 0 {
			t.Errorf("scene %d has invalid duration: %f", i+1, scene.DurationSec)
		}
		if scene.Mood == "" {
			t.Errorf("scene %d has empty mood", i+1)
		}
		if len(scene.Keywords) == 0 {
			t.Errorf("scene %d has no keywords", i+1)
		}
		if scene.VisualPrompt == "" {
			t.Errorf("scene %d has empty visual_prompt", i+1)
		}
	}

	// Check total duration is approximately target
	totalDuration := 0.0
	for _, scene := range result.Scenes {
		totalDuration += scene.DurationSec
	}
	if totalDuration < 100 || totalDuration > 140 {
		t.Errorf("total duration %f is too far from target 120", totalDuration)
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int // minimum expected keywords
	}{
		{
			name: "normal text",
			text: "The ancient empire of Mali was one of the richest civilizations in history.",
			want: 3,
		},
		{
			name: "short text",
			text: "Gold and power.",
			want: 1,
		},
		{
			name: "empty text",
			text: "",
			want: 1, // Should return default keywords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractKeywords(tt.text)
			if len(keywords) < tt.want {
				t.Errorf("extractKeywords() got %d keywords, want at least %d", len(keywords), tt.want)
			}
		})
	}
}

func TestGenerateSSML(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "single sentence",
			text: "Hello world.",
			want: "<speak>Hello world.</speak>",
		},
		{
			name: "multiple sentences",
			text: "First sentence. Second sentence.",
			want: "<speak>First sentence.<break time=\"400ms\"/> Second sentence.</speak>",
		},
		{
			name: "question",
			text: "What happened? Nobody knows.",
			want: "<speak>What happened?<break time=\"500ms\"/> Nobody knows.</speak>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSSML(tt.text)
			if got != tt.want {
				t.Errorf("generateSSML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSceneResponse_Validation(t *testing.T) {
	tests := []struct {
		name    string
		scene   contracts.SceneResponse
		wantErr bool
	}{
		{
			name: "valid scene",
			scene: contracts.SceneResponse{
				SceneNumber:   1,
				NarrationText: "Test narration",
				DurationSec:   10.0,
				Mood:          "dramatic",
				Keywords:      []string{"test"},
				VisualPrompt:  "test prompt",
				AssetStrategy: "ai_or_archive",
			},
			wantErr: false,
		},
		{
			name: "invalid - zero duration",
			scene: contracts.SceneResponse{
				SceneNumber:   1,
				NarrationText: "Test narration",
				DurationSec:   0,
				Mood:          "dramatic",
				Keywords:      []string{"test"},
				VisualPrompt:  "test prompt",
				AssetStrategy: "ai_or_archive",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty narration",
			scene: contracts.SceneResponse{
				SceneNumber:   1,
				NarrationText: "",
				DurationSec:   10.0,
				Mood:          "dramatic",
				Keywords:      []string{"test"},
				VisualPrompt:  "test prompt",
				AssetStrategy: "ai_or_archive",
			},
			wantErr: true,
		},
		{
			name: "invalid - duration too long",
			scene: contracts.SceneResponse{
				SceneNumber:   1,
				NarrationText: "Test",
				DurationSec:   30.0, // Max is 25
				Mood:          "dramatic",
				Keywords:      []string{"test"},
				VisualPrompt:  "test prompt",
				AssetStrategy: "ai_or_archive",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScene(tt.scene)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateScene() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validateScene(scene contracts.SceneResponse) error {
	if scene.NarrationText == "" {
		return &ValidationError{Field: "narration_text", Message: "narration text is required"}
	}
	if scene.DurationSec <= 0 {
		return &ValidationError{Field: "duration_sec", Message: "duration must be positive"}
	}
	if scene.DurationSec > 25 {
		return &ValidationError{Field: "duration_sec", Message: "duration cannot exceed 25 seconds"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
