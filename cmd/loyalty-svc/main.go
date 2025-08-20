package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaihedrick/go-loyalty-benefits/internal/loyalty"
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

	logger.Info("Starting Loyalty Service...")

	// Load configuration
	cfg, err := config.Load("loyalty-svc")
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug: Print loaded configuration
	logger.Infof("=== LOYALTY SERVICE CONFIG DEBUG ===")
	logger.Infof("App Name: '%s'", cfg.App.Name)
	logger.Infof("HTTP Address: '%s'", cfg.App.HTTPAddr)
	logger.Infof("Database Host: '%s'", cfg.Database.Postgres.Host)
	logger.Infof("Database Port: '%d'", cfg.Database.Postgres.Port)
	logger.Infof("Database Name: '%s'", cfg.Database.Postgres.Database)
	logger.Infof("Database User: '%s'", cfg.Database.Postgres.Username)
	logger.Infof("Database Password: '%s' (length: %d)", cfg.Database.Postgres.Password, len(cfg.Database.Postgres.Password))
	logger.Infof("Database SSL Mode: '%s'", cfg.Database.Postgres.SSLMode)
	logger.Infof("JWT Secret: '%s' (length: %d)", cfg.Security.JWT.Secret, len(cfg.Security.JWT.Secret))
	logger.Infof("JWT Issuer: '%s'", cfg.Security.JWT.Issuer)
	logger.Infof("JWT Audience: '%s'", cfg.Security.JWT.Audience)
	logger.Infof("JWT Expiration: '%s'", cfg.Security.JWT.Expiration)
	logger.Infof("=== END CONFIG DEBUG ===")

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

	logger.Infof("Connected to PostgreSQL database %s on %s:%d", cfg.Database.Postgres.Database, cfg.Database.Postgres.Host, cfg.Database.Postgres.Port)

	// Initialize loyalty service
	loyaltyService := loyalty.NewService(cfg, logger)

	// Set database connection
	loyaltyService.SetDatabase(db)

	// Add routes
	server.AddRoutes(loyaltyService.Routes)

	// Start server
	go func() {
		logger.Infof("Starting HTTP server on %s", cfg.App.HTTPAddr)
		if err := server.Start(); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
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
		logger.Errorf("Error during server shutdown: %v", err)
	}

	logger.Info("Loyalty Service stopped")
}
