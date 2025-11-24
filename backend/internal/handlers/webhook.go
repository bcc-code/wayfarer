package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type WebhookHandler struct {
	DB *database.DB
}

// ContentEventRequest represents the incoming webhook payload for content events
type ContentEventRequest struct {
	PersonID        string   `json:"person_id" binding:"required"`
	ContentID       string   `json:"content_id" binding:"required"`
	ReadingPlanID   *string  `json:"reading_plan_id"`
	ContentProgress *float64 `json:"content_progress"`
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

	// Validate content_progress if provided
	var contentProgress *float32
	if req.ContentProgress != nil {
		progress := *req.ContentProgress
		// Invalid values (< 0.01 or > 1.1) result in NULL
		if progress >= 0.01 && progress <= 1.1 {
			// Valid value - convert to float32 for database
			progressFloat32 := float32(progress)
			contentProgress = &progressFloat32
		}
		// If invalid, contentProgress remains nil (NULL in database)
	}

	// Generate ULID for the event
	eventID := ulid.NewContentEventID()
	receivedAt := time.Now()

	slog.Info("webhook: creating content event",
		"event_id", eventID,
		"person_id", req.PersonID,
		"content_id", req.ContentID,
		"source", sourceStr,
	)

	// Convert receivedAt to pgtype.Timestamptz
	var receivedAtPg pgtype.Timestamptz
	if err := receivedAtPg.Scan(receivedAt); err != nil {
		slog.Error("webhook: failed to convert timestamp", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Prepare reading_plan_id (nullable)
	readingPlanID := ""
	if req.ReadingPlanID != nil {
		readingPlanID = *req.ReadingPlanID
	}

	// Insert event into database
	event, err := h.DB.Queries.CreateExternalContentEvent(ctx, sqlc.CreateExternalContentEventParams{
		ID:              eventID,
		Personid:        personPgUUID,
		Contentid:       req.ContentID,
		Readingplanid:   readingPlanID,
		Source:          sourceStr,
		Receivedat:      receivedAtPg,
		Contentprogress: contentProgress,
	})
	if err != nil {
		slog.Error("webhook: failed to create content event",
			"error", err,
			"event_id", eventID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	slog.Info("webhook: content event created successfully",
		"event_id", event.ID,
		"person_id", req.PersonID,
		"source", sourceStr,
	)

	// Return success response with no body
	c.Status(http.StatusCreated)
}
