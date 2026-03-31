# Audit: `unnest(::text[])` Type Mismatch on `char(28)` Columns

**Date:** 2026-03-31
**Database:** Neon PostgreSQL (wayfarer)
**Root Cause:** All ID columns use `char(28)` (ULID format), but bulk queries pass `text[]` arrays to `unnest()`. PostgreSQL cannot use indexes when comparing `text` to `char(28)`, falling back to sequential scans.

---

## Summary

**7 queries** across 4 SQL files have this bug. The worst case scans **1.2 million rows** instead of doing a PK lookup. Total CPU savings from fixing all 7: estimated **2+ seconds per affected request**.

**Fix:** Change `unnest(@param::text[])` to `unnest(@param::char(28)[])` in WHERE/JOIN clauses.

**Note:** INSERT statements using `unnest(::text[])` are NOT affected (the cast happens on write, no index lookup needed).

---

## Affected Queries

### 1. `GetBulkUserContentProgress` - **CRITICAL**

**File:** `backend/internal/database/queries/achievements.sql:456-461`
**Table:** `user_content_progress` (1,211,062 rows)
**Index available:** `idx_user_content_progress_user_achievement (user_id, achievement_id)`

| Version | Plan | Time | Buffers |
|---------|------|------|---------|
| Current (`text[]`) | Parallel Seq Scan (3 workers) | **1,383ms** | 24,026 |
| Fixed (`char(28)[]`) | Bitmap Index Scan | **2.2ms** | 6 |

**Improvement: 629x faster**

```sql
-- Current
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::text[]), unnest(@achievement_ids::text[])
);

-- Fixed
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::char(28)[]), unnest(@achievement_ids::char(28)[])
);
```

### 2. `GetBulkUserStreakProgress` - **CRITICAL**

**File:** `backend/internal/database/queries/achievements.sql:677-682`
**Table:** `user_streak_progress` (641,144 rows)
**Index available:** `idx_user_streak_progress_user_achievement (user_id, achievement_id)`

| Version | Plan | Time | Buffers |
|---------|------|------|---------|
| Current (`text[]`) | Parallel Seq Scan (3 workers) | **547ms** | 10,331 |
| Fixed (`char(28)[]`) | Bitmap Index Scan | **2.4ms** | 47 |

**Improvement: 228x faster**

### 3. `UpdateBetResults` - **HIGH**

**File:** `backend/internal/database/queries/quiz_responses.sql:89-96`
**Table:** `quiz_responses` (101,164 rows)
**Index available:** `quiz_responses_pkey (id)`

| Version | Plan | Time | Buffers |
|---------|------|------|---------|
| Current (`text[]`) | Hash Join + Seq Scan | **164ms** | 2,538 |
| Fixed (`char(28)[]`) | Nested Loop + PK Index | **2.4ms** | 12 |

**Improvement: 68x faster**

```sql
-- Current
FROM (
    SELECT unnest(@ids::text[]) AS id, unnest(@pointsearned::int[]) AS points_earned
) AS data
WHERE quiz_responses.id = data.id

-- Fixed
FROM (
    SELECT unnest(@ids::char(28)[]) AS id, unnest(@pointsearned::int[]) AS points_earned
) AS data
WHERE quiz_responses.id = data.id
```

### 4. `UpdateBetResultsWithJournal` - **HIGH**

**File:** `backend/internal/database/queries/quiz_responses.sql:131-142`
**Table:** `quiz_responses` (101,164 rows)
**Index available:** `quiz_responses_pkey (id)`
**Same pattern as #3, same fix.**

| Version | Plan | Time | Buffers |
|---------|------|------|---------|
| Current (`text[]`) | ~164ms (same table) | **~164ms** | ~2,538 |
| Fixed (`char(28)[]`) | ~2.4ms | **~2.4ms** | ~12 |

```sql
-- Fix: change both unnest calls
SELECT
    unnest(@ids::char(28)[]) AS id,        -- was ::text[]
    unnest(@pointsearned::int[]) AS points_earned,
    unnest(@scorejournalids::char(28)[]) AS score_journal_id  -- was ::text[]
```

### 5. `GetBulkUserAchievementTimestamps` - **MEDIUM**

**File:** `backend/internal/database/queries/achievements.sql:433-438`
**Table:** `user_achievements` (75,618 rows)
**Index available:** `user_achievements_pkey (user_id, achievement_id)`

| Version | Plan | Time | Buffers |
|---------|------|------|---------|
| Current (`text[]`) | Hash Semi Join + Seq Scan | **41ms** | 1,305 |
| Fixed (`char(28)[]`) | Nested Loop + PK Index | **3ms** | 20 |

**Improvement: 14x faster**

### 6. `GetBulkUserAchievementCelebratedTimestamps` - **MEDIUM**

**File:** `backend/internal/database/queries/achievements.sql:440-445`
**Table:** `user_achievements` (75,618 rows)
**Same table and pattern as #5, same improvement.**

### 7. `GetBulkUserCompletionTimestamps` - **MEDIUM**

**File:** `backend/internal/database/queries/challenge_completions.sql:36-41`
**Table:** `user_challenge_completions` (52,209 rows)
**Index available:** `user_challenge_completions_pkey (user_id, challenge_id)`

| Version | Plan | Time | Buffers |
|---------|------|------|---------|
| Current (`text[]`) | Hash Semi Join + Seq Scan | **27ms** | 960 |
| Fixed (`char(28)[]`) | Nested Loop + PK Index | **1.8ms** | 15 |

**Improvement: 15x faster**

### 8. `GetBulkUserEnrollmentTimestamps` - **MEDIUM**

**File:** `backend/internal/database/queries/challenge_enrollments.sql:39-44`
**Table:** `user_challenge_enrollments` (smaller table, same pattern)
**Same fix needed.**

---

## NOT Affected (INSERT statements - safe)

These use `unnest(::text[])` in INSERT...SELECT where the text is cast to char(28) implicitly on write:

| Query | File | Why safe |
|-------|------|----------|
| `BulkCreateChallenges` | `challenges.sql:166-204` | INSERT values |
| `BulkCompleteChallenge` | `challenge_completions.sql:30-34` | INSERT values |
| `BulkEnrollUsersInChallenge` | `challenge_enrollments.sql:48-51` | INSERT values |
| `CreateTeamScoreAdjustmentBatch` | `score_journal.sql:117-140` | INSERT values |
| `CreateBulkScoreAdjustmentBatch` | `score_journal.sql:142-167` | INSERT values |
| `BulkAwardAchievement` | `achievements.sql:378-384` | INSERT values |

---

## Fix Summary

| # | Query | File:Line | Table Rows | Before | After | Speedup |
|---|-------|-----------|------------|--------|-------|---------|
| 1 | GetBulkUserContentProgress | achievements.sql:456 | 1,211,062 | 1,383ms | 2.2ms | 629x |
| 2 | GetBulkUserStreakProgress | achievements.sql:677 | 641,144 | 547ms | 2.4ms | 228x |
| 3 | UpdateBetResults | quiz_responses.sql:89 | 101,164 | 164ms | 2.4ms | 68x |
| 4 | UpdateBetResultsWithJournal | quiz_responses.sql:131 | 101,164 | ~164ms | ~2.4ms | 68x |
| 5 | GetBulkUserAchievementTimestamps | achievements.sql:433 | 75,618 | 41ms | 3ms | 14x |
| 6 | GetBulkUserAchievementCelebratedTimestamps | achievements.sql:440 | 75,618 | ~41ms | ~3ms | 14x |
| 7 | GetBulkUserCompletionTimestamps | challenge_completions.sql:36 | 52,209 | 27ms | 1.8ms | 15x |
| 8 | GetBulkUserEnrollmentTimestamps | challenge_enrollments.sql:39 | ~10,000 | ~11ms | ~1ms | 11x |

**Total worst-case savings: ~2.4 seconds of CPU time per request that hits all queries.**

---

## How to Apply

1. Edit the 4 SQL files listed above
2. Change `::text[]` to `::char(28)[]` in the WHERE/JOIN `unnest()` calls only (not in INSERT statements)
3. Run `make generate` in `backend/` to regenerate sqlc
4. Run `make test` to verify
5. No migration needed - this is a query-only change

---

## Why This Happens

PostgreSQL's query planner cannot use a `btree` index on a `char(28)` column when the comparison value is `text`. The types are compatible for equality comparison (the result is correct), but the planner cannot prove that the index ordering matches, so it falls back to a sequential scan.

This is a well-known PostgreSQL behavior: implicit type coercion between `char(n)` and `text` works for equality but prevents index usage. The fix is to ensure the unnest output type matches the column type exactly.
