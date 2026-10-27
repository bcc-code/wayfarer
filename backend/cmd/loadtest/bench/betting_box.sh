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
#   betting_box.sh reset-multi [buckets] [webhook_url] [points]
#                                  the betting quiz with NO session, plus
#                                  load-test churches sized by buckets (default
#                                  "885:1-9,90:10-49,20:50-99,5:100-500" =
#                                  1000 churches, ~10k players), each with one
#                                  church admin, for k6/betting-multisession.js.
#                                  webhook_url registers the plugin's
#                                  quiz_session_finished receiver. Stop the
#                                  server first.
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
    remove_previous_run
    create_quiz "$points"
    create_single_session
}

remove_previous_run() {
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
    # The responses go first: quiz_responses.score_journal_id references the
    # journal (ON DELETE SET NULL) without an index, so every deleted journal
    # row scans quiz_responses. With the run's responses still there (~50k),
    # deleting ~90k journal rows took ~7 min on one core.
    psql_q -v quiz_id="'$QUIZ_ID'" <<'SQL'
BEGIN;
CREATE TEMP TABLE run_responses ON COMMIT DROP AS
    SELECT r.id FROM quiz_responses r
    JOIN quiz_submissions s ON s.id = r.submission_id
    WHERE s.quiz_id = :quiz_id;
DELETE FROM quiz_responses WHERE id IN (SELECT id FROM run_responses);
DELETE FROM quiz_submissions WHERE quiz_id = :quiz_id;
ALTER TABLE score_journal DISABLE TRIGGER trigger_score_journal_leaderboard;
DELETE FROM score_journal
WHERE (source_type = 'BET' AND source_id IN (SELECT id FROM run_responses))
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
}

create_quiz() {
    local points=$1
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
SQL
}

create_single_session() {
    psql_q -v admin="'$ADMIN_USER'" <<'SQL'
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

# DEFAULT_BUCKETS: 5 big churches (100-500 players), 20 medium (50-99),
# 90 small (10-49) and 885 tiny (1-9): 1000 churches, ~10k players
DEFAULT_BUCKETS="885:1-9,90:10-49,20:50-99,5:100-500"

# church_sizes <buckets> prints "<index>\t<players>" for each load-test church.
# buckets is a comma-separated list of <churches>:<min>-<max>; each bucket's
# sizes are spread evenly from min to max (fixed, not random, so every run
# has the same mix). Sizes ascend with the index.
church_sizes() {
    echo "$1" | tr ',' '\n' | awk -F'[:-]' '
    NF != 3 || $1 < 1 || $2 < 1 || $3 < $2 { print "bad bucket: " $0 > "/dev/stderr"; exit 1 }
    { for (j = 0; j < $1; j++) print ($1 == 1 ? $3 : int($2 + ($3 - $2) * j / ($1 - 1) + 0.5)) }' |
        sort -n | awk '{ printf "%d\t%d\n", NR, $1 }'
}

# reset-multi: the betting quiz with no sessions, plus load-test churches
# sized by <buckets>, holding players (users in the project with points), and
# one CHURCH_ADMIN each. The admin VUs of k6/betting-multisession.js create,
# open, grant, lock and finish one session per church through GraphQL.
# Users moved into the load-test churches are restored to their own church
# first, so the reset can be rerun.
cmd_reset_multi() {
    local buckets=${1:-$DEFAULT_BUCKETS} webhook_url=${2:-} points=${3:-10}
    remove_previous_run
    create_quiz "$points"

    local table churches players sizes
    table=$(church_sizes "$buckets")
    churches=$(echo "$table" | wc -l | tr -d ' ')
    players=$(echo "$table" | awk -F'\t' '{ s += $2 } END { print s }')
    echo "=== $churches load-test churches, $players players (buckets $buckets) ==="
    sizes=$(echo "$table" | awk -F'\t' '{printf "%s(%d,%d)", (NR>1?",":""), $1, $2}')
    # Right after the bulk seed the planner has no statistics yet: the first
    # reset then took ~10 min instead of ~6s
    psql_q -c "ANALYZE users" -c "ANALYZE user_projects" -c "ANALYZE score_journal" -c "ANALYZE churches"
    psql_q -v project_id="'$PROJECT_ID'" -v admin="'$ADMIN_USER'" -v players="$players" -v churches="$churches" <<SQL
BEGIN;
-- Undo the previous multi reset
CREATE TABLE IF NOT EXISTS loadtest_church_backup (
    user_id CHAR(28) PRIMARY KEY, church_id CHAR(28) NOT NULL);
UPDATE users u SET church_id = b.church_id FROM loadtest_church_backup b WHERE u.id = b.user_id;
DELETE FROM loadtest_church_backup;
DELETE FROM churches WHERE id LIKE 'CH01LOADTESTCHURCH%';  -- cascades admin roles

CREATE TEMP TABLE sizes (idx INT, size INT) ON COMMIT DROP;
INSERT INTO sizes VALUES $sizes;

INSERT INTO churches (id, name, country, category, external_id)
SELECT 'CH01LOADTESTCHURCH' || lpad(idx::text, 10, '0'), 'Load Test Church ' || idx, 'NO',
       CASE WHEN size >= 50 THEN 'XL' WHEN size >= 15 THEN 'L' ELSE 'S' END, 900000 + idx
FROM sizes;

-- Players: users in the project with points, in id order, filling the
-- churches by cumulative size
CREATE TEMP TABLE bounds ON COMMIT DROP AS
    SELECT idx, sum(size) OVER (ORDER BY idx) - size AS lo, sum(size) OVER (ORDER BY idx) AS hi FROM sizes;
CREATE TEMP TABLE players ON COMMIT DROP AS
    SELECT u.id, u.church_id, row_number() OVER (ORDER BY u.id) - 1 AS rn
    FROM users u
    JOIN user_projects up ON up.user_id = u.id AND up.project_id = :project_id
    WHERE u.id IN (SELECT user_id FROM score_journal WHERE project_id = :project_id
                   GROUP BY user_id HAVING sum(points) > 0)
    ORDER BY u.id LIMIT :players;
-- Admins: users outside the project (so they are never players), one per church
CREATE TEMP TABLE admins ON COMMIT DROP AS
    SELECT u.id, u.church_id, row_number() OVER (ORDER BY u.id) AS idx
    FROM users u
    WHERE u.id <> :admin
      AND NOT EXISTS (SELECT 1 FROM user_projects up WHERE up.user_id = u.id AND up.project_id = :project_id)
    ORDER BY u.id LIMIT :churches;

INSERT INTO loadtest_church_backup (user_id, church_id)
SELECT id, church_id FROM players UNION ALL SELECT id, church_id FROM admins;

UPDATE users u SET church_id = 'CH01LOADTESTCHURCH' || lpad(b.idx::text, 10, '0')
FROM players p JOIN bounds b ON p.rn >= b.lo AND p.rn < b.hi
WHERE u.id = p.id;
UPDATE users u SET church_id = 'CH01LOADTESTCHURCH' || lpad(a.idx::text, 10, '0')
FROM admins a WHERE u.id = a.id;

INSERT INTO user_roles (id, user_id, role, church_id, assigned_by, assigned_at)
SELECT 'UR01LOADTESTADMIN' || lpad(a.idx::text, 11, '0'), a.id, 'CHURCH_ADMIN',
       'CH01LOADTESTCHURCH' || lpad(a.idx::text, 10, '0'), :admin, now()
FROM admins a;

DO \$\$ BEGIN
    IF (SELECT count(*) FROM players) < $players THEN
        RAISE EXCEPTION 'only % users in the project have points, need %', (SELECT count(*) FROM players), $players;
    END IF;
    IF (SELECT count(*) FROM admins) < $churches THEN
        RAISE EXCEPTION 'only % users outside the project for admins, need %', (SELECT count(*) FROM admins), $churches;
    END IF;
END \$\$;

ALTER TABLE score_journal DISABLE TRIGGER trigger_score_journal_leaderboard;
SELECT regenerate_leaderboards();
ALTER TABLE score_journal ENABLE TRIGGER trigger_score_journal_leaderboard;
COMMIT;
SQL

    # The ladder_to_heaven plugin receives quiz_session_finished as a webhook
    # from the same server, as in prod. Without it every finish would skip the
    # plugin's work.
    psql_q -c "DELETE FROM webhooks WHERE id = 'WH01LOADTESTMULTI00000000000'"
    if [ -n "$webhook_url" ]; then
        psql_q -c "INSERT INTO webhooks (id, project_id, name, url, event_type)
            VALUES ('WH01LOADTESTMULTI00000000000', '$PROJECT_ID', 'Load test plugin receiver', '$webhook_url', 'quiz_session_finished')"
    fi

    $PSQL -Atc "SELECT 'churches=' || count(*) || ' players=' || sum(n) || ' min=' || min(n) || ' median=' || percentile_disc(0.5) WITHIN GROUP (ORDER BY n) || ' max=' || max(n) || ' churches_100plus=' || count(*) FILTER (WHERE n >= 100)
        FROM (SELECT c.id, count(up.user_id) AS n FROM churches c
              JOIN users u ON u.church_id = c.id
              JOIN user_projects up ON up.user_id = u.id AND up.project_id = '$PROJECT_ID'
              WHERE c.id LIKE 'CH01LOADTESTCHURCH%' GROUP BY c.id) t"
    $PSQL -Atc "SELECT 'church admins=' || count(*) FROM user_roles WHERE id LIKE 'UR01LOADTESTADMIN%'"
}

case "${1:-}" in
seed) shift; cmd_seed "$@" ;;
reset) shift; cmd_reset "$@" ;;
reset-multi) shift; cmd_reset_multi "$@" ;;
*)
    sed -n '2,24p' "$0" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
