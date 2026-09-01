package main

import (
	"fmt"
	"strings"
	"time"
)

// Cache modes decide how the raw query text is prepared per step.
const (
	// cacheCold appends a per-step nonce comment so the branch's
	// whole-response cache (keyed on sha256 of the raw query text) can never
	// hit. Used for the correctness battery: compares resolver output.
	cacheCold = "cold"
	// cacheWarm sends the identical text twice per side and compares the
	// SECOND responses — asserts the warm path equals the cold path.
	cacheWarm = "warm"
	// cacheVerbatim sends the exact text once. Required for probes where a
	// cache-key collision is the point.
	cacheVerbatim = "verbatim"
)

// Step is one lockstep A→B request.
type Step struct {
	Name  string
	User  string // "u<N>" (seeded user N), "admin", or "" for unauthenticated
	Query string // key into queries; empty for REST steps
	Vars  map[string]any

	// REST step (plugin webhooks, /metrics/http). Used when Query is empty.
	Method string
	Path   string
	Body   map[string]any // JSON body; HMAC-signed when Signed is true
	Signed bool

	CacheMode string // default cacheCold
	// Capture extracts values from each side's response data into per-side
	// variables usable in later Vars/Body as "{{name}}". Paths are dotted,
	// arrays indexed numerically ("orderedQuestions.0.id").
	Capture map[string]string
	// ExpectDiverge marks a pinned probe step: divergence between A and B is
	// the *expected* outcome. An unexpected MATCH is flagged loudly.
	ExpectDiverge string // probe id, e.g. "P1-cross-user-leak"
	// Check names a per-side assertion evaluated on both responses (see
	// checks.go). Its failures are reported per side.
	Check string
	// SkipDiff skips A/B comparison (setup steps whose only purpose is the
	// side effect + capture, e.g. reads that prime a cache).
	SkipDiff bool
	// SleepBefore pauses before firing (e.g. waiting out the 30s TTL).
	SleepBefore time.Duration
}

// Scenario is a named sequence of steps sharing per-side captured variables.
type Scenario struct {
	Name        string
	Description string
	RedactPaths []string
	SortPaths   []string
	Steps       []Step
}

// Fixed IDs from the bench dataset (gendata.sql mkid) and the loadtest quiz
// fixture (setup_fixtures.sh).
const (
	loadtestQuizID  = "QZ01LOADTESTFREETEXT00000000"
	loadtestChalID  = "CL01LOADTESTFREETEXT00000000"
	loadtestSessID  = "QN01LOADTESTFREETEXT00000000"
	loadtestQFree1  = "QQ01LOADTESTFREETEXT00000001"
	loadtestQFree2  = "QQ01LOADTESTFREETEXT00000002"
	loadtestQMulti  = "QQ01LOADTESTFREETEXT00000003"
	loadtestQNumber = "QQ01LOADTESTFREETEXT00000004"
	loadtestAnswerA = "QA01LOADTESTFREETEXT00000001"
	loadtestAnswerB = "QA01LOADTESTFREETEXT00000002"
)

// Seeded mkid() IDs from gendata.sql.
var (
	projectID        = seededID("PR", 1)
	eventID          = seededID("EV", 1)
	loadtestTeamID   = seededID("TM", 10)
	adjustTargetUser = seededID("US", 300)
	filterChurchID   = seededID("CH", 5)
	filterSuperTeam  = seededID("ST", 3)
)

// Read-battery users: a spread of seeded users. Users 1..6300 have one team,
// 1..554 have two teams, per gendata.sql team_members inserts.
var readUsers = []string{"u2", "u3", "u100", "u500", "u700", "u2000", "u5000", "u6400", "u9000", "u10500"}

func standingsVars(entity string, first int) map[string]any {
	return map[string]any{"entityType": entity, "filter": nil, "first": first}
}

func buildScenarios() []Scenario {
	var scenarios []Scenario

	// ---------------- Battery R: reads, cold + warm ----------------
	for _, mode := range []string{cacheCold, cacheWarm} {
		var steps []Step
		add := func(name, user, query string, vars map[string]any) {
			steps = append(steps, Step{Name: name, User: user, Query: query, Vars: vars, CacheMode: mode})
		}
		for _, u := range readUsers[:4] {
			add("GetMe-"+u, u, "GetMe", nil)
			add("CurrentProject-"+u, u, "CurrentProject", nil)
			add("ProfilePage-"+u, u, "ProfilePage", map[string]any{"ageFilter": nil})
			add("ActiveChallengesPage-"+u, u, "ActiveChallengesPage", nil)
			add("CompletedChallengesPage-"+u, u, "CompletedChallengesPage", nil)
			add("StandingsUnitPage-"+u, u, "StandingsUnitPage", nil)
			add("StandingsPage-"+u, u, "StandingsPage", nil)
		}
		for _, u := range readUsers[4:] {
			add("ChallengePage-"+u, u, "ChallengePage", map[string]any{"challengeId": loadtestChalID})
			add("GetQuiz-"+u, u, "GetQuiz", map[string]any{"id": loadtestQuizID})
			add("StandingsLocalPage-"+u, u, "StandingsLocalPage", map[string]any{"filter": nil, "first": 50})
		}
		u := readUsers[0]
		for _, entity := range []string{"PERSONS", "TEAMS", "SUPER_TEAMS", "CHURCHES"} {
			add("StandingsGlobal-"+entity, u, "StandingsGlobalPage", standingsVars(entity, 50))
		}
		// Filter coverage on PERSONS (every LeaderboardFilter field).
		filters := []struct {
			name   string
			filter map[string]any
		}{
			{"minScore", map[string]any{"minScore": 10}},
			{"maxScore", map[string]any{"maxScore": 200}},
			{"churchId", map[string]any{"churchId": filterChurchID}},
			{"country", map[string]any{"country": "NO"}},
			{"churchCategory", map[string]any{"churchCategory": "L"}},
			{"gender", map[string]any{"gender": "FEMALE"}},
			{"ageRange", map[string]any{"ageRange": map[string]any{"min": 15, "max": 25}}},
			{"teamId", map[string]any{"teamId": loadtestTeamID}},
			{"superTeamId", map[string]any{"superTeamId": filterSuperTeam}},
			{"combined", map[string]any{"gender": "MALE", "ageRange": map[string]any{"min": 13, "max": 36}, "minScore": 1}},
		}
		for _, f := range filters {
			add("Filter-"+f.name, readUsers[1], "StandingsGlobalPage",
				map[string]any{"entityType": "PERSONS", "filter": f.filter, "first": 50})
		}
		// ProfilePage with an actual filter.
		add("ProfilePage-ageFilter", readUsers[2], "ProfilePage",
			map[string]any{"ageFilter": map[string]any{"ageRange": map[string]any{"min": 13, "max": 25}}})

		scenarios = append(scenarios, Scenario{
			Name:        "R-reads-" + mode,
			Description: "Read battery (" + mode + " cache path): bootstrap, profile, challenges, quiz, standings, filters",
			RedactPaths: []string{"firebaseToken.token"},
			Steps:       steps,
		})
	}

	// Pagination round-trip: each side follows its OWN endCursor; cursor
	// values themselves are opaque and excluded from the diff.
	scenarios = append(scenarios, Scenario{
		Name:        "R-pagination",
		Description: "Leaderboard cursor round-trip: page 2 fetched with each side's own endCursor",
		RedactPaths: []string{"pageInfo.startCursor", "pageInfo.endCursor"},
		Steps: []Step{
			{Name: "page1", User: "u2", Query: "LeaderboardPageForward",
				Vars:    map[string]any{"entityType": "PERSONS", "first": 20, "after": nil},
				Capture: map[string]string{"cursor": "myCurrentProject.leaderboard.pageInfo.endCursor"}},
			{Name: "page2-own-cursor", User: "u2", Query: "LeaderboardPageForward",
				Vars: map[string]any{"entityType": "PERSONS", "first": 20, "after": "{{cursor}}"}},
			{Name: "teams-page1", User: "u2", Query: "LeaderboardPageForward",
				Vars:    map[string]any{"entityType": "TEAMS", "first": 15, "after": nil},
				Capture: map[string]string{"tcursor": "myCurrentProject.leaderboard.pageInfo.endCursor"}},
			{Name: "teams-page2", User: "u2", Query: "LeaderboardPageForward",
				Vars: map[string]any{"entityType": "TEAMS", "first": 15, "after": "{{tcursor}}"}},
		},
	})

	// ---------------- Battery W: writes ----------------
	// Each scenario uses dedicated virgin users (high indexes, in-project per
	// gendata user_projects 1..10588) so both sides start symmetric.

	scenarios = append(scenarios, Scenario{
		Name:        "W1-enroll",
		Description: "Enroll in quiz challenge, immediately re-read challenge pages and points",
		Steps: []Step{
			{Name: "prime-active", User: "u10001", Query: "ActiveChallengesPage"},
			{Name: "prime-challenge", User: "u10001", Query: "ChallengePage", Vars: map[string]any{"challengeId": loadtestChalID}},
			{Name: "enroll", User: "u10001", Query: "EnrollInChallenge", Vars: map[string]any{"challengeId": loadtestChalID}},
			{Name: "reread-active", User: "u10001", Query: "ActiveChallengesPage"},
			{Name: "reread-challenge", User: "u10001", Query: "ChallengePage", Vars: map[string]any{"challengeId": loadtestChalID}},
			{Name: "reread-profile", User: "u10001", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}},
			{Name: "reread-standings-me", User: "u10001", Query: "StandingsGlobalPage", Vars: standingsVars("PERSONS", 10)},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W2-quiz-lifecycle",
		Description: "Session quiz: start, answer free-text x2 + multi-predefined + number, finalize, re-read points",
		Steps: []Step{
			{Name: "start-session", User: "u10002", Query: "StartQuizSession",
				Vars:    map[string]any{"sessionId": loadtestSessID},
				Capture: map[string]string{"subId": "startQuizSession.id"}},
			{Name: "answer-free1", User: "u10002", Query: "SubmitQuizAnswer",
				Vars: map[string]any{"submissionId": "{{subId}}", "input": map[string]any{
					"questionId": loadtestQFree1, "textResponse": "Patience with my siblings", "timeSpentSeconds": 12}}},
			{Name: "answer-free2", User: "u10002", Query: "SubmitQuizAnswer",
				Vars: map[string]any{"submissionId": "{{subId}}", "input": map[string]any{
					"questionId": loadtestQFree2, "textResponse": "Read one chapter every morning", "timeSpentSeconds": 9}}},
			{Name: "answer-multi", User: "u10002", Query: "SubmitQuizAnswer",
				Vars: map[string]any{"submissionId": "{{subId}}", "input": map[string]any{
					"questionId": loadtestQMulti, "selectedAnswerIds": []any{loadtestAnswerA, loadtestAnswerB}, "timeSpentSeconds": 7}}},
			{Name: "answer-number", User: "u10002", Query: "SubmitQuizAnswer",
				Vars: map[string]any{"submissionId": "{{subId}}", "input": map[string]any{
					"questionId": loadtestQNumber, "numberResponse": 42, "timeSpentSeconds": 4}}},
			{Name: "finalize", User: "u10002", Query: "FinalizeQuiz", Vars: map[string]any{"submissionId": "{{subId}}"}},
			{Name: "reread-profile", User: "u10002", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}},
			{Name: "reread-challenge", User: "u10002", Query: "ChallengePage", Vars: map[string]any{"challengeId": loadtestChalID}},
			{Name: "reread-standings", User: "u10002", Query: "StandingsGlobalPage", Vars: standingsVars("PERSONS", 10)},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W3-ordering-betting",
		Description: "Admin authors ordering+betting quiz via mutations on both sides; user plays it (bet + ordering answer)",
		Steps: []Step{
			{Name: "create-challenge", User: "admin", Query: "CreateChallenge",
				Vars: map[string]any{"projectId": projectID, "eventId": eventID, "input": map[string]any{
					"type": "QUIZ", "name": "Diff Ordering Betting Challenge", "buttonText": "Take Quiz",
					"description": "<p>A/B differential test challenge</p>"}},
				Capture: map[string]string{"chalId": "createChallenge.id"}},
			{Name: "publish", User: "admin", Query: "PublishChallenge",
				Vars: map[string]any{"id": "{{chalId}}", "publishedAt": "2026-08-24T00:00:00Z"}},
			{Name: "visible", User: "admin", Query: "SetChallengeVisibility",
				Vars: map[string]any{"id": "{{chalId}}", "visibleAt": "2026-08-24T00:00:00Z"}},
			{Name: "create-quiz", User: "admin", Query: "CreateQuiz",
				Vars: map[string]any{"input": map[string]any{
					"projectId": projectID, "challengeId": "{{chalId}}", "name": "Diff Ordering Betting Quiz",
					"description": "A/B test quiz", "randomizeQuestions": false, "revealCorrectAnswers": true,
					"allowRetakes": false, "completionPoints": 15}},
				Capture: map[string]string{"quizId": "createQuiz.id"}},
			{Name: "add-ordering-q", User: "admin", Query: "AddQuizQuestion",
				Vars: map[string]any{"quizId": "{{quizId}}", "input": map[string]any{
					"questionType": "ORDERING", "questionText": "Order these events:", "questionOrder": 0, "points": 5,
					"orderingItems": []any{
						map[string]any{"itemText": "Creation", "correctOrder": 1},
						map[string]any{"itemText": "Flood", "correctOrder": 2},
						map[string]any{"itemText": "Exodus", "correctOrder": 3},
					}}},
				Capture: map[string]string{
					"ordQId":   "addQuizQuestion.id",
					"ordItem1": "addQuizQuestion.orderingItems.0.id",
					"ordItem2": "addQuizQuestion.orderingItems.1.id",
					"ordItem3": "addQuizQuestion.orderingItems.2.id"}},
			{Name: "add-betting-q", User: "admin", Query: "AddQuizQuestion",
				Vars: map[string]any{"quizId": "{{quizId}}", "input": map[string]any{
					"questionType": "PREDEFINED", "questionText": "Bet on this one?", "questionOrder": 1, "points": 10,
					"bettingEnabled": true, "bettingMinAbsolute": 5, "bettingMaxAbsolute": 100,
					"predefinedAnswers": []any{
						map[string]any{"answerText": "Yes", "isCorrect": true, "answerOrder": 0},
						map[string]any{"answerText": "No", "isCorrect": false, "answerOrder": 1},
					}}},
				Capture: map[string]string{
					"betQId":       "addQuizQuestion.id",
					"betAnswerYes": "addQuizQuestion.predefinedAnswers.0.id"}},
			{Name: "give-points", User: "admin", Query: "CreateScoreAdjustment",
				Vars: map[string]any{"input": map[string]any{
					"projectId": projectID, "userId": seededUserID(10003), "points": 50, "reason": "Betting stake for A/B test"}}},
			{Name: "create-session", User: "admin", Query: "CreateQuizSession",
				Vars:    map[string]any{"input": map[string]any{"quizId": "{{quizId}}"}},
				Capture: map[string]string{"sessId": "createQuizSession.id"}},
			{Name: "grant-access", User: "admin", Query: "GrantQuizSessionAccess",
				Vars: map[string]any{"input": map[string]any{"sessionId": "{{sessId}}", "userIds": []any{seededUserID(10003)}}}},
			{Name: "open-session", User: "admin", Query: "OpenQuizSession", Vars: map[string]any{"id": "{{sessId}}"}},
			{Name: "start-session", User: "u10003", Query: "StartQuizSession",
				Vars:    map[string]any{"sessionId": "{{sessId}}"},
				Capture: map[string]string{"subId": "startQuizSession.id"}},
			{Name: "answer-ordering", User: "u10003", Query: "SubmitQuizAnswer",
				Vars: map[string]any{"submissionId": "{{subId}}", "input": map[string]any{
					"questionId": "{{ordQId}}", "submittedOrder": []any{"{{ordItem1}}", "{{ordItem2}}", "{{ordItem3}}"},
					"timeSpentSeconds": 8}}},
			{Name: "answer-bet", User: "u10003", Query: "SubmitQuizAnswer",
				Vars: map[string]any{"submissionId": "{{subId}}", "input": map[string]any{
					"questionId": "{{betQId}}", "selectedAnswerIds": []any{"{{betAnswerYes}}"}, "betAmount": 25,
					"timeSpentSeconds": 5}}},
			{Name: "finalize", User: "u10003", Query: "FinalizeQuiz", Vars: map[string]any{"submissionId": "{{subId}}"}},
			{Name: "reread-profile", User: "u10003", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}},
			{Name: "reread-challenge", User: "u10003", Query: "ChallengePage", Vars: map[string]any{"challengeId": "{{chalId}}"}},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W4-score-adjustment",
		Description: "Admin score adjustment, user immediately re-reads points and standings",
		Steps: []Step{
			{Name: "prime-profile", User: "u300", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}},
			{Name: "adjust", User: "admin", Query: "CreateScoreAdjustment",
				Vars: map[string]any{"input": map[string]any{
					"projectId": projectID, "userId": adjustTargetUser, "points": 37, "reason": "A/B differential test"}}},
			{Name: "reread-profile", User: "u300", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}},
			{Name: "reread-standings-me", User: "u300", Query: "StandingsGlobalPage", Vars: standingsVars("PERSONS", 10)},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W5-plugin-team-rename",
		Description: "Ladder to Heaven team-name-changed webhook, then team re-read",
		Steps: []Step{
			{Name: "webhook", Method: "POST", Path: "/plugins/ladder-to-heaven/team-name-changed", Signed: true,
				Body: map[string]any{
					"event_type": "team.name.changed", "timestamp": "2026-08-25T10:00:00Z", "project_id": projectID,
					"data": map[string]any{"team_id": loadtestTeamID, "old_name": "Team 10", "new_name": "The Renamed Ten"}}},
			{Name: "reread-unit", User: "u10", Query: "StandingsUnitPage"},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W6-plugin-content-event",
		Description: "Ladder to Heaven content-event webhook, then points re-read",
		Steps: []Step{
			{Name: "webhook", Method: "POST", Path: "/plugins/ladder-to-heaven/content-event", Signed: true,
				Body: map[string]any{
					"event_type": "content.consumed", "timestamp": "2026-08-25T10:05:00Z", "project_id": projectID,
					"user": map[string]any{"id": seededUserID(10004)},
					"data": map[string]any{"task_id": "task-diff-1", "content_progress": 1.0, "consumed_at": "2026-08-25T10:04:00Z"}}},
			{Name: "reread-profile", User: "u10004", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W7-plugin-superteam-preview",
		Description: "Superteam distribution preview (read-only, deterministic over identical data)",
		Steps: []Step{
			{Name: "preview", User: "admin", Method: "GET", Path: "/plugins/ladder-to-heaven/preview-superteams?project_id=PR00000000000000000000000001"},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "W8-plugin-quiz-finalized",
		Description: "Ladder to Heaven quiz-finalized webhook (betting settlement path)",
		Steps: []Step{
			{Name: "webhook", Method: "POST", Path: "/plugins/ladder-to-heaven/quiz-finalized", Signed: true,
				Body: map[string]any{
					"event_type": "quiz.finalized", "timestamp": "2026-08-25T10:10:00Z", "project_id": projectID,
					"data": map[string]any{
						"session_id": loadtestSessID, "session_name": "Load Test Free-Text Session",
						"quiz_id": loadtestQuizID, "quiz_name": "Load Test Free-Text Quiz",
						"challenge_id": loadtestChalID, "finished_at": "2026-08-25T10:09:00Z",
						"results": []any{map[string]any{
							"user_id": seededUserID(10002), "completed": true,
							"score": 3, "max_score": 4, "score_percentage": 75.0, "auto_submitted": false}}}}},
		},
	})

	// ---------------- Battery P: pinned probes ----------------
	// Run each probe against freshly restarted services (run.sh handles
	// restarts) so the branch cache is provably empty at step 1.

	scenarios = append(scenarios, Scenario{
		Name:        "P1-cross-user-leak",
		Description: "Byte-identical single-top-level query from two users within the TTL: does user 2 get user 1's `me`?",
		Steps: []Step{
			{Name: "warm-as-u2", User: "u2", Query: "LeakProbe", CacheMode: cacheVerbatim, Check: "meIsCaller", SkipDiff: true},
			{Name: "read-as-u3", User: "u3", Query: "LeakProbe", CacheMode: cacheVerbatim, Check: "meIsCaller",
				ExpectDiverge: "P1"},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "P2-variable-collision",
		Description: "Same query text, different variables within the TTL: is `first` honored on the second call?",
		Steps: []Step{
			{Name: "first-5", User: "u2", Query: "VarCollisionProbe", Vars: map[string]any{"first": 5},
				CacheMode: cacheVerbatim, Check: "edgeCountIsFirst"},
			{Name: "first-50", User: "u2", Query: "VarCollisionProbe", Vars: map[string]any{"first": 50},
				CacheMode: cacheVerbatim, Check: "edgeCountIsFirst", ExpectDiverge: "P2"},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "P3-stale-after-enroll",
		Description: "Prime per-user caches, enroll, immediately re-read the exact same documents: any stale fields on B?",
		Steps: []Step{
			// Verbatim (not nonce'd): the internal Ristretto caches keyed by
			// user/project are under test here; the response cache does not
			// apply to these multi-top-level documents.
			{Name: "prime-challenge", User: "u10006", Query: "ChallengePage", Vars: map[string]any{"challengeId": loadtestChalID}, CacheMode: cacheVerbatim},
			{Name: "prime-active", User: "u10006", Query: "ActiveChallengesPage", CacheMode: cacheVerbatim},
			{Name: "prime-profile", User: "u10006", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}, CacheMode: cacheVerbatim},
			{Name: "enroll", User: "u10006", Query: "EnrollInChallenge", Vars: map[string]any{"challengeId": loadtestChalID}},
			{Name: "reread-challenge", User: "u10006", Query: "ChallengePage", Vars: map[string]any{"challengeId": loadtestChalID}, CacheMode: cacheVerbatim, ExpectDiverge: "P3"},
			{Name: "reread-active", User: "u10006", Query: "ActiveChallengesPage", CacheMode: cacheVerbatim, ExpectDiverge: "P3"},
			{Name: "reread-profile", User: "u10006", Query: "ProfilePage", Vars: map[string]any{"ageFilter": nil}, CacheMode: cacheVerbatim, ExpectDiverge: "P3"},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "P4-ttl-staleness",
		Description: "Admin edits the project info message; is a warmed single-top-level read stale within 30s and fresh after?",
		Steps: []Step{
			{Name: "warm", User: "u2", Query: "ProjectInfoProbe", CacheMode: cacheVerbatim},
			{Name: "update-info", User: "admin", Query: "UpdateProjectInfoMessage",
				Vars: map[string]any{"id": projectID, "input": map[string]any{"infoMessage": "Updated by A/B probe P4 (fix verification pass)"}}},
			{Name: "reread-within-ttl", User: "u2", Query: "ProjectInfoProbe", CacheMode: cacheVerbatim, ExpectDiverge: "P4"},
			{Name: "reread-after-ttl", User: "u2", Query: "ProjectInfoProbe", CacheMode: cacheVerbatim, SleepBefore: 31 * time.Second},
		},
	})

	scenarios = append(scenarios, Scenario{
		Name:        "P5-metrics-endpoint",
		Description: "Unauthenticated GET /metrics/http: new endpoint on B only (status compared, body skipped)",
		Steps: []Step{
			{Name: "metrics", Method: "GET", Path: "/metrics/http", ExpectDiverge: "P5", Check: "statusOnly"},
		},
	})

	return scenarios
}

// applyVars resolves "{{name}}" templates against a side's captured variables.
func applyVars(v any, captured map[string]string) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			r, err := applyVars(e, captured)
			if err != nil {
				return nil, err
			}
			out[k] = r
		}
		return out, nil
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			r, err := applyVars(e, captured)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil
	case string:
		if strings.HasPrefix(t, "{{") && strings.HasSuffix(t, "}}") {
			name := strings.TrimSpace(t[2 : len(t)-2])
			val, ok := captured[name]
			if !ok {
				return nil, fmt.Errorf("unresolved capture variable %q", name)
			}
			return val, nil
		}
		return t, nil
	default:
		return v, nil
	}
}

func splitPath(path string) []string {
	return strings.Split(path, ".")
}

// capturePath walks a dotted path (numeric segments index arrays) into a
// decoded response and returns the value as a string.
func capturePath(data any, path string) (string, error) {
	cur := data
	for _, seg := range splitPath(path) {
		switch t := cur.(type) {
		case map[string]any:
			v, ok := t[seg]
			if !ok {
				return "", fmt.Errorf("capture path %q: key %q not found", path, seg)
			}
			cur = v
		case []any:
			var idx int
			if _, err := fmt.Sscanf(seg, "%d", &idx); err != nil || idx < 0 || idx >= len(t) {
				return "", fmt.Errorf("capture path %q: bad array index %q", path, seg)
			}
			cur = t[idx]
		default:
			return "", fmt.Errorf("capture path %q: cannot descend into %T at %q", path, cur, seg)
		}
	}
	s, ok := cur.(string)
	if !ok {
		return "", fmt.Errorf("capture path %q: value is %T, want string", path, cur)
	}
	return s, nil
}
