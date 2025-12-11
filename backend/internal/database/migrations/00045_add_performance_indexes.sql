-- +goose Up
-- +goose StatementBegin

-- Composite index for score_journal leaderboard aggregation queries
-- Covers: LEFT JOIN score_journal sj ON sj.user_id = fu.id AND sj.project_id = @projectid
CREATE INDEX IF NOT EXISTS idx_score_journal_project_user
ON score_journal(project_id, user_id) INCLUDE (points);

-- Composite index for church filtering in leaderboard queries
-- Covers: WHERE c.country = @country AND c.category = @churchcategory
CREATE INDEX IF NOT EXISTS idx_churches_country_category
ON churches(country, category);

-- Composite index for consent history lookups
-- Covers: DISTINCT ON (user_id, consent_key) ORDER BY user_id, consent_key, occurred_at DESC
CREATE INDEX IF NOT EXISTS idx_user_consent_history_user_key_time
ON user_consent_history(user_id, consent_key, occurred_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_user_consent_history_user_key_time;
DROP INDEX IF EXISTS idx_churches_country_category;
DROP INDEX IF EXISTS idx_score_journal_project_user;

-- +goose StatementEnd
