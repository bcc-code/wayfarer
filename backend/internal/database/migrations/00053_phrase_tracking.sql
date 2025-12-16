-- +goose Up
-- +goose StatementBegin

-- Hash tracking to avoid re-sending unchanged content to Phrase
CREATE TABLE translation_hashes (
    collection VARCHAR(100) PRIMARY KEY,
    hash BYTEA NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Async job tracking (replaces Redis for Phrase webhook processing)
CREATE TABLE phrase_async_jobs (
    async_request_id VARCHAR(100) PRIMARY KEY,
    job_uid VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Index for cleanup of old entries
CREATE INDEX idx_phrase_async_jobs_created_at ON phrase_async_jobs(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS phrase_async_jobs;
DROP TABLE IF EXISTS translation_hashes;

-- +goose StatementEnd
