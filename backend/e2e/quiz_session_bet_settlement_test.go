package e2e

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuizSessionBetSettlement covers the core payout of PREDEFINED bets when
// an admin finishes a quiz session.
func TestQuizSessionBetSettlement(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	data, err := dbMgr.Seed(ctx, 42, testutil.DefaultSeedConfig())
	require.NoError(t, err)

	winnerID := data.UserIDs[0]
	loserID := data.UserIDs[1]
	adminUserID := data.UserIDs[2]
	retryID := data.UserIDs[3]
	raceOpenID := data.UserIDs[4]
	raceCompletedID := data.UserIDs[5]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	router, cleanup, err := testutil.SetupTestServer(ctx, dbMgr)
	require.NoError(t, err)
	defer cleanup()

	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	winnerToken, err := testutil.GenerateUserToken(winnerID)
	require.NoError(t, err)
	loserToken, err := testutil.GenerateUserToken(loserID)
	require.NoError(t, err)
	retryToken, err := testutil.GenerateUserToken(retryID)
	require.NoError(t, err)
	raceOpenToken, err := testutil.GenerateUserToken(raceOpenID)
	require.NoError(t, err)
	raceCompletedToken, err := testutil.GenerateUserToken(raceCompletedID)
	require.NoError(t, err)

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	admin := func(t *testing.T, query string, vars map[string]any, out any) {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, query, vars)
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		if out != nil {
			require.NoError(t, resp.UnmarshalData(out))
		}
	}

	// --- quiz with a challenge ---
	var challenge struct {
		CreateChallenge struct{ ID string } `json:"createChallenge"`
	}
	admin(t, `mutation($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
		createChallenge(projectId: $projectId, eventId: $eventId, input: $input) { id }
	}`, map[string]any{
		"projectId": projectID,
		"eventId":   eventID,
		"input": map[string]any{
			"type": "QUIZ", "name": "Settlement Challenge", "buttonText": "Go", "description": "<p>x</p>",
		},
	}, &challenge)
	challengeID := challenge.CreateChallenge.ID
	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	admin(t, `mutation($id: ID!, $at: DateTime!) { publishChallenge(id: $id, publishedAt: $at) { id } }`,
		map[string]any{"id": challengeID, "at": past}, nil)
	admin(t, `mutation($id: ID!, $at: DateTime!) { setChallengeVisibility(id: $id, visibleAt: $at) { id } }`,
		map[string]any{"id": challengeID, "at": past}, nil)

	var quiz struct {
		CreateQuiz struct{ ID string } `json:"createQuiz"`
	}
	admin(t, `mutation($input: CreateQuizInput!) { createQuiz(input: $input) { id } }`, map[string]any{
		"input": map[string]any{
			"projectId": projectID, "challengeId": challengeID, "name": "Settlement Quiz", "description": "x",
			"randomizeQuestions": false, "revealCorrectAnswers": true, "allowRetakes": false, "completionPoints": 0,
		},
	}, &quiz)
	quizID := quiz.CreateQuiz.ID

	type predefinedQuestion struct {
		ID                       string   `json:"id"`
		BettingMultiplierCorrect *float64 `json:"bettingMultiplierCorrect"`
		BettingMultiplierWrong   *float64 `json:"bettingMultiplierWrong"`
		PredefinedAnswers        []struct {
			ID         string `json:"id"`
			AnswerText string `json:"answerText"`
		} `json:"predefinedAnswers"`
	}
	addPredefined := func(t *testing.T, order int, multipliers map[string]any) (q predefinedQuestion, correctID, wrongID string) {
		input := map[string]any{
			"questionType": "PREDEFINED", "questionText": "Question?", "questionOrder": order,
			"bettingEnabled": true,
			"predefinedAnswers": []map[string]any{
				{"answerText": "right", "isCorrect": true, "answerOrder": 0},
				{"answerText": "wrong", "isCorrect": false, "answerOrder": 1},
			},
		}
		for k, v := range multipliers {
			input[k] = v
		}
		var res struct {
			AddQuizQuestion predefinedQuestion `json:"addQuizQuestion"`
		}
		admin(t, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
			addQuizQuestion(quizId: $quizId, input: $input) {
				... on PredefinedQuestion {
					id bettingMultiplierCorrect bettingMultiplierWrong
					predefinedAnswers { id answerText }
				}
			}
		}`, map[string]any{"quizId": quizID, "input": input}, &res)
		q = res.AddQuizQuestion
		for _, a := range q.PredefinedAnswers {
			if a.AnswerText == "right" {
				correctID = a.ID
			} else {
				wrongID = a.ID
			}
		}
		require.NotEmpty(t, correctID)
		require.NotEmpty(t, wrongID)
		return q, correctID, wrongID
	}

	// Custom payout: correct 3x, wrong half back
	customQ, customRight, customWrong := addPredefined(t, 0, map[string]any{
		"bettingMultiplierCorrect": 3.0,
		"bettingMultiplierWrong":   0.5,
	})
	// Default payout: correct 2x, wrong 0
	defaultQ, defaultRight, defaultWrong := addPredefined(t, 1, nil)
	// Answered without a selection: never graded, so never settled
	ungradedQ, _, _ := addPredefined(t, 3, nil)

	// ORDERING bets are left to the plugin and must not be settled by the core
	var ordering struct {
		AddQuizQuestion struct {
			ID            string `json:"id"`
			OrderingItems []struct {
				ID string `json:"id"`
			} `json:"orderingItems"`
		} `json:"addQuizQuestion"`
	}
	admin(t, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
		addQuizQuestion(quizId: $quizId, input: $input) {
			... on OrderingQuestion { id orderingItems { id } }
		}
	}`, map[string]any{"quizId": quizID, "input": map[string]any{
		"questionType": "ORDERING", "questionText": "Order", "questionOrder": 2, "bettingEnabled": true,
		"orderingItems": []map[string]any{{"itemText": "a", "correctOrder": 0}, {"itemText": "b", "correctOrder": 1}},
	}}, &ordering)
	orderingQID := ordering.AddQuizQuestion.ID
	orderingIDs := []string{ordering.AddQuizQuestion.OrderingItems[0].ID, ordering.AddQuizQuestion.OrderingItems[1].ID}

	t.Run("multipliers are stored and returned", func(t *testing.T) {
		require.NotNil(t, customQ.BettingMultiplierCorrect)
		require.NotNil(t, customQ.BettingMultiplierWrong)
		assert.Equal(t, 3.0, *customQ.BettingMultiplierCorrect)
		assert.Equal(t, 0.5, *customQ.BettingMultiplierWrong)
		assert.Nil(t, defaultQ.BettingMultiplierCorrect)
		assert.Nil(t, defaultQ.BettingMultiplierWrong)
	})

	t.Run("wrong multiplier above correct is rejected", func(t *testing.T) {
		resp := client.WithAuth(adminToken).MustExecute(t, `mutation($id: ID!, $input: UpdateQuizQuestionInput!) {
			updateQuizQuestion(id: $id, input: $input) { id }
		}`, map[string]any{"id": customQ.ID, "input": map[string]any{
			"bettingMultiplierCorrect": 1.0, "bettingMultiplierWrong": 2.0,
		}})
		assert.True(t, resp.HasErrors())
	})

	updateMultipliers := func(t *testing.T, questionID string, input map[string]any) bool {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, `mutation($id: ID!, $input: UpdateQuizQuestionInput!) {
			updateQuizQuestion(id: $id, input: $input) { id }
		}`, map[string]any{"id": questionID, "input": input})
		return !resp.HasErrors()
	}

	t.Run("only wrong above the default correct is rejected", func(t *testing.T) {
		// defaultQ has no correct multiplier, so 2.0 applies
		assert.False(t, updateMultipliers(t, defaultQ.ID, map[string]any{"bettingMultiplierWrong": 3.0}))
	})

	t.Run("multiplier above 100 or with 3 decimals is rejected", func(t *testing.T) {
		assert.False(t, updateMultipliers(t, defaultQ.ID, map[string]any{"bettingMultiplierCorrect": 100.01}))
		assert.False(t, updateMultipliers(t, defaultQ.ID, map[string]any{"bettingMultiplierCorrect": 1.234}))
	})

	t.Run("only wrong is checked against the stored correct", func(t *testing.T) {
		// customQ stores correct 3.0, so 2.5 is fine even though it exceeds the default
		assert.True(t, updateMultipliers(t, customQ.ID, map[string]any{"bettingMultiplierWrong": 2.5}))
		// restore for the payout checks below
		require.True(t, updateMultipliers(t, customQ.ID, map[string]any{"bettingMultiplierWrong": 0.5}))
	})

	// --- give users points to bet with ---
	for _, uid := range []string{winnerID, loserID, retryID, raceOpenID, raceCompletedID} {
		admin(t, `mutation($input: CreateScoreAdjustmentInput!) { createScoreAdjustment(input: $input) { id } }`,
			map[string]any{"input": map[string]any{"projectId": projectID, "userId": uid, "points": 1000, "reason": "bets"}}, nil)
	}

	// --- session ---
	createOpenSession := func(t *testing.T, userIDs []string) string {
		t.Helper()
		var session struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		admin(t, `mutation($input: CreateQuizSessionInput!) { createQuizSession(input: $input) { id } }`,
			map[string]any{"input": map[string]any{"quizId": quizID}}, &session)
		id := session.CreateQuizSession.ID
		admin(t, `mutation($input: GrantQuizSessionAccessInput!) { grantQuizSessionAccess(input: $input) }`,
			map[string]any{"input": map[string]any{"sessionId": id, "userIds": userIDs}}, nil)
		admin(t, `mutation($id: ID!) { openQuizSession(id: $id) { id } }`, map[string]any{"id": id}, nil)
		return id
	}
	sessionID := createOpenSession(t, []string{winnerID, loserID, retryID})

	answer := func(t *testing.T, token, submissionID string, input map[string]any) string {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
			submitQuizAnswer(submissionId: $submissionId, input: $input) { id }
		}`, map[string]any{"submissionId": submissionID, "input": input})
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		var res struct {
			SubmitQuizAnswer struct{ ID string } `json:"submitQuizAnswer"`
		}
		require.NoError(t, resp.UnmarshalData(&res))
		return res.SubmitQuizAnswer.ID
	}
	start := func(t *testing.T, token, sessionID string) string {
		t.Helper()
		resp := client.WithAuth(token).MustExecute(t, `mutation($sessionId: ID!) { startQuizSession(sessionId: $sessionId) { id } }`,
			map[string]any{"sessionId": sessionID})
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		var res struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		require.NoError(t, resp.UnmarshalData(&res))
		return res.StartQuizSession.ID
	}

	winnerSub := start(t, winnerToken, sessionID)
	loserSub := start(t, loserToken, sessionID)
	retrySub := start(t, retryToken, sessionID)

	winnerCustom := answer(t, winnerToken, winnerSub, map[string]any{"questionId": customQ.ID, "selectedAnswerIds": []string{customRight}, "betAmount": 100})
	winnerDefault := answer(t, winnerToken, winnerSub, map[string]any{"questionId": defaultQ.ID, "selectedAnswerIds": []string{defaultRight}, "betAmount": 40})
	winnerOrdering := answer(t, winnerToken, winnerSub, map[string]any{"questionId": orderingQID, "submittedOrder": orderingIDs, "betAmount": 10})
	loserCustom := answer(t, loserToken, loserSub, map[string]any{"questionId": customQ.ID, "selectedAnswerIds": []string{customWrong}, "betAmount": 100})
	loserDefault := answer(t, loserToken, loserSub, map[string]any{"questionId": defaultQ.ID, "selectedAnswerIds": []string{defaultWrong}, "betAmount": 40})
	winnerUngraded := answer(t, winnerToken, winnerSub, map[string]any{"questionId": ungradedQ.ID, "betAmount": 20})
	retryBet := answer(t, retryToken, retrySub, map[string]any{"questionId": customQ.ID, "selectedAnswerIds": []string{customRight}, "betAmount": 100})

	admin(t, `mutation($id: ID!) { lockQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)

	type settled struct {
		PointsEarned   *int32
		ScoreJournalID *string
	}
	getResponse := func(t *testing.T, responseID string) settled {
		t.Helper()
		var s settled
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT points_earned, score_journal_id FROM quiz_responses WHERE id = $1`, responseID,
		).Scan(&s.PointsEarned, &s.ScoreJournalID))
		return s
	}
	betJournal := func(t *testing.T, responseID string) (count int, sum int) {
		t.Helper()
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT count(*), COALESCE(sum(points), 0) FROM score_journal WHERE source_type = 'BET' AND source_id = $1`, responseID,
		).Scan(&count, &sum))
		return count, sum
	}

	t.Run("nothing is paid out before the session finishes", func(t *testing.T) {
		for _, id := range []string{winnerCustom, loserCustom} {
			count, _ := betJournal(t, id)
			assert.Zero(t, count)
			assert.Nil(t, getResponse(t, id).ScoreJournalID)
		}
	})

	// Two finishes at once (e.g. a double click) must still pay out each bet once
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.WithAuth(adminToken).MustExecute(t, `mutation($id: ID!) { finishQuizSession(id: $id) { id } }`,
				map[string]any{"id": sessionID})
		}()
	}
	wg.Wait()

	tests := []struct {
		name       string
		responseID string
		wantNet    int
	}{
		{name: "correct, custom 3x", responseID: winnerCustom, wantNet: 200},
		{name: "wrong, custom half back", responseID: loserCustom, wantNet: -50},
		{name: "correct, default 2x", responseID: winnerDefault, wantNet: 40},
		{name: "wrong, default bet lost", responseID: loserDefault, wantNet: -40},
	}
	for _, tt := range tests {
		t.Run("settled: "+tt.name, func(t *testing.T) {
			resp := getResponse(t, tt.responseID)
			require.NotNil(t, resp.PointsEarned)
			assert.Equal(t, int32(tt.wantNet), *resp.PointsEarned)
			require.NotNil(t, resp.ScoreJournalID)

			count, sum := betJournal(t, tt.responseID)
			assert.Equal(t, 2, count, "stake and winnings entries")
			assert.Equal(t, tt.wantNet, sum)
		})
	}

	t.Run("bet without an answer is not settled", func(t *testing.T) {
		count, _ := betJournal(t, winnerUngraded)
		assert.Zero(t, count)
		assert.Nil(t, getResponse(t, winnerUngraded).ScoreJournalID)
	})

	t.Run("ORDERING bet is not settled by the core", func(t *testing.T) {
		count, _ := betJournal(t, winnerOrdering)
		assert.Zero(t, count)
		assert.Nil(t, getResponse(t, winnerOrdering).ScoreJournalID)
	})

	t.Run("finishing again does not pay out twice", func(t *testing.T) {
		admin(t, `mutation($id: ID!) { finishQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
		for _, tt := range tests {
			count, sum := betJournal(t, tt.responseID)
			assert.Equal(t, 2, count)
			assert.Equal(t, tt.wantNet, sum)
		}
	})

	t.Run("finishing again retries a failed auto-submission and settlement", func(t *testing.T) {
		// Undo the first finish for one user, as if auto-submission had failed
		// (settlement is then skipped): submission open, bet unpaid
		_, err := dbMgr.DB.Pool.Exec(ctx,
			`UPDATE quiz_responses SET score_journal_id = NULL, points_earned = NULL WHERE id = $1`, retryBet)
		require.NoError(t, err)
		_, err = dbMgr.DB.Pool.Exec(ctx, `DELETE FROM score_journal WHERE source_type = 'BET' AND source_id = $1`, retryBet)
		require.NoError(t, err)
		_, err = dbMgr.DB.Pool.Exec(ctx,
			`UPDATE quiz_submissions SET completed_at = NULL, auto_submitted = false WHERE id = $1`, retrySub)
		require.NoError(t, err)

		admin(t, `mutation($id: ID!) { finishQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)

		var completed bool
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT completed_at IS NOT NULL FROM quiz_submissions WHERE id = $1`, retrySub,
		).Scan(&completed))
		assert.True(t, completed, "submission auto-submitted on retry")

		resp := getResponse(t, retryBet)
		require.NotNil(t, resp.PointsEarned)
		assert.Equal(t, int32(200), *resp.PointsEarned)
		count, sum := betJournal(t, retryBet)
		assert.Equal(t, 2, count)
		assert.Equal(t, 200, sum)
	})

	// An answer write that holds its submission lock while the session is
	// finished must be committed before settlement, also for a submission the
	// user already completed (auto-submit skips those)
	raceCases := []struct {
		name      string
		userID    string
		token     string
		completed bool
	}{
		{name: "open submission", userID: raceOpenID, token: raceOpenToken},
		{name: "completed submission", userID: raceCompletedID, token: raceCompletedToken, completed: true},
	}
	for _, tc := range raceCases {
		t.Run("an answer in flight while finishing is settled: "+tc.name, func(t *testing.T) {
			raceSessionID := createOpenSession(t, []string{tc.userID})
			raceSub := start(t, tc.token, raceSessionID)
			if tc.completed {
				_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE quiz_submissions SET completed_at = now() WHERE id = $1`, raceSub)
				require.NoError(t, err)
			}
			admin(t, `mutation($id: ID!) { lockQuizSession(id: $id) { id } }`, map[string]any{"id": raceSessionID}, nil)

			// Hold the submission lock and store a bet, as an in-flight answer write does
			tx, err := dbMgr.DB.Pool.Begin(ctx)
			require.NoError(t, err)
			defer func() { _ = tx.Rollback(ctx) }()
			_, err = tx.Exec(ctx, `SELECT id FROM quiz_submissions WHERE id = $1 FOR UPDATE`, raceSub)
			require.NoError(t, err)
			raceBet := ulid.NewQuizResponseID()
			_, err = tx.Exec(ctx, `INSERT INTO quiz_responses (id, submission_id, question_id, selected_answer_ids, is_correct, bet_amount)
				VALUES ($1, $2, $3, $4, true, 50)`, raceBet, raceSub, defaultQ.ID, `["`+defaultRight+`"]`)
			require.NoError(t, err)

			type finishResult struct {
				resp *testutil.GraphQLResponse
				err  error
			}
			finished := make(chan finishResult, 1)
			go func() {
				resp, err := client.WithAuth(adminToken).Execute(ctx, `mutation($id: ID!) { finishQuizSession(id: $id) { id } }`,
					map[string]any{"id": raceSessionID})
				finished <- finishResult{resp: resp, err: err}
			}()

			// Finishing must wait for the answer's transaction
			require.Eventually(t, func() bool {
				var waiting int
				_ = dbMgr.DB.Pool.QueryRow(ctx,
					`SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND query LIKE '%WaitForSessionSubmissionLocks%'`,
				).Scan(&waiting)
				return waiting > 0
			}, 10*time.Second, 10*time.Millisecond)
			require.NoError(t, tx.Commit(ctx))

			select {
			case res := <-finished:
				require.NoError(t, res.err)
				require.False(t, res.resp.HasErrors(), res.resp.ErrorMessage())
			case <-time.After(10 * time.Second):
				t.Fatal("finishQuizSession did not return")
			}

			resp := getResponse(t, raceBet)
			require.NotNil(t, resp.PointsEarned)
			assert.Equal(t, int32(50), *resp.PointsEarned, "default 2x on a correct answer")
			count, sum := betJournal(t, raceBet)
			assert.Equal(t, 2, count)
			assert.Equal(t, 50, sum)
		})
	}
}
