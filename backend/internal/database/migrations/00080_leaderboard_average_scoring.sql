-- +goose Up
-- +goose StatementBegin

-- Migration: Change leaderboard scoring from SUM to AVERAGE for Teams, SuperTeams, and Churches
-- This adds total_points and member_count columns to enable average calculation:
-- score = total_points / member_count (rounded to nearest integer)
-- Person leaderboards remain unchanged (no averaging)

-- ==================== Add Columns to 6 Leaderboard Tables ====================

ALTER TABLE leaderboard_project_teams
    ADD COLUMN total_points BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN member_count INT NOT NULL DEFAULT 0;

ALTER TABLE leaderboard_event_teams
    ADD COLUMN total_points BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN member_count INT NOT NULL DEFAULT 0;

ALTER TABLE leaderboard_project_superteams
    ADD COLUMN total_points BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN member_count INT NOT NULL DEFAULT 0;

ALTER TABLE leaderboard_event_superteams
    ADD COLUMN total_points BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN member_count INT NOT NULL DEFAULT 0;

ALTER TABLE leaderboard_project_churches
    ADD COLUMN total_points BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN member_count INT NOT NULL DEFAULT 0;

ALTER TABLE leaderboard_event_churches
    ADD COLUMN total_points BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN member_count INT NOT NULL DEFAULT 0;

-- ==================== Create Helper Function for Safe Division ====================

CREATE OR REPLACE FUNCTION calculate_average_score(
    p_total_points BIGINT,
    p_member_count INT
) RETURNS BIGINT AS $$
BEGIN
    IF p_member_count <= 0 THEN
        RETURN 0;
    END IF;
    RETURN ROUND(p_total_points::NUMERIC / p_member_count)::BIGINT;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ==================== Update Helper Functions ====================

-- Update team leaderboard helper
CREATE OR REPLACE FUNCTION update_team_leaderboard(
    p_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
DECLARE
    v_member_count INT;
    v_total_points BIGINT;
BEGIN
    -- Update project team leaderboard
    IF p_project_id IS NOT NULL THEN
        -- Get member count for this team (all team members)
        SELECT COUNT(*)::INT INTO v_member_count
        FROM team_members tm
        WHERE tm.team_id = p_team_id;

        -- Upsert and recalculate
        INSERT INTO leaderboard_project_teams (project_id, team_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            p_project_id,
            p_team_id,
            p_points_delta,
            v_member_count,
            calculate_average_score(p_points_delta, v_member_count),
            NOW(),
            p_score_at
        )
        ON CONFLICT (project_id, team_id)
        DO UPDATE SET
            total_points = leaderboard_project_teams.total_points + p_points_delta,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_project_teams.total_points + p_points_delta, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_teams.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event team leaderboard
    IF p_event_id IS NOT NULL THEN
        -- Get member count for this team who are in this event
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = p_event_id
        WHERE tm.team_id = p_team_id;

        -- Upsert and recalculate
        INSERT INTO leaderboard_event_teams (event_id, team_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            p_event_id,
            p_team_id,
            p_points_delta,
            v_member_count,
            calculate_average_score(p_points_delta, v_member_count),
            NOW(),
            p_score_at
        )
        ON CONFLICT (event_id, team_id)
        DO UPDATE SET
            total_points = leaderboard_event_teams.total_points + p_points_delta,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_event_teams.total_points + p_points_delta, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_teams.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Update superteam leaderboard helper
CREATE OR REPLACE FUNCTION update_superteam_leaderboard(
    p_super_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
DECLARE
    v_member_count INT;
    v_total_points BIGINT;
BEGIN
    -- Update project superteam leaderboard
    IF p_project_id IS NOT NULL THEN
        -- Get member count: distinct users across all teams in this superteam
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id
        WHERE t.super_team_id = p_super_team_id
          AND t.project_id = p_project_id;

        -- Upsert and recalculate
        INSERT INTO leaderboard_project_superteams (project_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            p_project_id,
            p_super_team_id,
            p_points_delta,
            v_member_count,
            calculate_average_score(p_points_delta, v_member_count),
            NOW(),
            p_score_at
        )
        ON CONFLICT (project_id, super_team_id)
        DO UPDATE SET
            total_points = leaderboard_project_superteams.total_points + p_points_delta,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_project_superteams.total_points + p_points_delta, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_superteams.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event superteam leaderboard
    IF p_event_id IS NOT NULL THEN
        -- Get member count: distinct users across all teams in this superteam who are in the event
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id
        INNER JOIN events e ON e.project_id = t.project_id AND e.id = p_event_id
        INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = p_event_id
        WHERE t.super_team_id = p_super_team_id;

        -- Upsert and recalculate
        INSERT INTO leaderboard_event_superteams (event_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            p_event_id,
            p_super_team_id,
            p_points_delta,
            v_member_count,
            calculate_average_score(p_points_delta, v_member_count),
            NOW(),
            p_score_at
        )
        ON CONFLICT (event_id, super_team_id)
        DO UPDATE SET
            total_points = leaderboard_event_superteams.total_points + p_points_delta,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_event_superteams.total_points + p_points_delta, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_superteams.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Update church leaderboard helper
CREATE OR REPLACE FUNCTION update_church_leaderboard(
    p_church_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
DECLARE
    v_member_count INT;
    v_total_points BIGINT;
BEGIN
    -- Update project church leaderboard
    IF p_project_id IS NOT NULL THEN
        -- Get member count: users with this church_id who are in the project
        SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
        FROM users u
        INNER JOIN user_projects up ON up.user_id = u.id AND up.project_id = p_project_id
        WHERE u.church_id = p_church_id;

        -- Upsert and recalculate
        INSERT INTO leaderboard_project_churches (project_id, church_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            p_project_id,
            p_church_id,
            p_points_delta,
            v_member_count,
            calculate_average_score(p_points_delta, v_member_count),
            NOW(),
            p_score_at
        )
        ON CONFLICT (project_id, church_id)
        DO UPDATE SET
            total_points = leaderboard_project_churches.total_points + p_points_delta,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_project_churches.total_points + p_points_delta, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event church leaderboard
    IF p_event_id IS NOT NULL THEN
        -- Get member count: users with this church_id who are in the event
        SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
        FROM users u
        INNER JOIN user_events ue ON ue.user_id = u.id AND ue.event_id = p_event_id
        WHERE u.church_id = p_church_id;

        -- Upsert and recalculate
        INSERT INTO leaderboard_event_churches (event_id, church_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            p_event_id,
            p_church_id,
            p_points_delta,
            v_member_count,
            calculate_average_score(p_points_delta, v_member_count),
            NOW(),
            p_score_at
        )
        ON CONFLICT (event_id, church_id)
        DO UPDATE SET
            total_points = leaderboard_event_churches.total_points + p_points_delta,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_event_churches.total_points + p_points_delta, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_churches.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ==================== Membership Change Triggers ====================

-- Function to recalculate team average when membership changes
CREATE OR REPLACE FUNCTION trigger_recalculate_team_average_on_membership_change()
RETURNS TRIGGER AS $$
DECLARE
    v_team_id CHAR(28);
    v_project_id CHAR(28);
    v_super_team_id CHAR(28);
    v_member_count INT;
BEGIN
    -- Get the team_id from the operation
    v_team_id := COALESCE(NEW.team_id, OLD.team_id);

    -- Get team's project_id and super_team_id
    SELECT t.project_id, t.super_team_id
    INTO v_project_id, v_super_team_id
    FROM teams t
    WHERE t.id = v_team_id;

    IF v_project_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Update project team leaderboard if it exists
    SELECT COUNT(*)::INT INTO v_member_count
    FROM team_members tm
    WHERE tm.team_id = v_team_id;

    UPDATE leaderboard_project_teams
    SET member_count = v_member_count,
        score = calculate_average_score(total_points, v_member_count),
        updated_at = NOW()
    WHERE project_id = v_project_id AND team_id = v_team_id;

    -- Update project superteam leaderboard if it exists
    IF v_super_team_id IS NOT NULL THEN
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id
        WHERE t.super_team_id = v_super_team_id
          AND t.project_id = v_project_id;

        UPDATE leaderboard_project_superteams
        SET member_count = v_member_count,
            score = calculate_average_score(total_points, v_member_count),
            updated_at = NOW()
        WHERE project_id = v_project_id AND super_team_id = v_super_team_id;
    END IF;

    -- Update event leaderboards for all events in this project
    -- Team event leaderboards
    UPDATE leaderboard_event_teams let
    SET member_count = (
            SELECT COUNT(DISTINCT tm.user_id)::INT
            FROM team_members tm
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = let.event_id
            WHERE tm.team_id = v_team_id
        ),
        score = calculate_average_score(let.total_points, (
            SELECT COUNT(DISTINCT tm.user_id)::INT
            FROM team_members tm
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = let.event_id
            WHERE tm.team_id = v_team_id
        )),
        updated_at = NOW()
    WHERE let.team_id = v_team_id;

    -- Superteam event leaderboards
    IF v_super_team_id IS NOT NULL THEN
        UPDATE leaderboard_event_superteams les
        SET member_count = (
                SELECT COUNT(DISTINCT tm.user_id)::INT
                FROM team_members tm
                INNER JOIN teams t ON t.id = tm.team_id
                INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = les.event_id
                WHERE t.super_team_id = v_super_team_id
            ),
            score = calculate_average_score(les.total_points, (
                SELECT COUNT(DISTINCT tm.user_id)::INT
                FROM team_members tm
                INNER JOIN teams t ON t.id = tm.team_id
                INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = les.event_id
                WHERE t.super_team_id = v_super_team_id
            )),
            updated_at = NOW()
        WHERE les.super_team_id = v_super_team_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_team_members_recalculate_average
AFTER INSERT OR DELETE ON team_members
FOR EACH ROW
EXECUTE FUNCTION trigger_recalculate_team_average_on_membership_change();

-- Function to recalculate church average when user joins/leaves project
CREATE OR REPLACE FUNCTION trigger_recalculate_church_average_on_project_membership_change()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id CHAR(28);
    v_project_id CHAR(28);
    v_church_id CHAR(28);
    v_member_count INT;
BEGIN
    v_user_id := COALESCE(NEW.user_id, OLD.user_id);
    v_project_id := COALESCE(NEW.project_id, OLD.project_id);

    -- Get user's church_id
    SELECT u.church_id INTO v_church_id
    FROM users u
    WHERE u.id = v_user_id;

    IF v_church_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Recalculate church member count for this project
    SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
    FROM users u
    INNER JOIN user_projects up ON up.user_id = u.id AND up.project_id = v_project_id
    WHERE u.church_id = v_church_id;

    UPDATE leaderboard_project_churches
    SET member_count = v_member_count,
        score = calculate_average_score(total_points, v_member_count),
        updated_at = NOW()
    WHERE project_id = v_project_id AND church_id = v_church_id;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_projects_recalculate_church_average
AFTER INSERT OR DELETE ON user_projects
FOR EACH ROW
EXECUTE FUNCTION trigger_recalculate_church_average_on_project_membership_change();

-- Function to recalculate averages when user joins/leaves event
CREATE OR REPLACE FUNCTION trigger_recalculate_averages_on_event_membership_change()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id CHAR(28);
    v_event_id CHAR(28);
    v_church_id CHAR(28);
    v_member_count INT;
    team_record RECORD;
BEGIN
    v_user_id := COALESCE(NEW.user_id, OLD.user_id);
    v_event_id := COALESCE(NEW.event_id, OLD.event_id);

    -- Get user's church_id
    SELECT u.church_id INTO v_church_id
    FROM users u
    WHERE u.id = v_user_id;

    -- Update church event leaderboard
    IF v_church_id IS NOT NULL THEN
        SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
        FROM users u
        INNER JOIN user_events ue ON ue.user_id = u.id AND ue.event_id = v_event_id
        WHERE u.church_id = v_church_id;

        UPDATE leaderboard_event_churches
        SET member_count = v_member_count,
            score = calculate_average_score(total_points, v_member_count),
            updated_at = NOW()
        WHERE event_id = v_event_id AND church_id = v_church_id;
    END IF;

    -- Update team event leaderboards for teams this user belongs to
    FOR team_record IN
        SELECT t.id AS team_id, t.super_team_id
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id
        INNER JOIN events e ON e.project_id = t.project_id AND e.id = v_event_id
        WHERE tm.user_id = v_user_id
    LOOP
        -- Update team event leaderboard
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = v_event_id
        WHERE tm.team_id = team_record.team_id;

        UPDATE leaderboard_event_teams
        SET member_count = v_member_count,
            score = calculate_average_score(total_points, v_member_count),
            updated_at = NOW()
        WHERE event_id = v_event_id AND team_id = team_record.team_id;

        -- Update superteam event leaderboard
        IF team_record.super_team_id IS NOT NULL THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN teams t ON t.id = tm.team_id
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = v_event_id
            WHERE t.super_team_id = team_record.super_team_id;

            UPDATE leaderboard_event_superteams
            SET member_count = v_member_count,
                score = calculate_average_score(total_points, v_member_count),
                updated_at = NOW()
            WHERE event_id = v_event_id AND super_team_id = team_record.super_team_id;
        END IF;
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_events_recalculate_averages
AFTER INSERT OR DELETE ON user_events
FOR EACH ROW
EXECUTE FUNCTION trigger_recalculate_averages_on_event_membership_change();

-- ==================== Update Regenerate Leaderboards Function ====================

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

-- +goose Down
-- +goose StatementBegin

-- Drop new triggers
DROP TRIGGER IF EXISTS trigger_user_events_recalculate_averages ON user_events;
DROP TRIGGER IF EXISTS trigger_user_projects_recalculate_church_average ON user_projects;
DROP TRIGGER IF EXISTS trigger_team_members_recalculate_average ON team_members;

-- Drop new trigger functions
DROP FUNCTION IF EXISTS trigger_recalculate_averages_on_event_membership_change();
DROP FUNCTION IF EXISTS trigger_recalculate_church_average_on_project_membership_change();
DROP FUNCTION IF EXISTS trigger_recalculate_team_average_on_membership_change();

-- Drop new columns
ALTER TABLE leaderboard_project_teams DROP COLUMN IF EXISTS total_points;
ALTER TABLE leaderboard_project_teams DROP COLUMN IF EXISTS member_count;
ALTER TABLE leaderboard_event_teams DROP COLUMN IF EXISTS total_points;
ALTER TABLE leaderboard_event_teams DROP COLUMN IF EXISTS member_count;
ALTER TABLE leaderboard_project_superteams DROP COLUMN IF EXISTS total_points;
ALTER TABLE leaderboard_project_superteams DROP COLUMN IF EXISTS member_count;
ALTER TABLE leaderboard_event_superteams DROP COLUMN IF EXISTS total_points;
ALTER TABLE leaderboard_event_superteams DROP COLUMN IF EXISTS member_count;
ALTER TABLE leaderboard_project_churches DROP COLUMN IF EXISTS total_points;
ALTER TABLE leaderboard_project_churches DROP COLUMN IF EXISTS member_count;
ALTER TABLE leaderboard_event_churches DROP COLUMN IF EXISTS total_points;
ALTER TABLE leaderboard_event_churches DROP COLUMN IF EXISTS member_count;

-- Drop helper function
DROP FUNCTION IF EXISTS calculate_average_score(BIGINT, INT);

-- Restore original helper functions (from migration 00069)
CREATE OR REPLACE FUNCTION update_team_leaderboard(
    p_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_team_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, team_id)
        DO UPDATE SET
            score = leaderboard_project_teams.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_teams.last_score_at, p_score_at), p_score_at);
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_teams (event_id, team_id, score, updated_at, last_score_at)
        VALUES (p_event_id, p_team_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (event_id, team_id)
        DO UPDATE SET
            score = leaderboard_event_teams.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_teams.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_superteam_leaderboard(
    p_super_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_super_team_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, super_team_id)
        DO UPDATE SET
            score = leaderboard_project_superteams.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_superteams.last_score_at, p_score_at), p_score_at);
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_superteams (event_id, super_team_id, score, updated_at, last_score_at)
        VALUES (p_event_id, p_super_team_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (event_id, super_team_id)
        DO UPDATE SET
            score = leaderboard_event_superteams.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_superteams.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_church_leaderboard(
    p_church_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_church_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, church_id)
        DO UPDATE SET
            score = leaderboard_project_churches.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, p_score_at), p_score_at);
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_churches (event_id, church_id, score, updated_at, last_score_at)
        VALUES (p_event_id, p_church_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (event_id, church_id)
        DO UPDATE SET
            score = leaderboard_event_churches.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_churches.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Restore original regenerate function (from migration 00069)
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

    INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at, last_score_at)
    SELECT
        t.project_id,
        t.id AS team_id,
        COALESCE(SUM(sj.points), 0) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM teams t
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = t.project_id
    GROUP BY t.project_id, t.id
    HAVING COALESCE(SUM(sj.points), 0) > 0;
    GET DIAGNOSTICS project_teams_count = ROW_COUNT;

    INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at, last_score_at)
    SELECT
        st.project_id,
        st.id AS super_team_id,
        COALESCE(SUM(sj.points), 0) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM super_teams st
    INNER JOIN teams t ON t.super_team_id = st.id AND t.project_id = st.project_id
    INNER JOIN team_members tm ON tm.team_id = t.id
    INNER JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = st.project_id
    GROUP BY st.project_id, st.id
    HAVING COALESCE(SUM(sj.points), 0) > 0;
    GET DIAGNOSTICS project_superteams_count = ROW_COUNT;

    INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at, last_score_at)
    SELECT
        sj.project_id,
        u.church_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE u.church_id IS NOT NULL
    GROUP BY sj.project_id, u.church_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS project_churches_count = ROW_COUNT;

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

    INSERT INTO leaderboard_event_teams (event_id, team_id, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        tm.team_id,
        SUM(sj.points) AS score,
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

    INSERT INTO leaderboard_event_superteams (event_id, super_team_id, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        t.super_team_id,
        SUM(sj.points) AS score,
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

    INSERT INTO leaderboard_event_churches (event_id, church_id, score, updated_at, last_score_at)
    SELECT
        sj.event_id,
        u.church_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at,
        MAX(sj.created_at) AS last_score_at
    FROM score_journal sj
    INNER JOIN users u ON sj.user_id = u.id
    WHERE sj.event_id IS NOT NULL
    AND u.church_id IS NOT NULL
    GROUP BY sj.event_id, u.church_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS event_churches_count = ROW_COUNT;

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
