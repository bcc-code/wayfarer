package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// leaderboardConfigsByProjectBatchFunc batches loading leaderboard configs by project IDs.
// Returns ALL configs (including inactive) — non-admin visibility filtering happens in the resolver.
func leaderboardConfigsByProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.LeaderboardConfig] {
	return func(ctx context.Context, projectIDs []string) []*dataloader.Result[[]*model.LeaderboardConfig] {
		configsByProject, missingProjectIDs := partitionLeaderboardConfigsByProjectCache(projectIDs, c)

		if len(missingProjectIDs) > 0 {
			rows, err := db.Queries.GetLeaderboardConfigsByProjectIDs(ctx, missingProjectIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.LeaderboardConfig], len(projectIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.LeaderboardConfig]{Error: err}
				}
				return results
			}
			storeLeaderboardConfigsByProjectInCache(missingProjectIDs, rows, configsByProject, c)
		}

		results := make([]*dataloader.Result[[]*model.LeaderboardConfig], len(projectIDs))
		for i, projectID := range projectIDs {
			configs := configsByProject[projectID]
			if configs == nil {
				configs = []*model.LeaderboardConfig{}
			}
			results[i] = &dataloader.Result[[]*model.LeaderboardConfig]{Data: configs}
		}
		return results
	}
}

// partitionLeaderboardConfigsByProjectCache splits the requested (possibly duplicated)
// project IDs into those already served from cache and the deduplicated set that still
// needs a DB round-trip.
func partitionLeaderboardConfigsByProjectCache(projectIDs []string, c *cache.CacheWithRegistry) (cached map[string][]*model.LeaderboardConfig, missing []string) {
	cached = make(map[string][]*model.LeaderboardConfig)
	missing = []string{}
	seen := make(map[string]bool)

	for _, projectID := range projectIDs {
		if seen[projectID] {
			continue
		}
		seen[projectID] = true

		cacheKey := cache.LeaderboardConfigsByProjectKey(projectID)
		if val, ok := c.Get(cacheKey); ok {
			if configs, ok := val.([]*model.LeaderboardConfig); ok {
				cached[projectID] = configs
				continue
			}
		}
		missing = append(missing, projectID)
	}

	return cached, missing
}

// storeLeaderboardConfigsByProjectInCache groups freshly fetched rows by project ID and
// writes each requested-but-missing project's (possibly empty) config list into both the
// result map and the cache, so a project with zero configs is cached as an empty slice
// rather than remaining a permanent cache miss on subsequent loads.
func storeLeaderboardConfigsByProjectInCache(missingProjectIDs []string, rows []*sqlc.LeaderboardConfig, configsByProject map[string][]*model.LeaderboardConfig, c *cache.CacheWithRegistry) {
	for _, row := range rows {
		config := ConvertRowToLeaderboardConfig(row)
		configsByProject[row.ProjectID] = append(configsByProject[row.ProjectID], config)
	}

	for _, projectID := range missingProjectIDs {
		configs := configsByProject[projectID]
		if configs == nil {
			configs = []*model.LeaderboardConfig{}
		}
		configsByProject[projectID] = configs
		c.Set(cache.LeaderboardConfigsByProjectKey(projectID), configs)
	}
}
