package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

func (r *Resolver) adminCheckAchievementProgress(ctx context.Context, targetUserID string, achievementID string) (*model.AdminAchievementProgress, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user not authenticated")
	}
	if !r.RoleService.IsSuperAdmin(ctx, currentUserID) {
		return nil, fmt.Errorf("permission denied: superadmin role required")
	}

	// Load achievement
	thunk := r.Loaders.AchievementByIDLoader.Load(ctx, achievementID)
	achievement, err := thunk()
	if err != nil {
		return nil, fmt.Errorf("failed to load achievement: %w", err)
	}

	// Determine achievement type
	var isContent, isStreak bool
	switch achievement.(type) {
	case *model.ContentAchievement:
		isContent = true
	case *model.StreakAchievement:
		isStreak = true
	default:
		return nil, fmt.Errorf("achievement must be a content or streak achievement")
	}

	// Check if user already has the achievement
	hasAchievement, err := r.DB.Queries.CheckUserHasAchievement(ctx, sqlc.CheckUserHasAchievementParams{
		UserID:        targetUserID,
		AchievementID: achievementID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check achievement status: %w", err)
	}

	var awardedAt *scalars.DateTime
	if hasAchievement {
		timestamps, err := r.DB.Queries.GetUserAchievementTimestamps(ctx, sqlc.GetUserAchievementTimestampsParams{
			Userid:         targetUserID,
			AchievementIds: []string{achievementID},
		})
		if err == nil && len(timestamps) > 0 && timestamps[0].AchievedAt.Valid {
			awardedAt = &scalars.DateTime{Time: timestamps[0].AchievedAt.Time}
		}
	}

	// Build progress items based on achievement type
	var items []model.AdminAchievementItemProgress

	if isContent {
		items, err = r.buildContentProgressItems(ctx, targetUserID, achievementID)
	} else if isStreak {
		items, err = r.buildStreakProgressItems(ctx, targetUserID, achievementID)
	}
	if err != nil {
		return nil, err
	}

	completedCount := 0
	for _, item := range items {
		if item.Completed {
			completedCount++
		}
	}

	return &model.AdminAchievementProgress{
		Achievement:    achievement,
		AlreadyAwarded: hasAchievement,
		AwardedAt:      awardedAt,
		Items:          items,
		CompletedCount: completedCount,
		TotalCount:     len(items),
	}, nil
}

func (r *Resolver) buildContentProgressItems(ctx context.Context, targetUserID string, achievementID string) ([]model.AdminAchievementItemProgress, error) {
	// Get items with deadlines
	dbItems, err := r.DB.Queries.GetContentItemsWithDeadlines(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content items: %w", err)
	}

	// Get user progress
	progress, err := r.DB.Queries.GetUserContentProgress(ctx, sqlc.GetUserContentProgressParams{
		UserID:         targetUserID,
		AchievementIds: []string{achievementID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user content progress: %w", err)
	}

	// Build progress map keyed by external_content_id
	progressMap := make(map[string]*sqlc.UserContentProgress)
	for _, p := range progress {
		progressMap[p.ExternalContentID] = p
	}

	items := make([]model.AdminAchievementItemProgress, 0, len(dbItems))
	for _, dbItem := range dbItems {
		item := model.AdminAchievementItemProgress{
			ContentItem: &model.ContentItem{
				ID:                dbItem.ID,
				ExternalContentID: dbItem.ExternalContentID,
				SortOrder:         int(dbItem.SortOrder),
			},
		}

		// Row in user_content_progress = completed
		if p, ok := progressMap[dbItem.ExternalContentID]; ok {
			item.Completed = true
			if p.CompletedAt.Valid {
				item.CompletedAt = &scalars.DateTime{Time: p.CompletedAt.Time}
			}
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *Resolver) buildStreakProgressItems(ctx context.Context, targetUserID string, achievementID string) ([]model.AdminAchievementItemProgress, error) {
	// Get items with deadlines
	dbItems, err := r.DB.Queries.GetStreakItemsWithDeadlines(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get streak items: %w", err)
	}

	// Get user progress
	progress, err := r.DB.Queries.GetUserStreakProgress(ctx, sqlc.GetUserStreakProgressParams{
		UserID:         targetUserID,
		AchievementIds: []string{achievementID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user streak progress: %w", err)
	}

	// Build progress map keyed by external_content_id
	progressMap := make(map[string]*sqlc.UserStreakProgress)
	for _, p := range progress {
		progressMap[p.ExternalContentID] = p
	}

	items := make([]model.AdminAchievementItemProgress, 0, len(dbItems))
	for _, dbItem := range dbItems {
		item := model.AdminAchievementItemProgress{
			ContentItem: &model.ContentItem{
				ID:                dbItem.ID,
				ExternalContentID: dbItem.ExternalContentID,
				SortOrder:         int(dbItem.SortOrder),
			},
		}

		// Set deadline if present
		if dbItem.CompleteBy.Valid {
			item.CompleteBy = &scalars.DateTime{Time: dbItem.CompleteBy.Time}
		}

		// Row in user_streak_progress = completed within deadline
		// (the system enforces deadlines at write time)
		if p, ok := progressMap[dbItem.ExternalContentID]; ok {
			item.Completed = true
			withinDeadline := true
			item.CompletedWithinDeadline = &withinDeadline
			if p.CompletedAt.Valid {
				item.CompletedAt = &scalars.DateTime{Time: p.CompletedAt.Time}
			}
		} else if dbItem.CompleteBy.Valid && time.Now().After(dbItem.CompleteBy.Time) {
			tooLate := false
			item.CompletedWithinDeadline = &tooLate
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *Resolver) adminExternalContentEvents(ctx context.Context, targetUserID string, externalContentID string) ([]model.AdminExternalContentEvent, error) {
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user not authenticated")
	}
	if !r.RoleService.IsSuperAdmin(ctx, currentUserID) {
		return nil, fmt.Errorf("permission denied: superadmin role required")
	}

	// Get user's person_uuid
	user, err := r.DB.Queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	// Get external content's task_id
	externalContent, err := r.DB.Queries.GetExternalContentByID(ctx, externalContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load external content: %w", err)
	}

	// Query events
	events, err := r.DB.Queries.GetExternalContentEventsByPersonAndTaskID(ctx, sqlc.GetExternalContentEventsByPersonAndTaskIDParams{
		Personid: user.PersonUuid,
		Taskid:   externalContent.TaskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get content events: %w", err)
	}

	result := make([]model.AdminExternalContentEvent, 0, len(events))
	for _, e := range events {
		event := model.AdminExternalContentEvent{
			ID:         e.ID,
			TaskID:     e.TaskID,
			Source:     e.Source,
			ReceivedAt: scalars.DateTime{Time: e.ReceivedAt.Time},
		}

		if e.PlanID != nil {
			event.PlanID = *e.PlanID
		}

		if e.ConsumedAt.Valid {
			event.ConsumedAt = &scalars.DateTime{Time: e.ConsumedAt.Time}
		}

		if e.ContentProgress != nil {
			cp := float64(*e.ContentProgress)
			event.ContentProgress = &cp
		}

		result = append(result, event)
	}

	return result, nil
}
