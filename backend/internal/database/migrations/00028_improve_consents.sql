-- +goose Up
-- +goose StatementBegin

-- Migration: Improve consent system
-- - Add URL field to consents
-- - Add remote consent tracking fields
-- - Transform user_consents to user_consent_history with rejection tracking

-- 1. Add new columns to consents table
ALTER TABLE consents
ADD COLUMN url VARCHAR(500),
ADD COLUMN managed_by VARCHAR(100),
ADD COLUMN is_remote BOOLEAN DEFAULT false NOT NULL;

CREATE INDEX idx_consents_is_remote ON consents(is_remote) WHERE is_remote = true;

-- 2. Rename user_consents to user_consent_history
ALTER TABLE user_consents RENAME TO user_consent_history;

-- 3. Update ID prefix constraint to accept both UC (legacy) and UH (new) prefixes
ALTER TABLE user_consent_history
DROP CONSTRAINT user_consents_id_check;

ALTER TABLE user_consent_history
ADD CONSTRAINT user_consent_history_id_check
CHECK (id ~ '^U[CH][0-9A-Z]{26}$');

-- 4. Add new columns to user_consent_history
ALTER TABLE user_consent_history
ADD COLUMN action VARCHAR(20) NOT NULL DEFAULT 'ACCEPTED',
ADD COLUMN consent_key VARCHAR(100) NOT NULL DEFAULT '',
ADD COLUMN occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
ADD COLUMN source VARCHAR(100),
ADD COLUMN external_consent_id VARCHAR(255),
ADD COLUMN external_timestamp TIMESTAMPTZ;

-- 5. Backfill consent_key from consents table
UPDATE user_consent_history uch
SET consent_key = c.key
FROM consents c
WHERE uch.consent_id = c.id;

-- 6. Backfill occurred_at from accepted_at
UPDATE user_consent_history
SET occurred_at = accepted_at;

-- 7. Remove defaults after backfill
ALTER TABLE user_consent_history
ALTER COLUMN consent_key DROP DEFAULT,
ALTER COLUMN action DROP DEFAULT,
ALTER COLUMN occurred_at DROP DEFAULT;

-- 8. Add constraint for action enum
ALTER TABLE user_consent_history
ADD CONSTRAINT check_action CHECK (action IN ('ACCEPTED', 'REJECTED'));

-- 9. Drop old columns
ALTER TABLE user_consent_history
DROP COLUMN accepted_at,
DROP COLUMN created_at;

-- 10. Drop old unique constraint (allow multiple actions over time)
ALTER TABLE user_consent_history
DROP CONSTRAINT user_consents_user_id_consent_id_key;

-- 11. Update indexes
DROP INDEX IF EXISTS idx_user_consents_user;
DROP INDEX IF EXISTS idx_user_consents_consent;

CREATE INDEX idx_user_consent_history_user ON user_consent_history(user_id);
CREATE INDEX idx_user_consent_history_consent ON user_consent_history(consent_id);
CREATE INDEX idx_user_consent_history_key ON user_consent_history(consent_key);
CREATE INDEX idx_user_consent_history_user_key ON user_consent_history(user_id, consent_key);
CREATE INDEX idx_user_consent_history_occurred ON user_consent_history(occurred_at);

-- 12. Add unique index for idempotency of remote consents
CREATE UNIQUE INDEX idx_user_consent_history_remote_idempotent
ON user_consent_history(user_id, consent_key, external_consent_id)
WHERE external_consent_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse migration: Revert consent system changes

-- 1. Drop unique remote consent index
DROP INDEX IF EXISTS idx_user_consent_history_remote_idempotent;

-- 2. Drop new indexes
DROP INDEX IF EXISTS idx_user_consent_history_occurred;
DROP INDEX IF EXISTS idx_user_consent_history_user_key;
DROP INDEX IF EXISTS idx_user_consent_history_key;
DROP INDEX IF EXISTS idx_user_consent_history_consent;
DROP INDEX IF EXISTS idx_user_consent_history_user;

-- 3. Recreate old indexes
CREATE INDEX idx_user_consents_user ON user_consent_history(user_id);
CREATE INDEX idx_user_consents_consent ON user_consent_history(consent_id);

-- 4. Add back old columns with data
ALTER TABLE user_consent_history
ADD COLUMN accepted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- 5. Backfill accepted_at from occurred_at
UPDATE user_consent_history
SET accepted_at = occurred_at,
    created_at = occurred_at;

-- 6. Drop defaults after backfill
ALTER TABLE user_consent_history
ALTER COLUMN accepted_at DROP DEFAULT,
ALTER COLUMN created_at DROP DEFAULT;

-- 7. Restore unique constraint
ALTER TABLE user_consent_history
ADD CONSTRAINT user_consents_user_id_consent_id_key UNIQUE (user_id, consent_id);

-- 8. Drop action constraint
ALTER TABLE user_consent_history
DROP CONSTRAINT check_action;

-- 9. Drop new columns
ALTER TABLE user_consent_history
DROP COLUMN external_timestamp,
DROP COLUMN external_consent_id,
DROP COLUMN source,
DROP COLUMN occurred_at,
DROP COLUMN consent_key,
DROP COLUMN action;

-- 10. Revert ID prefix constraint (back to UC only)
ALTER TABLE user_consent_history
DROP CONSTRAINT user_consent_history_id_check;

ALTER TABLE user_consent_history
ADD CONSTRAINT user_consents_id_check
CHECK (id ~ '^UC[0-9A-Z]{26}$');

-- 11. Rename table back
ALTER TABLE user_consent_history RENAME TO user_consents;

-- 12. Drop consent table additions
DROP INDEX IF EXISTS idx_consents_is_remote;

ALTER TABLE consents
DROP COLUMN is_remote,
DROP COLUMN managed_by,
DROP COLUMN url;

-- +goose StatementEnd
