-- +goose Up
-- +goose StatementBegin

-- Fix: Leaderboard trigger now updates ALL teams a user belongs to
-- Previously, when a user was in multiple teams in the same project,
-- the trigger would only update ONE random team because SELECT ... INTO
-- only captures a single row.
--
-- This migration changes the trigger to loop through all teams the user
-- belongs to in the target project.

CREATE OR REPLACE FUNCTION trigger_update_leaderboard_from_score_journal()
RETURNS TRIGGER AS $$
DECLARE
    points_delta BIGINT;
    target_user_id CHAR(28);
    target_project_id CHAR(28);
    target_event_id CHAR(28);
    score_timestamp TIMESTAMPTZ;
    user_church_id CHAR(28);
    user_exists BOOLEAN;
    team_record RECORD;
    processed_super_teams CHAR(28)[];
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

    -- Check if user exists and get their church_id
    SELECT TRUE, u.church_id
    INTO user_exists, user_church_id
    FROM users u
    WHERE u.id = target_user_id;

    -- If user doesn't exist, skip leaderboard updates
    IF NOT COALESCE(user_exists, FALSE) THEN
        RETURN NULL;
    END IF;

    -- Update person leaderboard (always single entry per user)
    PERFORM update_person_leaderboard(
        target_user_id,
        target_project_id,
        target_event_id,
        points_delta,
        score_timestamp
    );

    -- Update church leaderboard (always single entry per user's church)
    IF user_church_id IS NOT NULL THEN
        PERFORM update_church_leaderboard(
            user_church_id,
            target_project_id,
            target_event_id,
            points_delta,
            score_timestamp
        );
    END IF;

    -- Initialize array to track processed super_teams (to avoid duplicates)
    processed_super_teams := ARRAY[]::CHAR(28)[];

    -- Loop through ALL teams the user belongs to in the target project
    FOR team_record IN
        SELECT t.id AS team_id, t.super_team_id
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id AND t.project_id = target_project_id
        WHERE tm.user_id = target_user_id
    LOOP
        -- Update team leaderboard for each team
        PERFORM update_team_leaderboard(
            team_record.team_id,
            target_project_id,
            target_event_id,
            points_delta,
            score_timestamp
        );

        -- Update superteam leaderboard if the team has one and we haven't processed it yet
        IF team_record.super_team_id IS NOT NULL
           AND NOT (team_record.super_team_id = ANY(processed_super_teams)) THEN
            PERFORM update_superteam_leaderboard(
                team_record.super_team_id,
                target_project_id,
                target_event_id,
                points_delta,
                score_timestamp
            );
            -- Mark this super_team as processed
            processed_super_teams := array_append(processed_super_teams, team_record.super_team_id);
        END IF;
    END LOOP;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore previous trigger function (from migration 00069)
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

-- +goose StatementEnd
