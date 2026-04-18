# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Build Worker binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# API image
FROM alpine:3.19 AS api

# Install ffmpeg with fonts for subtitle rendering
RUN apk add --no-cache ca-certificates tzdata ffmpeg \
    fontconfig ttf-dejavu ttf-liberation font-noto \
    && fc-cache -f

WORKDIR /app

COPY --from=builder /api /app/api

EXPOSE 8080

CMD ["/app/api"]

# Worker image
FROM alpine:3.19 AS worker

# Install ffmpeg with fonts for subtitle rendering
RUN apk add --no-cache ca-certificates tzdata ffmpeg \
    fontconfig ttf-dejavu ttf-liberation font-noto \
    && fc-cache -f

WORKDIR /app

COPY --from=builder /worker /app/worker

CMD ["/app/worker"]
