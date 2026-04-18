package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	chatAPIURL = "https://api.openai.com/v1/chat/completions"
)

// Config holds OpenAI LLM configuration
type Config struct {
	APIKey string
	Model  string // gpt-4o, gpt-4-turbo, etc.
}

// Client is an OpenAI LLM client
type Client struct {
	config Config
	client *http.Client
}

// New creates a new OpenAI LLM client
func New(cfg Config) *Client {
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}

	return &Client{
		config: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete sends a chat completion request
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.config.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is not configured")
	}

	req := ChatRequest{
		Model: c.config.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   4096,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w\nBody: %s", err, string(body))
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}

// CompleteJSON sends a request and parses JSON response into target
func (c *Client) CompleteJSON(ctx context.Context, systemPrompt, userPrompt string, target interface{}) error {
	response, err := c.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	// Try to extract JSON from response (handle markdown code blocks)
	jsonStr := extractJSON(response)

	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w\nResponse: %s", err, response)
	}

	return nil
}

// extractJSON extracts JSON from a response that might be wrapped in markdown
func extractJSON(s string) string {
	// Try to find JSON in code blocks
	start := -1
	end := -1

	// Look for ```json or ```
	for i := 0; i < len(s)-3; i++ {
		if s[i:i+3] == "```" {
			if start == -1 {
				// Find the end of the opening line
				for j := i + 3; j < len(s); j++ {
					if s[j] == '\n' {
						start = j + 1
						break
					}
				}
			} else {
				end = i
				break
			}
		}
	}

	if start != -1 && end != -1 {
		return s[start:end]
	}

	// Try to find raw JSON object
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			// Find matching closing brace
			depth := 0
			for j := i; j < len(s); j++ {
				if s[j] == '{' {
					depth++
				} else if s[j] == '}' {
					depth--
					if depth == 0 {
						return s[i : j+1]
					}
				}
			}
		}
	}

	return s
}
