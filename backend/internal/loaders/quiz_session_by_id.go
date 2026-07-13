package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// quizSessionByIDBatchFunc batches loading quiz sessions by ID
func quizSessionByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*sqlc.QuizSession] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*sqlc.QuizSession] {
		// Check cache first
		sessions := make(map[string]*sqlc.QuizSession)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.QuizSessionKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if session, ok := cached.(*sqlc.QuizSession); ok {
					sessions[id] = session
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizSessionsByIDs(ctx, missingIDs)
			if err != nil {
				results := make([]*dataloader.Result[*sqlc.QuizSession], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*sqlc.QuizSession]{Error: err}
				}
				return results
			}

			// Index by ID and cache. The short TTL bounds staleness for any
			// state-transition path that misses InvalidateQuizSession.
			for _, row := range rows {
				session := row // Copy to avoid pointer issues
				sessions[row.ID] = session
				c.SetWithTTL(cache.QuizSessionKey(row.ID), session, QuizSessionAccessTTL)
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[*sqlc.QuizSession], len(ids))
		for i, id := range ids {
			results[i] = &dataloader.Result[*sqlc.QuizSession]{Data: sessions[id]}
		}
		return results
	}
}
