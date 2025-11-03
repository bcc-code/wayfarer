package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/gin-gonic/gin"
)

// JWTAuth is a middleware that validates JWT tokens and extracts user information
func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Log the authorization header for debugging
		slog.Debug("JWT middleware",
			"authorization", authHeader,
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

		// Extract token from "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: Implement actual JWT validation here
		// For now, we'll parse a mock structure for development
		// In production, use a proper JWT library like github.com/golang-jwt/jwt

		// Mock JWT parsing for development
		// Expected format: token contains role information
		// In a real implementation, you would:
		// 1. Validate the token signature
		// 2. Check expiration
		// 3. Extract claims

		// For development, we'll accept tokens in format: "user:<userId>" or "admin:<userId>" or "m2m:<systemId>"
		var userID string
		var role string

		if strings.HasPrefix(token, "user:") {
			role = "user"
			userID = strings.TrimPrefix(token, "user:")
		} else if strings.HasPrefix(token, "admin:") {
			role = "admin"
			userID = strings.TrimPrefix(token, "admin:")
		} else if strings.HasPrefix(token, "m2m:") {
			role = "m2m"
			userID = strings.TrimPrefix(token, "m2m:")
		} else {
			// Default to user role for any other token (for backward compatibility during development)
			role = "user"
			userID = token
		}

		// Set user context for GraphQL resolvers
		c.Set("user_id", userID)
		c.Set("user_role", role)

		slog.Debug("JWT parsed",
			"user_id", userID,
			"role", role,
		)

		c.Next()
	}
}

// RequireAuth is a stricter middleware that rejects requests without valid JWT
// Not currently used, but available for future use
func RequireAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// TODO: Implement actual JWT validation
		// For now, just check that something was provided

		c.Next()
	}
}
