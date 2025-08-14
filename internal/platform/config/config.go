package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Security SecurityConfig `mapstructure:"security"`
	OTel     OTelConfig     `mapstructure:"otel"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name             string        `mapstructure:"name"`
	HTTPAddr         string        `mapstructure:"http_addr"`
	LogLevel         string        `mapstructure:"log_level"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout"`
	Environment      string        `mapstructure:"environment"`
	Version          string        `mapstructure:"version"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Mongo    MongoConfig    `mapstructure:"mongo"`
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int    `mapstructure:"max_conns"`
}

// MongoConfig holds MongoDB configuration
type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
	PoolSize int    `mapstructure:"pool_size"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers   []string `mapstructure:"brokers"`
	ClientID  string   `mapstructure:"client_id"`
	GroupID   string   `mapstructure:"group_id"`
	Version   string   `mapstructure:"version"`
	Topics    Topics   `mapstructure:"topics"`
}

// Topics holds Kafka topic names
type Topics struct {
	PointsEarned      string `mapstructure:"points_earned"`
	RedemptionRequest string `mapstructure:"redemption_request"`
	RedemptionComplete string `mapstructure:"redemption_complete"`
	RedemptionFailed  string `mapstructure:"redemption_failed"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
	MTLS MTLSConfig `mapstructure:"mtls"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Issuer     string        `mapstructure:"issuer"`
	Audience   string        `mapstructure:"audience"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// MTLSConfig holds mTLS configuration
type MTLSConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	CAFile   string `mapstructure:"ca_file"`
}

// OTelConfig holds OpenTelemetry configuration
type OTelConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ServiceName  string `mapstructure:"service_name"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
}

// Load loads configuration from environment variables and config files
func Load(serviceName string) (*Config, error) {
	viper.SetDefault("app.name", serviceName)
	viper.SetDefault("app.http_addr", ":8080")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.shutdown_timeout", "15s")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.version", "1.0.0")

	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.ssl_mode", "disable")
	viper.SetDefault("database.postgres.max_conns", 10)

	viper.SetDefault("database.mongo.timeout", "10s")

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.version", "2.8.0")
	viper.SetDefault("kafka.topics.points_earned", "points.earned.v1")
	viper.SetDefault("kafka.topics.redemption_request", "redemption.requested.v1")
	viper.SetDefault("kafka.topics.redemption_complete", "redemption.completed.v1")
	viper.SetDefault("kafka.topics.redemption_failed", "redemption.failed.v1")

	viper.SetDefault("security.jwt.expiration", "24h")
	viper.SetDefault("security.mtls.enabled", false)

	viper.SetDefault("otel.enabled", true)
	viper.SetDefault("otel.otlp_endpoint", "http://localhost:4317")

	// Set environment variable prefix
	viper.SetEnvPrefix(strings.ToUpper(serviceName))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to read config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(fmt.Sprintf("./cmd/%s", serviceName))

	// Ignore config file not found errors
	viper.ReadInConfig()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetDSN returns the PostgreSQL connection string
func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
}

// GetMongoURI returns the MongoDB connection URI
func (c *MongoConfig) GetMongoURI() string {
	if c.URI != "" {
		return c.URI
	}
	return fmt.Sprintf("mongodb://localhost:27017/%s", c.Database)
}
