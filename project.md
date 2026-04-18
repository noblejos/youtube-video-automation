
# YouTube Video Automation System - Project Documentation

## Project Overview

This is a **YouTube Video Automation System** for automated generation of faceless history videos. It's a distributed Go backend with a React frontend that orchestrates a multi-stage video production pipeline using AI services.

### Core Architecture

The system uses a **workflow-driven architecture** with two separate binaries:
- **API Server** (`cmd/api`): Accepts HTTP requests, creates projects, enqueues jobs
- **Worker** (`cmd/worker`): Processes jobs from queue, executes AI generation tasks

**Key Design Pattern**: Application-owned state machine with retry support. The workflow service (`internal/workflow`) orchestrates project state transitions and job enqueueing, while handlers (`internal/handlers`) execute individual job types.

### Project Lifecycle State Machine

```
CREATED → SCRIPT_GENERATING → SCRIPT_READY → SCENES_GENERATING → SCENES_READY 
→ VOICE_GENERATING → VOICE_READY → ASSETS_GENERATING → ASSETS_READY 
→ SUBTITLES_GENERATING → SUBTITLES_READY → RENDERING → RENDER_READY 
→ REVIEW_PACKAGED → IN_REVIEW → APPROVED/REJECTED
```

Each state transition is triggered by job completion callbacks in the workflow service.

### Data Flow

1. **Project Creation**: User submits topic → API creates project record → Workflow service enqueues `SCRIPT_GENERATION` job
2. **Script Generation**: Worker handler generates script via LLM → Saves to DB → Calls `workflow.OnScriptGenerated()`
3. **Scene Breakdown**: Workflow enqueues `SCENES_GENERATION` → Worker breaks script into 6-8 scenes with visual prompts
4. **Parallel Asset Generation**: For each scene, workflow enqueues both `VOICE_GENERATION` and `IMAGE_GENERATION` jobs simultaneously
5. **Completion Check**: Each completed asset job calls `workflow.CheckAssetCompletion()` → When all assets ready, proceeds to subtitles
6. **Subtitles**: Worker generates SRT subtitles from audio timing data
7. **Render**: FFmpeg assembles scenes with Ken Burns effect, overlays subtitles
8. **Review**: Package is marked for human approval/rejection

### Provider Abstraction

The system uses **provider interfaces** for swappable implementations:

- **Voice Provider** (`internal/voice/voice.go`): 
  - `mock`: Generates dummy audio files for testing
  - `polly`: AWS Polly TTS (see `internal/providers/voice/polly/`)

- **Image Provider** (`internal/images/images.go`):
  - `mock`: Returns colored placeholder images
  - `openai`: DALL-E 3 image generation (see `internal/providers/image/openai/`)

- **LLM Provider** (for script/scene generation):
  - `mock`: Returns hardcoded script structures
  - `openai`: GPT-4 with structured prompts (see `internal/providers/llm/openai/`)

- **Storage Provider** (`internal/storage/storage.go`):
  - `local`: File-based storage in `./storage/` directory
  - (S3 implementation ready for future)

- **Queue Backend** (`internal/jobs/queue.go`):
  - `memory`: In-memory queue (single process, development)
  - `redis`: Distributed queue for multi-worker deployments

### Database Schema

PostgreSQL with the following core tables:
- `projects`: Main project records with status tracking
- `scripts`: Generated scripts with structured sections (hook, setup, build, turning_point, collapse, conclusion)
- `scenes`: Individual scenes with narration_text, visual_prompt, mood, keywords
- `audio_files`: Voice-over files with duration and speech marks
- `assets`: Generated images/videos
- `subtitles`: SRT subtitle files
- `renders`: Final video outputs
- `jobs`: Job queue persistence with retry tracking

**Important**: The `jobs` table persists job state for both in-memory and Redis queues. Job status transitions: `QUEUED → RUNNING → SUCCEEDED/FAILED`. Failed jobs retry up to `max_attempts` (default 3) before permanently failing.

## Common Development Commands

### Building
```bash
make build              # Build both API and Worker binaries to ./build/
make build-api          # Build only API server
make build-worker       # Build only Worker
```

### Running Locally
```bash
# Start PostgreSQL first (required)
docker run -d --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=youtube_automation \
  -p 5432:5432 \
  postgres:16-alpine

# Run migrations
make migrate-up

# Start API server (port 8080)
make run-api

# In separate terminal, start worker
make run-worker
```

**Important**: The API server does NOT process jobs itself. You must run the worker separately for job processing to occur.

### Docker Compose (Full Stack)
```bash
make docker-up          # Start postgres, redis, api, worker, frontend
make docker-logs        # View logs
make docker-down        # Stop all services
```

Docker Compose automatically runs migrations on postgres startup via volume mount to `migrations/`.

### Testing
```bash
make test               # Run all tests
make test-coverage      # Generate coverage report (coverage.html)

# Test the API
make test-project       # Create a test project via curl
make check-project      # Check project status (prompts for ID)
make get-manifest       # Get full project manifest
```

### Database Operations
```bash
make migrate-up         # Apply migrations
make migrate-down       # Rollback migrations
make db-reset           # Rollback then re-apply

# Manual query
psql postgres://postgres:postgres@localhost:5432/youtube_automation?sslmode=disable
```

**Migration Pattern**: SQL files in `migrations/` directory. Only one migration exists (`001_initial_schema`). New migrations should follow pattern `00X_description.up.sql` and `00X_description.down.sql`.

### Frontend Development
```bash
cd frontend
npm install
npm run dev             # Start Vite dev server (port 5173)
npm run build           # Build for production
npm run preview         # Preview production build
```

Frontend is React 19 + TypeScript + Vite. Connects to API at `http://localhost:8080`.

## Configuration

All configuration is environment-based. Copy `.env.example` to `.env` and modify:

### Critical Settings
- `DATABASE_URL`: PostgreSQL connection string (required)
- `QUEUE_BACKEND`: `memory` (dev) or `redis` (production)
- `REDIS_URL`: Redis connection (required if `QUEUE_BACKEND=redis`)
- `WORKER_COUNT`: Number of concurrent workers (default 3, only applies to worker process)

### Provider Configuration
- `LLM_PROVIDER`: `mock` or `openai` (script/scene generation)
- `VOICE_PROVIDER`: `mock` or `polly`
- `IMAGE_PROVIDER`: `mock` or `openai`
- `OPENAI_API_KEY`: Required if using openai for LLM or images
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`: Required if using polly

### Render Settings
- `FFMPEG_PATH`: Path to FFmpeg binary (default `ffmpeg`)
- `USE_MOCK_RENDER`: `true` to skip actual FFmpeg execution (dev mode)
- `RENDER_FPS`: Video frame rate (default 30)
- `RENDER_WIDTH`, `RENDER_HEIGHT`: Video dimensions (default 1080x1920 for 9:16)

**Development Defaults**: System runs fully in mock mode by default (no API keys needed). Mock providers generate fake but structurally valid data for testing the full pipeline.

## Key Implementation Details

### Job Queue and Retry Logic

Jobs are processed by the worker with automatic retry on failure:

1. Worker pulls job from queue (via `queue.Start()`)
2. Handler executes job logic
3. On success: Job marked `SUCCEEDED`, workflow callback triggered
4. On failure: Job retries up to `max_attempts`, then marked `FAILED` permanently
5. Permanent failure triggers `workflow.MarkProjectFailed()`

**Handler Registration**: `handlers.RegisterHandlers(queue)` maps job types to handler functions. Each handler is responsible for:
- Parsing job payload
- Executing business logic (calling service methods)
- Returning error to trigger retry or calling workflow callback on success

### Scene Generation Strategy

Scenes are generated from scripts using LLM with structured prompts. Each scene includes:
- `narration_text`: Exact text for voice-over
- `visual_prompt`: Detailed DALL-E prompt for image generation
- `mood`: Emotional tone (e.g., "mysterious", "triumphant")
- `keywords`: JSON array of visual elements
- `duration_sec`: Calculated from narration length

**Important**: Scene count is typically 6-8 for 120-second videos. The `CheckAssetCompletion` function waits until `COUNT(audio_files) >= COUNT(scenes)` AND `COUNT(assets) >= COUNT(scenes)` before proceeding.

### FFmpeg Rendering

The render service (`internal/render`) uses FFmpeg to:
1. Apply Ken Burns effect (slow zoom/pan) to images
2. Overlay audio narration
3. Add subtitle burn-in
4. Export to 9:16 vertical video

**Mock Mode**: When `USE_MOCK_RENDER=true`, the render service generates a dummy video file without invoking FFmpeg (fast for development).

### Storage Organization

Local storage structure:
```
./storage/
  projects/{project_id}/
    scripts/
      script.json
    audio/
      scene_001.mp3
      scene_002.mp3
    images/
      scene_001.png
    subtitles/
      video.srt
    renders/
      draft.mp4
```

**Storage Keys**: All database records reference files via `storage_key` (relative path). Storage provider implementations resolve these to actual file paths or S3 URLs.

## API Endpoints

### Create Project
```bash
POST /projects
Content-Type: application/json

{
  "topic": "The Rise and Fall of Mansa Musa",
  "channel_style": "dramatic_history_shorts",
  "target_duration_sec": 120,
  "aspect_ratio": "9:16",
  "script": "optional user-provided script",
  "title": "optional video title",
  "review_required": true
}
```

### Get Project Status
```bash
GET /projects/{project_id}
```

### Get Project Manifest
```bash
GET /projects/{project_id}/manifest
```
Returns complete project data including script, scenes, audio files, assets, and render info.

### Approve Project
```bash
POST /projects/{project_id}/approve
Content-Type: application/json

{
  "notes": "Looks good",
  "acted_by": "reviewer@example.com"
}
```

### Reject Project
```bash
POST /projects/{project_id}/reject
Content-Type: application/json

{
  "notes": "Needs revision",
  "acted_by": "reviewer@example.com"
}
```

### Retry Failed Project
```bash
POST /projects/{project_id}/retry
```
Re-queues all failed jobs for the project.

**User-Provided Scripts**: If `script` field is provided in project creation, the system skips `SCRIPT_GENERATION` job and proceeds directly to scene generation. The script must be plain text (not structured JSON).

## Project Structure

```
├── cmd/
│   ├── api/              # API server entrypoint
│   └── worker/           # Worker entrypoint
├── internal/
│   ├── api/              # HTTP handlers and routing
│   ├── config/           # Configuration loading
│   ├── db/               # Database connection
│   ├── handlers/         # Job handlers (script, scene, voice, image, render)
│   ├── images/           # Image service and repository
│   ├── jobs/             # Job queue implementations (memory, redis)
│   ├── projects/         # Project repository
│   ├── providers/        # Provider implementations
│   │   ├── image/        # mock, openai
│   │   ├── llm/          # openai (script/scene generators)
│   │   ├── storage/      # local (future: s3)
│   │   └── voice/        # mock, polly
│   ├── render/           # Video rendering service
│   ├── scenes/           # Scene service and repository
│   ├── scripts/          # Script service and repository
│   ├── storage/          # Storage interface
│   ├── subtitles/        # Subtitle service and repository
│   ├── voice/            # Voice service and repository
│   └── workflow/         # Workflow orchestration (state machine)
├── pkg/
│   ├── contracts/        # API request/response models
│   ├── logger/           # Structured logging (slog wrapper)
│   └── models/           # Domain models (Project, Scene, Job, etc.)
├── migrations/           # SQL migration files
├── deploy/
│   ├── compose/          # docker-compose.yml and .env
│   └── docker/           # Dockerfile
├── frontend/             # React + TypeScript + Vite UI
└── storage/              # Local file storage (gitignored)
```

## Testing Strategy

- **Unit Tests**: Service-level tests with mocked repositories (e.g., `internal/workflow/workflow_test.go`)
- **Integration Tests**: Not yet implemented (future: test with real Postgres + Redis)
- **Manual Testing**: Use `make test-project` to create end-to-end test projects

**Testing with Mocks**: All mock providers are designed to return realistic data instantly. A full project creation → render cycle completes in ~5-10 seconds with mocks enabled.

## Common Pitfalls

1. **Forgetting to start worker**: API will enqueue jobs but nothing happens until worker is running
2. **Wrong queue backend**: If API uses `QUEUE_BACKEND=redis` but worker uses `memory`, jobs won't be processed
3. **Missing FFmpeg**: If `USE_MOCK_RENDER=false` but FFmpeg not installed, render jobs will fail
4. **Scene asset mismatch**: The system expects exactly one audio file and one image per scene. If generation fails partway, `CheckAssetCompletion` will never trigger
5. **Go version**: Project requires Go 1.26.1+ (see go.mod)
6. **Database not running**: Always start PostgreSQL before API/worker
7. **Port conflicts**: API uses 8080, frontend uses 3000, postgres uses 5432, redis uses 6379

## Extending the System

### Adding a New Provider

1. Define interface in `internal/{domain}/{domain}.go` (e.g., `voice.Provider`)
2. Implement provider in `internal/providers/{domain}/{provider_name}/`
3. Register provider in worker startup (`cmd/worker/main.go`)
4. Add config variables to `internal/config/config.go`
5. Update `.env.example` with new settings

Example: Adding Google Cloud TTS
```go
// 1. Interface already exists in internal/voice/voice.go
type Provider interface {
    Generate(ctx context.Context, text string, voice string) (*GeneratedAudio, error)
}

// 2. Implement in internal/providers/voice/google/google.go
type GoogleProvider struct { ... }
func (g *GoogleProvider) Generate(...) { ... }

// 3. Register in cmd/worker/main.go
case "google":
    voiceProvider, err = voicegoogle.New(...)
```

### Adding a New Job Type

1. Define constant in `pkg/models/models.go` (e.g., `JobTypeTranscriptionCheck`)
2. Create payload struct in `internal/jobs/payloads.go`
3. Add handler function in `internal/handlers/handlers.go`
4. Register handler in `RegisterHandlers()`
5. Add workflow method to enqueue job in `internal/workflow/workflow.go`
6. Update state machine documentation

Example: Adding background music generation
```go
// 1. In pkg/models/models.go
const JobTypeMusicGeneration = "MUSIC_GENERATION"

// 2. In internal/jobs/payloads.go
type MusicJobPayload struct {
    JobPayload
    Mood     string
    Duration int
}

// 3. In internal/handlers/handlers.go
func (h *JobHandlers) HandleMusicGeneration(ctx context.Context, job *models.Job) error {
    payload, err := jobs.ParsePayload[jobs.MusicJobPayload](job)
    // ... generate music
    return h.workflowService.OnMusicGenerated(ctx, payload.ProjectID)
}

// 4. Register
queue.RegisterHandler(models.JobTypeMusicGeneration, h.HandleMusicGeneration)
```

### Modifying the State Machine

The workflow service (`internal/workflow/workflow.go`) controls state transitions via `projectRepo.UpdateStatus()` calls. To add new states:
1. Add constants to `pkg/models/models.go`
2. Update `OnXCompleted()` methods to enqueue next job
3. Ensure `current_step` field is updated for UI tracking
4. Update database if new columns needed

## Architecture Decisions

- **Separate API and Worker**: Allows horizontal scaling of job processing without scaling HTTP handling
- **Postgres for queue persistence**: Even in-memory queue writes to `jobs` table for crash recovery
- **Provider abstraction**: Avoids vendor lock-in, enables local development without cloud credentials
- **Structured script format**: Scripts use narrative structure (hook, setup, build, turning_point, collapse, conclusion) for consistency
- **Parallel asset generation**: Voice and images generated simultaneously per scene for performance
- **Review gate**: Human approval required before publishing (configurable via `review_required` flag)
- **No ORM**: Direct SQL with pgx for performance and control
- **Chi router**: Lightweight HTTP router with middleware support

## Dependencies

**Core**:
- `github.com/go-chi/chi/v5` v5.2.5: HTTP router
- `github.com/jackc/pgx/v5` v5.9.1: PostgreSQL driver (no ORM)
- `github.com/google/uuid` v1.6.0: UUID generation
- `github.com/redis/go-redis/v9` v9.18.0: Redis client

**AWS**:
- `github.com/aws/aws-sdk-go-v2/service/polly`: AWS Polly text-to-speech
- `github.com/aws/aws-sdk-go-v2/config`: AWS configuration

**Frontend**:
- React 19.2.4
- React Router DOM 7.14.0
- Vite 8.0.1
- TypeScript 5.9.3

**Note**: OpenAI SDK is not vendored. HTTP calls are made directly via `net/http` in provider implementations (`internal/providers/llm/openai/` and `internal/providers/image/openai/`).

## Logging

Structured logging via `pkg/logger`:
- Uses `slog` from Go standard library
- Contextual `project_id` added via `logger.WithProjectID(ctx, id)`
- Log levels: `Info`, `Warn`, `Error`, `Fatal`
- Development mode: Human-readable text format
- Production mode: JSON format for log aggregation

**Best Practice**: Always include `project_id` in logs for traceability across distributed workers.

Example:
```go
ctx = logger.WithProjectID(ctx, project.ID.String())
logger.Info(ctx, "project created", "topic", project.Topic)
logger.Error(ctx, "failed to generate script", "error", err)
```

## Security Considerations

- **No authentication/authorization**: This is a backend service. Add auth middleware before production
- **Environment variables**: Never commit `.env` files with real credentials
- **SQL injection**: Protected by parameterized queries via pgx
- **Storage access**: Local storage has no access controls. Use S3 with IAM in production

## Performance Characteristics

- **Mock mode**: Full pipeline completes in 5-10 seconds
- **With real providers**: 
  - Script generation: ~10-30 seconds (GPT-4)
  - Scene generation: ~10-30 seconds (GPT-4)
  - Voice per scene: ~2-5 seconds (Polly)
  - Image per scene: ~10-20 seconds (DALL-E 3)
  - Subtitle generation: <1 second
  - Render: 5-30 seconds depending on scene count
- **Total with AI**: ~2-5 minutes for 120-second video

**Optimization opportunities**:
- Increase `WORKER_COUNT` for parallel job processing
- Use Redis queue backend for multi-worker deployments
- Cache common LLM responses
- Use faster image models (DALL-E 2, Stable Diffusion)

## Future Enhancements

- S3 storage provider for production
- YouTube upload integration
- Background music generation
- Thumbnail generation
- A/B testing for script variations
- Analytics and reporting
- Webhook notifications
- Admin dashboard improvements
- Batch project creation
- Template system for different video styles
