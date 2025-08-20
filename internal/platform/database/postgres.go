package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
	MaxConns int
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config *PostgresConfig, logger *logrus.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(config.MaxConns)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Infof("Connected to PostgreSQL database %s on %s:%d", config.Database, config.Host, config.Port)

	return &PostgresDB{
		pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
		db.logger.Info("PostgreSQL connection pool closed")
	}
}

// GetPool returns the underlying connection pool
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Ping checks if the database is accessible
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Exec executes a query without returning rows
func (db *PostgresDB) Exec(ctx context.Context, sql string, arguments ...interface{}) error {
	_, err := db.pool.Exec(ctx, sql, arguments...)
	return err
}

// Query executes a query and returns rows
func (db *PostgresDB) Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, arguments...)
}

// QueryRow executes a query and returns a single row
func (db *PostgresDB) QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, arguments...)
}

// Begin starts a transaction
func (db *PostgresDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// BeginTx starts a transaction with options
func (db *PostgresDB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.pool.BeginTx(ctx, txOptions)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}
