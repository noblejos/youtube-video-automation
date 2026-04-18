package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/gama/youtube-video-automation/pkg/logger"
	"github.com/gama/youtube-video-automation/pkg/models"
)

const (
	jobQueueKey     = "jobs:queue"
	jobDataPrefix   = "jobs:data:"
	jobLockPrefix   = "jobs:lock:"
	lockTTL         = 5 * time.Minute
	pollInterval    = 500 * time.Millisecond
)

// RedisQueue is a Redis-backed job queue for distributed processing
type RedisQueue struct {
	client   *redis.Client
	handlers map[string]Handler
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
	wg       sync.WaitGroup

	workerCount int

	// Callback for job state changes (for database persistence)
	onJobStateChange func(ctx context.Context, job *models.Job) error
	// Callback for permanent job failures
	onJobFailedPermanently func(ctx context.Context, job *models.Job, err error)
}

// RedisQueueConfig holds Redis queue configuration
type RedisQueueConfig struct {
	RedisURL    string
	WorkerCount int
}

// NewRedisQueue creates a new Redis-backed job queue
func NewRedisQueue(cfg RedisQueueConfig, onJobStateChange func(ctx context.Context, job *models.Job) error) (*RedisQueue, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	workerCount := cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = 3
	}

	return &RedisQueue{
		client:           client,
		handlers:         make(map[string]Handler),
		stopCh:           make(chan struct{}),
		workerCount:      workerCount,
		onJobStateChange: onJobStateChange,
	}, nil
}

// SetOnJobFailedPermanently sets the callback for permanent job failures
func (q *RedisQueue) SetOnJobFailedPermanently(callback func(ctx context.Context, job *models.Job, err error)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.onJobFailedPermanently = callback
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *models.Job) error {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	job.Status = models.JobStatusQueued
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job data and add to queue atomically
	pipe := q.client.Pipeline()
	pipe.Set(ctx, jobDataPrefix+job.ID.String(), jobData, 24*time.Hour)
	pipe.LPush(ctx, jobQueueKey, job.ID.String())
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Persist to database
	if q.onJobStateChange != nil {
		if err := q.onJobStateChange(ctx, job); err != nil {
			logger.Error(ctx, "failed to persist job state", "error", err)
		}
	}

	logger.Info(ctx, "job enqueued to Redis",
		"job_id", job.ID,
		"job_type", job.JobType,
		"project_id", job.ProjectID,
	)

	return nil
}

// Dequeue gets the next job from the queue
func (q *RedisQueue) Dequeue(ctx context.Context) (*models.Job, error) {
	// Block-pop from queue with timeout
	result, err := q.client.BRPop(ctx, pollInterval, jobQueueKey).Result()
	if err == redis.Nil {
		return nil, nil // No job available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue: %w", err)
	}

	jobID := result[1]

	// Get job data
	jobData, err := q.client.Get(ctx, jobDataPrefix+jobID).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	var job models.Job
	if err := json.Unmarshal(jobData, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Try to acquire lock
	lockKey := jobLockPrefix + jobID
	locked, err := q.client.SetNX(ctx, lockKey, "locked", lockTTL).Result()
	if err != nil {
		// Re-queue the job
		q.client.LPush(ctx, jobQueueKey, jobID)
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		// Job is being processed by another worker, skip
		return nil, nil
	}

	// Update job status
	job.Status = models.JobStatusRunning
	job.AttemptCount++
	job.UpdatedAt = time.Now()

	// Save updated job
	updatedData, _ := json.Marshal(&job)
	q.client.Set(ctx, jobDataPrefix+jobID, updatedData, 24*time.Hour)

	if q.onJobStateChange != nil {
		if err := q.onJobStateChange(ctx, &job); err != nil {
			logger.Error(ctx, "failed to persist job state", "error", err)
		}
	}

	return &job, nil
}

// Complete marks a job as completed
func (q *RedisQueue) Complete(ctx context.Context, jobID uuid.UUID, result json.RawMessage) error {
	jobKey := jobDataPrefix + jobID.String()
	lockKey := jobLockPrefix + jobID.String()

	// Get current job
	jobData, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	var job models.Job
	if err := json.Unmarshal(jobData, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	job.Status = models.JobStatusSucceeded
	job.Result = result
	job.UpdatedAt = time.Now()

	// Update and release lock
	updatedData, _ := json.Marshal(&job)
	pipe := q.client.Pipeline()
	pipe.Set(ctx, jobKey, updatedData, 24*time.Hour)
	pipe.Del(ctx, lockKey)
	pipe.Exec(ctx)

	if q.onJobStateChange != nil {
		if err := q.onJobStateChange(ctx, &job); err != nil {
			logger.Error(ctx, "failed to persist job state", "error", err)
		}
	}

	logger.Info(ctx, "job completed",
		"job_id", jobID,
		"job_type", job.JobType,
	)

	return nil
}

// Fail marks a job as failed
func (q *RedisQueue) Fail(ctx context.Context, jobID uuid.UUID, jobErr error) error {
	jobKey := jobDataPrefix + jobID.String()
	lockKey := jobLockPrefix + jobID.String()

	// Get current job
	jobData, err := q.client.Get(ctx, jobKey).Bytes()
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	var job models.Job
	if err := json.Unmarshal(jobData, &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	errMsg := jobErr.Error()
	job.ErrorMessage = &errMsg
	job.UpdatedAt = time.Now()

	var isPermanent bool

	// Check if we should retry
	if job.AttemptCount < job.MaxAttempts {
		job.Status = models.JobStatusQueued
		logger.Warn(ctx, "job failed, will retry",
			"job_id", jobID,
			"attempt", job.AttemptCount,
			"max_attempts", job.MaxAttempts,
			"error", jobErr,
		)

		// Re-queue for retry
		updatedData, _ := json.Marshal(&job)
		pipe := q.client.Pipeline()
		pipe.Set(ctx, jobKey, updatedData, 24*time.Hour)
		pipe.LPush(ctx, jobQueueKey, jobID.String())
		pipe.Del(ctx, lockKey)
		pipe.Exec(ctx)
	} else {
		job.Status = models.JobStatusFailed
		isPermanent = true
		logger.Error(ctx, "job failed permanently",
			"job_id", jobID,
			"attempts", job.AttemptCount,
			"error", jobErr,
		)

		// Just update status and release lock
		updatedData, _ := json.Marshal(&job)
		pipe := q.client.Pipeline()
		pipe.Set(ctx, jobKey, updatedData, 24*time.Hour)
		pipe.Del(ctx, lockKey)
		pipe.Exec(ctx)
	}

	if q.onJobStateChange != nil {
		if err := q.onJobStateChange(ctx, &job); err != nil {
			logger.Error(ctx, "failed to persist job state", "error", err)
		}
	}

	// Call permanent failure callback
	q.mu.RLock()
	callback := q.onJobFailedPermanently
	q.mu.RUnlock()

	if isPermanent && callback != nil {
		callback(ctx, &job, jobErr)
	}

	return nil
}

// RegisterHandler registers a handler for a job type
func (q *RedisQueue) RegisterHandler(jobType string, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = handler
}

// Start starts processing jobs with multiple workers
func (q *RedisQueue) Start(ctx context.Context) error {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return fmt.Errorf("queue already running")
	}
	q.running = true
	q.mu.Unlock()

	// Start multiple workers
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}

	logger.Info(ctx, "Redis queue started", "workers", q.workerCount)

	return nil
}

// Stop stops processing jobs
func (q *RedisQueue) Stop() error {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return nil
	}
	q.running = false
	q.mu.Unlock()

	close(q.stopCh)
	q.wg.Wait()

	return q.client.Close()
}

func (q *RedisQueue) worker(ctx context.Context, workerID int) {
	defer q.wg.Done()

	logger.Info(ctx, "worker started", "worker_id", workerID)

	for {
		select {
		case <-q.stopCh:
			logger.Info(ctx, "worker stopping", "worker_id", workerID)
			return
		case <-ctx.Done():
			return
		default:
			job, err := q.Dequeue(ctx)
			if err != nil {
				logger.Error(ctx, "failed to dequeue job", "worker_id", workerID, "error", err)
				time.Sleep(time.Second)
				continue
			}
			if job == nil {
				continue
			}

			q.processJob(ctx, job, workerID)
		}
	}
}

func (q *RedisQueue) processJob(ctx context.Context, job *models.Job, workerID int) {
	q.mu.RLock()
	handler, ok := q.handlers[job.JobType]
	q.mu.RUnlock()

	if !ok {
		logger.Error(ctx, "no handler registered for job type",
			"job_type", job.JobType,
			"worker_id", workerID,
		)
		q.Fail(ctx, job.ID, fmt.Errorf("no handler for job type: %s", job.JobType))
		return
	}

	jobCtx := logger.WithProjectID(ctx, job.ProjectID.String())
	jobCtx = logger.WithJobID(jobCtx, job.ID.String())
	jobCtx = logger.WithStage(jobCtx, job.JobType)

	logger.Info(jobCtx, "processing job", "worker_id", workerID)

	if err := handler(jobCtx, job); err != nil {
		q.Fail(ctx, job.ID, err)
		return
	}

	result, _ := json.Marshal(map[string]string{"status": "completed"})
	q.Complete(ctx, job.ID, result)
}

// Close closes the Redis connection
func (q *RedisQueue) Close() error {
	return q.client.Close()
}
