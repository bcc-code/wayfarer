package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// leaderboardConfigByIDBatchFunc batches loading leaderboard configs by IDs
func leaderboardConfigByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.LeaderboardConfig] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.LeaderboardConfig] {
		configMap := make(map[string]*model.LeaderboardConfig)
		missingIDs := []string{}
		seen := make(map[string]bool)

		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true

			cacheKey := cache.LeaderboardConfigKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if config, ok := cached.(*model.LeaderboardConfig); ok {
					configMap[id] = config
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetLeaderboardConfigsByIDs(ctx, missingIDs)
			if err != nil {
				results := make([]*dataloader.Result[*model.LeaderboardConfig], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.LeaderboardConfig]{Error: err}
				}
				return results
			}

			for _, row := range rows {
				config := ConvertRowToLeaderboardConfig(row)
				configMap[row.ID] = config
				c.Set(cache.LeaderboardConfigKey(row.ID), config)
			}
		}

		results := make([]*dataloader.Result[*model.LeaderboardConfig], len(ids))
		for i, id := range ids {
			if config, ok := configMap[id]; ok {
				results[i] = &dataloader.Result[*model.LeaderboardConfig]{Data: config}
			} else {
				results[i] = &dataloader.Result[*model.LeaderboardConfig]{
					Error: fmt.Errorf("leaderboard config not found: %s", id),
				}
			}
		}
		return results
	}
}

// ConvertRowToLeaderboardConfig converts a database row to the GraphQL model
func ConvertRowToLeaderboardConfig(row *sqlc.LeaderboardConfig) *model.LeaderboardConfig {
	var filter *string
	if len(row.Filter) > 0 {
		f := string(row.Filter)
		filter = &f
	}

	return &model.LeaderboardConfig{
		ID:         row.ID,
		ProjectID:  row.ProjectID,
		EventID:    row.EventID,
		Name:       row.Name,
		Slug:       row.Slug,
		EntityType: model.LeaderboardEntityType(row.EntityType),
		Filter:     filter,
		SortOrder:  int(row.SortOrder),
		IsActive:   row.IsActive,
		CreatedAt:  scalars.DateTime{Time: row.CreatedAt.Time},
		UpdatedAt:  scalars.DateTime{Time: row.UpdatedAt.Time},
	}
}
