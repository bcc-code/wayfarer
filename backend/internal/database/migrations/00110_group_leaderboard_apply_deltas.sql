-- +goose Up
-- Group a drain batch's deltas per (user, project, event) before applying.
-- Each apply_score_journal_delta call fans out to ~38 statements (person,
-- church, teams, superteams, with member-count recalculation). Bet
-- settlement queues ~10 deltas per user (stake + winnings per bet), so after
-- a 10k-user betting session the one-delta-per-row drain left the
-- leaderboards ~2 minutes behind. Deltas are additive and last_score_at
-- takes the latest timestamp, so a group's sum (applied once, with its
-- latest score_at) ends in the same state as applying each row.
-- Still returns the number of queue rows consumed (the worker keeps
-- draining while a full batch comes back).

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION drain_leaderboard_apply_queue(batch_size INT)
RETURNS INT AS $$
DECLARE
    applied INT;
BEGIN
    WITH batch AS (
        DELETE FROM leaderboard_apply_queue
        WHERE id IN (
            SELECT id FROM leaderboard_apply_queue
            ORDER BY id
            LIMIT batch_size
            FOR UPDATE SKIP LOCKED
        )
        RETURNING user_id, project_id, event_id, points_delta, score_at
    ), grouped AS (
        SELECT user_id, project_id, event_id,
               SUM(points_delta)::BIGINT AS points_delta,
               MAX(score_at) AS score_at,
               count(*)::INT AS n
        FROM batch
        GROUP BY user_id, project_id, event_id
    )
    SELECT COALESCE(SUM(n), 0)::INT INTO applied
    FROM (
        -- Volatile, so evaluated for every group even though only n is used
        SELECT n, apply_score_journal_delta(user_id, project_id, event_id, points_delta, score_at)
        FROM grouped
        ORDER BY user_id
    ) applied_groups;
    RETURN applied;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION drain_leaderboard_apply_queue(batch_size INT)
RETURNS INT AS $$
DECLARE
    applied INT;
BEGIN
    WITH batch AS (
        DELETE FROM leaderboard_apply_queue
        WHERE id IN (
            SELECT id FROM leaderboard_apply_queue
            ORDER BY id
            LIMIT batch_size
            FOR UPDATE SKIP LOCKED
        )
        RETURNING user_id, project_id, event_id, points_delta, score_at
    )
    SELECT count(*) INTO applied
    FROM (
        SELECT apply_score_journal_delta(user_id, project_id, event_id, points_delta, score_at)
        FROM batch
    ) applied_rows;
    RETURN applied;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
