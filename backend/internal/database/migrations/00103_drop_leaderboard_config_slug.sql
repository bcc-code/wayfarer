-- +goose Up
-- +goose StatementBegin

-- 00102 was edited after it had already been applied: code review dropped
-- `slug` (and its unique index) from the CREATE TABLE, and tightened
-- created_at/updated_at to NOT NULL. Goose records 00102 as applied, so it will
-- never re-run to correct a database that migrated before that edit — such a
-- database still has `slug NOT NULL` with no default, and every insert fails
-- with "null value in column slug violates not-null constraint".
--
-- Every statement here is conditional or idempotent, so this is a no-op on a
-- database created from the amended 00102.

DROP INDEX IF EXISTS idx_leaderboard_configs_project_slug;

ALTER TABLE leaderboard_configs DROP COLUMN IF EXISTS slug;

-- Safe on rows that already exist: both columns have defaulted to now() since
-- the table was created, so neither can hold a NULL to trip over.
ALTER TABLE leaderboard_configs ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE leaderboard_configs ALTER COLUMN updated_at SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- `slug` is deliberately not restored. It had no consumers — there was no
-- bySlug lookup anywhere — and a NOT NULL column cannot be added back to a
-- populated table without inventing values for it. Only the nullability
-- tightening is reversible.

ALTER TABLE leaderboard_configs ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE leaderboard_configs ALTER COLUMN updated_at DROP NOT NULL;

-- +goose StatementEnd
