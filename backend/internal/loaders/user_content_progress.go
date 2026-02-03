package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// userContentProgressBatchFunc batches loading user content progress by (user_id, achievement_id) pairs.
// Returns a slice of progress rows per key (empty slice if no progress).
func userContentProgressBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserAchievementKey) []*dataloader.Result[[]*sqlc.UserContentProgress] {
	return func(ctx context.Context, keys []UserAchievementKey) []*dataloader.Result[[]*sqlc.UserContentProgress] {
		progressByKey := make(map[string][]*sqlc.UserContentProgress)
		var missingKeys []UserAchievementKey

		for _, key := range keys {
			cacheKey := cache.UserContentProgressKey(key.UserID, key.AchievementID)

			if cached, ok := c.Get(cacheKey); ok {
				if rows, ok := cached.([]*sqlc.UserContentProgress); ok {
					progressByKey[key.String()] = rows
					continue
				}
				// Handle cached nil/empty (no progress for this user+achievement)
				if cached == nil {
					progressByKey[key.String()] = []*sqlc.UserContentProgress{}
					continue
				}
			}

			missingKeys = append(missingKeys, key)
		}

		if len(missingKeys) > 0 {
			userIDs := make([]string, len(missingKeys))
			achievementIDs := make([]string, len(missingKeys))
			for i, key := range missingKeys {
				userIDs[i] = key.UserID
				achievementIDs[i] = key.AchievementID
			}

			rows, err := db.Queries.GetBulkUserContentProgress(ctx, sqlc.GetBulkUserContentProgressParams{
				UserIds:        userIDs,
				AchievementIds: achievementIDs,
			})
			if err != nil {
				results := make([]*dataloader.Result[[]*sqlc.UserContentProgress], len(keys))
				for i := range results {
					results[i] = &dataloader.Result[[]*sqlc.UserContentProgress]{Error: err}
				}
				return results
			}

			// Group results by (user_id, achievement_id) pair
			rowsByKey := make(map[string][]*sqlc.UserContentProgress)
			for _, row := range rows {
				key := fmt.Sprintf("%s:%s", row.UserID, row.AchievementID)
				rowsByKey[key] = append(rowsByKey[key], row)
			}

			// Populate cache and result map for all missing keys
			for _, key := range missingKeys {
				keyStr := key.String()
				progress := rowsByKey[keyStr] // nil if no rows
				if progress == nil {
					progress = []*sqlc.UserContentProgress{}
				}
				progressByKey[keyStr] = progress
				c.Set(cache.UserContentProgressKey(key.UserID, key.AchievementID), progress)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*sqlc.UserContentProgress], len(keys))
		for i, key := range keys {
			results[i] = &dataloader.Result[[]*sqlc.UserContentProgress]{Data: progressByKey[key.String()]}
		}
		return results
	}
}
