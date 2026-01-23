package ladder_to_heaven

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// cryptexQuerier defines the database operations needed by the cryptex token handler.
type cryptexQuerier interface {
	GetUserByID(ctx context.Context, id string) (*sqlc.GetUserByIDRow, error)
	GetChurchByID(ctx context.Context, id string) (*sqlc.GetChurchByIDRow, error)
}

// cryptexSettingsProvider defines the settings operations needed by the cryptex token handler.
type cryptexSettingsProvider interface {
	GetCurrentProjectID(ctx context.Context) (string, error)
}

// cryptexTokenHandler handles requests for Cryptex JWT tokens.
type cryptexTokenHandler struct {
	db              *database.DB
	settingsService *services.SettingsService
	secretKey       string
	jwtConfig       config.JWTConfig

	// For testing - when set, these override the default implementations
	testQuerier  cryptexQuerier
	testSettings cryptexSettingsProvider
}

// CryptexClaims represents the JWT claims for Cryptex tokens.
type CryptexClaims struct {
	UserName   string `json:"userName"`
	UserID     string `json:"userId"`
	ChurchID   string `json:"churchId"`
	ChurchName string `json:"churchName"`
	ProjectID  string `json:"projectId"`
	jwt.RegisteredClaims
}

// allowedRoles are the roles that can request a Cryptex token.
var allowedRoles = []string{"church_admin", "admin", "superadmin"}

// handle generates a JWT token for Cryptex integration.
func (h *cryptexTokenHandler) handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if feature is enabled
	if h.secretKey == "" {
		slog.Warn("cryptex_token: feature disabled, PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_SECRET_KEY not set")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "feature disabled"})
		return
	}

	// Get user ID from JWT middleware context
	userIDValue, exists := c.Get("user_id")
	if !exists || userIDValue == nil {
		slog.Warn("cryptex_token: user not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		slog.Warn("cryptex_token: invalid user_id in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Get user roles from JWT middleware context
	userRolesValue, exists := c.Get("user_roles")
	if !exists || userRolesValue == nil {
		slog.Warn("cryptex_token: no roles found for user", "user_id", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	userRoles, ok := userRolesValue.([]string)
	if !ok {
		slog.Warn("cryptex_token: invalid user_roles in context", "user_id", userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Check if user has at least one of the allowed roles
	if !hasAllowedRole(userRoles) {
		slog.Warn("cryptex_token: user lacks required role",
			"user_id", userID,
			"user_roles", userRoles,
			"required_any", allowedRoles,
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Use test implementations if set, otherwise use real ones
	querier := h.getQuerier()
	settings := h.getSettings()

	// Fetch user data from database
	user, err := querier.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("cryptex_token: failed to get user",
			"error", err,
			"user_id", userID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user data"})
		return
	}

	// Fetch church data from database
	church, err := querier.GetChurchByID(ctx, user.ChurchID)
	if err != nil {
		slog.Error("cryptex_token: failed to get church",
			"error", err,
			"church_id", user.ChurchID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve church data"})
		return
	}

	// Get current project ID from settings
	projectID, err := settings.GetCurrentProjectID(ctx)
	if err != nil {
		slog.Error("cryptex_token: failed to get current project ID",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve project configuration"})
		return
	}

	// Create claims with 6-month validity
	now := time.Now()
	expiresAt := now.AddDate(0, 6, 0) // 6 months

	claims := CryptexClaims{
		UserName:   user.Name,
		UserID:     user.ID,
		ChurchID:   user.ChurchID,
		ChurchName: church.Name,
		ProjectID:  projectID,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"LADD-Cryptex"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    h.jwtConfig.Issuer,
		},
	}

	// Sign the token using HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		slog.Error("cryptex_token: failed to sign token",
			"error", err,
			"user_id", userID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	slog.Info("cryptex_token: token generated successfully",
		"user_id", userID,
		"church_id", user.ChurchID,
		"project_id", projectID,
	)

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// getQuerier returns the querier to use (test mock or real implementation).
func (h *cryptexTokenHandler) getQuerier() cryptexQuerier {
	if h.testQuerier != nil {
		return h.testQuerier
	}
	return h.db.Queries
}

// getSettings returns the settings provider to use (test mock or real implementation).
func (h *cryptexTokenHandler) getSettings() cryptexSettingsProvider {
	if h.testSettings != nil {
		return h.testSettings
	}
	return h.settingsService
}

// hasAllowedRole checks if the user has at least one of the allowed roles.
func hasAllowedRole(userRoles []string) bool {
	for _, userRole := range userRoles {
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return true
			}
		}
	}
	return false
}
