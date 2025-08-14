package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

// Server represents an HTTP server
type Server struct {
	router *chi.Mux
	server *http.Server
	logger *logrus.Logger
	config *ServerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
	AllowedMethods  []string
	AllowedHeaders  []string
}

// NewServer creates a new HTTP server with default configuration
func NewServer(config *ServerConfig, logger *logrus.Logger) *Server {
	if config == nil {
		config = &ServerConfig{
			Addr:            ":8080",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 15 * time.Second,
			AllowedOrigins:  []string{"*"},
			AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		}
	}

	router := chi.NewRouter()
	
	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(config.WriteTimeout))
	
	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   config.AllowedMethods,
		AllowedHeaders:   config.AllowedHeaders,
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	router.Get("/healthz", healthCheck)
	
	server := &http.Server{
		Addr:         config.Addr,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		router: router,
		server: server,
		logger: logger,
		config: config,
	}
}

// Router returns the Chi router for adding routes
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Infof("Starting HTTP server on %s", s.config.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	s.logger.Info("Shutting down HTTP server...")
	return s.server.Shutdown(shutdownCtx)
}

// healthCheck handles health check requests
func healthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "go-loyalty-benefits",
	})
}

// AddRoutes adds routes to the server
func (s *Server) AddRoutes(routes func(*chi.Mux)) {
	routes(s.router)
}

// AddMiddleware adds middleware to the server
func (s *Server) AddMiddleware(middleware func(http.Handler) http.Handler) {
	s.router.Use(middleware)
}

// GetServer returns the underlying http.Server
func (s *Server) GetServer() *http.Server {
	return s.server
}
