package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/pagination"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
)

// getScoreJournal retrieves paginated score journal entries for a user in a project.
// Used by both MyScoreJournal (user-facing) and ScoreJournal (admin-facing) resolvers.
func (r *queryResolver) getScoreJournal(ctx context.Context, projectID string, userID string, filter *model.ScoreJournalFilter, first *int, after *string, last *int, before *string) (*model.ScoreJournalConnection, error) {
	// Decode cursors if provided
	var afterCursor, beforeCursor *string
	if after != nil && *after != "" {
		decoded, err := pagination.DecodeCursor(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid after cursor: %w", err)
		}
		afterCursor = &decoded
	}
	if before != nil && *before != "" {
		decoded, err := pagination.DecodeCursor(*before)
		if err != nil {
			return nil, fmt.Errorf("invalid before cursor: %w", err)
		}
		beforeCursor = &decoded
	}

	// Determine requested limit and fetch one extra for hasMore detection
	requestedLimit := 50 // default
	if first != nil {
		requestedLimit = *first
	} else if last != nil {
		requestedLimit = *last
	}
	queryLimit := requestedLimit + 1

	// Determine pagination direction
	isBackward := last != nil

	// Build filter params
	params := sqlc.GetScoreJournalFilteredParams{
		ProjectID:  projectID,
		UserID:     userID,
		Querylimit: int32(queryLimit),
		Isbackward: isBackward,
	}

	// Set optional filters
	if filter != nil {
		if filter.EventID != nil {
			params.EventID = *filter.EventID
		}
		if filter.ChallengeID != nil {
			params.ChallengeID = *filter.ChallengeID
		}
		if filter.SourceType != nil {
			params.SourceType = string(*filter.SourceType)
		}
	}

	// Set cursor params
	if afterCursor != nil {
		params.Aftercursor = *afterCursor
	}
	if beforeCursor != nil {
		params.Beforecursor = *beforeCursor
	}

	// Get entries
	rows, err := r.DB.Queries.GetScoreJournalFiltered(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get score journal: %w", err)
	}

	// Get total count
	countParams := sqlc.CountScoreJournalFilteredParams{
		ProjectID: projectID,
		UserID:    userID,
	}
	if filter != nil {
		if filter.EventID != nil {
			countParams.EventID = *filter.EventID
		}
		if filter.ChallengeID != nil {
			countParams.ChallengeID = *filter.ChallengeID
		}
		if filter.SourceType != nil {
			countParams.SourceType = string(*filter.SourceType)
		}
	}

	totalCount, err := r.DB.Queries.CountScoreJournalFiltered(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count score journal: %w", err)
	}

	// Check if there are more results
	hasMore := len(rows) > requestedLimit
	entries := rows
	if hasMore {
		// Trim the extra record we fetched for hasMore detection
		entries = rows[:requestedLimit]
	}

	// If backward pagination, reverse the results
	if last != nil {
		for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
			entries[i], entries[j] = entries[j], entries[i]
		}
	}

	// Convert to GraphQL model
	modelEntries := make([]*model.ScoreJournal, len(entries))
	for i, entry := range entries {
		// Determine source type
		sourceType := model.ScoreSourceTypeManual
		if entry.SourceType == "ACHIEVEMENT" {
			sourceType = model.ScoreSourceTypeAchievement
		}

		modelEntries[i] = &model.ScoreJournal{
			ID:          entry.ID,
			ProjectID:   entry.ProjectID,
			UserID:      entry.UserID,
			EventID:     entry.EventID,
			ChallengeID: entry.ChallengeID,
			Points:      int(entry.Points),
			SourceType:  sourceType,
			SourceID:    entry.SourceID,
			Reason:      entry.Reason,
			AwardedByID: entry.AwardedBy,
			CreatedAt:   scalars.DateTime{Time: entry.CreatedAt.Time},
		}
	}

	// Build the connection
	connection := pagination.BuildScoreJournalConnection(pagination.BuildScoreJournalConnectionParams{
		ScoreJournals:   modelEntries,
		RequestedFirst:  first,
		RequestedLast:   last,
		RequestedAfter:  after,
		RequestedBefore: before,
		TotalCount:      int(totalCount),
		HasMore:         hasMore,
	})

	return connection, nil
}
