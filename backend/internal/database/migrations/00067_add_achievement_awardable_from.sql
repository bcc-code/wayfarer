-- +goose Up
ALTER TABLE achievements ADD COLUMN awardable_from TIMESTAMPTZ;

CREATE INDEX idx_achievements_awardable_from ON achievements(awardable_from)
WHERE awardable_from IS NOT NULL;

COMMENT ON COLUMN achievements.awardable_from IS 'Earliest time the achievement can be awarded. NULL means always awardable.';

-- +goose Down
DROP INDEX IF EXISTS idx_achievements_awardable_from;
ALTER TABLE achievements DROP COLUMN IF EXISTS awardable_from;
