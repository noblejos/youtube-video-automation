package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gama/youtube-video-automation/internal/storage"
	"github.com/gama/youtube-video-automation/internal/workflow"
	"github.com/gama/youtube-video-automation/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	router   *chi.Mux
	handlers *Handlers
	server   *http.Server
}

// NewServer creates a new HTTP server
func NewServer(workflow *workflow.Service, storage storage.Provider) *Server {
	handlers := NewHandlers(workflow, storage)
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(requestLogger)

	// Routes
	router.Get("/health", handlers.HealthCheck)

	router.Route("/projects", func(r chi.Router) {
		r.Post("/", handlers.CreateProject)
		r.Get("/{id}", handlers.GetProject)
		r.Get("/{id}/manifest", handlers.GetProjectManifest)
		r.Get("/{id}/download", handlers.DownloadVideo)
		r.Post("/{id}/approve", handlers.ApproveProject)
		r.Post("/{id}/reject", handlers.RejectProject)
		r.Post("/{id}/retry", handlers.RetryProject)
	})

	return &Server{
		router:   router,
		handlers: handlers,
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info(context.Background(), "starting HTTP server", "addr", addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			logger.Info(r.Context(), "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
