package voice

import (
	"context"

	"github.com/gama/youtube-video-automation/pkg/contracts"
)

// Provider defines the interface for voice/TTS generation
type Provider interface {
	// Generate generates audio from text
	Generate(ctx context.Context, req contracts.VoiceRequest) (*contracts.VoiceResult, error)

	// Name returns the provider name
	Name() string
}
