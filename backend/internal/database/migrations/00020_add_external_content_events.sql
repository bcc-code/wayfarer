-- +goose Up
-- +goose StatementBegin

-- Migration: Add external_content_events table
-- This table stores content completion events received from external systems

CREATE TABLE external_content_events (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CE[0-9A-Z]{26}$'),
    person_id UUID NOT NULL,
    content_id TEXT NOT NULL,
    reading_plan_id TEXT,
    source TEXT NOT NULL,
    received_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_external_content_events_person ON external_content_events (person_id);
CREATE INDEX idx_external_content_events_source ON external_content_events (source);
CREATE INDEX idx_external_content_events_received_at ON external_content_events (received_at);

COMMENT ON TABLE external_content_events IS 'Stores content completion events from external systems';
COMMENT ON COLUMN external_content_events.person_id IS 'Brunstad TV person_id (UUID)';
COMMENT ON COLUMN external_content_events.content_id IS 'External content identifier';
COMMENT ON COLUMN external_content_events.reading_plan_id IS 'External reading plan identifier (nullable)';
COMMENT ON COLUMN external_content_events.source IS 'API key identifier that submitted this event';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS external_content_events;

-- +goose StatementEnd
