package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuizSessionReset asserts that resetQuizSessionSubmission only throws
// away an in-progress attempt: once a submission is finalized its points are
// in the score journal, so resetting it would let the user score again.
func TestQuizSessionReset(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	adminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)

	userID := data.UserIDs[0]
	userToken, err := testutil.GenerateUserToken(userID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	// ---------------------------------------------------------------- helpers

	// exec runs a query and returns its error message ("" on success)
	exec := func(t *testing.T, token, query string, vars map[string]any, out any) string {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, query, vars)
		if resp.HasErrors() {
			return resp.ErrorMessage()
		}
		if out != nil {
			require.NoError(t, resp.UnmarshalData(out))
		}
		return ""
	}
	admin := func(t *testing.T, query string, vars map[string]any, out any) {
		t.Helper()
		require.Empty(t, exec(t, adminToken, query, vars, out))
	}

	// newQuiz creates a quiz worth 100 points (50 for completing, 50 for the
	// one question) that does not allow retakes, and returns the quiz ID and
	// the question's correct answer
	newQuiz := func(t *testing.T, name string) (quizID, questionID, rightAnswerID string) {
		t.Helper()
		var challenge struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		admin(t, `mutation($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
			createChallenge(projectId: $projectId, eventId: $eventId, input: $input) { id }
		}`, map[string]any{
			"projectId": projectID,
			"eventId":   eventID,
			"input":     map[string]any{"type": "QUIZ", "name": name, "buttonText": "Go", "description": "<p>x</p>"},
		}, &challenge)

		var quiz struct {
			CreateQuiz struct{ ID string } `json:"createQuiz"`
		}
		admin(t, `mutation($input: CreateQuizInput!) { createQuiz(input: $input) { id } }`,
			map[string]any{"input": map[string]any{
				"projectId": projectID, "challengeId": challenge.CreateChallenge.ID, "name": name, "description": "x",
				"randomizeQuestions": false, "revealCorrectAnswers": true, "allowRetakes": false, "completionPoints": 50,
			}}, &quiz)

		var question struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID        string `json:"id"`
					IsCorrect bool   `json:"isCorrect"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		admin(t, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
			addQuizQuestion(quizId: $quizId, input: $input) {
				... on PredefinedQuestion { id predefinedAnswers { id isCorrect } }
			}
		}`, map[string]any{"quizId": quiz.CreateQuiz.ID, "input": map[string]any{
			"questionType": "PREDEFINED", "questionText": "Question?", "questionOrder": 0, "points": 50,
			"predefinedAnswers": []map[string]any{
				{"answerText": "right", "isCorrect": true, "answerOrder": 0},
				{"answerText": "wrong", "isCorrect": false, "answerOrder": 1},
			},
		}}, &question)
		for _, a := range question.AddQuizQuestion.PredefinedAnswers {
			if a.IsCorrect {
				rightAnswerID = a.ID
			}
		}
		require.NotEmpty(t, rightAnswerID)
		return quiz.CreateQuiz.ID, question.AddQuizQuestion.ID, rightAnswerID
	}

	// openSession creates a session for quizID, grants the user access and opens it
	openSession := func(t *testing.T, quizID string) string {
		t.Helper()
		var session struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		admin(t, `mutation($input: CreateQuizSessionInput!) { createQuizSession(input: $input) { id } }`,
			map[string]any{"input": map[string]any{"quizId": quizID}}, &session)
		id := session.CreateQuizSession.ID
		admin(t, `mutation($input: GrantQuizSessionAccessInput!) { grantQuizSessionAccess(input: $input) }`,
			map[string]any{"input": map[string]any{"sessionId": id, "userIds": []string{userID}}}, nil)
		admin(t, `mutation($id: ID!) { openQuizSession(id: $id) { id } }`, map[string]any{"id": id}, nil)
		return id
	}

	start := func(t *testing.T, sessionID string) string {
		t.Helper()
		var res struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.Empty(t, exec(t, userToken, `mutation($sessionId: ID!) { startQuizSession(sessionId: $sessionId) { id } }`,
			map[string]any{"sessionId": sessionID}, &res))
		return res.StartQuizSession.ID
	}
	answerRight := func(t *testing.T, submissionID, questionID, rightAnswerID string) {
		t.Helper()
		require.Empty(t, exec(t, userToken, `mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
			submitQuizAnswer(submissionId: $submissionId, input: $input) { id }
		}`, map[string]any{"submissionId": submissionID, "input": map[string]any{
			"questionId": questionID, "selectedAnswerIds": []string{rightAnswerID},
		}}, nil))
	}

	const finalizeMutation = `mutation($id: ID!) { finalizeQuiz(submissionId: $id) { id } }`
	const resetMutation = `mutation($id: ID!) { resetQuizSessionSubmission(sessionId: $id) }`

	// quizJournal returns the number and sum of the user's journal entries for the quiz
	quizJournal := func(t *testing.T, quizID string) (count int, sum int) {
		t.Helper()
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT count(*), COALESCE(sum(points), 0) FROM score_journal
			 WHERE source_type = 'QUIZ' AND source_id = $1 AND user_id = $2`, quizID, userID,
		).Scan(&count, &sum))
		return count, sum
	}
	// submission returns whether the user has a submission in the session and
	// whether it is completed
	submission := func(t *testing.T, sessionID string) (exists, completed bool) {
		t.Helper()
		err := dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT completed_at IS NOT NULL FROM quiz_submissions WHERE session_id = $1 AND user_id = $2`,
			sessionID, userID,
		).Scan(&completed)
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false
		}
		require.NoError(t, err)
		return true, completed
	}

	// ------------------------------------------------------------------ tests

	t.Run("a completed submission can't be reset to score again", func(t *testing.T) {
		quizID, questionID, right := newQuiz(t, "Reset after finalize")
		sessionID := openSession(t, quizID)

		sub := start(t, sessionID)
		answerRight(t, sub, questionID, right)
		require.Empty(t, exec(t, userToken, finalizeMutation, map[string]any{"id": sub}, nil))
		count, sum := quizJournal(t, quizID)
		require.Equal(t, 1, count)
		require.Equal(t, 100, sum)

		msg := exec(t, userToken, resetMutation, map[string]any{"id": sessionID}, nil)
		assert.Contains(t, msg, "cannot reset a completed submission")

		exists, completed := submission(t, sessionID)
		assert.True(t, exists, "the completed submission is kept")
		assert.True(t, completed)

		// Starting again hands back the completed submission, which can't be
		// finalized a second time
		assert.Equal(t, sub, start(t, sessionID))
		msg = exec(t, userToken, finalizeMutation, map[string]any{"id": sub}, nil)
		assert.Contains(t, msg, "submission already completed")

		count, sum = quizJournal(t, quizID)
		assert.Equal(t, 1, count, "the quiz is journaled once")
		assert.Equal(t, 100, sum, "the user earns the quiz's 100 points once")
	})

	t.Run("resetting with no submission is a no-op", func(t *testing.T) {
		quizID, _, _ := newQuiz(t, "Reset without submission")
		sessionID := openSession(t, quizID)

		var res struct {
			ResetQuizSessionSubmission bool `json:"resetQuizSessionSubmission"`
		}
		require.Empty(t, exec(t, userToken, resetMutation, map[string]any{"id": sessionID}, &res))
		assert.True(t, res.ResetQuizSessionSubmission)
	})

	// raceOnRowLock runs finalize and reset concurrently in a fixed order: a
	// test transaction holds the submission's row lock, the first mutation
	// queues on it, then the second, then the lock is released. Postgres hands
	// the row to the waiters in the order they queued. Returns each mutation's
	// error message ("" on success).
	raceOnRowLock := func(t *testing.T, submissionID, sessionID string, finalizeFirst bool) (finalizeMsg, resetMsg string) {
		t.Helper()
		tx, err := dbMgr.DB.Pool.Begin(ctx)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback(ctx) }()
		_, err = tx.Exec(ctx, `SELECT 1 FROM quiz_submissions WHERE id = $1 FOR UPDATE`, submissionID)
		require.NoError(t, err)

		run := func(query string, vars map[string]any, msg *string) {
			resp, err := client.WithAuth(userToken).Execute(ctx, query, vars)
			switch {
			case err != nil:
				*msg = err.Error()
			case resp.HasErrors():
				*msg = resp.ErrorMessage()
			}
		}
		finalize := func() { run(finalizeMutation, map[string]any{"id": submissionID}, &finalizeMsg) }
		reset := func() { run(resetMutation, map[string]any{"id": sessionID}, &resetMsg) }
		first, second := reset, finalize
		if finalizeFirst {
			first, second = finalize, reset
		}

		// waitForLockWaiters waits until n backends are blocked on a lock
		waitForLockWaiters := func(n int) {
			require.Eventually(t, func() bool {
				var waiting int
				err := dbMgr.DB.Pool.QueryRow(ctx,
					`SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'`,
				).Scan(&waiting)
				return err == nil && waiting == n
			}, 10*time.Second, 10*time.Millisecond)
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); first() }()
		waitForLockWaiters(1)
		go func() { defer wg.Done(); second() }()
		waitForLockWaiters(2)
		require.NoError(t, tx.Commit(ctx))
		wg.Wait()
		return finalizeMsg, resetMsg
	}

	t.Run("reset waiting on a running finalize does not delete the finalized submission", func(t *testing.T) {
		quizID, questionID, right := newQuiz(t, "Finalize then reset race")
		sessionID := openSession(t, quizID)
		sub := start(t, sessionID)
		answerRight(t, sub, questionID, right)

		finalizeMsg, resetMsg := raceOnRowLock(t, sub, sessionID, true)

		assert.Empty(t, finalizeMsg)
		assert.Contains(t, resetMsg, "cannot reset a completed submission")
		exists, completed := submission(t, sessionID)
		assert.True(t, exists, "the finalized submission is kept")
		assert.True(t, completed)
		count, sum := quizJournal(t, quizID)
		assert.Equal(t, 1, count)
		assert.Equal(t, 100, sum)
	})

	t.Run("finalize waiting on a running reset awards no points", func(t *testing.T) {
		quizID, questionID, right := newQuiz(t, "Reset then finalize race")
		sessionID := openSession(t, quizID)
		sub := start(t, sessionID)
		answerRight(t, sub, questionID, right)

		finalizeMsg, resetMsg := raceOnRowLock(t, sub, sessionID, false)

		assert.Empty(t, resetMsg)
		assert.NotEmpty(t, finalizeMsg, "finalizing a deleted submission fails")
		exists, _ := submission(t, sessionID)
		assert.False(t, exists, "the in-progress submission is reset")
		count, _ := quizJournal(t, quizID)
		assert.Zero(t, count, "no points for a deleted submission")
	})
}
