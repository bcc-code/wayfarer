# Quiz Results in the Admin Panel

Branch: `feature/admin-quiz-results`

## Goal

The people running a quiz need to read the current results off a screen, understand
them without explanation, and lift the numbers — absolutes or percentages — straight
into a keynote.

## Decisions

Settled with the product owner before implementation started. Each one closes a
question the data model leaves genuinely open, so they are recorded here rather
than only in the code.

| Question | Decision |
| --- | --- |
| Scope of a results view | **Whole quiz, all sessions pooled.** One results page per quiz, every completed submission counted together. |
| NUMBER questions (no correct answer is stored) | **Summary stats + histogram** — average, median, min, max, and bucketed distribution. |
| FREE_TEXT questions (not graded) | **Grouped by normalised text**, count + %, expandable to the raw list of every answer. |
| ORDERING questions (graded all-or-nothing) | **Fully-correct % + per-position accuracy** — for each item, how many placed it in its correct slot. |
| JSON questions | **Response count only**, with a note saying the type cannot be summarised automatically. Kept visible so the question numbering stays intact. |
| Individual participants | **Aggregates only.** No names, no per-person table — the page gets projected. |
| Betting | **Out of scope** for this pass. Can be added per question later. |
| Getting numbers out | **Copy + CSV.** A copy button per question (TSV, pastes into Keynote/Excel) plus one CSV download for the whole quiz. |

### Consequence of pooling all sessions

A user who took the quiz in two sessions is counted twice. That is accepted:
`submissionCount` counts completed submissions and `participantCount` counts
distinct users, so the page can show both and the difference is visible rather
than hidden.

## What already exists

- `quiz_submissions` — one per user per run; `completed_at IS NOT NULL` marks a
  finished one. `score` / `max_score` are already computed at finalisation.
- `quiz_responses` — one row per (submission, question). Polymorphic payload:
  - `selected_answer_ids` JSONB — PREDEFINED
  - `text_response` — FREE_TEXT
  - `number_response` DECIMAL — NUMBER
  - `json_response` JSONB — JSON **and** ORDERING (ORDERING stores the submitted
    order array here)
  - `is_correct` — set for PREDEFINED and ORDERING, NULL for the rest
- **ORDERING items live in `quiz_predefined_answers`**, where `answer_order` is the
  correct position. `orderingQuestionResolver.OrderingItems` converts them. Results
  code must use the same table — there is no separate ordering-items table.
- No aggregation queries exist yet; everything in `queries/quiz_*.sql` is row-level.
- Admin quiz pages today: `challenges/[challengeId]/quiz.vue` (edit) and
  `challenges/[challengeId]/sessions.vue` (session management). The challenge page
  header links to Sessions; Results belongs next to it.
- `@unovis/vue` is a dependency, and `AdminSparkline` shows the hand-rolled inline
  SVG approach. Bars here are simple enough for plain elements — no chart library.

## Design

### Where it lives

New page `admin/projects/[projectId]/challenges/[challengeId]/results.vue`, reached
from a **Resultater** button beside **Sesjoner** in the challenge page header
(`challenges/[challengeId]/index.vue`). Shown only for `QuizChallenge`.

### Page shape

```
Utfordring > Resultater

 Quiz-navn
 128 besvarelser · 121 deltakere · 3 sesjoner
 Snittscore 6,9 / 10  (69 %)                    [⤓ Last ned CSV]
 ─────────────────────────────────────────────────────────────
 Q1  Hva er hovedstaden?              PREDEFINED   [⧉ Kopier]
     Oslo     ████████████  31   74 %  ✓
     Bergen   ███            7   17 %
     Tromsø   ██             4    9 %
     42 av 128 svarte

 Q2  Hvor mange var med?                  NUMBER   [⧉ Kopier]
     Snitt 41,6 · Median 40 · 12–95
     10–25 ████ 6 ...

 Q3  Hva tar du med deg?                FREE_TEXT  [⧉ Kopier]
     "håp" ███████ 9  21 % ... [Vis alle 42 svar]
```

Every question block shares one bar-row primitive so the page reads as one system:
label, proportional bar, absolute, percentage, optional correct-answer tick.

### GraphQL

New file `gql/quiz_results.graphqls`. One query, admin-only. Note the directive rule
in the root `CLAUDE.md`: `@requireRole` goes **on mutations only** — this query is
authorised inside the resolver like the other admin quiz reads.

```graphql
extend type Query {
    quizResults(quizId: ID!): QuizResults!
}

type QuizResults {
    quiz: Quiz!
    submissionCount: Int!        # completed submissions
    participantCount: Int!       # distinct users
    sessionCount: Int!
    averageScore: Float
    averageMaxScore: Float
    averageScorePercentage: Float
    questions: [QuizQuestionResults!]!
}

interface QuizQuestionResults {
    question: QuizQuestion!
    responseCount: Int!          # submissions that answered this question
}

type PredefinedQuestionResults implements QuizQuestionResults {
    question: QuizQuestion!
    responseCount: Int!
    correctCount: Int!
    options: [PredefinedOptionResult!]!
}
type PredefinedOptionResult {
    answer: QuizPredefinedAnswer!
    count: Int!
    percentage: Float!
    isCorrect: Boolean!
}

type NumberQuestionResults implements QuizQuestionResults {
    question: QuizQuestion!
    responseCount: Int!
    average: Float
    median: Float
    min: Float
    max: Float
    buckets: [NumberBucket!]!
}
type NumberBucket { from: Float!  to: Float!  count: Int!  percentage: Float! }

type FreeTextQuestionResults implements QuizQuestionResults {
    question: QuizQuestion!
    responseCount: Int!
    distinctCount: Int!
    groups: [FreeTextGroup!]!    # desc by count
    responses: [String!]!        # raw, for the expand
}
type FreeTextGroup { text: String!  count: Int!  percentage: Float! }

type OrderingQuestionResults implements QuizQuestionResults {
    question: QuizQuestion!
    responseCount: Int!
    fullyCorrectCount: Int!
    items: [OrderingItemResult!]!
}
type OrderingItemResult {
    item: QuizOrderingItem!
    correctPosition: Int!
    correctlyPlacedCount: Int!
    percentage: Float!
}

type JsonQuestionResults implements QuizQuestionResults {
    question: QuizQuestion!
    responseCount: Int!
}
```

`percentage` is always a share of that question's `responseCount`, never of
`submissionCount` — a question people skipped must not read as unpopular answers.

### Aggregation strategy

Counting happens in SQL; shaping happens in Go behind pure functions so it can be
unit-tested without a database.

**SQL (new `queries/quiz_results.sql`)** — all filtered to
`quiz_id = @quizId AND submissions.completed_at IS NOT NULL`:

1. `GetQuizResultsSummary` `:one` — submissions, distinct users, distinct sessions,
   avg score, avg max_score.
2. `GetQuizResponseCountsByQuestion` `:many` — per question: responses, correct count.
3. `GetQuizPredefinedAnswerCounts` `:many` — per (question, answer) count, unnesting
   `selected_answer_ids` with `jsonb_array_elements_text`.
4. `GetQuizNumberResponses` `:many` — (question_id, value) rows.
5. `GetQuizFreeTextResponses` `:many` — (question_id, text) rows.
6. `GetQuizOrderingResponses` `:many` — (question_id, json_response) rows.

Six queries for the whole page regardless of question count — no per-question
round-trip. Queries 4–6 return raw rows because bucketing, grouping and
position-matching are shaping decisions, not counting ones.

**Go (new `internal/graph/api/quiz_results.go`, non-resolver file so `make generate`
cannot overwrite it)** — pure, table-testable helpers:

- `numberStats(values []float64) (avg, median, min, max float64)`
- `numberBuckets(values []float64, target int) []Bucket` — ~6 buckets; a single
  distinct value must yield one bucket, not a divide-by-zero
- `normalizeFreeText(s string) string` — trim, collapse inner whitespace, casefold
- `groupFreeText(responses []string) []FreeTextGroup` — groups by the normalised
  form, displays the most common original spelling, sorts by count desc then text
  asc so equal counts are stable
- `orderingAccuracy(submitted [][]string, correct []string) (fullyCorrect int, perPosition []int)`

### Export

- **Per question** — copy button writes TSV (`label\tcount\tpercentage`) to the
  clipboard. TSV because that is what pastes into Keynote and Excel as a table.
- **Whole quiz** — one CSV, one row per option/bucket/group, columns:
  `question_order, question_text, question_type, label, count, percentage`.
  Built client-side from the same query result; no extra endpoint.

Both live in a pure util (`layers/admin/app/utils/quizResultsExport.ts`) so the
formatting is unit-tested rather than asserted through the DOM.

## Tasks

### Backend

- [x] 1. `gql/quiz_results.graphqls` — types above; registered in `gqlgen.yml`
      (its `schema:` list is explicit, a new file there is **not** picked up
      automatically) and `make generate`
- [x] 2. `queries/quiz_results.sql` — the six aggregate queries; `make generate`
- [x] 3. `quiz_results.go` — pure helpers (stats, buckets, free-text grouping,
      ordering accuracy) plus the assembly and input-loading helpers
- [x] 4. Unit tests for every helper in 3, including the degenerate cases: no
      responses, one response, all-identical values, unanswered question
- [x] 5. `quizResults` resolver — authorises via `RoleService.CanManageProject`,
      fans out the queries, assembles per question type
- [ ] 6. E2E test covering the resolver end to end (`e2e/quiz_results_test.go`),
      alongside the existing `quiz_sessions_test.go` / `quiz_betting_test.go`.
      The package's convention is to unit-test pure helpers and cover resolvers
      in e2e — only one resolver in `internal/graph/api` has a mock-based test —
      so this is where resolver coverage belongs.
- [x] 7. `make fmt` && `make test` — all green

### Frontend

- [x] 8. `pnpm codegen` after the schema lands
- [x] 9. `AdminQuizResultBar.vue` — the shared bar row (label, bar, count, %, tick)
- [x] 10. Per-type blocks: `AdminQuizResultsPredefined/Number/FreeText/Ordering/Json.vue`,
      dispatched by `AdminQuizResultsQuestion.vue` (which also owns the card
      header and the copy button)
- [x] 11. `results.vue` page — summary header, question list, loading/error via
      `AdminQueryState`
- [x] 12. `quizResultsExport.ts` — TSV clipboard + CSV builder (pure)
- [x] 13. Unit tests for `quizResultsExport.ts` (37 cases)
- [x] 14. Component tests for the bar and every per-type block, including the
      zero-response and nothing-to-copy states (15 cases)
- [x] 15. **Resultater** button on the challenge page, beside **Sesjoner**
- [x] 16. `pnpm typecheck`, `pnpm lint`, `pnpm test`, `pnpm build` — all green

### Wrap-up

- [x] 17. Update this note with anything the implementation changed
- [x] 18. Note in `notes/quiz-system-implementation.md` that results aggregation exists

## What the implementation changed from the plan

- **`gqlgen.yml` has an explicit `schema:` list.** A new `gql/*.graphqls` file is
  silently ignored until it is added there — no error, the resolver stub simply
  never appears. `frontend/codegen.ts` globs `../gql/*.graphqls`, so the two
  disagree; only the backend needs the manual entry.
- **`question` and `answer` are embedded, not force-resolved.** The resolver has
  the questions and their answers in hand while assembling, so a `forceResolver`
  round-trip per row would have re-fetched what it already held. Only
  `QuizResults.quiz` stays a resolver field (with a `QuizID` extra field), since
  the page usually does not ask for it.
- **`GetQuizResultsSummary` returns `scored_count` and COALESCEd averages.**
  `AVG(...)::float8` over zero rows is NULL, which sqlc types as non-nullable
  `float64` and which then fails to scan — a quiz with no completed submissions
  is the *normal* state before a session runs, so this would have broken on
  first use. `scored_count` is what separates "nothing scored yet" (averages left
  null) from "the average really is zero".
- **Five queries, not six.** `GetQuizOrderingResponses` dropped its `is_correct`
  column: the all-or-nothing count already comes from
  `GetQuizResponseCountsByQuestion`, and two sources for one number can only
  disagree.
- **Bars are a single hue.** Shading each bar by its own value would double-encode
  length as colour; answer options are nominal, so there is no order for a ramp
  to carry. Correctness is an icon **and** the word "Riktig", never colour alone.
- **The CSV carries a UTF-8 BOM**, built via `String.fromCharCode(0xfeff)` —
  Prettier rewrites a `\uFEFF` escape into the invisible character itself, which
  then trips eslint's `no-irregular-whitespace`.
- **`test/unit/routes.test.ts` holds a committed route snapshot.** Adding a page
  fails it until the snapshot is updated (`vitest -u`).

## Open risks

- `schema.sql` is stale for quizzes (no sessions, no betting, no ORDERING in the
  type constraint). Read the migrations, not `schema.sql`, when writing the
  aggregate SQL. Worth a separate fix, out of scope here.
- Free-text grouping is naive by design — normalised exact match, no stemming or
  fuzzy clustering. If the camp answers in several languages this will under-group;
  revisit only if it actually bites.
- Bucket count for NUMBER is fixed at ~6. Step-based questions with few distinct
  values may look better as exact values; the shaping helper is the single place to
  change that if it reads badly in practice.
