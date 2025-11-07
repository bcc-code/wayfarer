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

// challengeByIDBatchFunc batches loading challenges by IDs
func challengeByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Challenge] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Challenge] {
		// Check cache first for each ID
		challengeMap := make(map[string]*model.Challenge)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ChallengeKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if challenge, ok := cached.(*model.Challenge); ok {
					challengeMap[id] = challenge
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetChallengesByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Challenge], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Challenge]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
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

				challengeMap[row.ID] = challenge
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.ChallengeKey(row.ID), challenge)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Challenge], len(ids))
		for i, id := range ids {
			if challenge, ok := challengeMap[id]; ok {
				results[i] = &dataloader.Result[*model.Challenge]{Data: challenge}
			} else {
				results[i] = &dataloader.Result[*model.Challenge]{
					Error: fmt.Errorf("challenge not found: %s", id),
				}
			}
		}
		return results
	}
}
