package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bcc-media/wayfarer/internal/authtoken"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/gin-gonic/gin"
)

// Context key types for user information
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserRolesKey is the context key for user roles (array)
	UserRolesKey contextKey = "user_roles"
	// SessionIDKey is the context key for the auth session ID (empty for
	// legacy and m2m tokens)
	SessionIDKey contextKey = "session_id"
	// LanguageKey is the context key for the preferred language
	LanguageKey contextKey = "language"
	// UserAgentKey is the context key for the User-Agent header
	UserAgentKey contextKey = "user_agent"
)

// WayfarerClaims represents the JWT claims issued by Wayfarer
type WayfarerClaims = authtoken.Claims

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

		tokenString := parts[1]

		// Parse and validate JWT
		claims, err := validateToken(tokenString, cfg)
		if err != nil {
			slog.Warn("Invalid JWT token",
				"error", err,
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user context for GraphQL resolvers
		c.Set("user_id", claims.UserID)
		c.Set("user_roles", claims.UserRoles)
		c.Set("session_id", claims.SessionID)

		slog.Debug("JWT validated",
			"user_id", claims.UserID,
			"roles", claims.UserRoles,
		)

		c.Next()
	}
}

// validateToken parses and validates a JWT token
func validateToken(tokenString string, cfg config.JWTConfig) (*WayfarerClaims, error) {
	return authtoken.Parse(tokenString, cfg)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserRoles retrieves the user roles from the context
func GetUserRoles(ctx context.Context) []string {
	userRoles, ok := ctx.Value(UserRolesKey).([]string)
	if !ok {
		return []string{}
	}
	return userRoles
}

// GetUserAgent retrieves the user agent from the context
func GetUserAgent(ctx context.Context) string {
	userAgent, ok := ctx.Value(UserAgentKey).(string)
	if !ok {
		return ""
	}
	return userAgent
}
