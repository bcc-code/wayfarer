package api

import (
	"context"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// quizSessionAccessTTL bounds how long cached quiz session access may lag behind
// session state changes that bypass invalidation (e.g. time-based transitions).
const quizSessionAccessTTL = 45 * time.Second

// cachedQuizAccess is the cached result of a bulk session-access check.
// Covered holds every quiz ID the check included; Accessible the subset the
// user has session access to. Both maps are shared across requests — treat as
// read-only.
type cachedQuizAccess struct {
	Covered    map[string]bool
	Accessible map[string]bool
}

// getUserEnrolledChallengeIDs returns the set of challenge IDs the user is
// enrolled in within the project, cached per (user, project) including the
// empty result. Enroll/unenroll mutations invalidate via InvalidateUser and
// InvalidateChallenge. The returned map is shared across requests — read-only.
func (r *Resolver) getUserEnrolledChallengeIDs(ctx context.Context, userID, projectID string) (map[string]bool, error) {
	cacheKey := cache.UserEnrolledChallengesKey(userID, projectID)
	if cached, ok := r.Cache.Get(cacheKey); ok {
		if ids, ok := cached.(map[string]bool); ok {
			return ids, nil
		}
	}

	enrolledIDs, err := r.DB.Queries.GetUserEnrolledChallengeIDsInProject(ctx, sqlc.GetUserEnrolledChallengeIDsInProjectParams{
		Userid:    userID,
		Projectid: projectID,
	})
	if err != nil {
		return nil, err
	}

	ids := make(map[string]bool, len(enrolledIDs))
	for _, id := range enrolledIDs {
		ids[id] = true
	}
	r.Cache.Set(cacheKey, ids)
	return ids, nil
}

// getUserAccessibleQuizIDs returns which of the given quiz IDs the user has
// session access to, cached per (user, project) with a short TTL. A cached
// entry is only used when it covers every requested quiz ID, so callers with
// different quiz sets (project-wide vs event-scoped) can share the key. The
// returned map is shared across requests — read-only.
func (r *Resolver) getUserAccessibleQuizIDs(ctx context.Context, userID, projectID string, quizIDs []string) (map[string]bool, error) {
	if len(quizIDs) == 0 {
		return map[string]bool{}, nil
	}

	cacheKey := cache.UserQuizSessionAccessKey(userID, projectID)
	if cached, ok := r.Cache.Get(cacheKey); ok {
		if access, ok := cached.(*cachedQuizAccess); ok {
			covered := true
			for _, qid := range quizIDs {
				if !access.Covered[qid] {
					covered = false
					break
				}
			}
			if covered {
				return access.Accessible, nil
			}
		}
	}

	accessibleQuizIDs, err := r.DB.Queries.GetBulkUserSessionAccessQuizIDs(ctx, sqlc.GetBulkUserSessionAccessQuizIDsParams{
		Quizids: quizIDs,
		Userid:  userID,
	})
	if err != nil {
		return nil, err
	}

	access := &cachedQuizAccess{
		Covered:    make(map[string]bool, len(quizIDs)),
		Accessible: make(map[string]bool, len(accessibleQuizIDs)),
	}
	for _, qid := range quizIDs {
		access.Covered[qid] = true
	}
	for _, qid := range accessibleQuizIDs {
		access.Accessible[qid] = true
	}
	r.Cache.SetWithTTL(cacheKey, access, quizSessionAccessTTL)
	return access.Accessible, nil
}
