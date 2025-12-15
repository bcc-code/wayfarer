-- +goose Up
-- +goose StatementBegin

-- Migration: Add pending_consent_events table
-- Stores consent events received before the user exists in Wayfarer.
-- These are processed when the user registers.

CREATE TABLE pending_consent_events (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^PC[0-9A-Z]{26}$'),
    members_id VARCHAR(255) NOT NULL,
    consent_key VARCHAR(255) NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('ACCEPTED', 'REJECTED')),
    occurred_at TIMESTAMPTZ NOT NULL,
    source VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_consent_events_members_id ON pending_consent_events(members_id);

COMMENT ON TABLE pending_consent_events IS 'Stores consent events for users not yet registered in Wayfarer';
COMMENT ON COLUMN pending_consent_events.members_id IS 'BCC Members person_id (UUID) of the user';
COMMENT ON COLUMN pending_consent_events.consent_key IS 'Key identifying the consent type';
COMMENT ON COLUMN pending_consent_events.action IS 'ACCEPTED or REJECTED';
COMMENT ON COLUMN pending_consent_events.occurred_at IS 'When the consent action occurred (from external system)';
COMMENT ON COLUMN pending_consent_events.source IS 'API key source that submitted this event';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pending_consent_events;

-- +goose StatementEnd
