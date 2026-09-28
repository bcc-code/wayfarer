package api

import (
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// buildQuizFilterParamsCursor converts GraphQL filter and cursor pagination params to database query parameters
func buildQuizFilterParamsCursor(filter *model.QuizFilter, first *int, after *string, last *int, before *string) (sqlc.GetQuizzesFilteredCursorParams, error) {
	params := sqlc.GetQuizzesFilteredCursorParams{}

	// Apply filters if provided
	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.ChallengeID != nil {
			params.Challengeid = *filter.ChallengeID
		}
	}

	// Handle cursor pagination
	isBackward := false
	var limit int

	if first != nil && last != nil {
		return params, fmt.Errorf("cannot specify both first and last")
	}

	if first != nil {
		limit = *first + 1 // Fetch one extra to determine hasNextPage
		isBackward = false
	} else if last != nil {
		limit = *last + 1 // Fetch one extra to determine hasPreviousPage
		isBackward = true
	} else {
		// Default page size
		limit = 11 // 10 items + 1 to check for next page
		isBackward = false
	}

	params.Querylimit = int32(limit)
	params.Isbackward = isBackward

	// Set cursors
	if after != nil && *after != "" {
		params.Aftercursor = *after
	}

	if before != nil && *before != "" {
		params.Beforecursor = *before
	}

	return params, nil
}

// buildCountQuizzesFilterParams converts GraphQL filter to database count query parameters
func buildCountQuizzesFilterParams(filter *model.QuizFilter) sqlc.CountQuizzesFilteredParams {
	params := sqlc.CountQuizzesFilteredParams{}

	if filter != nil {
		if filter.Ids != nil {
			params.Ids = filter.Ids
		}

		if filter.ProjectID != nil {
			params.Projectid = *filter.ProjectID
		}

		if filter.ChallengeID != nil {
			params.Challengeid = *filter.ChallengeID
		}
	}

	return params
}

// buildQuizCacheKeyParams converts filter and pagination parameters to a map for cache key generation
func buildQuizCacheKeyParams(filter *model.QuizFilter, first *int, after *string, last *int, before *string) map[string]string {
	params := make(map[string]string)

	// Add filter parameters
	if filter != nil {
		if len(filter.Ids) > 0 {
			params["ids"] = fmt.Sprintf("%v", filter.Ids)
		}
		if filter.ProjectID != nil {
			params["projectid"] = *filter.ProjectID
		}
		if filter.ChallengeID != nil {
			params["challengeid"] = *filter.ChallengeID
		}
	}

	// Add pagination parameters
	if first != nil {
		params["first"] = fmt.Sprintf("%d", *first)
	}
	if after != nil && *after != "" {
		params["after"] = *after
	}
	if last != nil {
		params["last"] = fmt.Sprintf("%d", *last)
	}
	if before != nil && *before != "" {
		params["before"] = *before
	}

	return params
}

// quizReorderStep is one question's new position within a reorder pass.
type quizReorderStep struct {
	ID    string
	Order int32
}

// quizReorderPasses builds the two passes a reorder needs.
//
// `quiz_questions` has UNIQUE (quiz_id, question_order), so writing final
// positions directly collides the moment two questions swap: the first UPDATE
// takes an order the second still holds. The first pass parks every question on
// a negative order — no real position uses those — which leaves the second pass
// free to assign 1..n in any arrangement.
func quizReorderPasses(questionIDs []string) [][]quizReorderStep {
	if len(questionIDs) == 0 {
		return nil
	}

	park := make([]quizReorderStep, len(questionIDs))
	final := make([]quizReorderStep, len(questionIDs))
	for i, id := range questionIDs {
		park[i] = quizReorderStep{ID: id, Order: int32(-(i + 1))}
		final[i] = quizReorderStep{ID: id, Order: int32(i + 1)}
	}

	return [][]quizReorderStep{park, final}
}
