package images

import (
	"context"

	"github.com/gama/youtube-video-automation/pkg/contracts"
)

// Provider defines the interface for image generation
type Provider interface {
	// Generate generates an image from a prompt
	Generate(ctx context.Context, req contracts.ImageRequest) (*contracts.ImageResult, error)

	// Name returns the provider name
	Name() string
}
