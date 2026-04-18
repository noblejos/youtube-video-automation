package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gama/youtube-video-automation/internal/images"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/pkg/contracts"
)

const (
	apiURL = "https://api.openai.com/v1/images/generations"
)

// Config holds OpenAI configuration
type Config struct {
	APIKey string
	Model  string // dall-e-2, dall-e-3, gpt-image-1, or gpt-image-1-mini
}

// Provider implements images.Provider using OpenAI DALL-E
type Provider struct {
	config  Config
	storage storage.Provider
	client  *http.Client
}

// New creates a new OpenAI DALL-E provider
func New(cfg Config, store storage.Provider) images.Provider {
	if cfg.Model == "" {
		cfg.Model = "dall-e-3"
	}

	return &Provider{
		config:  cfg,
		storage: store,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "openai"
}

// Generate generates an image using OpenAI DALL-E or GPT Image
func (p *Provider) Generate(ctx context.Context, req contracts.ImageRequest) (*contracts.ImageResult, error) {
	// Build request based on model type
	requestBody := map[string]interface{}{
		"model":  p.config.Model,
		"prompt": req.Prompt,
		"n":      1,
	}

	// Track size for metadata
	var size string

	// GPT Image models use different parameters
	isGPTImage := p.config.Model == "gpt-image-1" || p.config.Model == "gpt-image-1-mini"

	if isGPTImage {
		// GPT Image models use specific sizes
		switch req.AspectRatio {
		case "9:16":
			size = "1024x1536" // Portrait (closest to 9:16)
		case "16:9":
			size = "1536x1024" // Landscape
		default:
			size = "1024x1024" // Square
		}
		requestBody["size"] = size
		// GPT Image models support quality: low, medium, high
		if p.config.Model == "gpt-image-1-mini" {
			requestBody["quality"] = "medium"
		} else {
			requestBody["quality"] = "high"
		}
		// GPT Image models return base64 by default, no response_format needed
	} else {
		// DALL-E models use size parameter
		size = "1024x1792" // Default for 9:16 (portrait/vertical video)
		if req.AspectRatio == "16:9" {
			size = "1792x1024"
		} else if req.AspectRatio == "1:1" {
			size = "1024x1024"
		}

		// For DALL-E 2, sizes are limited
		if p.config.Model == "dall-e-2" {
			size = "1024x1024"
		}

		requestBody["size"] = size
		requestBody["response_format"] = "b64_json"

		// DALL-E 3 supports quality parameter
		if p.config.Model == "dall-e-3" {
			requestBody["quality"] = "standard" // or "hd"
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response - handle both b64_json and URL responses
	var result struct {
		Data []struct {
			B64JSON       string `json:"b64_json"`
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image data in response")
	}

	var imageData []byte

	if result.Data[0].B64JSON != "" {
		// Decode base64 image
		imageData, err = base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}
	} else if result.Data[0].URL != "" {
		// Download image from URL
		imageResp, err := p.client.Get(result.Data[0].URL)
		if err != nil {
			return nil, fmt.Errorf("failed to download image: %w", err)
		}
		defer imageResp.Body.Close()

		imageData, err = io.ReadAll(imageResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read image data: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no image data (b64_json or url) in response")
	}

	// Generate storage key
	storageKey := fmt.Sprintf("projects/%s/assets/scene_%s.png", req.ProjectID, req.SceneID)

	// Store the image
	if err := p.storage.Put(ctx, storageKey, "image/png", imageData); err != nil {
		return nil, fmt.Errorf("failed to store image: %w", err)
	}

	return &contracts.ImageResult{
		StorageKey: storageKey,
		Provider:   p.Name(),
		Metadata: map[string]interface{}{
			"model":          p.config.Model,
			"size":           size,
			"revised_prompt": result.Data[0].RevisedPrompt,
		},
	}, nil
}
