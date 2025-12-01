-- +goose Up
-- +goose StatementBegin

-- Consent definitions (English text in main table)
-- key: identifies consent type (e.g., "privacy_policy", "terms_of_service")
-- version: increments per key for tracking consent updates
-- published_at: null = draft, non-null = active consent
CREATE TABLE consents (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^CN[0-9A-Z]{26}$'),
    key VARCHAR(100) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (key, version)
);

CREATE INDEX idx_consents_key ON consents(key);
CREATE INDEX idx_consents_published ON consents(published_at) WHERE published_at IS NOT NULL;

-- Consent translations (shadow table pattern)
CREATE TABLE consent_translations (
    consent_id CHAR(28) NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    title VARCHAR(255),
    body TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (consent_id, language_code)
);

-- User consent acceptances
-- ON DELETE RESTRICT for consent_id: prevent deleting consents that users have accepted
CREATE TABLE user_consents (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^UC[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_id CHAR(28) NOT NULL REFERENCES consents(id) ON DELETE RESTRICT,
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, consent_id)
);

CREATE INDEX idx_user_consents_user ON user_consents(user_id);
CREATE INDEX idx_user_consents_consent ON user_consents(consent_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_consents;
DROP TABLE IF EXISTS consent_translations;
DROP TABLE IF EXISTS consents;

-- +goose StatementEnd
