package e2e

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/e2e/testutil"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuizBettingLoad hammers the bet write path (per-user submission locks,
// see betCheck) with many users and checks that it never fails: every bet is
// either accepted or rejected for the points, never lost to an error, and no
// score goes negative. A baseline scenario without bets shows what the bet
// check costs.
//
// Opt-in, it takes a while:
//
//	BETTING_LOADTEST=1 go test -v -count=1 -timeout 60m -run TestQuizBettingLoad ./e2e/
//
// Knobs: BETTING_LOAD_USERS (default 300), BETTING_LOAD_QUESTIONS (5),
// BETTING_LOAD_POOL (DB connections, default 25 = production default, which
// must not be raised), BETTING_LOAD_WINDOW (seconds over which users arrive,
// default 0 = all at once), BETTING_LOAD_INFLIGHT (max concurrent HTTP
// requests from the test, default 500; the local socket backlog resets
// connections well before the server is the limit).
func TestQuizBettingLoad(t *testing.T) {
	if os.Getenv("BETTING_LOADTEST") == "" {
		t.Skip("set BETTING_LOADTEST=1 to run the betting load test")
	}
	envInt := func(name string, def int) int {
		if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
			return v
		}
		return def
	}
	numUsers := envInt("BETTING_LOAD_USERS", 300)
	numQuestions := envInt("BETTING_LOAD_QUESTIONS", 5)
	poolSize := envInt("BETTING_LOAD_POOL", 25)
	window := time.Duration(envInt("BETTING_LOAD_WINDOW", 0)) * time.Second
	inflight := make(chan struct{}, envInt("BETTING_LOAD_INFLIGHT", 500))
	const startingScore = 100

	ctx := context.Background()
	dbMgr, _ := GetTestEnv()
	require.NoError(t, dbMgr.Clean(ctx))
	cfg := testutil.DefaultSeedConfig()
	cfg.NumUsers = numUsers + 5
	cfg.AchievementCompletionRate = 0 // not needed here, and the slowest part of seeding 10k users
	data, err := dbMgr.Seed(ctx, 42, cfg)
	require.NoError(t, err)

	adminUserID := data.UserIDs[2]
	require.NoError(t, dbMgr.AssignRole(ctx, adminUserID, testutil.RoleAdmin))

	// The server gets a production-sized pool; the test manager's pool is small
	db, err := database.Connect(ctx, config.DatabaseConfig{
		URL:             dbMgr.DSN,
		MaxOpenConns:    poolSize,
		MaxIdleConns:    poolSize,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	require.NoError(t, err)
	defer db.Close()
	serverMgr := *dbMgr
	serverMgr.DB = db
	router, cleanup, err := testutil.SetupTestServer(ctx, &serverMgr)
	require.NoError(t, err)
	defer cleanup()
	client := testutil.NewGraphQLClient(router)
	defer client.Close()

	adminToken, err := testutil.GenerateAdminToken(adminUserID)
	require.NoError(t, err)
	projectID := data.ProjectIDs[0]
	eventID := data.EventIDs[projectID][0]

	type user struct{ id, token string }
	var users []user
	for _, id := range data.UserIDs {
		if id == adminUserID || len(users) == numUsers {
			continue
		}
		token, err := testutil.GenerateUserToken(id)
		require.NoError(t, err)
		users = append(users, user{id: id, token: token})
	}
	require.Len(t, users, numUsers)
	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.id
	}

	// ---------------------------------------------------------------- helpers

	admin := func(t *testing.T, query string, vars map[string]any, out any) {
		t.Helper()
		resp := client.WithAuth(adminToken).MustExecute(t, query, vars)
		require.False(t, resp.HasErrors(), resp.ErrorMessage())
		if out != nil {
			require.NoError(t, resp.UnmarshalData(out))
		}
	}

	// resetScores sets every user's project score to startingScore
	resetScores := func(t *testing.T) {
		t.Helper()
		rows, err := dbMgr.DB.Pool.Query(ctx,
			`SELECT u, COALESCE((SELECT SUM(points) FROM score_journal WHERE user_id = u AND project_id = $2), 0)
			 FROM unnest($1::char(28)[]) AS u`, userIDs, projectID)
		require.NoError(t, err)
		var entries [][]any
		for rows.Next() {
			var id string
			var score int
			require.NoError(t, rows.Scan(&id, &score))
			if delta := startingScore - score; delta != 0 {
				entries = append(entries, []any{ulid.NewScoreJournalID(), projectID, id, delta, "MANUAL"})
			}
		}
		require.NoError(t, rows.Err())
		_, err = dbMgr.DB.Pool.CopyFrom(ctx, pgx.Identifier{"score_journal"},
			[]string{"id", "project_id", "user_id", "points", "source_type"}, pgx.CopyFromRows(entries))
		require.NoError(t, err)
	}

	type question struct{ id, wrong string }
	// newQuizSession creates a quiz with betting questions (or without, for the
	// baseline) and an open session all users can access
	newQuizSession := func(t *testing.T, name string, betting bool) (string, []question) {
		t.Helper()
		var challenge struct {
			CreateChallenge struct{ ID string } `json:"createChallenge"`
		}
		admin(t, `mutation($projectId: ID!, $eventId: ID!, $input: CreateChallengeInput!) {
			createChallenge(projectId: $projectId, eventId: $eventId, input: $input) { id }
		}`, map[string]any{"projectId": projectID, "eventId": eventID,
			"input": map[string]any{"type": "QUIZ", "name": name, "buttonText": "Go", "description": "<p>x</p>"}}, &challenge)
		past := time.Now().Add(-time.Hour).Format(time.RFC3339)
		admin(t, `mutation($id: ID!, $at: DateTime!) { publishChallenge(id: $id, publishedAt: $at) { id } }`,
			map[string]any{"id": challenge.CreateChallenge.ID, "at": past}, nil)
		var quiz struct {
			CreateQuiz struct{ ID string } `json:"createQuiz"`
		}
		admin(t, `mutation($input: CreateQuizInput!) { createQuiz(input: $input) { id } }`, map[string]any{"input": map[string]any{
			"projectId": projectID, "challengeId": challenge.CreateChallenge.ID, "name": name, "description": "x",
			"randomizeQuestions": false, "revealCorrectAnswers": true, "allowRetakes": false, "completionPoints": 0,
		}}, &quiz)

		questions := make([]question, numQuestions)
		for i := range questions {
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
			}`, map[string]any{"quizId": quiz.CreateQuiz.ID, "input": map[string]any{
				"questionType": "PREDEFINED", "questionText": fmt.Sprintf("Q%d", i), "questionOrder": i, "bettingEnabled": betting,
				"predefinedAnswers": []map[string]any{
					{"answerText": "right", "isCorrect": true, "answerOrder": 0},
					{"answerText": "wrong", "isCorrect": false, "answerOrder": 1},
				},
			}}, &res)
			questions[i].id = res.AddQuizQuestion.ID
			for _, a := range res.AddQuizQuestion.PredefinedAnswers {
				if a.AnswerText == "wrong" {
					questions[i].wrong = a.ID
				}
			}
		}

		var session struct {
			CreateQuizSession struct{ ID string } `json:"createQuizSession"`
		}
		admin(t, `mutation($input: CreateQuizSessionInput!) { createQuizSession(input: $input) { id } }`,
			map[string]any{"input": map[string]any{"quizId": quiz.CreateQuiz.ID}}, &session)
		sessionID := session.CreateQuizSession.ID
		admin(t, `mutation($input: GrantQuizSessionAccessInput!) { grantQuizSessionAccess(input: $input) }`,
			map[string]any{"input": map[string]any{"sessionId": sessionID, "userIds": userIDs}}, nil)
		admin(t, `mutation($id: ID!) { openQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
		return sessionID, questions
	}

	// startAll starts the session for every user (not measured); returns submission IDs by user
	startAll := func(t *testing.T, sessionID string) []string {
		t.Helper()
		subs := make([]string, len(users))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 32)
		var mu sync.Mutex
		var failures []string
		for i, u := range users {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				resp, err := client.WithAuth(u.token).Execute(ctx,
					`mutation($sessionId: ID!) { startQuizSession(sessionId: $sessionId) { id } }`,
					map[string]any{"sessionId": sessionID})
				var res struct {
					StartQuizSession struct{ ID string } `json:"startQuizSession"`
				}
				if err == nil && !resp.HasErrors() && resp.UnmarshalData(&res) == nil {
					subs[i] = res.StartQuizSession.ID
					return
				}
				mu.Lock()
				failures = append(failures, fmt.Sprint(err, resp))
				mu.Unlock()
			}()
		}
		wg.Wait()
		require.Empty(t, failures)
		return subs
	}

	finishSession := func(t *testing.T, sessionID string) {
		t.Helper()
		admin(t, `mutation($id: ID!) { lockQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
		admin(t, `mutation($id: ID!) { finishQuizSession(id: $id) { id } }`, map[string]any{"id": sessionID}, nil)
	}

	type bet struct {
		submissionID string
		question     question
		amount       int // 0 = answer without a bet
	}
	type outcome struct {
		user     int
		amount   int
		duration time.Duration
		msg      string
	}

	// run fires every user's waves: users arrive spread evenly over the window,
	// each sends its waves one after another, the bets within a wave in
	// parallel. All answers are wrong, so every accepted bet is lost and the
	// final score shows what was staked.
	run := func(t *testing.T, waves func(userIndex int) [][]bet) []outcome {
		t.Helper()
		var mu sync.Mutex
		var outcomes []outcome
		var wg sync.WaitGroup
		gate := make(chan struct{})
		for ui, u := range users {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-gate
				if window > 0 {
					time.Sleep(window * time.Duration(ui) / time.Duration(len(users)))
				}
				for _, wave := range waves(ui) {
					var waveWG sync.WaitGroup
					for _, b := range wave {
						waveWG.Add(1)
						go func() {
							defer waveWG.Done()
							input := map[string]any{"questionId": b.question.id, "selectedAnswerIds": []string{b.question.wrong}}
							if b.amount > 0 {
								input["betAmount"] = b.amount
							}
							inflight <- struct{}{}
							began := time.Now()
							resp, err := client.WithAuth(u.token).Execute(ctx, `mutation($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
								submitQuizAnswer(submissionId: $submissionId, input: $input) { id }
							}`, map[string]any{"submissionId": b.submissionID, "input": input})
							o := outcome{user: ui, amount: b.amount, duration: time.Since(began)}
							<-inflight
							switch {
							case err != nil:
								o.msg = err.Error()
							case resp.HasErrors():
								o.msg = resp.ErrorMessage()
							}
							mu.Lock()
							outcomes = append(outcomes, o)
							mu.Unlock()
						}()
					}
					waveWG.Wait()
				}
			}()
		}
		close(gate)
		wg.Wait()
		return outcomes
	}

	// report logs the numbers and checks no bet failed for anything but the points
	report := func(t *testing.T, outcomes []outcome, wall time.Duration) (acceptedByUser []int) {
		t.Helper()
		acceptedByUser = make([]int, len(users))
		var accepted, rejected int
		otherErrors := map[string]int{}
		durations := make([]time.Duration, 0, len(outcomes))
		for _, o := range outcomes {
			durations = append(durations, o.duration)
			switch {
			case o.msg == "":
				accepted++
				acceptedByUser[o.user]++
			case strings.Contains(o.msg, "exceeds available points"):
				rejected++
			default:
				otherErrors[o.msg]++
			}
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		pct := func(p float64) time.Duration {
			return durations[int(float64(len(durations)-1)*p)].Round(time.Millisecond)
		}
		t.Logf("%d users, %d answers in %s (%.0f/s), arrival window %s, pool %d", len(users), len(outcomes), wall.Round(time.Millisecond),
			float64(len(outcomes))/wall.Seconds(), window, poolSize)
		t.Logf("accepted %d, rejected for points %d, other errors %d", accepted, rejected, len(outcomes)-accepted-rejected)
		t.Logf("latency p50 %s, p95 %s, p99 %s, max %s", pct(0.50), pct(0.95), pct(0.99), pct(1))
		for msg, n := range otherErrors {
			t.Logf("  %d× %s", n, msg)
		}
		assert.Empty(t, otherErrors, "bets may only be rejected for the points")
		return acceptedByUser
	}

	// checkScores finishes the session(s) and checks nobody lost more than they had
	checkScores := func(t *testing.T, wantScore func(userIndex int) int) {
		t.Helper()
		rows, err := dbMgr.DB.Pool.Query(ctx,
			`SELECT u, COALESCE((SELECT SUM(points) FROM score_journal WHERE user_id = u AND project_id = $2), 0)
			 FROM unnest($1::char(28)[]) AS u`, userIDs, projectID)
		require.NoError(t, err)
		scores := map[string]int{}
		for rows.Next() {
			var id string
			var score int
			require.NoError(t, rows.Scan(&id, &score))
			scores[id] = score
		}
		require.NoError(t, rows.Err())
		var negative, wrong int
		for i, u := range users {
			if scores[u.id] < 0 {
				negative++
			}
			if scores[u.id] != wantScore(i) {
				wrong++
			}
		}
		assert.Zero(t, negative, "users with a negative score")
		assert.Zero(t, wrong, "users whose final score doesn't match their accepted bets")
	}

	measure := func(t *testing.T, waves func(int) [][]bet) []int {
		t.Helper()
		began := time.Now()
		outcomes := run(t, waves)
		return report(t, outcomes, time.Since(began))
	}

	// ================================================================ scenarios

	t.Run("baseline: everyone answers every question, no betting", func(t *testing.T) {
		resetScores(t)
		sessionID, questions := newQuizSession(t, "Load baseline", false)
		subs := startAll(t, sessionID)
		answered := measure(t, func(ui int) [][]bet {
			waves := make([][]bet, len(questions))
			for qi, q := range questions {
				waves[qi] = []bet{{subs[ui], q, 0}}
			}
			return waves
		})
		for _, n := range answered {
			if !assert.Equal(t, numQuestions, n, "every answer stored") {
				break
			}
		}
		finishSession(t, sessionID)
	})

	t.Run("camp moment: everyone answers every question, bets that fit", func(t *testing.T) {
		resetScores(t)
		sessionID, questions := newQuizSession(t, "Load camp moment", true)
		subs := startAll(t, sessionID)
		stake := startingScore / numQuestions
		accepted := measure(t, func(ui int) [][]bet {
			waves := make([][]bet, len(questions))
			for qi, q := range questions {
				waves[qi] = []bet{{subs[ui], q, stake}}
			}
			return waves
		})
		for ui, n := range accepted {
			if !assert.Equal(t, numQuestions, n, "every bet fits") {
				t.Logf("user %d", ui)
				break
			}
		}
		finishSession(t, sessionID)
		checkScores(t, func(ui int) int { return startingScore - accepted[ui]*stake })
	})

	t.Run("over-betting: everyone bets 60 on every question, one at a time", func(t *testing.T) {
		resetScores(t)
		sessionID, questions := newQuizSession(t, "Load over-betting", true)
		subs := startAll(t, sessionID)
		accepted := measure(t, func(ui int) [][]bet {
			waves := make([][]bet, len(questions))
			for qi, q := range questions {
				waves[qi] = []bet{{subs[ui], q, 60}}
			}
			return waves
		})
		for _, n := range accepted {
			if !assert.Equal(t, 1, n, "only one 60 bet fits in 100") {
				break
			}
		}
		finishSession(t, sessionID)
		checkScores(t, func(ui int) int { return startingScore - accepted[ui]*60 })
	})

	t.Run("double tap: everyone sends all answers at once, bets of 30", func(t *testing.T) {
		resetScores(t)
		sessionID, questions := newQuizSession(t, "Load double tap", true)
		subs := startAll(t, sessionID)
		accepted := measure(t, func(ui int) [][]bet {
			wave := make([]bet, len(questions))
			for qi, q := range questions {
				wave[qi] = bet{subs[ui], q, 30}
			}
			return [][]bet{wave}
		})
		want := min(numQuestions, startingScore/30)
		for _, n := range accepted {
			if !assert.Equal(t, want, n, "as many 30 bets as fit in 100") {
				break
			}
		}
		finishSession(t, sessionID)
		checkScores(t, func(ui int) int { return startingScore - accepted[ui]*30 })
	})

	t.Run("two devices: everyone bets 100 in two sessions at once", func(t *testing.T) {
		resetScores(t)
		sessionA, questionsA := newQuizSession(t, "Load two sessions A", true)
		sessionB, questionsB := newQuizSession(t, "Load two sessions B", true)
		subsA := startAll(t, sessionA)
		subsB := startAll(t, sessionB)
		accepted := measure(t, func(ui int) [][]bet {
			return [][]bet{{{subsA[ui], questionsA[0], 100}, {subsB[ui], questionsB[0], 100}}}
		})
		for _, n := range accepted {
			if !assert.Equal(t, 1, n, "only one 100 bet fits in 100") {
				break
			}
		}
		finishSession(t, sessionA)
		finishSession(t, sessionB)
		checkScores(t, func(ui int) int { return startingScore - accepted[ui]*100 })
	})
}
