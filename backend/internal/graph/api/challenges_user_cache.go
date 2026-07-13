package api

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/loaders"
)

// getUserEnrolledChallengeIDs returns the set of challenge IDs the user is
// enrolled in within the project, batched per project through
// UserEnrolledChallengeIDsLoader and cached per (user, project) including the
// empty result. Enroll/unenroll mutations invalidate via InvalidateUser and
// InvalidateChallenge. The returned map is shared across requests — read-only.
func (r *Resolver) getUserEnrolledChallengeIDs(ctx context.Context, userID, projectID string) (map[string]bool, error) {
	return r.Loaders.UserEnrolledChallengeIDsLoader.Load(ctx, loaders.UserProjectKey{
		UserID:    userID,
		ProjectID: projectID,
	})()
}

// getUserAccessibleQuizIDs returns which quiz IDs in the project the user has
// session access to, batched per project through UserAccessibleQuizIDsLoader
// and cached per (user, project) with a short TTL. The set is computed
// project-wide, so it covers any quiz subset callers filter against (the
// quizIDs parameter only signals whether there is anything to check). The
// returned map is shared across requests — read-only.
func (r *Resolver) getUserAccessibleQuizIDs(ctx context.Context, userID, projectID string, quizIDs []string) (map[string]bool, error) {
	if len(quizIDs) == 0 {
		return map[string]bool{}, nil
	}
	return r.Loaders.UserAccessibleQuizIDsLoader.Load(ctx, loaders.UserProjectKey{
		UserID:    userID,
		ProjectID: projectID,
	})()
}
