package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gama/youtube-video-automation/pkg/contracts"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

// Handler is a function that processes a job
type Handler func(ctx context.Context, job *models.Job) error

// Queue defines the job queue interface
type Queue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *models.Job) error

	// Dequeue gets the next job from the queue
	Dequeue(ctx context.Context) (*models.Job, error)

	// Complete marks a job as completed
	Complete(ctx context.Context, jobID uuid.UUID, result json.RawMessage) error

	// Fail marks a job as failed
	Fail(ctx context.Context, jobID uuid.UUID, err error) error

	// RegisterHandler registers a handler for a job type
	RegisterHandler(jobType string, handler Handler)

	// SetOnJobFailedPermanently sets callback for permanent failures
	SetOnJobFailedPermanently(callback func(ctx context.Context, job *models.Job, err error))

	// Start starts processing jobs
	Start(ctx context.Context) error

	// Stop stops processing jobs
	Stop() error
}

// MemoryQueue is an in-memory job queue for development/testing
type MemoryQueue struct {
	mu       sync.Mutex
	jobs     []*models.Job
	handlers map[string]Handler
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup

	// Callback for job state changes (for persistence)
	onJobStateChange func(ctx context.Context, job *models.Job) error
	// Callback for permanent job failures (after all retries exhausted)
	onJobFailedPermanently func(ctx context.Context, job *models.Job, err error)
}

// NewMemoryQueue creates a new in-memory job queue
func NewMemoryQueue(onJobStateChange func(ctx context.Context, job *models.Job) error) *MemoryQueue {
	return &MemoryQueue{
		jobs:             make([]*models.Job, 0),
		handlers:         make(map[string]Handler),
		stopCh:           make(chan struct{}),
		onJobStateChange: onJobStateChange,
	}
}

// SetOnJobFailedPermanently sets the callback for permanent job failures
func (q *MemoryQueue) SetOnJobFailedPermanently(callback func(ctx context.Context, job *models.Job, err error)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.onJobFailedPermanently = callback
}

// Enqueue adds a job to the queue
func (q *MemoryQueue) Enqueue(ctx context.Context, job *models.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	job.Status = models.JobStatusQueued
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	q.jobs = append(q.jobs, job)

	if q.onJobStateChange != nil {
		if err := q.onJobStateChange(ctx, job); err != nil {
			logger.Error(ctx, "failed to persist job state", "error", err)
		}
	}

	logger.Info(ctx, "job enqueued",
		"job_id", job.ID,
		"job_type", job.JobType,
		"project_id", job.ProjectID,
	)

	return nil
}

// Dequeue gets the next job from the queue
func (q *MemoryQueue) Dequeue(ctx context.Context) (*models.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, job := range q.jobs {
		if job.Status == models.JobStatusQueued {
			job.Status = models.JobStatusRunning
			job.AttemptCount++
			job.UpdatedAt = time.Now()
			q.jobs[i] = job

			if q.onJobStateChange != nil {
				if err := q.onJobStateChange(ctx, job); err != nil {
					logger.Error(ctx, "failed to persist job state", "error", err)
				}
			}

			return job, nil
		}
	}

	return nil, nil
}

// Complete marks a job as completed
func (q *MemoryQueue) Complete(ctx context.Context, jobID uuid.UUID, result json.RawMessage) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, job := range q.jobs {
		if job.ID == jobID {
			job.Status = models.JobStatusSucceeded
			job.Result = result
			job.UpdatedAt = time.Now()
			q.jobs[i] = job

			if q.onJobStateChange != nil {
				if err := q.onJobStateChange(ctx, job); err != nil {
					logger.Error(ctx, "failed to persist job state", "error", err)
				}
			}

			logger.Info(ctx, "job completed",
				"job_id", jobID,
				"job_type", job.JobType,
			)
			return nil
		}
	}

	return fmt.Errorf("job not found: %s", jobID)
}

// Fail marks a job as failed
func (q *MemoryQueue) Fail(ctx context.Context, jobID uuid.UUID, jobErr error) error {
	q.mu.Lock()

	var foundJob *models.Job
	var isPermanent bool

	for i, job := range q.jobs {
		if job.ID == jobID {
			errMsg := jobErr.Error()
			job.ErrorMessage = &errMsg
			job.UpdatedAt = time.Now()

			// Check if we should retry
			if job.AttemptCount < job.MaxAttempts {
				job.Status = models.JobStatusQueued // Re-queue for retry
				logger.Warn(ctx, "job failed, will retry",
					"job_id", jobID,
					"attempt", job.AttemptCount,
					"max_attempts", job.MaxAttempts,
					"error", jobErr,
				)
			} else {
				job.Status = models.JobStatusFailed
				isPermanent = true
				logger.Error(ctx, "job failed permanently",
					"job_id", jobID,
					"attempts", job.AttemptCount,
					"error", jobErr,
				)
			}

			q.jobs[i] = job
			foundJob = job

			if q.onJobStateChange != nil {
				if err := q.onJobStateChange(ctx, job); err != nil {
					logger.Error(ctx, "failed to persist job state", "error", err)
				}
			}

			break
		}
	}

	// Get the callback before unlocking
	callback := q.onJobFailedPermanently
	q.mu.Unlock()

	if foundJob == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Call the permanent failure callback outside the lock
	if isPermanent && callback != nil {
		callback(ctx, foundJob, jobErr)
	}

	return nil
}

// RegisterHandler registers a handler for a job type
func (q *MemoryQueue) RegisterHandler(jobType string, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = handler
}

// Start starts processing jobs
func (q *MemoryQueue) Start(ctx context.Context) error {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return fmt.Errorf("queue already running")
	}
	q.running = true
	q.mu.Unlock()

	q.wg.Add(1)
	go q.processLoop(ctx)

	return nil
}

// Stop stops processing jobs
func (q *MemoryQueue) Stop() error {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return nil
	}
	q.running = false
	q.mu.Unlock()

	close(q.stopCh)
	q.wg.Wait()

	return nil
}

func (q *MemoryQueue) processLoop(ctx context.Context) {
	defer q.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			job, err := q.Dequeue(ctx)
			if err != nil {
				logger.Error(ctx, "failed to dequeue job", "error", err)
				continue
			}
			if job == nil {
				continue
			}

			q.processJob(ctx, job)
		}
	}
}

func (q *MemoryQueue) processJob(ctx context.Context, job *models.Job) {
	q.mu.Lock()
	handler, ok := q.handlers[job.JobType]
	q.mu.Unlock()

	if !ok {
		logger.Error(ctx, "no handler registered for job type", "job_type", job.JobType)
		q.Fail(ctx, job.ID, fmt.Errorf("no handler for job type: %s", job.JobType))
		return
	}

	jobCtx := logger.WithProjectID(ctx, job.ProjectID.String())
	jobCtx = logger.WithJobID(jobCtx, job.ID.String())
	jobCtx = logger.WithStage(jobCtx, job.JobType)

	logger.Info(jobCtx, "processing job")

	if err := handler(jobCtx, job); err != nil {
		q.Fail(ctx, job.ID, err)
		return
	}

	// Marshal empty result if none provided
	result, _ := json.Marshal(map[string]string{"status": "completed"})
	q.Complete(ctx, job.ID, result)
}

// CreateJob creates a new job with default values
func CreateJob(projectID uuid.UUID, jobType string, payload interface{}) (*models.Job, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return &models.Job{
		ID:          uuid.New(),
		ProjectID:   projectID,
		JobType:     jobType,
		Status:      models.JobStatusQueued,
		AttemptCount: 0,
		MaxAttempts: 3,
		Payload:     payloadBytes,
	}, nil
}

// ParsePayload parses job payload into the given struct
func ParsePayload[T any](job *models.Job) (*T, error) {
	var payload T
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return &payload, nil
}

// ScriptJobPayload is the payload for script generation jobs
type ScriptJobPayload struct {
	contracts.JobPayload
	Topic        string `json:"topic"`
	ChannelStyle string `json:"channel_style"`
	TargetDuration int  `json:"target_duration"`
}

// SceneJobPayload is the payload for scene generation jobs
type SceneJobPayload struct {
	contracts.JobPayload
	ScriptID uuid.UUID `json:"script_id"`
}

// VoiceJobPayload is the payload for voice generation jobs
type VoiceJobPayload struct {
	contracts.JobPayload
	VoiceID  string `json:"voice_id"`
	Engine   string `json:"engine"`
}

// ImageJobPayload is the payload for image generation jobs
type ImageJobPayload struct {
	contracts.JobPayload
	AspectRatio  string `json:"aspect_ratio"`
	StyleProfile string `json:"style_profile"`
}
