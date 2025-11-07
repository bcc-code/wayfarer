package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// challengesByEventBatchFunc batches loading challenges by event IDs
func challengesByEventBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Challenge] {
	return func(ctx context.Context, eventIDs []string) []*dataloader.Result[[]*model.Challenge] {
		// Check cache first for each event ID
		challengesByEvent := make(map[string][]*model.Challenge)
		missingEventIDs := []string{}

		for _, eventID := range eventIDs {
			cacheKey := cache.ChallengesByEventKey(eventID)
			if cached, ok := c.Get(cacheKey); ok {
				if challenges, ok := cached.([]*model.Challenge); ok {
					challengesByEvent[eventID] = challenges
					continue
				}
			}
			missingEventIDs = append(missingEventIDs, eventID)
		}

		// Query database only for cache misses
		if len(missingEventIDs) > 0 {
			rows, err := db.Queries.GetChallengesByEventIDs(ctx, missingEventIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Challenge], len(eventIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Challenge]{Error: err}
				}
				return results
			}

			// Group challenges by event ID and convert to GraphQL model
			for _, row := range rows {
				var endTime *scalars.DateTime
				if row.EndTime.Valid {
					endTime = &scalars.DateTime{Time: row.EndTime.Time}
				}

				challenge := &model.Challenge{
					ID:          row.ID,
					Name:        row.Name,
					Description: scalars.HTML(row.Description),
					Image:       row.ImageUrl,
					ProjectID:   row.ProjectID,
					EventID:     row.EventID,
					URL:         row.Url,
					ButtonText:  row.ButtonText,
					PublishedAt: scalars.DateTime{Time: row.PublishedAt.Time},
					EndTime:     endTime,
				}
				challengesByEvent[*row.EventID] = append(challengesByEvent[*row.EventID], challenge)
			}

			// Populate cache for each event, including empty results
			for _, eventID := range missingEventIDs {
				challenges := challengesByEvent[eventID]
				if challenges == nil {
					challenges = []*model.Challenge{} // Empty slice, not nil
				}
				challengesByEvent[eventID] = challenges
				c.Set(cache.ChallengesByEventKey(eventID), challenges)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Challenge], len(eventIDs))
		for i, eventID := range eventIDs {
			challenges := challengesByEvent[eventID]
			if challenges == nil {
				challenges = []*model.Challenge{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Challenge]{Data: challenges}
		}
		return results
	}
}
