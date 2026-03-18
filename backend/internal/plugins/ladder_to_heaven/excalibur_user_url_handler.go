package ladder_to_heaven

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// excaliburQuerier defines the database operations needed by the excalibur user URL handler.
type excaliburQuerier interface {
	GetUserByID(ctx context.Context, id string) (*sqlc.GetUserByIDRow, error)
	GetChurchByID(ctx context.Context, id string) (*sqlc.GetChurchByIDRow, error)
}

// excaliburSettingsProvider defines the settings operations needed by the excalibur user URL handler.
type excaliburSettingsProvider interface {
	GetCurrentProjectID(ctx context.Context) (string, error)
}

// excaliburUserURLHandler handles requests for Excalibur user login URLs.
type excaliburUserURLHandler struct {
	db              *database.DB
	settingsService *services.SettingsService
	secretKey       string
	baseURL         string
	jwtConfig       config.JWTConfig

	// For testing - when set, these override the default implementations
	testQuerier  excaliburQuerier
	testSettings excaliburSettingsProvider
}

// ExcaliburClaims represents the JWT claims for Excalibur tokens.
type ExcaliburClaims struct {
	UserName   string `json:"userName"`
	UserID     string `json:"userId"`
	ChurchID   string `json:"churchId"`
	ChurchName string `json:"churchName"`
	ProjectID  string `json:"projectId"`
	jwt.RegisteredClaims
}

// handle generates an Excalibur user login URL with a JWT token.
func (h *excaliburUserURLHandler) handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get required path parameter
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	// Check if feature is enabled
	if h.secretKey == "" {
		slog.Warn("excalibur_user_url: feature disabled, PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_SECRET_KEY not set")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "feature disabled"})
		return
	}

	if h.baseURL == "" {
		slog.Warn("excalibur_user_url: feature disabled, PLUGIN_LADDER_TO_HEAVEN_EXCALIBUR_BASE_URL not set")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "feature disabled"})
		return
	}

	// Get user ID from JWT middleware context
	userIDValue, exists := c.Get("user_id")
	if !exists || userIDValue == nil {
		slog.Warn("excalibur_user_url: user not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		slog.Warn("excalibur_user_url: invalid user_id in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Use test implementations if set, otherwise use real ones
	querier := h.getQuerier()
	settings := h.getSettings()

	// Fetch user data from database
	user, err := querier.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("excalibur_user_url: failed to get user",
			"error", err,
			"user_id", userID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user data"})
		return
	}

	// Fetch church data from database
	church, err := querier.GetChurchByID(ctx, user.ChurchID)
	if err != nil {
		slog.Error("excalibur_user_url: failed to get church",
			"error", err,
			"church_id", user.ChurchID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve church data"})
		return
	}

	// Get current project ID from settings
	projectID, err := settings.GetCurrentProjectID(ctx)
	if err != nil {
		slog.Error("excalibur_user_url: failed to get current project ID",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve project configuration"})
		return
	}

	// Create claims with 6-month validity
	now := time.Now()
	expiresAt := now.AddDate(0, 6, 0) // 6 months

	claims := ExcaliburClaims{
		UserName:   user.Name,
		UserID:     user.ID,
		ChurchID:   user.ChurchID,
		ChurchName: church.Name,
		ProjectID:  projectID,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"excalibur"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    h.jwtConfig.Issuer,
		},
	}

	// Sign the token using HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		slog.Error("excalibur_user_url: failed to sign token",
			"error", err,
			"user_id", userID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Build the user login URL, ensuring exactly one slash between base URL and path
	baseURL := strings.TrimSuffix(h.baseURL, "/")
	normalizedPath := "/" + strings.TrimPrefix(path, "/")
	userURL := fmt.Sprintf("%s%s?token=%s", baseURL, normalizedPath, url.QueryEscape(tokenString))

	slog.Info("excalibur_user_url: redirecting to Excalibur",
		"user_id", userID,
		"church_id", user.ChurchID,
		"project_id", projectID,
	)

	c.Redirect(http.StatusFound, userURL)
}

// getQuerier returns the querier to use (test mock or real implementation).
func (h *excaliburUserURLHandler) getQuerier() excaliburQuerier {
	if h.testQuerier != nil {
		return h.testQuerier
	}
	return h.db.Queries
}

// getSettings returns the settings provider to use (test mock or real implementation).
func (h *excaliburUserURLHandler) getSettings() excaliburSettingsProvider {
	if h.testSettings != nil {
		return h.testSettings
	}
	return h.settingsService
}
