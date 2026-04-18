package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gama/youtube-video-automation/internal/config"
	"github.com/gama/youtube-video-automation/internal/db"
	"github.com/gama/youtube-video-automation/internal/handlers"
	"github.com/gama/youtube-video-automation/internal/images"
	"github.com/gama/youtube-video-automation/internal/jobs"
	"github.com/gama/youtube-video-automation/internal/projects"
	imagemock "github.com/gama/youtube-video-automation/internal/providers/image/mock"
	imageopenai "github.com/gama/youtube-video-automation/internal/providers/image/openai"
	llmopenai "github.com/gama/youtube-video-automation/internal/providers/llm/openai"
	localstore "github.com/gama/youtube-video-automation/internal/providers/storage/local"
	voicemock "github.com/gama/youtube-video-automation/internal/providers/voice/mock"
	voicepolly "github.com/gama/youtube-video-automation/internal/providers/voice/polly"
	"github.com/gama/youtube-video-automation/internal/render"
	"github.com/gama/youtube-video-automation/internal/scenes"
	"github.com/gama/youtube-video-automation/internal/scripts"
	"github.com/gama/youtube-video-automation/internal/subtitles"
	"github.com/gama/youtube-video-automation/internal/voice"
	"github.com/gama/youtube-video-automation/internal/workflow"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(ctx, "failed to load config", "error", err)
	}

	// Initialize logger
	logger.Init(cfg.IsDevelopment())
	logger.Info(ctx, "starting worker", "env", cfg.AppEnv)

	// Connect to database
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal(ctx, "failed to connect to database", "error", err)
	}
	defer database.Close()

	// Initialize storage
	storage, err := localstore.New(cfg.StorageBaseDir)
	if err != nil {
		logger.Fatal(ctx, "failed to initialize storage", "error", err)
	}

	// Initialize repositories
	projectRepo := projects.NewRepository(database.Pool)
	scriptRepo := scripts.NewRepository(database.Pool)
	sceneRepo := scenes.NewRepository(database.Pool)
	jobRepo := jobs.NewRepository(database.Pool)
	audioRepo := voice.NewRepository(database.Pool)
	assetRepo := images.NewRepository(database.Pool)
	subtitleRepo := subtitles.NewRepository(database.Pool)
	renderRepo := render.NewRepository(database.Pool)

	// Initialize voice provider
	var voiceProvider voice.Provider
	switch cfg.VoiceProvider {
	case "polly":
		var err error
		voiceProvider, err = voicepolly.New(ctx, voicepolly.Config{
			Region:          cfg.AWSRegion,
			AccessKeyID:     cfg.AWSAccessKeyID,
			SecretAccessKey: cfg.AWSSecretAccessKey,
			DefaultVoice:    cfg.PollyDefaultVoice,
			Engine:          cfg.PollyEngine,
		}, storage)
		if err != nil {
			logger.Fatal(ctx, "failed to initialize Polly provider", "error", err)
		}
		logger.Info(ctx, "using AWS Polly voice provider")
	default:
		voiceProvider = voicemock.New(storage)
		logger.Info(ctx, "using mock voice provider")
	}

	// Initialize image provider
	var imageProvider images.Provider
	switch cfg.ImageProvider {
	case "openai":
		imageProvider = imageopenai.New(imageopenai.Config{
			APIKey: cfg.OpenAIAPIKey,
			Model:  cfg.OpenAIModel,
		}, storage)
		logger.Info(ctx, "using OpenAI DALL-E image provider", "model", cfg.OpenAIModel)
	default:
		imageProvider = imagemock.New(storage)
		logger.Info(ctx, "using mock image provider")
	}

	// Initialize LLM-based generators
	var scriptGenerator scripts.Generator
	var sceneGenerator scenes.Generator

	switch cfg.LLMProvider {
	case "openai":
		llmClient := llmopenai.New(llmopenai.Config{
			APIKey: cfg.OpenAIAPIKey,
			Model:  cfg.LLMModel,
		})
		scriptGenerator = llmopenai.NewScriptGenerator(llmClient)
		sceneGenerator = llmopenai.NewSceneGenerator(llmClient, llmopenai.VoiceConfig{
			Voice:  cfg.PollyDefaultVoice,
			Engine: cfg.PollyEngine,
		})
		logger.Info(ctx, "using OpenAI LLM for script/scene generation", "model", cfg.LLMModel)
	default:
		scriptGenerator = scripts.NewMockGenerator()
		sceneGenerator = scenes.NewMockGenerator()
		logger.Info(ctx, "using mock generators for script/scene generation")
	}

	// Initialize services
	scriptService := scripts.NewService(scriptRepo, scriptGenerator)
	sceneService := scenes.NewService(sceneRepo, sceneGenerator)
	voiceService := voice.NewService(audioRepo, voiceProvider)
	imageService := images.NewService(assetRepo, imageProvider)
	subtitleService := subtitles.NewService(subtitleRepo, storage)
	renderService := render.NewService(renderRepo, storage, render.Config{
		FFmpegPath: cfg.FFmpegPath,
		Width:      cfg.RenderWidth,
		Height:     cfg.RenderHeight,
		FPS:        cfg.RenderFPS,
	})

	// Initialize job queue with persistence callback
	var queue jobs.Queue
	var redisQueue *jobs.RedisQueue

	switch cfg.QueueBackend {
	case "redis":
		var err error
		redisQueue, err = jobs.NewRedisQueue(jobs.RedisQueueConfig{
			RedisURL:    cfg.RedisURL,
			WorkerCount: cfg.WorkerCount,
		}, func(ctx context.Context, job *models.Job) error {
			return jobRepo.Update(ctx, job)
		})
		if err != nil {
			logger.Fatal(ctx, "failed to initialize Redis queue", "error", err)
		}
		queue = redisQueue
		logger.Info(ctx, "using Redis queue backend", "workers", cfg.WorkerCount)
	default:
		memQueue := jobs.NewMemoryQueue(func(ctx context.Context, job *models.Job) error {
			return jobRepo.Update(ctx, job)
		})
		queue = memQueue
		logger.Info(ctx, "using in-memory queue backend")
	}

	// Initialize workflow service
	workflowService := workflow.NewService(
		projectRepo,
		scriptRepo,
		scriptService,
		sceneRepo,
		jobRepo,
		audioRepo,
		assetRepo,
		renderRepo,
		queue,
		workflow.VoiceConfig{
			DefaultVoice: cfg.PollyDefaultVoice,
			Engine:       cfg.PollyEngine,
		},
	)

	// Set up permanent job failure callback to mark projects as failed
	queue.SetOnJobFailedPermanently(func(ctx context.Context, job *models.Job, err error) {
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		}
		failErr := workflowService.MarkProjectFailed(ctx, job.ProjectID,
			"job "+job.JobType+" failed permanently: "+errMsg)
		if failErr != nil {
			logger.Error(ctx, "failed to mark project as failed", "error", failErr)
		}
	})

	// Initialize job handlers
	jobHandlers := handlers.NewJobHandlers(
		scriptService,
		sceneService,
		voiceService,
		imageService,
		subtitleService,
		renderService,
		workflowService,
		sceneRepo,
		cfg.UseMockRender,
	)
	jobHandlers.RegisterHandlers(queue)

	// Start job queue
	if err := queue.Start(ctx); err != nil {
		logger.Fatal(ctx, "failed to start job queue", "error", err)
	}

	logger.Info(ctx, "worker started, waiting for jobs...")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info(ctx, "shutting down worker...")
	queue.Stop()
}
