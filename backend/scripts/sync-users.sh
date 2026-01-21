#!/bin/bash
#
# Sync users from Members API
#
# Usage: ./scripts/sync-users.sh -k <api-key> -f <user-ids-file> [-u <base-url>] [-d]
#

set -euo pipefail

# Defaults
BASE_URL="http://localhost:8080"
DRY_RUN=false
PROCESS_PENDING=false
API_KEY=""
INPUT_FILE=""

usage() {
    cat <<EOF
Usage: $(basename "$0") -k <api-key> -f <user-ids-file> [options]

Sync users from Members API by calling the maintenance sync endpoint.

Required:
  -k, --api-key <key>    API key for authentication (from EXTERNAL_API_KEYS)
  -f, --file <path>      Path to file with user IDs (one per line)

Options:
  -u, --url <url>        Base URL (default: http://localhost:8080)
  -p, --process-pending  Process pending consent and content events
  -d, --dry-run          Print requests without executing
  -h, --help             Show this help message

Getting user IDs for unknown church:
  psql -c "SELECT u.id FROM users u JOIN churches c ON u.church_id = c.id WHERE c.external_id IS NULL ORDER BY u.id;" -t -A > unknown-church-users.txt

EOF
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--api-key)
            API_KEY="$2"
            shift 2
            ;;
        -f|--file)
            INPUT_FILE="$2"
            shift 2
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -p|--process-pending)
            PROCESS_PENDING=true
            shift
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

if [[ -z "$INPUT_FILE" ]]; then
    echo "Error: Input file is required (-f, --file)"
    exit 1
fi

if [[ ! -f "$INPUT_FILE" ]]; then
    echo "Error: Input file not found: $INPUT_FILE"
    exit 1
fi

# Counters (using temp files to persist across subshell)
RESULT_FILE=$(mktemp)
trap 'rm -f "$RESULT_FILE"' EXIT
echo "0 0" > "$RESULT_FILE"

# Build query string
QUERY_STRING=""
if $PROCESS_PENDING; then
    QUERY_STRING="?process_pending=true"
fi

echo "Syncing users from: $INPUT_FILE"
echo "Target endpoint: ${BASE_URL}/api/maintenance/sync-user/:user_id${QUERY_STRING}"
if $DRY_RUN; then
    echo "Mode: DRY RUN (no requests will be made)"
fi
if $PROCESS_PENDING; then
    echo "Mode: Processing pending consent and content events"
fi
echo ""

# Count total lines (excluding empty lines)
total_lines=$(grep -c -v '^[[:space:]]*$' "$INPUT_FILE" || echo 0)
echo "Total users to process: $total_lines"
echo ""

# Process file (skip empty lines)
current=0
while IFS= read -r user_id || [[ -n "$user_id" ]]; do
    # Skip empty lines
    if [[ -z "$user_id" || "$user_id" =~ ^[[:space:]]*$ ]]; then
        continue
    fi

    # Trim whitespace
    user_id=$(echo "$user_id" | tr -d '[:space:]')

    ((current++)) || true

    echo -n "[$current/$total_lines] Syncing user $user_id... "

    if $DRY_RUN; then
        echo "DRY RUN"
        echo "  POST ${BASE_URL}/api/maintenance/sync-user/${user_id}${QUERY_STRING}"
    else
        # Make the request
        http_code=$(curl -s -o /dev/null -w "%{http_code}" \
            -X POST "${BASE_URL}/api/maintenance/sync-user/${user_id}${QUERY_STRING}" \
            -H "Authorization: Bearer $API_KEY")

        # Read current counts
        read succeeded failed < "$RESULT_FILE"

        if [[ "$http_code" == "200" || "$http_code" == "204" ]]; then
            echo "OK ($http_code)"
            echo "$((succeeded + 1)) $failed" > "$RESULT_FILE"
        else
            echo "FAILED (HTTP $http_code)"
            echo "$succeeded $((failed + 1))" > "$RESULT_FILE"
        fi
    fi
done < "$INPUT_FILE"

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
