-- +goose Up
-- Async leaderboard apply: move the score_journal INSERT fan-out (38
-- statements per point award, 45x write amplification, measured ceiling
-- ~2,200 inserts/s) out of the request transaction. The hot-path trigger now
-- writes ONE row into an outbox queue; a background worker (see
-- internal/services/leaderboardapply) drains the queue in batches and calls
-- the same update_{person,church,team,superteam}_leaderboard helpers the
-- trigger called, so scoring semantics are unchanged and everything stays
-- durable in Postgres. Rare admin UPDATE/DELETE on score_journal keep the
-- old synchronous trigger.

CREATE TABLE leaderboard_apply_queue (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id CHAR(28) NOT NULL,
    project_id CHAR(28) NOT NULL,
    event_id CHAR(28),
    points_delta BIGINT NOT NULL,
    score_at TIMESTAMPTZ
);

-- +goose StatementBegin
-- The INSERT branch of trigger_update_leaderboard_from_score_journal(),
-- callable outside a trigger. Keep in sync with that function.
CREATE OR REPLACE FUNCTION apply_score_journal_delta(
    target_user_id CHAR(28),
    target_project_id CHAR(28),
    target_event_id CHAR(28),
    points_delta BIGINT,
    score_timestamp TIMESTAMPTZ
) RETURNS void AS $$
DECLARE
    user_church_id CHAR(28);
    user_exists BOOLEAN;
    team_record RECORD;
    processed_super_teams CHAR(28)[];
BEGIN
    SELECT TRUE, u.church_id
    INTO user_exists, user_church_id
    FROM users u
    WHERE u.id = target_user_id;

    IF NOT COALESCE(user_exists, FALSE) THEN
        RETURN;
    END IF;

    PERFORM update_person_leaderboard(
        target_user_id, target_project_id, target_event_id,
        points_delta, score_timestamp
    );

    IF user_church_id IS NOT NULL THEN
        PERFORM update_church_leaderboard(
            user_church_id, target_project_id, target_event_id,
            points_delta, score_timestamp
        );
    END IF;

    processed_super_teams := ARRAY[]::CHAR(28)[];

    FOR team_record IN
        SELECT t.id AS team_id, t.super_team_id
        FROM team_members tm
        INNER JOIN teams t ON t.id = tm.team_id AND t.project_id = target_project_id
        WHERE tm.user_id = target_user_id
    LOOP
        PERFORM update_team_leaderboard(
            team_record.team_id, target_project_id, target_event_id,
            points_delta, score_timestamp
        );

        IF team_record.super_team_id IS NOT NULL
           AND NOT (team_record.super_team_id = ANY(processed_super_teams)) THEN
            PERFORM update_superteam_leaderboard(
                team_record.super_team_id, target_project_id, target_event_id,
                points_delta, score_timestamp
            );
            processed_super_teams := array_append(processed_super_teams, team_record.super_team_id);
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- Hot-path trigger: one queue insert instead of the 38-statement fan-out.
CREATE OR REPLACE FUNCTION trigger_enqueue_leaderboard_apply()
RETURNS trigger AS $$
BEGIN
    INSERT INTO leaderboard_apply_queue (user_id, project_id, event_id, points_delta, score_at)
    VALUES (NEW.user_id, NEW.project_id, NEW.event_id, NEW.points, COALESCE(NEW.created_at, NOW()));
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- Drains up to batch_size queued deltas. FIFO by queue id; SKIP LOCKED makes
-- concurrent workers (e.g. two server instances) safe — deltas commute, so
-- interleaving is harmless. Returns the number of rows applied.
CREATE OR REPLACE FUNCTION drain_leaderboard_apply_queue(batch_size INT)
RETURNS INT AS $$
DECLARE
    applied INT;
BEGIN
    WITH batch AS (
        DELETE FROM leaderboard_apply_queue
        WHERE id IN (
            SELECT id FROM leaderboard_apply_queue
            ORDER BY id
            LIMIT batch_size
            FOR UPDATE SKIP LOCKED
        )
        RETURNING user_id, project_id, event_id, points_delta, score_at
    )
    SELECT count(*) INTO applied
    FROM (
        SELECT apply_score_journal_delta(user_id, project_id, event_id, points_delta, score_at)
        FROM batch
    ) applied_rows;
    RETURN applied;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Swap the score_journal trigger: INSERTs enqueue; UPDATE/DELETE (rare admin
-- paths) keep the synchronous fan-out with unchanged semantics.
DROP TRIGGER trigger_score_journal_leaderboard ON score_journal;

CREATE TRIGGER trigger_score_journal_leaderboard
AFTER UPDATE OR DELETE ON score_journal
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_from_score_journal();

CREATE TRIGGER trigger_score_journal_enqueue_apply
AFTER INSERT ON score_journal
FOR EACH ROW
EXECUTE FUNCTION trigger_enqueue_leaderboard_apply();

-- +goose Down
DROP TRIGGER trigger_score_journal_enqueue_apply ON score_journal;
DROP TRIGGER trigger_score_journal_leaderboard ON score_journal;

CREATE TRIGGER trigger_score_journal_leaderboard
AFTER INSERT OR UPDATE OR DELETE ON score_journal
FOR EACH ROW
EXECUTE FUNCTION trigger_update_leaderboard_from_score_journal();

DROP FUNCTION drain_leaderboard_apply_queue(INT);
DROP FUNCTION trigger_enqueue_leaderboard_apply();
DROP FUNCTION apply_score_journal_delta(CHAR(28), CHAR(28), CHAR(28), BIGINT, TIMESTAMPTZ);
DROP TABLE leaderboard_apply_queue;
