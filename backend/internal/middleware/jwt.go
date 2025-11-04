package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Context key types for user information
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey contextKey = "user_role"
)

// WayfarerClaims represents the JWT claims issued by Wayfarer
type WayfarerClaims struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
	jwt.RegisteredClaims
}

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
		c.Set("user_role", claims.UserRole)

		slog.Debug("JWT validated",
			"user_id", claims.UserID,
			"role", claims.UserRole,
		)

		c.Next()
	}
}

// validateToken parses and validates a JWT token
func validateToken(tokenString string, cfg config.JWTConfig) (*WayfarerClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &WayfarerClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*WayfarerClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Verify issuer if configured
	if cfg.Issuer != "" && claims.Issuer != cfg.Issuer {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
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
