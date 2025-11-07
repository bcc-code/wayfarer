package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// relevantDaysByStreakBatchFunc batches loading relevant days by streak IDs
func relevantDaysByStreakBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.DateRange] {
	return func(ctx context.Context, streakIDs []string) []*dataloader.Result[[]model.DateRange] {
		// Check cache first for each streak ID
		relevantDaysByStreak := make(map[string][]model.DateRange)
		missingStreakIDs := []string{}

		for _, streakID := range streakIDs {
			cacheKey := cache.RelevantDaysByStreakKey(streakID)
			if cached, ok := c.Get(cacheKey); ok {
				if relevantDays, ok := cached.([]model.DateRange); ok {
					relevantDaysByStreak[streakID] = relevantDays
					continue
				}
			}
			missingStreakIDs = append(missingStreakIDs, streakID)
		}

		// Query database only for cache misses
		if len(missingStreakIDs) > 0 {
			rows, err := db.Queries.GetRelevantDaysByStreakIDs(ctx, missingStreakIDs)
			if err != nil {
				results := make([]*dataloader.Result[[]model.DateRange], len(streakIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.DateRange]{Error: err}
				}
				return results
			}

			// Group relevant days by streak ID and convert to GraphQL model
			for _, row := range rows {
				dateRange := model.DateRange{
					Start: scalars.Date{Time: row.StartDate.Time},
					End:   scalars.Date{Time: row.EndDate.Time},
				}
				relevantDaysByStreak[row.StreakID] = append(relevantDaysByStreak[row.StreakID], dateRange)
			}

			// Populate cache for each streak, including empty results
			for _, streakID := range missingStreakIDs {
				relevantDays := relevantDaysByStreak[streakID]
				if relevantDays == nil {
					relevantDays = []model.DateRange{} // Empty slice, not nil
				}
				relevantDaysByStreak[streakID] = relevantDays
				c.Set(cache.RelevantDaysByStreakKey(streakID), relevantDays)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[[]model.DateRange], len(streakIDs))
		for i, streakID := range streakIDs {
			relevantDays := relevantDaysByStreak[streakID]
			if relevantDays == nil {
				relevantDays = []model.DateRange{} // Return empty slice, not nil
			}
			results[i] = &dataloader.Result[[]model.DateRange]{Data: relevantDays}
		}
		return results
	}
}
