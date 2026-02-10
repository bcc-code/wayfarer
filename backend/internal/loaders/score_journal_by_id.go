package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// scoreJournalByIDBatchFunc batches loading score journal entries by IDs
func scoreJournalByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.ScoreJournal] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.ScoreJournal] {
		// Check cache first for each ID
		journalMap := make(map[string]*model.ScoreJournal)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ScoreJournalKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if journal, ok := cached.(*model.ScoreJournal); ok {
					journalMap[id] = journal
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetScoreJournalByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.ScoreJournal], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.ScoreJournal]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				journal := &model.ScoreJournal{
					ID:          row.ID,
					Points:      int(row.Points),
					SourceType:  convertSourceType(row.SourceType),
					Reason:      row.Reason,
					CreatedAt:   scalars.DateTime{Time: row.CreatedAt.Time},
					ProjectID:   row.ProjectID,
					EventID:     row.EventID,
					ChallengeID: row.ChallengeID,
					AwardedByID: row.AwardedBy,
				}

				journalMap[row.ID] = journal
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.ScoreJournalKey(row.ID), journal)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.ScoreJournal], len(ids))
		for i, id := range ids {
			if journal, ok := journalMap[id]; ok {
				results[i] = &dataloader.Result[*model.ScoreJournal]{Data: journal}
			} else {
				results[i] = &dataloader.Result[*model.ScoreJournal]{
					Error: fmt.Errorf("score journal entry not found: %s", id),
				}
			}
		}
		return results
	}
}

// convertSourceType converts database source type string to GraphQL enum
func convertSourceType(sourceType string) model.ScoreSourceType {
	switch sourceType {
	case "ACHIEVEMENT":
		return model.ScoreSourceTypeAchievement
	case "CHALLENGE":
		return model.ScoreSourceTypeChallenge
	case "EVENT":
		return model.ScoreSourceTypeEvent
	case "MANUAL":
		return model.ScoreSourceTypeManual
	case "BET":
		return model.ScoreSourceTypeBet
	default:
		return model.ScoreSourceTypeManual
	}
}
