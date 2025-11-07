package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// eventsByProjectBatchFunc batches loading events by project IDs
func eventsByProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.Event] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]*model.Event] {
		// Check cache first for each project ID
		eventsByProject := make(map[string][]*model.Event)
		missingProjectIDs := []string{}

		for _, projectID := range projectIDs {
			cacheKey := cache.EventsByProjectKey(projectID)
			if cached, ok := c.Get(cacheKey); ok {
				if events, ok := cached.([]*model.Event); ok {
					eventsByProject[projectID] = events
					continue
				}
			}
			missingProjectIDs = append(missingProjectIDs, projectID)
		}

		// Query database only for cache misses
		if len(missingProjectIDs) > 0 {
			rows, err := db.Queries.GetEventsByProjectIDs(ctx, missingProjectIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.Event], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.Event]{Error: err}
				}
				return results
			}

			// Group events by project ID and convert to GraphQL model
			for _, row := range rows {
				event := &model.Event{
					ID:          row.ID,
					ProjectID:   row.ProjectID,
					Name:        row.Name,
					Description: row.Description,
					StartDate:   scalars.DateTime{Time: row.StartDate.Time},
					EndDate:     scalars.DateTime{Time: row.EndDate.Time},
				}
				eventsByProject[row.ProjectID] = append(eventsByProject[row.ProjectID], event)
			}

			// Populate cache for each project, including empty results
			for _, projectID := range missingProjectIDs {
				events := eventsByProject[projectID]
				if events == nil {
					events = []*model.Event{} // Empty slice, not nil
				}
				eventsByProject[projectID] = events
				c.Set(cache.EventsByProjectKey(projectID), events)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*model.Event], len(projectIDs))
		for i, projectID := range projectIDs {
			events := eventsByProject[projectID]
			if events == nil {
				events = []*model.Event{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]*model.Event]{Data: events}
		}
		return results
	}
}
