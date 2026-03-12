-- +goose Up
-- +goose StatementBegin

-- Bulk jobs table for tracking async bulk operations
CREATE TABLE bulk_jobs (
    id CHAR(28) PRIMARY KEY,
    operation_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_by CHAR(28) NOT NULL REFERENCES users(id),
    project_id CHAR(28) REFERENCES projects(id),
    input_params JSONB NOT NULL,
    total_count INT NOT NULL DEFAULT 0,
    processed_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    error_details JSONB,
    created_at TIMESTAMPTZ DEFAULT now(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    message_id VARCHAR(255)
);

-- Index for querying jobs by status
CREATE INDEX idx_bulk_jobs_status ON bulk_jobs(status);

-- Index for querying jobs by creator
CREATE INDEX idx_bulk_jobs_created_by ON bulk_jobs(created_by);

-- Index for querying jobs by project
CREATE INDEX idx_bulk_jobs_project_id ON bulk_jobs(project_id) WHERE project_id IS NOT NULL;

-- Index for querying recent pending/processing jobs
CREATE INDEX idx_bulk_jobs_status_created_at ON bulk_jobs(status, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS bulk_jobs;

-- +goose StatementEnd
