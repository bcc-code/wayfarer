# Betting Frontend — Genericity & Reusability

**Date:** 2026-10-22
**Question asked:** the betting module was built for one specific quiz question (PC26 Game Night,
a single ORDERING question). The backend has since been generalized. How generic and reusable is
the **frontend** today?

**Verdict:** the frontend has **not followed the backend's generalization**. The betting
_component_ (`QuizBettingModule.vue`) is close to generic in its betting mode, but everything that
feeds it is still ORDERING-shaped. Concretely:

- **ORDERING** — fully working, the only complete path (place → change → locked → inline results).
- **PREDEFINED** — can place a bet, but cannot change it, and **never sees the result**: after the
  session finishes the user is routed to the score screen, which says nothing about the bet. The
  backend settles these bets (since 2026-10-07); the UI just doesn't show it.
- **NUMBER / FREE_TEXT** — the slider can never render (hardcoded action mode), so betting is
  silently a no-op even with `bettingEnabled` on. The admin editor offers it anyway.
- **JSON** — the question component is an 11-line stub (`<div />`); betting is moot there.
- **Payout curve is unreachable** — `bettingMultiplierCorrect` / `bettingMultiplierWrong` exist in
  the schema and in `generated.ts`, but are selected by **no query document** and used **nowhere**
  in `app/` or `layers/`. The admin cannot set them, so in practice every question runs on the
  backend defaults (2.0× / 0×) and the UI re-derives the multiplier from the net points.

  > **Partly resolved 2026-10-27:** the result screen now shows the real payout, via
  > `QuizResponse.settledBetMultiplier`. The admin still cannot *set* a curve (step 5), and the odds are
  > still not shown before betting.

See `[[betting-module-genericity-review]]` for the original full-stack review (2026-09-30) and
`[[betting-edge-cases]]` for the rules the backend now enforces.

---

## Layer-by-layer (frontend only)

| Layer                                                 | Generic?                         | Notes                                                                                                                                                                                    |
| ----------------------------------------------------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GraphQL fragments (`app/graphql/fragments/quiz.gql`)  | ✅ Yes                           | Betting config selected on the `QuizQuestion` interface (lines 20–25, 61–65), so every type gets it. Multipliers missing.                                                                |
| `QuizBettingModule.vue` — betting mode                | ✅ Yes                           | Props are pure numbers (`availablePoints`, min/max %/abs). No question-type awareness.                                                                                                   |
| `QuizBettingModule.vue` — results mode                | ✅ Yes (2026-10-27)              | Takes `outcome` + `multiplier`. The derived multiplier is gone; `isWin` follows the verdict.                                                                                            |
| `resolveFooterState` (`useQuizViewState.ts`)          | ⚠️ Mode-gated                    | Betting is only visible in `session-*` modes; `'normal'` returns `{ visible: false }` unconditionally (line 107).                                                                        |
| `QuizOrderingQuestion.vue`                            | ✅ Full flow                     | `session-betting` + `canChangeBet: true`.                                                                                                                                                |
| `QuizPredefinedQuestion.vue`                          | ⚠️ Partial                       | Computes `session-betting`, but `canChangeBet: false` (line 152) — one-shot bets.                                                                                                        |
| `QuizNumberQuestion.vue` / `QuizFreeTextQuestion.vue` | ❌ No                            | `mode: props.readonly ? 'review' : 'normal'` (lines 94 / 93) — module can never show.                                                                                                    |
| `QuizChallenge.vue` wiring                            | ⚠️ Partly (2026-10-27)           | The client-side grading is gone — verdict and payout come from the response. `isSingleOrderingQuestion` still gates the results view (finding 1, still open).                           |
| Admin question editor                                 | ⚠️ Generic but incomplete        | Betting block renders for every question type (no warning that 3 of 5 can't use it); no multiplier fields; strings hardcoded Norwegian.                                                  |
| Admin session operation                               | ✅ Yes (since the sessions page) | `admin/projects/[projectId]/challenges/[challengeId]/sessions.vue` has the full lifecycle. The PC26 page is now a convenience wrapper, not the only path.                                |
| i18n                                                  | ⚠️ Ordering-worded + dead keys   | `correctCount` reads "You had {count} correct answers"; 4 unused keys in 14 locale files.                                                                                                |

---

## Findings

### 1. `QuizChallenge.vue:315` still routes results on "single ORDERING question"

```ts
// Special case for PC26 Game Night betting
const isSingleOrderingQuestion = computed(
  () =>
    questions.value.length === 1 &&
    questions.value[0]?.__typename === "OrderingQuestion",
);
```

`resolveQuizViewState` only shows the inline question view (and therefore the betting results
module) for a FINISHED session when this is true. For every other betting quiz — a PREDEFINED
question, or more than one question — a completed submission routes to `'results'`, i.e.
`QuizResult.vue`, which contains **no betting display at all** (its only `bet` match is
`betterLuckNextTime`).

**Consequence:** a user who bets on a PREDEFINED question is paid out by the backend but is never
shown the stake, the odds, or the win/loss. Their points simply change.

This is the single biggest blocker to the frontend being generic: betting results have no home
outside the one-ordering-question layout.

### 2. Correctness is computed client-side, ORDERING-only — and it leaks the answer

> **Resolved 2026-10-27.**

`bettingCorrectCount` (`QuizChallenge.vue:338`) re-implements grading in the browser: it sorts
`question.orderingItems` by `correctOrder` and counts positional matches. This only exists because
the module wants `correctCount`/`totalCount`.

Two problems:

- **It can't generalize.** There is no `correctOrder` equivalent for NUMBER / FREE_TEXT, and for a
  single-answer PREDEFINED question "N of M correct" is the wrong frame anyway (see 6). The backend
  already stores the verdict — `quizgrading` sets `is_correct` on the response — so the client
  should read it, not recompute it.
- **It forces correct answers onto the wire.** To count positions, the client needs `correctOrder`
  while the question is still live. The backend therefore gates it
  (`quizzes.resolvers.go:2737` — `correctOrder` is `nil` unless the most recent session is
  `FINISHED`). **`QuizPredefinedAnswer.isCorrect` has no such gate**
  (`quizzes.resolvers.go:2787` returns `&obj.IsCorrectValue` unconditionally), and the player-facing
  `PredefinedAnswerFields` fragment selects it — verified in `generated.ts`, the `ChallengePage`
  query's `predefinedAnswers` includes `isCorrect`. So on the question type the backend was
  generalized _for_, a user betting in an OPEN session is handed the correct answer in the response
  payload. `revealCorrectAnswers` only controls rendering, not the data.

  This is a backend gate, but it is listed here because the frontend's client-side-grading design is
  what created the requirement, and the gap only becomes exploitable once betting moves to
  PREDEFINED.

  > **Resolved 2026-10-27:** `QuizPredefinedAnswer.isCorrect` is now gated, and `isCorrect` was
  > hoisted onto the `QuizResponse` interface. See the 2026-10-27 entry in
  > `[[betting-module-genericity-review]]`.

### 3. The betting module can never render for NUMBER and FREE_TEXT

`resolveFooterState` makes `bettingModule` visible only in `session-betting`, `session-locked` and
`session-results`. `QuizNumberQuestion.vue:94` and `QuizFreeTextQuestion.vue:93` hardcode
`mode: props.readonly ? 'review' : 'normal'`, so they can never reach those modes — even inside an
OPEN session. They still receive and forward `:bet-amount="isBettingEnabled ? currentBetAmount : undefined"`
(`QuizChallenge.vue:681`, `:711`), but `currentBetAmount` stays `0` because no slider exists.

Since R2 (bet 0 = no bet) this no longer blocks submission — it just means betting is configurable
in admin and silently inert for these types. Nothing warns the admin.

**Root cause:** betting is wired through the per-question-type action state, but betting is
question-type-independent by nature. The session state alone determines the betting phase. Hoisting
betting out of `QuizActionState` (or giving every component the same three session modes) removes
this whole class of bug.

### 4. PREDEFINED bets are one-shot

`QuizPredefinedQuestion.vue:152` pins `canChangeBet: false` while `QuizOrderingQuestion.vue:301`
sets `true`. There is no product reason recorded for the difference; it reads as "ORDERING was the
one that got built". The backend allows updating the answer and the bet while the session is OPEN
(R4) for all types — though note `updateQuizResponse` itself still only rewrites ORDERING answers
(finding 3 in `[[betting-module-genericity-review]]`), so enabling this needs the backend change too.

### 5. The slider's maximum doesn't match the server's

`challenges/[challengeId].vue:78` passes `:user-score="data.myCurrentProject?.myPoints ?? 0"`, and
that value goes straight into `:available-points` (`QuizChallenge.vue:810`, `:824`, `:840`). The
backend validates against **score − open stakes** (`GetUserAvailableBetPoints`, R1).

With one betting question per session this never diverged. With several — the generic case, and the
reason R1 exists — a user with open bets is offered a slider range the server will reject, and the
rejection arrives only on submit. Already filed as an open follow-up in `[[betting-edge-cases]]`
("expose the available points, e.g. `availableBetPoints`"); repeating it here because it is a
frontend-visible correctness bug, not just a nicety.

Percentage limits compound it: `effectiveMax` (line 45) takes a percentage of the _wrong_ base.

### 6. Result copy and props assume "N of M positions"

> **Resolved 2026-10-27.**

- `quiz.betting.correctCount` = _"You had {count} correct answers."_ — for a single-answer
  PREDEFINED or NUMBER question there is one answer, not a count.
- `correctCount` / `totalCount` are the module's only correctness inputs. A generic module wants
  an outcome (`correct: boolean`, or a normalized ratio), optionally with a pre-rendered label from
  the parent.
- When both are `null` — i.e. every non-ORDERING type — the module silently drops the odds row and
  the "Betting results" row, leaving only "Your bet" and the big number. That fallback is
  accidental, not designed.

### 7. The multiplier is re-derived in the view, and the real one is unreachable

> **Resolved 2026-10-27.**

`QuizBettingModule.vue:84`: `multiplier = (pointsEarned + betAmount) / betAmount`. This reconstructs
the backend's payout from net points — a second, independent implementation of the formula.

Since 2026-10-07 the question carries the real numbers (`bettingMultiplierCorrect`,
`bettingMultiplierWrong`, fixed-point hundredths server-side). The frontend has the types from
`pnpm codegen` but **selects them in no query and uses them nowhere** (verified: zero hits outside
`generated.ts`). So:

- the admin editor cannot set a payout curve — every question silently uses 2.0× / 0×;
- the displayed `x 1.25` is a reconstruction that will drift the moment a curve is configurable or
  a payout floors;
- the user is never shown the odds _before_ betting, which is the one number a bettor wants.

Also `isWin = pointsEarned >= 0` (line 73) renders break-even as a win with `+0`. No current
multiplier yields exactly 1.0×; a configurable curve would.

### 8. i18n leftovers (unchanged since the first review)

- `quiz.betting.yourPoints` = **`"Dine poeng"`** — Norwegian sitting in `en_us.json`.
- `sessionLocked`, `results`, `yourPoints`, `winnings` — **all four unused**; present in 14 of the
  21 locale files (the other 7 never received them).

### 9. Admin editor gaps

`AdminQuizQuestionEditor.vue:514–580` renders the betting block for every question type with no
indication that NUMBER / FREE_TEXT / JSON can't use it. No multiplier fields (7). Labels are
hardcoded Norwegian (`"Aktiver betting"`, `"Min %"`, `"Maks poeng"`) — consistent with the rest of
admin, noted only as the first wall a non-Norwegian project hits.

`quizSave.ts:74` does send `clearBettingMinAbsolute` / `clearBettingMaxAbsolute`; per
`[[betting-module-genericity-review]]` the backend still ignores them, so clearing an absolute limit
in the UI appears to work and doesn't.

Validation that _is_ generic: `quizQuestionValidation.ts:56` blocks points + betting on the same
question (R8), in Norwegian, before the backend rejects it.

### 10. What is genuinely reusable today

Worth stating plainly, since most of the above is negative:

- **The betting-mode UI is sound.** `availablePoints` + four limit props + `v-model` is a clean,
  domain-neutral contract; `effectiveMin`/`effectiveMax` correctly take the stricter of percentage
  and absolute; the `hasStake` empty state handles the degenerate range. 16 component tests cover
  it, and `useQuizViewState.test.ts` covers the footer matrix for both ORDERING and PREDEFINED
  across OPEN / LOCKED / FINISHED.
- **Config plumbing is type-agnostic end to end** — interface-level fragment → `currentQuestion.betting*`
  → module props. Nothing there needs changing to support a new question type.
- **Session operation is no longer PC26-only.** `sessions.vue` (632 lines) exposes create / open /
  lock / reopen / finish / grant-access generically. The PC26 page
  (`my-church/gamenights/gamenight-[gamenight].vue`, 653 lines) remains hardcoded — `v-for="gamenight in 2"`,
  a literal `pc26.bcc.media` films URL, a date literal, `frontendConfig` keys
  `gamenight_N_betting_quiz_id`, a `plugins/ladder-to-heaven/cryptex-admin-url` fetch — but it is a
  run-of-show wrapper for one event, not the mechanism. Fine to keep; just don't mistake it for the
  generic path.

---

## Suggested order of work

Ordered so each step is independently shippable:

1. **Give betting results a home outside the single-ORDERING layout.** Either show the bet outcome
   on `QuizResult.vue`, or drop `isSingleOrderingQuestion` in favour of "FINISHED + any question had
   a bet". Without this, PREDEFINED betting is invisible to users. (Fixes 1.)
2. ~~**Take correctness and the multiplier from the API.**~~ (done 2026-10-27, in two parts:
   the `isCorrect` gate and interface hoist in the morning, then the result screen itself.)
   `QuizBettingModule` now takes `outcome` + `multiplier`; `bettingCorrectCount` /
   `bettingTotalCount` and the derived multiplier are gone; the payout comes from a new
   `QuizResponse.settledBetMultiplier`. (Fixed 2, 6, 7 — see the 2026-10-27 entries in
   `[[betting-module-genericity-review]]`.)
3. **Hoist betting out of the per-type action state** so the slider follows the session, not the
   question component. (Fixes 3, and most of 4.)
4. **Expose `availableBetPoints`** and feed it to `:available-points`. (Fixes 5.)
5. **Admin:** multiplier fields, a hint (or a disable) for types that can't bet, honour the
   `clear*` flags backend-side. (Fixes 9.)
6. **i18n cleanup:** drop the 4 dead keys from 14 files, fix `yourPoints`, reword `correctCount`.
   (Fixes 8.)
