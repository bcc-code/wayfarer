-- +goose Up
-- +goose StatementBegin

-- Update trigger function to skip score journal entries for achievements with no points
CREATE OR REPLACE FUNCTION create_score_journal_entry_for_achievement()
RETURNS TRIGGER AS $$
DECLARE
    achievement_points INT;
    achievement_project_id CHAR(28);
    achievement_event_id CHAR(28);
    achievement_challenge_id CHAR(28);
    journal_id CHAR(28);
BEGIN
    -- Get achievement details
    SELECT points, project_id, event_id, challenge_id
    INTO achievement_points, achievement_project_id, achievement_event_id, achievement_challenge_id
    FROM achievements
    WHERE id = NEW.achievement_id;

    -- Skip journal entry if achievement has no points
    IF achievement_points IS NULL OR achievement_points = 0 THEN
        RETURN NEW;
    END IF;

    -- Generate journal entry ID (SJ prefix + 26 character ULID)
    journal_id := 'SJ' || UPPER(SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT) FROM 1 FOR 26));

    -- Create journal entry
    INSERT INTO score_journal (
        id,
        project_id,
        user_id,
        event_id,
        challenge_id,
        points,
        source_type,
        source_id,
        reason,
        created_at
    ) VALUES (
        journal_id,
        achievement_project_id,
        NEW.user_id,
        achievement_event_id,
        achievement_challenge_id,
        achievement_points,
        'ACHIEVEMENT',
        NEW.achievement_id,
        NULL,
        NEW.achieved_at
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original trigger function that creates entries for all achievements
CREATE OR REPLACE FUNCTION create_score_journal_entry_for_achievement()
RETURNS TRIGGER AS $$
DECLARE
    achievement_points INT;
    achievement_project_id CHAR(28);
    achievement_event_id CHAR(28);
    achievement_challenge_id CHAR(28);
    journal_id CHAR(28);
BEGIN
    -- Get achievement details
    SELECT points, project_id, event_id, challenge_id
    INTO achievement_points, achievement_project_id, achievement_event_id, achievement_challenge_id
    FROM achievements
    WHERE id = NEW.achievement_id;

    -- Generate journal entry ID (SJ prefix + 26 character ULID)
    journal_id := 'SJ' || UPPER(SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT) FROM 1 FOR 26));

    -- Create journal entry
    INSERT INTO score_journal (
        id,
        project_id,
        user_id,
        event_id,
        challenge_id,
        points,
        source_type,
        source_id,
        reason,
        created_at
    ) VALUES (
        journal_id,
        achievement_project_id,
        NEW.user_id,
        achievement_event_id,
        achievement_challenge_id,
        achievement_points,
        'ACHIEVEMENT',
        NEW.achievement_id,
        NULL,
        NEW.achieved_at
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd
