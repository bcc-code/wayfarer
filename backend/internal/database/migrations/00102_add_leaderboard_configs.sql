-- +goose Up
-- +goose StatementBegin

CREATE TABLE leaderboard_configs (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^LC[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('PERSONS', 'TEAMS', 'SUPERTEAMS', 'CHURCHES')),
    filter JSONB,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_leaderboard_configs_project ON leaderboard_configs(project_id);
CREATE INDEX idx_leaderboard_configs_event ON leaderboard_configs(event_id);

CREATE TRIGGER update_leaderboard_configs_updated_at BEFORE UPDATE ON leaderboard_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS leaderboard_configs;

-- +goose StatementEnd
