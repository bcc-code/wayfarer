-- Enrollment Operations (idempotent with ON CONFLICT DO NOTHING)

-- name: EnrollUserInChallenge :one
INSERT INTO user_challenge_enrollments (user_id, challenge_id)
VALUES (@userid::text, @challengeid::text)
ON CONFLICT (user_id, challenge_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING enrolled_at;

-- name: UnenrollUserFromChallenge :exec
DELETE FROM user_challenge_enrollments
WHERE user_id = @userid::char(28) AND challenge_id = @challengeid::char(28);

-- name: IsUserEnrolledInChallenge :one
SELECT EXISTS(
    SELECT 1
    FROM user_challenge_enrollments
    WHERE user_id = @userid::char(28) AND challenge_id = @challengeid::char(28)
) AS is_enrolled;

-- name: GetUserEnrollmentTimestamp :one
SELECT enrolled_at
FROM user_challenge_enrollments
WHERE user_id = @userid::char(28) AND challenge_id = @challengeid::char(28);

-- name: GetEnrolledUsersForChallenge :many
SELECT user_id, enrolled_at
FROM user_challenge_enrollments
WHERE challenge_id = @challengeid::char(28)
ORDER BY enrolled_at ASC;

-- Batch operations for dataloader

-- name: GetUserEnrollmentTimestamps :many
SELECT challenge_id, enrolled_at
FROM user_challenge_enrollments
WHERE user_id = @userid::char(28)
  AND challenge_id = ANY(@challengeids::char(28)[]);

-- name: GetBulkUserEnrollmentTimestamps :many
SELECT user_id, challenge_id, enrolled_at
FROM user_challenge_enrollments
WHERE (user_id, challenge_id) IN (
    SELECT unnest(@userids::char(28)[]), unnest(@challengeids::char(28)[])
);

-- Bulk enrollment for admin/M2M

-- name: BulkEnrollUsersInChallenge :exec
INSERT INTO user_challenge_enrollments (user_id, challenge_id)
SELECT unnest(@userids::text[]), @challengeid::text
ON CONFLICT (user_id, challenge_id) DO NOTHING;

-- name: BulkUnenrollUsersFromChallenge :exec
DELETE FROM user_challenge_enrollments
WHERE challenge_id = @challengeid::char(28)
  AND user_id = ANY(@userids::char(28)[]);

-- name: GetUserEnrolledChallengeIDsInProject :many
SELECT uce.challenge_id
FROM user_challenge_enrollments uce
JOIN challenges c ON c.id = uce.challenge_id
WHERE uce.user_id = @userid::char(28)
  AND c.project_id = @projectid::char(28);
