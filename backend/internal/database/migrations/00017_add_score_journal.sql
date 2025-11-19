-- +goose Up
-- +goose StatementBegin

-- Create score_journal table to track all point awards
-- Points are always awarded in context of a project, optionally an event and/or challenge
CREATE TABLE score_journal (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SJ[0-9A-Z]{26}$'),

    -- Required relationships
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Optional relationships
    event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL,
    challenge_id CHAR(28) REFERENCES challenges(id) ON DELETE SET NULL,

    -- Points and source tracking
    points INT NOT NULL,
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL')),
    source_id CHAR(28),  -- achievement_id if source_type is ACHIEVEMENT

    -- Metadata
    reason TEXT,
    awarded_by CHAR(28) REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Create indexes for performance
CREATE INDEX idx_score_journal_project ON score_journal(project_id);
CREATE INDEX idx_score_journal_user ON score_journal(user_id);
CREATE INDEX idx_score_journal_event ON score_journal(event_id);
CREATE INDEX idx_score_journal_challenge ON score_journal(challenge_id);
CREATE INDEX idx_score_journal_source ON score_journal(source_type, source_id);
CREATE INDEX idx_score_journal_time ON score_journal(created_at);

-- Create trigger function to auto-populate journal when achievements are awarded
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
    -- For now, use a simple approach - in production you'd want to generate a proper ULID
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

-- Attach trigger to user_achievements table
CREATE TRIGGER trigger_create_score_journal_entry
AFTER INSERT ON user_achievements
FOR EACH ROW
EXECUTE FUNCTION create_score_journal_entry_for_achievement();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_create_score_journal_entry ON user_achievements;
DROP FUNCTION IF EXISTS create_score_journal_entry_for_achievement();

-- Drop indexes
DROP INDEX IF EXISTS idx_score_journal_time;
DROP INDEX IF EXISTS idx_score_journal_source;
DROP INDEX IF EXISTS idx_score_journal_challenge;
DROP INDEX IF EXISTS idx_score_journal_event;
DROP INDEX IF EXISTS idx_score_journal_user;
DROP INDEX IF EXISTS idx_score_journal_project;

-- Drop table
DROP TABLE IF EXISTS score_journal;

-- +goose StatementEnd
