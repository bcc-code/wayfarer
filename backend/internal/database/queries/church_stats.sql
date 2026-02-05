-- name: GetChurchAgeGroupStats :many
-- Get average scores grouped by age for users who are in teams and belong to a specific church
WITH church_team_users AS (
    SELECT DISTINCT
        u.id AS user_id,
        u.birthdate,
        EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate) AS age
    FROM users u
    INNER JOIN team_members tm ON u.id = tm.user_id
    INNER JOIN teams t ON tm.team_id = t.id
    WHERE u.church_id = @churchid::text
      AND t.project_id = @projectid::text
      AND u.birthdate IS NOT NULL
),
user_scores AS (
    SELECT
        ctu.user_id,
        ctu.age,
        COALESCE(SUM(sj.points), 0)::bigint AS total_score
    FROM church_team_users ctu
    LEFT JOIN score_journal sj ON sj.user_id = ctu.user_id AND sj.project_id = @projectid::text
    GROUP BY ctu.user_id, ctu.age
),
age_groups AS (
    SELECT
        us.user_id,
        us.total_score,
        CASE
            WHEN us.age >= 12 AND us.age <= 18 THEN '12 - 18'
            WHEN us.age >= 19 AND us.age <= 25 THEN '19 - 25'
            WHEN us.age >= 26 AND us.age <= 36 THEN '26 - 36'
            WHEN us.age >= 37 AND us.age <= 59 THEN '37 - 59'
            WHEN us.age >= 60 THEN '60+'
            ELSE NULL
        END AS age_group
    FROM user_scores us
)
SELECT
    age_group,
    COUNT(*)::int AS user_count,
    COALESCE(AVG(total_score), 0)::float8 AS average_score
FROM age_groups
WHERE age_group IS NOT NULL
GROUP BY age_group
ORDER BY
    CASE age_group
        WHEN '12 - 18' THEN 1
        WHEN '19 - 25' THEN 2
        WHEN '26 - 36' THEN 3
        WHEN '37 - 59' THEN 4
        WHEN '60+' THEN 5
    END;

-- name: CountChurchUsersInTeams :one
-- Count total users from a church who are in teams for a specific project
SELECT COUNT(DISTINCT u.id)::int AS total_users
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
INNER JOIN teams t ON tm.team_id = t.id
WHERE u.church_id = @churchid::text
  AND t.project_id = @projectid::text;

-- name: GetChurchUserScores :many
-- Get scores for all users from a church who are in teams for a specific project
SELECT
    u.id AS user_id,
    COALESCE(u.display_name, u.name) AS name,
    COALESCE(SUM(sj.points), 0)::bigint AS total_score
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
INNER JOIN teams t ON tm.team_id = t.id
LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.project_id = @projectid::text
WHERE u.church_id = @churchid::text
  AND t.project_id = @projectid::text
GROUP BY u.id, u.display_name, u.name
ORDER BY total_score ASC;
