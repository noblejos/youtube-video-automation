package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Project statuses
const (
	ProjectStatusCreated            = "CREATED"
	ProjectStatusScriptGenerating   = "SCRIPT_GENERATING"
	ProjectStatusScriptReady        = "SCRIPT_READY"
	ProjectStatusScenesGenerating   = "SCENES_GENERATING"
	ProjectStatusScenesReady        = "SCENES_READY"
	ProjectStatusVoiceGenerating    = "VOICE_GENERATING"
	ProjectStatusVoiceReady         = "VOICE_READY"
	ProjectStatusAssetsGenerating   = "ASSETS_GENERATING"
	ProjectStatusAssetsReady        = "ASSETS_READY"
	ProjectStatusSubtitlesGenerating = "SUBTITLES_GENERATING"
	ProjectStatusSubtitlesReady     = "SUBTITLES_READY"
	ProjectStatusRendering          = "RENDERING"
	ProjectStatusRenderReady        = "RENDER_READY"
	ProjectStatusReviewPackaged     = "REVIEW_PACKAGED"
	ProjectStatusInReview           = "IN_REVIEW"
	ProjectStatusApproved           = "APPROVED"
	ProjectStatusRejected           = "REJECTED"
	ProjectStatusPublishing         = "PUBLISHING"
	ProjectStatusPublished          = "PUBLISHED"
	ProjectStatusFailed             = "FAILED"
	ProjectStatusCancelled          = "CANCELLED"
)

// Scene statuses
const (
	SceneStatusPending = "PENDING"
	SceneStatusReady   = "READY"
	SceneStatusFailed  = "FAILED"
)

// Job statuses
const (
	JobStatusQueued    = "QUEUED"
	JobStatusRunning   = "RUNNING"
	JobStatusSucceeded = "SUCCEEDED"
	JobStatusFailed    = "FAILED"
	JobStatusRetrying  = "RETRYING"
	JobStatusCancelled = "CANCELLED"
)

// Job types
const (
	JobTypeScriptGeneration   = "SCRIPT_GENERATION"
	JobTypeScenesGeneration   = "SCENES_GENERATION"
	JobTypeVoiceGeneration    = "VOICE_GENERATION"
	JobTypeImageGeneration    = "IMAGE_GENERATION"
	JobTypeSubtitleGeneration = "SUBTITLE_GENERATION"
	JobTypeRender             = "RENDER"
	JobTypeReviewPackage      = "REVIEW_PACKAGE"
)

// Review actions
const (
	ReviewActionApprove  = "APPROVE"
	ReviewActionReject   = "REJECT"
	ReviewActionRerender = "RERENDER"
)

// Asset types
const (
	AssetTypeImage      = "IMAGE"
	AssetTypeAudio      = "AUDIO"
	AssetTypeVideo      = "VIDEO"
	AssetTypeSubtitle   = "SUBTITLE"
	AssetTypeBackground = "BACKGROUND"
)

// Project represents a video generation project
type Project struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	ExternalID        string     `json:"external_id" db:"external_id"`
	Topic             string     `json:"topic" db:"topic"`
	Title             *string    `json:"title,omitempty" db:"title"`
	ChannelStyle      string     `json:"channel_style" db:"channel_style"`
	TargetDurationSec int        `json:"target_duration_sec" db:"target_duration_sec"`
	AspectRatio       string     `json:"aspect_ratio" db:"aspect_ratio"`
	Status            string     `json:"status" db:"status"`
	ReviewRequired    bool       `json:"review_required" db:"review_required"`
	CurrentStep       *string    `json:"current_step,omitempty" db:"current_step"`
	ErrorMessage      *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// Script represents a generated script for a project
type Script struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	ProjectID         uuid.UUID        `json:"project_id" db:"project_id"`
	Hook              string           `json:"hook" db:"hook"`
	SetupText         string           `json:"setup_text" db:"setup_text"`
	BuildText         string           `json:"build_text" db:"build_text"`
	TurningPointText  string           `json:"turning_point_text" db:"turning_point_text"`
	CollapseText      string           `json:"collapse_text" db:"collapse_text"`
	ConclusionText    string           `json:"conclusion_text" db:"conclusion_text"`
	FullScript        string           `json:"full_script" db:"full_script"`
	RawModelResponse  json.RawMessage  `json:"raw_model_response,omitempty" db:"raw_model_response"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

// Scene represents a single scene in a video
type Scene struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	ProjectID      uuid.UUID       `json:"project_id" db:"project_id"`
	SceneNumber    int             `json:"scene_number" db:"scene_number"`
	Status         string          `json:"status" db:"status"`
	StoryRole      string          `json:"story_role" db:"story_role"`           // hook, setup, build, turning_point, collapse, conclusion
	EnergyLevel    string          `json:"energy_level" db:"energy_level"`       // low, medium, high
	NarrationText  string          `json:"narration_text" db:"narration_text"`
	SSMLText       *string         `json:"ssml_text,omitempty" db:"ssml_text"`
	DurationSec    float64         `json:"duration_sec" db:"duration_sec"`
	StartTimeSec   float64         `json:"start_time_sec" db:"start_time_sec"`
	Mood           string          `json:"mood" db:"mood"`
	Keywords       json.RawMessage `json:"keywords" db:"keywords"`
	VisualPrompt   string          `json:"visual_prompt" db:"visual_prompt"`
	NegativePrompt string          `json:"negative_prompt" db:"negative_prompt"`
	CameraMotion   string          `json:"camera_motion" db:"camera_motion"`
	TransitionIn   string          `json:"transition_in" db:"transition_in"`
	TransitionOut  string          `json:"transition_out" db:"transition_out"`
	AssetStrategy  string          `json:"asset_strategy" db:"asset_strategy"`
	TransitionType *string         `json:"transition_type,omitempty" db:"transition_type"`
	AudioConfig    json.RawMessage `json:"audio_config,omitempty" db:"audio_config"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// Asset represents a media asset (image, audio, video)
type Asset struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	ProjectID  uuid.UUID       `json:"project_id" db:"project_id"`
	SceneID    *uuid.UUID      `json:"scene_id,omitempty" db:"scene_id"`
	AssetType  string          `json:"asset_type" db:"asset_type"`
	Provider   string          `json:"provider" db:"provider"`
	StorageKey string          `json:"storage_key" db:"storage_key"`
	MimeType   string          `json:"mime_type" db:"mime_type"`
	SourceURL  *string         `json:"source_url,omitempty" db:"source_url"`
	PromptUsed *string         `json:"prompt_used,omitempty" db:"prompt_used"`
	Metadata   json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// AudioFile represents a generated audio file
type AudioFile struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ProjectID      uuid.UUID `json:"project_id" db:"project_id"`
	SceneID        uuid.UUID `json:"scene_id" db:"scene_id"`
	Provider       string    `json:"provider" db:"provider"`
	VoiceID        string    `json:"voice_id" db:"voice_id"`
	Engine         string    `json:"engine" db:"engine"`
	StorageKey     string    `json:"storage_key" db:"storage_key"`
	DurationMs     *int      `json:"duration_ms,omitempty" db:"duration_ms"`
	SpeechMarksKey *string   `json:"speech_marks_key,omitempty" db:"speech_marks_key"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Subtitle represents generated subtitles
type Subtitle struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ProjectID  uuid.UUID `json:"project_id" db:"project_id"`
	Format     string    `json:"format" db:"format"`
	StorageKey string    `json:"storage_key" db:"storage_key"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Render represents a rendered video
type Render struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	ProjectID  uuid.UUID       `json:"project_id" db:"project_id"`
	RenderType string          `json:"render_type" db:"render_type"`
	StorageKey string          `json:"storage_key" db:"storage_key"`
	Metadata   json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// Job represents a background job
type Job struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	ProjectID    uuid.UUID       `json:"project_id" db:"project_id"`
	JobType      string          `json:"job_type" db:"job_type"`
	Status       string          `json:"status" db:"status"`
	AttemptCount int             `json:"attempt_count" db:"attempt_count"`
	MaxAttempts  int             `json:"max_attempts" db:"max_attempts"`
	Payload      json.RawMessage `json:"payload" db:"payload"`
	Result       json.RawMessage `json:"result,omitempty" db:"result"`
	ErrorMessage *string         `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// ReviewAction represents a review action taken on a project
type ReviewAction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	Action    string    `json:"action" db:"action"`
	Notes     *string   `json:"notes,omitempty" db:"notes"`
	ActedBy   *string   `json:"acted_by,omitempty" db:"acted_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
