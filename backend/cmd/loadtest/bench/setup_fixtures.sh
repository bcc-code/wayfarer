#!/bin/bash
set -e
export PATH=/usr/local/go/bin:$PATH
export PGPASSWORD=bench
PSQL="psql -h localhost -U bench -d wayfarer_bench"

PROJECT_ID="PR00000000000000000000000001"

echo "=== 1. Synthetic Firebase service account (local RSA signing only) ==="
if [ ! -f /opt/wayfarer/fake-sa.json ]; then
  openssl genrsa -out /tmp/sa.key 2048 2>/dev/null
  KEY=$(python3 -c "print(open('/tmp/sa.key').read().replace(chr(10),'\\\\n'))")
  cat > /opt/wayfarer/fake-sa.json <<EOF
{
  "type": "service_account",
  "project_id": "wayfarer-loadtest",
  "private_key_id": "loadtestkeyid0000000000000000000000000000",
  "private_key": "${KEY}",
  "client_email": "loadtest@wayfarer-loadtest.iam.gserviceaccount.com",
  "client_id": "000000000000000000000",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/loadtest%40wayfarer-loadtest.iam.gserviceaccount.com"
}
EOF
  python3 -c "import json;json.load(open('/opt/wayfarer/fake-sa.json'));print('  service account JSON valid')"
fi

echo ""
echo "=== 2. Verify project exists ==="
$PSQL -Atc "SELECT id, name FROM projects WHERE id='${PROJECT_ID}';"

echo ""
echo "=== 3. Insert quiz fixture (parameterised points) ==="
# $1 = completion_points
POINTS=${1:-0}
$PSQL -q -v project_id="'${PROJECT_ID}'" -v points="${POINTS}" <<'SQL'
DELETE FROM quiz_submissions WHERE quiz_id = 'QZ01LOADTESTFREETEXT00000000';
DELETE FROM quizzes WHERE id = 'QZ01LOADTESTFREETEXT00000000';
DELETE FROM challenges WHERE id = 'CL01LOADTESTFREETEXT00000000';

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

$PSQL -Atc "SELECT 'quiz completion_points=' || completion_points FROM quizzes WHERE id='QZ01LOADTESTFREETEXT00000000';"
$PSQL -Atc "SELECT 'session access grants=' || count(*) FROM quiz_session_access WHERE session_id='QN01LOADTESTFREETEXT00000000';"
$PSQL -Atc "SELECT 'users in project=' || count(*) FROM user_projects WHERE project_id='${PROJECT_ID}';"
