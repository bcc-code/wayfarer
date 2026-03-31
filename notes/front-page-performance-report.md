# Front Page Database Performance Report

**Date:** 2026-03-31
**Database:** Neon PostgreSQL (wayfarer)
**Target:** Reduce CPU load from front page queries

---

## 1. Executive Summary

The front page (`ProfilePage` query) spends **98% of its database time** on a single leaderboard query that takes **362ms**. The root cause is a correlated consent-check subquery executed 12,552 times (once per user). Denormalizing the consent status reduces the full leaderboard query to **21ms (17x faster)**, and using a rank-only query for the "me" entry brings it to **~5ms (75x faster)**.

---

## 2. Current Data Scale

| Table | Rows | Role |
|-------|------|------|
| `score_journal` | 484,492 | Point transactions |
| `user_achievements` | 75,618 | Achievement awards |
| `user_challenge_completions` | 52,209 | Challenge completions |
| `user_consent_history` | 19,979 | Consent tracking |
| `users` | 13,162 | All users |
| `leaderboard_project_persons` | 12,552 | Pre-computed scores |
| `user_projects` | 10,588 | Project memberships |
| `team_members` | 6,854 | Team memberships |
| `teams` | 790 | Teams |
| `churches` | 179 | Churches |
| `achievements` | 15 | Achievement definitions |
| `projects` | 1 | "Ladder to Heaven" |

---

## 3. What the Front Page Loads

The `ProfilePage` GraphQL query (`frontend/app/graphql/queries/pages/profile.gql`) requests:

```graphql
query ProfilePage($ageFilter: LeaderboardFilter) {
  me { id, name, consentStatus { pendingConsents { ... } } }
  myCurrentProject {
    id, name, infoMessage { ... }, branding { ... }
    achievements { id, name, points, hidden, achievedAt, celebratedAt, ... }
    myPoints
    leaderboard(entityType: PERSONS, filter: $ageFilter) {
      me { rank }
    }
    myTeam { superTeam { id, name, color, imageObject { ... } } }
  }
}
```

---

## 4. Query-by-Query Benchmark

All benchmarks run against production data on the Neon pooler endpoint.

### 4.1 Fast Queries (not a concern)

| Operation | Resolver Path | Measured | Why it's fast |
|-----------|--------------|----------|---------------|
| `me` | `UserByIDLoader` | <1ms | PK lookup on `users` |
| `myCurrentProject` | `ProjectByIDLoader` | <1ms | PK lookup on `projects` |
| `achievements` (15 rows) | `AchievementsByProjectLoader` | ~1ms | Dataloader batched, small result set |
| `achievedAt`/`celebratedAt` | `UserAchievementTimestampLoader` | ~1ms | Dataloader batched |
| `myPoints` | `GetUserScore` | **2.4ms** | Covering index `idx_score_journal_project_user` |
| `myTeam.superTeam` | `TeamsByUserLoader` | <1ms | Dataloader |

**GetUserScore EXPLAIN:**
```
Index Only Scan using idx_score_journal_project_user on score_journal
  Index Cond: (project_id = '...', user_id = '...')
  Heap Fetches: 32
  Execution Time: 2.382 ms
```

### 4.2 The Bottleneck: `leaderboard.me.rank` (362ms)

**Code path:**
1. `projects.resolvers.go:468` -> `LeaderboardService.GetProjectLeaderboard()`
2. `leaderboard.go:136` -> `getProjectPersonLeaderboard()`
3. Checks cache (key includes project + entity type + filter hash)
4. On miss: calls `GetFullProjectPersonLeaderboard` SQL -- **loads ALL 9,008 consented users**
5. `findMeInLeaderboard()` -- linear scan of full array to find current user
6. Caches result for 5 minutes

**The front page only needs `rank` (one integer) but loads 9,008 rows.**

**EXPLAIN output for `GetFullProjectPersonLeaderboard`:**
```
Sort  (actual time=361.111..361.526 rows=9008)
  Sort Key: rank, last_score_at DESC, name
  -> WindowAgg  (actual time=6.212..356.518 rows=9008)
    -> Nested Loop  (actual time=6.206..350.438 rows=9008)
      -> Nested Loop  (actual time=6.191..335.302 rows=9008)
        -> Index Only Scan on leaderboard_project_persons  (actual time=3.243..28.348 rows=12552)
        -> Index Scan on users  (actual time=0.024 rows=1 loops=12552)
          Filter: EXISTS(SubPlan 3)    <-- CONSENT CHECK
          SubPlan 3:
            -> Index Scan on user_consent_history  (loops=12552)
              InitPlan 2:
                -> Index Only Scan on user_consent_history  (loops=12552)
Execution Time: 362.306 ms
```

**Buffer analysis:**
| Component | Shared Buffers Hit | % of Total |
|-----------|-------------------|------------|
| Consent MAX(occurred_at) subquery | 51,484 | 33% |
| Consent action check | 27,274 | 17% |
| Users PK lookups | 116,414 | 75% |
| Leaderboard table scan | 12,579 | 8% |
| **Total** | **156,016 hits + 798 reads** | |

The consent correlated subquery alone accounts for **50% of the I/O** (78,758 buffer hits across 12,552 iterations).

---

## 5. Tested Alternatives

### 5.1 `FindMyProjectPersonPosition` (existing but unused)

This query exists in `leaderboards.sql:80-100` but the service layer never calls it. It skips the consent check entirely:

```sql
WITH ranked_scores AS (
    SELECT u.id AS entity_id, ..., lpp.score,
        DENSE_RANK() OVER (ORDER BY lpp.score DESC) AS rank
    FROM leaderboard_project_persons lpp
    INNER JOIN users u ON lpp.user_id = u.id
    INNER JOIN churches c ON u.church_id = c.id
    WHERE lpp.project_id = $1 AND lpp.score >= 1
)
SELECT * FROM ranked_scores WHERE entity_id = $2;
```

**Result: 27ms** (still computes DENSE_RANK for all 12,552 users)

### 5.2 COUNT DISTINCT rank calculation (no consent)

```sql
WITH my_score AS (
    SELECT score FROM leaderboard_project_persons
    WHERE project_id = $1 AND user_id = $2
)
SELECT (COUNT(DISTINCT lpp.score) + 1) AS rank
FROM leaderboard_project_persons lpp
WHERE lpp.project_id = $1
  AND lpp.score > (SELECT score FROM my_score)
  AND lpp.score >= 1;
```

**Result: 4.5ms** -- uses index scan on `leaderboard_project_persons_pkey`, no window function needed.

### 5.3 COUNT DISTINCT with consent CTE

Added consent filtering via CTE to the rank-only query:

```sql
WITH my_score AS (...),
consented_users AS (
    SELECT DISTINCT uch.user_id
    FROM user_consent_history uch
    WHERE uch.consent_key = 'leaderboard_consent'
      AND uch.action = 'ACCEPTED'
      AND uch.occurred_at = (SELECT MAX(...))
)
SELECT (COUNT(DISTINCT lpp.score) + 1) AS rank
FROM leaderboard_project_persons lpp
WHERE ... AND lpp.user_id IN (SELECT user_id FROM consented_users);
```

**Result: 211ms** -- consent CTE scans 9,251 consent history rows with correlated subquery.

### 5.4 COUNT DISTINCT with age filter (no consent)

```sql
WITH my_score AS (...)
SELECT (COUNT(DISTINCT lpp.score) + 1) AS rank
FROM leaderboard_project_persons lpp
INNER JOIN users u ON lpp.user_id = u.id
WHERE ... AND EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate) BETWEEN 15 AND 25;
```

**Result: 15ms** -- hash join with users table for age filter.

### 5.5 Denormalized consent on leaderboard table

Created temp table with `has_consent BOOLEAN` column pre-computed, indexed with partial index `WHERE has_consent = true`.

**Rank-only query:**
```sql
WITH my_score AS (...)
SELECT (COUNT(DISTINCT lpp.score) + 1) AS rank
FROM lpp_with_consent lpp
WHERE ... AND lpp.has_consent = true;
```

**Result: 4.8ms** -- bitmap index scan on the partial index.

**Full leaderboard query (for paginated views):**
```sql
SELECT u.id, ..., DENSE_RANK() OVER (ORDER BY lpp.score DESC) AS rank
FROM lpp_with_consent lpp
INNER JOIN users u ON lpp.user_id = u.id
INNER JOIN churches c ON u.church_id = c.id
WHERE ... AND lpp.has_consent = true
ORDER BY rank ASC, last_score_at DESC;
```

**Result: 21ms** -- hash joins, no correlated subqueries.

---

## 6. Results Summary

| Query Variant | Time | vs. Current |
|---------------|------|-------------|
| **Current: full leaderboard + consent subquery** | **362ms** | baseline |
| FindMyPosition (existing, unused, no consent) | 27ms | 13x faster |
| Rank-only COUNT DISTINCT (no consent) | 4.5ms | 80x faster |
| Rank-only COUNT DISTINCT (consent via CTE) | 211ms | 1.7x faster |
| Rank-only COUNT DISTINCT (age filter, no consent) | 15ms | 24x faster |
| **Rank-only + denormalized consent** | **4.8ms** | **75x faster** |
| **Full leaderboard + denormalized consent** | **21ms** | **17x faster** |

---

## 7. Recommendations

### R1: Denormalize `has_consent` onto `leaderboard_project_persons` [HIGH IMPACT]

**Impact: 362ms -> 21ms (full leaderboard), 4.8ms (rank-only)**

Add a `has_consent BOOLEAN DEFAULT false` column to `leaderboard_project_persons`. Update it via:
- A trigger on `user_consent_history` INSERT/UPDATE
- Backfill existing data in the migration

Add a partial index:
```sql
CREATE INDEX idx_lpp_project_score_consented
ON leaderboard_project_persons (project_id, score DESC)
WHERE has_consent = true;
```

Replace the consent EXISTS subquery in all `GetFull*Leaderboard` queries with `AND lpp.has_consent = true`.

**Why this works:** The consent check is currently a correlated subquery that runs once per row (12,552 times), each requiring 2 index lookups on `user_consent_history`. Moving it to a boolean eliminates all 25,104 index lookups.

**Risk:** Consent status could become stale if the trigger isn't set up correctly. Since consent changes are rare (users accept once), the risk is low. The trigger on `user_consent_history` makes it eventually consistent within the same transaction.

### R2: Use rank-only query for front page "me" entry [HIGH IMPACT]

**Impact: avoids loading 9,008 rows when only 1 integer is needed**

The front page only requests `leaderboard.me.rank`. The current code loads the entire leaderboard into memory (9,008 rows -> JSON marshal -> Ristretto cache -> JSON unmarshal -> linear scan).

Create a new SQL query `GetMyProjectPersonRank` using the COUNT DISTINCT approach:
```sql
WITH my_score AS (
    SELECT score FROM leaderboard_project_persons
    WHERE project_id = @projectid AND user_id = @userid
)
SELECT COALESCE(
    (SELECT COUNT(DISTINCT lpp.score) + 1
     FROM leaderboard_project_persons lpp
     WHERE lpp.project_id = @projectid
       AND lpp.score > (SELECT score FROM my_score)
       AND lpp.score >= 1
       AND lpp.has_consent = true),
    0
) AS rank;
```

In the service layer, add a method like `GetMyRank(ctx, projectID, userID)` that:
1. Checks a cache key `leaderboard:myrank:{projectID}:{userID}` (short TTL, ~2 min)
2. On miss, runs the rank-only query (~5ms)
3. Returns just the rank integer

The resolver for `leaderboard.me` can use this fast path when no pagination/edges are requested.

### R3: Denormalize `birth_year` onto `leaderboard_project_persons` [MEDIUM IMPACT]

**Impact: eliminates JOIN to users table for age-filtered leaderboards**

The age filter requires joining `users` to get `birthdate`. Adding `birth_year SMALLINT` to `leaderboard_project_persons` removes this join. Update the `update_person_leaderboard` PL/pgSQL function to populate it from `users.birthdate` when inserting/updating.

Filter becomes:
```sql
WHERE birth_year BETWEEN (EXTRACT(YEAR FROM CURRENT_DATE) - @maxage)
                    AND (EXTRACT(YEAR FROM CURRENT_DATE) - @minage)
```

### R4: Increase leaderboard cache TTL for unfiltered leaderboard [LOW EFFORT]

Current TTL is 5 minutes for all leaderboard cache entries. The unfiltered full leaderboard (used by paginated views) could use a 10-15 minute TTL since score changes are incremental. The rank-only cache (R2) can stay at 2 minutes for more responsive updates.

### R5: Consider deduplicating consent indexes [LOW EFFORT]

`user_consent_history` has 9 indexes, several with overlapping prefixes:
- `idx_user_consent_history_user` (user_id)
- `idx_user_consent_history_user_key` (user_id, consent_key)
- `idx_user_consent_history_user_key_time` (user_id, consent_key, occurred_at DESC)
- `idx_user_consent_history_user_key_action_time` (user_id, consent_key, action, occurred_at DESC)

The first two are strict prefixes of the latter two and could potentially be dropped, reducing write amplification on INSERT (each consent insert updates 9 indexes).

---

## 8. Expected Combined Impact

| Scenario | Before | After R1+R2 | Improvement |
|----------|--------|-------------|-------------|
| Front page (cache miss, no age filter) | 362ms | ~5ms | 98.6% reduction |
| Front page (cache miss, age filter) | 362ms+ | ~15ms | 95.8% reduction |
| Front page (cache hit) | <1ms | <1ms | Same |
| Full leaderboard page (cache miss) | 362ms | ~21ms | 94.2% reduction |
| DB CPU per front page request | ~156K buffer hits | ~200 buffer hits | 99.9% reduction |

---

## 9. Existing Indexes (Reference)

### `leaderboard_project_persons`
```sql
-- PK
CREATE UNIQUE INDEX leaderboard_project_persons_pkey ON leaderboard_project_persons (project_id, user_id);
-- Score ordering
CREATE INDEX idx_leaderboard_project_persons_score ON leaderboard_project_persons (project_id, score DESC, last_score_at DESC NULLS LAST, user_id);
```

### `score_journal`
```sql
CREATE INDEX idx_score_journal_project_user ON score_journal (project_id, user_id) INCLUDE (points);  -- covering index, used by GetUserScore
CREATE INDEX idx_score_journal_project ON score_journal (project_id);
CREATE INDEX idx_score_journal_user ON score_journal (user_id);
CREATE INDEX idx_score_journal_event ON score_journal (event_id);
CREATE INDEX idx_score_journal_challenge ON score_journal (challenge_id);
CREATE INDEX idx_score_journal_source ON score_journal (source_type, source_id);
CREATE INDEX idx_score_journal_time ON score_journal (created_at);
```

### `user_consent_history`
```sql
CREATE INDEX idx_user_consent_history_user ON user_consent_history (user_id);
CREATE INDEX idx_user_consent_history_user_key ON user_consent_history (user_id, consent_key);
CREATE INDEX idx_user_consent_history_user_key_time ON user_consent_history (user_id, consent_key, occurred_at DESC);
CREATE INDEX idx_user_consent_history_user_key_action_time ON user_consent_history (user_id, consent_key, action, occurred_at DESC);
-- + 5 more indexes
```

---

## 10. Leaderboard Update Mechanism

The `leaderboard_project_persons` table is updated via a PostgreSQL trigger on `score_journal`:

```sql
-- Trigger fires on INSERT/UPDATE/DELETE of score_journal
CREATE TRIGGER trigger_score_journal_leaderboard
  AFTER INSERT OR UPDATE OR DELETE ON score_journal
  FOR EACH ROW EXECUTE FUNCTION trigger_update_leaderboard_from_score_journal();
```

The `update_person_leaderboard` function uses UPSERT:
```sql
INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
VALUES (p_project_id, p_user_id, p_points_delta, NOW())
ON CONFLICT (project_id, user_id) DO UPDATE SET
    score = leaderboard_project_persons.score + p_points_delta,
    updated_at = NOW();
```

This is the natural place to also set `has_consent` and `birth_year` when implementing R1 and R3.
