-- +goose Up
-- +goose StatementBegin

-- Settings table stores runtime configuration as key-value pairs
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value_text TEXT,
    value_int BIGINT,
    value_bool BOOLEAN,
    value_float DOUBLE PRECISION,
    value_json JSONB,
    value_type TEXT NOT NULL CHECK (value_type IN ('text', 'int', 'bool', 'float', 'json')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE settings IS 'Runtime configuration key-value store';
COMMENT ON COLUMN settings.key IS 'Unique setting identifier (e.g., current_project_id, log_level)';
COMMENT ON COLUMN settings.value_type IS 'Indicates which value column to use: text, int, bool, float, or json';
COMMENT ON COLUMN settings.description IS 'Human-readable description of the setting';

-- Auto-update updated_at timestamp
CREATE TRIGGER settings_updated_at
    BEFORE UPDATE ON settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default settings
INSERT INTO settings (key, value_text, value_type, description) VALUES
    ('current_project_id', 'PR01K9VZ8684DP9R4W3ZEV526E5X', 'text', 'Default project ID for unauthenticated queries');

INSERT INTO settings (key, value_text, value_type, description) VALUES
    ('log_level', 'info', 'text', 'Logging verbosity: debug, info, warn, error');

INSERT INTO settings (key, value_bool, value_type, description) VALUES
    ('db_log_queries', false, 'bool', 'Toggle query logging for debugging');

INSERT INTO settings (key, value_bool, value_type, description) VALUES
    ('otel_enabled', true, 'bool', 'Toggle OpenTelemetry tracing on/off');

INSERT INTO settings (key, value_float, value_type, description) VALUES
    ('otel_sampling_ratio', 0.1, 'float', 'OpenTelemetry sampling rate (0.0-1.0)');

INSERT INTO settings (key, value_bool, value_type, description) VALUES
    ('ssf_debug_mode', false, 'bool', 'SSF API debug logging');

-- Index for faster lookups by type (likely unnecessary given small table size, but follows best practices)
CREATE INDEX idx_settings_value_type ON settings(value_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS settings CASCADE;

-- +goose StatementEnd
