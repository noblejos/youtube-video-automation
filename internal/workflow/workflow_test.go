package workflow

import (
	"context"
	"testing"

	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// MockProjectRepository is a mock implementation of project repository
type MockProjectRepository struct {
	projects map[string]*models.Project
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[string]*models.Project),
	}
}

func (r *MockProjectRepository) Create(ctx context.Context, project *models.Project) error {
	r.projects[project.ID.String()] = project
	return nil
}

func (r *MockProjectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	if p, ok := r.projects[id]; ok {
		return p, nil
	}
	return nil, nil
}

func TestProjectStatusTransitions(t *testing.T) {
	tests := []struct {
		name         string
		currentState string
		nextState    string
		valid        bool
	}{
		{"created to script generating", models.ProjectStatusCreated, models.ProjectStatusScriptGenerating, true},
		{"script generating to script ready", models.ProjectStatusScriptGenerating, models.ProjectStatusScriptReady, true},
		{"script ready to scenes generating", models.ProjectStatusScriptReady, models.ProjectStatusScenesGenerating, true},
		{"scenes generating to scenes ready", models.ProjectStatusScenesGenerating, models.ProjectStatusScenesReady, true},
		{"scenes ready to voice generating", models.ProjectStatusScenesReady, models.ProjectStatusVoiceGenerating, true},
		{"voice ready to assets generating", models.ProjectStatusVoiceReady, models.ProjectStatusAssetsGenerating, true},
		{"assets ready to subtitles generating", models.ProjectStatusAssetsReady, models.ProjectStatusSubtitlesGenerating, true},
		{"subtitles ready to rendering", models.ProjectStatusSubtitlesReady, models.ProjectStatusRendering, true},
		{"render ready to review packaged", models.ProjectStatusRenderReady, models.ProjectStatusReviewPackaged, true},
		{"in review to approved", models.ProjectStatusInReview, models.ProjectStatusApproved, true},
		{"in review to rejected", models.ProjectStatusInReview, models.ProjectStatusRejected, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify state transitions are valid
			if !isValidTransition(tt.currentState, tt.nextState) && tt.valid {
				t.Errorf("expected transition from %s to %s to be valid", tt.currentState, tt.nextState)
			}
		})
	}
}

func isValidTransition(from, to string) bool {
	validTransitions := map[string][]string{
		models.ProjectStatusCreated:             {models.ProjectStatusScriptGenerating, models.ProjectStatusFailed},
		models.ProjectStatusScriptGenerating:    {models.ProjectStatusScriptReady, models.ProjectStatusFailed},
		models.ProjectStatusScriptReady:         {models.ProjectStatusScenesGenerating, models.ProjectStatusFailed},
		models.ProjectStatusScenesGenerating:    {models.ProjectStatusScenesReady, models.ProjectStatusFailed},
		models.ProjectStatusScenesReady:         {models.ProjectStatusVoiceGenerating, models.ProjectStatusFailed},
		models.ProjectStatusVoiceGenerating:     {models.ProjectStatusVoiceReady, models.ProjectStatusAssetsGenerating, models.ProjectStatusFailed},
		models.ProjectStatusVoiceReady:          {models.ProjectStatusAssetsGenerating, models.ProjectStatusFailed},
		models.ProjectStatusAssetsGenerating:    {models.ProjectStatusAssetsReady, models.ProjectStatusFailed},
		models.ProjectStatusAssetsReady:         {models.ProjectStatusSubtitlesGenerating, models.ProjectStatusFailed},
		models.ProjectStatusSubtitlesGenerating: {models.ProjectStatusSubtitlesReady, models.ProjectStatusFailed},
		models.ProjectStatusSubtitlesReady:      {models.ProjectStatusRendering, models.ProjectStatusFailed},
		models.ProjectStatusRendering:           {models.ProjectStatusRenderReady, models.ProjectStatusFailed},
		models.ProjectStatusRenderReady:         {models.ProjectStatusReviewPackaged, models.ProjectStatusFailed},
		models.ProjectStatusReviewPackaged:      {models.ProjectStatusInReview, models.ProjectStatusFailed},
		models.ProjectStatusInReview:            {models.ProjectStatusApproved, models.ProjectStatusRejected},
		models.ProjectStatusApproved:            {models.ProjectStatusPublishing},
		models.ProjectStatusPublishing:          {models.ProjectStatusPublished, models.ProjectStatusFailed},
	}

	validStates, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, valid := range validStates {
		if valid == to {
			return true
		}
	}

	return false
}

func TestCreateProjectRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     contracts.CreateProjectRequest
		wantErr bool
	}{
		{
			name: "valid request with all fields",
			req: contracts.CreateProjectRequest{
				Topic:             "The Rise and Fall of Mansa Musa",
				ChannelStyle:      "dramatic_history_shorts",
				TargetDurationSec: 120,
				AspectRatio:       "9:16",
			},
			wantErr: false,
		},
		{
			name: "valid request with minimal fields",
			req: contracts.CreateProjectRequest{
				Topic: "Ancient Egypt",
			},
			wantErr: false,
		},
		{
			name:    "invalid request - empty topic",
			req:     contracts.CreateProjectRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateProjectRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateProjectRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validateCreateProjectRequest(req contracts.CreateProjectRequest) error {
	if req.Topic == "" {
		return &ValidationError{Field: "topic", Message: "topic is required"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
