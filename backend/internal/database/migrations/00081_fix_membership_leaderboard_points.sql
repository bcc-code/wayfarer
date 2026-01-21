-- +goose Up
-- +goose StatementBegin

-- Fix: Membership change triggers now add/subtract member's existing points to/from total_points
-- Previously, these triggers only updated member_count and recalculated score,
-- but did NOT transfer the member's existing points to/from total_points.

-- ==================== Team Membership Trigger ====================
-- When a user joins/leaves a team, we need to:
-- 1. Update member_count (existing)
-- 2. Add/subtract user's project points to/from leaderboard_project_teams.total_points (NEW)
-- 3. Add/subtract user's project points to/from leaderboard_project_superteams.total_points (NEW)
-- 4. Add/subtract user's event points to/from leaderboard_event_teams.total_points (NEW)
-- 5. Add/subtract user's event points to/from leaderboard_event_superteams.total_points (NEW)

CREATE OR REPLACE FUNCTION trigger_recalculate_team_average_on_membership_change()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id CHAR(28);
    v_team_id CHAR(28);
    v_project_id CHAR(28);
    v_super_team_id CHAR(28);
    v_member_count INT;
    v_user_project_points BIGINT;
    v_user_event_points BIGINT;
    v_points_multiplier INT;
    v_last_score_at TIMESTAMPTZ;
    event_record RECORD;
BEGIN
    -- Determine operation: INSERT adds points (+1), DELETE subtracts (-1)
    IF TG_OP = 'INSERT' THEN
        v_user_id := NEW.user_id;
        v_team_id := NEW.team_id;
        v_points_multiplier := 1;
    ELSE
        v_user_id := OLD.user_id;
        v_team_id := OLD.team_id;
        v_points_multiplier := -1;
    END IF;

    -- Get team's project_id and super_team_id
    SELECT t.project_id, t.super_team_id
    INTO v_project_id, v_super_team_id
    FROM teams t
    WHERE t.id = v_team_id;

    IF v_project_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Get user's total project points from score_journal
    SELECT COALESCE(SUM(sj.points), 0), MAX(sj.created_at)
    INTO v_user_project_points, v_last_score_at
    FROM score_journal sj
    WHERE sj.user_id = v_user_id
      AND sj.project_id = v_project_id;

    -- ==== Update Project Team Leaderboard ====
    SELECT COUNT(*)::INT INTO v_member_count
    FROM team_members tm
    WHERE tm.team_id = v_team_id;

    -- Upsert: create row if it doesn't exist (on INSERT with points), or update existing
    IF TG_OP = 'INSERT' AND v_user_project_points > 0 THEN
        INSERT INTO leaderboard_project_teams (project_id, team_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            v_project_id,
            v_team_id,
            v_user_project_points,
            v_member_count,
            calculate_average_score(v_user_project_points, v_member_count),
            NOW(),
            v_last_score_at
        )
        ON CONFLICT (project_id, team_id)
        DO UPDATE SET
            total_points = leaderboard_project_teams.total_points + v_user_project_points,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_project_teams.total_points + v_user_project_points, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_teams.last_score_at, v_last_score_at), v_last_score_at);
    ELSE
        -- DELETE or INSERT with 0 points: just update existing row
        UPDATE leaderboard_project_teams
        SET total_points = total_points + (v_user_project_points * v_points_multiplier),
            member_count = v_member_count,
            score = calculate_average_score(total_points + (v_user_project_points * v_points_multiplier), v_member_count),
            updated_at = NOW()
        WHERE project_id = v_project_id AND team_id = v_team_id;
    END IF;

    -- ==== Update Project SuperTeam Leaderboard ====
    IF v_super_team_id IS NOT NULL THEN
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id
        WHERE t.super_team_id = v_super_team_id
          AND t.project_id = v_project_id;

        IF TG_OP = 'INSERT' AND v_user_project_points > 0 THEN
            INSERT INTO leaderboard_project_superteams (project_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (
                v_project_id,
                v_super_team_id,
                v_user_project_points,
                v_member_count,
                calculate_average_score(v_user_project_points, v_member_count),
                NOW(),
                v_last_score_at
            )
            ON CONFLICT (project_id, super_team_id)
            DO UPDATE SET
                total_points = leaderboard_project_superteams.total_points + v_user_project_points,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_project_superteams.total_points + v_user_project_points, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_project_superteams.last_score_at, v_last_score_at), v_last_score_at);
        ELSE
            UPDATE leaderboard_project_superteams
            SET total_points = total_points + (v_user_project_points * v_points_multiplier),
                member_count = v_member_count,
                score = calculate_average_score(total_points + (v_user_project_points * v_points_multiplier), v_member_count),
                updated_at = NOW()
            WHERE project_id = v_project_id AND super_team_id = v_super_team_id;
        END IF;
    END IF;

    -- ==== Update Event Leaderboards ====
    -- For each event in this project that the user participates in
    FOR event_record IN
        SELECT e.id AS event_id
        FROM events e
        INNER JOIN user_events ue ON ue.event_id = e.id AND ue.user_id = v_user_id
        WHERE e.project_id = v_project_id
    LOOP
        -- Get user's event points
        SELECT COALESCE(SUM(sj.points), 0), MAX(sj.created_at)
        INTO v_user_event_points, v_last_score_at
        FROM score_journal sj
        WHERE sj.user_id = v_user_id
          AND sj.event_id = event_record.event_id;

        -- Update event team leaderboard
        SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
        FROM team_members tm
        INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = event_record.event_id
        WHERE tm.team_id = v_team_id;

        IF TG_OP = 'INSERT' AND v_user_event_points > 0 THEN
            INSERT INTO leaderboard_event_teams (event_id, team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (
                event_record.event_id,
                v_team_id,
                v_user_event_points,
                v_member_count,
                calculate_average_score(v_user_event_points, v_member_count),
                NOW(),
                v_last_score_at
            )
            ON CONFLICT (event_id, team_id)
            DO UPDATE SET
                total_points = leaderboard_event_teams.total_points + v_user_event_points,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_event_teams.total_points + v_user_event_points, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_teams.last_score_at, v_last_score_at), v_last_score_at);
        ELSE
            UPDATE leaderboard_event_teams
            SET total_points = total_points + (v_user_event_points * v_points_multiplier),
                member_count = v_member_count,
                score = calculate_average_score(total_points + (v_user_event_points * v_points_multiplier), v_member_count),
                updated_at = NOW()
            WHERE event_id = event_record.event_id AND team_id = v_team_id;
        END IF;

        -- Update event superteam leaderboard
        IF v_super_team_id IS NOT NULL THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN teams t ON t.id = tm.team_id
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = event_record.event_id
            WHERE t.super_team_id = v_super_team_id;

            IF TG_OP = 'INSERT' AND v_user_event_points > 0 THEN
                INSERT INTO leaderboard_event_superteams (event_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
                VALUES (
                    event_record.event_id,
                    v_super_team_id,
                    v_user_event_points,
                    v_member_count,
                    calculate_average_score(v_user_event_points, v_member_count),
                    NOW(),
                    v_last_score_at
                )
                ON CONFLICT (event_id, super_team_id)
                DO UPDATE SET
                    total_points = leaderboard_event_superteams.total_points + v_user_event_points,
                    member_count = v_member_count,
                    score = calculate_average_score(leaderboard_event_superteams.total_points + v_user_event_points, v_member_count),
                    updated_at = NOW(),
                    last_score_at = GREATEST(COALESCE(leaderboard_event_superteams.last_score_at, v_last_score_at), v_last_score_at);
            ELSE
                UPDATE leaderboard_event_superteams
                SET total_points = total_points + (v_user_event_points * v_points_multiplier),
                    member_count = v_member_count,
                    score = calculate_average_score(total_points + (v_user_event_points * v_points_multiplier), v_member_count),
                    updated_at = NOW()
                WHERE event_id = event_record.event_id AND super_team_id = v_super_team_id;
            END IF;
        END IF;
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ==================== Project Membership Trigger ====================
-- When a user joins/leaves a project, we need to:
-- 1. Update member_count for their church (existing)
-- 2. Add/subtract user's project points to/from leaderboard_project_churches.total_points (NEW)

CREATE OR REPLACE FUNCTION trigger_recalculate_church_average_on_project_membership_change()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id CHAR(28);
    v_project_id CHAR(28);
    v_church_id CHAR(28);
    v_member_count INT;
    v_user_project_points BIGINT;
    v_points_multiplier INT;
    v_last_score_at TIMESTAMPTZ;
BEGIN
    -- Determine operation
    IF TG_OP = 'INSERT' THEN
        v_user_id := NEW.user_id;
        v_project_id := NEW.project_id;
        v_points_multiplier := 1;
    ELSE
        v_user_id := OLD.user_id;
        v_project_id := OLD.project_id;
        v_points_multiplier := -1;
    END IF;

    -- Get user's church_id
    SELECT u.church_id INTO v_church_id
    FROM users u
    WHERE u.id = v_user_id;

    IF v_church_id IS NULL THEN
        RETURN NULL;
    END IF;

    -- Get user's total project points from score_journal
    SELECT COALESCE(SUM(sj.points), 0), MAX(sj.created_at)
    INTO v_user_project_points, v_last_score_at
    FROM score_journal sj
    WHERE sj.user_id = v_user_id
      AND sj.project_id = v_project_id;

    -- Recalculate church member count for this project
    SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
    FROM users u
    INNER JOIN user_projects up ON up.user_id = u.id AND up.project_id = v_project_id
    WHERE u.church_id = v_church_id;

    -- Upsert: create row if it doesn't exist, or update existing
    IF TG_OP = 'INSERT' AND v_user_project_points > 0 THEN
        INSERT INTO leaderboard_project_churches (project_id, church_id, total_points, member_count, score, updated_at, last_score_at)
        VALUES (
            v_project_id,
            v_church_id,
            v_user_project_points,
            v_member_count,
            calculate_average_score(v_user_project_points, v_member_count),
            NOW(),
            v_last_score_at
        )
        ON CONFLICT (project_id, church_id)
        DO UPDATE SET
            total_points = leaderboard_project_churches.total_points + v_user_project_points,
            member_count = v_member_count,
            score = calculate_average_score(leaderboard_project_churches.total_points + v_user_project_points, v_member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, v_last_score_at), v_last_score_at);
    ELSE
        UPDATE leaderboard_project_churches
        SET total_points = total_points + (v_user_project_points * v_points_multiplier),
            member_count = v_member_count,
            score = calculate_average_score(total_points + (v_user_project_points * v_points_multiplier), v_member_count),
            updated_at = NOW()
        WHERE project_id = v_project_id AND church_id = v_church_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ==================== Event Membership Trigger ====================
-- When a user joins/leaves an event, we need to:
-- 1. Update member_count for their church (existing)
-- 2. Add/subtract user's event points to/from leaderboard_event_churches.total_points (NEW)
-- 3. Update member_count for their team(s) (existing)
-- 4. Add/subtract user's event points to/from leaderboard_event_teams.total_points (NEW)
-- 5. Update member_count for their superteam(s) (existing)
-- 6. Add/subtract user's event points to/from leaderboard_event_superteams.total_points (NEW)

CREATE OR REPLACE FUNCTION trigger_recalculate_averages_on_event_membership_change()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id CHAR(28);
    v_event_id CHAR(28);
    v_church_id CHAR(28);
    v_member_count INT;
    v_user_event_points BIGINT;
    v_points_multiplier INT;
    v_last_score_at TIMESTAMPTZ;
    team_record RECORD;
BEGIN
    -- Determine operation
    IF TG_OP = 'INSERT' THEN
        v_user_id := NEW.user_id;
        v_event_id := NEW.event_id;
        v_points_multiplier := 1;
    ELSE
        v_user_id := OLD.user_id;
        v_event_id := OLD.event_id;
        v_points_multiplier := -1;
    END IF;

    -- Get user's church_id
    SELECT u.church_id INTO v_church_id
    FROM users u
    WHERE u.id = v_user_id;

    -- Get user's event points from score_journal
    SELECT COALESCE(SUM(sj.points), 0), MAX(sj.created_at)
    INTO v_user_event_points, v_last_score_at
    FROM score_journal sj
    WHERE sj.user_id = v_user_id
      AND sj.event_id = v_event_id;

    -- ==== Update Church Event Leaderboard ====
    IF v_church_id IS NOT NULL THEN
        SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
        FROM users u
        INNER JOIN user_events ue ON ue.user_id = u.id AND ue.event_id = v_event_id
        WHERE u.church_id = v_church_id;

        IF TG_OP = 'INSERT' AND v_user_event_points > 0 THEN
            INSERT INTO leaderboard_event_churches (event_id, church_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (
                v_event_id,
                v_church_id,
                v_user_event_points,
                v_member_count,
                calculate_average_score(v_user_event_points, v_member_count),
                NOW(),
                v_last_score_at
            )
            ON CONFLICT (event_id, church_id)
            DO UPDATE SET
                total_points = leaderboard_event_churches.total_points + v_user_event_points,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_event_churches.total_points + v_user_event_points, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_churches.last_score_at, v_last_score_at), v_last_score_at);
        ELSE
            UPDATE leaderboard_event_churches
            SET total_points = total_points + (v_user_event_points * v_points_multiplier),
                member_count = v_member_count,
                score = calculate_average_score(total_points + (v_user_event_points * v_points_multiplier), v_member_count),
                updated_at = NOW()
            WHERE event_id = v_event_id AND church_id = v_church_id;
        END IF;
    END IF;

    -- ==== Update Team and SuperTeam Event Leaderboards ====
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

        IF TG_OP = 'INSERT' AND v_user_event_points > 0 THEN
            INSERT INTO leaderboard_event_teams (event_id, team_id, total_points, member_count, score, updated_at, last_score_at)
            VALUES (
                v_event_id,
                team_record.team_id,
                v_user_event_points,
                v_member_count,
                calculate_average_score(v_user_event_points, v_member_count),
                NOW(),
                v_last_score_at
            )
            ON CONFLICT (event_id, team_id)
            DO UPDATE SET
                total_points = leaderboard_event_teams.total_points + v_user_event_points,
                member_count = v_member_count,
                score = calculate_average_score(leaderboard_event_teams.total_points + v_user_event_points, v_member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_teams.last_score_at, v_last_score_at), v_last_score_at);
        ELSE
            UPDATE leaderboard_event_teams
            SET total_points = total_points + (v_user_event_points * v_points_multiplier),
                member_count = v_member_count,
                score = calculate_average_score(total_points + (v_user_event_points * v_points_multiplier), v_member_count),
                updated_at = NOW()
            WHERE event_id = v_event_id AND team_id = team_record.team_id;
        END IF;

        -- Update superteam event leaderboard
        IF team_record.super_team_id IS NOT NULL THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN teams t ON t.id = tm.team_id
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = v_event_id
            WHERE t.super_team_id = team_record.super_team_id;

            IF TG_OP = 'INSERT' AND v_user_event_points > 0 THEN
                INSERT INTO leaderboard_event_superteams (event_id, super_team_id, total_points, member_count, score, updated_at, last_score_at)
                VALUES (
                    v_event_id,
                    team_record.super_team_id,
                    v_user_event_points,
                    v_member_count,
                    calculate_average_score(v_user_event_points, v_member_count),
                    NOW(),
                    v_last_score_at
                )
                ON CONFLICT (event_id, super_team_id)
                DO UPDATE SET
                    total_points = leaderboard_event_superteams.total_points + v_user_event_points,
                    member_count = v_member_count,
                    score = calculate_average_score(leaderboard_event_superteams.total_points + v_user_event_points, v_member_count),
                    updated_at = NOW(),
                    last_score_at = GREATEST(COALESCE(leaderboard_event_superteams.last_score_at, v_last_score_at), v_last_score_at);
            ELSE
                UPDATE leaderboard_event_superteams
                SET total_points = total_points + (v_user_event_points * v_points_multiplier),
                    member_count = v_member_count,
                    score = calculate_average_score(total_points + (v_user_event_points * v_points_multiplier), v_member_count),
                    updated_at = NOW()
                WHERE event_id = v_event_id AND super_team_id = team_record.super_team_id;
            END IF;
        END IF;
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original trigger functions from migration 00080

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

-- +goose StatementEnd
