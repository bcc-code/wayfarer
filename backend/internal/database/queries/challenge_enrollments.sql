-- Enrollment Operations (idempotent with ON CONFLICT DO NOTHING)

-- name: EnrollUserInChallenge :one
INSERT INTO user_challenge_enrollments (user_id, challenge_id)
VALUES (@userid::text, @challengeid::text)
ON CONFLICT (user_id, challenge_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING enrolled_at;

-- name: UnenrollUserFromChallenge :exec
DELETE FROM user_challenge_enrollments
WHERE user_id = @userid::text AND challenge_id = @challengeid::text;

-- name: IsUserEnrolledInChallenge :one
SELECT EXISTS(
    SELECT 1
    FROM user_challenge_enrollments
    WHERE user_id = @userid::text AND challenge_id = @challengeid::text
) AS is_enrolled;

-- name: GetUserEnrollmentTimestamp :one
SELECT enrolled_at
FROM user_challenge_enrollments
WHERE user_id = @userid::text AND challenge_id = @challengeid::text;

-- name: GetEnrolledUsersForChallenge :many
SELECT user_id, enrolled_at
FROM user_challenge_enrollments
WHERE challenge_id = @challengeid::text
ORDER BY enrolled_at ASC;

-- Batch operations for dataloader

-- name: GetUserEnrollmentTimestamps :many
SELECT challenge_id, enrolled_at
FROM user_challenge_enrollments
WHERE user_id = @userid::text
  AND challenge_id = ANY(@challengeids::text[]);

-- name: GetBulkUserEnrollmentTimestamps :many
SELECT user_id, challenge_id, enrolled_at
FROM user_challenge_enrollments
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::text[]), unnest(@challengeids::text[])
);

-- Bulk enrollment for admin/M2M

-- name: BulkEnrollUsersInChallenge :exec
INSERT INTO user_challenge_enrollments (user_id, challenge_id)
SELECT unnest(@userids::text[]), @challengeid::text
ON CONFLICT (user_id, challenge_id) DO NOTHING;

-- name: BulkUnenrollUsersFromChallenge :exec
DELETE FROM user_challenge_enrollments
WHERE challenge_id = @challengeid::text
  AND user_id = ANY(@userids::text[]);

-- name: GetUserEnrolledChallengeIDsInProject :many
SELECT uce.challenge_id
FROM user_challenge_enrollments uce
JOIN challenges c ON c.id = uce.challenge_id
WHERE uce.user_id = @userid::text
  AND c.project_id = @projectid::text;
