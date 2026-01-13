-- +goose Up
-- +goose StatementBegin

-- Fix: Check if user exists before updating leaderboards
-- This prevents FK constraint violations when score_journal entries reference
-- users that have been deleted or don't exist

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
    user_exists BOOLEAN;
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

    -- Check if user exists and get their church, team, and superteam for aggregation
    SELECT
        TRUE,
        u.church_id,
        tm.team_id,
        t.super_team_id
    INTO user_exists, user_church_id, user_team_id, user_super_team_id
    FROM users u
    LEFT JOIN team_members tm ON tm.user_id = u.id
    LEFT JOIN teams t ON t.id = tm.team_id AND t.project_id = target_project_id
    WHERE u.id = target_user_id;

    -- If user doesn't exist, skip leaderboard updates
    IF NOT COALESCE(user_exists, FALSE) THEN
        RETURN NULL;
    END IF;

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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original function without user existence check
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

-- +goose StatementEnd
