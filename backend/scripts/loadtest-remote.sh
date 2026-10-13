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
#   LOADTEST_K6_SCRIPT  k6 script, relative to cmd/loadtest
#                       (k6/freetext-quiz-rampspike.js)
#   LOADTEST_SERVICE    systemd unit of the server (wayfarer); "active" means
#                       the live blue/green color (wayfarer@$(cat active))
#   LOADTEST_PSQL       psql command on the box (bench user on wayfarer_bench)
#   LOADTEST_PREP_CMD   command on the box that resets the test data; default
#                       resets the free-text quiz (completion points = $1)
#   LOADTEST_TOKENGEN_ARGS  extra tokengen flags (e.g. -auth0-count 13000)
#   LOADTEST_SSH_KEY    optional key file to use instead of the ssh agent
#   LOADTEST_GENERATOR  may list several hosts (space-separated): each runs
#                       one shard (SHARDS/SHARD env) of the k6 script
#   LOADTEST_SAMPLE_SQL optional query sampled on the box once per second
#                       during a run (e.g. settled bets): server/sql-sample.tsv
#   LOADTEST_EXPORT_SQL "name|query" lines exported as server/<name>.csv after
#                       the run (exact rows for later analysis)
#   LOADTEST_SETTLE_SCRIPT  k6 script run once on the first generator after
#                       all shards finish (shards then get SETTLE=0)
#   LOADTEST_HTTPSTATS_FILE server HTTP stats file copied into the run
#   LOADTEST_WAIT_SQL   query polled after k6 until it returns 0 (work that
#                       outlives the request, e.g. bet settlement)
#
# Every run writes cmd/loadtest/results/<label>-<UTC timestamp>/: manifest,
# per-shard k6 CSV + summary JSON + log + CPU, server/box CPU, SQL samples,
# exports and pg_stat_statements. See cmd/loadtest/analyzerun for the derived files.
set -euo pipefail

LOADTEST_SSH=${LOADTEST_SSH:-root@49.12.121.62}
LOADTEST_BASE_URL=${LOADTEST_BASE_URL:-http://49.12.121.62:8080}
LOADTEST_REMOTE_DIR=${LOADTEST_REMOTE_DIR:-/opt/wayfarer}
LOADTEST_TOKENS=${LOADTEST_TOKENS:-13000}
LOADTEST_WARMUP=${LOADTEST_WARMUP:-18}
LOADTEST_GENERATOR=${LOADTEST_GENERATOR:-}
LOADTEST_GENERATOR_DIR=${LOADTEST_GENERATOR_DIR:-/root/wayfarer-loadtest}
LOADTEST_K6_SCRIPT=${LOADTEST_K6_SCRIPT:-k6/freetext-quiz-rampspike.js}
LOADTEST_SERVICE=${LOADTEST_SERVICE:-wayfarer}
LOADTEST_PSQL=${LOADTEST_PSQL:-PGPASSWORD=bench psql -h localhost -U bench -d wayfarer_bench}
LOADTEST_PREP_CMD=${LOADTEST_PREP_CMD:-}
LOADTEST_TOKENGEN_ARGS=${LOADTEST_TOKENGEN_ARGS:-}
LOADTEST_SAMPLE_SQL=${LOADTEST_SAMPLE_SQL:-}
LOADTEST_EXPORT_SQL=${LOADTEST_EXPORT_SQL:-}
LOADTEST_SETTLE_SCRIPT=${LOADTEST_SETTLE_SCRIPT:-}
LOADTEST_HTTPSTATS_FILE=${LOADTEST_HTTPSTATS_FILE:-}
LOADTEST_WAIT_SQL=${LOADTEST_WAIT_SQL:-}

QUIZ_ID='QZ01LOADTESTFREETEXT00000000'
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
BACKEND_DIR=$(dirname "$SCRIPT_DIR")
LOADTEST_DIR=$BACKEND_DIR/cmd/loadtest
RESULTS_DIR=$LOADTEST_DIR/results
REMOTE_RESULTS=$LOADTEST_REMOTE_DIR/results

# One multiplexed connection per host: the prod firewall role rate-limits SSH
# (ufw limit, 6 connections / 30s) and a run opens many more than that
SSH_OPTS=(-o BatchMode=yes -o ControlMaster=auto -o "ControlPath=/tmp/loadtest-ssh-%C" -o ControlPersist=120)
# LOADTEST_SSH_KEY: use this key file only, bypassing the ssh agent
if [ -n "${LOADTEST_SSH_KEY:-}" ]; then
    SSH_OPTS+=(-o IdentityAgent=none -o IdentitiesOnly=yes -i "$LOADTEST_SSH_KEY")
fi
rssh() { ssh "${SSH_OPTS[@]}" "$LOADTEST_SSH" "$@"; }
rscp() { scp -q "${SSH_OPTS[@]}" "$@"; }
rpsql() {
    # shellcheck disable=SC2029  # $1 expands locally on purpose
    rssh "$LOADTEST_PSQL -q -Atc \"$1\""
}
# service resolves LOADTEST_SERVICE on the box (blue/green: the active color)
service() {
    if [ "$LOADTEST_SERVICE" = active ]; then
        echo "wayfarer@$(rssh "cat $LOADTEST_REMOTE_DIR/active")"
    else
        echo "$LOADTEST_SERVICE"
    fi
}

cmd_config() {
    # Built here and shipped, so the box needs no Go toolchain
    echo "Building tokengen (linux/amd64)..."
    (cd "$BACKEND_DIR" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$RESULTS_DIR/tokengen.linux" ./cmd/loadtest/tokengen)
    rssh "mkdir -p $LOADTEST_REMOTE_DIR/loadtest"
    rscp "$RESULTS_DIR/tokengen.linux" "$LOADTEST_SSH:$LOADTEST_REMOTE_DIR/loadtest/tokengen"
    echo "Minting $LOADTEST_TOKENS tokens on $LOADTEST_SSH (baseUrl=$LOADTEST_BASE_URL)..."
    # shellcheck disable=SC2029  # locals expand here on purpose
    rssh "set -a; . $LOADTEST_REMOTE_DIR/wayfarer.env; set +a; \
        cd $LOADTEST_REMOTE_DIR/loadtest && \
        ./tokengen -limit $LOADTEST_TOKENS -base-url $LOADTEST_BASE_URL -output ./config.json $LOADTEST_TOKENGEN_ARGS"
    rscp "$LOADTEST_SSH:$LOADTEST_REMOTE_DIR/loadtest/config.json" "$LOADTEST_DIR/config.json"
    echo "Fetched $LOADTEST_DIR/config.json:"
    jq -r '" baseUrl: \(.baseUrl)\n tokens:  \(.tokens | length)"' "$LOADTEST_DIR/config.json"
}

cmd_prep() {
    local points=${1:-10}
    local svc
    svc=$(service)
    if [ -n "$LOADTEST_PREP_CMD" ]; then
        # Server stopped while the data is reset: its background workers
        # (e.g. the leaderboard drain) would race the reset's deletes
        echo "Stopping $svc, resetting test data ($LOADTEST_PREP_CMD $points)..."
        rssh "systemctl stop $svc && $LOADTEST_PREP_CMD $points"
    else
        echo "Resetting quiz state (completion_points=$points)..."
        rpsql "UPDATE quizzes SET completion_points=$points WHERE id='$QUIZ_ID';" >/dev/null
        rpsql "DELETE FROM quiz_responses WHERE submission_id IN (SELECT id FROM quiz_submissions WHERE quiz_id='$QUIZ_ID');" >/dev/null
        rpsql "DELETE FROM quiz_submissions WHERE quiz_id='$QUIZ_ID';" >/dev/null
    fi
    echo "Restarting $svc..."
    rssh "systemctl restart $svc"
    echo "Waiting ${LOADTEST_WARMUP}s for token-warmer boot pass..."
    sleep "$LOADTEST_WARMUP"
    curl -sf --max-time 5 "$LOADTEST_BASE_URL/health" >/dev/null || {
        echo "server did not come back healthy at $LOADTEST_BASE_URL/health" >&2
        exit 1
    }
}

# Samplers write to $RUN_REMOTE on the box (and $GEN_DIR/<run> on each
# generator), 1s resolution, ISO timestamps, capped at 1h; stopped after k6.
start_samplers() {
    local svc
    svc=$(service)
    rssh "mkdir -p $RUN_REMOTE && chmod 777 $RUN_REMOTE; \
        SRVPID=\$(systemctl show -p MainPID --value $svc); \
        export S_TIME_FORMAT=ISO LC_ALL=C; \
        nohup pidstat -u -r -h -p \$SRVPID 1 3600 > $RUN_REMOTE/server.pidstat 2>&1 & \
        nohup pidstat -u -h -C postgres 1 3600 > $RUN_REMOTE/postgres.pidstat 2>&1 & \
        nohup mpstat 1 3600 > $RUN_REMOTE/server-box.mpstat 2>&1 & \
        true"
    if [ -n "$LOADTEST_SAMPLE_SQL" ]; then
        # The query travels as an argument, so its quotes need no escaping
        printf '%s\n' '#!/bin/bash' \
            "for i in \$(seq 3600); do echo \"\$(date -u +%Y-%m-%dT%H:%M:%SZ) \$($LOADTEST_PSQL -Atc \"\$1\")\"; sleep 1; done" |
            rssh "cat > $RUN_REMOTE/sql-sampler.sh"
        # shellcheck disable=SC2029  # locals expand here on purpose
        rssh "nohup bash $RUN_REMOTE/sql-sampler.sh $(printf %q "$LOADTEST_SAMPLE_SQL") \
            > $RUN_REMOTE/sql-sample.tsv 2>&1 & echo \$! > $RUN_REMOTE/sql-sampler.pid"
    fi
    # Reset query stats so pg_stat_statements covers exactly this run
    rpsql "SELECT pg_stat_statements_reset();" >/dev/null 2>&1 || true
}

stop_samplers() {
    rssh "pkill -x pidstat; pkill -x mpstat; \
        [ -f $RUN_REMOTE/sql-sampler.pid ] && kill \$(cat $RUN_REMOTE/sql-sampler.pid) 2>/dev/null; \
        rm -f $RUN_REMOTE/sql-sampler.pid; true"
}

# collect_server_data copies the box's samplers and exports the run's rows
# and query stats as CSV (LOADTEST_EXPORT_SQL: "name|query" lines)
collect_server_data() {
    rpsql "\\copy (SELECT * FROM pg_stat_statements WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database()) ORDER BY total_exec_time DESC) TO '$RUN_REMOTE/pg_stat_statements.csv' CSV HEADER" \
        >/dev/null 2>&1 || echo "  (pg_stat_statements not available)"
    if [ -n "$LOADTEST_EXPORT_SQL" ]; then
        local name query
        while IFS='|' read -r name query; do
            [ -n "$name" ] || continue
            # </dev/null: ssh would otherwise eat the rest of the loop's input
            rpsql "\\copy ($query) TO '$RUN_REMOTE/$name.csv' CSV HEADER" </dev/null >/dev/null ||
                echo "  export $name failed"
        done <<< "$LOADTEST_EXPORT_SQL"
    fi
    if [ -n "$LOADTEST_HTTPSTATS_FILE" ]; then
        rssh "cp $LOADTEST_HTTPSTATS_FILE $RUN_REMOTE/httpstats.jsonl 2>/dev/null; true"
    fi
    rssh "cd $RUN_REMOTE && chmod a+r * 2>/dev/null; true"
    mkdir -p "$RUN_DIR/server"
    rscp "$LOADTEST_SSH:$RUN_REMOTE/*" "$RUN_DIR/server/" || true
}

# write_manifest records what produced the data: code, target, knobs, hosts
write_manifest() {
    local started=$1 finished=$2 k6_exit=$3
    shift 3
    local server_env
    # Non-secret server settings only
    server_env=$(rssh "grep -E '^(ENVIRONMENT|DB_|SERVER_|TLS_AUTO_DOMAINS|GOGC|GOMEMLIMIT|LOG_LEVEL|OTEL_ENABLED|HTTP_STATS)' $LOADTEST_REMOTE_DIR/wayfarer.env $LOADTEST_REMOTE_DIR/env.* 2>/dev/null" || true)
    local pg_settings
    pg_settings=$(rpsql "SELECT name || '=' || setting FROM pg_settings WHERE name IN ('max_connections','shared_buffers','effective_cache_size','work_mem','synchronous_commit','max_wal_size','random_page_cost','server_version')" 2>/dev/null || true)
    local hosts
    hosts=$(for h in "$LOADTEST_SSH" "${GENERATORS[@]}"; do
        ssh "${SSH_OPTS[@]}" "$h" 'echo "$(hostname) nproc=$(nproc) mem=$(free -g | awk "/Mem:/{print \$2}")G kernel=$(uname -r)"' 2>/dev/null | sed "s|^|$h |"
    done)
    local commit dirty
    commit=$(git -C "$BACKEND_DIR" rev-parse HEAD)
    dirty=$(git -C "$BACKEND_DIR" status --porcelain | wc -l | tr -d ' ')
    jq -n \
        --arg label "$LABEL" --arg run_id "$RUN_ID" \
        --arg started "$started" --arg finished "$finished" \
        --argjson k6_exit "$k6_exit" --arg k6_script "$LOADTEST_K6_SCRIPT" \
        --arg k6_args "$*" --arg base_url "$LOADTEST_BASE_URL" \
        --arg server "$LOADTEST_SSH" --arg generators "${GENERATORS[*]}" \
        --arg commit "$commit" --argjson dirty "$dirty" \
        --arg server_env "$server_env" --arg pg "$pg_settings" --arg hosts "$hosts" \
        'def lines: split("\n") | map(select(test("\\S")));
        {
            label: $label, run_id: $run_id,
            started_utc: $started, finished_utc: $finished,
            k6_exit: $k6_exit, k6_script: $k6_script,
            k6_args: $k6_args, base_url: $base_url,
            server: $server, generators: ($generators | split(" ") | map(select(. != ""))),
            git_commit: $commit, git_uncommitted_files: $dirty,
            server_env: ($server_env | lines), postgres: ($pg | lines),
            hosts: ($hosts | lines)
        }' > "$RUN_DIR/manifest.json"
}

# run_generator <host> <shard> <shards> [k6 args...]: one k6 shard on a
# generator, with its own CPU samplers; outputs land in $GEN_DIR/$RUN_ID
run_generator() {
    local host=$1 shard=$2 shards=$3
    shift 3
    local out=$LOADTEST_GENERATOR_DIR/$RUN_ID name=shard$shard
    local settle_env=""
    [ -n "$LOADTEST_SETTLE_SCRIPT" ] && settle_env="--env SETTLE=0"
    # shellcheck disable=SC2029  # locals expand here on purpose
    ssh "${SSH_OPTS[@]}" "$host" \
        "ulimit -n 65535; mkdir -p $out; cd $LOADTEST_GENERATOR_DIR && \
         export S_TIME_FORMAT=ISO LC_ALL=C; \
         (pidstat -u -r -h -C k6 1 3600 > $out/$name.k6.pidstat 2>&1 &); \
         (mpstat 1 3600 > $out/$name.box.mpstat 2>&1 &); \
         k6 run --env BASE_URL=$LOADTEST_BASE_URL --env SHARDS=$shards --env SHARD=$shard $settle_env $* \
            --out csv=$out/$name.csv.gz --summary-export=$out/$name.summary.json \
            $LOADTEST_K6_SCRIPT > $out/$name.k6.log 2>&1; rc=\$?; \
         pkill -x pidstat; pkill -x mpstat; exit \$rc"
}

cmd_run() {
    LABEL=${1:?usage: loadtest-remote.sh run <label> [k6 args...]}
    shift
    RUN_ID=$LABEL-$(date -u +%Y%m%dT%H%M%SZ)
    RUN_DIR=$RESULTS_DIR/$RUN_ID
    RUN_REMOTE=$REMOTE_RESULTS/$RUN_ID
    read -r -a GENERATORS <<< "$LOADTEST_GENERATOR"
    mkdir -p "$RUN_DIR"
    echo "Run $RUN_ID -> $RUN_DIR"

    cmd_prep
    local started
    started=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    start_samplers

    local k6_exit=0
    if [ ${#GENERATORS[@]} -gt 0 ]; then
        local shards=${#GENERATORS[@]} i pids=()
        for i in "${!GENERATORS[@]}"; do
            echo "Syncing loadtest tree to ${GENERATORS[$i]}..."
            rsync -az --delete -e "ssh ${SSH_OPTS[*]}" "$LOADTEST_DIR/k6" "$LOADTEST_DIR/config.json" \
                "${GENERATORS[$i]}:$LOADTEST_GENERATOR_DIR/"
        done
        echo "Running $shards k6 shard(s) against $LOADTEST_BASE_URL..."
        for i in "${!GENERATORS[@]}"; do
            run_generator "${GENERATORS[$i]}" "$i" "$shards" "$@" &
            pids+=($!)
        done
        for i in "${!pids[@]}"; do
            local rc=0
            wait "${pids[$i]}" || rc=$?
            echo "  shard $i on ${GENERATORS[$i]} exited $rc"
            [ "$rc" -ne 0 ] && k6_exit=$rc
        done
        if [ -n "$LOADTEST_SETTLE_SCRIPT" ]; then
            echo "All shards done; settling on ${GENERATORS[0]} ($LOADTEST_SETTLE_SCRIPT)..."
            local out=$LOADTEST_GENERATOR_DIR/$RUN_ID rc=0
            # shellcheck disable=SC2029  # locals expand here on purpose
            ssh "${SSH_OPTS[@]}" "${GENERATORS[0]}" \
                "cd $LOADTEST_GENERATOR_DIR && k6 run --env BASE_URL=$LOADTEST_BASE_URL \
                    --out csv=$out/settle.csv.gz --summary-export=$out/settle.summary.json \
                    $LOADTEST_SETTLE_SCRIPT > $out/settle.k6.log 2>&1" || rc=$?
            echo "  settle exited $rc"
            [ "$rc" -ne 0 ] && k6_exit=$rc
        fi
        for i in "${!GENERATORS[@]}"; do
            mkdir -p "$RUN_DIR/generator$i"
            rscp "${GENERATORS[$i]}:$LOADTEST_GENERATOR_DIR/$RUN_ID/*" "$RUN_DIR/generator$i/" || true
        done
    else
        echo "Running k6 locally against $LOADTEST_BASE_URL..."
        k6 run --env "BASE_URL=$LOADTEST_BASE_URL" "$@" \
            --out "csv=$RUN_DIR/local.csv.gz" --summary-export="$RUN_DIR/local.summary.json" \
            "$LOADTEST_DIR/$LOADTEST_K6_SCRIPT" 2>&1 | tee "$RUN_DIR/local.k6.log" || k6_exit=$?
    fi

    if [ -n "$LOADTEST_WAIT_SQL" ]; then
        # Work can outlive the k6 request (e.g. settlement past the server's
        # write timeout): keep sampling until the query returns 0 (max 30 min)
        echo "Waiting for the server to finish ($LOADTEST_WAIT_SQL)..."
        local left i
        for i in $(seq 1 360); do
            left=$(rpsql "$LOADTEST_WAIT_SQL" 2>/dev/null || echo "?")
            [ "$left" = 0 ] && break
            [ $((i % 6)) = 1 ] && echo "  $left left"
            sleep 5
        done
        echo "  done waiting ($left left)"
    fi
    stop_samplers
    local finished
    finished=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    echo "Collecting server data..."
    collect_server_data
    write_manifest "$started" "$finished" "$k6_exit" "$@"

    echo ""
    echo "=== $RUN_ID summary ==="
    echo "k6 exit=$k6_exit"
    (cd "$BACKEND_DIR" && go run ./cmd/loadtest/analyzerun "$RUN_DIR") || true
    echo "Data: $RUN_DIR"
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
