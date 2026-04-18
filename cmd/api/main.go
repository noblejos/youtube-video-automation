package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gama/youtube-video-automation/internal/api"
	"github.com/gama/youtube-video-automation/internal/config"
	"github.com/gama/youtube-video-automation/internal/db"
	"github.com/gama/youtube-video-automation/internal/images"
	"github.com/gama/youtube-video-automation/internal/jobs"
	"github.com/gama/youtube-video-automation/internal/projects"
	localstore "github.com/gama/youtube-video-automation/internal/providers/storage/local"
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
	logger.Info(ctx, "starting API server", "env", cfg.AppEnv)

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
	_ = subtitles.NewRepository(database.Pool) // Used by worker
	renderRepo := render.NewRepository(database.Pool)

	// Initialize script service (needed for user-provided scripts)
	scriptGenerator := scripts.NewMockGenerator() // Actual generation happens in worker
	scriptService := scripts.NewService(scriptRepo, scriptGenerator)

	// Initialize job queue with persistence callback
	var queue jobs.Queue
	switch cfg.QueueBackend {
	case "redis":
		redisQueue, err := jobs.NewRedisQueue(jobs.RedisQueueConfig{
			RedisURL:    cfg.RedisURL,
			WorkerCount: 0, // API doesn't process jobs, only enqueues
		}, func(ctx context.Context, job *models.Job) error {
			return jobRepo.Update(ctx, job)
		})
		if err != nil {
			logger.Fatal(ctx, "failed to initialize Redis queue", "error", err)
		}
		queue = redisQueue
		logger.Info(ctx, "using Redis queue backend")
	default:
		queue = jobs.NewMemoryQueue(func(ctx context.Context, job *models.Job) error {
			return jobRepo.Update(ctx, job)
		})
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

	// API only enqueues jobs - worker processes them
	// No need to register handlers or start queue processing here

	// Create and start HTTP server
	server := api.NewServer(workflowService, storage)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info(ctx, "shutting down...")

		// Shutdown server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error(ctx, "server shutdown error", "error", err)
		}
	}()

	// Start server
	if err := server.Start(":" + cfg.Port); err != nil {
		logger.Info(ctx, "server stopped", "error", err)
	}
}
