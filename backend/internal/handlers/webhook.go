package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type WebhookHandler struct {
	DB                        *database.DB
	WebhookService            *webhooks.Service
	ContentAchievementService *services.ContentAchievementService
}

// ContentEventRequest represents the incoming webhook payload for content events
type ContentEventRequest struct {
	PersonID        string    `json:"person_id" binding:"required"`
	TaskID          string    `json:"task_id" binding:"required"`
	PlanID          *string   `json:"plan_id"`
	Timestamp       time.Time `json:"timestamp" binding:"required"`
	ContentProgress *float64  `json:"content_progress"`
	Force           bool      `json:"force"`
}

// HandleContentEvent handles POST requests for external content completion events
func (h *WebhookHandler) HandleContentEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract API key source from middleware context
	source, ok := c.Get("api_key_source")
	if !ok {
		slog.Error("webhook: missing api_key_source in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	sourceStr := source.(string)

	// Parse request body
	var req ContentEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("webhook: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Validate and parse person_id as UUID
	if _, err := uuid.Parse(req.PersonID); err != nil {
		slog.Warn("webhook: invalid person_id UUID", "person_id", req.PersonID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person_id format, must be a valid UUID"})
		return
	}

	// Convert to pgtype.UUID
	var personPgUUID pgtype.UUID
	if err := personPgUUID.Scan(req.PersonID); err != nil {
		slog.Error("webhook: failed to convert UUID", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	planID := ""
	if req.PlanID != nil {
		planID = *req.PlanID
	}

	slog.Info("webhook: processing content event",
		"person_id", req.PersonID,
		"task_id", req.TaskID,
		"source", sourceStr,
	)

	err := h.ContentAchievementService.StoreAndProcessContentEvent(
		ctx,
		"",
		personPgUUID,
		req.TaskID,
		planID,
		req.ContentProgress,
		req.Timestamp,
		sourceStr,
		req.Force,
	)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateContentEvent) {
			c.JSON(http.StatusOK, gin.H{"status": "duplicate, skipped"})
			return
		}
		slog.Error("webhook: failed to process content event",
			"error", err,
			"person_id", req.PersonID,
			"task_id", req.TaskID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process event"})
		return
	}

	c.Status(http.StatusCreated)
}
