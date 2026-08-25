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

// The whole-response cache stores per-user entries under
// gqlresponse:user:{userID}: — an enrollment must drop the enrolling user's
// entries (their cached ActiveChallengesPage embeds pre-enroll state) while a
// bystander's entries and shared entries survive.
func TestInvalidateUserChallengeEnrollment_DropsUsersResponseEntries(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	const (
		me        = "US00000000000000000000000001"
		other     = "US00000000000000000000000002"
		projectID = "PR00000000000000000000000001"
		challenge = "CL00000000000000000000000001"
	)

	mine := GQLResponseUserPrefix(me) + "qhash:vhash:en"
	theirs := GQLResponseUserPrefix(other) + "qhash:vhash:en"
	shared := PrefixGQLResponseShared + "qhash:vhash:en"
	for _, k := range []string{mine, theirs, shared} {
		require.True(t, c.Set(k, "cached"))
	}
	c.Wait()

	c.InvalidateUserChallengeEnrollment(me, projectID, challenge)
	c.Wait()

	_, ok := c.Get(mine)
	assert.False(t, ok, "enrolling user's response entry must be dropped")
	_, ok = c.Get(theirs)
	assert.True(t, ok, "bystander's response entry must survive")
	_, ok = c.Get(shared)
	assert.True(t, ok, "shared response entry must survive")
}

// invalidateUserLocal (and thus InvalidateUser broadcasts) must also clear the
// user's response entries, and invalidateProjectLocal clears every response
// entry so admin project edits surface immediately.
func TestResponseEntryInvalidation_UserAndProject(t *testing.T) {
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	defer c.Close()

	const me = "US00000000000000000000000001"
	userEntry := GQLResponseUserPrefix(me) + "qhash:vhash:en"
	sharedEntry := PrefixGQLResponseShared + "qhash:vhash:en"
	require.True(t, c.Set(userEntry, "cached"))
	require.True(t, c.Set(sharedEntry, "cached"))
	c.Wait()

	c.invalidateUserLocal(me)
	c.Wait()
	_, ok := c.Get(userEntry)
	assert.False(t, ok, "user invalidation must drop the user's response entries")
	_, ok = c.Get(sharedEntry)
	assert.True(t, ok, "user invalidation must not drop shared entries")

	require.True(t, c.Set(userEntry, "cached"))
	c.Wait()
	c.invalidateProjectLocal("PR00000000000000000000000001")
	c.Wait()
	_, ok = c.Get(userEntry)
	assert.False(t, ok, "project invalidation must drop all response entries")
	_, ok = c.Get(sharedEntry)
	assert.False(t, ok, "project invalidation must drop shared entries")
}
