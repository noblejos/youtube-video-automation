package contracts

import (
	"encoding/json"

	"github.com/google/uuid"
)

// CreateProjectRequest represents a request to create a new project
type CreateProjectRequest struct {
	Topic             string `json:"topic,omitempty"`              // Required if script not provided
	Title             string `json:"title,omitempty"`              // Optional video title
	Script            string `json:"script,omitempty"`             // Full script text (skips AI generation)
	ChannelStyle      string `json:"channel_style,omitempty"`
	TargetDurationSec int    `json:"target_duration_sec,omitempty"`
	AspectRatio       string `json:"aspect_ratio,omitempty"`
	ReviewRequired    *bool  `json:"review_required,omitempty"`
	VoiceID           string `json:"voice_id,omitempty"`           // AWS Polly voice ID (e.g., "Matthew", "Joanna")
	VoiceEngine       string `json:"voice_engine,omitempty"`       // "standard", "neural", "generative", or "long-form"
}

// ProjectResponse represents a project response
type ProjectResponse struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ExternalID  string    `json:"external_id"`
	Status      string    `json:"status"`
	CurrentStep string    `json:"current_step,omitempty"`
	Topic       string    `json:"topic"`
	Title       string    `json:"title,omitempty"`
	CreatedAt   string    `json:"created_at"`
}

// ScriptResponse represents a script generation result
type ScriptResponse struct {
	Title            string `json:"title"`
	Hook             string `json:"hook"`
	SetupText        string `json:"setup_text"`
	BuildText        string `json:"build_text"`
	TurningPointText string `json:"turning_point_text"`
	CollapseText     string `json:"collapse_text"`
	ConclusionText   string `json:"conclusion_text"`
	FullScript       string `json:"full_script"`
}

// VisualConfig represents visual generation configuration for a scene
type VisualConfig struct {
	Type           string `json:"type"`                      // ai_generated, stock, archive
	Prompt         string `json:"prompt"`                    // Image generation prompt
	NegativePrompt string `json:"negative_prompt,omitempty"` // What to avoid
	CameraMotion   string `json:"camera_motion,omitempty"`   // slow_zoom_in, pan_left, etc.
	TransitionIn   string `json:"transition_in,omitempty"`   // fade, cut, dissolve
	TransitionOut  string `json:"transition_out,omitempty"`  // fade, cut, dissolve
}

// AudioConfig represents audio configuration for a scene
type AudioConfig struct {
	Voice               string `json:"voice,omitempty"`
	Engine              string `json:"engine,omitempty"`
	BackgroundMusicMood string `json:"background_music_mood,omitempty"`
}

// SubtitleConfig represents subtitle configuration for a scene
type SubtitleConfig struct {
	Enabled bool   `json:"enabled"`
	Style   string `json:"style,omitempty"` // lower_third_center, etc.
}

// SceneResponse represents a single scene
type SceneResponse struct {
	SceneNumber   int      `json:"scene_number"`
	StartTimeSec  float64  `json:"start_time_sec,omitempty"`
	DurationSec   float64  `json:"duration_sec"`
	StoryRole     string   `json:"story_role,omitempty"` // hook, setup, build, turning_point, collapse, conclusion
	Mood          string   `json:"mood"`
	EnergyLevel   string   `json:"energy_level,omitempty"` // low, medium, high
	NarrationText string   `json:"narration_text"`
	SSMLText      string   `json:"ssml_text,omitempty"`
	Keywords      []string `json:"keywords"`
	// Legacy field for backwards compatibility
	VisualPrompt  string `json:"visual_prompt,omitempty"`
	AssetStrategy string `json:"asset_strategy,omitempty"`
	// New rich configuration
	Visual    *VisualConfig   `json:"visual,omitempty"`
	Audio     *AudioConfig    `json:"audio,omitempty"`
	Subtitles *SubtitleConfig `json:"subtitles,omitempty"`
}

// ScenesResponse represents all scenes for a project
type ScenesResponse struct {
	ProjectID    uuid.UUID       `json:"project_id"`
	AspectRatio  string          `json:"aspect_ratio"`
	StyleProfile string          `json:"style_profile"`
	Scenes       []SceneResponse `json:"scenes"`
}

// VoiceRequest represents a voice generation request
type VoiceRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	SceneID   uuid.UUID `json:"scene_id"`
	Provider  string    `json:"provider"`
	VoiceID   string    `json:"voice_id"`
	Engine    string    `json:"engine"`
	TextType  string    `json:"text_type"`
	Text      string    `json:"text"`
}

// VoiceResult represents a voice generation result
type VoiceResult struct {
	StorageKey     string `json:"storage_key"`
	DurationMs     int    `json:"duration_ms"`
	SpeechMarksKey string `json:"speech_marks_key,omitempty"`
}

// ImageRequest represents an image generation request
type ImageRequest struct {
	ProjectID    uuid.UUID `json:"project_id"`
	SceneID      uuid.UUID `json:"scene_id"`
	Provider     string    `json:"provider"`
	Prompt       string    `json:"prompt"`
	AspectRatio  string    `json:"aspect_ratio"`
	StyleProfile string    `json:"style_profile"`
}

// ImageResult represents an image generation result
type ImageResult struct {
	StorageKey string                 `json:"storage_key"`
	Provider   string                 `json:"provider"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RenderManifest represents the render output manifest
type RenderManifest struct {
	ProjectID     uuid.UUID          `json:"project_id"`
	Title         string             `json:"title"`
	AspectRatio   string             `json:"aspect_ratio"`
	SceneCount    int                `json:"scene_count"`
	SceneClips    []SceneClipInfo    `json:"scene_clips"`
	SubtitleKey   string             `json:"subtitle_key"`
	DraftVideoKey string             `json:"draft_video_key"`
	DurationSec   float64            `json:"duration_sec"`
}

// SceneClipInfo represents info about a rendered scene clip
type SceneClipInfo struct {
	SceneNumber int     `json:"scene_number"`
	ClipKey     string  `json:"clip_key"`
	ImageKey    string  `json:"image_key"`
	AudioKey    string  `json:"audio_key"`
	DurationSec float64 `json:"duration_sec"`
}

// ReviewPackage represents the review output package
type ReviewPackage struct {
	ProjectID      uuid.UUID `json:"project_id"`
	DraftVideoKey  string    `json:"draft_video_key"`
	SubtitleKey    string    `json:"subtitle_key"`
	ManifestKey    string    `json:"manifest_key"`
	SceneClipsDir  string    `json:"scene_clips_dir"`
	AssetsDir      string    `json:"assets_dir"`
	AudioDir       string    `json:"audio_dir"`
}

// ApproveRequest represents a project approval request
type ApproveRequest struct {
	Notes   string `json:"notes,omitempty"`
	ActedBy string `json:"acted_by,omitempty"`
}

// RejectRequest represents a project rejection request
type RejectRequest struct {
	Notes   string `json:"notes,omitempty"`
	ActedBy string `json:"acted_by,omitempty"`
}

// RetryRequest represents a retry request for a failed project
type RetryRequest struct {
	Stage string `json:"stage,omitempty"` // Optional: specific stage to retry
}

// JobPayload represents generic job payload
type JobPayload struct {
	ProjectID uuid.UUID       `json:"project_id"`
	SceneID   *uuid.UUID      `json:"scene_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ManifestResponse represents a full project manifest
type ManifestResponse struct {
	Project     ProjectResponse   `json:"project"`
	Script      *ScriptResponse   `json:"script,omitempty"`
	Scenes      []SceneResponse   `json:"scenes,omitempty"`
	AudioFiles  []AudioFileInfo   `json:"audio_files,omitempty"`
	Assets      []AssetInfo       `json:"assets,omitempty"`
	Subtitles   *SubtitleInfo     `json:"subtitles,omitempty"`
	Render      *RenderInfo       `json:"render,omitempty"`
}

// AudioFileInfo represents audio file info in manifest
type AudioFileInfo struct {
	SceneNumber int    `json:"scene_number"`
	StorageKey  string `json:"storage_key"`
	DurationMs  int    `json:"duration_ms"`
}

// AssetInfo represents asset info in manifest
type AssetInfo struct {
	SceneNumber int    `json:"scene_number"`
	AssetType   string `json:"asset_type"`
	StorageKey  string `json:"storage_key"`
	Provider    string `json:"provider"`
}

// SubtitleInfo represents subtitle info in manifest
type SubtitleInfo struct {
	Format     string `json:"format"`
	StorageKey string `json:"storage_key"`
}

// RenderInfo represents render info in manifest
type RenderInfo struct {
	DraftVideoKey string  `json:"draft_video_key"`
	DurationSec   float64 `json:"duration_sec"`
}
