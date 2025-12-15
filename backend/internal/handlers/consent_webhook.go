package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ConsentWebhookHandler handles external consent event notifications
type ConsentWebhookHandler struct {
	DB    *database.DB
	Cache *cache.CacheWithRegistry
}

// ConsentEventRequest represents the incoming webhook payload for consent events
type ConsentEventRequest struct {
	MembersID  string    `json:"members_id" binding:"required"`
	ConsentKey string    `json:"consent_key" binding:"required"`
	Action     string    `json:"action" binding:"required,oneof=ACCEPTED REJECTED"`
	Timestamp  time.Time `json:"timestamp" binding:"required"`
}

// ConsentEventResponse represents the response for a consent event
type ConsentEventResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"` // "created" or "pending"
}

// HandleConsentEvent handles POST requests for external consent events
func (h *ConsentWebhookHandler) HandleConsentEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract API key source from middleware context
	source, ok := c.Get("api_key_source")
	if !ok {
		slog.Error("consent_webhook: missing api_key_source in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	sourceStr := source.(string)

	// Parse request body
	var req ConsentEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("consent_webhook: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate members_id as UUID
	if _, err := uuid.Parse(req.MembersID); err != nil {
		slog.Warn("consent_webhook: invalid members_id UUID", "members_id", req.MembersID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid members_id format, must be a valid UUID"})
		return
	}

	slog.Info("consent_webhook: processing consent event",
		"members_id", req.MembersID,
		"consent_key", req.ConsentKey,
		"action", req.Action,
		"source", sourceStr,
	)

	// Try to get user by members_id
	user, err := h.DB.Queries.GetUserByMembersID(ctx, req.MembersID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// User not found - store as pending consent event
			h.storePendingConsentEvent(c, req, sourceStr)
			return
		}
		slog.Error("consent_webhook: failed to get user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// User exists - get consent and create history record
	consent, err := h.DB.Queries.GetLatestPublishedConsentByKey(ctx, req.ConsentKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("consent_webhook: consent not found",
				"consent_key", req.ConsentKey,
				"source", sourceStr,
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error":       "consent not found",
				"consent_key": req.ConsentKey,
			})
			return
		}
		slog.Error("consent_webhook: failed to get consent", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Create consent history record
	historyID := ulid.NewUserConsentHistoryID()

	var occurredAt pgtype.Timestamptz
	if err := occurredAt.Scan(req.Timestamp); err != nil {
		slog.Error("consent_webhook: failed to convert timestamp", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	_, err = h.DB.Queries.CreateUserConsentHistory(ctx, sqlc.CreateUserConsentHistoryParams{
		ID:         historyID,
		UserID:     user.ID,
		ConsentID:  consent.ID,
		ConsentKey: req.ConsentKey,
		Action:     req.Action,
		OccurredAt: occurredAt,
		Source:     &sourceStr,
	})
	if err != nil {
		slog.Error("consent_webhook: failed to create consent history",
			"error", err,
			"user_id", user.ID,
			"consent_key", req.ConsentKey,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record consent event"})
		return
	}

	// Invalidate user consent cache
	h.Cache.Delete(cache.UserConsentsKey(user.ID))

	slog.Info("consent_webhook: consent event created successfully",
		"history_id", historyID,
		"user_id", user.ID,
		"consent_key", req.ConsentKey,
		"action", req.Action,
		"source", sourceStr,
	)

	c.JSON(http.StatusCreated, ConsentEventResponse{
		ID:     historyID,
		Status: "created",
	})
}

// storePendingConsentEvent stores a consent event for a user that doesn't exist yet
func (h *ConsentWebhookHandler) storePendingConsentEvent(c *gin.Context, req ConsentEventRequest, source string) {
	ctx := c.Request.Context()

	pendingID := ulid.NewPendingConsentEventID()

	var occurredAt pgtype.Timestamptz
	if err := occurredAt.Scan(req.Timestamp); err != nil {
		slog.Error("consent_webhook: failed to convert timestamp for pending event", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	_, err := h.DB.Queries.CreatePendingConsentEvent(ctx, sqlc.CreatePendingConsentEventParams{
		ID:         pendingID,
		MembersID:  req.MembersID,
		ConsentKey: req.ConsentKey,
		Action:     req.Action,
		OccurredAt: occurredAt,
		Source:     &source,
	})
	if err != nil {
		slog.Error("consent_webhook: failed to create pending consent event",
			"error", err,
			"members_id", req.MembersID,
			"consent_key", req.ConsentKey,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record consent event"})
		return
	}

	slog.Info("consent_webhook: pending consent event created (user not yet registered)",
		"pending_id", pendingID,
		"members_id", req.MembersID,
		"consent_key", req.ConsentKey,
		"action", req.Action,
		"source", source,
	)

	c.JSON(http.StatusCreated, ConsentEventResponse{
		ID:     pendingID,
		Status: "pending",
	})
}
