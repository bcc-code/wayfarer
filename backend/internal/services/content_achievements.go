package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ContentAchievementService handles processing of content-based achievements.
// It is used by both the webhook handler (for real-time events) and the auth handler
// (for processing pending events when a user registers).
type ContentAchievementService struct {
	DB             *database.DB
	Cache          *cache.CacheWithRegistry
	PushService    *push.Service
	Loaders        *loaders.Loaders
	WebhookService *webhooks.Service
}

// ProcessContentEvent processes a content event for a user, awarding achievements if applicable.
// It also dispatches webhooks if the webhook service is configured.
func (s *ContentAchievementService) ProcessContentEvent(ctx context.Context, userID string, taskID string) {
	s.processAchievements(ctx, userID, taskID)
}

// ProcessPendingContentEvents processes all pending content events for a newly registered user.
// This is called when a user registers and there are content events that arrived before they existed.
func (s *ContentAchievementService) ProcessPendingContentEvents(ctx context.Context, userID string, personUUID pgtype.UUID) {
	events, err := s.DB.Queries.GetExternalContentEventsForProcessing(ctx, personUUID)
	if err != nil {
		slog.Error("content_achievements: failed to get external content events for processing",
			"user_id", userID,
			"error", err,
		)
		return
	}

	if len(events) == 0 {
		return
	}

	slog.Info("content_achievements: processing pending external content events for new user",
		"user_id", userID,
		"count", len(events),
	)

	// Get user data for webhook dispatch
	user, err := s.DB.Queries.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("content_achievements: failed to get user for webhook dispatch",
			"user_id", userID,
			"error", err,
		)
		return
	}

	for _, event := range events {
		// Dispatch webhook for this event
		if s.WebhookService != nil {
			userData := webhooks.NewUserData(user)

			var progress float64
			if event.ContentProgress != nil {
				progress = float64(*event.ContentProgress)
			}

			planID := ""
			if event.PlanID != nil {
				planID = *event.PlanID
			}

			var consumedAt string
			if event.ConsumedAt.Valid {
				consumedAt = event.ConsumedAt.Time.Format(time.RFC3339)
			}

			go s.WebhookService.DispatchGlobalExternalContentEvent(context.Background(), userData, webhooks.ExternalContentEventData{
				TaskID:          event.TaskID,
				PlanID:          planID,
				ContentProgress: progress,
				ConsumedAt:      consumedAt,
			})
		}

		// Process achievement for this event
		s.processAchievements(ctx, userID, event.TaskID)
	}
}

// processAchievements handles achievement progress and auto-award logic for a content event.
// It silently skips processing if content is not found in the database.
func (s *ContentAchievementService) processAchievements(ctx context.Context, userID, taskID string) {
	// Get external content by task_id
	content, err := s.DB.Queries.GetExternalContentByTaskID(ctx, taskID)
	if err != nil {
		slog.Debug("content_achievements: external content not found for achievement processing",
			"task_id", taskID, "error", err)
		return
	}

	// Get all published achievements containing this content
	achievements, err := s.DB.Queries.GetPublishedContentAchievementsByExternalContent(ctx, content.ID)
	if err != nil {
		slog.Error("content_achievements: failed to get achievements", "error", err)
		return
	}

	if len(achievements) == 0 {
		return
	}

	// Mark content completed for all achievements
	err = s.DB.Queries.MarkContentItemCompletedForAllAchievements(ctx, sqlc.MarkContentItemCompletedForAllAchievementsParams{
		UserID:            userID,
		ExternalContentID: content.ID,
	})
	if err != nil {
		slog.Error("content_achievements: failed to mark content completed", "error", err)
		return
	}

	slog.Info("content_achievements: marked content completed",
		"user_id", userID,
		"external_content_id", content.ID,
		"achievement_count", len(achievements))

	// Check and award completed achievements
	for _, achievement := range achievements {
		if s.Cache != nil {
			s.Cache.Delete(cache.UserContentProgressKey(userID, achievement.ID))
		}

		// Check if all items are completed
		items, err := s.DB.Queries.GetContentItemsByAchievementIDs(ctx, []string{achievement.ID})
		if err != nil {
			slog.Error("content_achievements: failed to get content items", "error", err, "achievement_id", achievement.ID)
			continue
		}

		progress, err := s.DB.Queries.GetUserContentProgressForAchievement(ctx, sqlc.GetUserContentProgressForAchievementParams{
			UserID:        userID,
			AchievementID: achievement.ID,
		})
		if err != nil {
			slog.Error("content_achievements: failed to get user progress", "error", err, "achievement_id", achievement.ID)
			continue
		}

		if len(progress) == len(items) {
			s.awardAchievement(ctx, userID, achievement)
		}
	}
}

// awardAchievement awards an achievement to a user, creates a score journal entry,
// and invalidates relevant caches. Only awards if not already awarded.
func (s *ContentAchievementService) awardAchievement(ctx context.Context, userID string, achievement *sqlc.GetPublishedContentAchievementsByExternalContentRow) {
	// Check if achievement is awardable based on awardable_from timestamp
	// Silently skip if not yet available (progress is still tracked, but award is delayed)
	if achievement.AwardableFrom.Valid && achievement.AwardableFrom.Time.After(time.Now()) {
		slog.Debug("content_achievements: achievement is not yet available for awarding, skipping",
			"user_id", userID, "achievement_id", achievement.ID, "awardable_from", achievement.AwardableFrom.Time)
		return
	}

	// Check if user already has this achievement
	hasAchievement, err := s.DB.Queries.CheckUserHasAchievement(ctx, sqlc.CheckUserHasAchievementParams{
		UserID:        userID,
		AchievementID: achievement.ID,
	})
	if err != nil {
		slog.Error("content_achievements: failed to check if user has achievement", "error", err,
			"user_id", userID, "achievement_id", achievement.ID)
		return
	}

	if hasAchievement {
		slog.Debug("content_achievements: user already has achievement, skipping",
			"user_id", userID, "achievement_id", achievement.ID)
		return
	}

	// Ensure user is in the project (required for leaderboard)
	err = s.DB.Queries.JoinProject(ctx, sqlc.JoinProjectParams{
		Userid:    userID,
		Projectid: achievement.ProjectID,
	})
	if err != nil {
		slog.Error("content_achievements: failed to add user to project", "error", err,
			"user_id", userID, "project_id", achievement.ProjectID)
		// Continue anyway - achievement can still be awarded
	}

	// Award the achievement
	err = s.DB.Queries.AwardUserAchievementIdempotent(ctx, sqlc.AwardUserAchievementIdempotentParams{
		UserID:        userID,
		AchievementID: achievement.ID,
	})
	if err != nil {
		slog.Error("content_achievements: failed to award achievement", "error", err,
			"user_id", userID, "achievement_id", achievement.ID)
		return
	}

	// Check if score journal entry already exists for this achievement
	journalExists, err := s.DB.Queries.CheckScoreJournalEntryExists(ctx, sqlc.CheckScoreJournalEntryExistsParams{
		UserID:     userID,
		SourceType: "ACHIEVEMENT",
		SourceID:   achievement.ID,
	})
	if err != nil {
		slog.Error("content_achievements: failed to check score journal entry", "error", err,
			"user_id", userID, "achievement_id", achievement.ID)
	}

	// Only create score journal entry if it doesn't exist
	if !journalExists {
		journalID := ulid.NewScoreJournalID()
		_, err = s.DB.Queries.CreateScoreJournalEntry(ctx, sqlc.CreateScoreJournalEntryParams{
			ID:         journalID,
			ProjectID:  achievement.ProjectID,
			UserID:     userID,
			EventID:    achievement.EventID,
			Points:     achievement.Points,
			SourceType: "ACHIEVEMENT",
			SourceID:   &achievement.ID,
		})
		if err != nil {
			slog.Error("content_achievements: failed to create score journal entry", "error", err,
				"user_id", userID, "achievement_id", achievement.ID)
			// Achievement was awarded, but score journal failed - log but continue
		}
	} else {
		slog.Debug("content_achievements: score journal entry already exists, skipping",
			"user_id", userID, "achievement_id", achievement.ID)
	}

	// Invalidate caches
	if s.Cache != nil {
		s.Cache.InvalidateUser(userID)
		s.Cache.InvalidateAchievement(achievement.ID)
		s.Cache.InvalidateProject(achievement.ProjectID)
		if achievement.EventID != nil {
			s.Cache.InvalidateEvent(*achievement.EventID)
		}
		s.Cache.Delete(cache.UserAchievementTimestampKey(userID, achievement.ID))
	}

	slog.Info("content_achievements: awarded achievement",
		"user_id", userID,
		"achievement_id", achievement.ID,
		"points", achievement.Points)

	// Send push notification in background with translated content
	if s.PushService != nil && s.Loaders != nil {
		go push.SendTranslatedAchievementNotification(s.PushService, s.Loaders, userID, push.AchievementInfo{
			ID:               achievement.ID,
			Name:             achievement.Name,
			NotificationText: achievement.NotificationText,
			ImageCompleted:   achievement.ImageCompleted,
		})
	}
}
