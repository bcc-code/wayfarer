package middleware

import (
	"log/slog"
	"net/http"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/gin-gonic/gin"
)

// JWTAuth is a middleware that validates JWT tokens
// Currently a placeholder that accepts all requests and logs headers
func JWTAuth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Log the authorization header for debugging
		slog.Debug("JWT middleware",
			"authorization", authHeader,
			"path", c.Request.URL.Path,
		)

		// TODO: Implement actual JWT validation
		// For now, accept all requests
		if authHeader == "" {
			slog.Warn("Request without Authorization header", "path", c.Request.URL.Path)
		}

		// TODO: Extract user ID from JWT and set in context
		// c.Set("user_id", userID)

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
