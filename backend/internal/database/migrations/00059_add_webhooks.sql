-- +goose Up
-- +goose StatementBegin

CREATE TABLE webhooks (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^WH[0-9A-Z]{26}$'),
    project_id CHAR(28) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(2000) NOT NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('external_content_event', 'points_awarded')),
    include_user_data BOOLEAN DEFAULT true NOT NULL,
    include_event_data BOOLEAN DEFAULT true NOT NULL,
    active BOOLEAN DEFAULT true NOT NULL,
    secret VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_webhooks_project ON webhooks(project_id);
CREATE INDEX idx_webhooks_event_active ON webhooks(project_id, event_type) WHERE active = true;

CREATE TABLE webhook_logs (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^WL[0-9A-Z]{26}$'),
    webhook_id CHAR(28) NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    request_payload JSONB NOT NULL,
    response_status_code INT,
    response_body TEXT,
    duration_ms INT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_webhook_logs_webhook ON webhook_logs(webhook_id);
CREATE INDEX idx_webhook_logs_created ON webhook_logs(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS webhook_logs;
DROP TABLE IF EXISTS webhooks;

-- +goose StatementEnd
