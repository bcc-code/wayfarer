package loaders

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserQuizKey(t *testing.T) {
	key := UserQuizKey{
		UserID: "US01K8XV6VK9ED2GBZSQ2VDTAT8T",
		QuizID: "QZ01K8XV6VK9ED2GBZSQ2VDTAT8T",
	}

	assert.Equal(t, "US01K8XV6VK9ED2GBZSQ2VDTAT8T:QZ01K8XV6VK9ED2GBZSQ2VDTAT8T", key.String())
	assert.Equal(t, key, key.Raw())
}

func TestUserActiveQuizSessionCacheHit(t *testing.T) {
	c := newTestCache(t)

	key := UserQuizKey{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", QuizID: "QZ01AAAAAAAAAAAAAAAAAAAAAAAA"}
	session := &sqlc.QuizSession{ID: "QS01AAAAAAAAAAAAAAAAAAAAAAAA", QuizID: key.QuizID, State: "OPEN"}
	c.SetWithTTL(cache.UserActiveQuizSessionKey(key.UserID, key.QuizID), session, QuizSessionAccessTTL)
	c.Wait()

	// db is nil: a cache miss would panic inside the batch function, so a
	// clean result proves the cache-hit path never touches the database
	batch := userActiveQuizSessionBatchFunc(nil, c)
	results := batch(context.Background(), []UserQuizKey{key})

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	require.NotNil(t, results[0].Data)
	assert.Equal(t, "QS01AAAAAAAAAAAAAAAAAAAAAAAA", results[0].Data.ID)
}

func TestUserActiveQuizSessionNegativeCacheHit(t *testing.T) {
	c := newTestCache(t)

	key := UserQuizKey{UserID: "US01AAAAAAAAAAAAAAAAAAAAAAAA", QuizID: "QZ01AAAAAAAAAAAAAAAAAAAAAAAA"}
	// "No active session" is cached as a typed nil pointer
	c.SetWithTTL(cache.UserActiveQuizSessionKey(key.UserID, key.QuizID), (*sqlc.QuizSession)(nil), QuizSessionAccessTTL)
	c.Wait()

	batch := userActiveQuizSessionBatchFunc(nil, c)
	results := batch(context.Background(), []UserQuizKey{key})

	require.Len(t, results, 1)
	require.NoError(t, results[0].Error)
	assert.Nil(t, results[0].Data)
}
