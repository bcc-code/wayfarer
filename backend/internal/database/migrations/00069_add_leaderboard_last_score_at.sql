-- +goose Up
-- +goose StatementBegin

-- Migration: Add last_score_at column to leaderboard tables for tie-breaking
-- When scores are tied, entries will be sorted by most recent score change (newest first)

-- ==================== Add Columns to Leaderboard Tables ====================

ALTER TABLE leaderboard_project_persons ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_project_teams ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_project_superteams ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_project_churches ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_event_persons ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_event_teams ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_event_superteams ADD COLUMN last_score_at TIMESTAMPTZ;
ALTER TABLE leaderboard_event_churches ADD COLUMN last_score_at TIMESTAMPTZ;

-- ==================== Update Indexes for New Sort Order ====================
-- New sort: score DESC, last_score_at DESC NULLS LAST, entity_id

DROP INDEX IF EXISTS idx_leaderboard_project_persons_score;
CREATE INDEX idx_leaderboard_project_persons_score
    ON leaderboard_project_persons(project_id, score DESC, last_score_at DESC NULLS LAST, user_id);

DROP INDEX IF EXISTS idx_leaderboard_project_teams_score;
CREATE INDEX idx_leaderboard_project_teams_score
    ON leaderboard_project_teams(project_id, score DESC, last_score_at DESC NULLS LAST, team_id);

DROP INDEX IF EXISTS idx_leaderboard_project_superteams_score;
CREATE INDEX idx_leaderboard_project_superteams_score
    ON leaderboard_project_superteams(project_id, score DESC, last_score_at DESC NULLS LAST, super_team_id);

DROP INDEX IF EXISTS idx_leaderboard_project_churches_score;
CREATE INDEX idx_leaderboard_project_churches_score
    ON leaderboard_project_churches(project_id, score DESC, last_score_at DESC NULLS LAST, church_id);

DROP INDEX IF EXISTS idx_leaderboard_event_persons_score;
CREATE INDEX idx_leaderboard_event_persons_score
    ON leaderboard_event_persons(event_id, score DESC, last_score_at DESC NULLS LAST, user_id);

DROP INDEX IF EXISTS idx_leaderboard_event_teams_score;
CREATE INDEX idx_leaderboard_event_teams_score
    ON leaderboard_event_teams(event_id, score DESC, last_score_at DESC NULLS LAST, team_id);

DROP INDEX IF EXISTS idx_leaderboard_event_superteams_score;
CREATE INDEX idx_leaderboard_event_superteams_score
    ON leaderboard_event_superteams(event_id, score DESC, last_score_at DESC NULLS LAST, super_team_id);

DROP INDEX IF EXISTS idx_leaderboard_event_churches_score;
CREATE INDEX idx_leaderboard_event_churches_score
    ON leaderboard_event_churches(event_id, score DESC, last_score_at DESC NULLS LAST, church_id);

-- ==================== Update Helper Functions ====================

-- Update person leaderboard helper
CREATE OR REPLACE FUNCTION update_person_leaderboard(
    p_user_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    -- Update project person leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_user_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, user_id)
        DO UPDATE SET
            score = leaderboard_project_persons.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_persons.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event person leaderboard
    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at, last_score_at)
        VALUES (p_event_id, p_user_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (event_id, user_id)
        DO UPDATE SET
            score = leaderboard_event_persons.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_event_persons.last_score_at, p_score_at), p_score_at);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Update team leaderboard helper
CREATE OR REPLACE FUNCTION update_team_leaderboard(
    p_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    -- Update project team leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_team_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, team_id)
        DO UPDATE SET
            score = leaderboard_project_teams.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_teams.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event team leaderboard
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

-- Update superteam leaderboard helper
CREATE OR REPLACE FUNCTION update_superteam_leaderboard(
    p_super_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    -- Update project superteam leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_super_team_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, super_team_id)
        DO UPDATE SET
            score = leaderboard_project_superteams.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_superteams.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event superteam leaderboard
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

-- Update church leaderboard helper
CREATE OR REPLACE FUNCTION update_church_leaderboard(
    p_church_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT,
    p_score_at TIMESTAMPTZ
) RETURNS VOID AS $$
BEGIN
    -- Update project church leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at, last_score_at)
        VALUES (p_project_id, p_church_id, p_points_delta, NOW(), p_score_at)
        ON CONFLICT (project_id, church_id)
        DO UPDATE SET
            score = leaderboard_project_churches.score + p_points_delta,
            updated_at = NOW(),
            last_score_at = GREATEST(COALESCE(leaderboard_project_churches.last_score_at, p_score_at), p_score_at);
    END IF;

    -- Update event church leaderboard
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

-- ==================== Update Trigger Function ====================

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_from_score_journal()
RETURNS TRIGGER AS $$
DECLARE
    points_delta BIGINT;
    target_user_id CHAR(28);
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    score_timestamp TIMESTAMPTZ;
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
        score_timestamp := COALESCE(NEW.created_at, NOW());
    ELSIF (TG_OP = 'UPDATE') THEN
        points_delta := NEW.points - OLD.points;
        target_user_id := NEW.user_id;
        target_project_id := NEW.project_id;
        target_event_id := NEW.event_id;
        score_timestamp := COALESCE(NEW.created_at, NOW());
    ELSIF (TG_OP = 'DELETE') THEN
        points_delta := -OLD.points;
        target_user_id := OLD.user_id;
        target_project_id := OLD.project_id;
        target_event_id := OLD.event_id;
        -- For deletes, we don't update last_score_at (keep existing value)
        score_timestamp := NULL;
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
        points_delta,
        score_timestamp
    );

    -- Update church leaderboard
    IF user_church_id IS NOT NULL THEN
        PERFORM update_church_leaderboard(
            user_church_id,
            target_project_id,
            target_event_id,
            points_delta,
            score_timestamp
        );
    END IF;

    -- Update team leaderboard
    IF user_team_id IS NOT NULL THEN
        PERFORM update_team_leaderboard(
            user_team_id,
            target_project_id,
            target_event_id,
            points_delta,
            score_timestamp
        );
    END IF;

    -- Update superteam leaderboard
    IF user_super_team_id IS NOT NULL THEN
        PERFORM update_superteam_leaderboard(
            user_super_team_id,
            target_project_id,
            target_event_id,
            points_delta,
            score_timestamp
        );
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ==================== Update Regenerate Function ====================

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

    -- ==================== Regenerate Project SuperTeam Leaderboards ====================
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

    -- ==================== Regenerate Project Church Leaderboards ====================
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

    -- ==================== Regenerate Event Person Leaderboards ====================
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

    -- ==================== Regenerate Event SuperTeam Leaderboards ====================
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

    -- ==================== Regenerate Event Church Leaderboards ====================
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

-- Drop columns from leaderboard tables
ALTER TABLE leaderboard_project_persons DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_project_teams DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_project_superteams DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_project_churches DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_event_persons DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_event_teams DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_event_superteams DROP COLUMN IF EXISTS last_score_at;
ALTER TABLE leaderboard_event_churches DROP COLUMN IF EXISTS last_score_at;

-- Restore original indexes
DROP INDEX IF EXISTS idx_leaderboard_project_persons_score;
CREATE INDEX idx_leaderboard_project_persons_score
    ON leaderboard_project_persons(project_id, score DESC, user_id);

DROP INDEX IF EXISTS idx_leaderboard_project_teams_score;
CREATE INDEX idx_leaderboard_project_teams_score
    ON leaderboard_project_teams(project_id, score DESC, team_id);

DROP INDEX IF EXISTS idx_leaderboard_project_superteams_score;
CREATE INDEX idx_leaderboard_project_superteams_score
    ON leaderboard_project_superteams(project_id, score DESC, super_team_id);

DROP INDEX IF EXISTS idx_leaderboard_project_churches_score;
CREATE INDEX idx_leaderboard_project_churches_score
    ON leaderboard_project_churches(project_id, score DESC, church_id);

DROP INDEX IF EXISTS idx_leaderboard_event_persons_score;
CREATE INDEX idx_leaderboard_event_persons_score
    ON leaderboard_event_persons(event_id, score DESC, user_id);

DROP INDEX IF EXISTS idx_leaderboard_event_teams_score;
CREATE INDEX idx_leaderboard_event_teams_score
    ON leaderboard_event_teams(event_id, score DESC, team_id);

DROP INDEX IF EXISTS idx_leaderboard_event_superteams_score;
CREATE INDEX idx_leaderboard_event_superteams_score
    ON leaderboard_event_superteams(event_id, score DESC, super_team_id);

DROP INDEX IF EXISTS idx_leaderboard_event_churches_score;
CREATE INDEX idx_leaderboard_event_churches_score
    ON leaderboard_event_churches(event_id, score DESC, church_id);

-- Restore original helper functions (4 parameters, no timestamp)
CREATE OR REPLACE FUNCTION update_person_leaderboard(
    p_user_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
        VALUES (p_project_id, p_user_id, p_points_delta, NOW())
        ON CONFLICT (project_id, user_id)
        DO UPDATE SET
            score = leaderboard_project_persons.score + p_points_delta,
            updated_at = NOW();
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_persons (event_id, user_id, score, updated_at)
        VALUES (p_event_id, p_user_id, p_points_delta, NOW())
        ON CONFLICT (event_id, user_id)
        DO UPDATE SET
            score = leaderboard_event_persons.score + p_points_delta,
            updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_team_leaderboard(
    p_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at)
        VALUES (p_project_id, p_team_id, p_points_delta, NOW())
        ON CONFLICT (project_id, team_id)
        DO UPDATE SET
            score = leaderboard_project_teams.score + p_points_delta,
            updated_at = NOW();
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_teams (event_id, team_id, score, updated_at)
        VALUES (p_event_id, p_team_id, p_points_delta, NOW())
        ON CONFLICT (event_id, team_id)
        DO UPDATE SET
            score = leaderboard_event_teams.score + p_points_delta,
            updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_superteam_leaderboard(
    p_super_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at)
        VALUES (p_project_id, p_super_team_id, p_points_delta, NOW())
        ON CONFLICT (project_id, super_team_id)
        DO UPDATE SET
            score = leaderboard_project_superteams.score + p_points_delta,
            updated_at = NOW();
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_superteams (event_id, super_team_id, score, updated_at)
        VALUES (p_event_id, p_super_team_id, p_points_delta, NOW())
        ON CONFLICT (event_id, super_team_id)
        DO UPDATE SET
            score = leaderboard_event_superteams.score + p_points_delta,
            updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_church_leaderboard(
    p_church_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at)
        VALUES (p_project_id, p_church_id, p_points_delta, NOW())
        ON CONFLICT (project_id, church_id)
        DO UPDATE SET
            score = leaderboard_project_churches.score + p_points_delta,
            updated_at = NOW();
    END IF;

    IF p_event_id IS NOT NULL THEN
        INSERT INTO leaderboard_event_churches (event_id, church_id, score, updated_at)
        VALUES (p_event_id, p_church_id, p_points_delta, NOW())
        ON CONFLICT (event_id, church_id)
        DO UPDATE SET
            score = leaderboard_event_churches.score + p_points_delta,
            updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Restore original trigger function
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

    SELECT u.church_id, tm.team_id, t.super_team_id
    INTO user_church_id, user_team_id, user_super_team_id
    FROM users u
    LEFT JOIN team_members tm ON tm.user_id = u.id
    LEFT JOIN teams t ON t.id = tm.team_id AND t.project_id = target_project_id
    WHERE u.id = target_user_id;

    PERFORM update_person_leaderboard(
        target_user_id,
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

    IF user_team_id IS NOT NULL THEN
        PERFORM update_team_leaderboard(
            user_team_id,
            target_project_id,
            target_event_id,
            points_delta
        );
    END IF;

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

-- Restore original regenerate function
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
        sj.project_id,
        sj.user_id,
        SUM(sj.points) AS score,
        NOW() AS updated_at
    FROM score_journal sj
    GROUP BY sj.project_id, sj.user_id
    HAVING SUM(sj.points) > 0;
    GET DIAGNOSTICS project_persons_count = ROW_COUNT;

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
