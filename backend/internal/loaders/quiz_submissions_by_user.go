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

// quizSubmissionsByUserBatchFunc batches loading quiz submissions by user IDs
func quizSubmissionsByUserBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[[]*model.QuizSubmission] {
	return func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.QuizSubmission] {
		ctx, span := otel.StartDataloaderSpan(ctx, "QuizSubmissionsByUser", len(userIDs))
		defer span.End()

		// Check cache first for each user ID
		submissionsMap := make(map[string][]*model.QuizSubmission)
		missingIDs := []string{}

		for _, userID := range userIDs {
			cacheKey := cache.QuizSubmissionsByUserKey(userID)
			if cached, ok := c.Get(cacheKey); ok {
				if submissions, ok := cached.([]*model.QuizSubmission); ok {
					submissionsMap[userID] = submissions
					continue
				}
			}
			missingIDs = append(missingIDs, userID)
		}

		// Record cache statistics
		otel.RecordCacheHitMiss(span, len(userIDs)-len(missingIDs), len(missingIDs))

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetQuizSubmissionsByUserIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[[]*model.QuizSubmission], len(userIDs))
				for i := range results {
					results[i] = &dataloader.Result[[]*model.QuizSubmission]{Error: err}
				}
				return results
			}

			// Group submissions by user ID
			for _, row := range rows {
				var completedAt *scalars.DateTime
				if row.CompletedAt.Valid {
					completedAt = &scalars.DateTime{Time: row.CompletedAt.Time}
				}

				var expiresAt *scalars.DateTime
				if row.ExpiresAt.Valid {
					expiresAt = &scalars.DateTime{Time: row.ExpiresAt.Time}
				}

				var score, maxScore, pointsAwarded *int
				if row.Score != nil {
					s := int(*row.Score)
					score = &s
				}
				if row.MaxScore != nil {
					ms := int(*row.MaxScore)
					maxScore = &ms
				}
				if row.PointsAwarded != nil {
					pa := int(*row.PointsAwarded)
					pointsAwarded = &pa
				}

				// Parse question order JSON
				var questionOrder []string
				if err := json.Unmarshal(row.QuestionOrder, &questionOrder); err != nil {
					// Skip malformed submissions
					continue
				}

				submission := &model.QuizSubmission{
					ID:            row.ID,
					StartedAt:     scalars.DateTime{Time: row.StartedAt.Time},
					CompletedAt:   completedAt,
					ExpiresAt:     expiresAt,
					QuestionOrder: questionOrder,
					Score:         score,
					MaxScore:      maxScore,
					PointsAwarded: pointsAwarded,
					AutoSubmitted: row.AutoSubmitted,
					// Fields for resolvers
					QuizID:    row.QuizID,
					SessionID: row.SessionID,
					UserID:    row.UserID,
				}

				submissionsMap[row.UserID] = append(submissionsMap[row.UserID], submission)
			}

			// Store in cache
			for userID, submissions := range submissionsMap {
				c.Set(cache.QuizSubmissionsByUserKey(userID), submissions)
			}
		}

		// Return results in the same order as input IDs
		results := make([]*dataloader.Result[[]*model.QuizSubmission], len(userIDs))
		for i, userID := range userIDs {
			if submissions, ok := submissionsMap[userID]; ok {
				results[i] = &dataloader.Result[[]*model.QuizSubmission]{Data: submissions}
			} else {
				results[i] = &dataloader.Result[[]*model.QuizSubmission]{Data: []*model.QuizSubmission{}}
			}
		}
		return results
	}
}
