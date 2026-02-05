package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/graph-gophers/dataloader/v7"
)

// quizQuestionsByQuizBatchFunc batches loading quiz questions by quiz IDs
func quizQuestionsByQuizBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.QuizQuestion] {
	return func(ctx context.Context, quizIDs []string) []*dataloader.Result[[]model.QuizQuestion] {
		ctx, span := otel.StartDataloaderSpan(ctx, "QuizQuestionsByQuiz", len(quizIDs))
		defer span.End()

		// Check cache first for each quiz ID
		questionsMap := make(map[string][]model.QuizQuestion)
		missingIDs := []string{}

		for _, quizID := range quizIDs {
			cacheKey := cache.QuizQuestionsByQuizKey(quizID)
			if cached, ok := c.Get(cacheKey); ok {
				if questions, ok := cached.([]model.QuizQuestion); ok {
					questionsMap[quizID] = questions
					continue
				}
			}
			missingIDs = append(missingIDs, quizID)
		}

		// Record cache statistics
		otel.RecordCacheHitMiss(span, len(quizIDs)-len(missingIDs), len(missingIDs))

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizQuestionsByQuizIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[[]model.QuizQuestion], len(quizIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.QuizQuestion]{Error: err}
				}
				return results
			}

			// Group questions by quiz ID
			for _, row := range rows {
				var question model.QuizQuestion

				// Convert points from *int32 to *int
				var points *int
				if row.Points != nil {
					p := int(*row.Points)
					points = &p
				}

				// Convert timeout_seconds from *int32 to *int
				var timeoutSeconds *int
				if row.TimeoutSeconds != nil {
					ts := int(*row.TimeoutSeconds)
					timeoutSeconds = &ts
				}

				switch row.QuestionType {
				case "PREDEFINED":
					allowMultiple := false
					if row.AllowMultipleSelection != nil {
						allowMultiple = *row.AllowMultipleSelection
					}
					question = &model.PredefinedQuestion{
						ID:                     row.ID,
						QuestionText:           row.QuestionText,
						QuestionOrder:          int(row.QuestionOrder),
						AllowMultipleSelection: allowMultiple,
						QuizID:                 row.QuizID,
						Points:                 points,
						TimeoutSeconds:         timeoutSeconds,
					}
				case "FREE_TEXT":
					question = &model.FreeTextQuestion{
						ID:             row.ID,
						QuestionText:   row.QuestionText,
						QuestionOrder:  int(row.QuestionOrder),
						QuizID:         row.QuizID,
						Points:         points,
						TimeoutSeconds: timeoutSeconds,
					}
				case "NUMBER":
					var minValue, maxValue, stepValue *float64
					if row.MinValue.Valid {
						val, _ := row.MinValue.Float64Value()
						fv := val.Float64
						minValue = &fv
					}
					if row.MaxValue.Valid {
						val, _ := row.MaxValue.Float64Value()
						fv := val.Float64
						maxValue = &fv
					}
					if row.StepValue.Valid {
						val, _ := row.StepValue.Float64Value()
						fv := val.Float64
						stepValue = &fv
					}
					question = &model.NumberQuestion{
						ID:             row.ID,
						QuestionText:   row.QuestionText,
						QuestionOrder:  int(row.QuestionOrder),
						MinValue:       minValue,
						MaxValue:       maxValue,
						StepValue:      stepValue,
						QuizID:         row.QuizID,
						Points:         points,
						TimeoutSeconds: timeoutSeconds,
					}
				case "JSON":
					question = &model.JSONQuestion{
						ID:             row.ID,
						QuestionText:   row.QuestionText,
						QuestionOrder:  int(row.QuestionOrder),
						QuizID:         row.QuizID,
						Points:         points,
						TimeoutSeconds: timeoutSeconds,
					}
				case "ORDERING":
					question = &model.OrderingQuestion{
						ID:             row.ID,
						QuestionText:   row.QuestionText,
						QuestionOrder:  int(row.QuestionOrder),
						QuizID:         row.QuizID,
						Points:         points,
						TimeoutSeconds: timeoutSeconds,
					}
				default:
					// Default to FreeTextQuestion for unknown types
					question = &model.FreeTextQuestion{
						ID:             row.ID,
						QuestionText:   row.QuestionText,
						QuestionOrder:  int(row.QuestionOrder),
						QuizID:         row.QuizID,
						Points:         points,
						TimeoutSeconds: timeoutSeconds,
					}
				}

				questionsMap[row.QuizID] = append(questionsMap[row.QuizID], question)
			}

			// Store in cache (questions are already ordered by question_order in SQL)
			for quizID, questions := range questionsMap {
				c.Set(cache.QuizQuestionsByQuizKey(quizID), questions)
			}
		}

		// Return results in the same order as input IDs
		results := make([]*dataloader.Result[[]model.QuizQuestion], len(quizIDs))
		for i, quizID := range quizIDs {
			if questions, ok := questionsMap[quizID]; ok {
				results[i] = &dataloader.Result[[]model.QuizQuestion]{Data: questions}
			} else {
				results[i] = &dataloader.Result[[]model.QuizQuestion]{Data: []model.QuizQuestion{}}
			}
		}
		return results
	}
}
