-- +goose Up
-- +goose StatementBegin

-- Migration: Fix leaderboard triggers and functions to use score_journal as single source of truth
-- This migration fixes the issue where leaderboards were reading from deleted tables
-- (team_achievements, super_team_achievements, score_adjustments)

-- ==================== Drop Obsolete Triggers and Functions ====================

-- Drop triggers on deleted tables
DROP TRIGGER IF EXISTS trigger_team_achievement_leaderboard ON team_achievements;
DROP TRIGGER IF EXISTS trigger_superteam_achievement_leaderboard ON super_team_achievements;
DROP TRIGGER IF EXISTS trigger_score_adjustment_leaderboard ON score_adjustments;
DROP TRIGGER IF EXISTS trigger_user_achievement_leaderboard ON user_achievements;

-- Drop obsolete trigger functions
DROP FUNCTION IF EXISTS trigger_update_leaderboard_team_achievement();
DROP FUNCTION IF EXISTS trigger_update_leaderboard_superteam_achievement();
DROP FUNCTION IF EXISTS trigger_update_leaderboard_score_adjustment();
DROP FUNCTION IF EXISTS trigger_update_leaderboard_user_achievement();

-- ==================== New Trigger Function: Score Journal Updates Leaderboards ====================

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_from_score_journal()
RETURNS TRIGGER AS $$
DECLARE
    points_delta BIGINT;
    target_user_id CHAR(28);
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    user_church_id CHAR(28);
    user_team_id CHAR(28);
    user_super_team_id CHAR(28);
BEGIN
    -- Determine operation and calculate delta
    IF (TG_OP = 'INSERT') THEN
        points_delta := NEW.points;
        target_user_id := NEW.user_id;
        target_project_id := NEW.project_id;
        target_event_id := NEW.event_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        points_delta := NEW.points - OLD.points;
        target_user_id := NEW.user_id;
        target_project_id := NEW.project_id;
        target_event_id := NEW.event_id;
    ELSIF (TG_OP = 'DELETE') THEN
        points_delta := -OLD.points;
        target_user_id := OLD.user_id;
        target_project_id := OLD.project_id;
        target_event_id := OLD.event_id;
    ELSE
        RETURN NULL;
    END IF;

    -- Get user's church, team, and superteam for aggregation
    SELECT u.church_id, tm.team_id, t.super_team_id
    INTO user_church_id, user_team_id, user_super_team_id
    FROM users u
    LEFT JOIN team_members tm ON tm.user_id = u.id
    LEFT JOIN teams t ON t.id = tm.team_id AND t.project_id = target_project_id
    WHERE u.id = target_user_id;

    -- Update person leaderboard
    PERFORM update_person_leaderboard(
        target_user_id,
        target_project_id,
        target_event_id,
        points_delta
    );

    -- Update church leaderboard
    IF user_church_id IS NOT NULL THEN
        PERFORM update_church_leaderboard(
            user_church_id,
            target_project_id,
            target_event_id,
            points_delta
        );
    END IF;

    -- Update team leaderboard
    IF user_team_id IS NOT NULL THEN
        PERFORM update_team_leaderboard(
            user_team_id,
            target_project_id,
            target_event_id,
            points_delta
        );
    END IF;

    -- Update superteam leaderboard
    IF user_super_team_id IS NOT NULL THEN
        PERFORM update_superteam_leaderboard(
            user_super_team_id,
            target_project_id,
            target_event_id,
            points_delta
        );
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on score_journal
CREATE TRIGGER trigger_score_journal_leaderboard
AFTER INSERT OR UPDATE OR DELETE ON score_journal
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_from_score_journal();

-- ==================== Rewrite Regenerate Leaderboards Function ====================

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
    -- Sum all points from score_journal for each user in each project
    INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
    SELECT
        sj.project_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
    FROM score_journal sj
    GROUP BY sj.project_id, sj.user_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS project_persons_count = ROW_COUNT;

    -- ==================== Regenerate Project Team Leaderboards ====================
    -- Sum all points from score_journal for team members in each project
    INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at)
    SELECT
        t.project_id,
        t.id AS team_id,
        COALESCE(SUM(sj.points), 0) AS score,
        NOW() AS updated_at
    FROM teams t
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = t.project_id
    GROUP BY t.project_id, t.id
    HAVING COALESCE(SUM(sj.points), 0) > 0;
    GET DIAGNOSTICS project_teams_count = ROW_COUNT;

    -- ==================== Regenerate Project SuperTeam Leaderboards ====================
    -- Sum all points from score_journal for superteam members (through teams) in each project
    INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at)
    SELECT
        st.project_id,
        st.id AS super_team_id,
        COALESCE(SUM(sj.points), 0) AS score,
        NOW() AS updated_at
    FROM super_teams st
    INNER JOIN teams t ON t.super_team_id = st.id AND t.project_id = st.project_id
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = st.project_id
    GROUP BY st.project_id, st.id
    HAVING COALESCE(SUM(sj.points), 0) > 0;
    GET DIAGNOSTICS project_superteams_count = ROW_COUNT;

    -- ==================== Regenerate Project Church Leaderboards ====================
    -- Sum all points from score_journal for each church in each project
    INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at)
    SELECT
        sj.project_id,
        u.church_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE u.church_id IS NOT NULL
    GROUP BY sj.project_id, u.church_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS project_churches_count = ROW_COUNT;

    -- ==================== Regenerate Event Person Leaderboards ====================
    -- Sum all points from score_journal for each user in each event
    INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at)
    SELECT
        sj.event_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
    FROM score_journal sj
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, sj.user_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_persons_count = ROW_COUNT;

    -- ==================== Regenerate Event Team Leaderboards ====================
    -- Sum all points from score_journal for team members in each event
    INSERT INTO leaderboard_event_teams (event_id, team_id, score, updated_at)
    SELECT
        sj.event_id,
        tm.team_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
    FROM score_journal sj
    INNER JOIN team_members tm ON tm.user_id = sj.user_id
    INNER JOIN teams t ON t.id = tm.team_id
    INNER JOIN events e ON e.id = sj.event_id AND e.project_id = t.project_id
    WHERE sj.event_id IS NOT NULL
    GROUP BY sj.event_id, tm.team_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_teams_count = ROW_COUNT;

    -- ==================== Regenerate Event SuperTeam Leaderboards ====================
    -- Sum all points from score_journal for superteam members (through teams) in each event
    INSERT INTO leaderboard_event_superteams (event_id, super_team_id, score, updated_at)
    SELECT
        sj.event_id,
        t.super_team_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
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
    -- Sum all points from score_journal for each church in each event
    INSERT INTO leaderboard_event_churches (event_id, church_id, score, updated_at)
    SELECT
        sj.event_id,
        u.church_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
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

-- +goose Down
-- +goose StatementBegin

-- Drop new trigger
DROP TRIGGER IF EXISTS trigger_score_journal_leaderboard ON score_journal;
DROP FUNCTION IF EXISTS trigger_update_leaderboard_from_score_journal();

-- Restore old triggers and functions (from migration 00012)
-- Note: These will fail if the old tables don't exist, but that's expected behavior

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_user_achievement()
RETURNS TRIGGER AS $$
DECLARE
    achievement_points BIGINT;
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    user_church_id CHAR(28);
    points_delta BIGINT;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        SELECT a.points, a.project_id, a.event_id
        INTO achievement_points, target_project_id, target_event_id
        FROM achievements a
        WHERE a.id = NEW.achievement_id;
        points_delta := achievement_points;
    ELSIF (TG_OP = 'DELETE') THEN
        SELECT a.points, a.project_id, a.event_id
        INTO achievement_points, target_project_id, target_event_id
        FROM achievements a
        WHERE a.id = OLD.achievement_id;
        points_delta := -achievement_points;
    ELSE
        RETURN NULL;
    END IF;

    SELECT church_id INTO user_church_id
    FROM users
    WHERE id = COALESCE(NEW.user_id, OLD.user_id);

    PERFORM update_person_leaderboard(
        COALESCE(NEW.user_id, OLD.user_id),
        target_project_id,
        target_event_id,
        points_delta
    );

    IF user_church_id IS NOT NULL THEN
        PERFORM update_church_leaderboard(
            user_church_id,
            target_project_id,
            target_event_id,
            points_delta
        );
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_achievement_leaderboard
AFTER INSERT OR DELETE ON user_achievements
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_user_achievement();

-- Restore old regenerate_leaderboards function
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
    TRUNCATE TABLE leaderboard_project_persons;
    TRUNCATE TABLE leaderboard_project_teams;
    TRUNCATE TABLE leaderboard_project_superteams;
    TRUNCATE TABLE leaderboard_project_churches;
    TRUNCATE TABLE leaderboard_event_persons;
    TRUNCATE TABLE leaderboard_event_teams;
    TRUNCATE TABLE leaderboard_event_superteams;
    TRUNCATE TABLE leaderboard_event_churches;

    INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
    SELECT
        up.project_id,
        up.user_id,
        COALESCE(SUM(a.points), 0) AS score,
        NOW() AS updated_at
    FROM user_projects up
    LEFT JOIN user_achievements ua ON ua.user_id = up.user_id
    LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.project_id = up.project_id
    GROUP BY up.project_id, up.user_id
    HAVING COALESCE(SUM(a.points), 0) > 0;
    GET DIAGNOSTICS project_persons_count = ROW_COUNT;

    INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at)
    SELECT
        up.project_id,
        u.church_id,
        COALESCE(SUM(a.points), 0) AS score,
        NOW() AS updated_at
    FROM user_projects up
    INNER JOIN users u ON up.user_id = u.id
    LEFT JOIN user_achievements ua ON ua.user_id = up.user_id
    LEFT JOIN achievements a ON ua.achievement_id = a.id AND a.project_id = up.project_id
    WHERE u.church_id IS NOT NULL
    GROUP BY up.project_id, u.church_id
    HAVING COALESCE(SUM(a.points), 0) > 0;
    GET DIAGNOSTICS project_churches_count = ROW_COUNT;

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

    RETURN QUERY
    SELECT 'leaderboard_project_persons'::TEXT, project_persons_count
    UNION ALL SELECT 'leaderboard_project_teams'::TEXT, 0::BIGINT
    UNION ALL SELECT 'leaderboard_project_superteams'::TEXT, 0::BIGINT
    UNION ALL SELECT 'leaderboard_project_churches'::TEXT, project_churches_count
    UNION ALL SELECT 'leaderboard_event_persons'::TEXT, event_persons_count
    UNION ALL SELECT 'leaderboard_event_teams'::TEXT, 0::BIGINT
    UNION ALL SELECT 'leaderboard_event_superteams'::TEXT, 0::BIGINT
    UNION ALL SELECT 'leaderboard_event_churches'::TEXT, event_churches_count;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd
