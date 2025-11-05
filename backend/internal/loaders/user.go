package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// userByIDBatchFunc batches user loading by IDs
func userByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.User] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.User] {
		// Check cache first for each ID
		userMap := make(map[string]*model.User)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.UserKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if user, ok := cached.(*model.User); ok {
					userMap[id] = user
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetUsersByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.User], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.User]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				// Convert birthdate to string (always valid since birthdate is required)
				birthdateStr := row.Birthdate.Time.Format("2006-01-02")

				user := &model.User{
					ID:        row.ID,
					MembersID: row.MembersID,
					Gender:    model.Gender(row.Gender),
					ChurchID:  row.ChurchID,
					Birthdate: birthdateStr,
					Email:     row.Email,
					Name:      row.Name,
					Image:     row.AvatarUrl,
				}

				userMap[row.ID] = user
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.UserKey(row.ID), user)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[*model.User], len(ids))
		for i, id := range ids {
			if user, ok := userMap[id]; ok {
				results[i] = &dataloader.Result[*model.User]{Data: user}
			} else {
				results[i] = &dataloader.Result[*model.User]{
					Error: fmt.Errorf("user not found: %s", id),
				}
			}
		}
		return results
	}
}
