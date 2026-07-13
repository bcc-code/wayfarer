package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserA     = "US01ARZ3NDEKTSV4RRFFQ69G5FAA"
	testUserB     = "US01ARZ3NDEKTSV4RRFFQ69G5FAB"
	testProject   = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testTeamID    = "TM01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testQuizID    = "QZ01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testChalleng  = "CL01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testSessionID = "QS01ARZ3NDEKTSV4RRFFQ69G5FAV"
)

func newLookupTestCache(t *testing.T) *CacheWithRegistry {
	t.Helper()
	c, err := NewCacheWithRegistry(DefaultConfig())
	require.NoError(t, err)
	return c
}

// setUserLookupKeys populates the per-user lookup caches for a user
func setUserLookupKeys(t *testing.T, c *CacheWithRegistry, userID string) {
	t.Helper()
	require.True(t, c.Set(UserTeamInProjectKey(userID, testProject), testTeamID))
	require.True(t, c.Set(UserEnrolledChallengesKey(userID, testProject), map[string]bool{testChalleng: true}))
	require.True(t, c.Set(UserQuizSessionAccessKey(userID, testProject), map[string]bool{testQuizID: true}))
	require.True(t, c.Set(UserActiveQuizSessionKey(userID, testQuizID), testSessionID))
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
	_, ok = c.Get(UserActiveQuizSessionKey(testUserA, testQuizID))
	assert.False(t, ok, "active session lookup should be invalidated for user A")
	_, ok = c.Get(UserActiveQuizSessionKey(testUserB, testQuizID))
	assert.False(t, ok, "active session lookup should be invalidated for user B")

	// Unrelated per-user caches must survive
	_, ok = c.Get(UserTeamInProjectKey(testUserA, testProject))
	assert.True(t, ok, "team lookup should remain")
	_, ok = c.Get(UserEnrolledChallengesKey(testUserA, testProject))
	assert.True(t, ok, "enrolled challenges should remain")
}

func TestInvalidateQuizRemovesQuizSessionAccessCaches(t *testing.T) {
	c := newLookupTestCache(t)
	setUserLookupKeys(t, c, testUserA)
	require.True(t, c.Set(QuizAchievementsByQuizKey(testQuizID), "criteria"))
	c.Wait()

	c.InvalidateQuiz(testQuizID)

	_, ok := c.Get(UserQuizSessionAccessKey(testUserA, testProject))
	assert.False(t, ok, "quiz session access should be invalidated on quiz change")
	_, ok = c.Get(UserActiveQuizSessionKey(testUserA, testQuizID))
	assert.False(t, ok, "active session lookup should be invalidated on quiz change")
	_, ok = c.Get(QuizAchievementsByQuizKey(testQuizID))
	assert.False(t, ok, "quiz achievement criteria should be invalidated on quiz change")
}

func TestInvalidateQuizSessionRemovesSessionRow(t *testing.T) {
	c := newLookupTestCache(t)
	require.True(t, c.Set(QuizSessionKey(testSessionID), "session-row"))
	require.True(t, c.Set(QuizSessionKey("QS01ARZ3NDEKTSV4RRFFQ69G5FAB"), "other-row"))
	c.Wait()

	c.InvalidateQuizSession(testSessionID)

	_, ok := c.Get(QuizSessionKey(testSessionID))
	assert.False(t, ok, "session row should be invalidated")
	_, ok = c.Get(QuizSessionKey("QS01ARZ3NDEKTSV4RRFFQ69G5FAB"))
	assert.True(t, ok, "other session rows should remain")
}

func TestExtractUserTagForActiveQuizSessionKey(t *testing.T) {
	userID, ok := ExtractUserTag(UserActiveQuizSessionKey(testUserA, testQuizID))
	assert.True(t, ok)
	assert.Equal(t, testUserA, userID)
}
