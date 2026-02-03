package loaders

import (
	"context"
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// userAchievementCelebratedTimestampBatchFunc batches loading user achievement celebrated timestamps by user ID and achievement ID
func userAchievementCelebratedTimestampBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserAchievementKey) []*dataloader.Result[*time.Time] {
	return func(ctx context.Context, keys []UserAchievementKey) []*dataloader.Result[*time.Time] {
		timestampsByKey := make(map[string]*time.Time)
		missingKeys := []UserAchievementKey{}

		for _, key := range keys {
			cacheKey := cache.UserAchievementCelebratedTimestampKey(key.UserID, key.AchievementID)

			if cached, ok := c.Get(cacheKey); ok {
				if ts, ok := cached.(*time.Time); ok {
					timestampsByKey[key.String()] = ts
					continue
				}
				// Handle cached nil (user hasn't celebrated this)
				if cached == nil {
					timestampsByKey[key.String()] = nil
					continue
				}
			}

			missingKeys = append(missingKeys, key)
		}

		// Query database for all missing keys in a single bulk query
		if len(missingKeys) > 0 {
			userIDs := make([]string, len(missingKeys))
			achievementIDs := make([]string, len(missingKeys))
			for i, key := range missingKeys {
				userIDs[i] = key.UserID
				achievementIDs[i] = key.AchievementID
			}

			rows, err := db.Queries.GetBulkUserAchievementCelebratedTimestamps(ctx, sqlc.GetBulkUserAchievementCelebratedTimestampsParams{
				UserIds:        userIDs,
				AchievementIds: achievementIDs,
			})
			if err != nil {
				results := make([]*dataloader.Result[*time.Time], len(keys))
				for i := range results {
					results[i] = &dataloader.Result[*time.Time]{Error: err}
				}
				return results
			}

			// Map results by (user_id, achievement_id) pair
			celebratedByKey := make(map[string]*time.Time)
			for _, row := range rows {
				key := fmt.Sprintf("%s:%s", row.UserID, row.AchievementID)
				if row.CelebratedAt.Valid {
					ts := row.CelebratedAt.Time
					celebratedByKey[key] = &ts
				}
			}

			// Populate cache and result map for all missing keys
			for _, key := range missingKeys {
				ts := celebratedByKey[key.String()] // nil if not celebrated
				timestampsByKey[key.String()] = ts
				c.Set(cache.UserAchievementCelebratedTimestampKey(key.UserID, key.AchievementID), ts)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[*time.Time], len(keys))
		for i, key := range keys {
			ts := timestampsByKey[key.String()]
			results[i] = &dataloader.Result[*time.Time]{Data: ts}
		}
		return results
	}
}
