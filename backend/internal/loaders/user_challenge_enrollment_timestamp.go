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

// userChallengeEnrollmentTimestampBatchFunc batches loading user challenge enrollment timestamps
func userChallengeEnrollmentTimestampBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserChallengeKey) []*dataloader.Result[*time.Time] {
	return func(ctx context.Context, keys []UserChallengeKey) []*dataloader.Result[*time.Time] {
		timestampsByKey := make(map[string]*time.Time)
		missingKeys := []UserChallengeKey{}

		for _, key := range keys {
			cacheKey := cache.UserChallengeEnrollmentKey(key.UserID, key.ChallengeID)

			if cached, ok := c.Get(cacheKey); ok {
				if ts, ok := cached.(*time.Time); ok {
					timestampsByKey[key.String()] = ts
					continue
				}
				// Handle cached nil (user hasn't enrolled in this)
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

			rows, err := db.Queries.GetBulkUserEnrollmentTimestamps(ctx, sqlc.GetBulkUserEnrollmentTimestampsParams{
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
			enrollmentByKey := make(map[string]*time.Time)
			for _, row := range rows {
				key := fmt.Sprintf("%s:%s", row.UserID, row.ChallengeID)
				if row.EnrolledAt.Valid {
					ts := row.EnrolledAt.Time
					enrollmentByKey[key] = &ts
				}
			}

			// Populate cache and result map for all missing keys
			for _, key := range missingKeys {
				ts := enrollmentByKey[key.String()] // nil if not enrolled
				timestampsByKey[key.String()] = ts
				c.Set(cache.UserChallengeEnrollmentKey(key.UserID, key.ChallengeID), ts)
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
