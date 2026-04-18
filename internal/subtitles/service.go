package subtitles

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// SpeechMark represents a word timing from Polly
type SpeechMark struct {
	Time  int    `json:"time"`  // Time in milliseconds
	Type  string `json:"type"`  // "word" or "sentence"
	Start int    `json:"start"` // Start character offset
	End   int    `json:"end"`   // End character offset
	Value string `json:"value"` // The word or sentence text
}

// Repository handles subtitle persistence
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new subtitle repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create creates a new subtitle record
func (r *Repository) Create(ctx context.Context, subtitle *models.Subtitle) error {
	if subtitle.ID == uuid.Nil {
		subtitle.ID = uuid.New()
	}
	subtitle.CreatedAt = time.Now()

	query := `
		INSERT INTO subtitles (id, project_id, format, storage_key, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		subtitle.ID,
		subtitle.ProjectID,
		subtitle.Format,
		subtitle.StorageKey,
		subtitle.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subtitle: %w", err)
	}

	return nil
}

// GetByProjectID retrieves subtitles for a project
func (r *Repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Subtitle, error) {
	query := `
		SELECT id, project_id, format, storage_key, created_at
		FROM subtitles WHERE project_id = $1
		ORDER BY created_at DESC LIMIT 1
	`

	var subtitle models.Subtitle
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&subtitle.ID,
		&subtitle.ProjectID,
		&subtitle.Format,
		&subtitle.StorageKey,
		&subtitle.CreatedAt,
	)
	if err != nil {
		return nil, nil
	}

	return &subtitle, nil
}

// Service handles subtitle generation
type Service struct {
	repo    *Repository
	storage storage.Provider
}

// NewService creates a new subtitle service
func NewService(repo *Repository, storage storage.Provider) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
	}
}

// GenerateWithSpeechMarks generates subtitles using Polly speech marks for word-accurate timing
func (s *Service) GenerateWithSpeechMarks(ctx context.Context, projectID uuid.UUID, scenes []*models.Scene, audioDurations map[uuid.UUID]int, speechMarksKeys map[uuid.UUID]string) (*models.Subtitle, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "subtitle_generation")

	logger.Info(ctx, "generating subtitles with speech marks", "scene_count", len(scenes))

	// Check if subtitles already exist (idempotency)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "subtitles already exist, returning existing")
		return existing, nil
	}

	// Generate SRT content using speech marks for accurate timing
	srtContent, err := s.generateSRTWithSpeechMarks(ctx, scenes, audioDurations, speechMarksKeys)
	if err != nil {
		// Fallback to duration-based if speech marks fail
		logger.Warn(ctx, "speech marks failed, falling back to duration-based", "error", err)
		srtContent = generateSRTWithDurations(scenes, audioDurations)
	}

	// Store subtitles
	storageKey := fmt.Sprintf("projects/%s/subtitles/subtitles.srt", projectID)
	if err := s.storage.Put(ctx, storageKey, "text/srt", []byte(srtContent)); err != nil {
		return nil, fmt.Errorf("failed to store subtitles: %w", err)
	}

	// Create subtitle record
	subtitle := &models.Subtitle{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Format:     "srt",
		StorageKey: storageKey,
	}

	if err := s.repo.Create(ctx, subtitle); err != nil {
		return nil, fmt.Errorf("failed to save subtitle: %w", err)
	}

	logger.Info(ctx, "subtitles generated successfully with speech marks", "storage_key", storageKey)

	return subtitle, nil
}

// generateSRTWithSpeechMarks creates SRT content using word-level speech marks
func (s *Service) generateSRTWithSpeechMarks(ctx context.Context, scenes []*models.Scene, audioDurations map[uuid.UUID]int, speechMarksKeys map[uuid.UUID]string) (string, error) {
	var builder strings.Builder
	var currentTimeMs int = 0
	subtitleIndex := 1

	for _, scene := range scenes {
		speechMarksKey, hasSpeechMarks := speechMarksKeys[scene.ID]

		if !hasSpeechMarks || speechMarksKey == "" {
			// No speech marks - use duration-based timing for this scene
			durationMs := int(scene.DurationSec * 1000)
			if actualDuration, ok := audioDurations[scene.ID]; ok && actualDuration > 0 {
				durationMs = actualDuration
			}

			builder.WriteString(fmt.Sprintf("%d\n", subtitleIndex))
			builder.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTimeMs(currentTimeMs), formatSRTTimeMs(currentTimeMs+durationMs)))
			builder.WriteString(scene.NarrationText + "\n\n")
			subtitleIndex++
			currentTimeMs += durationMs
			continue
		}

		// Fetch speech marks from storage
		speechMarksData, err := s.storage.Get(ctx, speechMarksKey)
		if err != nil {
			logger.Warn(ctx, "failed to fetch speech marks, using duration", "scene_id", scene.ID, "error", err)
			durationMs := int(scene.DurationSec * 1000)
			if actualDuration, ok := audioDurations[scene.ID]; ok && actualDuration > 0 {
				durationMs = actualDuration
			}
			builder.WriteString(fmt.Sprintf("%d\n", subtitleIndex))
			builder.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTimeMs(currentTimeMs), formatSRTTimeMs(currentTimeMs+durationMs)))
			builder.WriteString(scene.NarrationText + "\n\n")
			subtitleIndex++
			currentTimeMs += durationMs
			continue
		}

		// Parse speech marks
		var speechMarks []SpeechMark
		if err := json.Unmarshal(speechMarksData, &speechMarks); err != nil {
			logger.Warn(ctx, "failed to parse speech marks", "scene_id", scene.ID, "error", err)
			continue
		}

		// Filter to only word marks
		var wordMarks []SpeechMark
		for _, mark := range speechMarks {
			if mark.Type == "word" {
				wordMarks = append(wordMarks, mark)
			}
		}

		if len(wordMarks) == 0 {
			// No word marks - use full scene duration
			durationMs := int(scene.DurationSec * 1000)
			if actualDuration, ok := audioDurations[scene.ID]; ok && actualDuration > 0 {
				durationMs = actualDuration
			}
			builder.WriteString(fmt.Sprintf("%d\n", subtitleIndex))
			builder.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTimeMs(currentTimeMs), formatSRTTimeMs(currentTimeMs+durationMs)))
			builder.WriteString(scene.NarrationText + "\n\n")
			subtitleIndex++
			currentTimeMs += durationMs
			continue
		}

		// Group words into subtitle chunks (3-6 words per chunk for readability)
		chunks := groupWordsIntoChunks(wordMarks, 5)

		for _, chunk := range chunks {
			if len(chunk) == 0 {
				continue
			}

			// Start time is the first word's time
			startMs := currentTimeMs + chunk[0].Time

			// End time is the last word's time + estimated word duration
			lastWord := chunk[len(chunk)-1]
			// Estimate word duration: ~300ms per word or until next chunk
			endMs := currentTimeMs + lastWord.Time + 400

			// Build the text from words
			var words []string
			for _, mark := range chunk {
				words = append(words, mark.Value)
			}
			text := strings.Join(words, " ")

			builder.WriteString(fmt.Sprintf("%d\n", subtitleIndex))
			builder.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTimeMs(startMs), formatSRTTimeMs(endMs)))
			builder.WriteString(text + "\n\n")
			subtitleIndex++
		}

		// Update current time based on actual audio duration
		durationMs := int(scene.DurationSec * 1000)
		if actualDuration, ok := audioDurations[scene.ID]; ok && actualDuration > 0 {
			durationMs = actualDuration
		}
		currentTimeMs += durationMs
	}

	return builder.String(), nil
}

// groupWordsIntoChunks groups word marks into readable subtitle chunks
func groupWordsIntoChunks(wordMarks []SpeechMark, maxWordsPerChunk int) [][]SpeechMark {
	var chunks [][]SpeechMark
	var currentChunk []SpeechMark

	for _, mark := range wordMarks {
		currentChunk = append(currentChunk, mark)

		// Check if we should end this chunk
		shouldEnd := false

		// End on punctuation
		if strings.ContainsAny(mark.Value, ".,!?;:") {
			shouldEnd = true
		}

		// End if we've reached max words
		if len(currentChunk) >= maxWordsPerChunk {
			shouldEnd = true
		}

		if shouldEnd && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = nil
		}
	}

	// Add remaining words
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// SceneTiming holds timing info for a scene
type SceneTiming struct {
	SceneID     uuid.UUID
	SceneNumber int
	Text        string
	DurationMs  int // Actual audio duration in milliseconds
}

// Generate generates subtitles for a project
func (s *Service) Generate(ctx context.Context, projectID uuid.UUID, scenes []*models.Scene) (*models.Subtitle, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "subtitle_generation")

	logger.Info(ctx, "generating subtitles", "scene_count", len(scenes))

	// Check if subtitles already exist (idempotency)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "subtitles already exist, returning existing")
		return existing, nil
	}

	// Generate SRT content
	srtContent := generateSRT(scenes)

	logger.Info(ctx, "generated SRT content", "length", len(srtContent))

	// Store subtitles
	storageKey := fmt.Sprintf("projects/%s/subtitles/subtitles.srt", projectID)
	if err := s.storage.Put(ctx, storageKey, "text/srt", []byte(srtContent)); err != nil {
		return nil, fmt.Errorf("failed to store subtitles: %w", err)
	}

	// Create subtitle record
	subtitle := &models.Subtitle{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Format:     "srt",
		StorageKey: storageKey,
	}

	if err := s.repo.Create(ctx, subtitle); err != nil {
		return nil, fmt.Errorf("failed to save subtitle: %w", err)
	}

	logger.Info(ctx, "subtitles generated successfully", "storage_key", storageKey)

	return subtitle, nil
}

// GetByProjectID retrieves subtitles for a project
func (s *Service) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Subtitle, error) {
	return s.repo.GetByProjectID(ctx, projectID)
}

// GenerateWithAudioDurations generates subtitles using actual audio durations
func (s *Service) GenerateWithAudioDurations(ctx context.Context, projectID uuid.UUID, scenes []*models.Scene, audioDurations map[uuid.UUID]int) (*models.Subtitle, error) {
	ctx = logger.WithProjectID(ctx, projectID.String())
	ctx = logger.WithStage(ctx, "subtitle_generation")

	logger.Info(ctx, "generating subtitles with audio durations", "scene_count", len(scenes))

	// Check if subtitles already exist (idempotency)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		logger.Info(ctx, "subtitles already exist, returning existing")
		return existing, nil
	}

	// Generate SRT content using actual audio durations
	srtContent := generateSRTWithDurations(scenes, audioDurations)

	// Store subtitles
	storageKey := fmt.Sprintf("projects/%s/subtitles/subtitles.srt", projectID)
	if err := s.storage.Put(ctx, storageKey, "text/srt", []byte(srtContent)); err != nil {
		return nil, fmt.Errorf("failed to store subtitles: %w", err)
	}

	// Create subtitle record
	subtitle := &models.Subtitle{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Format:     "srt",
		StorageKey: storageKey,
	}

	if err := s.repo.Create(ctx, subtitle); err != nil {
		return nil, fmt.Errorf("failed to save subtitle: %w", err)
	}

	logger.Info(ctx, "subtitles generated successfully", "storage_key", storageKey)

	return subtitle, nil
}

func generateSRT(scenes []*models.Scene) string {
	var builder strings.Builder
	var currentTime float64 = 0

	for i, scene := range scenes {
		startTime := currentTime
		endTime := currentTime + scene.DurationSec

		// Split long text into subtitle chunks
		chunks := splitIntoSubtitleChunks(scene.NarrationText, scene.DurationSec)
		chunkDuration := scene.DurationSec / float64(len(chunks))

		for j, chunk := range chunks {
			chunkStart := startTime + float64(j)*chunkDuration
			chunkEnd := chunkStart + chunkDuration

			builder.WriteString(fmt.Sprintf("%d\n", i*10+j+1))
			builder.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTime(chunkStart), formatSRTTime(chunkEnd)))
			builder.WriteString(chunk + "\n\n")
		}

		currentTime = endTime
	}

	return builder.String()
}

func generateSRTWithDurations(scenes []*models.Scene, audioDurations map[uuid.UUID]int) string {
	var builder strings.Builder
	var currentTimeMs int = 0

	for i, scene := range scenes {
		// Get actual duration from audio, fallback to scene duration
		durationMs := int(scene.DurationSec * 1000)
		if actualDuration, ok := audioDurations[scene.ID]; ok && actualDuration > 0 {
			durationMs = actualDuration
		}

		startTimeMs := currentTimeMs
		endTimeMs := startTimeMs + durationMs

		// One subtitle per scene - matches the audio exactly
		builder.WriteString(fmt.Sprintf("%d\n", i+1))
		builder.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTimeMs(startTimeMs), formatSRTTimeMs(endTimeMs)))
		builder.WriteString(scene.NarrationText + "\n\n")

		currentTimeMs = endTimeMs
	}

	return builder.String()
}

func formatSRTTimeMs(ms int) string {
	hours := ms / 3600000
	ms = ms % 3600000
	minutes := ms / 60000
	ms = ms % 60000
	seconds := ms / 1000
	millis := ms % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, millis)
}

func splitIntoSubtitleChunks(text string, duration float64) []string {
	// Split text at natural breaks (sentences, commas) for better sync with speech
	// This creates chunks that align better with spoken phrases

	if strings.TrimSpace(text) == "" {
		return []string{text}
	}

	var chunks []string

	// First, split by sentences (period, exclamation, question mark)
	sentences := splitBySentences(text)

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// If sentence is short enough, use it as-is
		if len(sentence) <= 50 {
			chunks = append(chunks, sentence)
			continue
		}

		// Split longer sentences by commas or natural phrase breaks
		phrases := splitByPhrases(sentence)
		for _, phrase := range phrases {
			phrase = strings.TrimSpace(phrase)
			if phrase != "" {
				chunks = append(chunks, phrase)
			}
		}
	}

	// If we got no chunks, fall back to simple word splitting
	if len(chunks) == 0 {
		words := strings.Fields(text)
		var current []string
		for _, word := range words {
			current = append(current, word)
			if len(current) >= 6 {
				chunks = append(chunks, strings.Join(current, " "))
				current = nil
			}
		}
		if len(current) > 0 {
			chunks = append(chunks, strings.Join(current, " "))
		}
	}

	return chunks
}

func splitBySentences(text string) []string {
	var sentences []string
	var current strings.Builder

	for i, r := range text {
		current.WriteRune(r)

		// Check for sentence endings
		if r == '.' || r == '!' || r == '?' {
			// Make sure it's not an abbreviation (check if next char is space or end)
			if i+1 >= len(text) || text[i+1] == ' ' {
				sentences = append(sentences, current.String())
				current.Reset()
			}
		}
	}

	// Add any remaining text
	if current.Len() > 0 {
		sentences = append(sentences, current.String())
	}

	return sentences
}

func splitByPhrases(sentence string) []string {
	var phrases []string
	var current strings.Builder

	for _, r := range sentence {
		current.WriteRune(r)

		// Split on commas, semicolons, or dashes for natural phrase breaks
		if r == ',' || r == ';' || r == '—' || r == '-' {
			if current.Len() > 10 { // Only split if we have reasonable content
				phrases = append(phrases, current.String())
				current.Reset()
			}
		}

		// Also split if getting too long (max 50 chars per subtitle)
		if current.Len() >= 50 {
			// Try to split at last space
			text := current.String()
			lastSpace := strings.LastIndex(text, " ")
			if lastSpace > 20 {
				phrases = append(phrases, text[:lastSpace])
				current.Reset()
				current.WriteString(text[lastSpace+1:])
			}
		}
	}

	// Add remaining text
	if current.Len() > 0 {
		phrases = append(phrases, current.String())
	}

	return phrases
}

func formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}
