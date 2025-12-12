package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/graph-gophers/dataloader/v7"
)

// quizByChallengeIDBatchFunc batches loading quizzes by challenge IDs
func quizByChallengeIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[*model.Quiz] {
	return func(ctx context.Context, challengeIDs []string) []*dataloader.Result[*model.Quiz] {
		ctx, span := otel.StartDataloaderSpan(ctx, "QuizByChallengeID", len(challengeIDs))
		defer span.End()

		// Check cache first for each challenge ID
		quizMap := make(map[string]*model.Quiz)
		missingIDs := []string{}

		for _, challengeID := range challengeIDs {
			cacheKey := cache.QuizByChallengeKey(challengeID)
			if cached, ok := c.Get(cacheKey); ok {
				if quiz, ok := cached.(*model.Quiz); ok {
					quizMap[challengeID] = quiz
					continue
				}
			}
			missingIDs = append(missingIDs, challengeID)
		}

		// Record cache statistics
		otel.RecordCacheHitMiss(span, len(challengeIDs)-len(missingIDs), len(missingIDs))

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizzesByChallengeIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[*model.Quiz], len(challengeIDs))
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

				// Store in cache by challenge ID
				c.Set(cache.QuizByChallengeKey(row.ChallengeID), quiz)
				// Also store by quiz ID for consistency
				c.Set(cache.QuizKey(row.ID), quiz)
				quizMap[row.ChallengeID] = quiz
			}
		}

		// Return results in the same order as input challenge IDs
		results := make([]*dataloader.Result[*model.Quiz], len(challengeIDs))
		for i, challengeID := range challengeIDs {
			if quiz, ok := quizMap[challengeID]; ok {
				results[i] = &dataloader.Result[*model.Quiz]{Data: quiz}
			} else {
				results[i] = &dataloader.Result[*model.Quiz]{Data: nil}
			}
		}
		return results
	}
}
