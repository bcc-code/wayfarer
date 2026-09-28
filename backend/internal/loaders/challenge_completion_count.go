package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// challengeCompletionCountBatchFunc batches per-challenge completion counts.
// Deliberately not cached in Ristretto: the count changes on every completion
// and has no invalidation path, so it is only batched per request window.
func challengeCompletionCountBatchFunc(db *database.DB) func(context.Context, []string) []*dataloader.Result[int64] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[int64] {
		rows, err := db.Queries.GetBulkChallengeCompletionCounts(ctx, keys)
		if err != nil {
			results := make([]*dataloader.Result[int64], len(keys))
			for i := range results {
				results[i] = &dataloader.Result[int64]{Error: err}
			}
			return results
		}

		return mapChallengeCompletionCounts(keys, rows)
	}
}

// mapChallengeCompletionCounts maps grouped count rows back to the input keys
// in order; a challenge nobody has completed is absent from the rows and gets 0.
func mapChallengeCompletionCounts(keys []string, rows []*sqlc.GetBulkChallengeCompletionCountsRow) []*dataloader.Result[int64] {
	countByID := make(map[string]int64, len(rows))
	for _, row := range rows {
		countByID[row.ChallengeID] = row.CompletionCount
	}

	results := make([]*dataloader.Result[int64], len(keys))
	for i, key := range keys {
		results[i] = &dataloader.Result[int64]{Data: countByID[key]}
	}
	return results
}
