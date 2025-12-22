package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
// All environment variables are loaded once at startup
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	APIKey   APIKeyConfig
	Log      LogConfig
	Members  MembersConfig
	Auth0    Auth0Config
	OTEL     OTELConfig
	SSF      SSFConfig
	S3       S3Config
	VAPID    VAPIDConfig
	Phrase   PhraseConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string
	Port            int
	Environment     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	StaticFilesPath string // Path to frontend static files (empty to disable)
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
	Auth0JWKSURL      string
	Auth0Issuer       string
}

// APIKeyConfig holds API key authentication configuration for external systems
type APIKeyConfig struct {
	Keys map[string]string // map[source_identifier]api_key
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

// SSFConfig holds SSF API configuration
type SSFConfig struct {
	BaseURL   string        // SSF API base URL
	APIKey    string        // Bearer token for authentication
	DebugMode bool          // Enable verbose request/response logging
	Timeout   time.Duration // Request timeout
	SyncKey   string        // Static key for sync endpoint authentication
}

// S3Config holds AWS S3 configuration for file uploads
type S3Config struct {
	Bucket          string // S3 bucket name
	Region          string // AWS region
	AccessKeyID     string // AWS access key ID (used locally, optional on Cloud Run)
	SecretAccessKey string // AWS secret access key (used locally, optional on Cloud Run)
	PublicBaseURL   string // Public base URL for uploaded files
	RoleARN         string // AWS role ARN for OIDC auth on Cloud Run
}

// VAPIDConfig holds VAPID keys for web push notifications
type VAPIDConfig struct {
	PublicKey  string // VAPID public key (base64 URL-safe)
	PrivateKey string // VAPID private key (base64 URL-safe)
	Subject    string // Contact email (mailto:admin@example.com)
}

// PhraseConfig holds Phrase TMS (Translation Management System) configuration
type PhraseConfig struct {
	Enabled     bool     // Enable/disable Phrase integration
	BaseURL     string   // Phrase API base URL
	Username    string   // Phrase username for authentication
	Password    string   // Phrase password for authentication
	ProjectUID  string   // Phrase project UID (separate project for Wayfarer)
	CallbackURL string   // Webhook callback URL for completed translations
	UserUID     string   // Phrase user UID for notifications
	Debug       bool     // Enable verbose request/response logging
	Languages   []string // Target languages for translation (configurable list)
	ExportKey   string   // Static key for export endpoint authentication (like SSF.SyncKey)
}

// Load reads all environment variables and returns a Config struct
// This should be called once at application startup
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			Environment:     getEnv("ENVIRONMENT", "development"),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			StaticFilesPath: getEnv("STATIC_FILES_PATH", ""),
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
			Auth0JWKSURL:      getEnv("AUTH0_JWKS_URL", "https://login.bcc.no/.well-known/jwks.json"),
			Auth0Issuer:       getEnv("AUTH0_JWT_ISSUER", "https://login.bcc.no/"),
		},
		APIKey: APIKeyConfig{
			Keys: parseAPIKeys(getEnv("EXTERNAL_API_KEYS", "")),
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
		SSF: SSFConfig{
			BaseURL:   getEnv("SSF_API_BASE_URL", "https://api.sssf.life"),
			APIKey:    getEnv("SSF_API_KEY", ""),
			DebugMode: getEnvAsBool("SSF_DEBUG_MODE", false),
			Timeout:   getEnvAsDuration("SSF_API_TIMEOUT", 10*time.Second),
			SyncKey:   getEnv("SSF_SYNC_KEY", ""),
		},
		S3: S3Config{
			Bucket:          getEnv("AWS_S3_BUCKET", ""),
			Region:          getEnv("AWS_S3_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_S3_SECRET_ACCESS_KEY", ""),
			PublicBaseURL:   getEnv("AWS_S3_PUBLIC_BASE_URL", ""),
			RoleARN:         getEnv("AWS_S3_ROLE_ARN", ""),
		},
		VAPID: VAPIDConfig{
			PublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
			PrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
			Subject:    getEnv("VAPID_SUBJECT", ""),
		},
		Phrase: PhraseConfig{
			Enabled:     getEnvAsBool("PHRASE_ENABLED", false),
			BaseURL:     getEnv("PHRASE_BASE_URL", "https://cloud.memsource.com/web/api2"),
			Username:    getEnv("PHRASE_USERNAME", ""),
			Password:    getEnv("PHRASE_PASSWORD", ""),
			ProjectUID:  getEnv("PHRASE_PROJECT_UID", ""),
			CallbackURL: getEnv("PHRASE_CALLBACK_URL", ""),
			UserUID:     getEnv("PHRASE_USER_UID", ""),
			Debug:       getEnvAsBool("PHRASE_DEBUG", false),
			Languages:   parseLanguages(getEnv("PHRASE_TARGET_LANGUAGES", "da,de,en,fr,fi,hu,it,nl,pl,pt,ro,es,ru")),
			ExportKey:   getEnv("TRANSLATIONS_EXPORT_KEY", ""),
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

// parseAPIKeys parses API keys from environment variable format
// Format: "source1:key1,source2:key2,source3:key3"
// Returns a map of source identifier to API key
func parseAPIKeys(value string) map[string]string {
	keys := make(map[string]string)
	if value == "" {
		return keys
	}

	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			source := strings.TrimSpace(parts[0])
			key := strings.TrimSpace(parts[1])
			if source != "" && key != "" {
				keys[source] = key
			}
		}
	}

	return keys
}

// parseLanguages parses comma-separated language codes
// Format: "da,de,en,fr"
// Returns a slice of language codes
func parseLanguages(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	languages := make([]string, 0, len(parts))
	for _, part := range parts {
		if lang := strings.TrimSpace(part); lang != "" {
			languages = append(languages, lang)
		}
	}
	return languages
}
