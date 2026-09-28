#!/bin/bash
# Usage: runspike.sh <users> <window> <points> <label>
USERS=${1:-1000}; WINDOW=${2:-10}; POINTS=${3:-10}; LABEL=${4:-run}
export PGPASSWORD=bench
export S_TIME_FORMAT=ISO
export LC_ALL=C
PSQL="psql -h localhost -U bench -d wayfarer_bench -q -Atc"
OUT=/opt/wayfarer/results
mkdir -p $OUT

echo "############################################################"
echo "# $LABEL : users=$USERS window=${WINDOW}s completion_points=$POINTS"
echo "############################################################"

# set quiz points
$PSQL "UPDATE quizzes SET completion_points=$POINTS WHERE id='QZ01LOADTESTFREETEXT00000000';" >/dev/null
# clean prior submissions so every run starts equal
$PSQL "DELETE FROM quiz_responses WHERE submission_id IN (SELECT id FROM quiz_submissions WHERE quiz_id='QZ01LOADTESTFREETEXT00000000');" >/dev/null
$PSQL "DELETE FROM quiz_submissions WHERE quiz_id='QZ01LOADTESTFREETEXT00000000';" >/dev/null
SJ_BEFORE=$($PSQL "SELECT count(*) FROM score_journal;")

systemctl restart wayfarer; sleep 6

SRVPID=$(systemctl show -p MainPID --value wayfarer)
echo "server pid=$SRVPID  score_journal before=$SJ_BEFORE"

# CPU sampler: total + per-process group
( pidstat -u -p $SRVPID 2 200 > $OUT/$LABEL.server.cpu 2>&1 ) &
S1=$!
( pidstat -u -C "postgres" 2 200 > $OUT/$LABEL.pg.cpu 2>&1 ) &
S2=$!
( pidstat -u -C "k6" 2 200 > $OUT/$LABEL.k6.cpu 2>&1 ) &
S3=$!
( mpstat 2 200 > $OUT/$LABEL.all.cpu 2>&1 ) &
S4=$!

cd /opt/wayfarer/backend/cmd/loadtest
k6 run --env SPIKE_USERS=$USERS --env SPIKE_WINDOW=$WINDOW \
       --summary-export=$OUT/$LABEL.summary.json \
       k6/freetext-quiz-spike.js > $OUT/$LABEL.k6.log 2>&1
K6RC=$?

kill $S1 $S2 $S3 $S4 2>/dev/null; wait $S1 $S2 $S3 $S4 2>/dev/null

SJ_AFTER=$($PSQL "SELECT count(*) FROM score_journal;")
echo "score_journal after=$SJ_AFTER  (delta=$((SJ_AFTER-SJ_BEFORE)))"
echo "k6 exit=$K6RC"

echo ""
echo "--- latency (from k6) ---"
grep -E "http_req_duration|http_req_failed|http_reqs|graphql_errors|quiz_completions|quiz_failures|iterations\.|dropped_iterations|vus_max" $OUT/$LABEL.k6.log | head -22

echo ""
echo "--- peak CPU % (of 1200% = 12 threads) ---"
awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); if(c+0>m)m=c+0} END{printf "  server   peak %%CPU: %.0f (100 = 1 core)\n", m}' $OUT/$LABEL.server.cpu
awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); s[$1]+=c+0} END{for(t in s) if(s[t]>m)m=s[t]; printf "  postgres peak %%CPU: %.0f\n", m}' $OUT/$LABEL.pg.cpu
awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); s[$1]+=c+0} END{for(t in s) if(s[t]>m)m=s[t]; printf "  k6       peak %%CPU: %.0f\n", m}' $OUT/$LABEL.k6.cpu
awk '$2=="all" && $3 ~ /^[0-9.]/ {b=100-$NF; if(b>m)m=b} END{printf "  system   peak busy: %.1f%% of all 12 threads\n", m}' $OUT/$LABEL.all.cpu
echo ""
