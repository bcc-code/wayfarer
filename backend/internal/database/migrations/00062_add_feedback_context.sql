-- +goose Up
ALTER TABLE user_feedback ADD COLUMN locale TEXT;
ALTER TABLE user_feedback ADD COLUMN project_id CHAR(28) REFERENCES projects(id) ON DELETE SET NULL;
ALTER TABLE user_feedback ADD COLUMN timezone TEXT;

-- +goose Down
ALTER TABLE user_feedback DROP COLUMN locale;
ALTER TABLE user_feedback DROP COLUMN project_id;
ALTER TABLE user_feedback DROP COLUMN timezone;
