-- +goose Up
ALTER TABLE projects ADD COLUMN rules TEXT;
ALTER TABLE project_translations ADD COLUMN rules TEXT;

-- +goose Down
ALTER TABLE projects DROP COLUMN rules;
ALTER TABLE project_translations DROP COLUMN rules;
