package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// eventsByUserBatchFunc batches loading events by user IDs
func eventsByUserBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Event] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.Event] {
		// Check cache first for each user ID
		eventsByUser := make(map[string][]*model.Event)
		missingUserIDs := []string{}

		for _, userID := range userIDs {
			cacheKey := cache.EventsByUserKey(userID)
			if cached, ok := c.Get(cacheKey); ok {
				if events, ok := cached.([]*model.Event); ok {
					eventsByUser[userID] = events
					continue
				}
			}
			missingUserIDs = append(missingUserIDs, userID)
		}

		// Query database only for cache misses
		if len(missingUserIDs) > 0 {
			rows, err := db.Queries.GetEventsByUserIDs(ctx, missingUserIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Event], len(userIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Event]{Error: err}
				}
				return results
			}

			// Group events by user ID and convert to GraphQL model
			for _, row := range rows {
				event := &model.Event{
					ID:          row.ID,
					ProjectID:   row.ProjectID,
					Name:        row.Name,
					Description: row.Description,
					StartDate:   scalars.DateTime{Time: row.StartDate.Time},
					EndDate:     scalars.DateTime{Time: row.EndDate.Time},
				}
				eventsByUser[row.UserID] = append(eventsByUser[row.UserID], event)
			}

			// Populate cache for each user, including empty results
			for _, userID := range missingUserIDs {
				events := eventsByUser[userID]
				if events == nil {
					events = []*model.Event{} // Empty slice, not nil
				}
				eventsByUser[userID] = events
				c.Set(cache.EventsByUserKey(userID), events)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Event], len(userIDs))
		for i, userID := range userIDs {
			events := eventsByUser[userID]
			if events == nil {
				events = []*model.Event{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Event]{Data: events}
		}
		return results
	}
}
