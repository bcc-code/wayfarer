package e2e

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuizBettingEdgeCases asserts how quiz betting SHOULD behave across
// sessions, questions and answer paths. Failing subtests point at bugs; see
// notes/betting-edge-cases.md for the status of each rule.
//
// Rules (subtests name the rule they check):
//
//	R1 Open stakes are reserved: a bet may not exceed the score minus the
//	   user's other unsettled bets in the project, across questions and
//	   sessions. Betting never makes a score negative.
//	R2 A user can always answer a betting question. betAmount must be sent;
//	   a bet of 0 is always accepted and means no bet (the app's slider starts
//	   at 0). Limits only apply to real bets.
//	R3 Every stored bet is resolved at finish. A bet without an answer is void:
//	   its stake is returned. Betting can't be turned off while a question has
//	   open bets.
//	R4 Answers and bets are only accepted while the session is OPEN.
//	R5 A bet pays out at the odds shown when it was placed: multipliers can't
//	   change while a question has open bets.
//	R6 A question's answers can't be changed once users have answered it
//	   (resending them unchanged is fine and keeps them).
//	R7 Bet limits must be consistent (percentages 0..100, min <= max) and are
//	   rejected with a validation error, not a raw database error.
//	R8 A question can't award points and take bets at once. Question points
//	   are only awarded to users who finish the quiz themselves.
//
// Voiding a bet (question deleted, submission reset while OPEN) is a refund
// and is fine.
//
// Every subtest uses its own users so scores never mix between scenarios.
func TestQuizBettingEdgeCases(t *testing.T) {
	ctx := context.Background()
	dbMgr, _ := GetTestEnv()

	require.NoError(t, dbMgr.Clean(ctx))
	cfg := testutil.DefaultSeedConfig()
	cfg.NumUsers = 60
	data, err := dbMgr.Seed(ctx, 42, cfg)
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

	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	// ---------------------------------------------------------------- helpers

	type player struct{ id, token string }
	nextUser := 0
	newPlayer := func(t *testing.T) player {
		t.Helper()
		for nextUser < len(data.UserIDs) && data.UserIDs[nextUser] == adminUserID {
			nextUser++
		}
		require.Less(t, nextUser, len(data.UserIDs), "out of seeded users, raise cfg.NumUsers")
		id := data.UserIDs[nextUser]
		nextUser++
		token, err := testutil.GenerateUserToken(id)
		require.NoError(t, err)
		return player{id: id, token: token}
	}

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
		msg := exec(t, adminToken, query, vars, out)
		require.Empty(t, msg)
	}

	// score is the user's project score: the sum of their journal entries
	score := func(t *testing.T, userID string) int {
		t.Helper()
		var s int
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT COALESCE(SUM(points), 0) FROM score_journal WHERE user_id = $1 AND project_id = $2`,
			userID, projectID,
		).Scan(&s))
		return s
	}
	// setScore adjusts the user's project score to exactly target
	setScore := func(t *testing.T, userID string, target int) {
		t.Helper()
		delta := target - score(t, userID)
		if delta != 0 {
			admin(t, `mutation($input: CreateScoreAdjustmentInput!) { createScoreAdjustment(input: $input) { id } }`,
				map[string]any{"input": map[string]any{"projectId": projectID, "userId": userID, "points": delta, "reason": "betting test"}}, nil)
		}
		require.Equal(t, target, score(t, userID))
	}

	newQuiz := func(t *testing.T, name string, extra map[string]any) string {
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
		challengeID := challenge.CreateChallenge.ID
		past := time.Now().Add(-time.Hour).Format(time.RFC3339)
		admin(t, `mutation($id: ID!, $at: DateTime!) { publishChallenge(id: $id, publishedAt: $at) { id } }`,
			map[string]any{"id": challengeID, "at": past}, nil)
		admin(t, `mutation($id: ID!, $at: DateTime!) { setChallengeVisibility(id: $id, visibleAt: $at) { id } }`,
			map[string]any{"id": challengeID, "at": past}, nil)

		input := map[string]any{
			"projectId": projectID, "challengeId": challengeID, "name": name, "description": "x",
			"randomizeQuestions": false, "revealCorrectAnswers": true, "allowRetakes": false, "completionPoints": 0,
		}
		for k, v := range extra {
			input[k] = v
		}
		var quiz struct {
			CreateQuiz struct{ ID string } `json:"createQuiz"`
		}
		admin(t, `mutation($input: CreateQuizInput!) { createQuiz(input: $input) { id } }`,
			map[string]any{"input": input}, &quiz)
		return quiz.CreateQuiz.ID
	}

	// addQuestion adds any question type and returns its ID
	addQuestion := func(t *testing.T, quizID string, input map[string]any) string {
		t.Helper()
		var res struct {
			AddQuizQuestion struct{ ID string } `json:"addQuizQuestion"`
		}
		admin(t, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
			addQuizQuestion(quizId: $quizId, input: $input) { id }
		}`, map[string]any{"quizId": quizID, "input": input}, &res)
		return res.AddQuizQuestion.ID
	}

	type predefined struct{ ID, Right, Wrong string }
	// addPredefined adds a single-choice question with one right and one wrong
	// answer; betting is enabled unless extra overrides it
	addPredefined := func(t *testing.T, quizID string, order int, extra map[string]any) predefined {
		t.Helper()
		input := map[string]any{
			"questionType": "PREDEFINED", "questionText": fmt.Sprintf("Question %d?", order), "questionOrder": order,
			"bettingEnabled": true,
			"predefinedAnswers": []map[string]any{
				{"answerText": "right", "isCorrect": true, "answerOrder": 0},
				{"answerText": "wrong", "isCorrect": false, "answerOrder": 1},
			},
		}
		for k, v := range extra {
			input[k] = v
		}
		var res struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID         string `json:"id"`
					AnswerText string `json:"answerText"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		admin(t, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
			addQuizQuestion(quizId: $quizId, input: $input) {
				... on PredefinedQuestion { id predefinedAnswers { id answerText } }
			}
		}`, map[string]any{"quizId": quizID, "input": input}, &res)
		q := predefined{ID: res.AddQuizQuestion.ID}
		for _, a := range res.AddQuizQuestion.PredefinedAnswers {
			if a.AnswerText == "right" {
				q.Right = a.ID
			} else {
				q.Wrong = a.ID
			}
		}
		require.NotEmpty(t, q.Right)
		require.NotEmpty(t, q.Wrong)
		return q
	}

	addOrdering := func(t *testing.T, quizID string, order int) (id string, items []string) {
		t.Helper()
		var res struct {
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
			"questionType": "ORDERING", "questionText": "Order", "questionOrder": order, "bettingEnabled": true,
			"orderingItems": []map[string]any{{"itemText": "a", "correctOrder": 0}, {"itemText": "b", "correctOrder": 1}},
		}}, &res)
		for _, it := range res.AddQuizQuestion.OrderingItems {
			items = append(items, it.ID)
		}
		require.Len(t, items, 2)
		return res.AddQuizQuestion.ID, items
	}

	// openSession creates a session for quizID, grants the players access and opens it
	openSession := func(t *testing.T, quizID string, players ...player) string {
		t.Helper()
		var session struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		admin(t, `mutation($input: CreateQuizSessionInput!) { createQuizSession(input: $input) { id } }`,
			map[string]any{"input": map[string]any{"quizId": quizID}}, &session)
		id := session.CreateQuizSession.ID
		userIDs := make([]string, len(players))
		for i, p := range players {
			userIDs[i] = p.id
		}
		admin(t, `mutation($input: GrantQuizSessionAccessInput!) { grantQuizSessionAccess(input: $input) }`,
			map[string]any{"input": map[string]any{"sessionId": id, "userIds": userIDs}}, nil)
		admin(t, `mutation($id: ID!) { openQuizSession(id: $id) { id } }`, map[string]any{"id": id}, nil)
		return id
	}
	lock := func(t *testing.T, sessionID string) {
		t.Helper()
		admin(t, `mutation($id: ID!) { lockQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
	}
	reopen := func(t *testing.T, sessionID string) {
		t.Helper()
		admin(t, `mutation($id: ID!) { reopenQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
	}
	finish := func(t *testing.T, sessionID string) {
		t.Helper()
		admin(t, `mutation($id: ID!) { finishQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
	}
	lockAndFinish := func(t *testing.T, sessionID string) {
		t.Helper()
		lock(t, sessionID)
		finish(t, sessionID)
	}

	tryStart := func(t *testing.T, p player, sessionID string) (string, string) {
		t.Helper()
		var res struct {
			StartQuizSession struct{ ID string } `json:"startQuizSession"`
		}
		msg := exec(t, p.token, `mutation($sessionId: ID!) { startQuizSession(sessionId: $sessionId) { id } }`,
			map[string]any{"sessionId": sessionID}, &res)
		return res.StartQuizSession.ID, msg
	}
	start := func(t *testing.T, p player, sessionID string) string {
		t.Helper()
		id, msg := tryStart(t, p, sessionID)
		require.Empty(t, msg)
		return id
	}

	// trySubmit submits an answer and returns the response ID and the error message
	trySubmit := func(t *testing.T, p player, submissionID string, input map[string]any) (string, string) {
		t.Helper()
		var res struct {
			SubmitQuizAnswer struct{ ID string } `json:"submitQuizAnswer"`
		}
		msg := exec(t, p.token, `mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
			submitQuizAnswer(submissionId: $submissionId, input: $input) { id }
		}`, map[string]any{"submissionId": submissionID, "input": input}, &res)
		return res.SubmitQuizAnswer.ID, msg
	}
	submit := func(t *testing.T, p player, submissionID string, input map[string]any) string {
		t.Helper()
		id, msg := trySubmit(t, p, submissionID, input)
		require.Empty(t, msg)
		require.NotEmpty(t, id)
		return id
	}
	// pick builds an answer to a predefined question, right or wrong, with a bet
	pick := func(q predefined, right bool, bet int) map[string]any {
		answer := q.Wrong
		if right {
			answer = q.Right
		}
		return map[string]any{"questionId": q.ID, "selectedAnswerIds": []string{answer}, "betAmount": bet}
	}

	type storedResponse struct {
		BetAmount      *int32
		PointsEarned   *int32
		ScoreJournalID *string
		IsCorrect      *bool
	}
	response := func(t *testing.T, responseID string) storedResponse {
		t.Helper()
		var r storedResponse
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT bet_amount, points_earned, score_journal_id, is_correct FROM quiz_responses WHERE id = $1`, responseID,
		).Scan(&r.BetAmount, &r.PointsEarned, &r.ScoreJournalID, &r.IsCorrect))
		return r
	}
	responseExists := func(t *testing.T, responseID string) bool {
		t.Helper()
		var exists bool
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM quiz_responses WHERE id = $1)`, responseID,
		).Scan(&exists))
		return exists
	}
	betJournal := func(t *testing.T, responseID string) (count int, sum int) {
		t.Helper()
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT count(*), COALESCE(sum(points), 0) FROM score_journal WHERE source_type = 'BET' AND source_id = $1`, responseID,
		).Scan(&count, &sum))
		return count, sum
	}
	// assertSettled checks a bet was paid out once with the given net points
	assertSettled := func(t *testing.T, responseID string, wantNet int) {
		t.Helper()
		r := response(t, responseID)
		if assert.NotNil(t, r.ScoreJournalID, "bet should be settled") && assert.NotNil(t, r.PointsEarned) {
			assert.Equal(t, int32(wantNet), *r.PointsEarned, "points_earned holds the net bet result")
		}
		count, sum := betJournal(t, responseID)
		assert.Equal(t, 2, count, "stake and winnings journal entries")
		assert.Equal(t, wantNet, sum, "journal sum is the net bet result")
	}
	assertUnsettled := func(t *testing.T, responseID string) {
		t.Helper()
		assert.Nil(t, response(t, responseID).ScoreJournalID, "bet should not be settled")
		count, _ := betJournal(t, responseID)
		assert.Zero(t, count, "no bet journal entries")
	}

	// tryUpdateQuestion returns the error message ("" on success)
	tryUpdateQuestion := func(t *testing.T, questionID string, input map[string]any) string {
		t.Helper()
		return exec(t, adminToken, `mutation($id: ID!, $input: UpdateQuizQuestionInput!) { updateQuizQuestion(id: $id, input: $input) { id } }`,
			map[string]any{"id": questionID, "input": input}, nil)
	}

	// ============================================================ several questions

	t.Run("one session, several questions, mixed results are each settled", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 1000)

		quizID := newQuiz(t, "Mixed results", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)
		custom := map[string]any{"bettingMultiplierCorrect": 3.0, "bettingMultiplierWrong": 0.5}
		q3 := addPredefined(t, quizID, 2, custom)
		q4 := addPredefined(t, quizID, 3, custom)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r1 := submit(t, p, sub, pick(q1, true, 100))  // 2x: +100
		r2 := submit(t, p, sub, pick(q2, false, 200)) // 0x: -200
		r3 := submit(t, p, sub, pick(q3, true, 50))   // 3x: +100
		r4 := submit(t, p, sub, pick(q4, false, 100)) // 0.5x: -50

		assert.Equal(t, 1000, score(t, p.id), "nothing is paid out while the session is open")
		lockAndFinish(t, sessionID)

		assertSettled(t, r1, 100)
		assertSettled(t, r2, -200)
		assertSettled(t, r3, 100)
		assertSettled(t, r4, -50)
		assert.Equal(t, 950, score(t, p.id))
	})

	t.Run("questions the user never answers have no bet and cost nothing", func(t *testing.T) {
		p := newPlayer(t)
		idle := newPlayer(t) // has access but never starts
		setScore(t, p.id, 500)
		setScore(t, idle.id, 500)

		quizID := newQuiz(t, "Partially answered", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		addPredefined(t, quizID, 1, nil)
		addPredefined(t, quizID, 2, nil)

		sessionID := openSession(t, quizID, p, idle)
		sub := start(t, p, sessionID)
		r1 := submit(t, p, sub, pick(q1, true, 100))
		lockAndFinish(t, sessionID)

		assertSettled(t, r1, 100)
		var responses int
		var autoSubmitted bool
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT (SELECT count(*) FROM quiz_responses WHERE submission_id = s.id), s.auto_submitted
			 FROM quiz_submissions s WHERE s.id = $1`, sub,
		).Scan(&responses, &autoSubmitted))
		assert.Equal(t, 1, responses, "unanswered questions have no response row")
		assert.True(t, autoSubmitted, "the open submission is auto-submitted at finish")
		assert.Equal(t, 600, score(t, p.id))

		var idleSubmissions int
		require.NoError(t, dbMgr.DB.Pool.QueryRow(ctx,
			`SELECT count(*) FROM quiz_submissions WHERE session_id = $1 AND user_id = $2`, sessionID, idle.id,
		).Scan(&idleSubmissions))
		assert.Zero(t, idleSubmissions, "a user who never started has no submission")
		assert.Equal(t, 500, score(t, idle.id))
	})

	t.Run("R3: bet without a selected answer is void, the stake is returned", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 500)
		quizID := newQuiz(t, "No selection", nil)
		q := addPredefined(t, quizID, 0, nil)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r := submit(t, p, sub, map[string]any{"questionId": q.ID, "betAmount": 200})
		assert.Nil(t, response(t, r).IsCorrect, "no selection is not graded")

		lockAndFinish(t, sessionID)
		assertSettled(t, r, 0) // stake and refund entries, net 0
		var reasons []string
		rows, err := dbMgr.DB.Pool.Query(ctx,
			`SELECT reason FROM score_journal WHERE source_type = 'BET' AND source_id = $1 ORDER BY points`, r)
		require.NoError(t, err)
		for rows.Next() {
			var reason string
			require.NoError(t, rows.Scan(&reason))
			reasons = append(reasons, reason)
		}
		require.NoError(t, rows.Err())
		require.Len(t, reasons, 2)
		assert.NotEqual(t, reasons[0], reasons[1], "the refund entry says the stake was returned")
		assert.Equal(t, 500, score(t, p.id), "the stake is returned")
	})

	t.Run("bet with an empty selection is graded wrong and lost", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 500)
		quizID := newQuiz(t, "Empty selection", nil)
		q := addPredefined(t, quizID, 0, nil)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r := submit(t, p, sub, map[string]any{"questionId": q.ID, "selectedAnswerIds": []string{}, "betAmount": 200})
		if isCorrect := response(t, r).IsCorrect; assert.NotNil(t, isCorrect) {
			assert.False(t, *isCorrect)
		}

		lockAndFinish(t, sessionID)
		assertSettled(t, r, -200)
		assert.Equal(t, 300, score(t, p.id))
	})

	// ============================================================ betting more than you have

	t.Run("R1: bets on several questions cannot add up to more than the score", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Over-bet questions", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)
		q3 := addPredefined(t, quizID, 2, nil)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r1 := submit(t, p, sub, pick(q1, false, 60))
		_, msg := trySubmit(t, p, sub, pick(q2, false, 41))
		assert.NotEmpty(t, msg, "60 of 100 is already staked, 41 more must be rejected")
		r2 := submit(t, p, sub, pick(q2, false, 40)) // exactly what is left
		// Nothing left: R2 lets the user answer with 0
		r3, msg := trySubmit(t, p, sub, pick(q3, false, 1))
		assert.NotEmpty(t, msg, "the whole score is staked, any positive bet must be rejected")
		if msg != "" {
			r3 = submit(t, p, sub, pick(q3, false, 0))
		}

		lockAndFinish(t, sessionID)
		assertSettled(t, r1, -60)
		assertSettled(t, r2, -40)
		assertUnsettled(t, r3)
		assert.Equal(t, 0, score(t, p.id), "betting never makes the score negative")
	})

	t.Run("R1: bets in two open sessions share one score", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizA := newQuiz(t, "Over-bet sessions A", nil)
		qa := addPredefined(t, quizA, 0, nil)
		quizB := newQuiz(t, "Over-bet sessions B", nil)
		qb := addPredefined(t, quizB, 0, nil)

		sessionA := openSession(t, quizA, p)
		sessionB := openSession(t, quizB, p)
		ra := submit(t, p, start(t, p, sessionA), pick(qa, false, 100))
		subB := start(t, p, sessionB)
		_, msg := trySubmit(t, p, subB, pick(qb, false, 100))
		assert.NotEmpty(t, msg, "the 100 is staked in session A")

		lockAndFinish(t, sessionA)
		lockAndFinish(t, sessionB)
		assertSettled(t, ra, -100)
		assert.GreaterOrEqual(t, score(t, p.id), 0, "betting never makes the score negative")
	})

	// submitConcurrently sends all answers at once and returns the error
	// message per answer ("" when accepted)
	submitConcurrently := func(t *testing.T, p player, answers []struct {
		submissionID string
		input        map[string]any
	}) []string {
		t.Helper()
		msgs := make([]string, len(answers))
		var wg sync.WaitGroup
		startGate := make(chan struct{})
		for i, a := range answers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-startGate
				resp, err := client.WithAuth(p.token).Execute(ctx, `mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
					submitQuizAnswer(submissionId: $submissionId, input: $input) { id }
				}`, map[string]any{"submissionId": a.submissionID, "input": a.input})
				switch {
				case err != nil:
					msgs[i] = err.Error()
				case resp.HasErrors():
					msgs[i] = resp.ErrorMessage()
				}
			}()
		}
		close(startGate)
		wg.Wait()
		return msgs
	}
	type answer = struct {
		submissionID string
		input        map[string]any
	}
	assertOnlyBetRejections := func(t *testing.T, msgs []string) (accepted int) {
		t.Helper()
		for _, msg := range msgs {
			if msg == "" {
				accepted++
				continue
			}
			assert.Contains(t, msg, "exceeds available points", "rejected for the points, not a conflict or DB error")
		}
		return accepted
	}

	t.Run("R1: concurrent bets in two sessions cannot both spend the same points", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizA := newQuiz(t, "Concurrent sessions A", nil)
		qa := addPredefined(t, quizA, 0, nil)
		quizB := newQuiz(t, "Concurrent sessions B", nil)
		qb := addPredefined(t, quizB, 0, nil)
		sessionA := openSession(t, quizA, p)
		sessionB := openSession(t, quizB, p)
		subA := start(t, p, sessionA)
		subB := start(t, p, sessionB)

		msgs := submitConcurrently(t, p, []answer{
			{subA, pick(qa, false, 100)},
			{subB, pick(qb, false, 100)},
		})
		assert.Equal(t, 1, assertOnlyBetRejections(t, msgs), "exactly one 100 bet fits in 100 points")

		lockAndFinish(t, sessionA)
		lockAndFinish(t, sessionB)
		assert.Equal(t, 0, score(t, p.id), "betting never makes the score negative")
	})

	t.Run("R1: concurrent bets on several questions stay within the score", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Concurrent questions", nil)
		var answers []answer
		sessionID := openSession(t, quizID, p)
		questions := make([]predefined, 5)
		for i := range questions {
			questions[i] = addPredefined(t, quizID, i, nil)
		}
		sub := start(t, p, sessionID)
		for _, q := range questions {
			answers = append(answers, answer{sub, pick(q, false, 30)})
		}

		msgs := submitConcurrently(t, p, answers)
		assert.Equal(t, 3, assertOnlyBetRejections(t, msgs), "three 30 bets fit in 100 points, a fourth does not")

		lockAndFinish(t, sessionID)
		assert.Equal(t, 10, score(t, p.id))
	})

	t.Run("R1: an open stake in an unfinished session limits bets elsewhere", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Open stake elsewhere", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionA := openSession(t, quizID, p)
		submit(t, p, start(t, p, sessionA), pick(q, true, 70))
		lock(t, sessionA) // locked is still open: not paid out yet

		subB := start(t, p, openSession(t, quizID, p))
		_, msg := trySubmit(t, p, subB, pick(q, true, 31))
		assert.Contains(t, msg, "exceeds available points (30)")
		submit(t, p, subB, pick(q, true, 30))
	})

	t.Run("R1: a stake is released when its session finishes", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Stake released", nil)
		q := addPredefined(t, quizID, 0, nil)

		sessionA := openSession(t, quizID, p)
		submit(t, p, start(t, p, sessionA), pick(q, false, 30))
		lockAndFinish(t, sessionA) // lost: 70 left, nothing open
		sessionB := openSession(t, quizID, p)
		submit(t, p, start(t, p, sessionB), pick(q, true, 70))
	})

	t.Run("R2: with nothing to bet the user can still answer with a 0 bet", func(t *testing.T) {
		for _, start0 := range []int{0, -50} {
			t.Run(fmt.Sprintf("score %d", start0), func(t *testing.T) {
				p := newPlayer(t)
				setScore(t, p.id, start0)
				quizID := newQuiz(t, fmt.Sprintf("Nothing to bet %d", start0), nil)
				q := addPredefined(t, quizID, 0, nil)
				sessionID := openSession(t, quizID, p)
				sub := start(t, p, sessionID)

				_, msg := trySubmit(t, p, sub, pick(q, true, 1))
				assert.NotEmpty(t, msg, "no positive bet is possible")
				r, msg := trySubmit(t, p, sub, pick(q, true, 0))
				assert.Empty(t, msg, "the question must still be answerable")

				lockAndFinish(t, sessionID)
				if r != "" {
					assertUnsettled(t, r)
				}
				assert.Equal(t, start0, score(t, p.id))
			})
		}
	})

	t.Run("R2: when the minimum bet is more than the user has, a 0 bet is accepted", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 50)
		quizID := newQuiz(t, "Minimum unaffordable", nil)
		q := addPredefined(t, quizID, 0, map[string]any{"bettingMinAbsolute": 100})
		sub := start(t, p, openSession(t, quizID, p))

		_, msg := trySubmit(t, p, sub, pick(q, true, 50))
		assert.NotEmpty(t, msg, "below the minimum")
		_, msg = trySubmit(t, p, sub, pick(q, true, 0))
		assert.Empty(t, msg, "the question must still be answerable")
	})

	t.Run("bet equal to the score is accepted, one more is rejected", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 50)
		quizID := newQuiz(t, "Exact score", nil)
		q := addPredefined(t, quizID, 0, nil)
		sub := start(t, p, openSession(t, quizID, p))

		_, msg := trySubmit(t, p, sub, pick(q, true, 51))
		assert.Contains(t, msg, "exceeds available points (50)")
		// The rejected attempt stores nothing, so the question can still be answered
		r := submit(t, p, sub, pick(q, true, 50))
		assert.Equal(t, int32(50), *response(t, r).BetAmount)
	})

	// ============================================================ several sessions

	t.Run("finishing one session settles only its bets, the other waits", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 200)
		quizID := newQuiz(t, "Two sessions same quiz", nil)
		q := addPredefined(t, quizID, 0, nil)

		// Two sessions of the same quiz, both open at once
		session1 := openSession(t, quizID, p)
		session2 := openSession(t, quizID, p)
		r1 := submit(t, p, start(t, p, session1), pick(q, true, 100))
		r2 := submit(t, p, start(t, p, session2), pick(q, true, 100))

		lockAndFinish(t, session1)
		assertSettled(t, r1, 100)
		assertUnsettled(t, r2)
		assert.Equal(t, 300, score(t, p.id))

		lockAndFinish(t, session2)
		assertSettled(t, r2, 100)
		assert.Equal(t, 400, score(t, p.id))
	})

	t.Run("winnings from a finished session raise the limit in the next one", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Winnings carry over", nil)
		q := addPredefined(t, quizID, 0, nil)

		session1 := openSession(t, quizID, p)
		r1 := submit(t, p, start(t, p, session1), pick(q, true, 100))
		lockAndFinish(t, session1)
		assertSettled(t, r1, 100)

		sub2 := start(t, p, openSession(t, quizID, p))
		_, msg := trySubmit(t, p, sub2, pick(q, true, 201))
		assert.Contains(t, msg, "exceeds available points (200)")
		submit(t, p, sub2, pick(q, true, 200))
	})

	t.Run("losses from a finished session lower the limit in the next one", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Losses carry over", nil)
		q := addPredefined(t, quizID, 0, nil)

		session1 := openSession(t, quizID, p)
		submit(t, p, start(t, p, session1), pick(q, false, 70))
		lockAndFinish(t, session1)
		assert.Equal(t, 30, score(t, p.id))

		sub2 := start(t, p, openSession(t, quizID, p))
		_, msg := trySubmit(t, p, sub2, pick(q, true, 31))
		assert.Contains(t, msg, "exceeds available points (30)")
	})

	t.Run("reopened session: bets from before and after reopening are settled once", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 1000)
		quizID := newQuiz(t, "Reopened", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r1 := submit(t, p, sub, pick(q1, true, 100))
		lock(t, sessionID)
		reopen(t, sessionID)
		assertUnsettled(t, r1)
		r2 := submit(t, p, sub, pick(q2, false, 100))

		lockAndFinish(t, sessionID)
		finish(t, sessionID) // finishing again pays nothing twice
		assertSettled(t, r1, 100)
		assertSettled(t, r2, -100)
		assert.Equal(t, 1000, score(t, p.id))
	})

	// ============================================================ answer paths and session states

	t.Run("betting-enabled question cannot be answered without a bet", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Bet required", nil)
		q := addPredefined(t, quizID, 0, nil)
		sub := start(t, p, openSession(t, quizID, p))

		_, msg := trySubmit(t, p, sub, map[string]any{"questionId": q.ID, "selectedAnswerIds": []string{q.Right}})
		assert.Contains(t, msg, "bet is required when betting is enabled")
	})

	t.Run("R2: a bet of 0 means no bet, also with a minimum", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Zero bet", nil)
		q := addPredefined(t, quizID, 0, nil)
		withMinimum := addPredefined(t, quizID, 1, map[string]any{"bettingMinAbsolute": 20})
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)

		r := submit(t, p, sub, pick(q, true, 0))
		rMin := submit(t, p, sub, pick(withMinimum, false, 0))

		lockAndFinish(t, sessionID)
		assertUnsettled(t, r)
		assertUnsettled(t, rMin)
		assert.Equal(t, 100, score(t, p.id), "no bet, nothing won or lost")
	})

	t.Run("negative bet is rejected", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Negative bet", nil)
		q := addPredefined(t, quizID, 0, nil)
		sub := start(t, p, openSession(t, quizID, p))

		_, msg := trySubmit(t, p, sub, pick(q, true, -10))
		assert.Contains(t, msg, "cannot be negative")
	})

	t.Run("bet on a question without betting is rejected, a plain answer is not a bet", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Betting disabled", nil)
		q := addPredefined(t, quizID, 0, map[string]any{"bettingEnabled": false})
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)

		_, msg := trySubmit(t, p, sub, pick(q, true, 10))
		assert.Contains(t, msg, "betting is not enabled")

		r := submit(t, p, sub, map[string]any{"questionId": q.ID, "selectedAnswerIds": []string{q.Right}})
		lockAndFinish(t, sessionID)
		assertUnsettled(t, r)
		assert.Equal(t, 100, score(t, p.id))
	})

	t.Run("answering a question twice keeps the first answer and bet", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 1000)
		quizID := newQuiz(t, "Answered twice", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)

		first := submit(t, p, sub, pick(q, false, 50))
		second := submit(t, p, sub, pick(q, true, 100))
		assert.Equal(t, first, second, "the stored response is returned")
		stored := response(t, first)
		assert.Equal(t, int32(50), *stored.BetAmount)
		assert.False(t, *stored.IsCorrect)

		lockAndFinish(t, sessionID)
		assertSettled(t, first, -50)
		assert.Equal(t, 950, score(t, p.id))
	})

	t.Run("answering a second time with another bet cannot change the stored bet", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Answered twice, other bet", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)

		first := submit(t, p, sub, pick(q, false, 50))
		trySubmit(t, p, sub, pick(q, true, 500)) // rejected or ignored, either is fine
		trySubmit(t, p, sub, pick(q, true, 10))
		assert.Equal(t, int32(50), *response(t, first).BetAmount)

		lockAndFinish(t, sessionID)
		assertSettled(t, first, -50)
		assert.Equal(t, 50, score(t, p.id))
	})

	t.Run("R4: no answers or bets while the session is LOCKED", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Locked answers", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		lock(t, sessionID)

		// Answers may be revealed once locked; a bet now could be a sure win
		r, msg := trySubmit(t, p, sub, pick(q, true, 100))
		assert.NotEmpty(t, msg, "submitQuizAnswer must reject answers in a LOCKED session")

		finish(t, sessionID)
		if r != "" {
			assertUnsettled(t, r)
		}
		assert.Equal(t, 100, score(t, p.id))
	})

	t.Run("a LOCKED or FINISHED session cannot be started", func(t *testing.T) {
		p := newPlayer(t)
		quizID := newQuiz(t, "Start after lock", nil)
		addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		lock(t, sessionID)

		_, msg := tryStart(t, p, sessionID)
		assert.Contains(t, msg, "session is not open")
		finish(t, sessionID)
		_, msg = tryStart(t, p, sessionID)
		assert.Contains(t, msg, "session is not open")
	})

	t.Run("no bets after the session finished, even on unanswered questions", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Bet after finish", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		submit(t, p, sub, pick(q1, true, 10))
		lockAndFinish(t, sessionID)

		_, msg := trySubmit(t, p, sub, pick(q2, true, 10))
		assert.Contains(t, msg, "submission already completed")
	})

	t.Run("after finalizeQuiz no more bets, earlier bets are settled at finish", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Finalized early", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r1 := submit(t, p, sub, pick(q1, true, 100))

		msg := exec(t, p.token, `mutation($id: ID!) { finalizeQuiz(submissionId: $id) { id } }`,
			map[string]any{"id": sub}, nil)
		require.Empty(t, msg)
		assertUnsettled(t, r1)

		_, msg = trySubmit(t, p, sub, pick(q2, true, 10))
		assert.Contains(t, msg, "submission already completed")

		lockAndFinish(t, sessionID)
		assertSettled(t, r1, 100)
		assert.Equal(t, 200, score(t, p.id))
	})

	t.Run("R8: a question can't award points and take bets at once", func(t *testing.T) {
		quizID := newQuiz(t, "Points or bet", nil)
		addInput := func(extra map[string]any) map[string]any {
			input := map[string]any{
				"questionType": "PREDEFINED", "questionText": "Points?", "questionOrder": 0,
				"predefinedAnswers": []map[string]any{
					{"answerText": "right", "isCorrect": true, "answerOrder": 0},
					{"answerText": "wrong", "isCorrect": false, "answerOrder": 1},
				},
			}
			for k, v := range extra {
				input[k] = v
			}
			return input
		}
		msg := exec(t, adminToken, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
			addQuizQuestion(quizId: $quizId, input: $input) { id }
		}`, map[string]any{"quizId": quizID, "input": addInput(map[string]any{"points": 10, "bettingEnabled": true})}, nil)
		assert.Contains(t, msg, "a question with betting cannot also award points")

		betting := addPredefined(t, quizID, 1, nil)
		assert.Contains(t, tryUpdateQuestion(t, betting.ID, map[string]any{"points": 10}),
			"a question with betting cannot also award points")
		assert.Empty(t, tryUpdateQuestion(t, betting.ID, map[string]any{"points": 0}), "0 points is no points")

		pointsOnly := addPredefined(t, quizID, 2, map[string]any{"points": 10, "bettingEnabled": false})
		assert.Contains(t, tryUpdateQuestion(t, pointsOnly.ID, map[string]any{"bettingEnabled": true}),
			"a question with betting cannot also award points")

		// An older question with both stays editable as long as the edit doesn't touch them
		legacy := addPredefined(t, quizID, 3, nil)
		_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE quiz_questions SET points = 10 WHERE id = $1`, legacy.ID)
		require.NoError(t, err)
		assert.Empty(t, tryUpdateQuestion(t, legacy.ID, map[string]any{"questionText": "Fixed typo?"}))
	})

	t.Run("R8: question points only for users who finish the quiz themselves", func(t *testing.T) {
		finalizer := newPlayer(t)
		autoSubmitted := newPlayer(t)
		setScore(t, finalizer.id, 1000)
		setScore(t, autoSubmitted.id, 1000)
		quizID := newQuiz(t, "Points need finishing", nil)
		q := addPredefined(t, quizID, 0, map[string]any{"points": 10, "bettingEnabled": false})
		sessionID := openSession(t, quizID, finalizer, autoSubmitted)

		correct := map[string]any{"questionId": q.ID, "selectedAnswerIds": []string{q.Right}}
		subF := start(t, finalizer, sessionID)
		submit(t, finalizer, subF, correct)
		msg := exec(t, finalizer.token, `mutation($id: ID!) { finalizeQuiz(submissionId: $id) { id } }`,
			map[string]any{"id": subF}, nil)
		require.Empty(t, msg)
		submit(t, autoSubmitted, start(t, autoSubmitted, sessionID), correct)

		lockAndFinish(t, sessionID)
		assert.Equal(t, 1010, score(t, finalizer.id), "finished: question points")
		assert.Equal(t, 1000, score(t, autoSubmitted.id), "auto-submitted at session finish: no question points")
	})

	t.Run("expired submission rejects bets, bets placed in time are settled", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Expired", map[string]any{"timeoutSeconds": 3600})
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r1 := submit(t, p, sub, pick(q1, true, 100))

		_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE quiz_submissions SET expires_at = now() - interval '1 minute' WHERE id = $1`, sub)
		require.NoError(t, err)
		_, msg := trySubmit(t, p, sub, pick(q2, true, 10))
		assert.Contains(t, msg, "submission expired")

		lockAndFinish(t, sessionID)
		assertSettled(t, r1, 100)
		assert.Equal(t, 200, score(t, p.id))
	})

	t.Run("resetting the submission throws away the bets placed so far", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Reset", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)

		sub1 := start(t, p, sessionID)
		lost := submit(t, p, sub1, pick(q, false, 100))
		msg := exec(t, p.token, `mutation($id: ID!) { resetQuizSessionSubmission(sessionId: $id) }`,
			map[string]any{"id": sessionID}, nil)
		require.Empty(t, msg)
		assert.False(t, responseExists(t, lost), "the response and its bet are deleted with the submission")

		sub2 := start(t, p, sessionID)
		assert.NotEqual(t, sub1, sub2)
		won := submit(t, p, sub2, pick(q, true, 100))

		lockAndFinish(t, sessionID)
		assertSettled(t, won, 100)
		assert.Equal(t, 200, score(t, p.id))
	})

	t.Run("resetting is not possible once the session is LOCKED", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Reset locked", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		r := submit(t, p, start(t, p, sessionID), pick(q, false, 100))
		lock(t, sessionID)

		msg := exec(t, p.token, `mutation($id: ID!) { resetQuizSessionSubmission(sessionId: $id) }`,
			map[string]any{"id": sessionID}, nil)
		assert.Contains(t, msg, "can only reset submission while session is open")
		finish(t, sessionID)
		assertSettled(t, r, -100)
	})

	// ============================================================ question types

	t.Run("bets on FREE_TEXT, NUMBER and ORDERING questions are not settled by the core", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 300)
		quizID := newQuiz(t, "Other types", nil)
		freeTextID := addQuestion(t, quizID, map[string]any{
			"questionType": "FREE_TEXT", "questionText": "Why?", "questionOrder": 0, "bettingEnabled": true,
		})
		numberID := addQuestion(t, quizID, map[string]any{
			"questionType": "NUMBER", "questionText": "How many?", "questionOrder": 1, "bettingEnabled": true,
			"minValue": 0, "maxValue": 100, "stepValue": 1,
		})
		orderingID, items := addOrdering(t, quizID, 2)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		rText := submit(t, p, sub, map[string]any{"questionId": freeTextID, "textResponse": "because", "betAmount": 100})
		rNumber := submit(t, p, sub, map[string]any{"questionId": numberID, "numberResponse": 42.0, "betAmount": 100})
		rOrdering := submit(t, p, sub, map[string]any{"questionId": orderingID, "submittedOrder": items, "betAmount": 100})

		assert.Nil(t, response(t, rText).IsCorrect)
		assert.Nil(t, response(t, rNumber).IsCorrect)
		assert.True(t, *response(t, rOrdering).IsCorrect)

		lockAndFinish(t, sessionID)
		assertUnsettled(t, rText)
		assertUnsettled(t, rNumber)
		assertUnsettled(t, rOrdering) // left to the ladder_to_heaven plugin
		assert.Equal(t, 300, score(t, p.id))
	})

	t.Run("updateQuizAnswer on an ORDERING bet", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Update ordering bet", nil)
		orderingID, items := addOrdering(t, quizID, 0)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r := submit(t, p, sub, map[string]any{"questionId": orderingID, "submittedOrder": items, "betAmount": 60})

		update := func(input map[string]any) string {
			return exec(t, p.token, `mutation($id: ID!, $input: UpdateQuizAnswerInput!) {
				updateQuizAnswer(responseId: $id, input: $input) { id }
			}`, map[string]any{"id": r, "input": input}, nil)
		}
		reversed := []string{items[1], items[0]}

		// The new bet replaces the old one and is checked against the score alone
		assert.Empty(t, update(map[string]any{"submittedOrder": reversed, "betAmount": 100}))
		assert.Equal(t, int32(100), *response(t, r).BetAmount)
		assert.False(t, *response(t, r).IsCorrect)
		assert.Contains(t, update(map[string]any{"submittedOrder": items, "betAmount": 101}), "exceeds available points (100)")

		lock(t, sessionID)
		assert.Contains(t, update(map[string]any{"submittedOrder": items, "betAmount": 10}), "quiz session is LOCKED")
		finish(t, sessionID)
		assertUnsettled(t, r)
	})

	// ============================================================ payouts

	t.Run("multiplier edge values", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 1000)
		quizID := newQuiz(t, "Multiplier edges", nil)
		correctZero := addPredefined(t, quizID, 0, map[string]any{"bettingMultiplierCorrect": 0.0})
		refund := map[string]any{"bettingMultiplierCorrect": 1.0, "bettingMultiplierWrong": 1.0}
		refundRight := addPredefined(t, quizID, 1, refund)
		refundWrong := addPredefined(t, quizID, 2, refund)
		maxMult := addPredefined(t, quizID, 3, map[string]any{"bettingMultiplierCorrect": 100.0})
		fraction := addPredefined(t, quizID, 4, map[string]any{"bettingMultiplierCorrect": 1.5})
		wrongWins := addPredefined(t, quizID, 5, map[string]any{"bettingMultiplierCorrect": 3.0, "bettingMultiplierWrong": 1.25})

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		rZero := submit(t, p, sub, pick(correctZero, true, 100))
		rRefundRight := submit(t, p, sub, pick(refundRight, true, 100))
		rRefundWrong := submit(t, p, sub, pick(refundWrong, false, 100))
		rMax := submit(t, p, sub, pick(maxMult, true, 10))
		rFraction := submit(t, p, sub, pick(fraction, true, 3))
		rWrongWins := submit(t, p, sub, pick(wrongWins, false, 4))

		lockAndFinish(t, sessionID)
		assertSettled(t, rZero, -100)     // correct answer, 0x: still lost
		assertSettled(t, rRefundRight, 0) // 1x: money back
		assertSettled(t, rRefundWrong, 0) // 1x: money back
		assertSettled(t, rMax, 990)       // 100x
		assertSettled(t, rFraction, 1)    // floor(3 * 1.5) = 4
		assertSettled(t, rWrongWins, 1)   // wrong answer, 1.25x: floor(5) - 4
		assert.Equal(t, 1000-100+990+1+1, score(t, p.id))
	})

	t.Run("a bet that cannot be paid out does not block the others and is retried", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 1000)
		quizID := newQuiz(t, "Broken bet", nil)
		good := addPredefined(t, quizID, 0, nil)
		bad := addPredefined(t, quizID, 1, nil)

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		rGood := submit(t, p, sub, pick(good, true, 100))
		rBad := submit(t, p, sub, pick(bad, true, 100))
		// 2x of MaxInt32 does not fit the points range
		_, err := dbMgr.DB.Pool.Exec(ctx, `UPDATE quiz_responses SET bet_amount = $2 WHERE id = $1`, rBad, math.MaxInt32)
		require.NoError(t, err)

		lockAndFinish(t, sessionID)
		assertSettled(t, rGood, 100)
		assertUnsettled(t, rBad)
		assert.Equal(t, 1100, score(t, p.id))

		_, err = dbMgr.DB.Pool.Exec(ctx, `UPDATE quiz_responses SET bet_amount = 10 WHERE id = $1`, rBad)
		require.NoError(t, err)
		finish(t, sessionID)
		assertSettled(t, rBad, 10)
		assertSettled(t, rGood, 100)
		assert.Equal(t, 1110, score(t, p.id))
	})

	t.Run("multi-select: picking only one of two correct answers loses the bet", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Multi-select", nil)
		var res struct {
			AddQuizQuestion struct {
				ID                string `json:"id"`
				PredefinedAnswers []struct {
					ID         string `json:"id"`
					AnswerText string `json:"answerText"`
				} `json:"predefinedAnswers"`
			} `json:"addQuizQuestion"`
		}
		admin(t, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
			addQuizQuestion(quizId: $quizId, input: $input) {
				... on PredefinedQuestion { id predefinedAnswers { id answerText } }
			}
		}`, map[string]any{"quizId": quizID, "input": map[string]any{
			"questionType": "PREDEFINED", "questionText": "Pick all", "questionOrder": 0,
			"bettingEnabled": true, "allowMultipleSelection": true,
			"predefinedAnswers": []map[string]any{
				{"answerText": "a", "isCorrect": true, "answerOrder": 0},
				{"answerText": "b", "isCorrect": true, "answerOrder": 1},
				{"answerText": "c", "isCorrect": false, "answerOrder": 2},
			},
		}}, &res)
		ids := map[string]string{}
		for _, a := range res.AddQuizQuestion.PredefinedAnswers {
			ids[a.AnswerText] = a.ID
		}
		qID := res.AddQuizQuestion.ID

		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		r := submit(t, p, sub, map[string]any{"questionId": qID, "selectedAnswerIds": []string{ids["a"]}, "betAmount": 40})
		lockAndFinish(t, sessionID)
		assertSettled(t, r, -40) // all or nothing, no partial credit
		assert.Equal(t, 60, score(t, p.id))
	})

	// ============================================================ admin edits after bets were placed

	t.Run("R5: multipliers can't change while the question has open bets", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Multiplier changed", nil)
		q := addPredefined(t, quizID, 0, nil) // default 2x when the bet is placed
		sessionID := openSession(t, quizID, p)
		r := submit(t, p, start(t, p, sessionID), pick(q, true, 100))

		assert.Contains(t, tryUpdateQuestion(t, q.ID, map[string]any{"bettingMultiplierCorrect": 5.0}), "open bets")
		assert.Empty(t, tryUpdateQuestion(t, q.ID, map[string]any{"bettingMultiplierCorrect": 2.0}),
			"setting the default explicitly doesn't change any payout")
		lockAndFinish(t, sessionID)
		assertSettled(t, r, 100) // the 2x the user bet on
		assert.Equal(t, 200, score(t, p.id))

		assert.Empty(t, tryUpdateQuestion(t, q.ID, map[string]any{"bettingMultiplierCorrect": 5.0}),
			"allowed again once the bets are settled")
	})

	t.Run("R3: betting can't be turned off while the question has open bets", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Betting disabled later", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		r := submit(t, p, start(t, p, sessionID), pick(q, false, 100))

		assert.Contains(t, tryUpdateQuestion(t, q.ID, map[string]any{"bettingEnabled": false}), "open bets")
		assert.Empty(t, tryUpdateQuestion(t, q.ID, map[string]any{"bettingEnabled": true}), "unchanged is fine")
		lockAndFinish(t, sessionID)
		assertSettled(t, r, -100)
		assert.Equal(t, 0, score(t, p.id))

		assert.Empty(t, tryUpdateQuestion(t, q.ID, map[string]any{"bettingEnabled": false}),
			"allowed again once the bets are settled")
	})

	t.Run("R6: answers can't be changed once users answered the question", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Answer key changed", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		r := submit(t, p, start(t, p, sessionID), pick(q, true, 100))

		// Admin tries to fix the answer key: the answer the user picked would be wrong
		msg := tryUpdateQuestion(t, q.ID, map[string]any{"predefinedAnswers": []map[string]any{
			{"answerText": "right", "isCorrect": false, "answerOrder": 0},
			{"answerText": "wrong", "isCorrect": true, "answerOrder": 1},
		}})
		assert.Contains(t, msg, "after users have answered")
		lockAndFinish(t, sessionID)
		assertSettled(t, r, 100) // the original key stands
		assert.Equal(t, 200, score(t, p.id))
	})

	t.Run("R6: saving a question with its answers unchanged keeps the answers", func(t *testing.T) {
		// The admin editor sends all answers on every save
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Answers resent", nil)
		q := addPredefined(t, quizID, 0, nil)
		sessionID := openSession(t, quizID, p)
		submit(t, p, start(t, p, sessionID), pick(q, true, 10))

		msg := tryUpdateQuestion(t, q.ID, map[string]any{
			"questionText": "Fixed typo?",
			"predefinedAnswers": []map[string]any{
				{"answerText": "wrong", "isCorrect": false, "answerOrder": 1},
				{"answerText": "right", "isCorrect": true, "answerOrder": 0},
			},
		})
		assert.Empty(t, msg)
		var ids []string
		rows, err := dbMgr.DB.Pool.Query(ctx, `SELECT id FROM quiz_predefined_answers WHERE question_id = $1 ORDER BY answer_order`, q.ID)
		require.NoError(t, err)
		for rows.Next() {
			var id string
			require.NoError(t, rows.Scan(&id))
			ids = append(ids, id)
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, []string{q.Right, q.Wrong}, ids, "the answers keep their IDs, so stored selections stay valid")
	})

	t.Run("deleting a question voids its unsettled bets", func(t *testing.T) {
		p := newPlayer(t)
		setScore(t, p.id, 100)
		quizID := newQuiz(t, "Question deleted", nil)
		keep := addPredefined(t, quizID, 0, nil)
		drop := addPredefined(t, quizID, 1, nil)
		sessionID := openSession(t, quizID, p)
		sub := start(t, p, sessionID)
		rKeep := submit(t, p, sub, pick(keep, true, 50))
		rDrop := submit(t, p, sub, pick(drop, false, 50))

		admin(t, `mutation($id: ID!) { deleteQuizQuestion(id: $id) }`, map[string]any{"id": drop.ID}, nil)
		assert.False(t, responseExists(t, rDrop), "the bet is voided (refunded) with the question")

		lockAndFinish(t, sessionID)
		assertSettled(t, rKeep, 50)
		assert.Equal(t, 150, score(t, p.id))
	})

	t.Run("R7: inconsistent bet limits are rejected", func(t *testing.T) {
		quizID := newQuiz(t, "Bet limit config", nil)
		valid := addPredefined(t, quizID, 0, nil)
		tests := []struct {
			name   string
			limits map[string]any
		}{
			{name: "min percentage above max percentage", limits: map[string]any{"bettingMinPercentage": 60.0, "bettingMaxPercentage": 40.0}},
			{name: "min absolute above max absolute", limits: map[string]any{"bettingMinAbsolute": 100, "bettingMaxAbsolute": 10}},
			{name: "percentage above 100", limits: map[string]any{"bettingMaxPercentage": 150.0}},
			{name: "negative percentage", limits: map[string]any{"bettingMinPercentage": -10.0}},
			{name: "negative absolute", limits: map[string]any{"bettingMinAbsolute": -5}},
		}
		for i, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input := map[string]any{
					"questionType": "PREDEFINED", "questionText": "Limits?", "questionOrder": 10 + i, "bettingEnabled": true,
					"predefinedAnswers": []map[string]any{{"answerText": "right", "isCorrect": true, "answerOrder": 0}},
				}
				for k, v := range tt.limits {
					input[k] = v
				}
				msg := exec(t, adminToken, `mutation($quizId: ID!, $input: CreateQuizQuestionInput!) {
					addQuizQuestion(quizId: $quizId, input: $input) { id }
				}`, map[string]any{"quizId": quizID, "input": input}, nil)
				assert.NotEmpty(t, msg, "create must reject the limits")
				assert.NotContains(t, msg, "SQLSTATE", "a validation error, not a leaked database error")

				msg = tryUpdateQuestion(t, valid.ID, tt.limits)
				assert.NotEmpty(t, msg, "update must reject the limits")
				assert.NotContains(t, msg, "SQLSTATE", "a validation error, not a leaked database error")
			})
		}
	})

	t.Run("several users in one session are settled independently", func(t *testing.T) {
		players := []player{newPlayer(t), newPlayer(t), newPlayer(t)}
		quizID := newQuiz(t, "Many users", nil)
		q1 := addPredefined(t, quizID, 0, nil)
		q2 := addPredefined(t, quizID, 1, nil)
		for _, p := range players {
			setScore(t, p.id, 300)
		}

		sessionID := openSession(t, quizID, players...)
		subs := make([]string, len(players))
		for i, p := range players {
			subs[i] = start(t, p, sessionID)
		}
		// 0: both right, 1: both wrong, 2: one each, bets of 100 and 200
		submit(t, players[0], subs[0], pick(q1, true, 100))
		submit(t, players[0], subs[0], pick(q2, true, 200))
		submit(t, players[1], subs[1], pick(q1, false, 100))
		submit(t, players[1], subs[1], pick(q2, false, 200))
		submit(t, players[2], subs[2], pick(q1, true, 100))
		submit(t, players[2], subs[2], pick(q2, false, 200))

		lockAndFinish(t, sessionID)
		assert.Equal(t, 600, score(t, players[0].id))
		assert.Equal(t, 0, score(t, players[1].id))
		assert.Equal(t, 200, score(t, players[2].id))
	})
}
