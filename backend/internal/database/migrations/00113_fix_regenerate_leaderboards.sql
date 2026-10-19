-- +goose Up
-- +goose StatementBegin

-- Make the full rebuild agree with the incremental apply
-- (apply_score_journal_delta), which it disagreed with in two ways:
-- 1. SuperTeams: a user in several teams of one superteam had their points
--    summed once per team. The incremental apply counts them once.
-- 2. Rows whose points sum to zero or below were dropped (HAVING SUM > 0).
--    The incremental apply keeps them, and with betting a net loss is a real
--    score (a lost stake), not missing data.

CREATE OR REPLACE FUNCTION regenerate_leaderboards()
RETURNS TABLE (
    table_name TEXT,
    records_created BIGINT
) AS $$
DECLARE
    project_persons_count BIGINT;
    project_teams_count BIGINT;
    project_superteams_count BIGINT;
    project_churches_count BIGINT;
    event_persons_count BIGINT;
    event_teams_count BIGINT;
    event_superteams_count BIGINT;
    event_churches_count BIGINT;
BEGIN
    -- Truncate all leaderboard tables
    TRUNCATE TABLE leaderboard_project_persons;
    TRUNCATE TABLE leaderboard_project_teams;
    TRUNCATE TABLE leaderboard_project_superteams;
    TRUNCATE TABLE leaderboard_project_churches;
    TRUNCATE TABLE leaderboard_event_persons;
    TRUNCATE TABLE leaderboard_event_teams;
    TRUNCATE TABLE leaderboard_event_superteams;
    TRUNCATE TABLE leaderboard_event_churches;

    -- ==================== Regenerate Project Person Leaderboards ====================
    -- Persons still use SUM (no averaging)
    INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at, last_score_at)
    SELECT
        sj.project_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    GROUP BY sj.project_id, sj.user_id;
    GET DIAGNOSTICS project_persons_count = ROW_COUNT;

    -- ==================== Regenerate Project Team Leaderboards ====================
    -- Teams use AVERAGE: total_points / member_count
    INSERT INTO leaderboard_project_teams (project_id, team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        t.project_id,
        t.id AS team_id,
        COALESCE(SUM(sj.points), 0) AS total_points,
        (SELECT COUNT(*)::INT FROM team_members tm WHERE tm.team_id = t.id) AS member_count,
        calculate_average_score(
            COALESCE(SUM(sj.points), 0),
            (SELECT COUNT(*)::INT FROM team_members tm WHERE tm.team_id = t.id)
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM teams t
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = t.project_id
    GROUP BY t.project_id, t.id;
    GET DIAGNOSTICS project_teams_count = ROW_COUNT;

    -- ==================== Regenerate Project SuperTeam Leaderboards ====================
    -- SuperTeams use AVERAGE: total_points / distinct member_count
    INSERT INTO leaderboard_project_superteams (project_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        st.project_id,
        st.id AS super_team_id,
        COALESCE(SUM(sj.points), 0) AS total_points,
        (
            SELECT COUNT(DISTINCT tm2.user_id)::INT
            FROM team_members tm2
            INNER JOIN teams t2 ON t2.id = tm2.team_id
            WHERE t2.super_team_id = st.id AND t2.project_id = st.project_id
        ) AS member_count,
        calculate_average_score(
            COALESCE(SUM(sj.points), 0),
            (
                SELECT COUNT(DISTINCT tm2.user_id)::INT
                FROM team_members tm2
                INNER JOIN teams t2 ON t2.id = tm2.team_id
                WHERE t2.super_team_id = st.id AND t2.project_id = st.project_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM super_teams st
    -- Each member once, however many of the superteam's teams they are in
    INNER JOIN (
        SELECT DISTINCT t.super_team_id, t.project_id, tm.user_id
        FROM teams t
        INNER JOIN team_members tm ON tm.team_id = t.id
    ) stm ON stm.super_team_id = st.id AND stm.project_id = st.project_id
    INNER JOIN score_journal sj ON sj.user_id = stm.user_id AND sj.project_id = st.project_id
    GROUP BY st.project_id, st.id;
    GET DIAGNOSTICS project_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Project Church Leaderboards ====================
    -- Churches use AVERAGE: total_points / member_count (users in project with this church)
    INSERT INTO leaderboard_project_churches (project_id, church_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.project_id,
        u.church_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT u2.id)::INT
            FROM users u2
            INNER JOIN user_projects up2 ON up2.user_id = u2.id AND up2.project_id = sj.project_id
            WHERE u2.church_id = u.church_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT u2.id)::INT
                FROM users u2
                INNER JOIN user_projects up2 ON up2.user_id = u2.id AND up2.project_id = sj.project_id
                WHERE u2.church_id = u.church_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE u.church_id IS NOT NULL
    GROUP BY sj.project_id, u.church_id;
    GET DIAGNOSTICS project_churches_count = ROW_COUNT;

    -- ==================== Regenerate Event Person Leaderboards ====================
    -- Persons still use SUM (no averaging)
    INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, sj.user_id;
    GET DIAGNOSTICS event_persons_count = ROW_COUNT;

    -- ==================== Regenerate Event Team Leaderboards ====================
    -- Teams use AVERAGE: total_points / member_count (team members in event)
    INSERT INTO leaderboard_event_teams (event_id, team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        tm.team_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT tm2.user_id)::INT
            FROM team_members tm2
            INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
            WHERE tm2.team_id = tm.team_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT tm2.user_id)::INT
                FROM team_members tm2
                INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
                WHERE tm2.team_id = tm.team_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN team_members tm ON tm.user_id = sj.user_id
    INNER JOIN teams t ON t.id = tm.team_id
    INNER JOIN events e ON e.id = sj.event_id AND e.project_id = t.project_id
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, tm.team_id;
    GET DIAGNOSTICS event_teams_count = ROW_COUNT;

    -- ==================== Regenerate Event SuperTeam Leaderboards ====================
    -- SuperTeams use AVERAGE: total_points / distinct member_count (in event)
    INSERT INTO leaderboard_event_superteams (event_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        stm.super_team_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT tm2.user_id)::INT
            FROM team_members tm2
            INNER JOIN teams t2 ON t2.id = tm2.team_id
            INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
            WHERE t2.super_team_id = stm.super_team_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT tm2.user_id)::INT
                FROM team_members tm2
                INNER JOIN teams t2 ON t2.id = tm2.team_id
                INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
                WHERE t2.super_team_id = stm.super_team_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN events e ON e.id = sj.event_id
    -- Each member once, however many of the superteam's teams they are in
    INNER JOIN (
        SELECT DISTINCT t.super_team_id, t.project_id, tm.user_id
        FROM teams t
        INNER JOIN team_members tm ON tm.team_id = t.id
        WHERE t.super_team_id IS NOT NULL
    ) stm ON stm.user_id = sj.user_id AND stm.project_id = e.project_id
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, stm.super_team_id;
    GET DIAGNOSTICS event_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Event Church Leaderboards ====================
    -- Churches use AVERAGE: total_points / member_count (users in event with this church)
    INSERT INTO leaderboard_event_churches (event_id, church_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        u.church_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT u2.id)::INT
            FROM users u2
            INNER JOIN user_events ue2 ON ue2.user_id = u2.id AND ue2.event_id = sj.event_id
            WHERE u2.church_id = u.church_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT u2.id)::INT
                FROM users u2
                INNER JOIN user_events ue2 ON ue2.user_id = u2.id AND ue2.event_id = sj.event_id
                WHERE u2.church_id = u.church_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE sj.event_id IS NOT NULL
    AND u.church_id IS NOT NULL
    GROUP BY sj.event_id, u.church_id;
    GET DIAGNOSTICS event_churches_count = ROW_COUNT;

    -- Return summary
    RETURN QUERY
    SELECT 'leaderboard_project_persons'::TEXT, project_persons_count
    UNION ALL SELECT 'leaderboard_project_teams'::TEXT, project_teams_count
    UNION ALL SELECT 'leaderboard_project_superteams'::TEXT, project_superteams_count
    UNION ALL SELECT 'leaderboard_project_churches'::TEXT, project_churches_count
    UNION ALL SELECT 'leaderboard_event_persons'::TEXT, event_persons_count
    UNION ALL SELECT 'leaderboard_event_teams'::TEXT, event_teams_count
    UNION ALL SELECT 'leaderboard_event_superteams'::TEXT, event_superteams_count
    UNION ALL SELECT 'leaderboard_event_churches'::TEXT, event_churches_count;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the function from migration 00080
CREATE OR REPLACE FUNCTION regenerate_leaderboards()
RETURNS TABLE (
    table_name TEXT,
    records_created BIGINT
) AS $$
DECLARE
    project_persons_count BIGINT;
    project_teams_count BIGINT;
    project_superteams_count BIGINT;
    project_churches_count BIGINT;
    event_persons_count BIGINT;
    event_teams_count BIGINT;
    event_superteams_count BIGINT;
    event_churches_count BIGINT;
BEGIN
    -- Truncate all leaderboard tables
    TRUNCATE TABLE leaderboard_project_persons;
    TRUNCATE TABLE leaderboard_project_teams;
    TRUNCATE TABLE leaderboard_project_superteams;
    TRUNCATE TABLE leaderboard_project_churches;
    TRUNCATE TABLE leaderboard_event_persons;
    TRUNCATE TABLE leaderboard_event_teams;
    TRUNCATE TABLE leaderboard_event_superteams;
    TRUNCATE TABLE leaderboard_event_churches;

    -- ==================== Regenerate Project Person Leaderboards ====================
    -- Persons still use SUM (no averaging)
    INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at, last_score_at)
    SELECT
        sj.project_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    GROUP BY sj.project_id, sj.user_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS project_persons_count = ROW_COUNT;

    -- ==================== Regenerate Project Team Leaderboards ====================
    -- Teams use AVERAGE: total_points / member_count
    INSERT INTO leaderboard_project_teams (project_id, team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        t.project_id,
        t.id AS team_id,
        COALESCE(SUM(sj.points), 0) AS total_points,
        (SELECT COUNT(*)::INT FROM team_members tm WHERE tm.team_id = t.id) AS member_count,
        calculate_average_score(
            COALESCE(SUM(sj.points), 0),
            (SELECT COUNT(*)::INT FROM team_members tm WHERE tm.team_id = t.id)
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM teams t
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = t.project_id
    GROUP BY t.project_id, t.id
    HAVING COALESCE(SUM(sj.points), 0) > 0;
    GET DIAGNOSTICS project_teams_count = ROW_COUNT;

    -- ==================== Regenerate Project SuperTeam Leaderboards ====================
    -- SuperTeams use AVERAGE: total_points / distinct member_count
    INSERT INTO leaderboard_project_superteams (project_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        st.project_id,
        st.id AS super_team_id,
        COALESCE(SUM(sj.points), 0) AS total_points,
        (
            SELECT COUNT(DISTINCT tm2.user_id)::INT
            FROM team_members tm2
            INNER JOIN teams t2 ON t2.id = tm2.team_id
            WHERE t2.super_team_id = st.id AND t2.project_id = st.project_id
        ) AS member_count,
        calculate_average_score(
            COALESCE(SUM(sj.points), 0),
            (
                SELECT COUNT(DISTINCT tm2.user_id)::INT
                FROM team_members tm2
                INNER JOIN teams t2 ON t2.id = tm2.team_id
                WHERE t2.super_team_id = st.id AND t2.project_id = st.project_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM super_teams st
    INNER JOIN teams t ON t.super_team_id = st.id AND t.project_id = st.project_id
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = st.project_id
    GROUP BY st.project_id, st.id
    HAVING COALESCE(SUM(sj.points), 0) > 0;
    GET DIAGNOSTICS project_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Project Church Leaderboards ====================
    -- Churches use AVERAGE: total_points / member_count (users in project with this church)
    INSERT INTO leaderboard_project_churches (project_id, church_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.project_id,
        u.church_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT u2.id)::INT
            FROM users u2
            INNER JOIN user_projects up2 ON up2.user_id = u2.id AND up2.project_id = sj.project_id
            WHERE u2.church_id = u.church_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT u2.id)::INT
                FROM users u2
                INNER JOIN user_projects up2 ON up2.user_id = u2.id AND up2.project_id = sj.project_id
                WHERE u2.church_id = u.church_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE u.church_id IS NOT NULL
    GROUP BY sj.project_id, u.church_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS project_churches_count = ROW_COUNT;

    -- ==================== Regenerate Event Person Leaderboards ====================
    -- Persons still use SUM (no averaging)
    INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, sj.user_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_persons_count = ROW_COUNT;

    -- ==================== Regenerate Event Team Leaderboards ====================
    -- Teams use AVERAGE: total_points / member_count (team members in event)
    INSERT INTO leaderboard_event_teams (event_id, team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        tm.team_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT tm2.user_id)::INT
            FROM team_members tm2
            INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
            WHERE tm2.team_id = tm.team_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT tm2.user_id)::INT
                FROM team_members tm2
                INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
                WHERE tm2.team_id = tm.team_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN team_members tm ON tm.user_id = sj.user_id
    INNER JOIN teams t ON t.id = tm.team_id
    INNER JOIN events e ON e.id = sj.event_id AND e.project_id = t.project_id
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, tm.team_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_teams_count = ROW_COUNT;

    -- ==================== Regenerate Event SuperTeam Leaderboards ====================
    -- SuperTeams use AVERAGE: total_points / distinct member_count (in event)
    INSERT INTO leaderboard_event_superteams (event_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        t.super_team_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT tm2.user_id)::INT
            FROM team_members tm2
            INNER JOIN teams t2 ON t2.id = tm2.team_id
            INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
            WHERE t2.super_team_id = t.super_team_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT tm2.user_id)::INT
                FROM team_members tm2
                INNER JOIN teams t2 ON t2.id = tm2.team_id
                INNER JOIN user_events ue2 ON ue2.user_id = tm2.user_id AND ue2.event_id = sj.event_id
                WHERE t2.super_team_id = t.super_team_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN team_members tm ON tm.user_id = sj.user_id
    INNER JOIN teams t ON t.id = tm.team_id
    INNER JOIN events e ON e.id = sj.event_id AND e.project_id = t.project_id
    WHERE sj.event_id IS NOT NULL
    AND t.super_team_id IS NOT NULL
    GROUP BY sj.event_id, t.super_team_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Event Church Leaderboards ====================
    -- Churches use AVERAGE: total_points / member_count (users in event with this church)
    INSERT INTO leaderboard_event_churches (event_id, church_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        u.church_id,
        SUM(sj.points) AS total_points,
        (
            SELECT COUNT(DISTINCT u2.id)::INT
            FROM users u2
            INNER JOIN user_events ue2 ON ue2.user_id = u2.id AND ue2.event_id = sj.event_id
            WHERE u2.church_id = u.church_id
        ) AS member_count,
        calculate_average_score(
            SUM(sj.points),
            (
                SELECT COUNT(DISTINCT u2.id)::INT
                FROM users u2
                INNER JOIN user_events ue2 ON ue2.user_id = u2.id AND ue2.event_id = sj.event_id
                WHERE u2.church_id = u.church_id
            )
        ) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE sj.event_id IS NOT NULL
    AND u.church_id IS NOT NULL
    GROUP BY sj.event_id, u.church_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_churches_count = ROW_COUNT;

    -- Return summary
    RETURN QUERY
    SELECT 'leaderboard_project_persons'::TEXT, project_persons_count
    UNION ALL SELECT 'leaderboard_project_teams'::TEXT, project_teams_count
    UNION ALL SELECT 'leaderboard_project_superteams'::TEXT, project_superteams_count
    UNION ALL SELECT 'leaderboard_project_churches'::TEXT, project_churches_count
    UNION ALL SELECT 'leaderboard_event_persons'::TEXT, event_persons_count
    UNION ALL SELECT 'leaderboard_event_teams'::TEXT, event_teams_count
    UNION ALL SELECT 'leaderboard_event_superteams'::TEXT, event_superteams_count
    UNION ALL SELECT 'leaderboard_event_churches'::TEXT, event_churches_count;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd
