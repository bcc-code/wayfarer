-- +goose Up
CREATE INDEX idx_user_content_progress_user_achievement
ON user_content_progress (user_id, achievement_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_content_progress_user_achievement;
