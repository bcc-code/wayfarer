-- ==================== Maintenance Queries ====================
-- These queries support administrative maintenance operations

-- name: GetMissingContentProgressPreview :many
-- Get users with external_content_events that should have user_content_progress but don't.
-- Each row shows a user ID and how many missing events they have.
-- The User object is resolved via DataLoader.
SELECT
    u.id AS user_id,
    COUNT(DISTINCT ece.id)::int AS event_count
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
LEFT JOIN user_content_progress ucp ON
    ucp.user_id = u.id
    AND ucp.achievement_id = cai.achievement_id
    AND ucp.external_content_id = cai.external_content_id
WHERE ucp.user_id IS NULL
GROUP BY u.id
ORDER BY event_count DESC, u.id ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN 50 ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: CountMissingContentProgressUsers :one
-- Count how many distinct users are affected by missing content progress.
SELECT COUNT(DISTINCT u.id)
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
LEFT JOIN user_content_progress ucp ON
    ucp.user_id = u.id
    AND ucp.achievement_id = cai.achievement_id
    AND ucp.external_content_id = cai.external_content_id
WHERE ucp.user_id IS NULL;

-- name: CountMissingContentProgressEvents :one
-- Count how many external_content_events have missing user_content_progress.
SELECT COUNT(*)
FROM (
    SELECT DISTINCT ece.id
    FROM external_content_events ece
    INNER JOIN users u ON u.person_uuid = ece.person_id
    INNER JOIN external_content ec ON ec.task_id = ece.task_id
    INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
    LEFT JOIN user_content_progress ucp ON
        ucp.user_id = u.id
        AND ucp.achievement_id = cai.achievement_id
        AND ucp.external_content_id = cai.external_content_id
    WHERE ucp.user_id IS NULL
) AS missing;

-- name: GetMissingContentEventsForProcessing :many
-- Get all external_content_events that have a user but are missing user_content_progress.
-- Returns user_id and task_id so they can be processed via ContentAchievementService.
SELECT DISTINCT
    u.id AS user_id,
    ece.task_id
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
LEFT JOIN user_content_progress ucp ON
    ucp.user_id = u.id
    AND ucp.achievement_id = cai.achievement_id
    AND ucp.external_content_id = cai.external_content_id
WHERE ucp.user_id IS NULL;

-- name: GetMissingContentProgressUserIDs :many
-- Get distinct user IDs that have missing content progress.
-- Used to batch users into separate jobs.
SELECT DISTINCT u.id AS user_id
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
LEFT JOIN user_content_progress ucp ON
    ucp.user_id = u.id
    AND ucp.achievement_id = cai.achievement_id
    AND ucp.external_content_id = cai.external_content_id
WHERE ucp.user_id IS NULL
ORDER BY u.id ASC;

-- name: GetMissingContentEventsForUsers :many
-- Get events for specific users that are missing content progress.
-- Used by batched job processing.
SELECT DISTINCT
    u.id AS user_id,
    ece.task_id
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
LEFT JOIN user_content_progress ucp ON
    ucp.user_id = u.id
    AND ucp.achievement_id = cai.achievement_id
    AND ucp.external_content_id = cai.external_content_id
WHERE ucp.user_id IS NULL
AND u.id = ANY(@userids::text[]);

-- ==================== Missing Score Journal Queries ====================
-- These queries find external_content_events that are missing score_journal entries.
-- The score_journal.source_id references external_content.id for content achievements.

-- name: GetMissingScoreJournalPreview :many
-- Get users with external_content_events for an achievement that are missing score_journal entries.
-- Each row shows a user ID and how many events are missing score journal entries.
SELECT
    u.id AS user_id,
    COUNT(DISTINCT ece.id)::int AS event_count
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
WHERE cai.achievement_id = @achievementid::text
AND NOT EXISTS (
    SELECT 1
    FROM score_journal sj
    WHERE sj.user_id = u.id
    AND sj.source_id = ec.id
)
GROUP BY u.id
ORDER BY event_count DESC, u.id ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN 50 ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: CountMissingScoreJournalUsers :one
-- Count how many distinct users are affected by missing score journal entries for an achievement.
SELECT COUNT(DISTINCT u.id)
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
WHERE cai.achievement_id = @achievementid::text
AND NOT EXISTS (
    SELECT 1
    FROM score_journal sj
    WHERE sj.user_id = u.id
    AND sj.source_id = ec.id
);

-- name: CountMissingScoreJournalEvents :one
-- Count how many external_content_events are missing score journal entries for an achievement.
SELECT COUNT(*)
FROM (
    SELECT DISTINCT ece.id
    FROM external_content_events ece
    INNER JOIN users u ON u.person_uuid = ece.person_id
    INNER JOIN external_content ec ON ec.task_id = ece.task_id
    INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
    WHERE cai.achievement_id = @achievementid::text
    AND NOT EXISTS (
        SELECT 1
        FROM score_journal sj
        WHERE sj.user_id = u.id
        AND sj.source_id = ec.id
    )
) AS missing;

-- name: GetMissingScoreJournalUserIDs :many
-- Get distinct user IDs that have missing score journal entries for an achievement.
-- Used to batch users into separate jobs.
SELECT DISTINCT u.id AS user_id
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
WHERE cai.achievement_id = @achievementid::text
AND NOT EXISTS (
    SELECT 1
    FROM score_journal sj
    WHERE sj.user_id = u.id
    AND sj.source_id = ec.id
)
ORDER BY u.id ASC;

-- name: GetMissingScoreJournalForUsers :many
-- Get external_content_ids for specific users that are missing score journal entries.
-- Used by batched job processing.
SELECT DISTINCT
    u.id AS user_id,
    ec.id AS external_content_id
FROM external_content_events ece
INNER JOIN users u ON u.person_uuid = ece.person_id
INNER JOIN external_content ec ON ec.task_id = ece.task_id
INNER JOIN content_achievement_items cai ON cai.external_content_id = ec.id
WHERE cai.achievement_id = @achievementid::text
AND NOT EXISTS (
    SELECT 1
    FROM score_journal sj
    WHERE sj.user_id = u.id
    AND sj.source_id = ec.id
)
AND u.id = ANY(@userids::text[]);
