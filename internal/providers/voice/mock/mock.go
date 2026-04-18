package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/internal/voice"
	"github.com/gama/youtube-video-automation/pkg/contracts"
)

// Provider is a mock voice provider for testing
type Provider struct {
	storage storage.Provider
}

// New creates a new mock voice provider
func New(storage storage.Provider) voice.Provider {
	return &Provider{storage: storage}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "mock"
}

// Generate generates mock audio data
func (p *Provider) Generate(ctx context.Context, req contracts.VoiceRequest) (*contracts.VoiceResult, error) {
	// Calculate mock duration based on text length
	// Approximate: 150 words per minute, average 5 chars per word
	charCount := len(req.Text)
	wordsEstimate := float64(charCount) / 5.0
	durationMs := int((wordsEstimate / 150.0) * 60.0 * 1000.0)
	if durationMs < 1000 {
		durationMs = 1000 // Minimum 1 second
	}

	// Generate storage key
	storageKey := fmt.Sprintf("projects/%s/audio/scene_%s.mp3", req.ProjectID, req.SceneID)

	// Create mock MP3 data (just a placeholder)
	mockData := generateMockMP3Header(durationMs)

	// Store the mock audio
	if err := p.storage.Put(ctx, storageKey, "audio/mpeg", mockData); err != nil {
		return nil, fmt.Errorf("failed to store mock audio: %w", err)
	}

	return &contracts.VoiceResult{
		StorageKey: storageKey,
		DurationMs: durationMs,
	}, nil
}

// generateMockMP3Header generates a minimal valid-ish MP3 header
// This is just for testing - not a real MP3 file
func generateMockMP3Header(durationMs int) []byte {
	// Create a simple placeholder with metadata comment
	header := []byte{
		0xFF, 0xFB, 0x90, 0x00, // MP3 frame header
	}
	// Add some padding to make it look like audio data
	padding := make([]byte, 1024)
	for i := range padding {
		padding[i] = 0x00
	}
	return append(header, padding...)
}

// MockProviderWithDelay adds artificial delay for testing
type MockProviderWithDelay struct {
	*Provider
	delay time.Duration
}

// NewWithDelay creates a mock provider with artificial delay
func NewWithDelay(storage storage.Provider, delay time.Duration) voice.Provider {
	return &MockProviderWithDelay{
		Provider: &Provider{storage: storage},
		delay:    delay,
	}
}

// Generate generates mock audio with delay
func (p *MockProviderWithDelay) Generate(ctx context.Context, req contracts.VoiceRequest) (*contracts.VoiceResult, error) {
	select {
	case <-time.After(p.delay):
		return p.Provider.Generate(ctx, req)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
