-- +goose Up
ALTER TABLE achievements ADD COLUMN sort_order INT NOT NULL DEFAULT 0;

-- Initialize sort_order based on created_at order within each project
WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY created_at) - 1 AS new_order
    FROM achievements
)
UPDATE achievements SET sort_order = ranked.new_order
FROM ranked WHERE achievements.id = ranked.id;

-- +goose Down
ALTER TABLE achievements DROP COLUMN sort_order;
