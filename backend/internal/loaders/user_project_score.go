package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// UserProjectKey combines user ID and project ID for dataloader key
type UserProjectKey struct {
	UserID    string
	ProjectID string
}

func (k UserProjectKey) String() string {
	return fmt.Sprintf("%s:%s", k.UserID, k.ProjectID)
}

func (k UserProjectKey) Raw() interface{} {
	return k
}

// userProjectScoreBatchFunc batches user project score lookups into a single
// grouped query. Scores are intentionally not stored in Ristretto: they change
// on every point award and have no invalidation path, so we only batch per
// request window.
func userProjectScoreBatchFunc(db *database.DB) func(context.Context, []UserProjectKey) []*dataloader.Result[int64] {
	return func(ctx context.Context, keys []UserProjectKey) []*dataloader.Result[int64] {
		userIDs := make([]string, len(keys))
		projectIDs := make([]string, len(keys))
		for i, key := range keys {
			userIDs[i] = key.UserID
			projectIDs[i] = key.ProjectID
		}

		rows, err := db.Queries.GetBulkUserProjectScores(ctx, sqlc.GetBulkUserProjectScoresParams{
			UserIds:    userIDs,
			ProjectIds: projectIDs,
		})
		if err != nil {
			results := make([]*dataloader.Result[int64], len(keys))
			for i := range results {
				results[i] = &dataloader.Result[int64]{Error: err}
			}
			return results
		}

		return mapUserProjectScores(keys, rows)
	}
}

// mapUserProjectScores maps grouped score rows back to the input keys in
// order; keys without a journal entry get a score of 0.
func mapUserProjectScores(keys []UserProjectKey, rows []*sqlc.GetBulkUserProjectScoresRow) []*dataloader.Result[int64] {
	scoreByKey := make(map[string]int64, len(rows))
	for _, row := range rows {
		scoreByKey[fmt.Sprintf("%s:%s", row.UserID, row.ProjectID)] = row.TotalScore
	}

	results := make([]*dataloader.Result[int64], len(keys))
	for i, key := range keys {
		results[i] = &dataloader.Result[int64]{Data: scoreByKey[key.String()]}
	}
	return results
}
