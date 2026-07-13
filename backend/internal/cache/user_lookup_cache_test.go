package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserA    = "US01ARZ3NDEKTSV4RRFFQ69G5FAA"
	testUserB    = "US01ARZ3NDEKTSV4RRFFQ69G5FAB"
	testProject  = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testTeamID   = "TM01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testQuizID   = "QZ01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testChalleng = "CL01ARZ3NDEKTSV4RRFFQ69G5FAV"
)

func newLookupTestCache(t *testing.T) *CacheWithRegistry {
	t.Helper()
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	return c
}

// setUserLookupKeys populates all three per-user lookup caches for a user
func setUserLookupKeys(t *testing.T, c *CacheWithRegistry, userID string) {
	t.Helper()
	require.True(t, c.Set(UserTeamInProjectKey(userID, testProject), testTeamID))
	require.True(t, c.Set(UserEnrolledChallengesKey(userID, testProject), map[string]bool{testChalleng: true}))
	require.True(t, c.Set(UserQuizSessionAccessKey(userID, testProject), map[string]bool{testQuizID: true}))
	c.Wait()
}

func TestInvalidateUserRemovesPerUserLookupCaches(t *testing.T) {
	c := newLookupTestCache(t)
	setUserLookupKeys(t, c, testUserA)
	setUserLookupKeys(t, c, testUserB)

	c.InvalidateUser(testUserA)

	_, ok := c.Get(UserTeamInProjectKey(testUserA, testProject))
	assert.False(t, ok, "user A team lookup should be invalidated")
	_, ok = c.Get(UserEnrolledChallengesKey(testUserA, testProject))
	assert.False(t, ok, "user A enrolled challenges should be invalidated")
	_, ok = c.Get(UserQuizSessionAccessKey(testUserA, testProject))
	assert.False(t, ok, "user A quiz session access should be invalidated")

	// Other users' entries must survive
	_, ok = c.Get(UserTeamInProjectKey(testUserB, testProject))
	assert.True(t, ok, "user B team lookup should remain")
	_, ok = c.Get(UserEnrolledChallengesKey(testUserB, testProject))
	assert.True(t, ok, "user B enrolled challenges should remain")
	_, ok = c.Get(UserQuizSessionAccessKey(testUserB, testProject))
	assert.True(t, ok, "user B quiz session access should remain")
}

func TestInvalidateChallengeRemovesEnrollmentAndQuizAccessCaches(t *testing.T) {
	c := newLookupTestCache(t)
	setUserLookupKeys(t, c, testUserA)

	c.InvalidateChallenge(testChalleng, testProject, nil)

	_, ok := c.Get(UserEnrolledChallengesKey(testUserA, testProject))
	assert.False(t, ok, "enrolled challenges should be invalidated on challenge change")
	_, ok = c.Get(UserQuizSessionAccessKey(testUserA, testProject))
	assert.False(t, ok, "quiz session access should be invalidated on challenge change")

	// Team membership is unrelated to challenge changes
	_, ok = c.Get(UserTeamInProjectKey(testUserA, testProject))
	assert.True(t, ok, "team lookup should remain on challenge change")
}

func TestInvalidateQuizSessionAccess(t *testing.T) {
	c := newLookupTestCache(t)
	setUserLookupKeys(t, c, testUserA)
	setUserLookupKeys(t, c, testUserB)

	c.InvalidateQuizSessionAccess()

	_, ok := c.Get(UserQuizSessionAccessKey(testUserA, testProject))
	assert.False(t, ok, "quiz session access should be invalidated for user A")
	_, ok = c.Get(UserQuizSessionAccessKey(testUserB, testProject))
	assert.False(t, ok, "quiz session access should be invalidated for user B")

	// Unrelated per-user caches must survive
	_, ok = c.Get(UserTeamInProjectKey(testUserA, testProject))
	assert.True(t, ok, "team lookup should remain")
	_, ok = c.Get(UserEnrolledChallengesKey(testUserA, testProject))
	assert.True(t, ok, "enrolled challenges should remain")
}

func TestInvalidateQuizRemovesQuizSessionAccessCaches(t *testing.T) {
	c := newLookupTestCache(t)
	setUserLookupKeys(t, c, testUserA)

	c.InvalidateQuiz(testQuizID)

	_, ok := c.Get(UserQuizSessionAccessKey(testUserA, testProject))
	assert.False(t, ok, "quiz session access should be invalidated on quiz change")
}
