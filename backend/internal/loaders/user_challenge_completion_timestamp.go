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

// UserChallengeKey combines user ID and challenge ID for dataloader key
type UserChallengeKey struct {
	UserID      string
	ChallengeID string
}

func (k UserChallengeKey) String() string {
	return fmt.Sprintf("%s:%s", k.UserID, k.ChallengeID)
}

func (k UserChallengeKey) Raw() interface{} {
	return k
}

// userChallengeCompletionTimestampBatchFunc batches loading user challenge completion timestamps
func userChallengeCompletionTimestampBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserChallengeKey) []*dataloader.Result[*time.Time] {
	return func(ctx context.Context, keys []UserChallengeKey) []*dataloader.Result[*time.Time] {
		timestampsByKey := make(map[string]*time.Time)
		missingKeys := []UserChallengeKey{}

		for _, key := range keys {
			cacheKey := cache.UserChallengeCompletionKey(key.UserID, key.ChallengeID)

			if cached, ok := c.Get(cacheKey); ok {
				if ts, ok := cached.(*time.Time); ok {
					timestampsByKey[key.String()] = ts
					continue
				}
				// Handle cached nil (user hasn't completed this)
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
			challengeIDs := make([]string, len(missingKeys))
			for i, key := range missingKeys {
				userIDs[i] = key.UserID
				challengeIDs[i] = key.ChallengeID
			}

			rows, err := db.Queries.GetBulkUserCompletionTimestamps(ctx, sqlc.GetBulkUserCompletionTimestampsParams{
				Userids:      userIDs,
				Challengeids: challengeIDs,
			})
			if err != nil {
				results := make([]*dataloader.Result[*time.Time], len(keys))
				for i := range results {
					results[i] = &dataloader.Result[*time.Time]{Error: err}
				}
				return results
			}

			// Map results by (user_id, challenge_id) pair
			completionByKey := make(map[string]*time.Time)
			for _, row := range rows {
				key := fmt.Sprintf("%s:%s", row.UserID, row.ChallengeID)
				if row.CompletedAt.Valid {
					ts := row.CompletedAt.Time
					completionByKey[key] = &ts
				}
			}

			// Populate cache and result map for all missing keys
			for _, key := range missingKeys {
				ts := completionByKey[key.String()] // nil if not completed
				timestampsByKey[key.String()] = ts
				c.Set(cache.UserChallengeCompletionKey(key.UserID, key.ChallengeID), ts)
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
