package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// quizAnswersByQuestionBatchFunc batches loading predefined answers by question IDs
func quizAnswersByQuestionBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.QuizPredefinedAnswer] {
	return func(ctx context.Context, questionIDs []string) []*dataloader.Result[[]*model.QuizPredefinedAnswer] {
		// Check cache first for each question ID
		answersMap := make(map[string][]*model.QuizPredefinedAnswer)
		missingIDs := []string{}

		for _, questionID := range questionIDs {
			cacheKey := cache.QuizAnswersByQuestionKey(questionID)
			if cached, ok := c.Get(cacheKey); ok {
				if answers, ok := cached.([]*model.QuizPredefinedAnswer); ok {
					answersMap[questionID] = answers
					continue
				}
			}
			missingIDs = append(missingIDs, questionID)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetPredefinedAnswersByQuestionIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[[]*model.QuizPredefinedAnswer], len(questionIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.QuizPredefinedAnswer]{Error: err}
				}
				return results
			}

			// Group answers by question ID
			for _, row := range rows {
				answer := &model.QuizPredefinedAnswer{
					ID:          row.ID,
					AnswerText:  row.AnswerText,
					AnswerOrder: int(row.AnswerOrder),
					// IsCorrect is handled by field resolver that checks reveal_correct_answers
					// Fields for resolvers
					QuestionID:     row.QuestionID,
					IsCorrectValue: row.IsCorrect,
				}

				answersMap[row.QuestionID] = append(answersMap[row.QuestionID], answer)
			}

			// Store in cache (answers are already ordered by answer_order in SQL)
			for questionID, answers := range answersMap {
				c.Set(cache.QuizAnswersByQuestionKey(questionID), answers)
			}
		}

		// Return results in the same order as input IDs
		results := make([]*dataloader.Result[[]*model.QuizPredefinedAnswer], len(questionIDs))
		for i, questionID := range questionIDs {
			if answers, ok := answersMap[questionID]; ok {
				results[i] = &dataloader.Result[[]*model.QuizPredefinedAnswer]{Data: answers}
			} else {
				results[i] = &dataloader.Result[[]*model.QuizPredefinedAnswer]{Data: []*model.QuizPredefinedAnswer{}}
			}
		}
		return results
	}
}
