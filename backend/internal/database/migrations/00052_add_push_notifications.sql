-- +goose Up
-- +goose StatementBegin

-- Push subscriptions (user devices)
CREATE TABLE push_subscriptions (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PS[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);

-- Trigger to update updated_at
CREATE TRIGGER set_push_subscriptions_updated_at
    BEFORE UPDATE ON push_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Per-type notification preferences
CREATE TABLE push_notification_preferences (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, notification_type)
);

-- Trigger to update updated_at
CREATE TRIGGER set_push_notification_preferences_updated_at
    BEFORE UPDATE ON push_notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Audit log for sent notifications
CREATE TABLE push_notification_log (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PN[0-9A-Z]{26}$'),
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    url TEXT,
    data JSONB,
    target_criteria JSONB NOT NULL,
    sent_by CHAR(28) REFERENCES users(id) ON DELETE SET NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    total_recipients INT NOT NULL DEFAULT 0,
    successful_deliveries INT NOT NULL DEFAULT 0,
    failed_deliveries INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_push_notification_log_sent_at ON push_notification_log(sent_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS push_notification_log;
DROP TABLE IF EXISTS push_notification_preferences;
DROP TABLE IF EXISTS push_subscriptions;

-- +goose StatementEnd
