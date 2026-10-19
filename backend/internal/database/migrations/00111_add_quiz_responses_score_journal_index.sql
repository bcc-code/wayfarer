-- +goose NO TRANSACTION
-- quiz_responses.score_journal_id references score_journal (ON DELETE SET
-- NULL). Without an index every deleted journal row (achievement revokes,
-- manual score deletes, test resets) scans all of quiz_responses; deleting a
-- betting session's ~90k journal rows took ~7 minutes on one core.
-- Partial: only settled bets have a journal ID. CONCURRENTLY so the build
-- doesn't block answer writes while the previous blue/green color serves.

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_quiz_responses_score_journal
    ON quiz_responses (score_journal_id)
    WHERE score_journal_id IS NOT NULL;

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_quiz_responses_score_journal;
