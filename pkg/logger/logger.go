package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	projectIDKey contextKey = "project_id"
	jobIDKey     contextKey = "job_id"
	sceneIDKey   contextKey = "scene_id"
	stageKey     contextKey = "stage"
	providerKey  contextKey = "provider"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Init initializes the logger with the given options
func Init(isDevelopment bool) {
	var handler slog.Handler
	if isDevelopment {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	defaultLogger = slog.New(handler)
}

// WithProjectID adds project_id to context
func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}

// WithJobID adds job_id to context
func WithJobID(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, jobIDKey, jobID)
}

// WithSceneID adds scene_id to context
func WithSceneID(ctx context.Context, sceneID string) context.Context {
	return context.WithValue(ctx, sceneIDKey, sceneID)
}

// WithStage adds stage to context
func WithStage(ctx context.Context, stage string) context.Context {
	return context.WithValue(ctx, stageKey, stage)
}

// WithProvider adds provider to context
func WithProvider(ctx context.Context, provider string) context.Context {
	return context.WithValue(ctx, providerKey, provider)
}

// attrsFromContext extracts logging attributes from context
func attrsFromContext(ctx context.Context) []any {
	var attrs []any

	if v := ctx.Value(projectIDKey); v != nil {
		attrs = append(attrs, "project_id", v)
	}
	if v := ctx.Value(jobIDKey); v != nil {
		attrs = append(attrs, "job_id", v)
	}
	if v := ctx.Value(sceneIDKey); v != nil {
		attrs = append(attrs, "scene_id", v)
	}
	if v := ctx.Value(stageKey); v != nil {
		attrs = append(attrs, "stage", v)
	}
	if v := ctx.Value(providerKey); v != nil {
		attrs = append(attrs, "provider", v)
	}

	return attrs
}

// Info logs an info message
func Info(ctx context.Context, msg string, args ...any) {
	attrs := append(attrsFromContext(ctx), args...)
	defaultLogger.InfoContext(ctx, msg, attrs...)
}

// Debug logs a debug message
func Debug(ctx context.Context, msg string, args ...any) {
	attrs := append(attrsFromContext(ctx), args...)
	defaultLogger.DebugContext(ctx, msg, attrs...)
}

// Warn logs a warning message
func Warn(ctx context.Context, msg string, args ...any) {
	attrs := append(attrsFromContext(ctx), args...)
	defaultLogger.WarnContext(ctx, msg, attrs...)
}

// Error logs an error message
func Error(ctx context.Context, msg string, args ...any) {
	attrs := append(attrsFromContext(ctx), args...)
	defaultLogger.ErrorContext(ctx, msg, attrs...)
}

// Fatal logs a fatal message and exits
func Fatal(ctx context.Context, msg string, args ...any) {
	attrs := append(attrsFromContext(ctx), args...)
	defaultLogger.ErrorContext(ctx, msg, attrs...)
	os.Exit(1)
}
