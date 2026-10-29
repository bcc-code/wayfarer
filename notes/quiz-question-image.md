# Image per quiz question

A quiz question can carry an image, shown above the question text in the
participant app and in the admin preview.

## Storage

`quiz_questions.image_url VARCHAR(500)` — migration
`00113_add_quiz_question_image.sql`. Same shape and length as
`quizzes.image_url`; the value is a public URL produced by the file upload
endpoint (`AdminFileUpload`), not a stored blob.

The column is not translated. `quiz_question_translations` is untouched — an
image is not copy.

## GraphQL

Read side: `image: Image` on the `QuizQuestion` interface and on all five
implementations (`PredefinedQuestion`, `FreeTextQuestion`, `NumberQuestion`,
`JsonQuestion`, `OrderingQuestion`). It is a force-resolver field backed by the
`ImageURL` extra field on each model (see `gqlgen.yml`), resolved through
`resolveImageByURL` so the metadata loader fills in width/height/blurhash.

This is the same shape as every other image in the API: outputs are
`image: Image`, inputs take the URL as `imageUrl: String`. There is no raw URL
output field; admins read the current URL as `image { url }`. (Until the
rename, the object field was called `imageObject` and sat next to a deprecated
`image: String` on `Quiz`, `Challenge`, `User` and `LeaderboardEntry`; both
are gone.)

Write side: `imageUrl: String` on `CreateQuizQuestionInput` and
`UpdateQuizQuestionInput`.

**Clearing:** an empty string clears the image. `UpdateQuizQuestion`'s SQL is
`COALESCE(narg, column)` throughout, so a null means "leave alone" and could
never remove a value. Rather than add another `clearX` boolean, the resolver
treats a blank (or whitespace-only) `imageUrl` as an explicit clear and the query
takes a separate `clearimage` flag:

```sql
image_url = CASE WHEN @clearimage::bool THEN NULL
                 ELSE COALESCE(sqlc.narg('imageurl')::text, image_url) END
```

Note for anyone copying the `clearX` convention from this input:
`clearBettingMinAbsolute` / `clearBettingMaxAbsolute` are declared in the schema
but read by nothing — they are dead. Do not take them as a working example.

## Backend touch points

Adding a column to `quiz_questions` means touching every place a question row
becomes a model. There are two, and they do not share code:

- `internal/graph/api/quiz_helpers.go` — the `quizQuestionRow` interface, its
  three adapters (create/update/get-by-id) and the five `convertTo*Question`
  functions.
- `internal/loaders/quiz_questions_by_quiz.go` — a second, hand-written copy of
  the same switch for the batch loader.

Plus `internal/database/queries/quiz_questions.sql` (every SELECT list,
`RETURNING`, the INSERT and the UPDATE) and the five `Image` resolvers in
`quizzes.resolvers.go`.

## Frontend touch points

- `app/graphql/fragments/quiz.gql` — `image { ...ImageFields }` on both
  `QuizQuestionFields` (admin) and `QuizQuestionUserFields` (participant).
- `AdminQuizForm.vue` — `image?: string` on `QuizQuestionFormData`.
- `AdminQuizQuestionEditor.vue` — an `AdminFileUpload` field; carried through
  `handleSave` for every question type, not just one.
- `AdminQuizQuestionPreview.vue` — the preview feeds the same
  `QuizQuestionHeading` the participant sees, so the image appears there too.
- `QuizQuestionHeading.vue` — renders the image above the question with
  `DesignImage` (blurhash placeholder, intrinsic aspect ratio).
- `quizSave.ts` (`quizQuestionInput`) — sends the form's `image` as
  `imageUrl`, normalising undefined to `''` so removing an image in the editor
  actually clears it. The form state keeps the name `image`; only the GraphQL
  input is `imageUrl`.
- `challenges/[challengeId]/quiz.vue` — maps `image.url` back into the
  form state.

---

# Answer labels (A. B. C. / 1. 2. 3.)

A predefined question can number or letter its alternatives. The label is
**derived from the answer's position when it is rendered — never stored per
answer.**

## Why derived

Answers are always served `ORDER BY answer_order` and are never shuffled —
only questions are (`RandomizeQuestions`, `quiz_sessions.resolvers.go`). So a
label taken from position cannot disagree with the order the alternatives are
shown in. A stored label could: an answer labelled "C" rendering first is
worse than no label. A stored one would also need uniqueness validation and
would travel through Phrase inside the answer text.

Before this existed, organisers typed "A. " into the answer text by hand.
Dragging the answers in the editor then left the letters out of order, since
they were just characters someone had typed.

## Storage

`quiz_questions.answer_label_style VARCHAR(16) NOT NULL DEFAULT 'NONE'`,
CHECK-constrained to `NONE | NUMERIC | ALPHA` — migration
`00114_add_quiz_answer_label_style.sql`.

It is per question, not per quiz: a question keyed to a lettered image grid
wants `ALPHA` while the plain question next to it wants nothing.

## GraphQL

`enum QuizAnswerLabelStyle { NONE NUMERIC ALPHA }`.

`answerLabelStyle: QuizAnswerLabelStyle!` sits on `PredefinedQuestion` only,
not on the `QuizQuestion` interface — only predefined questions have
alternatives to label, the same reason `allowMultipleSelection` lives there.
The input field is nullable on both `CreateQuizQuestionInput` and
`UpdateQuizQuestionInput`; omitted means NONE on create and "leave alone" on
update.

An unrecognised stored value converts to `NONE` rather than erroring
(`convertAnswerLabelStyle` in `quiz_helpers.go`, `answerLabelStyle` in
`loaders/quiz_questions_by_quiz.go` — the loader keeps its own copy of the
row→model switch, as it does for every other question field).

## Rendering

`layers/user/app/utils/quizAnswerLabels.ts` → `quizAnswerLabel(index, style)`.
Alpha runs A..Z then AA, AB — spreadsheet style. More than 26 alternatives is
not expected, but running out of letters must repeat nothing rather than give
two answers the same label.

In `QuizAlternative` the label is **its own element, a sibling of the text
block** — not a prefix inside the title:

```
button (flex)
├── span            A.        ← own column, min-w-[2ch], shrink-0
└── span (flex-col)           ← title and description share this left edge
    ├── span        Kjøttgryten fra Egypt
    └── span        En gryte som symboliserer …
```

Glued to the front of the title, a label moves with each title's width and the
description starts at a different x than the title it explains. Two cva
variants carry the alignment: `alignStart` (a label or a description means the
text is left-aligned) and `withDescription` (`items-start`, so the label meets
the title rather than the paragraph's vertical centre).

## Touch points

Backend: `quiz_questions.sql` (every select list, RETURNING, the INSERT and
UPDATE), the `quizQuestionRow` interface and its three adapters,
`convertToPredefinedQuestion`, the batch loader, and `AddQuizQuestion` /
`UpdateQuizQuestion`.

Frontend: both fragments in `quiz.gql` (`... on PredefinedQuestion`),
`QuizQuestionFormData`, the `USelect` in `AdminQuizQuestionEditor` (saved in
the predefined branch of `handleSave`), `quizQuestionInput`, the admin quiz
page's form mapping, `AdminQuizQuestionPreview`, and `QuizPredefinedQuestion`,
which is the single place alternatives are rendered — review mode and the
admin preview go through it.

## Existing content

Answers that already have "A. " typed into their text will read "A. A. …" if
that question is switched to `ALPHA`. Default `NONE` leaves them as they are;
the prefixes have to come out of the text by hand when a question is switched.
