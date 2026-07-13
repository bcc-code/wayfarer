package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// UserQuizKey combines user ID and quiz ID for dataloader key
type UserQuizKey struct {
	UserID string
	QuizID string
}

func (k UserQuizKey) String() string {
	return fmt.Sprintf("%s:%s", k.UserID, k.QuizID)
}

func (k UserQuizKey) Raw() interface{} {
	return k
}

// userActiveQuizSessionBatchFunc batches lookups of a user's active (visible)
// session for a quiz, grouped by quiz. Results — including nil for "no active
// session" — are cached per (user, quiz) under cache.UserActiveQuizSessionKey
// with a short TTL; session mutations and invalidateQuizSessionAccessLocal
// drop the prefix. Cached sessions are shared across requests — read-only.
func userActiveQuizSessionBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserQuizKey) []*dataloader.Result[*sqlc.QuizSession] {
	return func(ctx context.Context, keys []UserQuizKey) []*dataloader.Result[*sqlc.QuizSession] {
		results := make([]*dataloader.Result[*sqlc.QuizSession], len(keys))

		missUserIDsByQuiz := make(map[string][]string)
		seenMiss := make(map[UserQuizKey]bool)
		for i, key := range keys {
			if cached, ok := c.Get(cache.UserActiveQuizSessionKey(key.UserID, key.QuizID)); ok {
				if session, ok := cached.(*sqlc.QuizSession); ok {
					results[i] = &dataloader.Result[*sqlc.QuizSession]{Data: session}
					continue
				}
			}
			if !seenMiss[key] {
				seenMiss[key] = true
				missUserIDsByQuiz[key.QuizID] = append(missUserIDsByQuiz[key.QuizID], key.UserID)
			}
		}

		type groupResult struct {
			sessionByUser map[string]*sqlc.QuizSession
			err           error
		}
		groupByQuiz := make(map[string]*groupResult, len(missUserIDsByQuiz))
		for quizID, userIDs := range missUserIDsByQuiz {
			rows, err := db.Queries.GetUsersActiveSessionForQuiz(ctx, sqlc.GetUsersActiveSessionForQuizParams{
				Quizid:  quizID,
				Userids: userIDs,
			})
			group := &groupResult{err: err}
			if err == nil {
				group.sessionByUser = make(map[string]*sqlc.QuizSession, len(rows))
				for _, row := range rows {
					group.sessionByUser[row.AccessUserID] = &sqlc.QuizSession{
						ID:        row.ID,
						QuizID:    row.QuizID,
						Name:      row.Name,
						State:     row.State,
						OpenAt:    row.OpenAt,
						LockAt:    row.LockAt,
						FinishAt:  row.FinishAt,
						CreatedBy: row.CreatedBy,
						CreatedAt: row.CreatedAt,
						UpdatedAt: row.UpdatedAt,
					}
				}
			}
			groupByQuiz[quizID] = group
		}

		for i, key := range keys {
			if results[i] != nil {
				continue
			}
			group := groupByQuiz[key.QuizID]
			if group.err != nil {
				results[i] = &dataloader.Result[*sqlc.QuizSession]{Error: group.err}
				continue
			}
			session := group.sessionByUser[key.UserID] // nil = no active session
			c.SetWithTTL(cache.UserActiveQuizSessionKey(key.UserID, key.QuizID), session, QuizSessionAccessTTL)
			results[i] = &dataloader.Result[*sqlc.QuizSession]{Data: session}
		}
		return results
	}
}
