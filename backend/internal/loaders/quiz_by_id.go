package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// quizByIDBatchFunc batches loading quizzes by IDs
func quizByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Quiz] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[*model.Quiz] {
		// Check cache first for each ID
		quizMap := make(map[string]*model.Quiz)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.QuizKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if quiz, ok := cached.(*model.Quiz); ok {
					quizMap[id] = quiz
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizzesByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Quiz], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[*model.Quiz]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				var publishedAt *scalars.DateTime
				if row.PublishedAt.Valid {
					publishedAt = &scalars.DateTime{Time: row.PublishedAt.Time}
				}

				var endTime *scalars.DateTime
				if row.EndTime.Valid {
					endTime = &scalars.DateTime{Time: row.EndTime.Time}
				}

				var timeoutSeconds *int
				if row.TimeoutSeconds != nil {
					ts := int(*row.TimeoutSeconds)
					timeoutSeconds = &ts
				}

				quiz := &model.Quiz{
					ID:                   row.ID,
					Name:                 row.Name,
					Description:          row.Description,
					Image:                row.ImageUrl,
					TimeoutSeconds:       timeoutSeconds,
					RandomizeQuestions:   row.RandomizeQuestions,
					RevealCorrectAnswers: row.RevealCorrectAnswers,
					AllowRetakes:         row.AllowRetakes,
					CompletionPoints:     int(row.CompletionPoints),
					PublishedAt:          publishedAt,
					EndTime:              endTime,
					// Fields for resolvers
					ProjectID:   row.ProjectID,
					ChallengeID: row.ChallengeID,
				}

				// Store in cache
				c.Set(cache.QuizKey(row.ID), quiz)
				quizMap[row.ID] = quiz
			}
		}

		// Return results in the same order as input IDs
		results := make([]*dataloader.Result[*model.Quiz], len(ids))
		for i, id := range ids {
			if quiz, ok := quizMap[id]; ok {
				results[i] = &dataloader.Result[*model.Quiz]{Data: quiz}
			} else {
				results[i] = &dataloader.Result[*model.Quiz]{Data: nil}
			}
		}
		return results
	}
}
