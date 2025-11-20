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

// eventByIDBatchFunc batches loading events by IDs
func eventByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Event] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Event] {
		// Check cache first for each ID
		eventMap := make(map[string]*model.Event)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.EventKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if event, ok := cached.(*model.Event); ok {
					eventMap[id] = event
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetEventsByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Event], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Event]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				event := &model.Event{
					ID:          row.ID,
					ProjectID:   row.ProjectID,
					Name:        row.Name,
					Description: row.Description,
					StartDate:   scalars.DateTime{Time: row.StartDate.Time},
					EndDate:     scalars.DateTime{Time: row.EndDate.Time},
				}

				eventMap[row.ID] = event
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.EventKey(row.ID), event)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.Event], len(ids))
		for i, id := range ids {
			if event, ok := eventMap[id]; ok {
				results[i] = &dataloader.Result[*model.Event]{Data: event}
			} else {
				results[i] = &dataloader.Result[*model.Event]{
					Error: fmt.Errorf("event not found: %s", id),
				}
			}
		}
		return results
	}
}
