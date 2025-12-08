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
		activitiesByKey := make(map[string][]*sqlc.UserStreakActivity)
		missingKeys := []UserStreakActivityKey{}

		for _, key := range keys {
			cacheKey := cache.UserStreakActivityKey(key.UserID, key.StreakID)

			if cached, ok := c.Get(cacheKey); ok {
				if activities, ok := cached.([]*sqlc.UserStreakActivity); ok {
					activitiesByKey[key.String()] = activities
					continue
				}
			}

			missingKeys = append(missingKeys, key)
		}

		// Query database for all missing keys in a single bulk query
		if len(missingKeys) > 0 {
			userIDs := make([]string, len(missingKeys))
			streakIDs := make([]string, len(missingKeys))
			for i, key := range missingKeys {
				userIDs[i] = key.UserID
				streakIDs[i] = key.StreakID
			}

			rows, err := db.Queries.GetBulkUserStreakActivities(ctx, sqlc.GetBulkUserStreakActivitiesParams{
				UserIds:   userIDs,
				StreakIds: streakIDs,
			})
			if err != nil {
				results := make([]*dataloader.Result[[]*sqlc.UserStreakActivity], len(keys))
				for i := range results {
					results[i] = &dataloader.Result[[]*sqlc.UserStreakActivity]{Error: err}
				}
				return results
			}

			// Group activities by (user_id, streak_id) pair
			activitiesByPair := make(map[string][]*sqlc.UserStreakActivity)
			for _, row := range rows {
				key := fmt.Sprintf("%s:%s", row.UserID, row.StreakID)
				activitiesByPair[key] = append(activitiesByPair[key], row)
			}

			// Populate cache and result map for all missing keys
			for _, key := range missingKeys {
				activities := activitiesByPair[key.String()]
				if activities == nil {
					activities = []*sqlc.UserStreakActivity{}
				}
				activitiesByKey[key.String()] = activities
				c.Set(cache.UserStreakActivityKey(key.UserID, key.StreakID), activities)
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
