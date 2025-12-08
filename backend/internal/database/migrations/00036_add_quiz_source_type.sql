-- +goose Up
-- +goose StatementBegin

-- Add QUIZ to score_journal source_type constraint
ALTER TABLE score_journal DROP CONSTRAINT IF EXISTS score_journal_source_type_check;
ALTER TABLE score_journal ADD CONSTRAINT score_journal_source_type_check
    CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL', 'QUIZ'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to original constraint without QUIZ
ALTER TABLE score_journal DROP CONSTRAINT IF EXISTS score_journal_source_type_check;
ALTER TABLE score_journal ADD CONSTRAINT score_journal_source_type_check
    CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL'));

-- +goose StatementEnd
