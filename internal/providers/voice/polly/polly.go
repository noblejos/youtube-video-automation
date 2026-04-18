package polly

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/internal/voice"
	"github.com/gama/youtube-video-automation/pkg/contracts"
)

// SpeechMark represents a single speech mark from Polly
type SpeechMark struct {
	Time  int    `json:"time"`  // Time in milliseconds
	Type  string `json:"type"`  // "word", "sentence", "ssml", or "viseme"
	Start int    `json:"start"` // Start character offset
	End   int    `json:"end"`   // End character offset
	Value string `json:"value"` // The word or sentence text
}

// Config holds AWS Polly configuration
type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	DefaultVoice    string
	Engine          string
}

// Provider implements voice.Provider using AWS Polly
type Provider struct {
	client  *polly.Client
	storage storage.Provider
	config  Config
}

// New creates a new AWS Polly provider
func New(ctx context.Context, cfg Config, store storage.Provider) (voice.Provider, error) {
	var awsCfg aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// Use explicit credentials
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		// Use default credential chain (IAM role, env vars, etc.)
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := polly.NewFromConfig(awsCfg)

	// Set defaults
	if cfg.DefaultVoice == "" {
		cfg.DefaultVoice = "Ayanda"
	}
	if cfg.Engine == "" {
		cfg.Engine = "standard"
	}

	return &Provider{
		client:  client,
		storage: store,
		config:  cfg,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "polly"
}

// sanitizeSSMLForNeural removes SSML tags not supported by neural voices
func sanitizeSSMLForNeural(ssml string) string {
	// Neural voices don't support: <emphasis>, <prosody>, <amazon:effect>
	// They do support: <speak>, <break>, <lang>, <mark>, <p>, <s>, <phoneme>, <say-as>, <sub>

	// Remove <emphasis> tags but keep content
	emphasisRegex := regexp.MustCompile(`<emphasis[^>]*>|</emphasis>`)
	ssml = emphasisRegex.ReplaceAllString(ssml, "")

	// Remove <prosody> tags but keep content
	prosodyRegex := regexp.MustCompile(`<prosody[^>]*>|</prosody>`)
	ssml = prosodyRegex.ReplaceAllString(ssml, "")

	// Remove <amazon:effect> tags but keep content
	effectRegex := regexp.MustCompile(`<amazon:effect[^>]*>|</amazon:effect>`)
	ssml = effectRegex.ReplaceAllString(ssml, "")

	// Clean up any double spaces
	ssml = strings.Join(strings.Fields(ssml), " ")

	return ssml
}

// Generate generates audio from text using AWS Polly
func (p *Provider) Generate(ctx context.Context, req contracts.VoiceRequest) (*contracts.VoiceResult, error) {
	// Determine voice ID
	voiceID := req.VoiceID
	if voiceID == "" {
		voiceID = p.config.DefaultVoice
	}

	// Determine engine
	engine := types.Engine(p.config.Engine)

	// Determine text type
	textType := types.TextTypeText
	text := req.Text
	if req.TextType == "ssml" {
		textType = types.TextTypeSsml
		// Sanitize SSML for neural engine (remove unsupported tags)
		if engine == types.EngineNeural {
			text = sanitizeSSMLForNeural(text)
		}
	}

	// Call Polly for audio
	audioInput := &polly.SynthesizeSpeechInput{
		OutputFormat: types.OutputFormatMp3,
		Text:         aws.String(text),
		VoiceId:      types.VoiceId(voiceID),
		Engine:       engine,
		TextType:     textType,
	}

	audioOutput, err := p.client.SynthesizeSpeech(ctx, audioInput)
	if err != nil {
		return nil, fmt.Errorf("polly synthesis failed: %w", err)
	}
	defer audioOutput.AudioStream.Close()

	// Read audio data
	audioData, err := io.ReadAll(audioOutput.AudioStream)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio stream: %w", err)
	}

	// Call Polly for speech marks (word-level timestamps)
	speechMarks, err := p.getSpeechMarks(ctx, text, voiceID, engine, textType)
	if err != nil {
		// Log but don't fail - speech marks are optional for subtitle sync
		fmt.Printf("warning: failed to get speech marks: %v\n", err)
		speechMarks = nil
	}

	// Generate storage keys
	storageKey := fmt.Sprintf("projects/%s/audio/scene_%s.mp3", req.ProjectID, req.SceneID)
	speechMarksKey := fmt.Sprintf("projects/%s/audio/scene_%s_speechmarks.json", req.ProjectID, req.SceneID)

	// Store the audio
	if err := p.storage.Put(ctx, storageKey, "audio/mpeg", audioData); err != nil {
		return nil, fmt.Errorf("failed to store audio: %w", err)
	}

	// Store speech marks if available
	if speechMarks != nil && len(speechMarks) > 0 {
		speechMarksJSON, _ := json.Marshal(speechMarks)
		if err := p.storage.Put(ctx, speechMarksKey, "application/json", speechMarksJSON); err != nil {
			// Log but don't fail
			fmt.Printf("warning: failed to store speech marks: %v\n", err)
		}
	}

	// Calculate duration from speech marks (most accurate) or estimate from file size
	var durationMs int
	if len(speechMarks) > 0 {
		// Get the time of the last speech mark
		lastMark := speechMarks[len(speechMarks)-1]
		durationMs = lastMark.Time + 500 // Add 500ms buffer after last word
	} else {
		// Fallback: estimate from file size (MP3 at ~128kbps)
		durationMs = int(float64(len(audioData)) / 16000.0 * 1000.0)
	}
	if durationMs < 1000 {
		durationMs = 1000
	}

	return &contracts.VoiceResult{
		StorageKey:     storageKey,
		DurationMs:     durationMs,
		SpeechMarksKey: speechMarksKey,
	}, nil
}

// getSpeechMarks fetches word-level timestamps from Polly
func (p *Provider) getSpeechMarks(ctx context.Context, text string, voiceID string, engine types.Engine, textType types.TextType) ([]SpeechMark, error) {
	input := &polly.SynthesizeSpeechInput{
		OutputFormat:    types.OutputFormatJson,
		Text:            aws.String(text),
		VoiceId:         types.VoiceId(voiceID),
		Engine:          engine,
		TextType:        textType,
		SpeechMarkTypes: []types.SpeechMarkType{types.SpeechMarkTypeWord, types.SpeechMarkTypeSentence},
	}

	output, err := p.client.SynthesizeSpeech(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("polly speech marks failed: %w", err)
	}
	defer output.AudioStream.Close()

	// Parse the NDJSON response (one JSON object per line)
	var speechMarks []SpeechMark
	scanner := bufio.NewScanner(output.AudioStream)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var mark SpeechMark
		if err := json.Unmarshal([]byte(line), &mark); err != nil {
			continue // Skip malformed lines
		}
		speechMarks = append(speechMarks, mark)
	}

	return speechMarks, nil
}
