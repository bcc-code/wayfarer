package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// quizSubmissionByIDBatchFunc batches loading quiz submissions by ID
func quizSubmissionByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*sqlc.GetQuizSubmissionsByIDsRow] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*sqlc.GetQuizSubmissionsByIDsRow] {
		submissions := make(map[string]*sqlc.GetQuizSubmissionsByIDsRow)

		rows, err := db.Queries.GetQuizSubmissionsByIDs(ctx, ids)
		if err != nil {
			results := make([]*dataloader.Result[*sqlc.GetQuizSubmissionsByIDsRow], len(ids))
			for i := range results {
				results[i] = &dataloader.Result[*sqlc.GetQuizSubmissionsByIDsRow]{Error: err}
			}
			return results
		}

		for _, row := range rows {
			submissions[row.ID] = row
		}

		results := make([]*dataloader.Result[*sqlc.GetQuizSubmissionsByIDsRow], len(ids))
		for i, id := range ids {
			results[i] = &dataloader.Result[*sqlc.GetQuizSubmissionsByIDsRow]{Data: submissions[id]}
		}
		return results
	}
}
