package voice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Repository handles audio file persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new audio file repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new audio file record
func (r *Repository) Create(ctx context.Context, audio *models.AudioFile) error {
	if audio.ID == uuid.Nil {
		audio.ID = uuid.New()
	}
	audio.CreatedAt = time.Now()

	query := `
		INSERT INTO audio_files (
			id, project_id, scene_id, provider, voice_id, engine,
			storage_key, duration_ms, speech_marks_key, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.pool.Exec(ctx, query,
		audio.ID,
		audio.ProjectID,
		audio.SceneID,
		audio.Provider,
		audio.VoiceID,
		audio.Engine,
		audio.StorageKey,
		audio.DurationMs,
		audio.SpeechMarksKey,
		audio.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create audio file: %w", err)
	}

	return nil
}

// GetBySceneID retrieves the audio file for a scene
func (r *Repository) GetBySceneID(ctx context.Context, sceneID uuid.UUID) (*models.AudioFile, error) {
	query := `
		SELECT id, project_id, scene_id, provider, voice_id, engine,
			storage_key, duration_ms, speech_marks_key, created_at
		FROM audio_files WHERE scene_id = $1
		ORDER BY created_at DESC LIMIT 1
	`

	var audio models.AudioFile
	err := r.pool.QueryRow(ctx, query, sceneID).Scan(
		&audio.ID,
		&audio.ProjectID,
		&audio.SceneID,
		&audio.Provider,
		&audio.VoiceID,
		&audio.Engine,
		&audio.StorageKey,
		&audio.DurationMs,
		&audio.SpeechMarksKey,
		&audio.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get audio file: %w", err)
	}

	return &audio, nil
}

// GetByProjectID retrieves all audio files for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*models.AudioFile, error) {
	query := `
		SELECT af.id, af.project_id, af.scene_id, af.provider, af.voice_id, af.engine,
			af.storage_key, af.duration_ms, af.speech_marks_key, af.created_at
		FROM audio_files af
		JOIN scenes s ON af.scene_id = s.id
		WHERE af.project_id = $1
		ORDER BY s.scene_number
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio files: %w", err)
	}
	defer rows.Close()

	var audioFiles []*models.AudioFile
	for rows.Next() {
		var audio models.AudioFile
		err := rows.Scan(
			&audio.ID,
			&audio.ProjectID,
			&audio.SceneID,
			&audio.Provider,
			&audio.VoiceID,
			&audio.Engine,
			&audio.StorageKey,
			&audio.DurationMs,
			&audio.SpeechMarksKey,
			&audio.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audio file: %w", err)
		}
		audioFiles = append(audioFiles, &audio)
	}

	return audioFiles, nil
}

// DeleteBySceneID deletes audio files for a scene
func (r *Repository) DeleteBySceneID(ctx context.Context, sceneID uuid.UUID) error {
	query := `DELETE FROM audio_files WHERE scene_id = $1`

	_, err := r.pool.Exec(ctx, query, sceneID)
	if err != nil {
		return fmt.Errorf("failed to delete audio files: %w", err)
	}

	return nil
}

// CountByProjectID counts audio files for a project
func (r *Repository) CountByProjectID(ctx context.Context, projectID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM audio_files WHERE project_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audio files: %w", err)
	}

	return count, nil
}
