package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/gin-gonic/gin"
)

// APIKeySourceKey is the context key for API key source identifier
const APIKeySourceKey contextKey = "api_key_source"

// APIKeyAuth is a middleware that validates API keys and extracts source information
func APIKeyAuth(cfg config.APIKeyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		slog.Debug("API key middleware",
			"path", c.Request.URL.Path,
		)

		if authHeader == "" {
			slog.Warn("Request without Authorization header", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <key>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <key>",
			})
			c.Abort()
			return
		}

		providedKey := parts[1]

		// Validate API key and get source identifier
		source, valid := validateAPIKey(providedKey, cfg)
		if !valid {
			slog.Warn("Invalid API key",
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Set API key source in context for handlers
		c.Set("api_key_source", source)

		slog.Debug("API key validated",
			"source", source,
		)

		c.Next()
	}
}

// validateAPIKey validates the provided API key against configured keys
// Returns the source identifier and whether the key is valid
func validateAPIKey(providedKey string, cfg config.APIKeyConfig) (string, bool) {
	for source, validKey := range cfg.Keys {
		if providedKey == validKey {
			return source, true
		}
	}
	return "", false
}

// GetAPIKeySource retrieves the API key source identifier from the context
func GetAPIKeySource(ctx context.Context) (string, bool) {
	source, ok := ctx.Value(APIKeySourceKey).(string)
	return source, ok
}
