package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// articlesByAchievementBatchFunc batches loading articles by achievement IDs
func articlesByAchievementBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.Article] {
	return func(ctx context.Context, achievementIDs []string) []*dataloader.Result[[]model.Article] {
		// Check cache first for each achievement ID
		articlesByAchievement := make(map[string][]model.Article)
		missingAchievementIDs := []string{}

		for _, achievementID := range achievementIDs {
			cacheKey := cache.ArticlesByAchievementKey(achievementID)
			if cached, ok := c.Get(cacheKey); ok {
				if articles, ok := cached.([]model.Article); ok {
					articlesByAchievement[achievementID] = articles
					continue
				}
			}
			missingAchievementIDs = append(missingAchievementIDs, achievementID)
		}

		// Query database only for cache misses
		if len(missingAchievementIDs) > 0 {
			rows, err := db.Queries.GetArticlesByAchievementIDs(ctx, missingAchievementIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.Article], len(achievementIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.Article]{Error: err}
				}
				return results
			}

			// Group articles by achievement ID and convert to GraphQL model
			for _, row := range rows {
				article := model.Article{
					ID:     row.ID, // Use the database row ID, not ArticleID
					Title:  row.Title,
					Author: row.Author,
					URL:    row.Url,
				}
				articlesByAchievement[row.AchievementID] = append(articlesByAchievement[row.AchievementID], article)
			}

			// Populate cache for each achievement, including empty results
			for _, achievementID := range missingAchievementIDs {
				articles := articlesByAchievement[achievementID]
				if articles == nil {
					articles = []model.Article{} // Empty slice, not nil
				}
				articlesByAchievement[achievementID] = articles
				c.Set(cache.ArticlesByAchievementKey(achievementID), articles)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.Article], len(achievementIDs))
		for i, achievementID := range achievementIDs {
			articles := articlesByAchievement[achievementID]
			if articles == nil {
				articles = []model.Article{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.Article]{Data: articles}
		}
		return results
	}
}
