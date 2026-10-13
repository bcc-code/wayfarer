#!/bin/bash
# Betting load test on the dedicated load-test pair (Hetzner CCX33, Helsinki):
#   interact-lt-server  2.29.101.91  prod copy: native wayfarer@{blue,green},
#                                    Postgres 17, TLS via the server's own ACME
#                                    (Let's Encrypt) on 2-29-101-91.sslip.io
#   interact-lt-client  2.29.97.60   k6 generators (same DC, no NAT), one
#   interact-lt-client2 2.29.107.165 shard of the arrivals each
# See notes/loadtest-offbox.md ("Betting load test on the LT pair").
#
# Usage (from backend/):
#   loadtest-lt.sh provision   ansible: prod roles on the server, tuning + k6 on the client
#   loadtest-lt.sh env         write /opt/wayfarer/wayfarer.env on the server (once)
#   loadtest-lt.sh deploy      build a release like CI and blue/green deploy it
#   loadtest-lt.sh seed        prod-scale data + betting quiz (once, empty DB)
#   loadtest-lt.sh config      mint user/Auth0/admin tokens, fetch config.json
#   loadtest-lt.sh smoke       ~200 users (RAMP_SCALE=0.02), settles bets
#   loadtest-lt.sh run <label> [k6 --env args...]   full 10k betting rampspike
#
# Generated secrets (DB password, JWT secret, Auth0 load-test key) are kept in
# cmd/loadtest/results/lt-secrets.env (gitignored) so reruns reuse them; the
# dedicated SSH key for the pair is cmd/loadtest/results/lt_ed25519.
set -euo pipefail

LT_SERVER=${LT_SERVER:-root@2.29.101.91}
# k6 generators, one shard each (one CCX33 saturates before 10k arrivals)
LT_CLIENTS=${LT_CLIENTS:-root@2.29.97.60 root@2.29.107.165}
LT_DOMAIN=${LT_DOMAIN:-2-29-101-91.sslip.io}
LT_ADMIN_USER=US00000000000000000000000001

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
BACKEND_DIR=$(dirname "$SCRIPT_DIR")
REPO_DIR=$(dirname "$BACKEND_DIR")
LOADTEST_DIR=$BACKEND_DIR/cmd/loadtest
RESULTS_DIR=$LOADTEST_DIR/results
SECRETS=$RESULTS_DIR/lt-secrets.env
REMOTE=/opt/wayfarer

# Dedicated key for the LT pair (results/lt_ed25519, gitignored), so runs
# don't depend on the ssh agent; falls back to the agent when absent
LT_SSH_KEY=${LT_SSH_KEY:-$RESULTS_DIR/lt_ed25519}
SSH_OPTS=(-o BatchMode=yes -o ControlMaster=auto -o "ControlPath=/tmp/loadtest-ssh-%C" -o ControlPersist=120)
SSH_KEY_OPTS=()
if [ -f "$LT_SSH_KEY" ]; then
    SSH_KEY_OPTS=(-o IdentityAgent=none -o IdentitiesOnly=yes -i "$LT_SSH_KEY")
    SSH_OPTS+=("${SSH_KEY_OPTS[@]}")
else
    LT_SSH_KEY=
fi
rssh() { ssh "${SSH_OPTS[@]}" "$LT_SERVER" "$@"; }
rscp() { scp -q "${SSH_OPTS[@]}" "$@"; }

secrets() {
    mkdir -p "$RESULTS_DIR"
    if [ ! -f "$SECRETS" ]; then
        echo "Generating load-test secrets in $SECRETS..."
        local key
        key=$(openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 2>/dev/null | base64 | tr -d '\n')
        umask 077
        cat > "$SECRETS" <<EOF
LT_DB_PASSWORD=$(openssl rand -hex 24)
LT_JWT_SECRET=$(openssl rand -hex 32)
LT_AUTH0_KEY=$key
EOF
    fi
    # shellcheck disable=SC1090
    . "$SECRETS"
}

cmd_provision() {
    secrets
    cd "$REPO_DIR/ansible"
    local key_args=()
    if [ -n "$LT_SSH_KEY" ]; then
        key_args=(--private-key "$LT_SSH_KEY" --ssh-common-args "-o IdentityAgent=none -o IdentitiesOnly=yes")
    fi
    ansible-playbook -i inventory-loadtest.yml loadtest.yml -e "interact_db_password=$LT_DB_PASSWORD" "${key_args[@]}" "$@"
}

cmd_env() {
    secrets
    if rssh "test -f $REMOTE/wayfarer.env"; then
        echo "$REMOTE/wayfarer.env already exists on $LT_SERVER, leaving it alone"
        return
    fi
    echo "Writing $REMOTE/wayfarer.env on $LT_SERVER..."
    # Prod's env minus the integrations a load test must not touch (Firestore,
    # S3, Pub/Sub, push, mail, Phrase, SSF); pool/admin port come from env.<color>
    rssh "umask 077; cat > $REMOTE/wayfarer.env" <<EOF
ENVIRONMENT=production
DATABASE_URL=postgres://interact:$LT_DB_PASSWORD@127.0.0.1:5432/interact
DB_CONN_MAX_IDLE_TIME=10m
SERVER_PORT=443
SERVER_REUSEPORT=true
TLS_AUTO_DOMAINS=$LT_DOMAIN
TLS_AUTO_CACHE_DIR=$REMOTE/autocert
TLS_AUTO_HTTP_ADDR=:80
JWT_SECRET=$LT_JWT_SECRET
JWT_ISSUER=wayfarer
AUTH0_LOADTEST_PRIVATE_KEY=$LT_AUTH0_KEY
AUTH0_JWKS_URL=https://$LT_DOMAIN/jwks.json
AUTH0_JWT_ISSUER=https://login.bcc.no/
GOGC=600
GOMEMLIMIT=12GiB
LOG_LEVEL=warn
OTEL_ENABLED=false
PHRASE_ENABLED=false
PUBSUB_ENABLED=false
HTTP_STATS_FILE=$REMOTE/httpstats.jsonl
HTTP_STATS_INTERVAL=15s
EOF
}

cmd_deploy() {
    local sha
    sha=$(git -C "$REPO_DIR" rev-parse --short HEAD)
    if [ -n "$(git -C "$REPO_DIR" status --porcelain)" ]; then
        sha="$sha-dirty-$(date +%s)"
    fi
    local release=$RESULTS_DIR/release-$sha
    rm -rf "$release"
    mkdir -p "$release/static"
    echo "Building wayfarer-server (linux/amd64) for $sha..."
    (cd "$BACKEND_DIR" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$release/wayfarer-server" ./cmd/server)
    # The SPA is not part of the API load; LT_FRONTEND=1 bundles it like CI
    if [ "${LT_FRONTEND:-0}" = 1 ]; then
        (cd "$REPO_DIR/frontend" && pnpm run build)
        cp -r "$REPO_DIR/frontend/.output/public/." "$release/static/"
    fi
    tar -czf "$release.tar.gz" -C "$release" .
    echo "Shipping and deploying $sha..."
    rssh "mkdir -p $REMOTE/releases/$sha"
    rscp "$release.tar.gz" "$LT_SERVER:/tmp/release-$sha.tar.gz"
    rssh "tar -xzf /tmp/release-$sha.tar.gz -C $REMOTE/releases/$sha && rm /tmp/release-$sha.tar.gz && \
        $REMOTE/bin/deploy.sh $sha"
    echo "Waiting for the certificate and https://$LT_DOMAIN/health..."
    for _ in $(seq 1 30); do
        if curl -sf --max-time 5 "https://$LT_DOMAIN/health" >/dev/null; then
            echo "healthy over TLS:"
            echo | openssl s_client -connect "$LT_DOMAIN:443" -servername "$LT_DOMAIN" 2>/dev/null |
                openssl x509 -noout -text 2>/dev/null |
                grep -E "Issuer:|Subject:|Not After|Public Key Algorithm|ASN1 OID" || true
            return
        fi
        sleep 2
    done
    echo "https://$LT_DOMAIN/health did not come up" >&2
    exit 1
}

sync_bench() {
    rssh "mkdir -p $REMOTE/loadtest"
    rscp "$LOADTEST_DIR/bench/betting_box.sh" "$LOADTEST_DIR/bench/gendata.sql" "$LOADTEST_DIR/bench/gendata2.sql" \
        "$LT_SERVER:$REMOTE/loadtest/"
}

cmd_seed() {
    sync_bench
    rssh "$REMOTE/loadtest/betting_box.sh seed && $REMOTE/loadtest/betting_box.sh reset"
}

QUIZ=QZ01LOADTESTBETTING000000000
# Exact rows exported after every run (server/<name>.csv in the run dir)
EXPORT_SQL="bets|SELECT r.id AS response_id, s.user_id, r.question_id, r.bet_amount, r.is_correct, r.points_earned, r.answered_at, r.time_spent_seconds, r.score_journal_id, sj.created_at AS settled_at FROM quiz_responses r JOIN quiz_submissions s ON s.id = r.submission_id LEFT JOIN score_journal sj ON sj.id = r.score_journal_id WHERE s.quiz_id = '$QUIZ' ORDER BY r.answered_at
submissions|SELECT id, user_id, started_at, completed_at, score, max_score, points_awarded, auto_submitted FROM quiz_submissions WHERE quiz_id = '$QUIZ' ORDER BY started_at
bet_journal|SELECT sj.id, sj.user_id, sj.source_id AS response_id, sj.points, CASE WHEN sj.points < 0 THEN 'stake' ELSE 'winnings' END AS kind, sj.created_at FROM score_journal sj WHERE sj.source_type = 'BET' AND sj.source_id IN (SELECT r.id FROM quiz_responses r JOIN quiz_submissions s ON s.id = r.submission_id WHERE s.quiz_id = '$QUIZ') ORDER BY sj.created_at
quiz_journal|SELECT id, user_id, points, created_at FROM score_journal WHERE source_type = 'QUIZ' AND source_id = '$QUIZ' ORDER BY created_at"

# Shared knobs for scripts/loadtest-remote.sh. Warmup 40s: with
# AUTH0_LOADTEST_PRIVATE_KEY the server fetches its own /jwks.json, which
# fails at boot (not listening yet) and is retried every 30s; Auth0 dances
# before that get 401. Prod loads the real Auth0 JWKS at boot.
remote() {
    sync_bench
    LOADTEST_SSH=$LT_SERVER \
    LOADTEST_SSH_KEY=$LT_SSH_KEY \
    LOADTEST_GENERATOR="$LT_CLIENTS" \
    LOADTEST_SETTLE_SCRIPT=k6/betting-settle.js \
    LOADTEST_HTTPSTATS_FILE=$REMOTE/httpstats.jsonl \
    LOADTEST_EXPORT_SQL="$EXPORT_SQL" \
    LOADTEST_WAIT_SQL="SELECT count(*) FROM quiz_responses r JOIN quiz_submissions s ON s.id = r.submission_id WHERE s.quiz_id = '$QUIZ' AND r.bet_amount > 0 AND r.score_journal_id IS NULL" \
    LOADTEST_BASE_URL=https://$LT_DOMAIN \
    LOADTEST_REMOTE_DIR=$REMOTE \
    LOADTEST_SERVICE=active \
    LOADTEST_WARMUP=${LOADTEST_WARMUP:-40} \
    LOADTEST_PSQL="sudo -u postgres psql -d interact" \
    LOADTEST_PREP_CMD="$REMOTE/loadtest/betting_box.sh reset" \
    LOADTEST_K6_SCRIPT=k6/betting-quiz-rampspike.js \
    LOADTEST_TOKENGEN_ARGS="-auth0-count ${LOADTEST_TOKENS:-13000} -admin-user $LT_ADMIN_USER" \
    LOADTEST_SAMPLE_SQL="SELECT count(*) FILTER (WHERE r.score_journal_id IS NOT NULL) || ' settled / ' || count(*) || ' bets' FROM quiz_responses r JOIN quiz_submissions s ON s.id = r.submission_id WHERE s.quiz_id = 'QZ01LOADTESTBETTING000000000' AND r.bet_amount > 0" \
        "$SCRIPT_DIR/loadtest-remote.sh" "$@"
}

# Per-user check after a run: every user's settled bet entries net out to what
# their responses say they won or lost (stake + winnings journal rows per bet)
check_settlement() {
    echo ""
    echo "=== settlement check ==="
    rssh "sudo -u postgres psql -d interact -Atc \"
        WITH bets AS (
            SELECT r.id, r.bet_amount, r.points_earned, r.score_journal_id
            FROM quiz_responses r JOIN quiz_submissions s ON s.id = r.submission_id
            WHERE s.quiz_id = 'QZ01LOADTESTBETTING000000000' AND r.bet_amount > 0
        ), journal AS (
            SELECT source_id, SUM(points) AS net, count(*) AS n
            FROM score_journal WHERE source_type = 'BET' AND source_id IN (SELECT id FROM bets)
            GROUP BY source_id
        )
        SELECT 'bets=' || count(*)
            || ' unsettled=' || count(*) FILTER (WHERE b.score_journal_id IS NULL)
            || ' journal_mismatch=' || count(*) FILTER (WHERE b.score_journal_id IS NOT NULL
                AND (j.n IS DISTINCT FROM 2 OR j.net IS DISTINCT FROM b.points_earned))
            || ' net_points=' || COALESCE(SUM(j.net), 0)
        FROM bets b LEFT JOIN journal j ON j.source_id = b.id;\""
}

# Prod-like client behaviour minus Firebase (disabled on the LT server):
# 3% Auth0 dance, no Firebase token fetches
K6_DEFAULTS=(--env REALISM=1 --env SKIP_FIREBASE_TOKEN=1 --env FIREBASE_REFRESH_FRACTION=0)

case "${1:-}" in
provision) shift; cmd_provision "$@" ;;
env) shift; cmd_env ;;
deploy) shift; cmd_deploy ;;
seed) shift; cmd_seed ;;
config) shift; remote config ;;
smoke)
    shift
    rc=0
    remote smoke "${K6_DEFAULTS[@]}" "$@" || rc=$?
    check_settlement
    exit $rc
    ;;
run)
    shift
    label=${1:?usage: loadtest-lt.sh run <label> [k6 args...]}
    shift
    rc=0
    remote run "$label" "${K6_DEFAULTS[@]}" "$@" || rc=$?
    check_settlement
    exit $rc
    ;;
*)
    sed -n '2,19p' "$0" | sed 's/^# \{0,1\}//'
    exit 1
    ;;
esac
