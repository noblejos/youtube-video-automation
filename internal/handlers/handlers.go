package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/internal/images"
	"github.com/gama/youtube-video-automation/internal/jobs"
	"github.com/gama/youtube-video-automation/internal/render"
	"github.com/gama/youtube-video-automation/internal/scenes"
	"github.com/gama/youtube-video-automation/internal/scripts"
	"github.com/gama/youtube-video-automation/internal/subtitles"
	"github.com/gama/youtube-video-automation/internal/voice"
	"github.com/gama/youtube-video-automation/internal/workflow"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// JobHandlers registers all job handlers
type JobHandlers struct {
	scriptService   *scripts.Service
	sceneService    *scenes.Service
	voiceService    *voice.Service
	imageService    *images.Service
	subtitleService *subtitles.Service
	renderService   *render.Service
	workflowService *workflow.Service
	sceneRepo       *scenes.Repository
	useMockRender   bool
}

// NewJobHandlers creates a new job handlers instance
func NewJobHandlers(
	scriptService *scripts.Service,
	sceneService *scenes.Service,
	voiceService *voice.Service,
	imageService *images.Service,
	subtitleService *subtitles.Service,
	renderService *render.Service,
	workflowService *workflow.Service,
	sceneRepo *scenes.Repository,
	useMockRender bool,
) *JobHandlers {
	return &JobHandlers{
		scriptService:   scriptService,
		sceneService:    sceneService,
		voiceService:    voiceService,
		imageService:    imageService,
		subtitleService: subtitleService,
		renderService:   renderService,
		workflowService: workflowService,
		sceneRepo:       sceneRepo,
		useMockRender:   useMockRender,
	}
}

// RegisterHandlers registers all job handlers with the queue
func (h *JobHandlers) RegisterHandlers(queue jobs.Queue) {
	queue.RegisterHandler(models.JobTypeScriptGeneration, h.HandleScriptGeneration)
	queue.RegisterHandler(models.JobTypeScenesGeneration, h.HandleScenesGeneration)
	queue.RegisterHandler(models.JobTypeVoiceGeneration, h.HandleVoiceGeneration)
	queue.RegisterHandler(models.JobTypeImageGeneration, h.HandleImageGeneration)
	queue.RegisterHandler(models.JobTypeSubtitleGeneration, h.HandleSubtitleGeneration)
	queue.RegisterHandler(models.JobTypeRender, h.HandleRender)
	queue.RegisterHandler(models.JobTypeReviewPackage, h.HandleReviewPackage)
}

// HandleScriptGeneration handles script generation jobs
func (h *JobHandlers) HandleScriptGeneration(ctx context.Context, job *models.Job) error {
	payload, err := jobs.ParsePayload[jobs.ScriptJobPayload](job)
	if err != nil {
		return err
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	logger.Info(ctx, "handling script generation job")

	_, err = h.scriptService.Generate(ctx, payload.ProjectID, payload.Topic, payload.ChannelStyle, payload.TargetDuration)
	if err != nil {
		h.workflowService.MarkProjectFailed(ctx, payload.ProjectID, err.Error())
		return err
	}

	// Trigger next workflow step
	return h.workflowService.OnScriptGenerated(ctx, payload.ProjectID)
}

// HandleScenesGeneration handles scenes generation jobs
func (h *JobHandlers) HandleScenesGeneration(ctx context.Context, job *models.Job) error {
	payload, err := jobs.ParsePayload[jobs.SceneJobPayload](job)
	if err != nil {
		return err
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	logger.Info(ctx, "handling scenes generation job")

	// Get the script
	script, err := h.scriptService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}
	if script == nil {
		return fmt.Errorf("script not found for project %s", payload.ProjectID)
	}

	// Get project details for aspect ratio and style
	manifest, err := h.workflowService.GetProjectManifest(ctx, payload.ProjectID)
	if err != nil {
		return err
	}

	_, err = h.sceneService.Generate(ctx, payload.ProjectID, script,
		"9:16", // Default aspect ratio
		"dramatic_history_shorts",
		120, // Default duration
	)
	if err != nil {
		h.workflowService.MarkProjectFailed(ctx, payload.ProjectID, err.Error())
		return err
	}

	// Trigger next workflow step
	_ = manifest
	return h.workflowService.OnScenesGenerated(ctx, payload.ProjectID)
}

// HandleVoiceGeneration handles voice generation jobs
func (h *JobHandlers) HandleVoiceGeneration(ctx context.Context, job *models.Job) error {
	payload, err := jobs.ParsePayload[jobs.VoiceJobPayload](job)
	if err != nil {
		return err
	}

	if payload.SceneID == nil {
		return fmt.Errorf("scene_id is required for voice generation")
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	ctx = logger.WithSceneID(ctx, payload.SceneID.String())
	logger.Info(ctx, "handling voice generation job")

	// Get scene
	scene, err := h.sceneService.GetByID(ctx, *payload.SceneID)
	if err != nil {
		return err
	}

	// Get text to synthesize (prefer SSML if available)
	text := scene.NarrationText
	if scene.SSMLText != nil && *scene.SSMLText != "" {
		text = *scene.SSMLText
	}

	_, err = h.voiceService.Generate(ctx, payload.ProjectID, *payload.SceneID,
		text, payload.VoiceID, payload.Engine)
	if err != nil {
		return err
	}

	// Update scene status
	h.sceneService.UpdateStatus(ctx, *payload.SceneID, models.SceneStatusReady)

	// Check if all assets are complete
	return h.workflowService.CheckAssetCompletion(ctx, payload.ProjectID)
}

// HandleImageGeneration handles image generation jobs
func (h *JobHandlers) HandleImageGeneration(ctx context.Context, job *models.Job) error {
	payload, err := jobs.ParsePayload[jobs.ImageJobPayload](job)
	if err != nil {
		return err
	}

	if payload.SceneID == nil {
		return fmt.Errorf("scene_id is required for image generation")
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	ctx = logger.WithSceneID(ctx, payload.SceneID.String())
	logger.Info(ctx, "handling image generation job")

	// Get scene
	scene, err := h.sceneService.GetByID(ctx, *payload.SceneID)
	if err != nil {
		return err
	}

	_, err = h.imageService.Generate(ctx, payload.ProjectID, *payload.SceneID,
		scene.VisualPrompt, payload.AspectRatio, payload.StyleProfile)
	if err != nil {
		return err
	}

	// Check if all assets are complete
	return h.workflowService.CheckAssetCompletion(ctx, payload.ProjectID)
}

// HandleSubtitleGeneration handles subtitle generation jobs
func (h *JobHandlers) HandleSubtitleGeneration(ctx context.Context, job *models.Job) error {
	var payload struct {
		ProjectID uuid.UUID `json:"project_id"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	logger.Info(ctx, "handling subtitle generation job")

	// Get scenes
	sceneList, err := h.sceneService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}

	// Get audio files for durations and speech marks
	audioFiles, err := h.voiceService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}

	// Build maps from audio files
	audioDurations := make(map[uuid.UUID]int)
	speechMarksKeys := make(map[uuid.UUID]string)
	for _, audio := range audioFiles {
		if audio.DurationMs != nil {
			audioDurations[audio.SceneID] = *audio.DurationMs
		}
		if audio.SpeechMarksKey != nil && *audio.SpeechMarksKey != "" {
			speechMarksKeys[audio.SceneID] = *audio.SpeechMarksKey
		}
	}

	_, err = h.subtitleService.GenerateWithSpeechMarks(ctx, payload.ProjectID, sceneList, audioDurations, speechMarksKeys)
	if err != nil {
		h.workflowService.MarkProjectFailed(ctx, payload.ProjectID, err.Error())
		return err
	}

	// Trigger next workflow step
	return h.workflowService.OnSubtitlesGenerated(ctx, payload.ProjectID)
}

// HandleRender handles render jobs
func (h *JobHandlers) HandleRender(ctx context.Context, job *models.Job) error {
	var payload struct {
		ProjectID uuid.UUID `json:"project_id"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	logger.Info(ctx, "handling render job")

	// Get project manifest for title
	manifest, err := h.workflowService.GetProjectManifest(ctx, payload.ProjectID)
	if err != nil {
		return err
	}

	// Get scenes
	sceneList, err := h.sceneService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}

	// Get audio files
	audioFiles, err := h.voiceService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}
	audioByScene := make(map[uuid.UUID]string)
	for _, audio := range audioFiles {
		audioByScene[audio.SceneID] = audio.StorageKey
	}

	// Get image assets
	imageAssets, err := h.imageService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}
	imageByScene := make(map[uuid.UUID]string)
	for _, asset := range imageAssets {
		if asset.SceneID != nil {
			imageByScene[*asset.SceneID] = asset.StorageKey
		}
	}

	// Get subtitles
	subtitle, err := h.subtitleService.GetByProjectID(ctx, payload.ProjectID)
	if err != nil {
		return err
	}

	// Build scene assets
	var renderScenes []render.SceneAssets
	for _, scene := range sceneList {
		audioKey, ok := audioByScene[scene.ID]
		if !ok {
			return fmt.Errorf("audio not found for scene %d", scene.SceneNumber)
		}
		imageKey, ok := imageByScene[scene.ID]
		if !ok {
			return fmt.Errorf("image not found for scene %d", scene.SceneNumber)
		}

		renderScenes = append(renderScenes, render.SceneAssets{
			SceneNumber: scene.SceneNumber,
			ImageKey:    imageKey,
			AudioKey:    audioKey,
			DurationSec: scene.DurationSec,
		})
	}

	title := manifest.Project.Topic
	if manifest.Project.Title != "" {
		title = manifest.Project.Title
	}

	// Render (or mock render)
	if h.useMockRender {
		_, err = h.renderService.MockRender(ctx, payload.ProjectID, title, renderScenes, subtitle.StorageKey)
	} else {
		_, err = h.renderService.Render(ctx, payload.ProjectID, title, renderScenes, subtitle.StorageKey)
	}

	if err != nil {
		h.workflowService.MarkProjectFailed(ctx, payload.ProjectID, err.Error())
		return err
	}

	// Trigger next workflow step
	return h.workflowService.OnRenderComplete(ctx, payload.ProjectID)
}

// HandleReviewPackage handles review package jobs
func (h *JobHandlers) HandleReviewPackage(ctx context.Context, job *models.Job) error {
	var payload struct {
		ProjectID uuid.UUID `json:"project_id"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	ctx = logger.WithProjectID(ctx, payload.ProjectID.String())
	logger.Info(ctx, "handling review package job")

	// For now, just mark the project as ready for review
	// In a full implementation, this would package all artifacts

	return h.workflowService.OnReviewPackageReady(ctx, payload.ProjectID)
}
