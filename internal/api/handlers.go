package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/internal/workflow"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
)

// Handlers holds the HTTP handlers
type Handlers struct {
	workflow *workflow.Service
	storage  storage.Provider
}

// NewHandlers creates new HTTP handlers
func NewHandlers(workflow *workflow.Service, storage storage.Provider) *Handlers {
	return &Handlers{
		workflow: workflow,
		storage:  storage,
	}
}

// CreateProject handles POST /projects
func (h *Handlers) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req contracts.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Require either topic or script
	if req.Topic == "" && req.Script == "" {
		h.writeError(w, http.StatusBadRequest, "either topic or script is required", "")
		return
	}

	// If only script is provided, use title as topic for display purposes
	if req.Topic == "" && req.Script != "" {
		if req.Title != "" {
			req.Topic = req.Title
		} else {
			req.Topic = "Custom Script"
		}
	}

	project, err := h.workflow.CreateProject(ctx, &req)
	if err != nil {
		logger.Error(ctx, "failed to create project", "error", err)
		h.writeError(w, http.StatusInternalServerError, "failed to create project", err.Error())
		return
	}

	response := contracts.ProjectResponse{
		ProjectID:  project.ID,
		ExternalID: project.ExternalID,
		Status:     project.Status,
		Topic:      project.Topic,
		CreatedAt:  project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if project.Title != nil {
		response.Title = *project.Title
	}
	if project.CurrentStep != nil {
		response.CurrentStep = *project.CurrentStep
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// GetProject handles GET /projects/{id}
func (h *Handlers) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid project ID", err.Error())
		return
	}

	manifest, err := h.workflow.GetProjectManifest(ctx, id)
	if err != nil {
		logger.Error(ctx, "failed to get project", "error", err)
		h.writeError(w, http.StatusNotFound, "project not found", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, manifest.Project)
}

// GetProjectManifest handles GET /projects/{id}/manifest
func (h *Handlers) GetProjectManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid project ID", err.Error())
		return
	}

	manifest, err := h.workflow.GetProjectManifest(ctx, id)
	if err != nil {
		logger.Error(ctx, "failed to get project manifest", "error", err)
		h.writeError(w, http.StatusNotFound, "project not found", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, manifest)
}

// DownloadVideo handles GET /projects/{id}/download
func (h *Handlers) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid project ID", err.Error())
		return
	}

	manifest, err := h.workflow.GetProjectManifest(ctx, id)
	if err != nil {
		logger.Error(ctx, "failed to get project", "error", err)
		h.writeError(w, http.StatusNotFound, "project not found", err.Error())
		return
	}

	// Check if project has a render
	if manifest.Render == nil || manifest.Render.DraftVideoKey == "" {
		h.writeError(w, http.StatusNotFound, "video not ready", "project has not been rendered yet")
		return
	}

	// Get the video file from storage
	videoData, err := h.storage.Get(ctx, manifest.Render.DraftVideoKey)
	if err != nil {
		logger.Error(ctx, "failed to get video from storage", "error", err, "key", manifest.Render.DraftVideoKey)
		h.writeError(w, http.StatusInternalServerError, "failed to retrieve video", err.Error())
		return
	}

	// Set headers for file download - use title if available, fallback to external ID
	baseName := manifest.Project.ExternalID
	if manifest.Project.Title != "" {
		baseName = sanitizeFilename(manifest.Project.Title)
	}
	filename := fmt.Sprintf("%s.mp4", baseName)
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(videoData)))
	w.WriteHeader(http.StatusOK)
	w.Write(videoData)
}

// ApproveProject handles POST /projects/{id}/approve
func (h *Handlers) ApproveProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid project ID", err.Error())
		return
	}

	var req contracts.ApproveRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.workflow.ApproveProject(ctx, id, req.Notes, req.ActedBy); err != nil {
		logger.Error(ctx, "failed to approve project", "error", err)
		h.writeError(w, http.StatusBadRequest, "failed to approve project", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

// RejectProject handles POST /projects/{id}/reject
func (h *Handlers) RejectProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid project ID", err.Error())
		return
	}

	var req contracts.RejectRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.workflow.RejectProject(ctx, id, req.Notes, req.ActedBy); err != nil {
		logger.Error(ctx, "failed to reject project", "error", err)
		h.writeError(w, http.StatusBadRequest, "failed to reject project", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// RetryProject handles POST /projects/{id}/retry
func (h *Handlers) RetryProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid project ID", err.Error())
		return
	}

	if err := h.workflow.RetryProject(ctx, id); err != nil {
		logger.Error(ctx, "failed to retry project", "error", err)
		h.writeError(w, http.StatusBadRequest, "failed to retry project", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "retrying"})
}

// HealthCheck handles GET /health
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handlers) writeError(w http.ResponseWriter, status int, message, details string) {
	h.writeJSON(w, status, contracts.ErrorResponse{
		Error:   message,
		Details: details,
	})
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Replace characters that are problematic in filenames
	// Windows: \ / : * ? " < > |
	// Also replace spaces with underscores for cleaner filenames
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)
	name = invalidChars.ReplaceAllString(name, "")

	// Replace multiple spaces/underscores with single underscore
	name = regexp.MustCompile(`[\s_]+`).ReplaceAllString(name, "_")

	// Trim leading/trailing underscores and spaces
	name = strings.Trim(name, "_ ")

	// Limit length to 200 characters (leaving room for extension)
	if len(name) > 200 {
		name = name[:200]
	}

	// If empty after sanitization, use a default
	if name == "" {
		name = "video"
	}

	return name
}
