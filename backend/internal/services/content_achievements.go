package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrDuplicateContentEvent is returned when a content event with the same person_id and task_id already exists.
var ErrDuplicateContentEvent = errors.New("content event already exists for this person and task")

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

// getUserByPersonUUID resolves a user by person UUID, checking cache first.
func (s *ContentAchievementService) getUserByPersonUUID(ctx context.Context, personUUID pgtype.UUID) (*sqlc.GetUserByPersonUUIDRow, error) {
	if !personUUID.Valid {
		return nil, fmt.Errorf("invalid person UUID")
	}

	uuidStr := uuid.UUID(personUUID.Bytes).String()
	cacheKey := cache.UserByPersonUUIDKey(uuidStr)

	if s.Cache != nil {
		if cached, ok := s.Cache.Get(cacheKey); ok {
			if user, ok := cached.(*sqlc.GetUserByPersonUUIDRow); ok {
				return user, nil
			}
		}
	}

	user, err := s.DB.Queries.GetUserByPersonUUID(ctx, personUUID)
	if err != nil {
		return nil, err
	}

	if s.Cache != nil {
		s.Cache.Set(cacheKey, user)
	}

	return user, nil
}

// ProcessContentEvent processes a content event for a user, awarding achievements if applicable.
// It also dispatches webhooks if the webhook service is configured.
func (s *ContentAchievementService) ProcessContentEvent(ctx context.Context, userID string, taskID string) {
	s.processAchievements(ctx, userID, taskID)
}

// StoreAndProcessContentEvent stores a content event in the database and processes achievements.
// This is the shared pipeline used by both the webhook handler and the sync service.
// It generates an event ID, stores the event, dispatches webhooks (async), and processes achievements.
func (s *ContentAchievementService) StoreAndProcessContentEvent(
	ctx context.Context,
	userID string,
	personUUID pgtype.UUID,
	taskID string,
	planID string,
	contentProgress *float64,
	consumedAt time.Time,
	source string,
	force bool,
) error {
	if !force {
		exists, err := s.DB.Queries.ExternalContentEventExists(ctx, sqlc.ExternalContentEventExistsParams{
			Personid: personUUID,
			Taskid:   taskID,
		})
		if err != nil {
			return fmt.Errorf("failed to check for duplicate content event: %w", err)
		}
		if exists {
			return ErrDuplicateContentEvent
		}
	}

	// Validate content_progress if provided
	var contentProgressFloat32 *float32
	if contentProgress != nil {
		progress := *contentProgress
		if progress >= 0.01 && progress <= 1.1 {
			p32 := float32(progress)
			contentProgressFloat32 = &p32
		}
	}

	// Generate ULID for the event
	eventID := ulid.NewContentEventID()
	receivedAt := time.Now()

	// Convert timestamps to pgtype
	var receivedAtPg pgtype.Timestamptz
	if err := receivedAtPg.Scan(receivedAt); err != nil {
		return fmt.Errorf("failed to convert received_at timestamp: %w", err)
	}

	var consumedAtPg pgtype.Timestamptz
	if err := consumedAtPg.Scan(consumedAt); err != nil {
		return fmt.Errorf("failed to convert consumed_at timestamp: %w", err)
	}

	// Insert event into database
	_, err := s.DB.Queries.CreateExternalContentEvent(ctx, sqlc.CreateExternalContentEventParams{
		ID:              eventID,
		Personid:        personUUID,
		Taskid:          taskID,
		Planid:          planID,
		Source:          source,
		Receivedat:      receivedAtPg,
		Contentprogress: contentProgressFloat32,
		Consumedat:      consumedAtPg,
	})
	if err != nil {
		return fmt.Errorf("failed to create content event: %w", err)
	}

	// Dispatch external content event webhooks (async, only if user is identified)
	if s.WebhookService != nil {
		if user, err := s.getUserByPersonUUID(ctx, personUUID); err == nil {
			userData := webhooks.NewUserData(user)

			var progress float64
			if contentProgress != nil {
				progress = *contentProgress
			}

			go s.WebhookService.DispatchGlobalExternalContentEvent(context.Background(), userData, webhooks.ExternalContentEventData{
				TaskID:          taskID,
				PlanID:          planID,
				ContentProgress: progress,
				ConsumedAt:      consumedAt.Format(time.RFC3339),
			})
		}
	}

	// Process achievements
	if userID != "" {
		s.ProcessContentEvent(ctx, userID, taskID)
	} else {
		// Try to find user by person_uuid for achievement processing
		if user, err := s.getUserByPersonUUID(ctx, personUUID); err == nil {
			s.ProcessContentEvent(ctx, user.ID, taskID)
		}
	}

	return nil
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

	// Collect achievement IDs and build lookup map
	achievementIDs := make([]string, len(achievements))
	achievementsByID := make(map[string]*sqlc.GetPublishedContentAchievementsByExternalContentRow, len(achievements))
	for i, achievement := range achievements {
		achievementIDs[i] = achievement.ID
		achievementsByID[achievement.ID] = achievement
	}

	// Invalidate caches for all achievements
	if s.Cache != nil {
		for _, id := range achievementIDs {
			s.Cache.Delete(cache.UserContentProgressKey(userID, id))
		}
	}

	// Check which achievements user already has - skip those entirely
	alreadyAwarded, err := s.DB.Queries.GetUserAwardedAchievementIDs(ctx, sqlc.GetUserAwardedAchievementIDsParams{
		UserID:         userID,
		AchievementIds: achievementIDs,
	})
	if err != nil {
		slog.Error("content_achievements: failed to check awarded achievements", "error", err)
		return
	}

	// Filter out already-awarded achievements
	awardedSet := make(map[string]bool, len(alreadyAwarded))
	for _, id := range alreadyAwarded {
		awardedSet[id] = true
	}

	pendingIDs := make([]string, 0, len(achievementIDs))
	for _, id := range achievementIDs {
		if !awardedSet[id] {
			pendingIDs = append(pendingIDs, id)
		}
	}

	if len(pendingIDs) == 0 {
		return // All achievements already awarded
	}

	// Get item counts from cache, fetch missing from DB
	itemCounts := make(map[string]int32, len(pendingIDs))
	var uncachedIDs []string

	if s.Cache != nil {
		for _, id := range pendingIDs {
			if cached, ok := s.Cache.Get(cache.ContentItemCountKey(id)); ok {
				if count, ok := cached.(int32); ok {
					itemCounts[id] = count
					continue
				}
			}
			uncachedIDs = append(uncachedIDs, id)
		}
	} else {
		uncachedIDs = pendingIDs
	}

	// Fetch uncached item counts from DB
	if len(uncachedIDs) > 0 {
		dbCounts, err := s.DB.Queries.GetContentItemCounts(ctx, uncachedIDs)
		if err != nil {
			slog.Error("content_achievements: failed to get content item counts", "error", err)
			return
		}
		for _, c := range dbCounts {
			itemCounts[c.AchievementID] = c.ItemCount
			if s.Cache != nil {
				s.Cache.Set(cache.ContentItemCountKey(c.AchievementID), c.ItemCount)
			}
		}
	}

	// Get progress counts from DB (user-specific, not cached)
	progressCounts, err := s.DB.Queries.GetUserProgressCounts(ctx, sqlc.GetUserProgressCountsParams{
		UserID:         userID,
		AchievementIds: pendingIDs,
	})
	if err != nil {
		slog.Error("content_achievements: failed to get user progress counts", "error", err)
		return
	}

	progressByAchievement := make(map[string]int32, len(progressCounts))
	for _, p := range progressCounts {
		progressByAchievement[p.AchievementID] = p.ProgressCount
	}

	// Check and award completed achievements
	for _, id := range pendingIDs {
		itemCount := itemCounts[id]
		progressCount := progressByAchievement[id]

		if progressCount == itemCount && itemCount > 0 {
			achievement := achievementsByID[id]
			if achievement != nil {
				s.awardAchievement(ctx, userID, achievement)
			}
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
