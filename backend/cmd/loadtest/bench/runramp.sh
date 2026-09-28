#!/bin/bash
# Usage: runramp.sh <label> [warmup_seconds]
LABEL=${1:-ramp_opt}; WARM=${2:-18}
export PGPASSWORD=bench S_TIME_FORMAT=ISO LC_ALL=C
PSQL="psql -h localhost -U bench -d wayfarer_bench -q -Atc"
OUT=/opt/wayfarer/results; mkdir -p $OUT

$PSQL "UPDATE quizzes SET completion_points=10 WHERE id='QZ01LOADTESTFREETEXT00000000';" >/dev/null
$PSQL "DELETE FROM quiz_responses WHERE submission_id IN (SELECT id FROM quiz_submissions WHERE quiz_id='QZ01LOADTESTFREETEXT00000000');" >/dev/null
$PSQL "DELETE FROM quiz_submissions WHERE quiz_id='QZ01LOADTESTFREETEXT00000000';" >/dev/null
SJ_BEFORE=$($PSQL "SELECT count(*) FROM score_journal;")

systemctl restart wayfarer
echo "waiting ${WARM}s for token warmer boot pass..."
sleep $WARM
SRVPID=$(systemctl show -p MainPID --value wayfarer)
echo "server pid=$SRVPID"

( pidstat -u -p $SRVPID 2 150 > $OUT/$LABEL.server.cpu 2>&1 ) & S1=$!
( mpstat 2 150 > $OUT/$LABEL.all.cpu 2>&1 ) & S2=$!
( pidstat -u -C k6 2 150 > $OUT/$LABEL.k6.cpu 2>&1 ) & S3=$!

cd /opt/wayfarer/backend/cmd/loadtest
k6 run ${K6EXTRA:-} --out csv=$OUT/$LABEL.csv k6/freetext-quiz-rampspike.js > $OUT/$LABEL.k6.log 2>&1
echo "k6 exit=$?"
kill $S1 $S2 $S3 2>/dev/null; wait 2>/dev/null

SJ_AFTER=$($PSQL "SELECT count(*) FROM score_journal;")
echo "score_journal delta=$((SJ_AFTER-SJ_BEFORE))"
echo ""
grep -E "http_req_duration\.|http_req_failed|http_reqs|graphql_errors\.|quiz_completions|quiz_failures|dropped_iter" $OUT/$LABEL.k6.log | head -8
echo ""
echo "--- thresholds ---"
sed -n '/THRESHOLDS/,/TOTAL RESULTS/p' $OUT/$LABEL.k6.log | grep -E "✓|✗|name:" | head -24
echo ""
awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); if(c+0>m)m=c+0} END{printf "server peak %%CPU: %.0f\n", m}' $OUT/$LABEL.server.cpu
awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); s[$1]+=c+0} END{for(t in s) if(s[t]>m)m=s[t]; printf "k6 peak %%CPU: %.0f\n", m}' $OUT/$LABEL.k6.cpu
awk '$2=="all" && $3 ~ /^[0-9.]/ {b=100-$NF; if(b>m)m=b} END{printf "system peak busy: %.1f%%\n", m}' $OUT/$LABEL.all.cpu
