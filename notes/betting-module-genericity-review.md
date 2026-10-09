# Betting Module — Genericity & Reusability Review

**Date:** 2026-09-30
**Question asked:** betting should attach to _all_ quiz question types, but so far only ORDERING has used it. How generic is it really — in functionality and in language?

**Verdict:** the **storage and configuration layers are fully generic**; the **settlement (scoring) layer and the whole user-facing flow are hard-wired to ORDERING**. A predefined/number/free-text question with `bettingEnabled = true` can be authored today, but the user can never place a bet on it, and if they somehow did, nothing would ever pay it out.

---

## Layer-by-layer

| Layer                                                           | Generic?                                            | Notes                                                                                                                       |
| --------------------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| DB schema (`00090_add_question_betting.sql`)                    | ✅ Yes                                              | `betting_*` on `quiz_questions`, `bet_amount` on `quiz_responses`. No question-type coupling, constraints are type-neutral. |
| GraphQL schema (`gql/quizzes.graphqls`)                         | ✅ Yes                                              | All 5 betting fields repeated on every question type + input; `betAmount` on every response type + input.                   |
| Bet validation (`backend/internal/graph/api/quiz_betting.go`)   | ✅ Yes                                              | `ValidateBet` is purely numeric — available points (score − open stakes), %-limits, absolute limits. Zero question-type awareness. |
| Admin editor (`AdminQuizQuestionEditor.vue`)                    | ✅ Yes                                              | Betting block renders for _every_ `questionType`, not gated.                                                                |
| Submit resolver (`createQuizResponse`)                          | ✅ Yes                                              | Validates the bet for all types, inside the transaction that stores the answer (after the user's bet locks).               |
| **Update resolver (`updateQuizResponse`)**                      | ❌ **ORDERING only**                                |                                                                                                                             |
| **Settlement / payout**                                         | ❌ **ORDERING only, and lives in a project plugin** |                                                                                                                             |
| **User-facing flow (`QuizChallenge.vue`, question components)** | ❌ **effectively ORDERING only**                    |                                                                                                                             |
| **Language (i18n + prop naming)**                               | ⚠️ **ORDERING-shaped**                              |                                                                                                                             |

---

## Blocking findings (functionality)

### 1. Settlement lives inside a per-project plugin and is hardcoded to ORDERING

`backend/internal/plugins/ladder_to_heaven/quiz_finalized_handler.go` is the _only_ thing that
actually pays out a bet in production. It is not reusable:

- `processResponse` (~line 390) short-circuits:
  `if response.QuestionType != questionTypeOrdering || !response.BettingEnabled { return }`
- `calculateBetMultiplier(correctCount, totalItems)` is a fixed ladder —
  all correct `2.0x`, 2 correct `1.5x`, 1 correct `1.25x`, everything else `0x` — with a
  literal `expectedOrderingItems = 4` constant baked in for the "all wrong" penalty case.
- Correctness is measured as `countCorrectPositions` over predefined-answer IDs — a concept
  that only exists for ORDERING.
- The multiplier table is **not configurable from the question**; it is a Go constant block.

**Consequence:** enabling betting on a PREDEFINED question today silently does nothing at
finalization — the stake is never deducted and nothing is won. The response keeps
`points_earned = NULL` forever.

**To make generic:** the payout needs (a) a per-question-type "how correct was this?" function
returning a normalized ratio (`correctCount / totalCount`, or simply correct/incorrect for
binary types), and (b) a payout curve configured on the question rather than in constants.
It also needs to move out of `ladder_to_heaven` into a shared service (e.g.
`internal/services/betting`), with the plugin only choosing to _call_ it.

### 2. Two divergent, unreconciled settlement paths

- `recordBetResult` / `recordBetResults` mutations (`quizzes.resolvers.go:1745`, `:1826`) are
  fully type-agnostic — they take `pointsEarned` and write **one** `BET` journal entry with
  `ChallengeID` set.
- The plugin writes **two** entries (negative stake + positive winnings), sets `EventID` but
  **not** `ChallengeID`, and picks which journal ID to store back on the response.

So the same `source_type = 'BET'` means different things depending on which path ran, and the
score journal / `AdminUserScoreJournal` UI has to cope with both shapes. The mutations also skip
the plugin's guards: no idempotency check against an existing `score_journal_id`, no check that
`betting_enabled` was ever true. Whichever shape is chosen, one of the two should go.

### 3. `updateQuizResponse` only updates ORDERING answers

`quizzes.resolvers.go:1101` — `// Only handle ORDERING questions for now`. The bet amount
itself _is_ updated generically (line 1077), but the _answer_ is not, so the
"save bet → change bet → re-save" workflow is unusable for any other type.

### 4. The frontend never shows the betting module outside a session

`resolveFooterState` (`useQuizViewState.ts:103`) only makes `bettingModule` visible in the
`session-betting`, `session-locked` and `session-results` modes. `mode: 'normal'` returns
`{ visible: false }` unconditionally.

`QuizNumberQuestion.vue` and `QuizFreeTextQuestion.vue` hardcode `mode: props.readonly ? 'review' : 'normal'`
→ **their betting module can never render**. `QuizPredefinedQuestion.vue` does compute
`session-betting`, but pins `canChangeBet: false`, so bets are one-shot there.
`QuizJsonQuestion.vue` isn't wired for betting at all — `QuizChallenge.vue:683` omits the
`:bet-amount` prop that the other four get.

### 5. This combination is an outright submission failure, not just a missing feature

> **Resolved 2026-10-09:** a bet of 0 now always means "no bet" and limits only apply to real
> bets, so these answers go through (see `betting-edge-cases.md`, R2).

Free-text / number questions still receive `:bet-amount="isBettingEnabled ? currentBetAmount : undefined"`
(`QuizChallenge.vue:680`, `:710`) and forward it, but `currentBetAmount` can only ever be `0`
because no slider is rendered. If such a question has `bettingMinAbsolute` or
`bettingMinPercentage` set, `ValidateBet` rejects `0` and **the answer cannot be submitted at all**.
An admin can configure this from the editor today with no warning.

### 6. `ValidateBet`'s doc comment contradicts its code

> **Resolved 2026-10-09:** doc and code agree (0 = no bet, also on a question without betting).

```
// - If betting is disabled and bet is nil or 0, it's valid (no bet placed)
```

…but the code rejects _any_ non-nil bet when betting is disabled, including `0`:

```go
if betAmount == nil { return nil }
bet := *betAmount
if bet < 0 { ... }
if !config.BettingEnabled { return &BetValidationError{...} }  // fires for bet == 0
```

The inline comment `// At this point betting must be enabled (we checked above for nil/zero bets)`
is also false — zero is never checked. Today the UI hides this by passing `undefined`, but it is
a trap for the next question type wired up. (See `[[verify-before-assuming]]` — the comment is a
claim, the code is the evidence.)

### 7. Dead code: `ExtractBetConfigFromQuestion`

`quiz_betting.go:156` has zero callers. Both call sites in `quizzes.resolvers.go` (lines 850 and 1086) build `BetValidationConfig` inline from a raw sqlc row instead. It should either be deleted
or become the single construction path (it takes the `quizQuestionRow` interface, which the
`GetQuizQuestionByIDRow` used at both sites does not satisfy directly — that's why it went unused).

---

## Language findings

### 8. Result copy assumes "N of M positions correct"

`quiz.betting.correctCount` = _"You had {count} correct answers."_, plus `correctCountAll` /
`correctCountNone`. For a single-answer PREDEFINED or NUMBER question this reads wrong —
there is one answer, not a count. The module should either take a pre-rendered label from the
parent, or branch on `totalCount === 1`.

### 9. `QuizBettingModule` props are ordering-shaped

`correctCount` / `totalCount` are the module's only correctness inputs, and
`QuizChallenge.vue:338` (`bettingCorrectCount`) returns `null` unless both the question and the
response are `Ordering*`. A generic module wants something like
`{ outcomeRatio: number }` or an opaque `outcomeLabel: string`, not position counts.

### 10. The module re-derives the payout in the view

`resultAmount = pointsEarned + betAmount`, `multiplier = resultAmount / betAmount`. This
reconstructs the backend's multiplier from net points — a second, independent implementation of
the payout formula. If the backend curve ever changes (or a payout produces a non-terminating
ratio), the displayed `x 1.25` and the real ledger drift apart. The multiplier should come from
the API alongside `pointsEarned`.

Also `isWin = pointsEarned >= 0` treats break-even as a win and renders `+0`. No current
multiplier yields exactly `1.0x`, but a configurable curve would.

### 11. i18n leftovers in `frontend/i18n/locales/en_us.json`

- `quiz.betting.yourPoints` = **`"Dine poeng"`** — untranslated Norwegian sitting in the English locale.
- `quiz.betting.sessionLocked`, `results`, `yourPoints`, `winnings` — **all four are unused**
  (no reference anywhere in `app/` or `layers/`). Present in every locale file, so ~15 files carry dead keys.

### 12. Admin-side strings are hardcoded Norwegian

`AdminQuizQuestionEditor.vue`: `label="Aktiver betting"`, `"Min %"`, `"Maks %"`, `"Min poeng"`,
`"Maks poeng"`. Consistent with the rest of admin, so not a betting-specific defect — noting it
only because these are the strings a non-Norwegian project would hit first.

### 13. "Betting" as domain vocabulary

Every layer names this _betting_ / _bet_ / _stake_ / _winnings_ — including the user-visible
`bet_reason` journal strings (`"{challenge} - stake"`, `"{challenge} - winnings"`) that land in a
user's score history. For a youth-camp bible-study product, a neutral framing
(_wager_ → _stake points_, _winnings_ → _bonus_) may be worth a deliberate decision before this
spreads to more question types. Flagging as a product call, not a defect.

### 14. Naming inconsistency: `betting*` vs `bet*`

Question config uses `betting*` (`bettingEnabled`, `bettingMinAbsolute`); response data uses
`bet*` (`betAmount`). Internally consistent per side, but `BetValidationConfig` mixes them
(`BettingEnabled` + field name `betAmount` in errors). Minor.

---

## Other observations

- **`schema.sql` is stale for betting** — it has `'BET'` in the `score_journal.source_type` CHECK
  (line 416) but no `betting_*` columns on `quiz_questions` and no `bet_amount` on
  `quiz_responses`. Already flagged more broadly in `notes/quiz-results-admin.md`.
- **Admin preview can't show betting** — `AdminQuizQuestionPreview.vue:51` hardcodes
  `bettingEnabled: false`, so the editor's betting block has no preview. Already listed as an
  opportunity in `notes/admin-preview-opportunities.md`.
- **The special case is explicit in the code** — `QuizChallenge.vue:314`
  `// Special case for PC26 Game Night betting` gates `isSingleOrderingQuestion`, which
  `resolveQuizViewState` uses to decide whether a FINISHED session shows inline results or the
  score screen. Any second betting quiz shape will need this generalized.
- **Admin results are out of scope for betting** — `notes/quiz-results-admin.md` says betting
  "can be added per question later"; still true.

---

## Progress

### 2026-10-06 — shared settlement service created (`backend/internal/services/betting/`)

First version used per-question-type `Evaluator`/`Strategy` interfaces for partial payouts.
Replaced on 2026-10-07 by the all-or-nothing version below.

Mock: `services/betting/mocks/Querier.go`, listed in `.mockery.yml`. Note `make generate` does
**not** run mockery; it was generated with `go run github.com/vektra/mockery/v3@latest`
(v3.8.0), and the other mocks' template drift from that version was reverted.

### 2026-10-06 — one shared grader for `is_correct` (`backend/internal/services/quizgrading/`)

`quizgrading.Grade(questionType, Response, []Answer) *bool` is now the single place that decides
correctness. It is **all or nothing**: PREDEFINED (single and multiple choice) is correct only if
exactly the correct answers are selected (compared as sets), ORDERING only if every item is in
its position. FREE_TEXT / NUMBER / JSON return `nil` (not graded).

It replaced the copies in `SubmitQuizAnswer`, `UpdateQuizAnswer` and `CreateQuizSubmission`
(via `gradeQuizResponse` in `graph/api/quiz_grading.go`). Behaviour changes:

- `SubmitQuizAnswer` no longer counts selected `[A, A]` as correct when the correct set is `{A, B}`
  (the old check compared lengths only), and now counts `[A, A]` as correct when the correct set
  is `{A}` (previously wrong on length). Both now agree with `CreateQuizSubmission`.
- `SubmitQuizAnswer` returns an error if the answers can't be loaded, instead of silently
  grading the response as wrong.
- `UpdateQuizAnswer` loads answers via the cached `QuizAnswersByQuestionLoader` instead of a
  direct query.

The plugin's partial ORDERING payout (`countCorrectPositions`) is untouched and does not use the
grader.

### 2026-10-07 — PREDEFINED bets are paid out by the core when a session finishes

Scope decision: only PREDEFINED (single and multiple choice) bets, all or nothing. ORDERING
stays with the `ladder_to_heaven` plugin unchanged. Backend only; no frontend changes.

- **Migration `00108_add_question_betting_multipliers.sql`** — `quiz_questions.betting_multiplier_correct`
  and `betting_multiplier_wrong` (`NUMERIC(5,2)`, nullable, 0–100, `wrong <= correct`). NULL = default
  (correct 2.0, wrong 0). Exposed as `bettingMultiplierCorrect` / `bettingMultiplierWrong` on the
  `QuizQuestion` interface, all question types and both question inputs; validated by
  `ValidateBettingMultipliers` (`graph/api/quiz_betting.go`).
- **`services/betting`** — multipliers are fixed-point hundredths (2.5x = 250); `Winnings` computes
  `floor(stake * multiplier / 100)` with integers and returns `ErrPayoutOutOfRange` if it does not
  fit in int32 (the bet is then logged and left unsettled). `Multiplier(correct, multCorrect, multWrong)` and
  `Settle(ctx, Querier, SettleInput)`. Correctness is the response's stored `is_correct` (set by
  `quizgrading` on submit), so points and payout agree even if answers were edited later (editing
  answers replaces their IDs). Same journal shape as the plugin: stake (negative) + winnings
  (≥ 0), `ChallengeID` unset; `points_earned` is overwritten with the net bet result.
- **`FinishQuizSession`** calls `settleSessionBets` (`graph/api/quiz_session_betting.go`) after
  auto-submit and before the Firestore notification. Query `GetUnsettledSessionBets` selects
  responses with `bet_amount > 0`, `betting_enabled`, `is_correct IS NOT NULL` (a bet sent
  without an answer is ignored, not lost), `score_journal_id IS NULL` and
  `question_type = ANY(['PREDEFINED'])`. Each bet is settled in its own transaction (no partial
  payouts); a failing bet is logged and skipped. The result is stored with `SettleBetResult`
  (`WHERE score_journal_id IS NULL`, after the journal entries because of the FK); 0 rows returns
  `betting.ErrAlreadySettled` and the transaction is rolled back, so two concurrent
  `finishQuizSession` calls (`UpdateQuizSessionState` is unconditional) still pay out once. Afterwards: cache invalidation, Firestore
  `NotifyUserContent`, push `SendTranslatedBetResultNotificationCtx` (only when the quiz has a
  challenge), sent from one background goroutine with at most 16 users at a time
  (`betNotifyConcurrency`) and the request context via `context.WithoutCancel`. Settlement itself runs synchronously, so the admin's finish call
  waits for it.
- **Retry:** auto-submit + settlement live in `closeSessionAndSettleBets`. Settlement only runs
  after auto-submit succeeded, and calling `finishQuizSession` on a `FINISHED` session runs both
  again (no-ops once done), so a failed auto-submit or bet is retried by finishing again. Partial
  failures are logged and recorded on the span (`bets.failed`). A bet whose multiplier is invalid
  keeps failing until the multiplier is fixed.
- **Late answers:** `submitQuizAnswer` / `updateQuizAnswer` write under the submission row lock
  (`GetQuizSubmissionByIDForUpdate`, `graph/api/quiz_answers.go`) and re-check completion /
  session state under it. Before auto-submit, finishing runs `WaitForSessionSubmissionLocks`
  (`SELECT ... FOR SHARE` on all session submissions, completed ones included, outside a
  transaction) as a barrier, so an in-flight answer is committed before settlement and the
  `quiz_session_finished` webhook load it, or sees the session FINISHED and is rejected. The
  barrier is needed because auto-submit's `UPDATE` skips completed submissions, whose answers
  `updateQuizAnswer` can still change while the session is OPEN. A duplicate answer is detected
  under the lock and the stored response returned.
- **Tests:** `services/betting/service_test.go`, `TestValidateBettingMultipliers`,
  `e2e/quiz_session_bet_settlement_test.go` (custom/default multipliers, ORDERING untouched,
  finishing twice does not pay twice, finishing again retries a failed auto-submit and unpaid bet,
  an answer holding the submission lock while finishing is still settled, for open and completed
  submissions).

Known gaps:

- `points_earned` on a settled bet response holds the net bet result, not the question points
  (same as the plugin; the frontend reads it that way).
- `clearBettingMinAbsolute` / `clearBettingMaxAbsolute` exist in the schema and are sent by the
  admin UI but are ignored by the backend; multipliers likewise cannot be reset to NULL via
  `updateQuizQuestion` (the query uses COALESCE).
- Frontend not updated: `pnpm codegen`, admin editor fields, `QuizBettingModule` result copy.

## Suggested order of work, if this is to be made properly generic

1. Extract settlement out of `ladder_to_heaven` into a shared service; define a per-question-type
   `evaluate(response, question) → (correct, total)` and a question-configurable payout curve.
   Reconcile with `recordBetResult` so there is one journal shape.
2. ~~Fix `ValidateBet`'s zero-bet handling to match its documented contract~~ (done 2026-10-09);
   delete or wire up `ExtractBetConfigFromQuestion`.
3. Give every question component a real `session-betting` action mode (or hoist betting out of
   the per-type action state entirely — it is question-type-independent by nature).
4. Generalize `updateQuizResponse` beyond ORDERING.
5. Replace `correctCount`/`totalCount` on `QuizBettingModule` with a type-neutral outcome input;
   take the multiplier from the API rather than re-deriving it.
6. Clean the i18n: remove the 4 unused keys, fix `yourPoints`, reword `correctCount` for
   single-answer questions.
