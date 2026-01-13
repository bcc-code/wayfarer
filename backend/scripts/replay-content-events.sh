#!/bin/bash
#
# Replay external content events from a CSV export
#
# Usage: ./scripts/replay-content-events.sh -k <api-key> -f <csv-file> [-u <base-url>] [-d]
#

set -euo pipefail

# Defaults
BASE_URL="http://localhost:8080"
DRY_RUN=false
API_KEY=""
CSV_FILE=""

usage() {
    cat <<EOF
Usage: $(basename "$0") -k <api-key> -f <csv-file> [options]

Replay external content events from a CSV database export.

Required:
  -k, --api-key <key>    API key for authentication
  -f, --file <path>      Path to CSV file

Options:
  -u, --url <url>        Base URL (default: http://localhost:8080)
  -d, --dry-run          Print requests without executing
  -h, --help             Show this help message

CSV Format (database export):
  id,person_id,task_id,plan_id,source,received_at,content_progress,consumed_at

Timestamp formats supported:
  - RFC3339: 2025-12-23T00:33:49Z
  - Postgres: 2025-12-23 00:33:49.000000 +00:00

EOF
    exit 1
}

# Convert postgres timestamp to RFC3339
# Input:  2025-12-23 00:33:49.000000 +00:00
# Output: 2025-12-23T00:33:49+00:00
to_rfc3339() {
    local ts="$1"
    # Already RFC3339 format (contains T)
    if [[ "$ts" == *"T"* ]]; then
        echo "$ts"
        return
    fi
    # Convert postgres format to RFC3339
    # 1. Replace first space with T
    # 2. Remove microseconds (.000000)
    # 3. Remove space before timezone
    echo "$ts" | sed -E 's/ /T/; s/\.[0-9]+//; s/ ([+-])/\1/'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--api-key)
            API_KEY="$2"
            shift 2
            ;;
        -f|--file)
            CSV_FILE="$2"
            shift 2
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Error: Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required arguments
if [[ -z "$API_KEY" ]]; then
    echo "Error: API key is required (-k, --api-key)"
    exit 1
fi

if [[ -z "$CSV_FILE" ]]; then
    echo "Error: CSV file is required (-f, --file)"
    exit 1
fi

if [[ ! -f "$CSV_FILE" ]]; then
    echo "Error: CSV file not found: $CSV_FILE"
    exit 1
fi

ENDPOINT="${BASE_URL}/api/v1/content-events"

# Counters (using temp files to persist across subshell)
RESULT_FILE=$(mktemp)
trap 'rm -f "$RESULT_FILE"' EXIT
echo "0 0" > "$RESULT_FILE"

echo "Replaying events from: $CSV_FILE"
echo "Target endpoint: $ENDPOINT"
if $DRY_RUN; then
    echo "Mode: DRY RUN (no requests will be made)"
fi
echo ""

# Count total lines (minus header)
total_lines=$(($(wc -l < "$CSV_FILE") - 1))
echo "Total events to process: $total_lines"
echo ""

# Process CSV (skip header)
current=0
tail -n +2 "$CSV_FILE" | while IFS=',' read -r id person_id task_id plan_id source received_at content_progress consumed_at; do
    ((current++)) || true

    # Handle nullable fields
    if [[ -z "$plan_id" || "$plan_id" == "NULL" || "$plan_id" == '""' ]]; then
        plan_id_json="null"
    else
        # Remove surrounding quotes if present
        plan_id=$(echo "$plan_id" | sed 's/^"//;s/"$//')
        plan_id_json="\"$plan_id\""
    fi

    if [[ -z "$content_progress" || "$content_progress" == "NULL" ]]; then
        content_progress_json="null"
    else
        content_progress_json="$content_progress"
    fi

    # Remove surrounding quotes from fields if present
    person_id=$(echo "$person_id" | sed 's/^"//;s/"$//')
    task_id=$(echo "$task_id" | sed 's/^"//;s/"$//')
    consumed_at=$(echo "$consumed_at" | sed 's/^"//;s/"$//')

    # Convert timestamp to RFC3339 format
    consumed_at=$(to_rfc3339 "$consumed_at")

    # Build JSON payload
    payload=$(cat <<EOF
{
  "person_id": "$person_id",
  "task_id": "$task_id",
  "plan_id": $plan_id_json,
  "timestamp": "$consumed_at",
  "content_progress": $content_progress_json
}
EOF
)

    echo -n "[$current/$total_lines] Processing event for person $person_id, task $task_id... "

    if $DRY_RUN; then
        echo "DRY RUN"
        echo "  Payload: $(echo "$payload" | tr -d '\n' | tr -s ' ')"
    else
        # Make the request
        http_code=$(curl -s -o /dev/null -w "%{http_code}" \
            -X POST "$ENDPOINT" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $API_KEY" \
            -d "$payload")

        # Read current counts
        read succeeded failed < "$RESULT_FILE"

        if [[ "$http_code" == "201" ]]; then
            echo "OK (201)"
            echo "$((succeeded + 1)) $failed" > "$RESULT_FILE"
        else
            echo "FAILED (HTTP $http_code)"
            echo "$succeeded $((failed + 1))" > "$RESULT_FILE"
        fi
    fi
done

# Read final counts
read succeeded failed < "$RESULT_FILE"

echo ""
echo "========================================="
echo "Summary:"
echo "  Total processed: $total_lines"
if ! $DRY_RUN; then
    echo "  Succeeded: $succeeded"
    echo "  Failed: $failed"
fi
echo "========================================="
