package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// validTestSecret is a syntactically valid (>=32 byte) secret.
	const validTestSecret = "test-secret-that-is-at-least-32-bytes-long"

	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"SERVER_HOST", "SERVER_PORT", "DATABASE_URL", "LOG_LEVEL", "LOG_FORMAT", "JWT_SECRET",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Restore env vars after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	t.Run("loads config with defaults", func(t *testing.T) {
		// Unset vars whose defaults this subtest asserts, so ambient
		// environment (or a loaded .env) can't leak in.
		_ = os.Unsetenv("SERVER_HOST")
		_ = os.Unsetenv("SERVER_PORT")
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("LOG_FORMAT")
		_ = os.Setenv("DATABASE_URL", "postgresql://localhost/test")
		_ = os.Setenv("JWT_SECRET", validTestSecret)

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "postgresql://localhost/test", cfg.Database.URL)
		assert.Equal(t, "info", cfg.Log.Level)
		assert.Equal(t, "json", cfg.Log.Format)
	})

	t.Run("loads config from environment", func(t *testing.T) {
		_ = os.Setenv("SERVER_HOST", "localhost")
		_ = os.Setenv("SERVER_PORT", "3000")
		_ = os.Setenv("DATABASE_URL", "postgresql://localhost/wayfarer")
		_ = os.Setenv("LOG_LEVEL", "debug")
		_ = os.Setenv("LOG_FORMAT", "text")
		_ = os.Setenv("JWT_SECRET", validTestSecret)

		cfg, err := Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, 3000, cfg.Server.Port)
		assert.Equal(t, "postgresql://localhost/wayfarer", cfg.Database.URL)
		assert.Equal(t, "debug", cfg.Log.Level)
		assert.Equal(t, "text", cfg.Log.Format)
	})

	t.Run("returns error when DATABASE_URL is missing", func(t *testing.T) {
		_ = os.Unsetenv("DATABASE_URL")
		_ = os.Setenv("JWT_SECRET", validTestSecret)

		cfg, err := Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "DATABASE_URL")
	})

	t.Run("returns error when JWT_SECRET is missing", func(t *testing.T) {
		_ = os.Setenv("DATABASE_URL", "postgresql://localhost/test")
		_ = os.Unsetenv("JWT_SECRET")

		cfg, err := Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "JWT_SECRET")
	})
}

func TestValidateJWTSecret(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "rejects empty secret",
			secret:    "",
			wantErr:   true,
			errSubstr: "required",
		},
		{
			name:      "rejects secret shorter than minimum",
			secret:    strings.Repeat("a", minJWTSecretBytes-1),
			wantErr:   true,
			errSubstr: "at least",
		},
		{
			name:    "accepts secret at exactly the minimum length",
			secret:  strings.Repeat("a", minJWTSecretBytes),
			wantErr: false,
		},
		{
			name:    "accepts the historical hardcoded secret",
			secret:  "your-secret-key-for-signing-wayfarer-jwts",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJWTSecret(tt.secret)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"returns default when not set", "", 100, 100},
		{"returns parsed value", "200", 100, 200},
		{"returns default on invalid value", "invalid", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT_VAR"
			if tt.envValue != "" {
				_ = os.Setenv(key, tt.envValue)
				defer func() { _ = os.Unsetenv(key) }()
			}

			result := getEnvAsInt(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"returns default when not set", "", 5 * time.Second, 5 * time.Second},
		{"returns parsed value", "10s", 5 * time.Second, 10 * time.Second},
		{"returns default on invalid value", "invalid", 5 * time.Second, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_DURATION_VAR"
			if tt.envValue != "" {
				_ = os.Setenv(key, tt.envValue)
				defer func() { _ = os.Unsetenv(key) }()
			}

			result := getEnvAsDuration(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
