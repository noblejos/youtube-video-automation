package mock

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"time"

	"github.com/gama/youtube-video-automation/internal/images"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/pkg/contracts"
)

// Provider is a mock image provider for testing
type Provider struct {
	storage storage.Provider
}

// New creates a new mock image provider
func New(storage storage.Provider) images.Provider {
	return &Provider{storage: storage}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "mock"
}

// Generate generates a mock image
func (p *Provider) Generate(ctx context.Context, req contracts.ImageRequest) (*contracts.ImageResult, error) {
	// Determine dimensions based on aspect ratio
	width, height := getDimensions(req.AspectRatio)

	// Generate a simple colored image
	img := generateMockImage(width, height, req.Prompt)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode mock image: %w", err)
	}

	// Generate storage key
	storageKey := fmt.Sprintf("projects/%s/assets/scene_%s.png", req.ProjectID, req.SceneID)

	// Store the mock image
	if err := p.storage.Put(ctx, storageKey, "image/png", buf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to store mock image: %w", err)
	}

	return &contracts.ImageResult{
		StorageKey: storageKey,
		Provider:   "mock",
		Metadata: map[string]interface{}{
			"width":  width,
			"height": height,
		},
	}, nil
}

func getDimensions(aspectRatio string) (int, int) {
	switch aspectRatio {
	case "9:16":
		return 1080, 1920
	case "16:9":
		return 1920, 1080
	case "1:1":
		return 1080, 1080
	default:
		return 1080, 1920 // Default to vertical
	}
}

func generateMockImage(width, height int, prompt string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Generate a color based on prompt hash
	hash := 0
	for _, c := range prompt {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}

	// Create a gradient background
	baseColor := color.RGBA{
		R: uint8(50 + (hash%100)),
		G: uint8(50 + ((hash/100)%100)),
		B: uint8(80 + ((hash/10000)%100)),
		A: 255,
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Add some gradient effect
			factor := float64(y) / float64(height)
			r := uint8(float64(baseColor.R) * (1 - factor*0.3))
			g := uint8(float64(baseColor.G) * (1 - factor*0.3))
			b := uint8(float64(baseColor.B) * (1 - factor*0.2))
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// MockProviderWithDelay adds artificial delay for testing
type MockProviderWithDelay struct {
	*Provider
	delay time.Duration
}

// NewWithDelay creates a mock provider with artificial delay
func NewWithDelay(storage storage.Provider, delay time.Duration) images.Provider {
	return &MockProviderWithDelay{
		Provider: &Provider{storage: storage},
		delay:    delay,
	}
}

// Generate generates mock image with delay
func (p *MockProviderWithDelay) Generate(ctx context.Context, req contracts.ImageRequest) (*contracts.ImageResult, error) {
	select {
	case <-time.After(p.delay):
		return p.Provider.Generate(ctx, req)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
