package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
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

// previewMissingScoreJournal retrieves a preview of users with missing score journal entries for a content achievement.
func (r *Resolver) previewMissingScoreJournal(ctx context.Context, achievementID string, first *int, after *string) (*model.MissingScoreJournalPreview, error) {
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
	rows, err := r.DB.Queries.GetMissingScoreJournalPreview(ctx, sqlc.GetMissingScoreJournalPreviewParams{
		Achievementid: achievementID,
		Querylimit:    int32(limit),
		Queryoffset:   int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get missing score journal preview: %w", err)
	}

	// Get counts
	totalUsers, err := r.DB.Queries.CountMissingScoreJournalUsers(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to count affected users: %w", err)
	}

	totalEvents, err := r.DB.Queries.CountMissingScoreJournalEvents(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to count affected events: %w", err)
	}

	// Build affected users list with User objects loaded via DataLoader
	affectedUsers := make([]model.MissingScoreJournalUser, 0, len(rows))
	for _, row := range rows {
		userThunk := r.Loaders.UserByIDLoader.Load(ctx, row.UserID)
		user, err := userThunk()
		if err != nil {
			continue
		}

		affectedUsers = append(affectedUsers, model.MissingScoreJournalUser{
			User:       user,
			EventCount: int(row.EventCount),
		})
	}

	return &model.MissingScoreJournalPreview{
		AffectedUsers: affectedUsers,
		TotalUsers:    int(totalUsers),
		TotalEvents:   int(totalEvents),
	}, nil
}
