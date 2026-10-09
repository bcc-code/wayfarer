# Betting edge cases: intended rules and status

The betting edge-case tests assert the **intended** behavior. A failing test is a
bug to fix (or a rule to revisit), not a test to loosen. If a rule below is
wrong, change the rule here and in the test header together.

Tests:
- E2E: `backend/e2e/quiz_betting_edge_cases_test.go` (`TestQuizBettingEdgeCases`)
- Unit, real `ValidateBet`: `backend/internal/graph/api/quiz_betting_validate_test.go` (`TestValidateBet_Real`)
- Existing: `e2e/quiz_session_bet_settlement_test.go`, `e2e/quiz_betting_test.go`, `internal/services/betting/service_test.go`

```
go test -v -count=1 -run TestQuizBettingEdgeCases ./e2e/              # needs Docker
go test -v -count=1 -run TestValidateBet_Real ./internal/graph/api/
```

## How it works today (short)

1. `submitQuizAnswer` / `updateQuizAnswer` store the answer in one transaction. When it carries a
   bet, it first locks all the user's submissions in the project whose session isn't FINISHED
   (`LockUserBettableSubmissions`, in id order, before any other lock). Then
   `ValidateBet` checks the bet against the question limits and the user's **available points**:
   project score (sum of `score_journal`) minus their open stakes (`GetUserAvailableBetPoints`).
   A stake is open until it is paid out; in a FINISHED session it stays open only if the core
   will still settle it, other bets (ORDERING plugin, m2m) are released then. Percentage limits
   are a share of the available points. One user's concurrent bets queue on those locks and each
   sees the ones before it (READ COMMITTED); different users never wait for each other.
2. The bet is stored on `quiz_responses.bet_amount`. Nothing is journaled yet.
3. `finishQuizSession` → auto-submits open submissions → `settleSessionBets`
   writes a stake (−bet) and winnings (bet × multiplier, floored) journal entry
   per bet and stores the net on `points_earned` + `score_journal_id`.
   It settles only PREDEFINED questions, only graded responses (`is_correct` not NULL), only `bet_amount > 0`,
   and only while the question still has `betting_enabled`. The multiplier is read at finish time.

## Rules

| Rule | Intended behavior |
|---|---|
| R1 | Open stakes are reserved: a bet may not exceed score − the user's other unsettled bets in the project (across questions and sessions). Betting never makes a score negative |
| R2 | A user can always answer a betting question. A bet is required and must be ≥ 1, unless no positive bet is allowed (nothing available, minimum above what they have, max rounds to 0); then 0 is accepted |
| R3 | No free bets: a stored bet is settled at finish, or rejected when placed |
| R4 | Answers and bets only while the session is OPEN |
| R5 | A bet pays out at the odds shown when it was placed (a later multiplier change is rejected or doesn't apply) |
| R6 | Changing the answer key after bets is rejected or re-grades them |
| R7 | Bet limits are consistent (0..100 %, min ≤ max) and rejected with a validation error, not a raw DB error |
| R8 | Betting doesn't cost question points: finalized or auto-submitted, the user ends with question points + net bet |

Voiding a bet (question deleted, submission reset while OPEN) is a refund and is fine.

## Status (2026-10-08)

### Failing — bugs

| Test | What happens today |
|---|---|
| R2 zero bet when the user can bet | 0 accepted with 100 pts, stored and never settled (`ValidateBet` doc says 0 is rejected; code only rejects nil) |
| R2 nothing to bet (score −50) | 0 rejected ("0 exceeds available points (−50)"): the user can't answer betting questions at all. Score 0 works |
| R2 minimum above what the user has | 0 rejected ("below minimum (100)"): the user can't answer |
| R3 bet without a selection | `selectedAnswerIds` omitted → `is_correct` NULL → never settled, stake kept (free bet). An empty `[]` is graded wrong and works |
| R3 betting turned off after bets | Settlement filters on `betting_enabled`, so the bets are skipped: a lost bet costs nothing |
| R4 LOCKED session | `submitQuizAnswer` checks only the submission (completed/expired), not the session state; the bet is accepted and paid |
| R5 multiplier changed after bets | 2× → 5× after betting pays 5× |
| R6 answer key changed | No re-grade; the stored `is_correct` is used |
| R7 inconsistent limits | Rejected, but by DB CHECK constraints: the raw `SQLSTATE 23514` error and the constraint name reach the client |
| R8 question points + bet | Finalized user: 1000 + 10 + 100 = 1110. Auto-submitted user: 1100 (auto-submit journals no question points). Settlement also overwrites `points_earned` (10) with the net bet |
| Unit: disabled question, 0 bet | `ValidateBet` rejects it ("betting is not enabled"); reachable only through `updateQuizAnswer` |

### Fixed

- **R1** (2026-10-08): open stakes are reserved across questions and sessions; each bet write
  locks the user's bettable submissions first (per-user serialization). Covered by sequential and
  concurrent tests plus the load test. Without the lock, the load test's "two devices" scenario
  fails (19 of 300 users got both 100 bets accepted, negative scores); the 2-request e2e test alone
  is too timing-dependent to catch it reliably. Same-session bets were already serialized by the
  submission row lock. Error message changed from "exceeds current score" to "exceeds available points (N)".
  A first attempt with SERIALIZABLE transactions was dropped after the load test (below).

### Passing

Settlement per question with mixed results and custom multipliers; unanswered questions cost nothing; users who never start are unaffected;
two sessions settle independently; finished-session winnings/losses change the next limit; reopen → finish settles once;
bet required; negative bet rejected; bet on a non-betting question rejected; duplicate answer keeps the first bet;
LOCKED/FINISHED sessions can't be started; no answers after finish/finalize/expiry while earlier bets are settled;
reset while OPEN voids bets, reset while LOCKED rejected; FREE_TEXT/NUMBER/ORDERING bets are left to m2m/plugin;
`updateQuizAnswer` on ORDERING (own stake replaced, rejected while LOCKED); multiplier edge values (0×, 1×, 100×, 1.5× rounding, wrong 1.25×);
an unpayable bet doesn't block others and is retried; multi-select all-or-nothing; deleting a question voids its bets;
several users settled independently. Unit: absolute/percentage limits and floor rounding.

## Other notes

- `ValidateBet` is now pure (`config, available, bet`); the old tests in `quiz_betting_test.go` call it directly (the drifted copy `validateBetWithMockQueries` is gone).
- The frontend should show the available points as the bet maximum; it isn't exposed in GraphQL yet.
- The e2e test "zero bet rejected when betting enabled" in `quiz_betting_test.go` passes only because its question sets `bettingMinAbsolute: 100`.

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
