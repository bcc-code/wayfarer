-- +goose Up
ALTER TABLE quiz_predefined_answers ADD COLUMN description TEXT;
ALTER TABLE quiz_answer_translations ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE quiz_answer_translations DROP COLUMN description;
ALTER TABLE quiz_predefined_answers DROP COLUMN description;
