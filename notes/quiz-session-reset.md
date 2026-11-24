# Quiz session reset (`resetQuizSessionSubmission`)

## Rule

A user can reset (delete) **only an in-progress** submission, and only while
the session is `OPEN`. A finalized submission can't be reset: the mutation
returns `cannot reset a completed submission` and the submission is kept.
Resetting when there is no submission is a no-op and returns `true`.

## Why (fixed 2026-11)

`finalizeQuiz` writes the quiz's points to `score_journal`. The reset used to
delete the submission whatever its state, but never touched the journal. A
user could finalize, reset, start again and finalize again, earning a
100-point quiz's points any number of times. Because the completed submission
disappeared, this also got around `allowRetakes = false`. The frontend never
calls the mutation, but it is reachable through the API.

## Race safety

The completion check is part of the DELETE itself
(`DeleteInProgressSubmissionByUserAndSession`, `... AND completed_at IS NULL`,
`:execrows`), so there is no separate check-then-delete step to race:

- **Finalize first.** Finalize holds `FOR UPDATE` on the submission row. The
  DELETE waits on that lock, and after finalize commits, READ COMMITTED
  re-checks the WHERE against the updated row. The row is now completed, so
  nothing is deleted and the reset is rejected.
- **Reset first.** Finalize's `GetQuizSubmissionByIDForUpdate` finds no row
  and fails, so nothing is journaled.

`e2e/quiz_session_reset_test.go` tests both orderings deterministically. A
test transaction holds the row lock, the two mutations queue behind it in a
fixed order, and then the lock is released.

## Known gaps (not fixed here)

- **Concurrent `startQuizSession`.** `quiz_submissions` has no unique
  `(session_id, user_id)` (the `UNIQUE` in migration 00075 is on
  `quiz_session_access`). Two simultaneous starts can create two submissions,
  and each can be finalized for points. Fixing this needs a migration (a
  partial unique index), and existing duplicate rows would have to be checked
  first.
- **`allowRetakes` is not enforced by `startQuizSession`.** Within one session
  the user gets their existing submission back. Across several sessions of
  the same quiz, nothing stops a second completed attempt. Today `allowRetakes`
  is only read by the `Quiz.userCanStart` field resolver.
- **Finalize itself is idempotent per submission** (FOR UPDATE plus a
  re-check of `completed_at`). The journal's `source_id` is the quiz ID, so a
  unique constraint there would block legitimate retakes.
