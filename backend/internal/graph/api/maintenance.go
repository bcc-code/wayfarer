package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services"
)

// previewMissingContentProgress retrieves a preview of users with missing content progress records.
func (r *Resolver) previewMissingContentProgress(ctx context.Context, first *int, after *string) (*model.MissingContentProgressPreview, error) {
	limit := 50
	if first != nil && *first > 0 && *first <= 100 {
		limit = *first
	}

	offset := 0
	// After cursor is just the offset as a string for simplicity
	if after != nil && *after != "" {
		if _, err := fmt.Sscanf(*after, "%d", &offset); err == nil {
			offset++
		}
	}

	// Get preview rows
	rows, err := r.DB.Queries.GetMissingContentProgressPreview(ctx, sqlc.GetMissingContentProgressPreviewParams{
		Querylimit:  int32(limit),
		Queryoffset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get missing content progress preview: %w", err)
	}

	// Get counts
	totalUsers, err := r.DB.Queries.CountMissingContentProgressUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count affected users: %w", err)
	}

	totalEvents, err := r.DB.Queries.CountMissingContentProgressEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count affected events: %w", err)
	}

	// Build affected users list with User objects loaded via DataLoader
	affectedUsers := make([]model.MissingContentProgressUser, 0, len(rows))
	for _, row := range rows {
		userThunk := r.Loaders.UserByIDLoader.Load(ctx, row.UserID)
		user, err := userThunk()
		if err != nil {
			continue
		}

		affectedUsers = append(affectedUsers, model.MissingContentProgressUser{
			User:       user,
			EventCount: int(row.EventCount),
		})
	}

	return &model.MissingContentProgressPreview{
		AffectedUsers: affectedUsers,
		TotalUsers:    int(totalUsers),
		TotalEvents:   int(totalEvents),
	}, nil
}

// fixMissingContentProgress processes missing content events using the existing ContentAchievementService.
// This ensures all hooks (cache invalidation, webhooks, push notifications, Firebase, score journals) are triggered.
func (r *Resolver) fixMissingContentProgress(ctx context.Context) (*model.FixMissingContentProgressResult, error) {
	// Get all missing events with user and task info
	events, err := r.DB.Queries.GetMissingContentEventsForProcessing(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing content events: %w", err)
	}

	if len(events) == 0 {
		return &model.FixMissingContentProgressResult{
			UsersFixed:             0,
			ProgressRecordsCreated: 0,
			AchievementsAwarded:    0,
		}, nil
	}

	// Count unique users
	userSet := make(map[string]bool)
	for _, event := range events {
		userSet[event.UserID] = true
	}

	slog.Info("maintenance: processing missing content events",
		"event_count", len(events),
		"user_count", len(userSet))

	// Create ContentAchievementService to handle processing
	// This handles: progress records, cache invalidation, achievement awards,
	// score journals, push notifications, and Firebase notifications
	contentAchievementService := &services.ContentAchievementService{
		DB:             r.DB,
		Cache:          r.Cache,
		PushService:    r.PushService,
		Loaders:        r.Loaders,
		WebhookService: r.WebhookService,
	}

	// Process each event using the existing service
	for _, event := range events {
		contentAchievementService.ProcessContentEvent(ctx, event.UserID, event.TaskID)
	}

	// Notify Firebase for content updates for each affected user
	for userID := range userSet {
		go r.FirebaseService.NotifyUserContent(context.Background(), userID)
	}

	return &model.FixMissingContentProgressResult{
		UsersFixed:             len(userSet),
		ProgressRecordsCreated: len(events),
		AchievementsAwarded:    0, // Tracked internally by the service via logs
	}, nil
}
