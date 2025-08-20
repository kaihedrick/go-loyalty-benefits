package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaihedrick/go-loyalty-benefits/internal/loyalty"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/http"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load("loyalty-svc")
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level from config
	if level, err := logrus.ParseLevel(cfg.App.LogLevel); err == nil {
		logger.SetLevel(level)
	}

	logger.Info("Starting Loyalty Service...")

	// Create HTTP server
	serverConfig := &http.ServerConfig{
		Addr:            cfg.App.HTTPAddr,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: cfg.App.ShutdownTimeout,
	}

	server := http.NewServer(serverConfig, logger)

	// Initialize loyalty service
	loyaltyService := loyalty.NewService(cfg, logger)

	// Add routes
	server.AddRoutes(loyaltyService.Routes)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Errorf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Loyalty Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("Loyalty Service stopped")
}
