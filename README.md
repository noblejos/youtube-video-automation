# YouTube Video Automation System

A backend system for automated generation of faceless YouTube history videos.

## Features

- **Script Generation**: Generates structured short-form scripts from topic input
- **Scene Generation**: Breaks scripts into 6-8 scenes with narration, mood, keywords, and visual prompts
- **Voice Generation**: Text-to-speech using provider abstraction (AWS Polly or mock)
- **Image Generation**: AI image generation using provider abstraction (AWS Titan or mock)
- **Subtitle Generation**: Automatic SRT subtitle creation
- **Video Rendering**: FFmpeg-based video assembly with Ken Burns effect
- **Workflow Engine**: Application-owned state machine with retry support
- **Human Review**: Draft generation pipeline with approve/reject flow

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   API       │────▶│  Workflow   │────▶│   Queue     │
│   Server    │     │   Service   │     │   (Memory)  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                    ┌──────────────────────────┼──────────────────────────┐
                    │                          │                          │
              ┌─────▼─────┐            ┌───────▼───────┐          ┌───────▼───────┐
              │  Script   │            │    Voice      │          │    Image      │
              │  Service  │            │    Service    │          │    Service    │
              └───────────┘            └───────────────┘          └───────────────┘
                    │                          │                          │
              ┌─────▼─────┐            ┌───────▼───────┐          ┌───────▼───────┐
              │  Scene    │            │    Subtitle   │          │    Render     │
              │  Service  │            │    Service    │          │    Service    │
              └───────────┘            └───────────────┘          └───────────────┘
```

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- FFmpeg (for video rendering)
- Docker & Docker Compose (optional)

### Local Development Setup

1. **Clone and setup**:
   ```bash
   git clone <repo>
   cd youtube-video-automation
   make dev-setup
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Start PostgreSQL** (using Docker):
   ```bash
   docker run -d --name postgres \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=youtube_automation \
     -p 5432:5432 \
     postgres:16-alpine
   ```

4. **Run migrations**:
   ```bash
   make migrate-up
   ```

5. **Start the API server**:
   ```bash
   make run-api
   ```

### Using Docker Compose

```bash
# Build and start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## API Endpoints

### Create Project
```bash
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "The Rise and Fall of Mansa Musa",
    "channel_style": "dramatic_history_shorts",
    "target_duration_sec": 120,
    "aspect_ratio": "9:16"
  }'
```

### Get Project Status
```bash
curl http://localhost:8080/projects/{project_id}
```

### Get Project Manifest
```bash
curl http://localhost:8080/projects/{project_id}/manifest
```

### Approve Project
```bash
curl -X POST http://localhost:8080/projects/{project_id}/approve \
  -H "Content-Type: application/json" \
  -d '{"notes": "Looks good", "acted_by": "reviewer@example.com"}'
```

### Reject Project
```bash
curl -X POST http://localhost:8080/projects/{project_id}/reject \
  -H "Content-Type: application/json" \
  -d '{"notes": "Needs revision", "acted_by": "reviewer@example.com"}'
```

### Retry Failed Project
```bash
curl -X POST http://localhost:8080/projects/{project_id}/retry
```

## Project Lifecycle

```
CREATED
  └─▶ SCRIPT_GENERATING ─▶ SCRIPT_READY
                              └─▶ SCENES_GENERATING ─▶ SCENES_READY
                                                         └─▶ VOICE_GENERATING ─▶ VOICE_READY
                                                                                    └─▶ ASSETS_GENERATING ─▶ ASSETS_READY
                                                                                                               └─▶ SUBTITLES_GENERATING ─▶ SUBTITLES_READY
                                                                                                                                            └─▶ RENDERING ─▶ RENDER_READY
                                                                                                                                                               └─▶ REVIEW_PACKAGED ─▶ IN_REVIEW
                                                                                                                                                                                        ├─▶ APPROVED ─▶ PUBLISHING ─▶ PUBLISHED
                                                                                                                                                                                        └─▶ REJECTED
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `PORT` | HTTP server port | 8080 |
| `DATABASE_URL` | PostgreSQL connection URL | postgres://... |
| `STORAGE_BACKEND` | Storage backend (local/s3) | local |
| `STORAGE_BASE_DIR` | Local storage directory | ./storage |
| `VOICE_PROVIDER` | Voice provider (mock/aws_polly) | mock |
| `IMAGE_PROVIDER` | Image provider (mock/aws_titan) | mock |
| `FFMPEG_PATH` | Path to FFmpeg binary | ffmpeg |
| `RENDER_FPS` | Video frame rate | 30 |
| `RENDER_WIDTH` | Video width | 1080 |
| `RENDER_HEIGHT` | Video height | 1920 |

## Project Structure

```
├── cmd/
│   ├── api/          # API server entrypoint
│   └── worker/       # Worker entrypoint
├── internal/
│   ├── api/          # HTTP handlers
│   ├── config/       # Configuration
│   ├── db/           # Database connection
│   ├── handlers/     # Job handlers
│   ├── images/       # Image service
│   ├── jobs/         # Job queue
│   ├── projects/     # Project repository
│   ├── providers/    # Provider implementations
│   │   ├── image/
│   │   ├── storage/
│   │   └── voice/
│   ├── render/       # Video rendering
│   ├── scenes/       # Scene service
│   ├── scripts/      # Script service
│   ├── storage/      # Storage interface
│   ├── subtitles/    # Subtitle service
│   ├── voice/        # Voice service
│   └── workflow/     # Workflow orchestration
├── pkg/
│   ├── contracts/    # API contracts
│   ├── logger/       # Structured logging
│   └── models/       # Domain models
├── migrations/       # SQL migrations
└── deploy/           # Docker configuration
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Makefile Targets

```bash
make help           # Show all available targets
make build          # Build all binaries
make run-api        # Run API server
make run-worker     # Run worker
make test           # Run tests
make docker-up      # Start Docker services
make docker-down    # Stop Docker services
make migrate-up     # Run database migrations
make migrate-down   # Rollback migrations
make test-project   # Create a test project via API
```

## Provider Abstraction

The system uses provider interfaces for:

- **Voice**: `voice.Provider` - Implement for TTS services
- **Image**: `images.Provider` - Implement for image generation
- **Storage**: `storage.Provider` - Implement for file storage

Mock providers are included for local development without AWS credentials.

## License

MIT
# youtube-video-automation
