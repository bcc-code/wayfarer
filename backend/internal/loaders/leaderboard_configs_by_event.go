package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// leaderboardConfigsByEventBatchFunc batches loading leaderboard configs by event IDs.
// Returns ALL configs (including inactive) — non-admin visibility filtering happens in the resolver.
func leaderboardConfigsByEventBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.LeaderboardConfig] {
	return func(ctx context.Context, eventIDs []string) []*dataloader.Result[[]*model.LeaderboardConfig] {
		configsByEvent, missingEventIDs := partitionLeaderboardConfigsByEventCache(eventIDs, c)

		if len(missingEventIDs) > 0 {
			rows, err := db.Queries.GetLeaderboardConfigsByEventIDs(ctx, missingEventIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.LeaderboardConfig], len(eventIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.LeaderboardConfig]{Error: err}
				}
				return results
			}
			storeLeaderboardConfigsByEventInCache(missingEventIDs, rows, configsByEvent, c)
		}

		results := make([]*dataloader.Result[[]*model.LeaderboardConfig], len(eventIDs))
		for i, eventID := range eventIDs {
			configs := configsByEvent[eventID]
			if configs == nil {
				configs = []*model.LeaderboardConfig{}
			}
			results[i] = &dataloader.Result[[]*model.LeaderboardConfig]{Data: configs}
		}
		return results
	}
}

// partitionLeaderboardConfigsByEventCache splits the requested (possibly duplicated) event
// IDs into those already served from cache and the deduplicated set that still needs a DB
// round-trip.
func partitionLeaderboardConfigsByEventCache(eventIDs []string, c *cache.CacheWithRegistry) (cached map[string][]*model.LeaderboardConfig, missing []string) {
	cached = make(map[string][]*model.LeaderboardConfig)
	missing = []string{}
	seen := make(map[string]bool)

	for _, eventID := range eventIDs {
		if seen[eventID] {
			continue
		}
		seen[eventID] = true

		cacheKey := cache.LeaderboardConfigsByEventKey(eventID)
		if val, ok := c.Get(cacheKey); ok {
			if configs, ok := val.([]*model.LeaderboardConfig); ok {
				cached[eventID] = configs
				continue
			}
		}
		missing = append(missing, eventID)
	}

	return cached, missing
}

// storeLeaderboardConfigsByEventInCache groups freshly fetched rows by event ID and writes
// each requested-but-missing event's (possibly empty) config list into both the result map
// and the cache, so an event with zero configs is cached as an empty slice rather than
// remaining a permanent cache miss on subsequent loads.
func storeLeaderboardConfigsByEventInCache(missingEventIDs []string, rows []*sqlc.LeaderboardConfig, configsByEvent map[string][]*model.LeaderboardConfig, c *cache.CacheWithRegistry) {
	for _, row := range rows {
		// row.EventID is guaranteed non-nil: the query filters on
		// event_id = ANY(@event_ids), and NULL = ANY(non-null array) is
		// never true in Postgres, so NULL event_id rows are excluded.
		config := ConvertRowToLeaderboardConfig(row)
		configsByEvent[*row.EventID] = append(configsByEvent[*row.EventID], config)
	}

	for _, eventID := range missingEventIDs {
		configs := configsByEvent[eventID]
		if configs == nil {
			configs = []*model.LeaderboardConfig{}
		}
		configsByEvent[eventID] = configs
		c.Set(cache.LeaderboardConfigsByEventKey(eventID), configs)
	}
}
