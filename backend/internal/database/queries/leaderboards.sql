-- ==================== Project Person Leaderboard ====================

-- name: GetProjectPersonLeaderboard :many
WITH project_users AS MATERIALIZED (
    -- Start with users in THIS project (huge performance win)
    -- MATERIALIZED forces execution order to avoid scanning all users
    SELECT DISTINCT u.id, u.name, u.avatar_url, u.birthdate, u.church_id, u.gender, c.name AS church_name
    FROM user_projects up
    INNER JOIN users u ON up.user_id = u.id
    INNER JOIN churches c ON u.church_id = c.id
    WHERE up.project_id = @projectid::text
      AND (@churchid::text = '' OR u.church_id = @churchid::text)
      AND (@gender::text = '' OR u.gender = @gender::text)
      AND (@teamid::text = '' OR EXISTS (
          SELECT 1 FROM team_members tm
          WHERE tm.user_id = u.id AND tm.team_id = @teamid::text
      ))
      AND (@superteamid::text = '' OR EXISTS (
          SELECT 1 FROM team_members tm
          INNER JOIN teams t ON tm.team_id = t.id
          WHERE tm.user_id = u.id AND t.super_team_id = @superteamid::text
      ))
),
filtered_users AS MATERIALIZED (
    -- Apply age and church filters
    SELECT pu.id, pu.name, pu.avatar_url, pu.church_name
    FROM project_users pu
    WHERE (@minage::int IS NULL OR DATE_PART('year', AGE(pu.birthdate)) >= @minage::int)
      AND (@maxage::int IS NULL OR DATE_PART('year', AGE(pu.birthdate)) <= @maxage::int)
      AND (@country::text = '' OR EXISTS (
          SELECT 1 FROM churches c
          WHERE c.id = pu.church_id AND c.country = @country::text
      ))
      AND (@churchcategory::text = '' OR EXISTS (
          SELECT 1 FROM churches c
          WHERE c.id = pu.church_id AND c.category = @churchcategory::text
      ))
),
person_scores AS (
    SELECT
        fu.id AS entity_id,
        fu.name,
        fu.church_name,
        fu.avatar_url AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM filtered_users fu
    LEFT JOIN score_journal sj ON sj.user_id = fu.id AND sj.project_id = @projectid::text
    GROUP BY fu.id, fu.name, fu.church_name, fu.avatar_url
    HAVING COALESCE(SUM(sj.points), 0) >= 1
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        church_name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM person_scores
)
SELECT entity_id, name, church_name, image, score, rank
FROM ranked_scores
WHERE
    (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyProjectPersonPosition :one
WITH ranked_scores AS (
    SELECT
        u.id AS entity_id,
        u.name,
        c.name AS church_name,
        u.avatar_url AS image,
        lpp.score,
        RANK() OVER (ORDER BY lpp.score DESC, u.name ASC) AS rank
    FROM leaderboard_project_persons lpp
    INNER JOIN users u ON lpp.user_id = u.id
    INNER JOIN churches c ON u.church_id = c.id
    WHERE lpp.project_id = @projectid::text
      AND lpp.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpp.score <= @maxscore::int)
      AND (@churchid::text = '' OR u.church_id = @churchid::text)
)
SELECT entity_id, name, church_name, image, score, rank
FROM ranked_scores
WHERE entity_id = @userid::text;

-- name: GetFullProjectPersonLeaderboard :many
WITH ranked_scores AS (
    SELECT
        u.id AS entity_id,
        u.name,
        c.name AS church_name,
        u.avatar_url AS image,
        lpp.score,
        RANK() OVER (ORDER BY lpp.score DESC, u.name ASC) AS rank
    FROM leaderboard_project_persons lpp
    INNER JOIN users u ON lpp.user_id = u.id
    INNER JOIN churches c ON u.church_id = c.id
    WHERE lpp.project_id = @projectid::text
      AND lpp.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpp.score <= @maxscore::int)
      AND (@churchid::text = '' OR u.church_id = @churchid::text)
)
SELECT entity_id, name, church_name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountProjectPersonLeaderboard :one
WITH project_users AS MATERIALIZED (
    -- Start with users in THIS project
    SELECT DISTINCT u.id, u.birthdate, u.church_id, u.gender
    FROM user_projects up
    INNER JOIN users u ON up.user_id = u.id
    WHERE up.project_id = @projectid::text
      AND (@churchid::text = '' OR u.church_id = @churchid::text)
      AND (@gender::text = '' OR u.gender = @gender::text)
      AND (@teamid::text = '' OR EXISTS (
          SELECT 1 FROM team_members tm
          WHERE tm.user_id = u.id AND tm.team_id = @teamid::text
      ))
      AND (@superteamid::text = '' OR EXISTS (
          SELECT 1 FROM team_members tm
          INNER JOIN teams t ON tm.team_id = t.id
          WHERE tm.user_id = u.id AND t.super_team_id = @superteamid::text
      ))
)
SELECT COUNT(*)::bigint AS total
FROM project_users pu
WHERE (@minage::int IS NULL OR DATE_PART('year', AGE(pu.birthdate)) >= @minage::int)
  AND (@maxage::int IS NULL OR DATE_PART('year', AGE(pu.birthdate)) <= @maxage::int)
  AND (@country::text = '' OR EXISTS (
      SELECT 1 FROM churches c
      WHERE c.id = pu.church_id AND c.country = @country::text
  ))
  AND (@churchcategory::text = '' OR EXISTS (
      SELECT 1 FROM churches c
      WHERE c.id = pu.church_id AND c.category = @churchcategory::text
  ));

-- ==================== Project Team Leaderboard ====================

-- name: GetProjectTeamLeaderboard :many
WITH team_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM teams t
    INNER JOIN team_members tm ON t.id = tm.team_id
    INNER JOIN users u ON tm.user_id = u.id
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.project_id = @projectid::text
    WHERE
        t.project_id = @projectid::text
        AND (@superteamid::text = '' OR t.super_team_id = @superteamid::text)
    GROUP BY t.id, t.name
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM team_scores
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyProjectTeamPosition :one
WITH ranked_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        lpt.score,
        RANK() OVER (ORDER BY lpt.score DESC, t.name ASC) AS rank
    FROM leaderboard_project_teams lpt
    INNER JOIN teams t ON lpt.team_id = t.id
    WHERE lpt.project_id = @projectid::text
      AND lpt.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpt.score <= @maxscore::int)
),
user_team AS (
    SELECT team_id
    FROM team_members
    WHERE user_id = @userid::text
      AND team_id IN (SELECT id FROM teams WHERE project_id = @projectid::text)
    LIMIT 1
)
SELECT rs.entity_id, rs.name, rs.image, rs.score, rs.rank
FROM ranked_scores rs
INNER JOIN user_team ut ON rs.entity_id = ut.team_id;

-- name: GetFullProjectTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        lpt.score,
        RANK() OVER (ORDER BY lpt.score DESC, t.name ASC) AS rank
    FROM leaderboard_project_teams lpt
    INNER JOIN teams t ON lpt.team_id = t.id
    WHERE lpt.project_id = @projectid::text
      AND lpt.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpt.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountProjectTeamLeaderboard :one
SELECT COUNT(DISTINCT t.id)::bigint AS total
FROM teams t
WHERE
    t.project_id = @projectid::text
    AND (@superteamid::text = '' OR t.super_team_id = @superteamid::text);

-- ==================== Project SuperTeam Leaderboard ====================

-- name: GetProjectSuperTeamLeaderboard :many
WITH superteam_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM super_teams st
    INNER JOIN teams t ON t.super_team_id = st.id
    INNER JOIN team_members tm ON t.id = tm.team_id
    INNER JOIN users u ON tm.user_id = u.id
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.project_id = @projectid::text
    WHERE
        st.project_id = @projectid::text
    GROUP BY st.id, st.name
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM superteam_scores
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyProjectSuperTeamPosition :one
WITH ranked_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        lps.score,
        RANK() OVER (ORDER BY lps.score DESC, st.name ASC) AS rank
    FROM leaderboard_project_superteams lps
    INNER JOIN super_teams st ON lps.super_team_id = st.id
    WHERE lps.project_id = @projectid::text
      AND lps.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lps.score <= @maxscore::int)
),
user_superteam AS (
    SELECT t.super_team_id
    FROM team_members tm
    INNER JOIN teams t ON tm.team_id = t.id
    WHERE tm.user_id = @userid::text
      AND t.super_team_id IS NOT NULL
      AND t.project_id = @projectid::text
    LIMIT 1
)
SELECT rs.entity_id, rs.name, rs.image, rs.score, rs.rank
FROM ranked_scores rs
INNER JOIN user_superteam ust ON rs.entity_id = ust.super_team_id;

-- name: GetFullProjectSuperTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        lpst.score,
        RANK() OVER (ORDER BY lpst.score DESC, st.name ASC) AS rank
    FROM leaderboard_project_superteams lpst
    INNER JOIN super_teams st ON lpst.super_team_id = st.id
    WHERE lpst.project_id = @projectid::text
      AND lpst.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpst.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountProjectSuperTeamLeaderboard :one
SELECT COUNT(DISTINCT st.id)::bigint AS total
FROM super_teams st
WHERE st.project_id = @projectid::text;

-- ==================== Project Church Leaderboard ====================

-- name: GetProjectChurchLeaderboard :many
WITH church_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM churches c
    INNER JOIN users u ON c.id = u.church_id
    INNER JOIN user_projects up ON u.id = up.user_id
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.project_id = @projectid::text
    WHERE
        up.project_id = @projectid::text
        AND (@country::text = '' OR c.country = @country::text)
        AND (@churchcategory::text = '' OR c.category = @churchcategory::text)
    GROUP BY c.id, c.name
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM church_scores
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyProjectChurchPosition :one
WITH ranked_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        lpc.score,
        RANK() OVER (ORDER BY lpc.score DESC, c.name ASC) AS rank
    FROM leaderboard_project_churches lpc
    INNER JOIN churches c ON lpc.church_id = c.id
    WHERE lpc.project_id = @projectid::text
      AND lpc.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpc.score <= @maxscore::int)
),
user_church AS (
    SELECT church_id
    FROM users
    WHERE id = @userid::text
)
SELECT rs.entity_id, rs.name, rs.image, rs.score, rs.rank
FROM ranked_scores rs
INNER JOIN user_church uc ON rs.entity_id = uc.church_id;

-- name: GetFullProjectChurchLeaderboard :many
WITH ranked_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        lpc.score,
        RANK() OVER (ORDER BY lpc.score DESC, c.name ASC) AS rank
    FROM leaderboard_project_churches lpc
    INNER JOIN churches c ON lpc.church_id = c.id
    WHERE lpc.project_id = @projectid::text
      AND lpc.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lpc.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountProjectChurchLeaderboard :one
SELECT COUNT(DISTINCT c.id)::bigint AS total
FROM churches c
INNER JOIN users u ON c.id = u.church_id
INNER JOIN user_projects up ON u.id = up.user_id
WHERE
    up.project_id = @projectid::text
    AND (@country::text = '' OR c.country = @country::text)
    AND (@churchcategory::text = '' OR c.category = @churchcategory::text);

-- ==================== Event Person Leaderboard ====================

-- name: GetEventPersonLeaderboard :many
WITH person_scores AS (
    SELECT
        u.id AS entity_id,
        u.name,
        c.name AS church_name,
        u.avatar_url AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM users u
    INNER JOIN user_events ue ON u.id = ue.user_id
    INNER JOIN churches c ON u.church_id = c.id
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.event_id = @eventid::text
    WHERE
        ue.event_id = @eventid::text
        AND (@churchid::text = '' OR u.church_id = @churchid::text)
        AND (@country::text = '' OR EXISTS (
            SELECT 1 FROM churches ch WHERE ch.id = u.church_id AND ch.country = @country::text
        ))
        AND (@churchcategory::text = '' OR EXISTS (
            SELECT 1 FROM churches ch WHERE ch.id = u.church_id AND ch.category = @churchcategory::text
        ))
        AND (@gender::text = '' OR u.gender = @gender::text)
        AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
        AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    GROUP BY u.id, u.name, c.name, u.avatar_url
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        church_name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM person_scores
)
SELECT entity_id, name, church_name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyEventPersonPosition :one
WITH ranked_scores AS (
    SELECT
        u.id AS entity_id,
        u.name,
        c.name AS church_name,
        u.avatar_url AS image,
        lep.score,
        RANK() OVER (ORDER BY lep.score DESC, u.name ASC) AS rank
    FROM leaderboard_event_persons lep
    INNER JOIN users u ON lep.user_id = u.id
    INNER JOIN churches c ON u.church_id = c.id
    WHERE lep.event_id = @eventid::text
      AND lep.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lep.score <= @maxscore::int)
      AND (@churchid::text = '' OR u.church_id = @churchid::text)
)
SELECT entity_id, name, church_name, image, score, rank
FROM ranked_scores
WHERE entity_id = @userid::text;

-- name: GetFullEventPersonLeaderboard :many
WITH ranked_scores AS (
    SELECT
        u.id AS entity_id,
        u.name,
        c.name AS church_name,
        u.avatar_url AS image,
        lep.score,
        RANK() OVER (ORDER BY lep.score DESC, u.name ASC) AS rank
    FROM leaderboard_event_persons lep
    INNER JOIN users u ON lep.user_id = u.id
    INNER JOIN churches c ON u.church_id = c.id
    WHERE lep.event_id = @eventid::text
      AND lep.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lep.score <= @maxscore::int)
      AND (@churchid::text = '' OR u.church_id = @churchid::text)
)
SELECT entity_id, name, church_name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountEventPersonLeaderboard :one
SELECT COUNT(DISTINCT u.id)::bigint AS total
FROM users u
INNER JOIN user_events ue ON u.id = ue.user_id
WHERE
    ue.event_id = @eventid::text
    AND (@churchid::text = '' OR u.church_id = @churchid::text)
    AND (@country::text = '' OR EXISTS (
        SELECT 1 FROM churches c WHERE c.id = u.church_id AND c.country = @country::text
    ))
    AND (@churchcategory::text = '' OR EXISTS (
        SELECT 1 FROM churches c WHERE c.id = u.church_id AND c.category = @churchcategory::text
    ))
    AND (@gender::text = '' OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int);

-- ==================== Event Team Leaderboard ====================

-- name: GetEventTeamLeaderboard :many
WITH event_project AS (
    SELECT project_id FROM events WHERE id = @eventid::text
),
team_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM teams t
    CROSS JOIN event_project ep
    INNER JOIN team_members tm ON t.id = tm.team_id
    INNER JOIN users u ON tm.user_id = u.id
    INNER JOIN user_events ue ON u.id = ue.user_id AND ue.event_id = @eventid::text
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.event_id = @eventid::text
    WHERE
        t.project_id = ep.project_id
    GROUP BY t.id, t.name
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM team_scores
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyEventTeamPosition :one
WITH ranked_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        let.score,
        RANK() OVER (ORDER BY let.score DESC, t.name ASC) AS rank
    FROM leaderboard_event_teams let
    INNER JOIN teams t ON let.team_id = t.id
    WHERE let.event_id = @eventid::text
      AND let.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR let.score <= @maxscore::int)
),
user_team AS (
    SELECT tm.team_id
    FROM team_members tm
    INNER JOIN teams t ON tm.team_id = t.id
    INNER JOIN events e ON t.project_id = e.project_id AND e.id = @eventid::text
    WHERE tm.user_id = @userid::text
    LIMIT 1
)
SELECT rs.entity_id, rs.name, rs.image, rs.score, rs.rank
FROM ranked_scores rs
INNER JOIN user_team ut ON rs.entity_id = ut.team_id;

-- name: GetFullEventTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        t.id AS entity_id,
        t.name,
        NULL::text AS image,
        let.score,
        RANK() OVER (ORDER BY let.score DESC, t.name ASC) AS rank
    FROM leaderboard_event_teams let
    INNER JOIN teams t ON let.team_id = t.id
    WHERE let.event_id = @eventid::text
      AND let.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR let.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountEventTeamLeaderboard :one
WITH event_project AS (
    SELECT project_id FROM events WHERE id = @eventid::text
)
SELECT COUNT(DISTINCT t.id)::bigint AS total
FROM teams t
CROSS JOIN event_project ep
WHERE t.project_id = ep.project_id;

-- ==================== Event SuperTeam Leaderboard ====================

-- name: GetEventSuperTeamLeaderboard :many
WITH event_project AS (
    SELECT project_id FROM events WHERE id = @eventid::text
),
superteam_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM super_teams st
    CROSS JOIN event_project ep
    INNER JOIN teams t ON t.super_team_id = st.id
    INNER JOIN team_members tm ON t.id = tm.team_id
    INNER JOIN users u ON tm.user_id = u.id
    INNER JOIN user_events ue ON u.id = ue.user_id AND ue.event_id = @eventid::text
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.event_id = @eventid::text
    WHERE
        st.project_id = ep.project_id
    GROUP BY st.id, st.name
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM superteam_scores
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyEventSuperTeamPosition :one
WITH ranked_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        les.score,
        RANK() OVER (ORDER BY les.score DESC, st.name ASC) AS rank
    FROM leaderboard_event_superteams les
    INNER JOIN super_teams st ON les.super_team_id = st.id
    WHERE les.event_id = @eventid::text
      AND les.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR les.score <= @maxscore::int)
),
user_superteam AS (
    SELECT t.super_team_id
    FROM team_members tm
    INNER JOIN teams t ON tm.team_id = t.id
    INNER JOIN events e ON t.project_id = e.project_id AND e.id = @eventid::text
    WHERE tm.user_id = @userid::text
      AND t.super_team_id IS NOT NULL
    LIMIT 1
)
SELECT rs.entity_id, rs.name, rs.image, rs.score, rs.rank
FROM ranked_scores rs
INNER JOIN user_superteam ust ON rs.entity_id = ust.super_team_id;

-- name: GetFullEventSuperTeamLeaderboard :many
WITH ranked_scores AS (
    SELECT
        st.id AS entity_id,
        st.name,
        NULL::text AS image,
        lest.score,
        RANK() OVER (ORDER BY lest.score DESC, st.name ASC) AS rank
    FROM leaderboard_event_superteams lest
    INNER JOIN super_teams st ON lest.super_team_id = st.id
    WHERE lest.event_id = @eventid::text
      AND lest.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lest.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountEventSuperTeamLeaderboard :one
WITH event_project AS (
    SELECT project_id FROM events WHERE id = @eventid::text
)
SELECT COUNT(DISTINCT st.id)::bigint AS total
FROM super_teams st
CROSS JOIN event_project ep
WHERE st.project_id = ep.project_id;

-- ==================== Event Church Leaderboard ====================

-- name: GetEventChurchLeaderboard :many
WITH church_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        COALESCE(SUM(sj.points), 0)::bigint AS score
    FROM churches c
    INNER JOIN users u ON c.id = u.church_id
    INNER JOIN user_events ue ON u.id = ue.user_id
    LEFT JOIN score_journal sj ON sj.user_id = u.id AND sj.event_id = @eventid::text
    WHERE
        ue.event_id = @eventid::text
        AND (@country::text = '' OR c.country = @country::text)
        AND (@churchcategory::text = '' OR c.category = @churchcategory::text)
    GROUP BY c.id, c.name
),
ranked_scores AS (
    SELECT
        entity_id,
        name,
        image,
        score,
        RANK() OVER (ORDER BY score DESC, name ASC) AS rank
    FROM church_scores
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
WHERE
    score >= 1
    AND (@minscore::int IS NULL OR score >= @minscore::int)
    AND (@maxscore::int IS NULL OR score <= @maxscore::int)
    AND (sqlc.narg('afterrank')::bigint IS NULL OR rank > sqlc.narg('afterrank')::bigint)
    AND (sqlc.narg('beforerank')::bigint IS NULL OR rank < sqlc.narg('beforerank')::bigint)
ORDER BY rank ASC
LIMIT @querylimit::int;

-- name: FindMyEventChurchPosition :one
WITH ranked_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        lec.score,
        RANK() OVER (ORDER BY lec.score DESC, c.name ASC) AS rank
    FROM leaderboard_event_churches lec
    INNER JOIN churches c ON lec.church_id = c.id
    WHERE lec.event_id = @eventid::text
      AND lec.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lec.score <= @maxscore::int)
),
user_church AS (
    SELECT church_id
    FROM users
    WHERE id = @userid::text
)
SELECT rs.entity_id, rs.name, rs.image, rs.score, rs.rank
FROM ranked_scores rs
INNER JOIN user_church uc ON rs.entity_id = uc.church_id;

-- name: GetFullEventChurchLeaderboard :many
WITH ranked_scores AS (
    SELECT
        c.id AS entity_id,
        c.name,
        NULL::text AS image,
        lec.score,
        RANK() OVER (ORDER BY lec.score DESC, c.name ASC) AS rank
    FROM leaderboard_event_churches lec
    INNER JOIN churches c ON lec.church_id = c.id
    WHERE lec.event_id = @eventid::text
      AND lec.score >= COALESCE(@minscore::int, 1)
      AND (@maxscore::int IS NULL OR lec.score <= @maxscore::int)
)
SELECT entity_id, name, image, score, rank
FROM ranked_scores
ORDER BY rank ASC;

-- name: CountEventChurchLeaderboard :one
SELECT COUNT(DISTINCT c.id)::bigint AS total
FROM churches c
INNER JOIN users u ON c.id = u.church_id
INNER JOIN user_events ue ON u.id = ue.user_id
WHERE
    ue.event_id = @eventid::text
    AND (@country::text = '' OR c.country = @country::text)
    AND (@churchcategory::text = '' OR c.category = @churchcategory::text);
