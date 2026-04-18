package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string

	// Queue settings
	QueueBackend string
	RedisURL     string
	WorkerCount  int

	// Storage settings
	StorageBackend string
	StorageBucket  string
	StorageBaseDir string // For local storage

	// AWS settings
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	// Voice settings
	VoiceProvider    string
	PollyDefaultVoice string
	PollyEngine      string

	// Image settings
	ImageProvider string
	OpenAIAPIKey  string
	OpenAIModel   string

	// LLM settings (for script/scene generation)
	LLMProvider string
	LLMModel    string

	// Render settings
	RenderFPS    int
	RenderWidth  int
	RenderHeight int

	// Review settings
	ReviewRequiredDefault bool

	// FFmpeg settings
	FFmpegPath    string
	UseMockRender bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/youtube_automation?sslmode=disable"),

		QueueBackend: getEnv("QUEUE_BACKEND", "redis"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		WorkerCount:  getEnvInt("WORKER_COUNT", 3),

		StorageBackend: getEnv("STORAGE_BACKEND", "local"),
		StorageBucket:  getEnv("STORAGE_BUCKET", "youtube-automation"),
		StorageBaseDir: getEnv("STORAGE_BASE_DIR", "./storage"),

		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),

		VoiceProvider:     getEnv("VOICE_PROVIDER", "mock"),
		PollyDefaultVoice: getEnv("POLLY_DEFAULT_VOICE", "Ayanda"),
		PollyEngine:       getEnv("POLLY_ENGINE", "standard"),

		ImageProvider: getEnv("IMAGE_PROVIDER", "mock"),
		OpenAIAPIKey:  getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:   getEnv("OPENAI_MODEL", "dall-e-3"),

		LLMProvider: getEnv("LLM_PROVIDER", "mock"),
		LLMModel:    getEnv("LLM_MODEL", "gpt-4o"),

		RenderFPS:    getEnvInt("RENDER_FPS", 30),
		RenderWidth:  getEnvInt("RENDER_WIDTH", 1080),
		RenderHeight: getEnvInt("RENDER_HEIGHT", 1920),

		ReviewRequiredDefault: getEnvBool("REVIEW_REQUIRED_DEFAULT", true),

		FFmpegPath:    getEnv("FFMPEG_PATH", "ffmpeg"),
		UseMockRender: getEnvBool("USE_MOCK_RENDER", true),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
