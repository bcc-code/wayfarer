-- +goose Up
-- +goose StatementBegin

-- Add points value to quiz_questions
ALTER TABLE quiz_questions ADD COLUMN points INT;

-- Add points earned for individual responses
ALTER TABLE quiz_responses ADD COLUMN points_earned INT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE quiz_responses DROP COLUMN points_earned;
ALTER TABLE quiz_questions DROP COLUMN points;

-- +goose StatementEnd
