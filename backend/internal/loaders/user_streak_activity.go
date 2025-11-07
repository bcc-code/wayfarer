package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// UserStreakActivityKey combines user ID and streak ID for dataloader key
type UserStreakActivityKey struct {
	UserID   string
	StreakID string
}

func (k UserStreakActivityKey) String() string {
	return fmt.Sprintf("%s:%s", k.UserID, k.StreakID)
}

func (k UserStreakActivityKey) Raw() interface{} {
	return k
}

// userStreakActivityBatchFunc batches loading user streak activities by user ID and streak ID
func userStreakActivityBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserStreakActivityKey) []*dataloader.Result[[]*sqlc.UserStreakActivity] {
	return func(ctx context.Context, keys []UserStreakActivityKey) []*dataloader.Result[[]*sqlc.UserStreakActivity] {
		// Group keys by user ID to optimize queries
		activitiesByKey := make(map[string][]*sqlc.UserStreakActivity)
		missingKeys := []UserStreakActivityKey{}
		keysByUser := make(map[string][]UserStreakActivityKey)

		for _, key := range keys {
			cacheKey := cache.UserStreakActivityKey(key.UserID, key.StreakID)

			if cached, ok := c.Get(cacheKey); ok {
				if activities, ok := cached.([]*sqlc.UserStreakActivity); ok {
					activitiesByKey[key.String()] = activities
					continue
				}
			}

			missingKeys = append(missingKeys, key)
			keysByUser[key.UserID] = append(keysByUser[key.UserID], key)
		}

		// Query database for each user with missing keys
		if len(missingKeys) > 0 {
			for userID, userKeys := range keysByUser {
				streakIDs := make([]string, len(userKeys))
				for i, key := range userKeys {
					streakIDs[i] = key.StreakID
				}

				rows, err := db.Queries.GetUserStreakActivitiesForMultipleStreaks(ctx, sqlc.GetUserStreakActivitiesForMultipleStreaksParams{
					Userid:    userID,
					StreakIds: streakIDs,
				})
				if err != nil {
					// Return error for all keys
					results := make([]*dataloader.Result[[]*sqlc.UserStreakActivity], len(keys))
					for i := range results {
						results[i] = &dataloader.Result[[]*sqlc.UserStreakActivity]{Error: err}
					}
					return results
				}

				// Group activities by streak ID
				activitiesByStreak := make(map[string][]*sqlc.UserStreakActivity)
				for _, row := range rows {
					activitiesByStreak[row.StreakID] = append(activitiesByStreak[row.StreakID], row)
				}

				// Populate cache and result map for each key
				for _, key := range userKeys {
					activities := activitiesByStreak[key.StreakID]
					if activities == nil {
						activities = []*sqlc.UserStreakActivity{}
					}
					activitiesByKey[key.String()] = activities
					c.Set(cache.UserStreakActivityKey(key.UserID, key.StreakID), activities)
				}
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]*sqlc.UserStreakActivity], len(keys))
		for i, key := range keys {
			activities := activitiesByKey[key.String()]
			if activities == nil {
				activities = []*sqlc.UserStreakActivity{}
			}
			results[i] = &dataloader.Result[[]*sqlc.UserStreakActivity]{Data: activities}
		}
		return results
	}
}
