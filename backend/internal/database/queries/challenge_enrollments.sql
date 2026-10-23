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

-- name: BulkEnrollUsersInChallenge :many
-- Returns the users that were newly enrolled; users already enrolled are
-- skipped and not returned, so callers notify only the new ones.
INSERT INTO user_challenge_enrollments (user_id, challenge_id)
SELECT unnest(@userids::text[]), @challengeid::text
ON CONFLICT (user_id, challenge_id) DO NOTHING
RETURNING user_id;

-- name: GetUsersOutsideChurchAdminScope :many
-- The users in @userids that a church admin of @churchids may not enroll in a
-- challenge of @projectid: users who are not members of the project, or who
-- are neither in one of those churches nor on a team in the project that has
-- a member from one of them.
SELECT t.user_id::text AS user_id
FROM unnest(@userids::text[]) AS t(user_id)
WHERE NOT EXISTS (
    SELECT 1
    FROM user_projects up
    JOIN users u ON u.id = up.user_id
    WHERE up.user_id = t.user_id
      AND up.project_id = @projectid::char(28)
      AND (
          u.church_id = ANY(@churchids::char(28)[])
          OR EXISTS (
              SELECT 1
              FROM team_members tm
              JOIN teams tt ON tt.id = tm.team_id
              JOIN team_members tm2 ON tm2.team_id = tt.id
              JOIN users u2 ON u2.id = tm2.user_id
              WHERE tm.user_id = t.user_id
                AND tt.project_id = @projectid::char(28)
                AND u2.church_id = ANY(@churchids::char(28)[])
          )
      )
);

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

-- name: GetUsersEnrolledChallengeIDsInProject :many
SELECT uce.user_id, uce.challenge_id
FROM user_challenge_enrollments uce
JOIN challenges c ON c.id = uce.challenge_id
WHERE uce.user_id = ANY(@userids::char(28)[])
  AND c.project_id = @projectid::char(28);
