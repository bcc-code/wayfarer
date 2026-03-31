# Challenges Page Database Performance Report

**Date:** 2026-03-31
**Database:** Neon PostgreSQL (wayfarer)
**Target:** Reduce CPU load from challenges page queries

---

## 1. Summary

The challenges page is significantly lighter than the front page. The main queries are fast (~1ms each for the challenge list) but two dataloader queries suffer from a **type mismatch bug** that causes full table scans instead of PK lookups: `GetBulkUserCompletionTimestamps` and `GetBulkUserEnrollmentTimestamps`. Fixing the type cast brings these from **27ms -> 1.8ms** each.

---

## 2. What the Challenges Page Loads

### ActiveChallengesPage query
```graphql
query ActiveChallengesPage {
  myCurrentProject {
    myTeam { joinCode }
    activeChallenges { ...ChallengeFields }
  }
}
```

### ChallengeFields fragment includes (for quiz challenges):
- Quiz metadata (timeoutSeconds, allowRetakes, etc.)
- `userActiveSubmission` - active quiz submission
- `userActiveSession` - active quiz session
- `userSubmissions` - all user's submissions with scores

### ChallengePage (individual challenge detail):
- Same as above plus full `responses` with answers for each submission

---

## 3. Query-by-Query Benchmark

### 3.1 Challenge List Loading

| Operation | Query | Time | Notes |
|-----------|-------|------|-------|
| `ChallengesByProjectLoader` | `GetChallengesByProjectIDs` | **0.13ms** | Seq scan on 13 rows, trivial |
| `GetUserEnrolledChallengeIDsInProject` | JOIN enrollments + challenges | **3.9ms** | Index scan on enrollments PK |
| `GetBulkUserSessionAccessQuizIDs` | JOIN sessions + access | **5.6ms** | Index scan, 212 sessions total |

### 3.2 Challenge Filtering Pipeline

The `getFilteredChallenges` function in `challenges.go` executes these steps:

1. Load all project challenges via `ChallengesByProjectLoader` (0.13ms)
2. Batch load quizzes via `QuizByChallengeIDLoader` (~1ms)
3. Batch check session access via `GetBulkUserSessionAccessQuizIDs` (5.6ms)
4. Batch load enrolled IDs via `GetUserEnrolledChallengeIDsInProject` (3.9ms)
5. Filter visible challenges in-memory
6. Batch load completion timestamps via `UserChallengeCompletionTimestampLoader` (**27ms - BUG**)
7. Batch load enrollment timestamps via `UserChallengeEnrollmentTimestampLoader` (**27ms - BUG**)

**Total: ~65ms, of which 54ms (83%) is the two buggy bulk queries**

### 3.3 Quiz Data (per quiz challenge)

| Operation | Query | Time | Notes |
|-----------|-------|------|-------|
| `QuizSubmissionsByUserLoader` | `GetQuizSubmissionsByUserIDs` | **1.1ms** | Index scan, avg 5 subs/user |
| Quiz responses (detail page only) | `quiz_responses` by submission | **9.3ms** | Nested loop, avg 3 responses/sub |

---

## 4. The Bug: Type Mismatch in Bulk Queries

### Root Cause

Both `GetBulkUserCompletionTimestamps` and `GetBulkUserEnrollmentTimestamps` use this pattern:

```sql
SELECT user_id, challenge_id, completed_at
FROM user_challenge_completions
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::text[]), unnest(@challengeids::text[])
);
```

The `unnest` returns `text` values, but `user_id` and `challenge_id` are `char(28)` (`bpchar`). PostgreSQL **cannot use the PK index** (`user_challenge_completions_pkey ON (user_id, challenge_id)`) because of the type mismatch. Instead, it does a **Hash Semi Join with a full sequential scan** of the entire table.

### Evidence

```
-- Current (seq scan): 27ms, 960 buffers
Hash Semi Join  (actual time=27.981)
  -> Seq Scan on user_challenge_completions (rows=52209)

-- Fixed (PK index): 1.8ms, 15 buffers
Nested Loop  (actual time=1.763)
  -> Index Scan using user_challenge_completions_pkey (loops=5)
```

### The Fix

Cast the unnest output to match the column type:

```sql
-- Current (broken)
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::text[]), unnest(@challengeids::text[])
);

-- Fixed (uses PK index)
FROM unnest(@userids::char(28)[], @challengeids::char(28)[]) AS pairs(uid, cid)
JOIN user_challenge_completions ucc ON ucc.user_id = pairs.uid AND ucc.challenge_id = pairs.cid;
```

Or alternatively, keep the same pattern but cast:
```sql
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::char(28)[]), unnest(@challengeids::char(28)[])
);
```

### Affected Queries

| Query | File | Line | Table Scanned | Rows |
|-------|------|------|---------------|------|
| `GetBulkUserCompletionTimestamps` | `challenge_completions.sql` | 36-41 | `user_challenge_completions` | 52,209 |
| `GetBulkUserEnrollmentTimestamps` | `challenge_enrollments.sql` | 39-44 | `user_challenge_enrollments` | ~10,000 (smaller, but same bug) |

---

## 5. Quiz Submissions: Potential Issue at Scale

The `QuizSubmissionsByUserLoader` loads **all** of a user's quiz submissions across all quizzes, then filters in Go code per-quiz. Currently this is fine (avg 5 submissions/user, max 8), but if quizzes grow, this becomes an N+1 anti-pattern.

The `UserCanStart`, `UserActiveSubmission`, and `UserSubmissions` resolvers all call the same `QuizSubmissionsByUserLoader` and filter in memory. With the dataloader deduplication this works well - one DB query serves all three resolvers.

**Current:** Not a problem (34K total submissions, fast index scan)
**Future risk:** If submission count grows significantly per user

---

## 6. Data Scale Reference

| Table | Rows | Has PK Index? |
|-------|------|---------------|
| `user_challenge_completions` | 52,209 | Yes: `(user_id, challenge_id)` |
| `quiz_responses` | 101,164 | Yes: `(id)`, unique on `(submission_id, question_id)` |
| `quiz_session_access` | 75,061 | Yes: `(id)`, unique on `(session_id, user_id)` |
| `quiz_submissions` | 34,036 | Yes: `(id)`, idx on `(user_id)` |
| `challenges` | 13 | Yes: `(id)` |
| `quiz_sessions` | 212 | Yes: `(id)` |

---

## 7. Recommendations

### R1: Fix type mismatch in bulk completion/enrollment queries [HIGH IMPACT, LOW EFFORT]

**Impact: 27ms -> 1.8ms per query (15x faster), saves ~50ms per challenges page load**

Change `GetBulkUserCompletionTimestamps` in `challenge_completions.sql`:
```sql
-- FROM:
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::text[]), unnest(@challengeids::text[])
);

-- TO:
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::char(28)[]), unnest(@challengeids::char(28)[])
);
```

Same fix for `GetBulkUserEnrollmentTimestamps` in `challenge_enrollments.sql`.

After changing, run `make generate` to regenerate sqlc.

**Note:** This same `unnest + text[]` pattern may exist in other queries. A project-wide audit for `unnest(@...::text[])` on `char(28)` columns is recommended.

### R2: Audit all unnest patterns for the same type mismatch [MEDIUM IMPACT]

Search for all `unnest(@..::text[])` patterns in `backend/internal/database/queries/*.sql` and check if the target column uses `char(28)` (which is the case for all ULID-based IDs in this project). Any mismatch will cause the same seq scan behavior.

### R3: (Future) Consider scoping quiz submissions by quiz ID in the dataloader [LOW PRIORITY]

Currently `QuizSubmissionsByUserLoader` loads all submissions for a user regardless of quiz. If submission volume grows, consider a `QuizSubmissionsByUserAndQuizLoader` with key `(userID, quizID)`.

---

## 8. Expected Impact

| Change | Before | After | Improvement |
|--------|--------|-------|-------------|
| R1: Fix bulk completion query | 27ms | 1.8ms | 15x faster |
| R1: Fix bulk enrollment query | ~27ms | ~1.8ms | 15x faster |
| **Total challenges page** | **~65ms** | **~15ms** | **4x faster** |

The challenges page goes from ~65ms to ~15ms total DB time, with the remaining cost being the quiz session access check (5.6ms) and enrollment ID lookup (3.9ms).
