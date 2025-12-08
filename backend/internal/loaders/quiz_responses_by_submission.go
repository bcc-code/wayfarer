package loaders

import (
	"context"
	"encoding/json"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/graph-gophers/dataloader/v7"
)

// quizResponsesBySubmissionBatchFunc batches loading quiz responses by submission IDs
func quizResponsesBySubmissionBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.QuizResponse] {
	return func(ctx context.Context, submissionIDs []string) []*dataloader.Result[[]*model.QuizResponse] {
		// Check cache first for each submission ID
		responsesMap := make(map[string][]*model.QuizResponse)
		missingIDs := []string{}

		for _, submissionID := range submissionIDs {
			cacheKey := cache.QuizResponsesBySubmissionKey(submissionID)
			if cached, ok := c.Get(cacheKey); ok {
				if responses, ok := cached.([]*model.QuizResponse); ok {
					responsesMap[submissionID] = responses
					continue
				}
			}
			missingIDs = append(missingIDs, submissionID)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizResponsesBySubmissionIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[[]*model.QuizResponse], len(submissionIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.QuizResponse]{Error: err}
				}
				return results
			}

			// Group responses by submission ID
			for _, row := range rows {
				var selectedAnswerIDs []string
				if row.SelectedAnswerIds != nil {
					if err := json.Unmarshal(row.SelectedAnswerIds, &selectedAnswerIDs); err == nil {
						// Successfully parsed
					}
				}

				textResponse := row.TextResponse

				var numberResponse *float64
				if row.NumberResponse.Valid {
					val, _ := row.NumberResponse.Float64Value()
					fv := val.Float64
					numberResponse = &fv
				}

				var jsonResponse *string
				if row.JsonResponse != nil {
					jsonStr := string(row.JsonResponse)
					jsonResponse = &jsonStr
				}

				isCorrect := row.IsCorrect

				var answeredAt *scalars.DateTime
				if row.AnsweredAt.Valid {
					answeredAt = &scalars.DateTime{Time: row.AnsweredAt.Time}
				}

				var timeSpentSeconds *int
				if row.TimeSpentSeconds != nil {
					tss := int(*row.TimeSpentSeconds)
					timeSpentSeconds = &tss
				}

				response := &model.QuizResponse{
					ID:               row.ID,
					SelectedAnswerIds: selectedAnswerIDs,
					TextResponse:     textResponse,
					NumberResponse:   numberResponse,
					JSONResponse:     jsonResponse,
					IsCorrect:        isCorrect,
					AnsweredAt:       answeredAt,
					TimeSpentSeconds: timeSpentSeconds,
					// Fields for resolvers
					SubmissionID: row.SubmissionID,
					QuestionID:   row.QuestionID,
				}

				responsesMap[row.SubmissionID] = append(responsesMap[row.SubmissionID], response)
			}

			// Store in cache
			for submissionID, responses := range responsesMap {
				c.Set(cache.QuizResponsesBySubmissionKey(submissionID), responses)
			}
		}

		// Return results in the same order as input IDs
		results := make([]*dataloader.Result[[]*model.QuizResponse], len(submissionIDs))
		for i, submissionID := range submissionIDs {
			if responses, ok := responsesMap[submissionID]; ok {
				results[i] = &dataloader.Result[[]*model.QuizResponse]{Data: responses}
			} else {
				results[i] = &dataloader.Result[[]*model.QuizResponse]{Data: []*model.QuizResponse{}}
			}
		}
		return results
	}
}
