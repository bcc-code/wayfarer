-- Challenge Completion Operations

-- name: CompleteUserChallenge :one
INSERT INTO user_challenge_completions (user_id, challenge_id, completed_at)
VALUES (@userid::text, @challengeid::text, COALESCE(@completedat, now()))
ON CONFLICT (user_id, challenge_id) DO UPDATE SET completed_at = COALESCE(@completedat, now())
RETURNING completed_at;

-- name: UncompleteUserFromChallenge :exec
DELETE FROM user_challenge_completions
WHERE user_id = @userid::text AND challenge_id = @challengeid::text;

-- name: IsUserChallengeCompleted :one
SELECT EXISTS(
    SELECT 1
    FROM user_challenge_completions
    WHERE user_id = @userid::text AND challenge_id = @challengeid::text
) AS is_completed;

-- name: GetUserCompletionTimestamp :one
SELECT completed_at
FROM user_challenge_completions
WHERE user_id = @userid::text AND challenge_id = @challengeid::text;

-- name: GetCompletedUsersForChallenge :many
SELECT user_id, completed_at
FROM user_challenge_completions
WHERE challenge_id = @challengeid::text
ORDER BY completed_at ASC;

-- name: BulkCompleteChallenges :exec
INSERT INTO user_challenge_completions (user_id, challenge_id, completed_at)
SELECT unnest(@userids::text[]) as user_id, @challengeid::text as challenge_id, COALESCE(@completedat, now()) as completed_at
ON CONFLICT (user_id, challenge_id) DO NOTHING;
