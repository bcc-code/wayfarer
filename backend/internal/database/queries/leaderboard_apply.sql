-- name: DrainLeaderboardApplyQueue :one
-- Applies up to batch_size queued score deltas to the leaderboard tables
-- (see migration 00101) and returns how many were applied.
SELECT drain_leaderboard_apply_queue(@batch_size::int) AS applied;

-- name: CountLeaderboardApplyQueue :one
SELECT count(*) AS backlog FROM leaderboard_apply_queue;
