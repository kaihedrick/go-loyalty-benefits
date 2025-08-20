package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaihedrick/go-loyalty-benefits/internal/auth"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/database"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/http"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load("auth-svc")
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// DEBUG: Print loaded configuration values
	logger.Info("=== CONFIGURATION DEBUG ===")
	logger.Infof("App HTTP Addr: '%s'", cfg.App.HTTPAddr)
	logger.Infof("App Log Level: '%s'", cfg.App.LogLevel)
	logger.Infof("App Shutdown Timeout: '%v'", cfg.App.ShutdownTimeout)
	logger.Infof("Database Host: '%s'", cfg.Database.Postgres.Host)
	logger.Infof("Database Port: %d", cfg.Database.Postgres.Port)
	logger.Infof("Database Name: '%s'", cfg.Database.Postgres.Database)
	logger.Infof("Database User: '%s'", cfg.Database.Postgres.Username)
	logger.Infof("Database Password: '%s' (length: %d)",
		cfg.Database.Postgres.Password, len(cfg.Database.Postgres.Password))
	logger.Infof("Database SSL Mode: '%s'", cfg.Database.Postgres.SSLMode)
	logger.Infof("Database Max Conns: %d", cfg.Database.Postgres.MaxConns)
	logger.Infof("JWT Secret: '%s' (length: %d)",
		cfg.Security.JWT.Secret, len(cfg.Security.JWT.Secret))
	logger.Infof("JWT Issuer: '%s'", cfg.Security.JWT.Issuer)
	logger.Infof("JWT Audience: '%s'", cfg.Security.JWT.Audience)
	logger.Infof("JWT Expiration: '%v'", cfg.Security.JWT.Expiration)
	logger.Info("=== END CONFIG DEBUG ===")

	// Set log level from config
	if level, err := logrus.ParseLevel(cfg.App.LogLevel); err == nil {
		logger.SetLevel(level)
	}

	logger.Info("Starting Auth Service...")

	// Create HTTP server
	serverConfig := &http.ServerConfig{
		Addr:            cfg.App.HTTPAddr,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: cfg.App.ShutdownTimeout,
	}

	server := http.NewServer(serverConfig, logger)

	// Initialize database connection
	dbConfig := &database.PostgresConfig{
		Host:     cfg.Database.Postgres.Host,
		Port:     cfg.Database.Postgres.Port,
		Database: cfg.Database.Postgres.Database,
		Username: cfg.Database.Postgres.Username,
		Password: cfg.Database.Postgres.Password,
		SSLMode:  cfg.Database.Postgres.SSLMode,
		MaxConns: cfg.Database.Postgres.MaxConns,
	}

	db, err := database.NewPostgresDB(dbConfig, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize auth service
	authService := auth.NewService(cfg, logger)

	// Set database connection
	authService.SetDatabase(db)

	// Add routes
	server.AddRoutes(authService.Routes)

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

	logger.Info("Shutting down Auth Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("Auth Service stopped")
}
