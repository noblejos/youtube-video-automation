package openai

import (
	"context"
	"fmt"

	"github.com/gama/youtube-video-automation/pkg/contracts"
)

// ScriptGenerator generates scripts using OpenAI
type ScriptGenerator struct {
	client *Client
}

// NewScriptGenerator creates a new OpenAI script generator
func NewScriptGenerator(client *Client) *ScriptGenerator {
	return &ScriptGenerator{client: client}
}

const scriptSystemPrompt = `You are an expert short-form video scriptwriter specializing in dramatic historical content for platforms like YouTube Shorts and TikTok.

Your scripts should:
- Hook viewers in the first 2 seconds with a provocative statement or question
- Follow a dramatic narrative arc: hook → setup → build → turning point → collapse → conclusion
- Use vivid, cinematic language that paints mental pictures
- Keep sentences short and punchy for voiceover
- Target the specified duration (typically 60-120 seconds)
- Be historically accurate but dramatically engaging

Output ONLY valid JSON matching the requested format. No markdown, no explanations.`

const scriptUserPrompt = `Create a dramatic short-form video script about: "%s"

Target duration: %d seconds
Style: %s

Output this exact JSON structure:
{
  "title": "Catchy, dramatic title with emotional hook",
  "hook": "Opening line that immediately grabs attention (2-3 seconds when spoken)",
  "setup_text": "Establish the context and stakes (15-20 seconds)",
  "build_text": "Build tension and develop the story (25-30 seconds)",
  "turning_point_text": "The dramatic moment everything changes (15-20 seconds)",
  "collapse_text": "The aftermath and consequences (15-20 seconds)",
  "conclusion_text": "Powerful closing that resonates (10-15 seconds)",
  "full_script": "All sections combined as one flowing narrative"
}`

// Generate generates a script for the given topic
func (g *ScriptGenerator) Generate(ctx context.Context, topic, channelStyle string, targetDuration int) (*contracts.ScriptResponse, error) {
	userPrompt := fmt.Sprintf(scriptUserPrompt, topic, targetDuration, channelStyle)

	var result contracts.ScriptResponse
	if err := g.client.CompleteJSON(ctx, scriptSystemPrompt, userPrompt, &result); err != nil {
		return nil, fmt.Errorf("failed to generate script: %w", err)
	}

	return &result, nil
}
