-- Backfill script to assign random birthdates to users with NULL birthdate
-- This script should be run BEFORE applying migration 00006_make_birthdate_required.sql
--
-- Usage:
--   psql $DATABASE_URL -f internal/database/migrations/scripts/backfill_birthdates.sql

BEGIN;

-- Update users with NULL birthdate
-- Assigns random birthdates corresponding to ages between 13-80 years old
UPDATE users
SET birthdate = CURRENT_DATE
    - INTERVAL '1 year' * (13 + FLOOR(RANDOM() * 68)::INT)  -- Random years: 13-80
    - INTERVAL '1 month' * FLOOR(RANDOM() * 12)::INT         -- Random months: 0-11
    - INTERVAL '1 day' * FLOOR(RANDOM() * 28)::INT           -- Random days: 0-27
WHERE birthdate IS NULL;

-- Display the number of rows updated
SELECT FORMAT('Updated %s users with random birthdates', COUNT(*))
FROM users
WHERE birthdate IS NOT NULL;

COMMIT;

-- Verify no NULL birthdates remain
SELECT
    CASE
        WHEN COUNT(*) = 0 THEN 'SUCCESS: No users with NULL birthdate'
        ELSE FORMAT('WARNING: %s users still have NULL birthdate', COUNT(*))
    END AS result
FROM users
WHERE birthdate IS NULL;
