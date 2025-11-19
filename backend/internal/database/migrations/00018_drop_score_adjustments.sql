-- +goose Up
-- +goose StatementBegin

-- Drop score_adjustments table
-- Points are now tracked exclusively in score_journal
DROP TABLE IF EXISTS score_adjustments;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate score_adjustments table
CREATE TABLE IF NOT EXISTS score_adjustments (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SA[0-9A-Z]{26}$'),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('USER', 'TEAM', 'SUPER_TEAM')),
    entity_id CHAR(28) NOT NULL,
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    points INT NOT NULL,
    reason TEXT,
    adjusted_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_score_adjustments_entity ON score_adjustments(entity_type, entity_id);
CREATE INDEX idx_score_adjustments_project ON score_adjustments(project_id);
CREATE INDEX idx_score_adjustments_time ON score_adjustments(created_at);

-- +goose StatementEnd
