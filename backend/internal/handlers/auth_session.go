package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-gonic/gin"
)

// Error codes returned by the session endpoints so the frontend can decide
// between retrying, reloading stored tokens, and sending the user to login.
const (
	authCodeInvalidRefreshToken = "invalid_refresh_token"
	authCodeTokenRotated        = "token_rotated"
	authCodeRevoked             = "revoked"
)

// AuthRequestBodyLimit caps request bodies on the public /auth endpoints,
// which are reachable without authentication. BCC access tokens are a few
// KB at most (bounded to 8 KiB below); refresh tokens are
// services.RefreshTokenLength bytes.
const AuthRequestBodyLimit = 16 << 10

type exchangeRequest struct {
	Token string `json:"token" binding:"required,max=8192"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=128"`
}

// bindAuthRequest decodes a JSON body, answering 413 when the body limit
// was hit and 400 for anything else that doesn't validate.
func bindAuthRequest(c *gin.Context, req any, field string) bool {
	err := c.ShouldBindJSON(req)
	if err == nil {
		return true
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return false
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": field + " is required and must be a valid token"})
	return false
}

// Exchange handles POST /auth/exchange. It validates a login.bcc.no (or
// Brunstad TV) token, finds or creates the user, and opens a Wayfarer
// session. This is the only endpoint that depends on the external identity
// provider; afterwards the client renews through Refresh.
func (h *AuthHandler) Exchange(c *gin.Context) {
	var req exchangeRequest
	if !bindAuthRequest(c, &req, "token") {
		return
	}
	ctx := c.Request.Context()

	userID, authErr := h.authenticateExternalToken(ctx, req.Token)
	if authErr != nil {
		c.JSON(authErr.status, gin.H{"error": authErr.message})
		return
	}

	pair, err := h.SessionService.StartSession(ctx, userID, c.GetHeader("User-Agent"))
	if errors.Is(err, services.ErrUserRevoked) {
		slog.Warn("auth: exchange blocked by revocation list", "user_id", userID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access revoked", "code": authCodeRevoked})
		return
	}
	if err != nil {
		slog.Error("auth: failed to start session", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate authentication token"})
		return
	}

	slog.Info("auth: session started", "user_id", userID)
	c.JSON(http.StatusOK, pair)
}

// Refresh handles POST /auth/refresh. It rotates the refresh token and
// returns a new token pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if !bindAuthRequest(c, &req, "refresh_token") {
		return
	}

	pair, err := h.SessionService.Refresh(c.Request.Context(), req.RefreshToken, c.GetHeader("User-Agent"))
	switch {
	case err == nil:
		c.JSON(http.StatusOK, pair)
	case errors.Is(err, services.ErrRefreshTokenRotated):
		// Another tab or request rotated this token moments ago. The client
		// should use the tokens that request stored.
		c.JSON(http.StatusConflict, gin.H{"error": "refresh token already rotated", "code": authCodeTokenRotated})
	case errors.Is(err, services.ErrUserRevoked):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access revoked", "code": authCodeRevoked})
	case errors.Is(err, services.ErrRefreshTokenInvalid), errors.Is(err, services.ErrRefreshTokenReused):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token", "code": authCodeInvalidRefreshToken})
	default:
		slog.Error("auth: failed to refresh session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh session"})
	}
}

// Logout handles POST /auth/logout. It revokes the session behind the
// refresh token and always succeeds for unknown tokens.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	if !bindAuthRequest(c, &req, "refresh_token") {
		return
	}

	if err := h.SessionService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		slog.Error("auth: failed to revoke session on logout", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log out"})
		return
	}
	c.Status(http.StatusNoContent)
}
