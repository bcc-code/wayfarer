-- +goose Up
-- +goose StatementBegin

-- Migration: Add triggers to maintain leaderboard tables automatically
-- These triggers keep the leaderboard tables in sync with achievement changes

-- ==================== Helper Function: Update Person Leaderboards ====================

CREATE OR REPLACE FUNCTION update_person_leaderboard(
    p_user_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    -- Update project person leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_persons (project_id, user_id, score, updated_at)
        VALUES (p_project_id, p_user_id, p_points_delta, NOW())
        ON CONFLICT (project_id, user_id)
        DO UPDATE SET
            score = leaderboard_project_persons.score + p_points_delta,
            updated_at = NOW();
    END IF;

    -- Update event person leaderboard
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

-- ==================== Helper Function: Update Team Leaderboards ====================

CREATE OR REPLACE FUNCTION update_team_leaderboard(
    p_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    -- Update project team leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_teams (project_id, team_id, score, updated_at)
        VALUES (p_project_id, p_team_id, p_points_delta, NOW())
        ON CONFLICT (project_id, team_id)
        DO UPDATE SET
            score = leaderboard_project_teams.score + p_points_delta,
            updated_at = NOW();
    END IF;

    -- Update event team leaderboard
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

-- ==================== Helper Function: Update SuperTeam Leaderboards ====================

CREATE OR REPLACE FUNCTION update_superteam_leaderboard(
    p_super_team_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    -- Update project superteam leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_superteams (project_id, super_team_id, score, updated_at)
        VALUES (p_project_id, p_super_team_id, p_points_delta, NOW())
        ON CONFLICT (project_id, super_team_id)
        DO UPDATE SET
            score = leaderboard_project_superteams.score + p_points_delta,
            updated_at = NOW();
    END IF;

    -- Update event superteam leaderboard
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

-- ==================== Helper Function: Update Church Leaderboards ====================

CREATE OR REPLACE FUNCTION update_church_leaderboard(
    p_church_id CHAR(28),
    p_project_id CHAR(28),
    p_event_id CHAR(28),
    p_points_delta BIGINT
) RETURNS VOID AS $$
BEGIN
    -- Update project church leaderboard
    IF p_project_id IS NOT NULL THEN
        INSERT INTO leaderboard_project_churches (project_id, church_id, score, updated_at)
        VALUES (p_project_id, p_church_id, p_points_delta, NOW())
        ON CONFLICT (project_id, church_id)
        DO UPDATE SET
            score = leaderboard_project_churches.score + p_points_delta,
            updated_at = NOW();
    END IF;

    -- Update event church leaderboard
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

-- ==================== Trigger 1: User Achievements ====================

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_user_achievement()
RETURNS TRIGGER AS $$
DECLARE
    achievement_points BIGINT;
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    user_church_id CHAR(28);
    points_delta BIGINT;
BEGIN
    -- Determine operation and get achievement details
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

    -- Get user's church for church leaderboards
    SELECT church_id INTO user_church_id
    FROM users
    WHERE id = COALESCE(NEW.user_id, OLD.user_id);

    -- Update person leaderboard
    PERFORM update_person_leaderboard(
        COALESCE(NEW.user_id, OLD.user_id),
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

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_achievement_leaderboard
AFTER INSERT OR DELETE ON user_achievements
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_user_achievement();

-- ==================== Trigger 2: Team Achievements ====================

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_team_achievement()
RETURNS TRIGGER AS $$
DECLARE
    achievement_points BIGINT;
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    points_delta BIGINT;
BEGIN
    -- Determine operation and get achievement details
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

    -- Update team leaderboard
    PERFORM update_team_leaderboard(
        COALESCE(NEW.team_id, OLD.team_id),
        target_project_id,
        target_event_id,
        points_delta
    );

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_team_achievement_leaderboard
AFTER INSERT OR DELETE ON team_achievements
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_team_achievement();

-- ==================== Trigger 3: SuperTeam Achievements ====================

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_superteam_achievement()
RETURNS TRIGGER AS $$
DECLARE
    achievement_points BIGINT;
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    points_delta BIGINT;
BEGIN
    -- Determine operation and get achievement details
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

    -- Update superteam leaderboard
    PERFORM update_superteam_leaderboard(
        COALESCE(NEW.super_team_id, OLD.super_team_id),
        target_project_id,
        target_event_id,
        points_delta
    );

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_superteam_achievement_leaderboard
AFTER INSERT OR DELETE ON super_team_achievements
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_superteam_achievement();

-- ==================== Trigger 4: Score Adjustments ====================

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_score_adjustment()
RETURNS TRIGGER AS $$
DECLARE
    old_points BIGINT;
    new_points BIGINT;
    points_delta BIGINT;
    target_project_id CHAR(28);
    target_event_id CHAR(28);
BEGIN
    -- Determine operation and calculate delta
    IF (TG_OP = 'INSERT') THEN
        points_delta := NEW.points;
        target_project_id := NEW.project_id;
        target_event_id := NEW.event_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        points_delta := NEW.points - OLD.points;
        target_project_id := NEW.project_id;
        target_event_id := NEW.event_id;
    ELSIF (TG_OP = 'DELETE') THEN
        points_delta := -OLD.points;
        target_project_id := OLD.project_id;
        target_event_id := OLD.event_id;
    ELSE
        RETURN NULL;
    END IF;

    -- Update appropriate leaderboard based on entity_type
    IF COALESCE(NEW.entity_type, OLD.entity_type) = 'USER' THEN
        PERFORM update_person_leaderboard(
            COALESCE(NEW.entity_id, OLD.entity_id),
            target_project_id,
            target_event_id,
            points_delta
        );
    ELSIF COALESCE(NEW.entity_type, OLD.entity_type) = 'TEAM' THEN
        PERFORM update_team_leaderboard(
            COALESCE(NEW.entity_id, OLD.entity_id),
            target_project_id,
            target_event_id,
            points_delta
        );
    ELSIF COALESCE(NEW.entity_type, OLD.entity_type) = 'SUPER_TEAM' THEN
        PERFORM update_superteam_leaderboard(
            COALESCE(NEW.entity_id, OLD.entity_id),
            target_project_id,
            target_event_id,
            points_delta
        );
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_score_adjustment_leaderboard
AFTER INSERT OR UPDATE OR DELETE ON score_adjustments
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_score_adjustment();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_score_adjustment_leaderboard ON score_adjustments;
DROP TRIGGER IF EXISTS trigger_superteam_achievement_leaderboard ON super_team_achievements;
DROP TRIGGER IF EXISTS trigger_team_achievement_leaderboard ON team_achievements;
DROP TRIGGER IF EXISTS trigger_user_achievement_leaderboard ON user_achievements;

-- Drop functions
DROP FUNCTION IF EXISTS trigger_update_leaderboard_score_adjustment();
DROP FUNCTION IF EXISTS trigger_update_leaderboard_superteam_achievement();
DROP FUNCTION IF EXISTS trigger_update_leaderboard_team_achievement();
DROP FUNCTION IF EXISTS trigger_update_leaderboard_user_achievement();
DROP FUNCTION IF EXISTS update_church_leaderboard(CHAR(28), CHAR(28), CHAR(28), BIGINT);
DROP FUNCTION IF EXISTS update_superteam_leaderboard(CHAR(28), CHAR(28), CHAR(28), BIGINT);
DROP FUNCTION IF EXISTS update_team_leaderboard(CHAR(28), CHAR(28), CHAR(28), BIGINT);
DROP FUNCTION IF EXISTS update_person_leaderboard(CHAR(28), CHAR(28), CHAR(28), BIGINT);

-- +goose StatementEnd
