package loaders

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// churchBatchFunc batches church loading by IDs
func churchBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Church] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Church] {
		// Check cache first for each ID
		churchMap := make(map[string]*model.Church)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.ChurchKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if church, ok := cached.(*model.Church); ok {
					churchMap[id] = church
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetChurchesByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Church], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Church]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				church := &model.Church{
					ID:       row.ID,
					Name:     row.Name,
					Country:  row.Country,
					Category: model.ChurchCategory(row.Category),
				}

				churchMap[row.ID] = church
				// Churches rarely change, use longer TTL (30 minutes)
				c.SetWithTTL(cache.ChurchKey(row.ID), church, 30*time.Minute)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Church], len(ids))
		for i, id := range ids {
			if church, ok := churchMap[id]; ok {
				results[i] = &dataloader.Result[*model.Church]{Data: church}
			} else {
				results[i] = &dataloader.Result[*model.Church]{
					Error: fmt.Errorf("church not found: %s", id),
				}
			}
		}
		return results
	}
}
