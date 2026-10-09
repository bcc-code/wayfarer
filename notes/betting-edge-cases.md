# Betting edge cases: rules and status

The betting edge-case tests assert the **intended** behavior. A failing test is a
bug to fix (or a rule to revisit), not a test to loosen. If a rule below is
wrong, change the rule here and in the test header together.

Tests:
- E2E: `backend/e2e/quiz_betting_edge_cases_test.go` (`TestQuizBettingEdgeCases`, rules R1–R8 in its header)
- Load (opt-in): `backend/e2e/quiz_betting_load_test.go` (`TestQuizBettingLoad`, see below)
- Unit: `internal/graph/api/quiz_betting_validate_test.go` (`ValidateBet`, `positiveBetAllowed`, `betCheck`),
  `internal/graph/api/quiz_question_rules_test.go` (question settings, payout changes, answer comparison),
  `internal/services/betting/service_test.go` (payouts, void bets)
- Admin UI: `frontend/test/unit/quizQuestionValidation.test.ts`
- Existing: `e2e/quiz_session_bet_settlement_test.go`, `e2e/quiz_betting_test.go`

```
go test -v -count=1 -run TestQuizBettingEdgeCases ./e2e/              # needs Docker
go test -v -count=1 -run 'TestValidateBet|TestNewBetCheck|TestPositiveBetAllowed' ./internal/graph/api/
```

## How it works

1. `submitQuizAnswer` / `updateQuizAnswer` store the answer in one transaction, and only while the
   session is OPEN (checked under the submission lock). When the answer carries a bet, the transaction
   first locks all the user's submissions in the project whose session isn't FINISHED
   (`LockUserBettableSubmissions`, in id order, before any other lock). Then `ValidateBet` checks the
   bet against the question limits and the user's **available points**: project score (sum of
   `score_journal`) minus their open stakes (`GetUserAvailableBetPoints`). A stake is open until it is
   paid out; in a FINISHED session it stays open only if the core will still settle it (PREDEFINED,
   betting on), other bets (ORDERING plugin, m2m) are released then. Percentage limits are a share of
   the available points. One user's concurrent bets queue on those locks and each sees the ones
   before it (READ COMMITTED); different users never wait for each other.
2. The bet is stored on `quiz_responses.bet_amount`. Nothing is journaled yet.
3. `finishQuizSession` → auto-submits open submissions → `settleSessionBets` writes a stake (−bet)
   and a payout journal entry per bet and stores the net on `points_earned` + `score_journal_id`.
   Payout = bet × multiplier (floored). A bet without an answer (`is_correct` NULL) is **void**: the
   payout entry returns the stake ("… - stake returned"), net 0, no push notification.
   Only PREDEFINED questions with `betting_enabled`, only `bet_amount > 0`.
4. Question edits (`updateQuizQuestion`) can't turn betting off or change a payout while the question
   has open bets (`CountOpenBetsForQuestion`), and can't change its answers once anyone answered it
   (`QuestionHasResponses`). Answers resent unchanged (the admin editor sends them on every save) are
   left alone, so their IDs, stored selections and translations survive.

## Rules (all implemented, 2026-10-09)

| Rule | Behavior |
|---|---|
| R1 | Open stakes are reserved: a bet may not exceed score − the user's other open bets in the project (across questions and sessions). Betting never makes a score negative |
| R2 | A user can always answer a betting question. `betAmount` must be sent; a bet of 0 is always accepted and means no bet (the player app's slider starts at 0, free-text/number questions have no slider and send 0). Limits only apply to real bets |
| R3 | Every stored bet is resolved at finish. A bet without an answer is void and its stake returned. Betting can't be turned off while the question has open bets (deleting the question cancels them) |
| R4 | Answers and bets only while the session is OPEN |
| R5 | A bet pays out at the odds shown when placed: multipliers can't change while the question has open bets (setting the default explicitly is fine) |
| R6 | A question's answers / ordering items can't change once users answered it; resending them unchanged is fine |
| R7 | Bet limits are consistent (0..100 %, ≥ 0, min ≤ max) and rejected with a validation error, not a raw DB error |
| R8 | A question can't award points and take bets at once (create, or an update that touches points/betting; older questions with both stay editable otherwise; the admin editor checks it too). Question points only go to users who finish the quiz themselves — auto-submitted users get none, but their bets are settled |

Voiding a bet (question deleted, submission reset while OPEN, no answer given) is a refund and is fine.

## Fixed along the way

- **Answer IDs churned on every admin save.** The admin editor sends all answers on every question
  save; the backend deleted and recreated them each time with new IDs, so stored selections pointed to
  deleted answers and answer translations were lost — even when nothing changed. Unchanged answers are
  now left alone (R6).
- **Zero bets**: a minimum (absolute or %) used to reject 0, so users below the minimum (and
  free-text/number questions, which always send 0) couldn't answer at all; with R1 that would have
  hit anyone who bet everything early. 0 is now "no bet" everywhere. The hand-copied
  `validateBetWithMockQueries` that tested a drifted copy is gone (`ValidateBet` is pure now).
- **R1 first attempt** with SERIALIZABLE transactions failed under load and was replaced by per-user
  locks (below). Error message changed from "exceeds current score" to "exceeds available points (N)".

## Open / follow-ups

- The client doesn't know about open stakes: expose the available points (e.g. `availableBetPoints`)
  so the bet maximum in the UI matches the server.
- Refund reason is translated for `en` and `nb`; other languages fall back to the default (`nb`).
- Small window: an admin edit checks open bets, then commits; a bet placed in between isn't seen. Rare
  (admin edits during live betting), not locked on purpose.
- Bench-box load test with bets (k6 scenarios don't bet yet) before a big event, on Neon latencies.

## Load tests of the R1 fix (2026-10-08)

`backend/e2e/quiz_betting_load_test.go` (`TestQuizBettingLoad`, opt-in):

```
BETTING_LOADTEST=1 BETTING_LOAD_USERS=300 go test -v -count=1 -timeout 30m -run TestQuizBettingLoad ./e2e/
```

All users fire at once against a testcontainers Postgres through the real router, DB pool 25
(production default). It asserts no bet fails for anything but the points.

| Scenario (300 users) | Bets | Accepted | Rejected (points) | Gave up ("too many concurrent changes") | Retries/bet | p95 |
|---|---|---|---|---|---|---|
| Camp moment: 5 questions in turn, bets that fit | 1500 | 964 | 0 | **536 (36%)** | 5.6 | 1.04s |
| Over-betting: 60 on each question in turn | 1500 | 300 | 1045 | **155** | 1.8 | 1.04s |
| Double tap: 5 answers at once | 1500 | 800 | 386 | 0 (+ connection resets, harness limit) | 1.5 | 1.25s |

With 30 users nothing gave up, but the camp moment already needed 1.8 retries per bet although no
two users compete for the same points.

First attempt, SERIALIZABLE + retries (table above). **Conclusion: SERIALIZABLE is not viable on this path.** The conflicts are false positives between
different users. Likely cause (not verified in detail): SSI predicate locks are page-level on
indexes (and relation-level on sequential scans). Concurrent submissions/responses have
time-ordered ULID keys, so they land on the same index leaf pages and every user's
available-points read conflicts with every other user's insert. Raising retries only adds
latency. Replaced by per-user locking:

| Scenario (300 users, all at once) | Answers | Accepted | Rejected (points) | Errors | p95 |
|---|---|---|---|---|---|
| Baseline: 5 answers in turn, no betting | 1500 | 1500 | – | 0 | 119ms |
| Camp moment: 5 bets in turn that fit | 1500 | 1500 | 0 | 0 | 162ms |
| Over-betting: 60 on each in turn | 1500 | 300 | 1200 | 0 | 163ms |
| Double tap: 5 bets of 30 at once | 1500 | 900 | 600 | 0 | 232ms |
| Two devices: 100 in two sessions at once | 600 | 300 | 300 | 0 | 233ms |

| Scenario (10,000 users arriving within 10s) | Answers | Accepted | Rejected (points) | Errors | Throughput | p95 |
|---|---|---|---|---|---|---|
| Baseline: 5 answers in turn, no betting | 50,000 | 50,000 | – | 0 | 2,385/s | 252ms |
| Camp moment: 5 bets in turn that fit | 50,000 | 50,000 | 0 | 0 | 1,575/s | 364ms |
| Over-betting: 60 on each in turn | 50,000 | 10,000 | 40,000 | 0 | 1,897/s | 310ms |
| Double tap: 5 bets of 30 at once | 50,000 | 30,000 | 20,000 | 0 | 1,749/s | 326ms |
| Two devices: 100 in two sessions at once | 20,000 | 10,000 | 10,000 | 0 | 1,996/s | 97ms |

Run: `BETTING_LOADTEST=1 BETTING_LOAD_USERS=10000 BETTING_LOAD_WINDOW=10 go test -v -count=1 -timeout 115m -run TestQuizBettingLoad ./e2e/`
(~40 min, mostly seeding; achievement seeding is switched off in the test, it alone took ~1h for 10k users).

Caveats: local Docker Postgres (sub-millisecond statements), test client capped at 500 requests in flight,
so the server ran at saturation (offered load above what 25 connections serve). A bet adds two statements
(lock + available points): ~34% less throughput and +110ms p95 here. On Neon (30–90ms per statement)
those two round trips weigh more; merging them into one statement is the obvious next step if the
bench box shows it matters.

Re-run after R2–R8 (2026-10-09; R4 adds a session-state read to every answer write):

| Scenario (10,000 users arriving within 10s) | Answers | Accepted | Rejected (points) | Errors | Throughput | p95 |
|---|---|---|---|---|---|---|
| Baseline: 5 answers in turn, no betting | 50,000 | 50,000 | – | 0 | 2,154/s | 268ms |
| Camp moment: 5 bets in turn that fit | 50,000 | 50,000 | 0 | 0 | 1,524/s | 372ms |
| Over-betting: 60 on each in turn | 50,000 | 10,000 | 40,000 | 0 | 2,074/s | 276ms |
| Double tap: 5 bets of 30 at once | 50,000 | 30,000 | 20,000 | 0 | 1,698/s | 332ms |
| Two devices: 100 in two sessions at once | 20,000 | 10,000 | 10,000 | 0 | 1,928/s | 281ms |

The camp-moment p99 (625ms, max 961ms) overlapped with another e2e run on the same machine.
