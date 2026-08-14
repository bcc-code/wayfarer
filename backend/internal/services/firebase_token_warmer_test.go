package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testProjectID = "PR01ARZ3NDEKTSV4RRFFQ69G5FA"
	testChurchID  = "CH01ARZ3NDEKTSV4RRFFQ69G5FA"
)

func userRow(id, churchID string) *sqlc.GetUsersByIDsRow {
	return &sqlc.GetUsersByIDsRow{ID: id, ChurchID: churchID}
}

// newWarmerForTest builds a warmer with the paced delay removed, so tests do not
// sleep through the staggering that production relies on.
func newWarmerForTest(
	minter FirebaseTokenMinter,
	queries TokenWarmerQuerier,
	projects ProjectIDProvider,
) (*FirebaseTokenWarmer, *cacheHarness) {
	c := newTestCache()
	w := NewFirebaseTokenWarmer(minter, queries, c, projects)
	w.interval = 0 // the paced delay derives from interval; 0 => no sleeping
	return w, &cacheHarness{c}
}

type cacheHarness struct {
	c interface {
		Get(key string) (interface{}, bool)
	}
}

func (h *cacheHarness) token(userID string) (string, bool) {
	v, ok := h.c.Get(FirebaseTokenCacheKey(userID))
	if !ok {
		return "", false
	}
	resp, ok := v.(*model.FirebaseTokenResponse)
	if !ok {
		return "", false
	}
	return resp.Token, true
}

func TestFirebaseTokenCacheKey_MatchesResolverFormat(t *testing.T) {
	// The resolver reads this exact key; drift here silently disables warming.
	assert.Equal(t, "firebase_token:US123", FirebaseTokenCacheKey("US123"))
}

func TestNewFirebaseTokenWarmer_NilWhenDependencyMissing(t *testing.T) {
	c := newTestCache()
	minter := mocks.NewMockFirebaseTokenMinter(t)
	queries := mocks.NewMockTokenWarmerQuerier(t)
	projects := mocks.NewMockProjectIDProvider(t)

	assert.Nil(t, NewFirebaseTokenWarmer(nil, queries, c, projects))
	assert.Nil(t, NewFirebaseTokenWarmer(minter, nil, c, projects))
	assert.Nil(t, NewFirebaseTokenWarmer(minter, queries, nil, projects))
	assert.Nil(t, NewFirebaseTokenWarmer(minter, queries, c, nil))
	assert.NotNil(t, NewFirebaseTokenWarmer(minter, queries, c, projects))
}

func TestWarmAll_CachesATokenPerUser(t *testing.T) {
	ctx := context.Background()
	userIDs := []string{"US001", "US002", "US003"}

	projects := mocks.NewMockProjectIDProvider(t)
	projects.On("GetCurrentProjectID", ctx).Return(testProjectID, nil)

	queries := mocks.NewMockTokenWarmerQuerier(t)
	queries.On("GetUserIDsInProject", ctx, testProjectID).Return(userIDs, nil)
	queries.On("GetUsersByIDs", ctx, userIDs).Return([]*sqlc.GetUsersByIDsRow{
		userRow("US001", testChurchID),
		userRow("US002", testChurchID),
		userRow("US003", testChurchID),
	}, nil)

	minter := mocks.NewMockFirebaseTokenMinter(t)
	for _, id := range userIDs {
		minter.On("CreateCustomToken", ctx, id, testChurchID).Return("tok-"+id, nil)
	}

	w, h := newWarmerForTest(minter, queries, projects)

	n, err := w.warmAll(ctx, true)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	w.cache.Wait() // ristretto applies writes asynchronously
	for _, id := range userIDs {
		got, ok := h.token(id)
		assert.True(t, ok, "expected a cached token for %s", id)
		assert.Equal(t, "tok-"+id, got)
	}

	minted, warmed := w.Stats()
	assert.Equal(t, 3, minted)
	assert.Equal(t, 3, warmed)
}

func TestWarmAll_PacedPathAlsoCaches(t *testing.T) {
	ctx := context.Background()
	userIDs := []string{"US001", "US002"}

	projects := mocks.NewMockProjectIDProvider(t)
	projects.On("GetCurrentProjectID", ctx).Return(testProjectID, nil)

	queries := mocks.NewMockTokenWarmerQuerier(t)
	queries.On("GetUserIDsInProject", ctx, testProjectID).Return(userIDs, nil)
	queries.On("GetUsersByIDs", ctx, userIDs).Return([]*sqlc.GetUsersByIDsRow{
		userRow("US001", testChurchID),
		userRow("US002", testChurchID),
	}, nil)

	minter := mocks.NewMockFirebaseTokenMinter(t)
	minter.On("CreateCustomToken", ctx, mock.Anything, testChurchID).
		Return("tok", nil).Times(2)

	w, h := newWarmerForTest(minter, queries, projects)

	n, err := w.warmAll(ctx, false) // paced, not boot
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	w.cache.Wait()
	_, ok := h.token("US001")
	assert.True(t, ok)
}

func TestWarmAll_SkipsWhenNoCurrentProject(t *testing.T) {
	ctx := context.Background()

	projects := mocks.NewMockProjectIDProvider(t)
	projects.On("GetCurrentProjectID", ctx).Return("", nil)

	queries := mocks.NewMockTokenWarmerQuerier(t)
	minter := mocks.NewMockFirebaseTokenMinter(t)

	w, _ := newWarmerForTest(minter, queries, projects)

	n, err := w.warmAll(ctx, true)
	require.NoError(t, err)
	assert.Zero(t, n)
	// No user lookup and no signing should happen without a project.
	queries.AssertNotCalled(t, "GetUserIDsInProject", mock.Anything, mock.Anything)
	minter.AssertNotCalled(t, "CreateCustomToken", mock.Anything, mock.Anything, mock.Anything)
}

func TestWarmAll_PropagatesProjectAndUserErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("project lookup fails", func(t *testing.T) {
		projects := mocks.NewMockProjectIDProvider(t)
		projects.On("GetCurrentProjectID", ctx).Return("", errors.New("boom"))
		w, _ := newWarmerForTest(mocks.NewMockFirebaseTokenMinter(t), mocks.NewMockTokenWarmerQuerier(t), projects)

		_, err := w.warmAll(ctx, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "current project")
	})

	t.Run("user list fails", func(t *testing.T) {
		projects := mocks.NewMockProjectIDProvider(t)
		projects.On("GetCurrentProjectID", ctx).Return(testProjectID, nil)
		queries := mocks.NewMockTokenWarmerQuerier(t)
		queries.On("GetUserIDsInProject", ctx, testProjectID).Return(nil, errors.New("boom"))
		w, _ := newWarmerForTest(mocks.NewMockFirebaseTokenMinter(t), queries, projects)

		_, err := w.warmAll(ctx, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list project users")
	})
}

func TestWarmAll_MintFailureDoesNotAbortOtherUsers(t *testing.T) {
	ctx := context.Background()
	userIDs := []string{"US001", "US002"}

	projects := mocks.NewMockProjectIDProvider(t)
	projects.On("GetCurrentProjectID", ctx).Return(testProjectID, nil)

	queries := mocks.NewMockTokenWarmerQuerier(t)
	queries.On("GetUserIDsInProject", ctx, testProjectID).Return(userIDs, nil)
	queries.On("GetUsersByIDs", ctx, userIDs).Return([]*sqlc.GetUsersByIDsRow{
		userRow("US001", testChurchID),
		userRow("US002", testChurchID),
	}, nil)

	minter := mocks.NewMockFirebaseTokenMinter(t)
	minter.On("CreateCustomToken", ctx, "US001", testChurchID).Return("", errors.New("signing failed"))
	minter.On("CreateCustomToken", ctx, "US002", testChurchID).Return("tok-US002", nil)

	w, h := newWarmerForTest(minter, queries, projects)

	n, err := w.warmAll(ctx, true)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "the successful user should still be warmed")

	w.cache.Wait()
	_, ok := h.token("US001")
	assert.False(t, ok, "failed mint must not cache anything")
	got, ok := h.token("US002")
	assert.True(t, ok)
	assert.Equal(t, "tok-US002", got)
}

func TestWarmAll_BatchesUserLookups(t *testing.T) {
	ctx := context.Background()

	// 2,500 users must be fetched in 1,000-row batches, not one giant IN list.
	userIDs := make([]string, 2500)
	rows := make([]*sqlc.GetUsersByIDsRow, 0, 2500)
	for i := range userIDs {
		id := "US" + string(rune('A'+i%26)) + itoa(i)
		userIDs[i] = id
		rows = append(rows, userRow(id, testChurchID))
	}

	projects := mocks.NewMockProjectIDProvider(t)
	projects.On("GetCurrentProjectID", ctx).Return(testProjectID, nil)

	queries := mocks.NewMockTokenWarmerQuerier(t)
	queries.On("GetUserIDsInProject", ctx, testProjectID).Return(userIDs, nil)
	calls := 0
	queries.On("GetUsersByIDs", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			batch := args.Get(1).([]string)
			assert.LessOrEqual(t, len(batch), 1000, "batch larger than the 1000-row limit")
			calls++
		}).
		Return(func(_ context.Context, ids []string) []*sqlc.GetUsersByIDsRow {
			out := make([]*sqlc.GetUsersByIDsRow, 0, len(ids))
			for _, id := range ids {
				out = append(out, userRow(id, testChurchID))
			}
			return out
		}, func(_ context.Context, _ []string) error { return nil })

	minter := mocks.NewMockFirebaseTokenMinter(t)
	minter.On("CreateCustomToken", ctx, mock.Anything, testChurchID).Return("tok", nil)

	w, _ := newWarmerForTest(minter, queries, projects)

	n, err := w.warmAll(ctx, true)
	require.NoError(t, err)
	assert.Equal(t, 2500, n)
	assert.Equal(t, 3, calls, "2500 users should take ceil(2500/1000) = 3 batches")
}

func TestWarmer_StartAndStopAreSafe(t *testing.T) {
	ctx := context.Background()

	projects := mocks.NewMockProjectIDProvider(t)
	projects.On("GetCurrentProjectID", mock.Anything).Return(testProjectID, nil).Maybe()
	queries := mocks.NewMockTokenWarmerQuerier(t)
	queries.On("GetUserIDsInProject", mock.Anything, testProjectID).Return([]string{"US001"}, nil).Maybe()
	queries.On("GetUsersByIDs", mock.Anything, mock.Anything).
		Return([]*sqlc.GetUsersByIDsRow{userRow("US001", testChurchID)}, nil).Maybe()
	minter := mocks.NewMockFirebaseTokenMinter(t)
	minter.On("CreateCustomToken", mock.Anything, "US001", testChurchID).Return("tok", nil).Maybe()

	c := newTestCache()
	w := NewFirebaseTokenWarmer(minter, queries, c, projects)
	w.interval = time.Hour // no second pass during the test

	w.Start(ctx)
	w.Stop() // must not hang or panic

	// A nil warmer is a no-op, so callers can wire it unconditionally.
	var nilWarmer *FirebaseTokenWarmer
	nilWarmer.Start(ctx)
	nilWarmer.Stop()
	minted, warmed := nilWarmer.Stats()
	assert.Zero(t, minted)
	assert.Zero(t, warmed)
}

// itoa avoids pulling strconv into the table-building helper above.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
