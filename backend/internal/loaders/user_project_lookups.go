package loaders

import (
	"context"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// QuizSessionAccessTTL bounds how long cached quiz session access may lag behind
// session state changes that bypass invalidation (e.g. time-based transitions).
const QuizSessionAccessTTL = 45 * time.Second

// batchUserProjectLookup implements the shared shape of the per-(user, project)
// lookup loaders: check Ristretto per key, group cache misses by project, run
// one grouped query per project, cache every miss (including negative results)
// and return results in key order. Values are shared across requests — treat
// as read-only.
func batchUserProjectLookup[V any](
	ctx context.Context,
	c *cache.CacheWithRegistry,
	keys []UserProjectKey,
	cacheKey func(userID, projectID string) string,
	fetch func(ctx context.Context, projectID string, userIDs []string) (map[string]V, error),
	emptyValue func() V,
	store func(key string, value V),
) []*dataloader.Result[V] {
	results := make([]*dataloader.Result[V], len(keys))

	missUserIDsByProject := make(map[string][]string)
	seenMiss := make(map[UserProjectKey]bool)
	for i, key := range keys {
		if cached, ok := c.Get(cacheKey(key.UserID, key.ProjectID)); ok {
			if value, ok := cached.(V); ok {
				results[i] = &dataloader.Result[V]{Data: value}
				continue
			}
		}
		if !seenMiss[key] {
			seenMiss[key] = true
			missUserIDsByProject[key.ProjectID] = append(missUserIDsByProject[key.ProjectID], key.UserID)
		}
	}

	type groupResult struct {
		valueByUser map[string]V
		err         error
	}
	groupByProject := make(map[string]*groupResult, len(missUserIDsByProject))
	for projectID, userIDs := range missUserIDsByProject {
		valueByUser, err := fetch(ctx, projectID, userIDs)
		groupByProject[projectID] = &groupResult{valueByUser: valueByUser, err: err}
	}

	for i, key := range keys {
		if results[i] != nil {
			continue
		}
		group := groupByProject[key.ProjectID]
		if group.err != nil {
			results[i] = &dataloader.Result[V]{Error: group.err}
			continue
		}
		value, ok := group.valueByUser[key.UserID]
		if !ok {
			value = emptyValue()
		}
		store(cacheKey(key.UserID, key.ProjectID), value)
		results[i] = &dataloader.Result[V]{Data: value}
	}
	return results
}

// userTeamIDInProjectBatchFunc batches user→team membership lookups, grouped by
// project. The team ID is cached per (user, project) under
// cache.UserTeamInProjectKey including the negative result ("" = no team), so
// the key/value shape and InvalidateUser semantics match the previous
// direct-query path in the myTeam resolver.
func userTeamIDInProjectBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserProjectKey) []*dataloader.Result[string] {
	return func(ctx context.Context, keys []UserProjectKey) []*dataloader.Result[string] {
		return batchUserProjectLookup(ctx, c, keys,
			cache.UserTeamInProjectKey,
			func(ctx context.Context, projectID string, userIDs []string) (map[string]string, error) {
				rows, err := db.Queries.GetUserTeamsByProjectIDBulk(ctx, sqlc.GetUserTeamsByProjectIDBulkParams{
					Userids:   userIDs,
					Projectid: projectID,
				})
				if err != nil {
					return nil, err
				}
				teamByUser := make(map[string]string, len(rows))
				for _, row := range rows {
					teamByUser[row.UserID] = row.TeamID
				}
				return teamByUser, nil
			},
			func() string { return "" },
			func(key, teamID string) { c.Set(key, teamID) },
		)
	}
}

// userEnrolledChallengeIDsBatchFunc batches enrolled-challenge lookups, grouped
// by project. The ID set is cached per (user, project) under
// cache.UserEnrolledChallengesKey including empty sets; enroll/unenroll
// mutations invalidate via InvalidateUser and InvalidateChallenge.
func userEnrolledChallengeIDsBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserProjectKey) []*dataloader.Result[map[string]bool] {
	return func(ctx context.Context, keys []UserProjectKey) []*dataloader.Result[map[string]bool] {
		return batchUserProjectLookup(ctx, c, keys,
			cache.UserEnrolledChallengesKey,
			func(ctx context.Context, projectID string, userIDs []string) (map[string]map[string]bool, error) {
				rows, err := db.Queries.GetUsersEnrolledChallengeIDsInProject(ctx, sqlc.GetUsersEnrolledChallengeIDsInProjectParams{
					Userids:   userIDs,
					Projectid: projectID,
				})
				if err != nil {
					return nil, err
				}
				idsByUser := make(map[string]map[string]bool)
				for _, row := range rows {
					if idsByUser[row.UserID] == nil {
						idsByUser[row.UserID] = make(map[string]bool)
					}
					idsByUser[row.UserID][row.ChallengeID] = true
				}
				return idsByUser, nil
			},
			func() map[string]bool { return map[string]bool{} },
			func(key string, ids map[string]bool) { c.Set(key, ids) },
		)
	}
}

// userAccessibleQuizIDsBatchFunc batches quiz session-access lookups, grouped
// by project. A quiz ID's presence in the map means the user has access to a
// visible session; the value is true only when at least one of those sessions
// is still live (OPEN or LOCKED), false when all are FINISHED.
// The accessible quiz ID set is computed project-wide and cached
// per (user, project) under cache.UserQuizSessionAccessKey with a short TTL
// (session states also change on a schedule, without invalidation). Explicit
// invalidation stays with InvalidateQuizSessionAccess / invalidateQuizLocal.
func userAccessibleQuizIDsBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []UserProjectKey) []*dataloader.Result[map[string]bool] {
	return func(ctx context.Context, keys []UserProjectKey) []*dataloader.Result[map[string]bool] {
		return batchUserProjectLookup(ctx, c, keys,
			cache.UserQuizSessionAccessKey,
			func(ctx context.Context, projectID string, userIDs []string) (map[string]map[string]bool, error) {
				rows, err := db.Queries.GetBulkUsersSessionAccessQuizIDsByProject(ctx, sqlc.GetBulkUsersSessionAccessQuizIDsByProjectParams{
					Projectid: projectID,
					Userids:   userIDs,
				})
				if err != nil {
					return nil, err
				}
				idsByUser := make(map[string]map[string]bool)
				for _, row := range rows {
					if idsByUser[row.UserID] == nil {
						idsByUser[row.UserID] = make(map[string]bool)
					}
					idsByUser[row.UserID][row.QuizID] = row.HasLiveSession
				}
				return idsByUser, nil
			},
			func() map[string]bool { return map[string]bool{} },
			func(key string, ids map[string]bool) { c.SetWithTTL(key, ids, QuizSessionAccessTTL) },
		)
	}
}
