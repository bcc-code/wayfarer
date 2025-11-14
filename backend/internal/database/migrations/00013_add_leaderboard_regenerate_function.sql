-- +goose Up
-- +goose StatementBegin

-- Function to regenerate all leaderboard tables from scratch
-- Usage: SELECT * FROM regenerate_leaderboards();
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
    INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
    SELECT
        up.project_id,
        up.user_id,
        COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) AS score,
        NOW() AS updated_at
    FROM user_projects up
    LEFT JOIN user_achievements ua ON ua.user_id = up.user_id
    LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.project_id = up.project_id
    LEFT JOIN score_adjustments sa ON sa.entity_type = 'USER'
        AND sa.entity_id = up.user_id
        AND sa.project_id = up.project_id
    GROUP BY up.project_id, up.user_id
    HAVING COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) > 0;
    GET DIAGNOSTICS project_persons_count = ROW_COUNT;

    -- ==================== Regenerate Project Team Leaderboards ====================
    INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at)
    SELECT
        t.project_id,
        t.id AS team_id,
        COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) AS score,
        NOW() AS updated_at
    FROM teams t
    LEFT JOIN team_achievements ta ON ta.team_id = t.id
    LEFT JOIN achievements a ON ta.achievement_id = a.id AND a.project_id = t.project_id
    LEFT JOIN score_adjustments sa ON sa.entity_type = 'TEAM'
        AND sa.entity_id = t.id
        AND sa.project_id = t.project_id
    GROUP BY t.project_id, t.id
    HAVING COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) > 0;
    GET DIAGNOSTICS project_teams_count = ROW_COUNT;

    -- ==================== Regenerate Project SuperTeam Leaderboards ====================
    INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at)
    SELECT
        st.project_id,
        st.id AS super_team_id,
        COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) AS score,
        NOW() AS updated_at
    FROM super_teams st
    LEFT JOIN super_team_achievements sta ON sta.super_team_id = st.id
    LEFT JOIN achievements a ON sta.achievement_id = a.id AND a.project_id = st.project_id
    LEFT JOIN score_adjustments sa ON sa.entity_type = 'SUPER_TEAM'
        AND sa.entity_id = st.id
        AND sa.project_id = st.project_id
    GROUP BY st.project_id, st.id
    HAVING COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) > 0;
    GET DIAGNOSTICS project_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Project Church Leaderboards ====================
    INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at)
    SELECT
        up.project_id,
        u.church_id,
        COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) AS score,
        NOW() AS updated_at
    FROM user_projects up
    INNER JOIN users u ON up.user_id = u.id
    LEFT JOIN user_achievements ua ON ua.user_id = up.user_id
    LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.project_id = up.project_id
    LEFT JOIN score_adjustments sa ON sa.entity_type = 'USER'
        AND sa.entity_id = up.user_id
        AND sa.project_id = up.project_id
    WHERE u.church_id IS NOT NULL
    GROUP BY up.project_id, u.church_id
    HAVING COALESCE(SUM(a.points), 0) + COALESCE(SUM(sa.points), 0) > 0;
    GET DIAGNOSTICS project_churches_count = ROW_COUNT;

    -- ==================== Regenerate Event Person Leaderboards ====================
    INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at)
    SELECT
        ue.event_id,
        ue.user_id,
        COALESCE(SUM(a.points), 0) AS score,
        NOW() AS updated_at
    FROM user_events ue
    LEFT JOIN user_achievements ua ON ua.user_id = ue.user_id
    LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.event_id = ue.event_id
    GROUP BY ue.event_id, ue.user_id
    HAVING COALESCE(SUM(a.points), 0) > 0;
    GET DIAGNOSTICS event_persons_count = ROW_COUNT;

    -- ==================== Regenerate Event Team Leaderboards ====================
    INSERT INTO leaderboard_event_teams (event_id, team_id, score, updated_at)
    SELECT
        e.id AS event_id,
        t.id AS team_id,
        COALESCE(SUM(a.points), 0) AS score,
        NOW() AS updated_at
    FROM events e
    INNER JOIN teams t ON t.project_id = e.project_id
    LEFT JOIN team_achievements ta ON ta.team_id = t.id
    LEFT JOIN achievements a ON ta.achievement_id = a.id AND a.event_id = e.id
    GROUP BY e.id, t.id
    HAVING COALESCE(SUM(a.points), 0) > 0;
    GET DIAGNOSTICS event_teams_count = ROW_COUNT;

    -- ==================== Regenerate Event SuperTeam Leaderboards ====================
    INSERT INTO leaderboard_event_superteams (event_id, super_team_id, score, updated_at)
    SELECT
        e.id AS event_id,
        st.id AS super_team_id,
        COALESCE(SUM(a.points), 0) AS score,
        NOW() AS updated_at
    FROM events e
    INNER JOIN super_teams st ON st.project_id = e.project_id
    LEFT JOIN super_team_achievements sta ON sta.super_team_id = st.id
    LEFT JOIN achievements a ON sta.achievement_id = a.id AND a.event_id = e.id
    GROUP BY e.id, st.id
    HAVING COALESCE(SUM(a.points), 0) > 0;
    GET DIAGNOSTICS event_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Event Church Leaderboards ====================
    INSERT INTO leaderboard_event_churches (event_id, church_id, score, updated_at)
    SELECT
        ue.event_id,
        u.church_id,
        COALESCE(SUM(a.points), 0) AS score,
        NOW() AS updated_at
    FROM user_events ue
    INNER JOIN users u ON ue.user_id = u.id
    LEFT JOIN user_achievements ua ON ua.user_id = ue.user_id
    LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.event_id = ue.event_id
    WHERE u.church_id IS NOT NULL
    GROUP BY ue.event_id, u.church_id
    HAVING COALESCE(SUM(a.points), 0) > 0;
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

DROP FUNCTION IF EXISTS regenerate_leaderboards();

-- +goose StatementEnd
