package openai

import (
	"context"
	"fmt"

	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// SceneGenerator generates scenes using OpenAI
type SceneGenerator struct {
	client      *Client
	voiceConfig VoiceConfig
}

// VoiceConfig holds default voice settings
type VoiceConfig struct {
	Voice  string
	Engine string
}

// NewSceneGenerator creates a new OpenAI scene generator
func NewSceneGenerator(client *Client, voiceConfig VoiceConfig) *SceneGenerator {
	if voiceConfig.Voice == "" {
		voiceConfig.Voice = "Ayanda"
	}
	if voiceConfig.Engine == "" {
		voiceConfig.Engine = "standard"
	}
	return &SceneGenerator{
		client:      client,
		voiceConfig: voiceConfig,
	}
}

const sceneSystemPrompt = `You are an expert video production planner specializing in dramatic historical short-form content.

Your job is to break a script into scenes with detailed production specifications including:
- Precise timing for each scene
- SSML markup for natural voiceover with pauses and emphasis
- Detailed image generation prompts optimized for AI image generators (DALL-E, Flux)
- Camera motion and transition effects
- Mood and energy levels

For SSML:
- Use <break time="Xms"/> for pauses (200-500ms typical)
- Use <emphasis> for dramatic words
- Use <prosody rate="slow"> for impactful moments

For image prompts:
- Be specific and descriptive
- Include art style, lighting, composition
- Always include "no text, no watermark" in negative prompts
- Optimize for vertical 9:16 format

Output ONLY valid JSON. No markdown, no explanations.`

const sceneUserPromptTemplate = `Break this script into production-ready scenes:

SCRIPT:
%s

SETTINGS:
- Aspect ratio: %s
- Style profile: %s
- Target total duration: %d seconds
- Voice: %s
- Engine: %s

Output valid JSON with this structure (no markdown, just raw JSON):
- title: string
- aspect_ratio: string (use the aspect ratio from settings)
- style_profile: string (use the style profile from settings)
- total_duration_sec: number
- scenes: array of scene objects

Each scene object must have:
- scene_number: number (1, 2, 3, etc.)
- start_time_sec: number
- duration_sec: number
- story_role: string (hook, setup, build, turning_point, collapse, or conclusion)
- mood: string (dramatic, tense, somber, hopeful, etc.)
- energy_level: string (low, medium, high)
- narration_text: string (the spoken text)
- ssml_text: string (SSML markup with <speak> tags and <break time="Xms"/> for pauses)
- keywords: array of strings (5-7 relevant keywords)
- visual: object with type, prompt, negative_prompt, camera_motion, transition_in, transition_out
- audio: object with voice, engine, background_music_mood
- subtitles: object with enabled (boolean) and style

Create 5-7 scenes. Ensure durations add up to approximately %d seconds total.`

// GeneratedScenes represents the full scene generation response
type GeneratedScenes struct {
	Title            string                    `json:"title"`
	AspectRatio      string                    `json:"aspect_ratio"`
	StyleProfile     string                    `json:"style_profile"`
	TotalDurationSec float64                   `json:"total_duration_sec"`
	Scenes           []contracts.SceneResponse `json:"scenes"`
}

// Generate generates scenes from a script
func (g *SceneGenerator) Generate(ctx context.Context, script *models.Script, aspectRatio, styleProfile string, targetDuration int) (*contracts.ScenesResponse, error) {
	userPrompt := fmt.Sprintf(sceneUserPromptTemplate,
		script.FullScript,
		aspectRatio, styleProfile, targetDuration,
		g.voiceConfig.Voice, g.voiceConfig.Engine,
		targetDuration,
	)

	var result GeneratedScenes
	if err := g.client.CompleteJSON(ctx, sceneSystemPrompt, userPrompt, &result); err != nil {
		return nil, fmt.Errorf("failed to generate scenes: %w", err)
	}

	// Map legacy fields for backwards compatibility
	for i := range result.Scenes {
		scene := &result.Scenes[i]
		if scene.Visual != nil && scene.VisualPrompt == "" {
			scene.VisualPrompt = scene.Visual.Prompt
		}
		if scene.AssetStrategy == "" {
			scene.AssetStrategy = "ai_generated"
		}
	}

	return &contracts.ScenesResponse{
		AspectRatio:  result.AspectRatio,
		StyleProfile: result.StyleProfile,
		Scenes:       result.Scenes,
	}, nil
}
