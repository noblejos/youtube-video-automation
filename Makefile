.PHONY: build run test lint clean docker-build docker-up docker-down migrate

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
API_BINARY=api
WORKER_BINARY=worker

# Build directories
BUILD_DIR=./build

# Database
DATABASE_URL?=postgres://postgres:postgres@localhost:5432/youtube_automation?sslmode=disable

all: build

build: build-api build-worker

build-api:
	@echo "Building API..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(API_BINARY) ./cmd/api

build-worker:
	@echo "Building Worker..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(WORKER_BINARY) ./cmd/worker

run-api: build-api
	@echo "Running API..."
	$(BUILD_DIR)/$(API_BINARY)

run-worker: build-worker
	@echo "Running Worker..."
	$(BUILD_DIR)/$(WORKER_BINARY)

test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linter..."
	golangci-lint run ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker-compose -f deploy/compose/docker-compose.yml build

docker-up:
	@echo "Starting Docker containers..."
	docker-compose -f deploy/compose/docker-compose.yml up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose -f deploy/compose/docker-compose.yml down

docker-logs:
	docker-compose -f deploy/compose/docker-compose.yml logs -f

# Database commands
migrate-up:
	@echo "Running migrations..."
	psql $(DATABASE_URL) -f migrations/001_initial_schema.up.sql

migrate-down:
	@echo "Rolling back migrations..."
	psql $(DATABASE_URL) -f migrations/001_initial_schema.down.sql

db-reset: migrate-down migrate-up
	@echo "Database reset complete"

# Development helpers
dev-setup: deps
	@echo "Setting up development environment..."
	@mkdir -p storage
	@cp -n .env.example .env 2>/dev/null || true
	@echo "Development setup complete"

# Create test project
test-project:
	@echo "Creating test project..."
	curl -X POST http://localhost:8080/projects \
		-H "Content-Type: application/json" \
		-d '{"topic": "The Rise and Fall of Mansa Musa", "channel_style": "dramatic_history_shorts", "target_duration_sec": 120, "aspect_ratio": "9:16"}'

# Check project status
check-project:
	@echo "Enter project ID: " && read id && \
	curl -s http://localhost:8080/projects/$$id | jq

# Get project manifest
get-manifest:
	@echo "Enter project ID: " && read id && \
	curl -s http://localhost:8080/projects/$$id/manifest | jq

help:
	@echo "Available targets:"
	@echo "  build         - Build all binaries"
	@echo "  build-api     - Build API binary"
	@echo "  build-worker  - Build Worker binary"
	@echo "  run-api       - Build and run API"
	@echo "  run-worker    - Build and run Worker"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy modules"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-up     - Start Docker containers"
	@echo "  docker-down   - Stop Docker containers"
	@echo "  docker-logs   - View Docker logs"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  db-reset      - Reset database"
	@echo "  dev-setup     - Setup development environment"
	@echo "  test-project  - Create a test project"
	@echo "  check-project - Check project status"
	@echo "  get-manifest  - Get project manifest"
