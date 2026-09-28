#!/bin/bash
# Off-box load testing: run k6 on this machine against the bare-metal bench
# box, keeping server + Postgres remote. This is the "second machine" that
# notes/ram-first-architecture.html §18 calls for — on-box runs had k6 eating
# 1.7x the server's CPU and starving the thing being measured.
#
# Usage:
#   loadtest-remote.sh config              # mint tokens on the box, fetch config.json
#   loadtest-remote.sh prep [points]       # reset quiz state + restart server
#   loadtest-remote.sh run <label> [k6 --env args...]
#   loadtest-remote.sh smoke               # scaled-down run to verify plumbing
#
# Env knobs (defaults match the current bench box):
#   LOADTEST_SSH        ssh destination            (root@49.12.121.62)
#   LOADTEST_BASE_URL   URL k6 targets             (http://49.12.121.62:8080)
#   LOADTEST_REMOTE_DIR wayfarer dir on the box    (/opt/wayfarer)
#   LOADTEST_TOKENS     tokengen -limit            (13000, rampspike needs >=12100)
#   LOADTEST_WARMUP     seconds after restart for the Firebase token-warmer
#                       boot pass (18)
#   LOADTEST_GENERATOR  optional ssh destination to run k6 on instead of this
#                       machine (e.g. root@167.233.228.181). Needed for the
#                       full 10k spike: >~5k concurrent connections from one
#                       IP through an office NAT / across Hetzner's DDoS
#                       heuristics get their connections killed (EOFs). Use a
#                       VM in the same DC as the box; k6 output still streams
#                       here and results are fetched back.
set -euo pipefail

LOADTEST_SSH=${LOADTEST_SSH:-root@49.12.121.62}
LOADTEST_BASE_URL=${LOADTEST_BASE_URL:-http://49.12.121.62:8080}
LOADTEST_REMOTE_DIR=${LOADTEST_REMOTE_DIR:-/opt/wayfarer}
LOADTEST_TOKENS=${LOADTEST_TOKENS:-13000}
LOADTEST_WARMUP=${LOADTEST_WARMUP:-18}
LOADTEST_GENERATOR=${LOADTEST_GENERATOR:-}
LOADTEST_GENERATOR_DIR=${LOADTEST_GENERATOR_DIR:-/root/wayfarer-loadtest}

QUIZ_ID='QZ01LOADTESTFREETEXT00000000'
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
BACKEND_DIR=$(dirname "$SCRIPT_DIR")
LOADTEST_DIR=$BACKEND_DIR/cmd/loadtest
RESULTS_DIR=$LOADTEST_DIR/results
REMOTE_RESULTS=$LOADTEST_REMOTE_DIR/results

rssh() { ssh -o BatchMode=yes "$LOADTEST_SSH" "$@"; }
rpsql() {
    # shellcheck disable=SC2029  # $1 expands locally on purpose
    rssh "PGPASSWORD=bench psql -h localhost -U bench -d wayfarer_bench -q -Atc \"$1\""
}

cmd_config() {
    echo "Minting $LOADTEST_TOKENS tokens on $LOADTEST_SSH (baseUrl=$LOADTEST_BASE_URL)..."
    rssh "set -a; . $LOADTEST_REMOTE_DIR/wayfarer.env; set +a; \
        cd $LOADTEST_REMOTE_DIR/backend/cmd/loadtest && \
        go run ./tokengen -limit $LOADTEST_TOKENS -base-url $LOADTEST_BASE_URL -output ./config.json"
    scp -q "$LOADTEST_SSH:$LOADTEST_REMOTE_DIR/backend/cmd/loadtest/config.json" "$LOADTEST_DIR/config.json"
    echo "Fetched $LOADTEST_DIR/config.json:"
    python3 -c "import json;c=json.load(open('$LOADTEST_DIR/config.json'));print(' baseUrl:',c['baseUrl']);print(' tokens: ',len(c['tokens']))"
}

cmd_prep() {
    local points=${1:-10}
    echo "Resetting quiz state (completion_points=$points) and restarting wayfarer..."
    rpsql "UPDATE quizzes SET completion_points=$points WHERE id='$QUIZ_ID';" >/dev/null
    rpsql "DELETE FROM quiz_responses WHERE submission_id IN (SELECT id FROM quiz_submissions WHERE quiz_id='$QUIZ_ID');" >/dev/null
    rpsql "DELETE FROM quiz_submissions WHERE quiz_id='$QUIZ_ID';" >/dev/null
    rssh "systemctl restart wayfarer"
    echo "Waiting ${LOADTEST_WARMUP}s for token-warmer boot pass..."
    sleep "$LOADTEST_WARMUP"
    curl -sf --max-time 5 "$LOADTEST_BASE_URL/health" >/dev/null || {
        echo "server did not come back healthy at $LOADTEST_BASE_URL/health" >&2
        exit 1
    }
}

start_samplers() {
    local label=$1
    # 2s interval x 300 samples = 10 min hard cap; stopped early after k6 exits.
    rssh "mkdir -p $REMOTE_RESULTS; \
        SRVPID=\$(systemctl show -p MainPID --value wayfarer); \
        export S_TIME_FORMAT=ISO LC_ALL=C; \
        nohup pidstat -u -p \$SRVPID 2 300 > $REMOTE_RESULTS/$label.server.cpu 2>&1 & \
        nohup pidstat -u -C postgres 2 300 > $REMOTE_RESULTS/$label.pg.cpu 2>&1 & \
        nohup mpstat 2 300 > $REMOTE_RESULTS/$label.all.cpu 2>&1 & \
        true"
}

stop_samplers() {
    rssh "pkill -x pidstat; pkill -x mpstat; true"
}

cmd_run() {
    local label=${1:?usage: loadtest-remote.sh run <label> [k6 args...]}
    shift
    mkdir -p "$RESULTS_DIR"

    cmd_prep
    local sj_before
    sj_before=$(rpsql "SELECT count(*) FROM score_journal;")
    start_samplers "$label"

    local k6_exit=0
    if [ -n "$LOADTEST_GENERATOR" ]; then
        echo "Syncing loadtest tree to generator $LOADTEST_GENERATOR..."
        rsync -az --delete "$LOADTEST_DIR/k6" "$LOADTEST_DIR/config.json" \
            "$LOADTEST_GENERATOR:$LOADTEST_GENERATOR_DIR/"
        echo "Running k6 on $LOADTEST_GENERATOR against $LOADTEST_BASE_URL (label=$label)..."
        # shellcheck disable=SC2029  # locals expand here on purpose
        ssh -o BatchMode=yes "$LOADTEST_GENERATOR" \
            "ulimit -n 65535; cd $LOADTEST_GENERATOR_DIR && \
             k6 run --env BASE_URL=$LOADTEST_BASE_URL $* \
                --out csv=$label.csv.gz k6/freetext-quiz-rampspike.js" \
            2>&1 | tee "$RESULTS_DIR/$label.k6.log" || k6_exit=$?
        scp -q "$LOADTEST_GENERATOR:$LOADTEST_GENERATOR_DIR/$label.csv.gz" "$RESULTS_DIR/" || true
    else
        echo "Running k6 locally against $LOADTEST_BASE_URL (label=$label)..."
        k6 run --env "BASE_URL=$LOADTEST_BASE_URL" "$@" \
            --out "csv=$RESULTS_DIR/$label.csv" \
            "$LOADTEST_DIR/k6/freetext-quiz-rampspike.js" 2>&1 | tee "$RESULTS_DIR/$label.k6.log" || k6_exit=$?
    fi

    stop_samplers
    local sj_after
    sj_after=$(rpsql "SELECT count(*) FROM score_journal;")
    scp -q "$LOADTEST_SSH:$REMOTE_RESULTS/$label.*.cpu" "$RESULTS_DIR/" || true

    echo ""
    echo "=== $label summary ==="
    echo "k6 exit=$k6_exit  score_journal delta=$((sj_after - sj_before))"
    grep -E "http_req_duration\.|http_req_failed|http_reqs|graphql_errors\.|quiz_completions|quiz_failures|dropped_iter" \
        "$RESULTS_DIR/$label.k6.log" | head -8 || true
    awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); if(c+0>m)m=c+0} END{printf "server peak %%CPU: %.0f\n", m}' \
        "$RESULTS_DIR/$label.server.cpu" 2>/dev/null || true
    awk '$1 ~ /:/ && $NF != "Command" {c=$(NF-2); s[$1]+=c+0} END{for(t in s) if(s[t]>m)m=s[t]; printf "postgres peak %%CPU: %.0f\n", m}' \
        "$RESULTS_DIR/$label.pg.cpu" 2>/dev/null || true
    awk '$2=="all" && $3 ~ /^[0-9.]/ {b=100-$NF; if(b>m)m=b} END{printf "box peak busy: %.1f%%\n", m}' \
        "$RESULTS_DIR/$label.all.cpu" 2>/dev/null || true
    exit "$k6_exit"
}

cmd_smoke() {
    # ~200 arrivals instead of 10,000 — verifies tokens, network path and
    # summary plumbing without a real spike.
    cmd_run smoke --env RAMP_SCALE=0.02 "$@"
}

case "${1:-}" in
config) shift; cmd_config "$@" ;;
prep) shift; cmd_prep "$@" ;;
run) shift; cmd_run "$@" ;;
smoke) shift; cmd_smoke "$@" ;;
*)
    sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
