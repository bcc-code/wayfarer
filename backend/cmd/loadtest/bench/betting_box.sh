#!/bin/bash
# Data for the betting load test (k6/betting-quiz-rampspike.js) on a box
# provisioned with ansible/loadtest.yml (database `interact`, peer access as
# the postgres OS user). Runs ON the box as root.
#
#   betting_box.sh seed            prod-scale synthetic data (gendata*.sql, ~13k
#                                  users with points), admin role for user 1,
#                                  settings. On a migrated DB; safe to rerun.
#   betting_box.sh reset [points]  (re)create the betting quiz with an OPEN
#                                  session and access for every user, removing
#                                  the previous run's submissions and points.
#                                  Stop the server first (see loadtest-remote.sh).
#                                  points = quiz completion points (default 10).
#
# Env: PSQL (default "sudo -u postgres psql -d interact"),
#      BENCH_DIR (dir holding gendata.sql/gendata2.sql, default: this dir).
set -euo pipefail

PSQL=${PSQL:-"sudo -u postgres psql -d interact"}
BENCH_DIR=${BENCH_DIR:-$(cd "$(dirname "$0")" && pwd)}
PROJECT_ID=PR00000000000000000000000001
QUIZ_ID=QZ01LOADTESTBETTING000000000
CHALLENGE_ID=CL01LOADTESTBETTING000000000
SESSION_ID=QN01LOADTESTBETTING000000000
ADMIN_USER=US00000000000000000000000001

psql_q() { $PSQL -q -v ON_ERROR_STOP=1 "$@"; }

cmd_seed() {
    # Resumable: gendata.sql only runs on an empty DB; everything after it is
    # idempotent, so a seed that stopped halfway can simply be rerun
    if [ "$($PSQL -Atc 'SELECT count(*) FROM users')" = 0 ]; then
        echo "=== consent referenced by gendata.sql ==="
        # gendata.sql references mkid('CO', 1), but consents.id requires the CN prefix
        psql_q -c "INSERT INTO consents (id, key, version, title, body, published_at)
            VALUES ('CN00000000000000000000000001', 'leaderboard_consent', 1, 'Leaderboard consent', 'Benchmark consent body.', now())"
        echo "=== gendata.sql (takes a few minutes) ==="
        sed "s/mkid('CO', 1)/'CN00000000000000000000000001'/" "$BENCH_DIR/gendata.sql" | psql_q -f -
    else
        echo "=== users already loaded, skipping gendata.sql ==="
    fi
    echo "=== gendata2.sql ==="
    psql_q -f "$BENCH_DIR/gendata2.sql"
    echo "=== prod-shaped identities (Auth0 login path) ==="
    # Like prod users: numeric BCC person id, a person UUID and a church with
    # an external id. gendata.sql leaves these empty or as "M<n>", and
    # tokengen -auth0-count needs them to mint login.bcc.no-shaped tokens.
    psql_q \
        -c "UPDATE users SET members_id = substr(members_id, 2) WHERE members_id ~ '^M[0-9]+\$'" \
        -c "UPDATE users SET person_uuid = gen_random_uuid() WHERE person_uuid IS NULL" \
        -c "UPDATE churches c SET external_id = n.rn FROM (
                SELECT id, row_number() OVER (ORDER BY id) AS rn FROM churches) n
            WHERE c.id = n.id AND c.external_id IS NULL"
    echo "=== admin role for $ADMIN_USER, settings ==="
    psql_q -c "INSERT INTO user_roles (id, user_id, role, assigned_by, assigned_at)
        VALUES ('UR00000000000000000000000001', '$ADMIN_USER', 'ADMIN', '$ADMIN_USER', now())
        ON CONFLICT DO NOTHING"
    psql_q \
        -c "UPDATE settings SET value_text='$PROJECT_ID' WHERE key='current_project_id'" \
        -c "UPDATE settings SET value_bool=false WHERE key='otel_enabled'" \
        -c "UPDATE settings SET value_text='warn' WHERE key='log_level'"
    $PSQL -Atc "SELECT 'users=' || count(*) FROM users"
    $PSQL -Atc "SELECT 'score_journal=' || count(*) FROM score_journal"
}

cmd_reset() {
    local points=${1:-10}
    echo "=== removing the previous betting run ==="
    # Apply the leaderboard updates still queued from the last run first, so
    # the deletes below subtract points the leaderboards already have. Run
    # with the server stopped: its drain worker would race (and deadlock)
    # with the synchronous delete trigger.
    while [ "$($PSQL -Atc 'SELECT drain_leaderboard_apply_queue(5000)')" != 0 ]; do :; done
    # Points from the previous run first (DELETE keeps the leaderboards in
    # sync through the score_journal trigger), then the quiz cascade
    # A full run leaves ~90k BET journal rows. The per-row leaderboard
    # trigger made deleting them take minutes on one core, so it is off for
    # the delete and the leaderboards are rebuilt once instead (as
    # gendata2.sql does for its bulk load). The server is stopped meanwhile.
    psql_q -v quiz_id="'$QUIZ_ID'" <<'SQL'
BEGIN;
ALTER TABLE score_journal DISABLE TRIGGER trigger_score_journal_leaderboard;
DELETE FROM score_journal
WHERE (source_type = 'BET' AND source_id IN (
        SELECT r.id FROM quiz_responses r
        JOIN quiz_submissions s ON s.id = r.submission_id
        WHERE s.quiz_id = :quiz_id))
   OR (source_type = 'QUIZ' AND source_id = :quiz_id);
ALTER TABLE score_journal ENABLE TRIGGER trigger_score_journal_leaderboard;
SELECT regenerate_leaderboards();
COMMIT;
DELETE FROM quiz_responses WHERE submission_id IN (
    SELECT id FROM quiz_submissions WHERE quiz_id = :quiz_id);
DELETE FROM quiz_submissions WHERE quiz_id = :quiz_id;
DELETE FROM quiz_sessions WHERE quiz_id = :quiz_id;
DELETE FROM quizzes WHERE id = :quiz_id;
SQL
    psql_q -c "DELETE FROM challenges WHERE id = '$CHALLENGE_ID'"

    echo "=== creating the betting quiz (completion_points=$points) ==="
    psql_q -v project_id="'$PROJECT_ID'" -v points="$points" -v admin="'$ADMIN_USER'" <<'SQL'
INSERT INTO challenges (id, project_id, challenge_type, name, description, button_text,
    notification_text, published_at, visible_at, started_at)
VALUES ('CL01LOADTESTBETTING000000000', :project_id, 'QUIZ', 'Load Test Betting Quiz',
    'Quiz for the betting load test.', 'Open quiz', '', NOW(), NOW(), NOW());

INSERT INTO quizzes (id, project_id, challenge_id, name, description, timeout_seconds,
    randomize_questions, reveal_correct_answers, allow_retakes, completion_points)
VALUES ('QZ01LOADTESTBETTING000000000', :project_id, 'CL01LOADTESTBETTING000000000',
    'Load Test Betting Quiz', 'Five betting questions.', 1800, false, false, false, :points);

-- One question per limit style the app supports. The correct answer is
-- always answer_order 1, which k6 picks with probability CORRECT_RATE.
INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order,
    allow_multiple_selection, betting_enabled, betting_min_percentage, betting_max_percentage,
    betting_min_absolute, betting_max_absolute, betting_multiplier_correct, betting_multiplier_wrong)
VALUES
  ('QQ01LOADTESTBETTING000000001','QZ01LOADTESTBETTING000000000','PREDEFINED','Who led Israel out of Egypt?',1,false,true,NULL,25,NULL,NULL,NULL,NULL),
  ('QQ01LOADTESTBETTING000000002','QZ01LOADTESTBETTING000000000','PREDEFINED','How many days did it rain on the ark?',2,false,true,NULL,50,NULL,NULL,NULL,NULL),
  ('QQ01LOADTESTBETTING000000003','QZ01LOADTESTBETTING000000000','PREDEFINED','Who was swallowed by a great fish?',3,false,true,NULL,NULL,NULL,100,NULL,NULL),
  ('QQ01LOADTESTBETTING000000004','QZ01LOADTESTBETTING000000000','PREDEFINED','Which king built the first temple?',4,false,true,NULL,50,5,NULL,3.00,0.00),
  ('QQ01LOADTESTBETTING000000005','QZ01LOADTESTBETTING000000000','PREDEFINED','Where was Jesus born?',5,false,true,NULL,NULL,NULL,NULL,1.50,0.50);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order)
SELECT 'QA01LOADTESTBETTING0000000' || q || a,
       'QQ01LOADTESTBETTING00000000' || q,
       (ARRAY[['Moses','Aaron','Joshua'],['40','7','12'],['Jonah','Peter','Elijah'],
              ['Solomon','David','Saul'],['Bethlehem','Nazareth','Jerusalem']])[q][a],
       a = 1, a
FROM generate_series(1, 5) q, generate_series(1, 3) a;

INSERT INTO quiz_sessions (id, quiz_id, name, state, created_by)
VALUES ('QN01LOADTESTBETTING000000000','QZ01LOADTESTBETTING000000000',
    'Load Test Betting Session','OPEN', :admin);

INSERT INTO quiz_session_access (id, session_id, user_id, granted_by, source_type)
SELECT 'QX01LOADTESTBETX' || upper(lpad(to_hex(row_number() OVER (ORDER BY u.id))::text, 12, '0')),
    'QN01LOADTESTBETTING000000000', u.id, :admin, 'ALL'
FROM users u;
SQL
    $PSQL -Atc "SELECT 'session access grants=' || count(*) FROM quiz_session_access WHERE session_id='$SESSION_ID'"
    $PSQL -Atc "SELECT 'users with points to bet=' || count(*) FROM (SELECT user_id FROM score_journal WHERE project_id='$PROJECT_ID' GROUP BY user_id HAVING SUM(points) > 0) t"
}

case "${1:-}" in
seed) shift; cmd_seed "$@" ;;
reset) shift; cmd_reset "$@" ;;
*)
    sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
