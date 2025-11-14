package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
// All environment variables are loaded once at startup
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Members  MembersConfig
	Auth0    Auth0Config
	OTEL     OTELConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         int
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogQueries      bool // Enable query logging for debugging
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret            string
	Issuer            string
	BrunstadTVJWKSURL string
	BrunstadTVIssuer  string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// MembersConfig holds BCC Members API configuration
type MembersConfig struct {
	Domain string // Members API domain
}

// Auth0Config holds Auth0 client credentials configuration
type Auth0Config struct {
	Domain       string // Auth0 domain
	ClientID     string // Auth0 client ID
	ClientSecret string // Auth0 client secret
}

// OTELConfig holds OpenTelemetry configuration
type OTELConfig struct {
	Enabled          bool    // Enable/disable tracing
	ServiceName      string  // Service name for traces
	ServiceVersion   string  // Service version
	ExporterEndpoint string  // OTLP exporter endpoint (e.g., "localhost:4317")
	ExporterInsecure bool    // Use insecure connection (no TLS)
	SamplingRatio    float64 // Sampling ratio (0.0 to 1.0)
}

// Load reads all environment variables and returns a Config struct
// This should be called once at application startup
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 10*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 1*time.Minute),
			LogQueries:      getEnvAsBool("DB_LOG_QUERIES", false),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", ""),
			Issuer:            getEnv("JWT_ISSUER", "wayfarer"),
			BrunstadTVJWKSURL: getEnv("BRUNSTAD_TV_JWKS_URL", "https://api.brunstad.tv/.well-known/jwks.json"),
			BrunstadTVIssuer:  getEnv("BRUNSTAD_TV_JWT_ISSUER", "https://api.brunstad.tv/"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Members: MembersConfig{
			Domain: getEnv("MEMBERS_API_DOMAIN", ""),
		},
		Auth0: Auth0Config{
			Domain:       getEnv("AUTH0_DOMAIN", ""),
			ClientID:     getEnv("AUTH0_CLIENT_ID", ""),
			ClientSecret: getEnv("AUTH0_CLIENT_SECRET", ""),
		},
		OTEL: OTELConfig{
			Enabled:          getEnvAsBool("OTEL_ENABLED", false),
			ServiceName:      getEnv("OTEL_SERVICE_NAME", "wayfarer-backend"),
			ServiceVersion:   getEnv("OTEL_SERVICE_VERSION", "dev"),
			ExporterEndpoint: getEnv("OTEL_EXPORTER_ENDPOINT", "localhost:4317"),
			ExporterInsecure: getEnvAsBool("OTEL_EXPORTER_INSECURE", true),
			SamplingRatio:    getEnvAsFloat("OTEL_SAMPLING_RATIO", 1.0),
		},
	}

	// Validate required fields
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// Helper functions to read environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}
