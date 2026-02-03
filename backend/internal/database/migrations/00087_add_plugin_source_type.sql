-- +goose Up
ALTER TABLE score_journal
DROP CONSTRAINT score_journal_source_type_check;

ALTER TABLE score_journal
ADD CONSTRAINT score_journal_source_type_check
CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL', 'PLUGIN'));

-- +goose Down
ALTER TABLE score_journal
DROP CONSTRAINT score_journal_source_type_check;

ALTER TABLE score_journal
ADD CONSTRAINT score_journal_source_type_check
CHECK (source_type IN ('ACHIEVEMENT', 'MANUAL'));
