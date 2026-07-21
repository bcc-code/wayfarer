-- +goose Up
-- +goose StatementBegin

-- Perf: membership triggers and the score-journal church helper previously recomputed
-- member_count with a COUNT(DISTINCT) join on every row change. On a project with 10k+
-- members that recompute seq-scans user_projects per INSERT (~70ms avg for JoinProject
-- under load). member_count is now maintained incrementally (+1/-1) by the membership
-- triggers, and the score path reuses the stored member_count. The COUNT(DISTINCT) only
-- runs on the rare creation of a (project|event, church) leaderboard row.
--
-- Semantics change: leaderboard rows are now created/maintained even while a church has
-- 0 points. Reads filter score >= 1 (see queries/leaderboards.sql), so 0-point rows are
-- never returned to clients. regenerate_leaderboards() is updated to match (church rows
-- sourced from membership, not score_journal) and is re-run at the end as a backfill.

-- ==================== Project Membership Trigger ====================

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

    -- Hot path: adjust the existing row incrementally, no membership recount
    UPDATE leaderboard_project_churches
    SET total_points = total_points + (v_user_project_points * v_points_multiplier),
        member_count = GREATEST(member_count + v_points_multiplier, 0),
        score = calculate_average_score(
            total_points + (v_user_project_points * v_points_multiplier),
            GREATEST(member_count + v_points_multiplier, 0)),
        updated_at = NOW(),
        last_score_at = CASE WHEN TG_OP = 'INSERT'
            THEN GREATEST(COALESCE(last_score_at, v_last_score_at), v_last_score_at)
            ELSE last_score_at END
    WHERE project_id = v_project_id AND church_id = v_church_id;

    IF NOT FOUND AND TG_OP = 'INSERT' THEN
        -- First leaderboard row for this (project, church): compute the real count once.
        -- ON CONFLICT increments the stored count instead, so a concurrent creator wins safely.
        SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
        FROM users u
        INNER JOIN user_projects up ON up.user_id = u.id AND up.project_id = v_project_id
        WHERE u.church_id = v_church_id;

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
            member_count = leaderboard_project_churches.member_count + 1,
            score = calculate_average_score(leaderboard_project_churches.total_points + v_user_project_points, leaderboard_project_churches.member_count + 1),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, v_last_score_at), v_last_score_at);
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd
-- +goose StatementBegin

-- ==================== Event Membership Trigger ====================
-- Same incremental treatment for event churches, event teams and event superteams.
-- A user has at most one team per project (joinTeam removes prior memberships), so the
-- per-team loop touches each team/superteam at most once per membership change.

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
        UPDATE leaderboard_event_churches
        SET total_points = total_points + (v_user_event_points * v_points_multiplier),
            member_count = GREATEST(member_count + v_points_multiplier, 0),
            score = calculate_average_score(
                total_points + (v_user_event_points * v_points_multiplier),
                GREATEST(member_count + v_points_multiplier, 0)),
            updated_at = NOW(),
            last_score_at = CASE WHEN TG_OP = 'INSERT'
                THEN GREATEST(COALESCE(last_score_at, v_last_score_at), v_last_score_at)
                ELSE last_score_at END
        WHERE event_id = v_event_id AND church_id = v_church_id;

        IF NOT FOUND AND TG_OP = 'INSERT' THEN
            SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
            FROM users u
            INNER JOIN user_events ue ON ue.user_id = u.id AND ue.event_id = v_event_id
            WHERE u.church_id = v_church_id;

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
                member_count = leaderboard_event_churches.member_count + 1,
                score = calculate_average_score(leaderboard_event_churches.total_points + v_user_event_points, leaderboard_event_churches.member_count + 1),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_churches.last_score_at, v_last_score_at), v_last_score_at);
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
        UPDATE leaderboard_event_teams
        SET total_points = total_points + (v_user_event_points * v_points_multiplier),
            member_count = GREATEST(member_count + v_points_multiplier, 0),
            score = calculate_average_score(
                total_points + (v_user_event_points * v_points_multiplier),
                GREATEST(member_count + v_points_multiplier, 0)),
            updated_at = NOW(),
            last_score_at = CASE WHEN TG_OP = 'INSERT'
                THEN GREATEST(COALESCE(last_score_at, v_last_score_at), v_last_score_at)
                ELSE last_score_at END
        WHERE event_id = v_event_id AND team_id = team_record.team_id;

        IF NOT FOUND AND TG_OP = 'INSERT' THEN
            SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
            FROM team_members tm
            INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = v_event_id
            WHERE tm.team_id = team_record.team_id;

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
                member_count = leaderboard_event_teams.member_count + 1,
                score = calculate_average_score(leaderboard_event_teams.total_points + v_user_event_points, leaderboard_event_teams.member_count + 1),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_teams.last_score_at, v_last_score_at), v_last_score_at);
        END IF;

        -- Update superteam event leaderboard
        IF team_record.super_team_id IS NOT NULL THEN
            UPDATE leaderboard_event_superteams
            SET total_points = total_points + (v_user_event_points * v_points_multiplier),
                member_count = GREATEST(member_count + v_points_multiplier, 0),
                score = calculate_average_score(
                    total_points + (v_user_event_points * v_points_multiplier),
                    GREATEST(member_count + v_points_multiplier, 0)),
                updated_at = NOW(),
                last_score_at = CASE WHEN TG_OP = 'INSERT'
                    THEN GREATEST(COALESCE(last_score_at, v_last_score_at), v_last_score_at)
                    ELSE last_score_at END
            WHERE event_id = v_event_id AND super_team_id = team_record.super_team_id;

            IF NOT FOUND AND TG_OP = 'INSERT' THEN
                SELECT COUNT(DISTINCT tm.user_id)::INT INTO v_member_count
                FROM team_members tm
                INNER JOIN teams t ON t.id = tm.team_id
                INNER JOIN user_events ue ON ue.user_id = tm.user_id AND ue.event_id = v_event_id
                WHERE t.super_team_id = team_record.super_team_id;

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
                    member_count = leaderboard_event_superteams.member_count + 1,
                    score = calculate_average_score(leaderboard_event_superteams.total_points + v_user_event_points, leaderboard_event_superteams.member_count + 1),
                    updated_at = NOW(),
                    last_score_at = GREATEST(COALESCE(leaderboard_event_superteams.last_score_at, v_last_score_at), v_last_score_at);
            END IF;
        END IF;
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd
-- +goose StatementBegin

-- ==================== Score-Journal Church Helper ====================
-- Called on every score_journal insert. Reuses the stored member_count instead of
-- recomputing it; the membership triggers above keep that count current.

CREATE OR REPLACE FUNCTION update_church_leaderboard(
    p_church_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
DECLARE
    v_member_count INT;
BEGIN
    -- Update project church leaderboard
    IF p_project_id IS NOT NULL THEN
        UPDATE leaderboard_project_churches
        SET total_points = total_points + p_points_delta,
            score = calculate_average_score(total_points + p_points_delta, member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(last_score_at, p_score_at), p_score_at)
        WHERE project_id = p_project_id AND church_id = p_church_id;

        IF NOT FOUND THEN
            -- Row missing (score before any tracked membership): compute the count once
            SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
            FROM users u
            INNER JOIN user_projects up ON up.user_id = u.id AND up.project_id = p_project_id
            WHERE u.church_id = p_church_id;

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
                score = calculate_average_score(leaderboard_project_churches.total_points + p_points_delta, leaderboard_project_churches.member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;

    -- Update event church leaderboard
    IF p_event_id IS NOT NULL THEN
        UPDATE leaderboard_event_churches
        SET total_points = total_points + p_points_delta,
            score = calculate_average_score(total_points + p_points_delta, member_count),
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(last_score_at, p_score_at), p_score_at)
        WHERE event_id = p_event_id AND church_id = p_church_id;

        IF NOT FOUND THEN
            SELECT COUNT(DISTINCT u.id)::INT INTO v_member_count
            FROM users u
            INNER JOIN user_events ue ON ue.user_id = u.id AND ue.event_id = p_event_id
            WHERE u.church_id = p_church_id;

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
                score = calculate_average_score(leaderboard_event_churches.total_points + p_points_delta, leaderboard_event_churches.member_count),
                updated_at = NOW(),
                last_score_at = GREATEST(COALESCE(leaderboard_event_churches.last_score_at, p_score_at), p_score_at);
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd
-- +goose StatementBegin

-- ==================== Regenerate Function ====================
-- Church sections now source rows from membership (with points LEFT JOINed in) so that
-- every (project|event, church) with members gets a row with a correct member_count —
-- the invariant the incremental triggers rely on. Members' points only (a user who left
-- no longer contributes), matching what the membership triggers maintain.
-- Other sections are unchanged from 00080.

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
    -- Churches use AVERAGE over current members; rows exist for every church with
    -- members in the project, even at 0 points (reads filter score >= 1)
    WITH member_points AS (
        SELECT
            up.project_id,
            u.church_id,
            u.id AS user_id,
            COALESCE(SUM(sj.points), 0) AS points,
            MAX(sj.created_at) AS last_score_at
        FROM user_projects up
        INNER JOIN users u ON u.id = up.user_id
        LEFT JOIN score_journal sj ON sj.user_id = up.user_id AND sj.project_id = up.project_id
        WHERE u.church_id IS NOT NULL
        GROUP BY up.project_id, u.church_id, u.id
    )
    INSERT INTO leaderboard_project_churches (project_id, church_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        mp.project_id,
        mp.church_id,
        SUM(mp.points)::BIGINT AS total_points,
        COUNT(*)::INT AS member_count,
        calculate_average_score(SUM(mp.points)::BIGINT, COUNT(*)::INT) AS score,
        NOW() AS updated_at,
        MAX(mp.last_score_at) AS last_score_at
    FROM member_points mp
    GROUP BY mp.project_id, mp.church_id;
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
    -- Same membership-sourced shape as project churches
    WITH member_points AS (
        SELECT
            ue.event_id,
            u.church_id,
            u.id AS user_id,
            COALESCE(SUM(sj.points), 0) AS points,
            MAX(sj.created_at) AS last_score_at
        FROM user_events ue
        INNER JOIN users u ON u.id = ue.user_id
        LEFT JOIN score_journal sj ON sj.user_id = ue.user_id AND sj.event_id = ue.event_id
        WHERE u.church_id IS NOT NULL
        GROUP BY ue.event_id, u.church_id, u.id
    )
    INSERT INTO leaderboard_event_churches (event_id, church_id, total_points, member_count, score, updated_at, last_score_at)
    SELECT
        mp.event_id,
        mp.church_id,
        SUM(mp.points)::BIGINT AS total_points,
        COUNT(*)::INT AS member_count,
        calculate_average_score(SUM(mp.points)::BIGINT, COUNT(*)::INT) AS score,
        NOW() AS updated_at,
        MAX(mp.last_score_at) AS last_score_at
    FROM member_points mp
    GROUP BY mp.event_id, mp.church_id;
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

-- One-time backfill: rebuild leaderboards so existing church rows carry the member
-- counts the incremental triggers will maintain from here on.
-- +goose StatementBegin
SELECT regenerate_leaderboards();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the 00081 project membership trigger (recomputes member_count per change)

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

-- +goose StatementEnd
-- +goose StatementBegin

-- Restore the 00081 event membership trigger

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
-- +goose StatementBegin

-- Restore the 00080 church helper (recomputes member_count per score write)

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

-- +goose StatementEnd
-- +goose StatementBegin

-- Restore the 00080 regenerate function (church rows sourced from score_journal only)

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

-- Rebuild rows under the restored (score-journal-sourced) semantics so lingering
-- 0-point membership rows are removed.
-- +goose StatementBegin
SELECT regenerate_leaderboards();
-- +goose StatementEnd
