#!/bin/bash
# Orchestrates the full A/B differential run ON the test box.
#
# Assumes:
#   - systemd units wayfarer-a (:8080, DB $DB_A) and wayfarer-b (:8081, DB $DB_B)
#   - both DBs are fresh clones of the same seeded template (see notes)
#   - diffcmp binary and oracle.sh sit next to this script
#
# The write scenarios are NOT idempotent (enrollments, quiz submissions):
# run this exactly once per fresh DB pair.
set -euo pipefail

cd "$(dirname "$0")"

BASE_A="${BASE_A:-http://127.0.0.1:8080}"
BASE_B="${BASE_B:-http://127.0.0.1:8081}"
DB_A="${DB_A:-wayfarer_diff_a}"
DB_B="${DB_B:-wayfarer_diff_b}"
OUT="${OUT:-results-$(date +%Y%m%d-%H%M%S)}"
mkdir -p "$OUT"

FAILURES=0
note() { echo; echo "########## $1 ##########"; }

restart() {
    systemctl restart wayfarer-a wayfarer-b
    for url in "$BASE_A" "$BASE_B"; do
        for _ in $(seq 1 30); do
            curl -sf "$url/health" >/dev/null 2>&1 && break
            sleep 1
        done
    done
    sleep 5 # let the branch's Firebase token warmer boot pass settle
}

run_diffcmp() { # <label> <filter>
    local label="$1" filter="$2"
    note "diffcmp $label (filter: $filter)"
    if ./diffcmp -a "$BASE_A" -b "$BASE_B" -run "$filter" -out "$OUT/$label" \
        | tee "$OUT/$label.log"; then
        echo "[$label] OK"
    else
        echo "[$label] HAD UNEXPECTED RESULTS"
        FAILURES=$((FAILURES + 1))
    fi
}

run_oracle() { # <label>
    note "oracle $1"
    if ./oracle.sh "$DB_A" "$DB_B" "$1" | tee "$OUT/oracle-$1.log"; then
        echo "[oracle-$1] OK"
    else
        echo "[oracle-$1] DIVERGED"
        FAILURES=$((FAILURES + 1))
    fi
}

note "0. sanity: A vs A on the read battery (must be 100% MATCH)"
restart
if ./diffcmp -a "$BASE_A" -b "$BASE_A" -run '^R-reads-cold$' -out "$OUT/sanity-a-vs-a" \
    | tee "$OUT/sanity-a-vs-a.log"; then
    echo "[sanity] OK — harness normalization is sound"
else
    echo "[sanity] FAILED — normalization bug in the harness; aborting"
    exit 2
fi

run_oracle "pre-run"

note "1. read battery"
restart
run_diffcmp "reads" '^R-'

note "2. write battery"
restart
run_diffcmp "writes" '^W'
run_oracle "post-writes"

note "3. pinned probes (fresh caches per probe)"
for probe in P1-cross-user-leak P2-variable-collision P3-stale-after-enroll P4-ttl-staleness P5-metrics-endpoint; do
    restart
    run_diffcmp "$probe" "^${probe}\$"
done

note "4. read battery again, post-writes (write-induced read drift)"
restart
run_diffcmp "reads-postwrite" '^R-reads-cold$'
run_oracle "final"

note "DONE"
tar czf "$OUT.tar.gz" "$OUT"
echo "results: $OUT.tar.gz — $FAILURES phase(s) with unexpected results"
echo "(pinned-probe EXPECTED-DIVERGE verdicts are successes; check summary.json per phase)"
exit $((FAILURES > 0 ? 1 : 0))
