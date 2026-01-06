package ladder_to_heaven

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/bcc-media/wayfarer/internal/webhook"
	"github.com/gin-gonic/gin"
)

// contentEventHandler handles webhook requests for the Ladder to Heaven plugin.
type contentEventHandler struct {
	db            *database.DB
	cache         *cache.CacheWithRegistry
	achievementID string
	secretKey     string
}

// contentEventRequest matches the outbound WebhookPayload format
type contentEventRequest struct {
	EventType string                `json:"event_type" binding:"required"`
	Timestamp time.Time             `json:"timestamp" binding:"required"`
	ProjectID string                `json:"project_id" binding:"required"`
	User      *contentEventUserData `json:"user" binding:"required"`
	Data      contentEventData      `json:"data" binding:"required"`
}

// contentEventUserData contains user information from the webhook payload
type contentEventUserData struct {
	ID        string `json:"id" binding:"required"`
	MembersID string `json:"members_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

// contentEventData contains the event-specific data from the webhook payload
type contentEventData struct {
	TaskID          string  `json:"task_id" binding:"required"`
	PlanID          string  `json:"plan_id"`
	ContentProgress float64 `json:"content_progress"`
	ConsumedAt      string  `json:"consumed_at" binding:"required"`
}

const (
	// Points awarded when content is completed before or on the deadline
	pointsOnTime = 50
	// Points awarded when content is completed after the deadline
	pointsLate = 25
	// Source type for plugin entries in score_journal
	sourceTypePlugin = "PLUGIN"
)

// handle processes incoming content event webhook requests
func (h *contentEventHandler) handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if feature is enabled
	if h.achievementID == "" {
		slog.Debug("ladder_to_heaven: feature disabled, PLUGIN_LADDER_TO_HEAVEN_ACHIEVEMENT_ID not set")
		c.Status(http.StatusOK)
		return
	}

	// Read raw body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Warn("ladder_to_heaven: failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Verify webhook signature if secret key is configured
	if h.secretKey != "" {
		signature := c.GetHeader("X-Webhook-Signature")
		if !webhook.VerifySignature(body, signature, h.secretKey) {
			slog.Warn("ladder_to_heaven: invalid webhook signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Restore body for JSON binding
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	// Parse request body
	var req contentEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("ladder_to_heaven: invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	slog.Info("ladder_to_heaven: processing request",
		"user_id", req.User.ID,
		"task_id", req.Data.TaskID,
		"achievement_id", h.achievementID,
	)

	// Get external content by task_id
	content, err := h.db.Queries.GetExternalContentByTaskID(ctx, req.Data.TaskID)
	if err != nil {
		slog.Warn("ladder_to_heaven: external content not found",
			"task_id", req.Data.TaskID, "error", err)
		c.Status(http.StatusOK)
		return
	}

	// Check if content is part of the configured achievement
	inAchievement, err := h.db.Queries.CheckContentItemInAchievement(ctx, sqlc.CheckContentItemInAchievementParams{
		AchievementID:     h.achievementID,
		ExternalContentID: content.ID,
	})
	if err != nil {
		slog.Error("ladder_to_heaven: failed to check content in achievement",
			"error", err, "content_id", content.ID, "achievement_id", h.achievementID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !inAchievement {
		slog.Debug("ladder_to_heaven: content not in configured achievement",
			"content_id", content.ID, "achievement_id", h.achievementID)
		c.Status(http.StatusOK)
		return
	}

	// Get achievement details to get project_id
	achievement, err := h.db.Queries.GetAchievementByID(ctx, h.achievementID)
	if err != nil {
		slog.Error("ladder_to_heaven: failed to get achievement",
			"error", err, "achievement_id", h.achievementID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Auto-join user to project if not already a member
	err = h.db.Queries.JoinProject(ctx, sqlc.JoinProjectParams{
		Userid:    req.User.ID,
		Projectid: achievement.ProjectID,
	})
	if err != nil {
		slog.Warn("ladder_to_heaven: failed to join user to project",
			"error", err, "user_id", req.User.ID, "project_id", achievement.ProjectID)
		// Continue anyway - score journal can still be created
	}

	// Parse consumed_at timestamp
	consumedAt, err := time.Parse(time.RFC3339, req.Data.ConsumedAt)
	if err != nil {
		slog.Warn("ladder_to_heaven: invalid consumed_at timestamp",
			"consumed_at", req.Data.ConsumedAt, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid consumed_at timestamp format"})
		return
	}

	// Calculate points based on deadline
	points := calculateDeadlinePoints(consumedAt, content.CompleteBy.Time, content.CompleteBy.Valid)

	// Check if score journal entry already exists for this user + content combination
	journalExists, err := h.db.Queries.CheckScoreJournalEntryExists(ctx, sqlc.CheckScoreJournalEntryExistsParams{
		UserID:     req.User.ID,
		SourceType: sourceTypePlugin,
		SourceID:   content.ID,
	})
	if err != nil {
		slog.Error("ladder_to_heaven: failed to check score journal entry",
			"error", err, "user_id", req.User.ID, "content_id", content.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if journalExists {
		slog.Debug("ladder_to_heaven: score journal entry already exists",
			"user_id", req.User.ID, "content_id", content.ID)
		c.Status(http.StatusOK)
		return
	}

	// Create score journal entry
	journalID := ulid.NewScoreJournalID()

	// Look up the content title in the user's language, fallback to nb
	var reason *string
	user, err := h.db.Queries.GetUserByID(ctx, req.User.ID)
	if err != nil {
		slog.Warn("ladder_to_heaven: failed to get user, using no title",
			"error", err, "user_id", req.User.ID)
	} else if user.Language != "" {
		translation, err := h.db.Queries.GetExternalContentTranslation(ctx, sqlc.GetExternalContentTranslationParams{
			Externalcontentid: content.ID,
			Languagecode:      user.Language,
		})
		if err == nil && translation.Title != nil {
			reason = translation.Title
		}
	}

	// Fallback to Norwegian Bokmal (nb) if no translation found
	if reason == nil {
		translation, err := h.db.Queries.GetExternalContentTranslation(ctx, sqlc.GetExternalContentTranslationParams{
			Externalcontentid: content.ID,
			Languagecode:      "nb",
		})
		if err == nil && translation.Title != nil {
			reason = translation.Title
		}
	}

	_, err = h.db.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
		ID:         journalID,
		ProjectID:  achievement.ProjectID,
		UserID:     req.User.ID,
		EventID:    achievement.EventID,
		Points:     int32(points),
		SourceType: sourceTypePlugin,
		SourceID:   &content.ID,
		Reason:     reason,
	})
	if err != nil {
		slog.Error("ladder_to_heaven: failed to create score journal entry",
			"error", err, "user_id", req.User.ID, "content_id", content.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create score entry"})
		return
	}

	// Invalidate caches
	h.cache.InvalidateUser(req.User.ID)
	h.cache.InvalidateProject(achievement.ProjectID)
	if achievement.EventID != nil {
		h.cache.InvalidateEvent(*achievement.EventID)
	}

	slog.Info("ladder_to_heaven: awarded points",
		"user_id", req.User.ID,
		"content_id", content.ID,
		"points", points,
		"on_time", points == pointsOnTime)

	c.Status(http.StatusCreated)
}

// calculateDeadlinePoints determines points based on whether the completion was on time
func calculateDeadlinePoints(consumedAt time.Time, completeBy time.Time, hasDeadline bool) int {
	// If no deadline set, award full points
	if !hasDeadline {
		return pointsOnTime
	}

	// If completed on or before deadline, award full points
	if !consumedAt.After(completeBy) {
		return pointsOnTime
	}

	// Completed after deadline, award reduced points
	return pointsLate
}
