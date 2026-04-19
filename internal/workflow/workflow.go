package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/internal/images"
	"github.com/gama/youtube-video-automation/internal/jobs"
	"github.com/gama/youtube-video-automation/internal/projects"
	"github.com/gama/youtube-video-automation/internal/render"
	"github.com/gama/youtube-video-automation/internal/scenes"
	"github.com/gama/youtube-video-automation/internal/scripts"
	"github.com/gama/youtube-video-automation/internal/voice"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// VoiceConfig holds voice generation settings
type VoiceConfig struct {
	DefaultVoice string
	Engine       string
}

// Service handles workflow orchestration
type Service struct {
	projectRepo   *projects.Repository
	scriptRepo    *scripts.Repository
	scriptService *scripts.Service
	sceneRepo     *scenes.Repository
	jobRepo       *jobs.Repository
	audioRepo     *voice.Repository
	assetRepo     *images.Repository
	renderRepo    *render.Repository
	queue         jobs.Queue
	voiceConfig   VoiceConfig
}

// NewService creates a new workflow service
func NewService(
	projectRepo *projects.Repository,
	scriptRepo *scripts.Repository,
	scriptService *scripts.Service,
	sceneRepo *scenes.Repository,
	jobRepo *jobs.Repository,
	audioRepo *voice.Repository,
	assetRepo *images.Repository,
	renderRepo *render.Repository,
	queue jobs.Queue,
	voiceConfig VoiceConfig,
) *Service {
	if voiceConfig.DefaultVoice == "" {
		voiceConfig.DefaultVoice = "Ayanda"
	}
	if voiceConfig.Engine == "" {
		voiceConfig.Engine = "standard"
	}
	return &Service{
		projectRepo:   projectRepo,
		scriptRepo:    scriptRepo,
		scriptService: scriptService,
		sceneRepo:     sceneRepo,
		jobRepo:       jobRepo,
		audioRepo:     audioRepo,
		assetRepo:     assetRepo,
		renderRepo:    renderRepo,
		queue:         queue,
		voiceConfig:   voiceConfig,
	}
}

// CreateProject creates a new project and starts the workflow
func (s *Service) CreateProject(ctx context.Context, req *contracts.CreateProjectRequest) (*models.Project, error) {
	// Set defaults
	channelStyle := req.ChannelStyle
	if channelStyle == "" {
		channelStyle = "dramatic_history_shorts"
	}

	targetDuration := req.TargetDurationSec
	if targetDuration == 0 {
		targetDuration = 120
	}

	aspectRatio := req.AspectRatio
	if aspectRatio == "" {
		aspectRatio = "9:16"
	}

	reviewRequired := true
	if req.ReviewRequired != nil {
		reviewRequired = *req.ReviewRequired
	}

	// Voice settings - use from request or fall back to config defaults
	voiceID := req.VoiceID
	if voiceID == "" {
		voiceID = s.voiceConfig.DefaultVoice
	}

	voiceEngine := req.VoiceEngine
	if voiceEngine == "" {
		voiceEngine = s.voiceConfig.Engine
	}

	// Create project
	project := &models.Project{
		ID:                uuid.New(),
		ExternalID:        uuid.New().String()[:8], // Short ID for external reference
		Topic:             req.Topic,
		ChannelStyle:      channelStyle,
		TargetDurationSec: targetDuration,
		AspectRatio:       aspectRatio,
		VoiceID:           voiceID,
		VoiceEngine:       voiceEngine,
		Status:            models.ProjectStatusCreated,
		ReviewRequired:    reviewRequired,
	}

	// Set title - use provided title, or fall back to topic
	if req.Title != "" {
		project.Title = &req.Title
	} else if req.Topic != "" {
		project.Title = &req.Topic
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	ctx = logger.WithProjectID(ctx, project.ID.String())
	logger.Info(ctx, "project created", "topic", project.Topic)

	// If user provided a script, save it and skip to scene generation
	if req.Script != "" {
		logger.Info(ctx, "user-provided script, skipping AI generation")

		// Save the user's script
		script, err := s.scriptService.SaveUserScript(ctx, project.ID, req.Script, req.Title)
		if err != nil {
			return nil, fmt.Errorf("failed to save user script: %w", err)
		}

		// Skip script generation, go directly to scene generation
		if err := s.OnScriptGenerated(ctx, project.ID); err != nil {
			return nil, fmt.Errorf("failed to start scene generation: %w", err)
		}

		_ = script // script saved successfully
		return project, nil
	}

	// Start the workflow by enqueuing script generation
	if err := s.EnqueueScriptGeneration(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to enqueue script generation: %w", err)
	}

	return project, nil
}

// EnqueueScriptGeneration enqueues script generation job
func (s *Service) EnqueueScriptGeneration(ctx context.Context, project *models.Project) error {
	payload := jobs.ScriptJobPayload{
		JobPayload: contracts.JobPayload{
			ProjectID: project.ID,
		},
		Topic:          project.Topic,
		ChannelStyle:   project.ChannelStyle,
		TargetDuration: project.TargetDurationSec,
	}

	job, err := jobs.CreateJob(project.ID, models.JobTypeScriptGeneration, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return err
	}

	step := "SCRIPT_GENERATION"
	if err := s.projectRepo.UpdateStatus(ctx, project.ID, models.ProjectStatusScriptGenerating, &step); err != nil {
		return err
	}

	return nil
}

// OnScriptGenerated handles script generation completion
func (s *Service) OnScriptGenerated(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	// Update project status
	step := "SCENES_GENERATION"
	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusScriptReady, &step); err != nil {
		return err
	}

	// Get project for scene generation
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Get script for scene generation
	script, err := s.scriptRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	if script == nil {
		return fmt.Errorf("script not found for project %s", projectID)
	}

	// Update project title
	if project.Title == nil {
		// Extract title from script (in real implementation, this would be parsed from LLM response)
		title := fmt.Sprintf("Video: %s", project.Topic)
		if err := s.projectRepo.UpdateTitle(ctx, projectID, title); err != nil {
			logger.Warn(ctx, "failed to update project title", "error", err)
		}
	}

	// Enqueue scene generation
	return s.EnqueueSceneGeneration(ctx, project, script.ID)
}

// EnqueueSceneGeneration enqueues scene generation job
func (s *Service) EnqueueSceneGeneration(ctx context.Context, project *models.Project, scriptID uuid.UUID) error {
	payload := jobs.SceneJobPayload{
		JobPayload: contracts.JobPayload{
			ProjectID: project.ID,
		},
		ScriptID: scriptID,
	}

	job, err := jobs.CreateJob(project.ID, models.JobTypeScenesGeneration, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return err
	}

	step := "SCENES_GENERATION"
	if err := s.projectRepo.UpdateStatus(ctx, project.ID, models.ProjectStatusScenesGenerating, &step); err != nil {
		return err
	}

	return nil
}

// OnScenesGenerated handles scenes generation completion
func (s *Service) OnScenesGenerated(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	// Update project status
	step := "VOICE_IMAGE_GENERATION"
	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusScenesReady, &step); err != nil {
		return err
	}

	// Get project
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Get scenes
	sceneList, err := s.sceneRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// Enqueue voice and image generation for each scene
	for _, scene := range sceneList {
		if err := s.EnqueueVoiceGeneration(ctx, project, scene); err != nil {
			return fmt.Errorf("failed to enqueue voice generation for scene %d: %w", scene.SceneNumber, err)
		}
		if err := s.EnqueueImageGeneration(ctx, project, scene); err != nil {
			return fmt.Errorf("failed to enqueue image generation for scene %d: %w", scene.SceneNumber, err)
		}
	}

	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusVoiceGenerating, &step); err != nil {
		return err
	}

	return nil
}

// EnqueueVoiceGeneration enqueues voice generation job for a scene
func (s *Service) EnqueueVoiceGeneration(ctx context.Context, project *models.Project, scene *models.Scene) error {
	payload := jobs.VoiceJobPayload{
		JobPayload: contracts.JobPayload{
			ProjectID: project.ID,
			SceneID:   &scene.ID,
		},
		VoiceID: project.VoiceID,
		Engine:  project.VoiceEngine,
	}

	job, err := jobs.CreateJob(project.ID, models.JobTypeVoiceGeneration, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	return s.queue.Enqueue(ctx, job)
}

// EnqueueImageGeneration enqueues image generation job for a scene
func (s *Service) EnqueueImageGeneration(ctx context.Context, project *models.Project, scene *models.Scene) error {
	payload := jobs.ImageJobPayload{
		JobPayload: contracts.JobPayload{
			ProjectID: project.ID,
			SceneID:   &scene.ID,
		},
		AspectRatio:  project.AspectRatio,
		StyleProfile: project.ChannelStyle,
	}

	job, err := jobs.CreateJob(project.ID, models.JobTypeImageGeneration, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	return s.queue.Enqueue(ctx, job)
}

// CheckAssetCompletion checks if all voice and image assets are ready
func (s *Service) CheckAssetCompletion(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	// Get scene count
	sceneList, err := s.sceneRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}
	sceneCount := len(sceneList)

	// Check audio count
	audioCount, err := s.audioRepo.CountByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// Check image count
	imageCount, err := s.assetRepo.CountImagesByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	logger.Info(ctx, "checking asset completion",
		"scenes", sceneCount,
		"audio_files", audioCount,
		"images", imageCount,
	)

	// If all assets are ready, proceed to subtitle generation
	if audioCount >= sceneCount && imageCount >= sceneCount {
		step := "SUBTITLE_GENERATION"
		if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusAssetsReady, &step); err != nil {
			return err
		}

		return s.EnqueueSubtitleGeneration(ctx, projectID)
	}

	return nil
}

// EnqueueSubtitleGeneration enqueues subtitle generation job
func (s *Service) EnqueueSubtitleGeneration(ctx context.Context, projectID uuid.UUID) error {
	payload := contracts.JobPayload{
		ProjectID: projectID,
	}

	job, err := jobs.CreateJob(projectID, models.JobTypeSubtitleGeneration, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return err
	}

	step := "SUBTITLE_GENERATION"
	return s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusSubtitlesGenerating, &step)
}

// OnSubtitlesGenerated handles subtitle generation completion
func (s *Service) OnSubtitlesGenerated(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	step := "RENDER"
	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusSubtitlesReady, &step); err != nil {
		return err
	}

	return s.EnqueueRender(ctx, projectID)
}

// EnqueueRender enqueues render job
func (s *Service) EnqueueRender(ctx context.Context, projectID uuid.UUID) error {
	payload := contracts.JobPayload{
		ProjectID: projectID,
	}

	job, err := jobs.CreateJob(projectID, models.JobTypeRender, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return err
	}

	step := "RENDER"
	return s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusRendering, &step)
}

// OnRenderComplete handles render completion
func (s *Service) OnRenderComplete(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	step := "REVIEW_PACKAGE"
	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusRenderReady, &step); err != nil {
		return err
	}

	return s.EnqueueReviewPackage(ctx, projectID)
}

// EnqueueReviewPackage enqueues review package generation
func (s *Service) EnqueueReviewPackage(ctx context.Context, projectID uuid.UUID) error {
	payload := contracts.JobPayload{
		ProjectID: projectID,
	}

	job, err := jobs.CreateJob(projectID, models.JobTypeReviewPackage, payload)
	if err != nil {
		return err
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return err
	}

	return s.queue.Enqueue(ctx, job)
}

// OnReviewPackageReady handles review package completion
func (s *Service) OnReviewPackageReady(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	step := "IN_REVIEW"
	return s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusInReview, &step)
}

// ApproveProject approves a project
func (s *Service) ApproveProject(ctx context.Context, projectID uuid.UUID, notes, actedBy string) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.Status != models.ProjectStatusInReview && project.Status != models.ProjectStatusReviewPackaged {
		return fmt.Errorf("project is not in reviewable state: %s", project.Status)
	}

	step := "APPROVED"
	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusApproved, &step); err != nil {
		return err
	}

	logger.Info(ctx, "project approved", "acted_by", actedBy)
	return nil
}

// RejectProject rejects a project
func (s *Service) RejectProject(ctx context.Context, projectID uuid.UUID, notes, actedBy string) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.Status != models.ProjectStatusInReview && project.Status != models.ProjectStatusReviewPackaged {
		return fmt.Errorf("project is not in reviewable state: %s", project.Status)
	}

	step := "REJECTED"
	if err := s.projectRepo.UpdateStatus(ctx, projectID, models.ProjectStatusRejected, &step); err != nil {
		return err
	}

	logger.Info(ctx, "project rejected", "acted_by", actedBy, "notes", notes)
	return nil
}

// RetryProject retries failed jobs for a project
func (s *Service) RetryProject(ctx context.Context, projectID uuid.UUID) error {
	ctx = logger.WithProjectID(ctx, projectID.String())

	failedJobs, err := s.jobRepo.GetFailedJobs(ctx, projectID)
	if err != nil {
		return err
	}

	if len(failedJobs) == 0 {
		return fmt.Errorf("no failed jobs found for project")
	}

	for _, job := range failedJobs {
		// Reset job for retry
		job.Status = models.JobStatusQueued
		job.AttemptCount = 0
		job.ErrorMessage = nil
		job.Result = nil

		if err := s.jobRepo.Update(ctx, job); err != nil {
			return fmt.Errorf("failed to reset job %s: %w", job.ID, err)
		}

		if err := s.queue.Enqueue(ctx, job); err != nil {
			return fmt.Errorf("failed to re-enqueue job %s: %w", job.ID, err)
		}

		logger.Info(ctx, "job re-queued for retry", "job_id", job.ID, "job_type", job.JobType)
	}

	// Reset project status if it was failed
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.Status == models.ProjectStatusFailed {
		// Determine appropriate status based on failed job types
		// For simplicity, just set to the first failed job's stage
		if len(failedJobs) > 0 {
			var newStatus string
			switch failedJobs[0].JobType {
			case models.JobTypeScriptGeneration:
				newStatus = models.ProjectStatusScriptGenerating
			case models.JobTypeScenesGeneration:
				newStatus = models.ProjectStatusScenesGenerating
			case models.JobTypeVoiceGeneration, models.JobTypeImageGeneration:
				newStatus = models.ProjectStatusVoiceGenerating
			case models.JobTypeSubtitleGeneration:
				newStatus = models.ProjectStatusSubtitlesGenerating
			case models.JobTypeRender:
				newStatus = models.ProjectStatusRendering
			default:
				newStatus = models.ProjectStatusCreated
			}
			if err := s.projectRepo.UpdateStatus(ctx, projectID, newStatus, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

// MarkProjectFailed marks a project as failed
func (s *Service) MarkProjectFailed(ctx context.Context, projectID uuid.UUID, errorMsg string) error {
	ctx = logger.WithProjectID(ctx, projectID.String())
	logger.Error(ctx, "project failed", "error", errorMsg)
	return s.projectRepo.SetError(ctx, projectID, errorMsg)
}

// GetProject gets a project by ID
func (s *Service) GetProject(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	return s.projectRepo.GetByID(ctx, projectID)
}

// GetProjectManifest gets the full manifest for a project
func (s *Service) GetProjectManifest(ctx context.Context, projectID uuid.UUID) (*contracts.ManifestResponse, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	manifest := &contracts.ManifestResponse{
		Project: contracts.ProjectResponse{
			ProjectID:   project.ID,
			ExternalID:  project.ExternalID,
			Status:      project.Status,
			Topic:       project.Topic,
			CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	if project.Title != nil {
		manifest.Project.Title = *project.Title
	}
	if project.CurrentStep != nil {
		manifest.Project.CurrentStep = *project.CurrentStep
	}

	// Get script
	script, err := s.scriptRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if script != nil {
		manifest.Script = &contracts.ScriptResponse{
			Hook:             script.Hook,
			SetupText:        script.SetupText,
			BuildText:        script.BuildText,
			TurningPointText: script.TurningPointText,
			CollapseText:     script.CollapseText,
			ConclusionText:   script.ConclusionText,
			FullScript:       script.FullScript,
		}
	}

	// Get scenes
	sceneList, err := s.sceneRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, scene := range sceneList {
		var keywords []string
		if scene.Keywords != nil {
			json.Unmarshal(scene.Keywords, &keywords)
		}
		manifest.Scenes = append(manifest.Scenes, contracts.SceneResponse{
			SceneNumber:   scene.SceneNumber,
			NarrationText: scene.NarrationText,
			DurationSec:   scene.DurationSec,
			Mood:          scene.Mood,
			Keywords:      keywords,
			VisualPrompt:  scene.VisualPrompt,
			AssetStrategy: scene.AssetStrategy,
		})
	}

	// Get audio files
	audioFiles, err := s.audioRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for i, audio := range audioFiles {
		durationMs := 0
		if audio.DurationMs != nil {
			durationMs = *audio.DurationMs
		}
		manifest.AudioFiles = append(manifest.AudioFiles, contracts.AudioFileInfo{
			SceneNumber: i + 1,
			StorageKey:  audio.StorageKey,
			DurationMs:  durationMs,
		})
	}

	// Get assets
	assets, err := s.assetRepo.GetImagesByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for i, asset := range assets {
		manifest.Assets = append(manifest.Assets, contracts.AssetInfo{
			SceneNumber: i + 1,
			AssetType:   asset.AssetType,
			StorageKey:  asset.StorageKey,
			Provider:    asset.Provider,
		})
	}

	// Get render info
	renders, err := s.renderRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(renders) > 0 {
		latestRender := renders[0] // Already sorted by created_at DESC
		var metadata map[string]interface{}
		if latestRender.Metadata != nil {
			json.Unmarshal(latestRender.Metadata, &metadata)
		}
		durationSec := 0.0
		if d, ok := metadata["duration_sec"].(float64); ok {
			durationSec = d
		}
		manifest.Render = &contracts.RenderInfo{
			DraftVideoKey: latestRender.StorageKey,
			DurationSec:   durationSec,
		}
	}

	return manifest, nil
}
