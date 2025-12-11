package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// contentItemsByAchievementBatchFunc batches loading content items by achievement IDs
func contentItemsByAchievementBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.ContentItem] {
	return func(ctx context.Context, achievementIDs []string) []*dataloader.Result[[]*model.ContentItem] {
		// Check cache first
		itemsByAchievement := make(map[string][]*model.ContentItem)
		missingAchievementIDs := []string{}

		for _, achievementID := range achievementIDs {
			cacheKey := cache.ContentItemsByAchievementKey(achievementID)
			if cached, ok := c.Get(cacheKey); ok {
				if items, ok := cached.([]*model.ContentItem); ok {
					itemsByAchievement[achievementID] = items
					continue
				}
			}
			missingAchievementIDs = append(missingAchievementIDs, achievementID)
		}

		// Query database for cache misses
		if len(missingAchievementIDs) > 0 {
			rows, err := db.Queries.GetContentItemsByAchievementIDs(ctx, missingAchievementIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]*model.ContentItem], len(achievementIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.ContentItem]{Error: err}
				}
				return results
			}

			// Group by achievement ID
			for _, row := range rows {
				item := &model.ContentItem{
					ID:                row.ID,
					ExternalContentID: row.ExternalContentID,
					SortOrder:         int(row.SortOrder),
				}
				itemsByAchievement[row.AchievementID] = append(itemsByAchievement[row.AchievementID], item)
			}

			// Populate cache
			for _, achievementID := range missingAchievementIDs {
				items := itemsByAchievement[achievementID]
				if items == nil {
					items = []*model.ContentItem{}
				}
				itemsByAchievement[achievementID] = items
				c.Set(cache.ContentItemsByAchievementKey(achievementID), items)
			}
		}

		// Return results in order
		results := make([]*dataloader.Result[[]*model.ContentItem], len(achievementIDs))
		for i, achievementID := range achievementIDs {
			items := itemsByAchievement[achievementID]
			if items == nil {
				items = []*model.ContentItem{}
			}
			results[i] = &dataloader.Result[[]*model.ContentItem]{Data: items}
		}
		return results
	}
}
