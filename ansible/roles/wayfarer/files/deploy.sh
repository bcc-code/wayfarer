#!/bin/bash
# Proxyless blue/green deploy for the wayfarer server.
#
#   deploy.sh <sha>     activate /opt/wayfarer/releases/<sha>
#   deploy.sh rollback  re-activate the previously active release
#
# Both colors bind the same public port via SO_REUSEPORT; the kernel splits
# new connections during the overlap. Flow: start the inactive color on the
# new release -> poll its private /health -> mark active -> gracefully stop
# the old color. On a failed health check the new color is stopped and the
# old one never stops serving.
set -euo pipefail

BASE=/opt/wayfarer
ACTIVE_FILE=$BASE/active
HEALTH_TIMEOUT=${HEALTH_TIMEOUT:-60}

admin_port() { [ "$1" = blue ] && echo 9441 || echo 9442; }
other() { [ "$1" = blue ] && echo green || echo blue; }

active=$(cat "$ACTIVE_FILE" 2>/dev/null || echo none)
if [ "$active" = none ]; then target=blue; else target=$(other "$active"); fi

if [ "${1:-}" = rollback ]; then
    prev=$(readlink "$BASE/color/$target" 2>/dev/null || true)
    [ -n "$prev" ] || { echo "nothing to roll back to"; exit 1; }
    sha=$(basename "$prev")
    echo "rolling back to $sha"
else
    sha=${1:?usage: deploy.sh <sha>|rollback}
    [ -x "$BASE/releases/$sha/wayfarer-server" ] || { echo "release $sha not found"; exit 1; }
fi

echo "deploy: $sha -> $target (active: $active)"
ln -sfn "$BASE/releases/$sha" "$BASE/color/$target"
systemctl start "wayfarer@$target"

port=$(admin_port "$target")
echo "health gate: 127.0.0.1:$port (${HEALTH_TIMEOUT}s)"
for i in $(seq 1 "$HEALTH_TIMEOUT"); do
    if curl -sf -m 2 "http://127.0.0.1:$port/health" > /dev/null; then
        echo "$target healthy after ${i}s"
        echo "$target" > "$ACTIVE_FILE"
        systemctl enable -q "wayfarer@$target"
        if [ "$active" != none ]; then
            echo "retiring $active (graceful drain)"
            systemctl disable -q "wayfarer@$active" || true
            systemctl stop "wayfarer@$active"
        fi
        # Keep the last few releases for rollback; prune the rest.
        ls -1dt "$BASE"/releases/*/ 2>/dev/null | tail -n +6 | xargs -r rm -rf
        echo "deploy complete: $target @ $sha"
        exit 0
    fi
    sleep 1
done

echo "HEALTH GATE FAILED — stopping $target, $active keeps serving" >&2
journalctl -u "wayfarer@$target" -n 20 --no-pager >&2 || true
systemctl stop "wayfarer@$target" || true
exit 1
