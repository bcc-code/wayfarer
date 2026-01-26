package loaders

import (
	"context"
	"encoding/json"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/graph-gophers/dataloader/v7"
)

// quizResponsesBySubmissionBatchFunc batches loading quiz responses by submission IDs
func quizResponsesBySubmissionBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]model.QuizResponse] {
	return func(ctx context.Context, submissionIDs []string) []*dataloader.Result[[]model.QuizResponse] {
		ctx, span := otel.StartDataloaderSpan(ctx, "QuizResponsesBySubmission", len(submissionIDs))
		defer span.End()

		// Check cache first for each submission ID
		responsesMap := make(map[string][]model.QuizResponse)
		missingIDs := []string{}

		for _, submissionID := range submissionIDs {
			cacheKey := cache.QuizResponsesBySubmissionKey(submissionID)
			if cached, ok := c.Get(cacheKey); ok {
				if responses, ok := cached.([]model.QuizResponse); ok {
					responsesMap[submissionID] = responses
					continue
				}
			}
			missingIDs = append(missingIDs, submissionID)
		}

		// Record cache statistics
		otel.RecordCacheHitMiss(span, len(submissionIDs)-len(missingIDs), len(missingIDs))

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizResponsesBySubmissionIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[[]model.QuizResponse], len(submissionIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]model.QuizResponse]{Error: err}
				}
				return results
			}

			// Group responses by submission ID
			for _, row := range rows {
				var answeredAt *scalars.DateTime
				if row.AnsweredAt.Valid {
					answeredAt = &scalars.DateTime{Time: row.AnsweredAt.Time}
				}

				var timeSpentSeconds *int
				if row.TimeSpentSeconds != nil {
					tss := int(*row.TimeSpentSeconds)
					timeSpentSeconds = &tss
				}

				var response model.QuizResponse

				switch row.QuestionType {
				case "PREDEFINED":
					var selectedAnswerIDs []string
					if row.SelectedAnswerIds != nil {
						_ = json.Unmarshal(row.SelectedAnswerIds, &selectedAnswerIDs)
					}
					response = &model.PredefinedResponse{
						ID:                row.ID,
						SubmissionID:      row.SubmissionID,
						QuestionID:        row.QuestionID,
						SelectedAnswerIds: selectedAnswerIDs,
						IsCorrect:         row.IsCorrect,
						AnsweredAt:        answeredAt,
						TimeSpentSeconds:  timeSpentSeconds,
					}
				case "FREE_TEXT":
					textResponse := ""
					if row.TextResponse != nil {
						textResponse = *row.TextResponse
					}
					response = &model.FreeTextResponse{
						ID:               row.ID,
						SubmissionID:     row.SubmissionID,
						QuestionID:       row.QuestionID,
						TextResponse:     textResponse,
						AnsweredAt:       answeredAt,
						TimeSpentSeconds: timeSpentSeconds,
					}
				case "NUMBER":
					var numberResponse float64
					if row.NumberResponse.Valid {
						val, _ := row.NumberResponse.Float64Value()
						numberResponse = val.Float64
					}
					response = &model.NumberResponse{
						ID:               row.ID,
						SubmissionID:     row.SubmissionID,
						QuestionID:       row.QuestionID,
						NumberResponse:   numberResponse,
						AnsweredAt:       answeredAt,
						TimeSpentSeconds: timeSpentSeconds,
					}
				case "JSON":
					jsonResponse := ""
					if row.JsonResponse != nil {
						jsonResponse = string(row.JsonResponse)
					}
					response = &model.JSONResponse{
						ID:               row.ID,
						SubmissionID:     row.SubmissionID,
						QuestionID:       row.QuestionID,
						JSONResponse:     jsonResponse,
						AnsweredAt:       answeredAt,
						TimeSpentSeconds: timeSpentSeconds,
					}
				case "ORDERING":
					var submittedOrder []string
					if row.JsonResponse != nil {
						_ = json.Unmarshal(row.JsonResponse, &submittedOrder)
					}
					var pointsEarned *int
					if row.PointsEarned != nil {
						pe := int(*row.PointsEarned)
						pointsEarned = &pe
					}
					response = &model.OrderingResponse{
						ID:               row.ID,
						SubmissionID:     row.SubmissionID,
						QuestionID:       row.QuestionID,
						SubmittedOrder:   submittedOrder,
						IsCorrect:        row.IsCorrect,
						AnsweredAt:       answeredAt,
						TimeSpentSeconds: timeSpentSeconds,
						PointsEarned:     pointsEarned,
					}
				default:
					// Default to FreeTextResponse for unknown types
					textResponse := ""
					if row.TextResponse != nil {
						textResponse = *row.TextResponse
					}
					response = &model.FreeTextResponse{
						ID:               row.ID,
						SubmissionID:     row.SubmissionID,
						QuestionID:       row.QuestionID,
						TextResponse:     textResponse,
						AnsweredAt:       answeredAt,
						TimeSpentSeconds: timeSpentSeconds,
					}
				}

				responsesMap[row.SubmissionID] = append(responsesMap[row.SubmissionID], response)
			}

			// Store in cache
			for submissionID, responses := range responsesMap {
				c.Set(cache.QuizResponsesBySubmissionKey(submissionID), responses)
			}
		}

		// Return results in the same order as input IDs
		results := make([]*dataloader.Result[[]model.QuizResponse], len(submissionIDs))
		for i, submissionID := range submissionIDs {
			if responses, ok := responsesMap[submissionID]; ok {
				results[i] = &dataloader.Result[[]model.QuizResponse]{Data: responses}
			} else {
				results[i] = &dataloader.Result[[]model.QuizResponse]{Data: []model.QuizResponse{}}
			}
		}
		return results
	}
}
