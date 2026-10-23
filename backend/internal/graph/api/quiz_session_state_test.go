package api

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

// mockQuizSessionStateUpdater is a mock implementation of quizSessionStateUpdater
type mockQuizSessionStateUpdater struct {
	mock.Mock
}

func (m *mockQuizSessionStateUpdater) UpdateQuizSessionState(ctx context.Context, arg sqlc.UpdateQuizSessionStateParams) (*sqlc.QuizSession, error) {
	args := m.Called(ctx, arg)
	session, _ := args.Get(0).(*sqlc.QuizSession)
	return session, args.Error(1)
}

func TestTransitionQuizSession(t *testing.T) {
	ctx := context.Background()
	const sessionID = "QS01ARZ3NDEKTSV4RRFFQ69G5FAV"
	params := sqlc.UpdateQuizSessionStateParams{ID: sessionID, Fromstate: "LOCKED", State: "FINISHED"}

	t.Run("returns the row when the session was in the from state", func(t *testing.T) {
		q := &mockQuizSessionStateUpdater{}
		row := &sqlc.QuizSession{ID: sessionID, State: "FINISHED"}
		q.On("UpdateQuizSessionState", ctx, params).Return(row, nil)

		got, ok, err := transitionQuizSession(ctx, q, sessionID, "LOCKED", "FINISHED")

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Same(t, row, got)
		q.AssertExpectations(t)
	})

	t.Run("reports no transition when another call moved the session first", func(t *testing.T) {
		q := &mockQuizSessionStateUpdater{}
		q.On("UpdateQuizSessionState", ctx, params).Return(nil, pgx.ErrNoRows)

		got, ok, err := transitionQuizSession(ctx, q, sessionID, "LOCKED", "FINISHED")

		require.NoError(t, err)
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("returns other errors", func(t *testing.T) {
		q := &mockQuizSessionStateUpdater{}
		dbErr := errors.New("connection reset")
		q.On("UpdateQuizSessionState", ctx, params).Return(nil, dbErr)

		got, ok, err := transitionQuizSession(ctx, q, sessionID, "LOCKED", "FINISHED")

		require.ErrorIs(t, err, dbErr)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}
