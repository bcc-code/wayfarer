package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Self-enrollment used to call InvalidateChallenge, which falls back to
// DeletePrefix over five per-user prefixes and therefore discarded these caches
// for *every* user. Under a spike of concurrent self-enrollments that meant the
// enrollment and quiz-access caches never survived long enough to serve the
// ChallengePage request that immediately follows. These tests pin the narrow
// behaviour: the enrolling user's keys go, everyone else's stay.
func TestInvalidateUserChallengeEnrollment_DropsOnlyThatUsersKeys(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	const (
		me        = "US00000000000000000000000001"
		other     = "US00000000000000000000000002"
		projectID = "PR00000000000000000000000001"
		challenge = "CL00000000000000000000000001"
	)

	// Keys for the enrolling user, all five that an enrollment can affect.
	mine := []string{
		UserChallengeEnrollmentKey(me, challenge),
		UserChallengeCompletionKey(me, challenge),
		UserEnrolledChallengesKey(me, projectID),
		UserQuizSessionAccessKey(me, projectID),
		ActiveChallengesCountKey(me, projectID),
	}
	// The same shapes for a bystander, which must survive.
	theirs := []string{
		UserChallengeEnrollmentKey(other, challenge),
		UserChallengeCompletionKey(other, challenge),
		UserEnrolledChallengesKey(other, projectID),
		UserQuizSessionAccessKey(other, projectID),
		ActiveChallengesCountKey(other, projectID),
	}

	for _, k := range append(append([]string{}, mine...), theirs...) {
		c.Set(k, "cached")
	}
	c.Wait()

	for _, k := range mine {
		_, ok := c.Get(k)
		require.True(t, ok, "precondition: %s should be cached", k)
	}

	c.InvalidateUserChallengeEnrollment(me, projectID, challenge)
	c.Wait()

	for _, k := range mine {
		_, ok := c.Get(k)
		assert.False(t, ok, "expected %s to be invalidated", k)
	}
	for _, k := range theirs {
		_, ok := c.Get(k)
		assert.True(t, ok, "another user's %s must NOT be invalidated", k)
	}
}

func TestInvalidateUserChallengeEnrollment_LeavesUnrelatedCachesIntact(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	const (
		me        = "US00000000000000000000000001"
		projectID = "PR00000000000000000000000001"
		challenge = "CL00000000000000000000000001"
	)

	// Caches that a self-enrollment has no business touching.
	unrelated := map[string]string{
		ChallengeKey(challenge):           "challenge",
		ChallengesByProjectKey(projectID): "list",
		ProjectKey(projectID):             "project",
		UserKey(me):                       "user",
		FullLeaderboardKey("project", projectID, "persons", map[string]string{}): "board",
	}
	for k, v := range unrelated {
		c.Set(k, v)
	}
	c.Wait()

	c.InvalidateUserChallengeEnrollment(me, projectID, challenge)
	c.Wait()

	for k := range unrelated {
		_, ok := c.Get(k)
		assert.True(t, ok, "%s must survive a self-enrollment invalidation", k)
	}
}
