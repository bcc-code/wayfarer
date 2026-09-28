package api

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// getQuizAchievements returns the quiz achievement criteria rows for a quiz,
// cached under QuizAchievementsByQuizKey. The rows are static admin-managed
// data; quiz-achievement mutations and invalidateQuizLocal drop the key. The
// returned slice is shared across requests — read-only.
func (r *Resolver) getQuizAchievements(ctx context.Context, quizID string) ([]*sqlc.QuizAchievement, error) {
	cacheKey := cache.QuizAchievementsByQuizKey(quizID)
	if cached, ok := r.Cache.Get(cacheKey); ok {
		if rows, ok := cached.([]*sqlc.QuizAchievement); ok {
			return rows, nil
		}
	}

	rows, err := r.DB.Queries.GetQuizAchievementsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	r.Cache.Set(cacheKey, rows)
	return rows, nil
}
