#!/bin/bash
# One-time setup of the A/B differential environment on the test box.
# Runs FROM the dev machine; drives the box over SSH.
#
#   COMMIT_A / COMMIT_B are exported via `git archive` (no git auth needed on
#   the box), so both must be resolvable locally. The diffcmp harness itself
#   is cross-compiled locally and copied over — it is not part of either
#   commit under test.
#
# Usage: setup_box.sh [ssh-host]
set -euo pipefail

SSH_HOST="${1:-${LOADTEST_SSH:-root@49.12.121.62}}"
COMMIT_A="${COMMIT_A:-411fc319}"              # main baseline
COMMIT_B="${COMMIT_B:-ram-first-2026-08-25}"  # tagged branch tip (3057e17b)
REMOTE_DIR="/opt/wayfarer-diff"
DB_A="wayfarer_diff_a"
DB_B="wayfarer_diff_b"
JWT_SECRET="your-secret-key-for-signing-wayfarer-jwts"
WEBHOOK_SECRET="diff-webhook-secret"
QUIZ_POINTS=25

REPO_ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
HERE="$(cd "$(dirname "$0")" && pwd)"

echo "=== 1. Export both commits to the box ==="
ssh "$SSH_HOST" "rm -rf $REMOTE_DIR/a $REMOTE_DIR/b && mkdir -p $REMOTE_DIR/{a,b,bin,results}"
git -C "$REPO_ROOT" archive "$COMMIT_A" | ssh "$SSH_HOST" "tar -x -C $REMOTE_DIR/a"
git -C "$REPO_ROOT" archive "$COMMIT_B" | ssh "$SSH_HOST" "tar -x -C $REMOTE_DIR/b"

echo "=== 2. Cross-compile and copy the harness ==="
(cd "$HERE" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/diffcmp-linux .)
scp /tmp/diffcmp-linux "$SSH_HOST:$REMOTE_DIR/diffcmp"
scp "$HERE/oracle.sh" "$HERE/run.sh" "$SSH_HOST:$REMOTE_DIR/"
ssh "$SSH_HOST" "chmod +x $REMOTE_DIR/diffcmp $REMOTE_DIR/oracle.sh $REMOTE_DIR/run.sh"

echo "=== 3. Build both servers on the box ==="
ssh "$SSH_HOST" "export PATH=/usr/local/go/bin:\$PATH &&
  cd $REMOTE_DIR/a/backend && go build -o $REMOTE_DIR/bin/server-a ./cmd/server &&
  cd $REMOTE_DIR/b/backend && go build -o $REMOTE_DIR/bin/server-b ./cmd/server"

echo "=== 4. Fresh database A: create, migrate, seed ==="
ssh "$SSH_HOST" bash -s <<EOF
set -euo pipefail
export PATH=/usr/local/go/bin:\$PATH
export PGPASSWORD=bench

sudo -u postgres psql -c "DROP DATABASE IF EXISTS $DB_A" >/dev/null
sudo -u postgres psql -c "DROP DATABASE IF EXISTS $DB_B" >/dev/null
sudo -u postgres psql -c "CREATE DATABASE $DB_A OWNER bench" >/dev/null

cd $REMOTE_DIR/a/backend
DATABASE_URL="postgres://bench:bench@127.0.0.1:5432/$DB_A" go run ./cmd/migrate -direction up

PSQL="psql -h 127.0.0.1 -U bench -d $DB_A"

# The committed gendata.sql references mkid('CO', 1) as the consent id, but
# consents.id has a CHECK requiring the CN prefix — the original bench run had
# a pre-inserted CN consent (verified on the old wayfarer_bench DB). Recreate
# that row and remap the reference.
\$PSQL -q -v ON_ERROR_STOP=1 -c "INSERT INTO consents (id, key, version, title, body, published_at)
  VALUES ('CN00000000000000000000000001', 'leaderboard_consent', 1, 'Leaderboard consent', 'Benchmark consent body.', now())"
sed "s/mkid('CO', 1)/'CN00000000000000000000000001'/" $REMOTE_DIR/b/backend/cmd/loadtest/bench/gendata.sql \
  | \$PSQL -q -v ON_ERROR_STOP=1 -f -
\$PSQL -q -v ON_ERROR_STOP=1 -f $REMOTE_DIR/b/backend/cmd/loadtest/bench/gendata2.sql

# Quiz fixture (same SQL block as bench/setup_fixtures.sh, POINTS=$QUIZ_POINTS)
\$PSQL -q -v ON_ERROR_STOP=1 -v project_id="'PR00000000000000000000000001'" -v points="$QUIZ_POINTS" <<'SQL'
INSERT INTO challenges (id, project_id, challenge_type, name, description, button_text,
    notification_text, published_at, visible_at, started_at)
VALUES ('CL01LOADTESTFREETEXT00000000', :project_id, 'QUIZ', 'Load Test Free-Text Quiz',
    'Quiz for load testing.', 'Open quiz', '', NOW(), NOW(), NOW());

INSERT INTO quizzes (id, project_id, challenge_id, name, description, timeout_seconds,
    randomize_questions, reveal_correct_answers, allow_retakes, completion_points)
VALUES ('QZ01LOADTESTFREETEXT00000000', :project_id, 'CL01LOADTESTFREETEXT00000000',
    'Load Test Free-Text Quiz', 'Four questions.', 1800, false, false, true, :points);

INSERT INTO quiz_questions (id, quiz_id, question_type, question_text, question_order,
    allow_multiple_selection, min_value, max_value, step_value)
VALUES
  ('QQ01LOADTESTFREETEXT00000001','QZ01LOADTESTFREETEXT00000000','FREE_TEXT','What did you learn today?',1,false,NULL,NULL,NULL),
  ('QQ01LOADTESTFREETEXT00000002','QZ01LOADTESTFREETEXT00000000','FREE_TEXT','What will you do differently tomorrow?',2,false,NULL,NULL,NULL),
  ('QQ01LOADTESTFREETEXT00000003','QZ01LOADTESTFREETEXT00000000','PREDEFINED','Which of these did you practice this week? (select all)',3,true,NULL,NULL,NULL),
  ('QQ01LOADTESTFREETEXT00000004','QZ01LOADTESTFREETEXT00000000','NUMBER','How many minutes did you spend reading today?',4,false,1,100,1);

INSERT INTO quiz_predefined_answers (id, question_id, answer_text, is_correct, answer_order)
VALUES
  ('QA01LOADTESTFREETEXT00000001','QQ01LOADTESTFREETEXT00000003','Patience',true,1),
  ('QA01LOADTESTFREETEXT00000002','QQ01LOADTESTFREETEXT00000003','Gratitude',true,2),
  ('QA01LOADTESTFREETEXT00000003','QQ01LOADTESTFREETEXT00000003','Kindness',true,3),
  ('QA01LOADTESTFREETEXT00000004','QQ01LOADTESTFREETEXT00000003','Listening',true,4);

INSERT INTO quiz_sessions (id, quiz_id, name, state, created_by)
VALUES ('QN01LOADTESTFREETEXT00000000','QZ01LOADTESTFREETEXT00000000',
    'Load Test Free-Text Session','OPEN',(SELECT id FROM users ORDER BY id LIMIT 1));

INSERT INTO quiz_session_access (id, session_id, user_id, granted_by, source_type)
SELECT 'QX01LOADTESTFTAX' || upper(lpad(to_hex(row_number() OVER (ORDER BY u.id))::text, 12, '0')),
    'QN01LOADTESTFREETEXT00000000', u.id,
    (SELECT id FROM users ORDER BY id LIMIT 1), 'ALL'
FROM users u;
SQL

# The @requireRole directive resolves roles from user_roles, not the JWT:
# grant user 1 a global ADMIN role for the harness's admin mutations.
\$PSQL -q -v ON_ERROR_STOP=1 -c "INSERT INTO user_roles (id, user_id, role, assigned_by, assigned_at)
  VALUES ('UR00000000000000000000000001', 'US00000000000000000000000001', 'ADMIN', 'US00000000000000000000000001', now())"

# The settings migration seeds a dev-project current_project_id; point it at
# the bench project (matches the original wayfarer_bench setup) and quiet
# otel/logging the same way.
\$PSQL -q -v ON_ERROR_STOP=1 \
  -c "UPDATE settings SET value_text='PR00000000000000000000000001' WHERE key='current_project_id'" \
  -c "UPDATE settings SET value_bool=false WHERE key='otel_enabled'" \
  -c "UPDATE settings SET value_text='warn' WHERE key='log_level'"

# NOTE: bench/optimize_award_path.sql is deliberately NOT applied — the
# comparison runs both builds against the unmodified schema.
EOF

echo "=== 5. Clone A -> B (byte-identical, before any server connects) ==="
ssh "$SSH_HOST" "sudo -u postgres psql -c 'CREATE DATABASE $DB_B OWNER bench TEMPLATE $DB_A'"

echo "=== 6. systemd units + env files ==="
ssh "$SSH_HOST" bash -s <<EOF
set -euo pipefail
for side in a b; do
  if [ "\$side" = b ]; then port=8081; db=$DB_B; else port=8080; db=$DB_A; fi
  # Base env from the existing loadtest deployment, then override.
  grep -v -E '^(SERVER_PORT|DATABASE_URL|ENVIRONMENT|JWT_SECRET|JWT_ISSUER|PLUGIN_LADDER_TO_HEAVEN_SECRET_KEY|HTTP_STATS_FILE)=' /opt/wayfarer/wayfarer.env > $REMOTE_DIR/env-\$side 2>/dev/null || true
  cat >> $REMOTE_DIR/env-\$side <<ENV
SERVER_PORT=\$port
DATABASE_URL=postgres://bench:bench@127.0.0.1:5432/\$db
ENVIRONMENT=production
JWT_SECRET=$JWT_SECRET
JWT_ISSUER=wayfarer
PLUGIN_LADDER_TO_HEAVEN_SECRET_KEY=$WEBHOOK_SECRET
ENV
  cat > /etc/systemd/system/wayfarer-\$side.service <<UNIT
[Unit]
Description=Wayfarer A/B differential side \$side
After=network.target postgresql.service

[Service]
ExecStart=$REMOTE_DIR/bin/server-\$side
EnvironmentFile=$REMOTE_DIR/env-\$side
WorkingDirectory=$REMOTE_DIR/\$side/backend
Restart=on-failure

[Install]
WantedBy=multi-user.target
UNIT
done
systemctl daemon-reload
systemctl restart wayfarer-a wayfarer-b
sleep 8
curl -sf http://127.0.0.1:8080/health && echo " side A healthy"
curl -sf http://127.0.0.1:8081/health && echo " side B healthy"
EOF

echo "=== setup complete ==="
echo "run the comparison with: ssh $SSH_HOST 'cd $REMOTE_DIR && ./run.sh'"
