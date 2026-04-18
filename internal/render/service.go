package render

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Config holds render configuration
type Config struct {
	FFmpegPath string
	Width      int
	Height     int
	FPS        int
	TempDir    string
}

// Repository handles render persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new render repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new render record
func (r *Repository) Create(ctx context.Context, render *models.Render) error {
	if render.ID == uuid.Nil {
		render.ID = uuid.New()
	}
	render.CreatedAt = time.Now()

	query := `
		INSERT INTO renders (id, project_id, render_type, storage_key, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		render.ID,
		render.ProjectID,
		render.RenderType,
		render.StorageKey,
		render.Metadata,
		render.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create render: %w", err)
	}

	return nil
}

// GetByProjectID retrieves renders for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.Render, error) {
	query := `
		SELECT id, project_id, render_type, storage_key, metadata, created_at
		FROM renders WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get renders: %w", err)
	}
	defer rows.Close()

	var renders []*models.Render
	for rows.Next() {
		var render models.Render
		err := rows.Scan(
			&render.ID,
			&render.ProjectID,
			&render.RenderType,
			&render.StorageKey,
			&render.Metadata,
			&render.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan render: %w", err)
		}
		renders = append(renders, &render)
	}

	return renders, nil
}

// Service handles video rendering
type Service struct {
	repo    *Repository
	storage storage.Provider
	config  Config
}

// NewService creates a new render service
func NewService(repo *Repository, storage storage.Provider, config Config) *Service {
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	return &Service{
		repo:    repo,
		storage: storage,
		config:  config,
	}
}

// SceneAssets holds the assets needed to render a scene
type SceneAssets struct {
	SceneNumber int
	ImageKey    string
	AudioKey    string
	DurationSec float64
}

// Render renders the final video for a project
func (s *Service) Render(ctx context.Context, projectID uuid.UUID, title string, scenes []SceneAssets, subtitleKey string) (*contracts.RenderManifest, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "render")

	logger.Info(ctx, "starting render", "scene_count", len(scenes))

	// Create temp directory for this render
	tempDir := filepath.Join(s.config.TempDir, "render-"+projectID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download subtitles if available
	var subtitlePath string
	if subtitleKey != "" {
		subtitleData, err := s.storage.Get(ctx, subtitleKey)
		if err != nil {
			logger.Warn(ctx, "failed to get subtitles, rendering without", "error", err)
		} else {
			subtitlePath = filepath.Join(tempDir, "subtitles.srt")
			if err := os.WriteFile(subtitlePath, subtitleData, 0644); err != nil {
				logger.Warn(ctx, "failed to write subtitles, rendering without", "error", err)
				subtitlePath = ""
			}
		}
	}

	manifest := &contracts.RenderManifest{
		ProjectID:   projectID,
		Title:       title,
		AspectRatio: fmt.Sprintf("%d:%d", s.config.Width, s.config.Height),
		SceneCount:  len(scenes),
	}

	var sceneClipPaths []string
	var totalDuration float64

	// Render each scene
	for _, scene := range scenes {
		clipPath, err := s.renderScene(ctx, tempDir, scene)
		if err != nil {
			return nil, fmt.Errorf("failed to render scene %d: %w", scene.SceneNumber, err)
		}
		sceneClipPaths = append(sceneClipPaths, clipPath)

		// Store scene clip
		clipKey := fmt.Sprintf("projects/%s/render/scene_%03d.mp4", projectID, scene.SceneNumber)
		clipData, err := os.ReadFile(clipPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read scene clip: %w", err)
		}
		if err := s.storage.Put(ctx, clipKey, "video/mp4", clipData); err != nil {
			return nil, fmt.Errorf("failed to store scene clip: %w", err)
		}

		manifest.SceneClips = append(manifest.SceneClips, contracts.SceneClipInfo{
			SceneNumber: scene.SceneNumber,
			ClipKey:     clipKey,
			ImageKey:    scene.ImageKey,
			AudioKey:    scene.AudioKey,
			DurationSec: scene.DurationSec,
		})

		totalDuration += scene.DurationSec
	}

	// Concatenate all scene clips
	concatPath := filepath.Join(tempDir, "concat.mp4")
	if err := s.concatenateClips(ctx, sceneClipPaths, concatPath); err != nil {
		return nil, fmt.Errorf("failed to concatenate clips: %w", err)
	}

	// Burn subtitles into the final video
	draftPath := filepath.Join(tempDir, "draft.mp4")
	if subtitlePath != "" {
		if err := s.burnSubtitles(ctx, concatPath, subtitlePath, draftPath); err != nil {
			logger.Warn(ctx, "failed to burn subtitles, using video without", "error", err)
			// Fall back to concat without subtitles
			if err := os.Rename(concatPath, draftPath); err != nil {
				return nil, fmt.Errorf("failed to rename concat to draft: %w", err)
			}
		}
	} else {
		// No subtitles, just rename
		if err := os.Rename(concatPath, draftPath); err != nil {
			return nil, fmt.Errorf("failed to rename concat to draft: %w", err)
		}
	}

	// Store final draft
	draftKey := fmt.Sprintf("projects/%s/render/draft.mp4", projectID)
	draftData, err := os.ReadFile(draftPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read draft: %w", err)
	}
	if err := s.storage.Put(ctx, draftKey, "video/mp4", draftData); err != nil {
		return nil, fmt.Errorf("failed to store draft: %w", err)
	}

	manifest.SubtitleKey = subtitleKey
	manifest.DraftVideoKey = draftKey
	manifest.DurationSec = totalDuration

	// Store manifest
	manifestKey := fmt.Sprintf("projects/%s/render/manifest.json", projectID)
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := s.storage.Put(ctx, manifestKey, "application/json", manifestData); err != nil {
		return nil, fmt.Errorf("failed to store manifest: %w", err)
	}

	// Create render record
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"duration_sec": totalDuration,
		"scene_count":  len(scenes),
	})
	render := &models.Render{
		ID:         uuid.New(),
		ProjectID:  projectID,
		RenderType: "draft",
		StorageKey: draftKey,
		Metadata:   metadataJSON,
	}
	if err := s.repo.Create(ctx, render); err != nil {
		return nil, fmt.Errorf("failed to save render: %w", err)
	}

	logger.Info(ctx, "render completed successfully",
		"draft_key", draftKey,
		"duration_sec", totalDuration,
	)

	return manifest, nil
}

func (s *Service) renderScene(ctx context.Context, tempDir string, scene SceneAssets) (string, error) {
	// Download image and audio
	imageData, err := s.storage.Get(ctx, scene.ImageKey)
	if err != nil {
		return "", fmt.Errorf("failed to get image: %w", err)
	}
	imagePath := filepath.Join(tempDir, fmt.Sprintf("scene_%03d.png", scene.SceneNumber))
	if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	audioData, err := s.storage.Get(ctx, scene.AudioKey)
	if err != nil {
		return "", fmt.Errorf("failed to get audio: %w", err)
	}
	audioPath := filepath.Join(tempDir, fmt.Sprintf("scene_%03d.mp3", scene.SceneNumber))
	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		return "", fmt.Errorf("failed to write audio: %w", err)
	}

	// Output path
	outputPath := filepath.Join(tempDir, fmt.Sprintf("scene_%03d.mp4", scene.SceneNumber))

	// FFmpeg command to create video from image + audio with zoom effect
	// Using a slow zoom (Ken Burns effect)
	args := []string{
		"-loop", "1",
		"-i", imagePath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-c:a", "aac",
		"-b:a", "192k",
		"-pix_fmt", "yuv420p",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,zoompan=z='min(zoom+0.0005,1.2)':d=%d:x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':s=%dx%d",
			s.config.Width, s.config.Height,
			s.config.Width, s.config.Height,
			int(scene.DurationSec*float64(s.config.FPS)),
			s.config.Width, s.config.Height),
		"-shortest",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, s.config.FFmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(ctx, "FFmpeg error", "output", string(output), "error", err)
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	return outputPath, nil
}

func (s *Service) concatenateClips(ctx context.Context, clipPaths []string, outputPath string) error {
	// Create concat file
	concatFilePath := filepath.Join(filepath.Dir(outputPath), "concat.txt")
	var concatContent string
	for _, path := range clipPaths {
		concatContent += fmt.Sprintf("file '%s'\n", path)
	}
	if err := os.WriteFile(concatFilePath, []byte(concatContent), 0644); err != nil {
		return fmt.Errorf("failed to write concat file: %w", err)
	}

	// FFmpeg concat command
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatFilePath,
		"-c", "copy",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, s.config.FFmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(ctx, "FFmpeg concat error", "output", string(output), "error", err)
		return fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	return nil
}

func (s *Service) burnSubtitles(ctx context.Context, inputPath, subtitlePath, outputPath string) error {
	logger.Info(ctx, "burning subtitles into video", "subtitle_path", subtitlePath)

	// FFmpeg command to burn subtitles with styling
	// Using subtitles filter with force_style for readability
	// - White text with black outline for contrast on any background
	// - Font size 24 for mobile-friendly reading
	// - Alignment=2 forces bottom-center positioning
	// - MarginV=40 provides space from bottom edge
	// - Bold for better visibility
	// Note: ASS colors are in &HAABBGGRR format (alpha, blue, green, red)
	// White = &H00FFFFFF (00 alpha=opaque, FF blue, FF green, FF red)
	// The force_style value must be wrapped in single quotes for FFmpeg
	subtitleFilter := fmt.Sprintf("subtitles=%s:force_style='Fontsize=10,PrimaryColour=&H00FFFFFF,OutlineColour=&H00000000,BorderStyle=1,Outline=1,Shadow=1,Bold=1,Alignment=2,MarginV=20'",
		subtitlePath)

	args := []string{
		"-i", inputPath,
		"-vf", subtitleFilter,
		"-c:a", "copy",
		"-c:v", "libx264",
		"-preset", "fast",
		"-y",
		outputPath,
	}

	logger.Info(ctx, "running FFmpeg subtitle burn", "filter", subtitleFilter)

	cmd := exec.CommandContext(ctx, s.config.FFmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(ctx, "FFmpeg subtitle burn error", "output", string(output), "error", err)
		return fmt.Errorf("ffmpeg subtitle burn failed: %w", err)
	}

	logger.Info(ctx, "subtitles burned successfully")
	return nil
}

// MockRender creates a mock render for testing without FFmpeg
func (s *Service) MockRender(ctx context.Context, projectID uuid.UUID, title string, scenes []SceneAssets, subtitleKey string) (*contracts.RenderManifest, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "mock_render")

	logger.Info(ctx, "creating mock render", "scene_count", len(scenes))

	manifest := &contracts.RenderManifest{
		ProjectID:   projectID,
		Title:       title,
		AspectRatio: "9:16",
		SceneCount:  len(scenes),
	}

	var totalDuration float64

	for _, scene := range scenes {
		clipKey := fmt.Sprintf("projects/%s/render/scene_%03d.mp4", projectID, scene.SceneNumber)

		// Store a placeholder for the scene clip
		placeholder := []byte("mock video data")
		if err := s.storage.Put(ctx, clipKey, "video/mp4", placeholder); err != nil {
			return nil, fmt.Errorf("failed to store mock clip: %w", err)
		}

		manifest.SceneClips = append(manifest.SceneClips, contracts.SceneClipInfo{
			SceneNumber: scene.SceneNumber,
			ClipKey:     clipKey,
			ImageKey:    scene.ImageKey,
			AudioKey:    scene.AudioKey,
			DurationSec: scene.DurationSec,
		})

		totalDuration += scene.DurationSec
	}

	// Store mock draft
	draftKey := fmt.Sprintf("projects/%s/render/draft.mp4", projectID)
	if err := s.storage.Put(ctx, draftKey, "video/mp4", []byte("mock draft video")); err != nil {
		return nil, fmt.Errorf("failed to store mock draft: %w", err)
	}

	manifest.SubtitleKey = subtitleKey
	manifest.DraftVideoKey = draftKey
	manifest.DurationSec = totalDuration

	// Store manifest
	manifestKey := fmt.Sprintf("projects/%s/render/manifest.json", projectID)
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := s.storage.Put(ctx, manifestKey, "application/json", manifestData); err != nil {
		return nil, fmt.Errorf("failed to store manifest: %w", err)
	}

	// Create render record
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"duration_sec": totalDuration,
		"scene_count":  len(scenes),
		"mock":         true,
	})
	render := &models.Render{
		ID:         uuid.New(),
		ProjectID:  projectID,
		RenderType: "draft",
		StorageKey: draftKey,
		Metadata:   metadataJSON,
	}
	if err := s.repo.Create(ctx, render); err != nil {
		return nil, fmt.Errorf("failed to save render: %w", err)
	}

	logger.Info(ctx, "mock render completed", "duration_sec", totalDuration)

	return manifest, nil
}
