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

// UserAchievementKey combines user ID and achievement ID for dataloader key
type UserAchievementKey struct {
	UserID        string
	AchievementID string
}

func (k UserAchievementKey) String() string {
	return fmt.Sprintf("%s:%s", k.UserID, k.AchievementID)
}

func (k UserAchievementKey) Raw() interface{} {
	return k
}

// userAchievementTimestampBatchFunc batches loading user achievement timestamps by user ID and achievement ID
func userAchievementTimestampBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserAchievementKey) []*dataloader.Result[*time.Time] {
	return func(ctx context.Context, keys []UserAchievementKey) []*dataloader.Result[*time.Time] {
		// Group keys by user ID to optimize queries
		timestampsByKey := make(map[string]*time.Time)
		missingKeys := []UserAchievementKey{}
		keysByUser := make(map[string][]UserAchievementKey)

		for _, key := range keys {
			cacheKey := cache.UserAchievementTimestampKey(key.UserID, key.AchievementID)

			if cached, ok := c.Get(cacheKey); ok {
				if ts, ok := cached.(*time.Time); ok {
					timestampsByKey[key.String()] = ts
					continue
				}
				// Handle cached nil (user hasn't achieved this)
				if cached == nil {
					timestampsByKey[key.String()] = nil
					continue
				}
			}

			missingKeys = append(missingKeys, key)
			keysByUser[key.UserID] = append(keysByUser[key.UserID], key)
		}

		// Query database for each user with missing keys
		if len(missingKeys) > 0 {
			for userID, userKeys := range keysByUser {
				achievementIDs := make([]string, len(userKeys))
				for i, key := range userKeys {
					achievementIDs[i] = key.AchievementID
				}

				rows, err := db.Queries.GetUserAchievementTimestamps(ctx, sqlc.GetUserAchievementTimestampsParams{
					Userid:         userID,
					AchievementIds: achievementIDs,
				})
				if err != nil {
					// Return error for all keys
					results := make([]*dataloader.Result[*time.Time], len(keys))
					for i := range results {
						results[i] = &dataloader.Result[*time.Time]{Error: err}
					}
					return results
				}

				// Map results by achievement ID
				achievedAtByAchievement := make(map[string]*time.Time)
				for _, row := range rows {
					if row.AchievedAt.Valid {
						ts := row.AchievedAt.Time
						achievedAtByAchievement[row.AchievementID] = &ts
					}
				}

				// Populate cache and result map for each key
				for _, key := range userKeys {
					ts := achievedAtByAchievement[key.AchievementID] // nil if not achieved
					timestampsByKey[key.String()] = ts
					c.Set(cache.UserAchievementTimestampKey(key.UserID, key.AchievementID), ts)
				}
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
