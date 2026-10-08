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

1. `submitQuizAnswer` → `ValidateBet` checks the bet against the question limits
   and the user's **current project score** (sum of `score_journal`).
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
| R1 several questions | 100 pts: 60 + 41 + 1 all accepted; score ends at −2 |
| R1 two open sessions | 100 staked in each session; score ends at −100 |
| R2 zero bet when the user can bet | 0 accepted with 100 pts, stored and never settled (`ValidateBet` doc says 0 is rejected; code only rejects nil) |
| R2 nothing to bet (score −50) | 0 rejected ("0 exceeds current score (−50)"): the user can't answer betting questions at all. Score 0 works |
| R2 minimum above what the user has | 0 rejected ("below minimum (100)"): the user can't answer |
| R3 bet without a selection | `selectedAnswerIds` omitted → `is_correct` NULL → never settled, stake kept (free bet). An empty `[]` is graded wrong and works |
| R3 betting turned off after bets | Settlement filters on `betting_enabled`, so the bets are skipped: a lost bet costs nothing |
| R4 LOCKED session | `submitQuizAnswer` checks only the submission (completed/expired), not the session state; the bet is accepted and paid |
| R5 multiplier changed after bets | 2× → 5× after betting pays 5× |
| R6 answer key changed | No re-grade; the stored `is_correct` is used |
| R7 inconsistent limits | Rejected, but by DB CHECK constraints: the raw `SQLSTATE 23514` error and the constraint name reach the client |
| R8 question points + bet | Finalized user: 1000 + 10 + 100 = 1110. Auto-submitted user: 1100 (auto-submit journals no question points). Settlement also overwrites `points_earned` (10) with the net bet |
| Unit: disabled question, 0 bet | `ValidateBet` rejects it ("betting is not enabled"); reachable only through `updateQuizAnswer` |

Several assertions in a failing R1 subtest cascade from the first failure (e.g. the "40" bet returns the already-stored 41 response).

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

- `validateBetWithMockQueries` in `internal/graph/api/quiz_betting_test.go` is a hand-copied version of `ValidateBet` that has drifted (it rejects 0). Its tests don't test the real function.
- The e2e test "zero bet rejected when betting enabled" in `quiz_betting_test.go` passes only because its question sets `bettingMinAbsolute: 100`.
