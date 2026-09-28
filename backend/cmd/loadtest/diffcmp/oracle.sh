#!/bin/bash
# DB-state oracle for the A/B differential run: computes deterministic
# projections of the core tables in both databases and diffs them as text.
# Timestamps and runtime ULID primary keys are excluded from projections;
# multiset equality is asserted via ordered md5 digests.
#
# Usage: oracle.sh <db_a> <db_b> [label]
set -euo pipefail

DB_A="${1:?usage: oracle.sh <db_a> <db_b> [label]}"
DB_B="${2:?usage: oracle.sh <db_a> <db_b> [label]}"
LABEL="${3:-oracle}"
export PGPASSWORD="${PGPASSWORD:-bench}"
PGUSER="${PGUSER:-bench}"
PGHOST="${PGHOST:-127.0.0.1}"

SQLFILE=$(mktemp)
trap 'rm -f "$SQLFILE"' EXIT
cat > "$SQLFILE" <<'SQL'
-- Session-local normalizer: runtime ULIDs (2024+ timestamps start 01J/01K)
-- collapse to RUNTIME so per-side generated ids don't fail the digest; seeded
-- mkid() hex ids (00-prefixed) and *LOADTEST* fixture ids (01L…, L is not in
-- the Crockford alphabet) pass through.
CREATE FUNCTION pg_temp.norm(t text) RETURNS text AS
  $f$ SELECT regexp_replace(t, '01[HJKMNPQRSTVWXYZ][0123456789ABCDEFGHJKMNPQRSTVWXYZ]{23}$', 'RUNTIME') $f$ LANGUAGE sql;
SELECT 'score_journal' AS t, count(*) AS n, coalesce(sum(points),0) AS sum,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT user_id || ':' || points || ':' || source_type || ':' || coalesce(event_id,'-') || ':' || pg_temp.norm(coalesce(challenge_id,'-')) AS row_text, points FROM score_journal) s;

SELECT 'user_challenge_enrollments' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT user_id || ':' || pg_temp.norm(challenge_id) AS row_text FROM user_challenge_enrollments) s;

SELECT 'user_challenge_completions' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT user_id || ':' || pg_temp.norm(challenge_id) AS row_text FROM user_challenge_completions) s;

SELECT 'quiz_submissions' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT user_id || ':' || pg_temp.norm(quiz_id) || ':' || coalesce(score::text,'-') || ':' || coalesce(max_score::text,'-')
             || ':' || coalesce(points_awarded::text,'-') || ':' || (completed_at IS NOT NULL)::text AS row_text
      FROM quiz_submissions) s;

SELECT 'quiz_responses' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT pg_temp.norm(question_id) || ':' || coalesce(text_response,'-') || ':' || coalesce(number_response::text,'-')
             || ':' || coalesce(is_correct::text,'-') || ':' || coalesce(bet_amount::text,'-')
             || ':' || coalesce(points_earned::text,'-') AS row_text
      FROM quiz_responses) s;

SELECT 'quiz_session_access' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT pg_temp.norm(session_id) || ':' || user_id || ':' || source_type AS row_text FROM quiz_session_access) s;

SELECT 'teams(names)' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT id || ':' || name AS row_text FROM teams) s;

SELECT 'user_achievements' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT user_id || ':' || achievement_id AS row_text FROM user_achievements) s;

SELECT 'lb_project_persons' AS t, count(*) AS n, coalesce(sum(score),0) AS sum,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT user_id || ':' || score AS row_text, score FROM leaderboard_project_persons) s;

SELECT 'lb_project_teams' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT team_id || ':' || score || ':' || total_points || ':' || member_count AS row_text
      FROM leaderboard_project_teams) s;

SELECT 'lb_project_churches' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT church_id || ':' || score || ':' || total_points || ':' || member_count AS row_text
      FROM leaderboard_project_churches) s;

SELECT 'lb_project_superteams' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT super_team_id || ':' || score || ':' || total_points || ':' || member_count AS row_text
      FROM leaderboard_project_superteams) s;

SELECT 'challenges' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT name || ':' || challenge_type || ':' || (published_at IS NOT NULL)::text AS row_text FROM challenges) s;

SELECT 'quizzes' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT name || ':' || completion_points AS row_text FROM quizzes) s;

SELECT 'projects(info)' AS t, count(*) AS n,
       md5(coalesce(string_agg(row_text, '|' ORDER BY row_text), '')) AS digest
FROM (SELECT id || ':' || name || ':' || coalesce(info_message,'-') AS row_text FROM projects) s;
SQL

OUT_A=$(psql -h "$PGHOST" -U "$PGUSER" -d "$DB_A" -At -v ON_ERROR_STOP=1 -f "$SQLFILE")
OUT_B=$(psql -h "$PGHOST" -U "$PGUSER" -d "$DB_B" -At -v ON_ERROR_STOP=1 -f "$SQLFILE")

echo "--- $LABEL: DB oracle ($DB_A vs $DB_B) ---"
if DIFF=$(diff <(echo "$OUT_A") <(echo "$OUT_B")); then
    echo "IDENTICAL ($(echo "$OUT_A" | wc -l | tr -d ' ') projections)"
    exit 0
else
    echo "DIVERGED:"
    echo "$DIFF"
    exit 1
fi
