-- +goose Up
-- +goose StatementBegin

-- Add score_journal_id to quiz_responses to track bet result journal entries
ALTER TABLE quiz_responses
ADD COLUMN score_journal_id CHAR(28) REFERENCES score_journal(id) ON DELETE SET NULL;

-- Add BET to score_journal source_type constraint
ALTER TABLE score_journal DROP CONSTRAINT IF EXISTS score_journal_source_type_check;
ALTER TABLE score_journal ADD CONSTRAINT score_journal_source_type_check
    CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL', 'QUIZ', 'PLUGIN', 'BET'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove score_journal_id from quiz_responses
ALTER TABLE quiz_responses DROP COLUMN IF EXISTS score_journal_id;

-- Revert score_journal source_type constraint
ALTER TABLE score_journal DROP CONSTRAINT IF EXISTS score_journal_source_type_check;
ALTER TABLE score_journal ADD CONSTRAINT score_journal_source_type_check
    CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL', 'QUIZ', 'PLUGIN'));

-- +goose StatementEnd
