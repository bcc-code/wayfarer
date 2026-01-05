package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type WebhookHandler struct {
	DB             *database.DB
	Cache          *cache.CacheWithRegistry
	PushService    *push.Service
	Loaders        *loaders.Loaders
	WebhookService *webhooks.Service
}

// ContentEventRequest represents the incoming webhook payload for content events
type ContentEventRequest struct {
	PersonID        string    `json:"person_id" binding:"required"`
	TaskID          string    `json:"task_id" binding:"required"`
	PlanID          *string   `json:"plan_id"`
	Timestamp       time.Time `json:"timestamp" binding:"required"`
	ContentProgress *float64  `json:"content_progress"`
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
		"task_id", req.TaskID,
		"source", sourceStr,
	)

	// Convert receivedAt to pgtype.Timestamptz
	var receivedAtPg pgtype.Timestamptz
	if err := receivedAtPg.Scan(receivedAt); err != nil {
		slog.Error("webhook: failed to convert timestamp", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Prepare plan_id (nullable)
	planID := ""
	if req.PlanID != nil {
		planID = *req.PlanID
	}

	// Convert consumed_at timestamp to pgtype.Timestamptz
	var consumedAtPg pgtype.Timestamptz
	if err := consumedAtPg.Scan(req.Timestamp); err != nil {
		slog.Error("webhook: failed to convert consumed_at timestamp", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Insert event into database
	event, err := h.DB.Queries.CreateExternalContentEvent(ctx, sqlc.CreateExternalContentEventParams{
		ID:              eventID,
		Personid:        personPgUUID,
		Taskid:          req.TaskID,
		Planid:          planID,
		Source:          sourceStr,
		Receivedat:      receivedAtPg,
		Contentprogress: contentProgress,
		Consumedat:      consumedAtPg,
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

	// Dispatch external content event webhooks to all active projects (only if user is identified)
	if h.WebhookService != nil {
		if user, err := h.DB.Queries.GetUserByPersonUUID(ctx, personPgUUID); err == nil {
			name := ""
			if user.DisplayName != nil {
				name = *user.DisplayName
			}
			userData := &webhooks.UserData{
				ID:        user.ID,
				MembersID: user.MembersID,
				Email:     user.Email,
				Name:      name,
			}

			// Get content progress as float64
			var progress float64
			if contentProgress != nil {
				progress = float64(*contentProgress)
			}

			// Get plan ID
			planID := ""
			if req.PlanID != nil {
				planID = *req.PlanID
			}

			go h.WebhookService.DispatchGlobalExternalContentEvent(context.Background(), userData, webhooks.ExternalContentEventData{
				TaskID:          req.TaskID,
				PlanID:          planID,
				ContentProgress: progress,
				ConsumedAt:      req.Timestamp.Format(time.RFC3339),
			})
		}
	}

	// Process achievements for this content event
	h.processAchievements(ctx, personPgUUID, req.TaskID)

	// Return success response with no body
	c.Status(http.StatusCreated)
}

// processAchievements handles achievement progress and auto-award logic for content events.
// It silently skips processing if user or content is not found in the database.
func (h *WebhookHandler) processAchievements(ctx context.Context, personUUID pgtype.UUID, taskID string) {
	// 1. Get user by person_uuid
	user, err := h.DB.Queries.GetUserByPersonUUID(ctx, personUUID)
	if err != nil {
		slog.Warn("webhook: user not found for achievement processing",
			"person_uuid", personUUID, "error", err)
		return
	}

	// 2. Get external content by task_id
	content, err := h.DB.Queries.GetExternalContentByTaskID(ctx, taskID)
	if err != nil {
		slog.Warn("webhook: external content not found for achievement processing",
			"task_id", taskID, "error", err)
		return
	}

	// 3. Get all published achievements containing this content
	achievements, err := h.DB.Queries.GetPublishedContentAchievementsByExternalContent(ctx, content.ID)
	if err != nil {
		slog.Error("webhook: failed to get achievements", "error", err)
		return
	}

	if len(achievements) == 0 {
		return
	}

	// 4. Mark content completed for all achievements
	err = h.DB.Queries.MarkContentItemCompletedForAllAchievements(ctx, sqlc.MarkContentItemCompletedForAllAchievementsParams{
		UserID:            user.ID,
		ExternalContentID: content.ID,
	})
	if err != nil {
		slog.Error("webhook: failed to mark content completed", "error", err)
		return
	}

	slog.Info("webhook: marked content completed",
		"user_id", user.ID,
		"external_content_id", content.ID,
		"achievement_count", len(achievements))

	// 5. Check and award completed achievements
	for _, row := range achievements {
		h.Cache.Delete(cache.UserContentProgressKey(user.ID, row.ID))

		// Check if all items are completed
		items, err := h.DB.Queries.GetContentItemsByAchievementIDs(ctx, []string{row.ID})
		if err != nil {
			slog.Error("webhook: failed to get content items", "error", err, "achievement_id", row.ID)
			continue
		}

		progress, err := h.DB.Queries.GetUserContentProgressForAchievement(ctx, sqlc.GetUserContentProgressForAchievementParams{
			UserID:        user.ID,
			AchievementID: row.ID,
		})
		if err != nil {
			slog.Error("webhook: failed to get user progress", "error", err, "achievement_id", row.ID)
			continue
		}

		if len(progress) == len(items) {
			h.awardAchievement(ctx, user.ID, row)
		}
	}
}

// awardAchievement awards an achievement to a user, creates a score journal entry,
// and invalidates relevant caches. Only awards if not already awarded.
func (h *WebhookHandler) awardAchievement(ctx context.Context, userID string, achievement *sqlc.GetPublishedContentAchievementsByExternalContentRow) {
	// Check if user already has this achievement
	hasAchievement, err := h.DB.Queries.CheckUserHasAchievement(ctx, sqlc.CheckUserHasAchievementParams{
		UserID:        userID,
		AchievementID: achievement.ID,
	})
	if err != nil {
		slog.Error("webhook: failed to check if user has achievement", "error", err,
			"user_id", userID, "achievement_id", achievement.ID)
		return
	}

	if hasAchievement {
		slog.Debug("webhook: user already has achievement, skipping",
			"user_id", userID, "achievement_id", achievement.ID)
		return
	}

	// Ensure user is in the project (required for leaderboard)
	err = h.DB.Queries.JoinProject(ctx, sqlc.JoinProjectParams{
		Userid:    userID,
		Projectid: achievement.ProjectID,
	})
	if err != nil {
		slog.Error("webhook: failed to add user to project", "error", err,
			"user_id", userID, "project_id", achievement.ProjectID)
		// Continue anyway - achievement can still be awarded
	}

	// Award the achievement
	err = h.DB.Queries.AwardUserAchievementIdempotent(ctx, sqlc.AwardUserAchievementIdempotentParams{
		UserID:        userID,
		AchievementID: achievement.ID,
	})
	if err != nil {
		slog.Error("webhook: failed to award achievement", "error", err,
			"user_id", userID, "achievement_id", achievement.ID)
		return
	}

	// Check if score journal entry already exists for this achievement
	journalExists, err := h.DB.Queries.CheckScoreJournalEntryExists(ctx, sqlc.CheckScoreJournalEntryExistsParams{
		UserID:     userID,
		SourceType: "ACHIEVEMENT",
		SourceID:   achievement.ID,
	})
	if err != nil {
		slog.Error("webhook: failed to check score journal entry", "error", err,
			"user_id", userID, "achievement_id", achievement.ID)
	}

	// Only create score journal entry if it doesn't exist
	if !journalExists {
		journalID := ulid.NewScoreJournalID()
		_, err = h.DB.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
			ID:         journalID,
			ProjectID:  achievement.ProjectID,
			UserID:     userID,
			EventID:    achievement.EventID,
			Points:     achievement.Points,
			SourceType: "ACHIEVEMENT",
			SourceID:   &achievement.ID,
		})
		if err != nil {
			slog.Error("webhook: failed to create score journal entry", "error", err,
				"user_id", userID, "achievement_id", achievement.ID)
			// Achievement was awarded, but score journal failed - log but continue
		}
	} else {
		slog.Debug("webhook: score journal entry already exists, skipping",
			"user_id", userID, "achievement_id", achievement.ID)
	}

	// Invalidate caches
	h.Cache.InvalidateUser(userID)
	h.Cache.InvalidateAchievement(achievement.ID)
	h.Cache.InvalidateProject(achievement.ProjectID)
	if achievement.EventID != nil {
		h.Cache.InvalidateEvent(*achievement.EventID)
	}
	h.Cache.Delete(cache.UserAchievementTimestampKey(userID, achievement.ID))

	slog.Info("webhook: awarded achievement",
		"user_id", userID,
		"achievement_id", achievement.ID,
		"points", achievement.Points)

	// Send push notification in background with translated content
	if h.PushService != nil && h.Loaders != nil {
		go push.SendTranslatedAchievementNotification(h.PushService, h.Loaders, userID, push.AchievementInfo{
			ID:               achievement.ID,
			Name:             achievement.Name,
			NotificationText: achievement.NotificationText,
			ImageCompleted:   achievement.ImageCompleted,
		})
	}
}
