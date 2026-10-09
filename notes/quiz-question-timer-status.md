# Quiz question timers — current support

_Status audit, 2026-10-09. Question asked: "we have a max-time per question in the
backend, but we don't show it anywhere, do we?" Answer: correct._

Short version: the data model, the GraphQL API and the admin UI all support a
per-question time limit. The user-facing quiz player ignores it completely, and
nothing enforces it server-side.

## What exists

### Database

Two independent timeouts:

| Column                           | Meaning                   |
| -------------------------------- | ------------------------- |
| `quizzes.timeout_seconds`        | Budget for the whole quiz |
| `quiz_questions.timeout_seconds` | Budget for one question   |

The per-question column was introduced by
`backend/internal/database/migrations/00040_move_question_timeout_to_questions.sql`.
Before that, `quizzes.question_timeout_seconds` held a single value applied to
every question in the quiz, and a `CHECK` constraint made the two timeouts
mutually exclusive. Migration 00040 dropped both that column and the constraint,
so **the two timeouts can now be set together**.

### GraphQL

`timeoutSeconds: Int` sits on the `QuizQuestion` interface and is repeated on all
five implementations (`PredefinedQuestion`, `FreeTextQuestion`, `NumberQuestion`,
`JsonQuestion`, `OrderingQuestion`) in `gql/quizzes.graphqls`. Readable and
writable through the question create/update mutations.

### Admin UI

Fully wired, both levels:

- `AdminQuizQuestionEditor.vue:421` — per-question "Tidsbegrensning (sekunder)"
- `AdminQuizForm.vue:373` — quiz-level "Tidsbegrensning (sekunder)"

Both help texts promise that _"den strengeste av de to gjelder"_ (the strictest
of the two applies). **No code implements that rule.**

### Frontend data fetch

The value already reaches the client. `QuizQuestionUserFields`
(`frontend/app/graphql/fragments/quiz.gql:52`) selects `timeoutSeconds`, as does
the challenge page query. It is then unused.

## What is missing

### No timer UI

`QuizChallenge.vue`, the whole `frontend/layers/user/app/components/challenges/quiz/`
tree and `useQuizViewState.ts` contain no reference to `timeout`, `timer`,
`countdown` or elapsed time. The field is fetched and dropped.

### No per-question enforcement

Only the quiz-level timeout is enforced:

- `quiz_sessions.resolvers.go:776-780` computes `expires_at` from
  `quiz.TimeoutSeconds` when a submission starts.
- `quizzes.resolvers.go:809`, `:1191`, `:1250` reject answers/submits after that
  timestamp; `:2806` backs the `QuizSubmission.isExpired` field.

`quiz_questions.timeout_seconds` is read, converted
(`quiz_helpers.go:convertQuestionTimeoutSeconds`) and returned — never compared
against anything.

### `timeSpentSeconds` is decorative

The input exists on the answer mutations and the value is stored on
`quiz_responses`, but the frontend never sends it and the backend never validates
it against the question's timeout.

### `isExpired` / `expiresAt` are queried but unused

Both appear in `challenge.gql` and `quizes.gql` selection sets, but no `.vue` or
`.ts` file reads them. So even the quiz-level deadline is invisible to users —
they simply get an error when they submit too late.

### `auto_submitted` is not time-driven

`AutoSubmitSessionSubmissions` has exactly one caller
(`quiz_sessions.resolvers.go:383`, closing a live session from admin). Nothing
auto-submits on expiry.

## Implications for building a visible timer

Mostly frontend work:

1. Countdown component in the quiz player driven by `question.timeoutSeconds`,
   clamped by the session's `expiresAt` when both are set (this is where the
   "strictest of the two" promise would finally become true).
2. Auto-advance or auto-submit at zero.
3. Actually send `timeSpentSeconds`.

If it should be more than an honour system, the backend needs per-question
deadline tracking. **There is currently no server-side record of when a question
was shown** — `quiz_responses` only has `answered_at` — so a client can sit on a
question indefinitely. Enforcing per-question time requires either persisting a
"question served at" timestamp or deriving deadlines from a server-driven
question sequence.

## Stale notes

These predate migration 00040 and describe the two timeouts as mutually exclusive
with `questionTimeoutSeconds` living on the quiz:

- `notes/quiz-system-implementation.md` (lines ~71, 268)
- `notes/quiz_implementation_summary.md` (lines ~19, 174-175)

`notes/quiz_testing_guide.md:840` already records "timeout enforcement is
client-side" as a known gap — still accurate, and now understated: there is no
client-side enforcement either.

## Clock sync: use `Query.currentTime`, not the device clock

**Decision (2026-10-09): the timer UI must be driven by the server clock, not
`Date.now()` on its own.** Device clocks are wrong often enough — manually set,
drifting, in the wrong timezone offset, or deliberately rolled back — that a
countdown computed straight from `Date.now()` disagrees with the deadline the
backend will actually enforce. The user then either loses time they should have
had, or sees time remaining on an answer the server has already rejected as
expired.

### What the API gives us

`gql/schema.graphqls:12`:

```graphql
type Query {
  currentTime: DateTime!
}
```

Resolved in `backend/internal/graph/api/schema.resolvers.go:46` as
`time.Now().UTC()`. No auth, no arguments. **Nothing in the frontend queries it
today** (the `currentTime` hits in `ProjectInfoBanner.vue` are a local
variable, unrelated).

### How to use it

1. Query `currentTime` once when the quiz player mounts, alongside the quiz
   data, and record the client timestamp at the moment the response arrives.
2. Store the offset: `skew = serverTime - clientTimeAtResponse`. Round-trip
   latency makes this off by up to half an RTT; that is well inside tolerance
   for a seconds-granularity countdown.
3. Derive every deadline comparison from `Date.now() + skew` rather than
   `Date.now()`. That includes the per-question `timeoutSeconds` countdown and
   the session-level `expiresAt` clamp described above.
4. Tick with a monotonic source (`performance.now()` deltas) so a clock change
   mid-question doesn't make the countdown jump; re-fetch `currentTime` on
   resume from background/visibility change, since long suspensions are where
   drift actually accumulates.

The skew only needs to be correct to roughly a second, so one fetch per quiz
session is enough — no polling.

### Caveat

This keeps the _display_ honest; it is not enforcement. Until the backend
records when a question was served (see "Implications for building a visible
timer"), a client can still ignore the countdown entirely. Clock sync makes the
timer trustworthy for honest users, not tamper-proof.
