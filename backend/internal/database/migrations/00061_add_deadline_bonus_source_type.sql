-- +goose Up
-- +goose StatementBegin

-- Add PLUGIN to score_journal source_type constraint
ALTER TABLE score_journal DROP CONSTRAINT IF EXISTS score_journal_source_type_check;
ALTER TABLE score_journal ADD CONSTRAINT score_journal_source_type_check
    CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL', 'QUIZ', 'PLUGIN'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to constraint without PLUGIN
ALTER TABLE score_journal DROP CONSTRAINT IF EXISTS score_journal_source_type_check;
ALTER TABLE score_journal ADD CONSTRAINT score_journal_source_type_check
    CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL', 'QUIZ'));

-- +goose StatementEnd
