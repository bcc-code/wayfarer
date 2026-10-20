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

Read side: `imageObject: Image` on the `QuizQuestion` interface and on all five
implementations (`PredefinedQuestion`, `FreeTextQuestion`, `NumberQuestion`,
`JsonQuestion`, `OrderingQuestion`). It is a force-resolver field backed by the
`ImageURL` extra field on each model (see `gqlgen.yml`), resolved through
`resolveImageByURL` so the metadata loader fills in width/height/blurhash.

There is deliberately no raw `image: String` output field — `Quiz.image` and
`Challenge.image` only still exist because they predate `imageObject`, and both
are deprecated. Admins read the current URL as `imageObject { url }`.

Write side: `image: String` on `CreateQuizQuestionInput` and
`UpdateQuizQuestionInput`.

**Clearing:** an empty string clears the image. `UpdateQuizQuestion`'s SQL is
`COALESCE(narg, column)` throughout, so a null means "leave alone" and could
never remove a value. Rather than add another `clearX` boolean, the resolver
treats a blank (or whitespace-only) `image` as an explicit clear and the query
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
`RETURNING`, the INSERT and the UPDATE) and the five `ImageObject` resolvers in
`quizzes.resolvers.go`.

## Frontend touch points

- `app/graphql/fragments/quiz.gql` — `imageObject { ...ImageFields }` on both
  `QuizQuestionFields` (admin) and `QuizQuestionUserFields` (participant).
- `AdminQuizForm.vue` — `image?: string` on `QuizQuestionFormData`.
- `AdminQuizQuestionEditor.vue` — an `AdminFileUpload` field; carried through
  `handleSave` for every question type, not just one.
- `AdminQuizQuestionPreview.vue` — the preview feeds the same
  `QuizQuestionHeading` the participant sees, so the image appears there too.
- `QuizQuestionHeading.vue` — renders the image above the question with
  `DesignImage` (blurhash placeholder, intrinsic aspect ratio).
- `quizSave.ts` (`quizQuestionInput`) — sends `image`, normalising undefined to
  `''` so removing an image in the editor actually clears it.
- `challenges/[challengeId]/quiz.vue` — maps `imageObject.url` back into the
  form state.
