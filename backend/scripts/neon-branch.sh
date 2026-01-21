#!/bin/bash
set -e

# Check if neon CLI is installed
if ! command -v neon &> /dev/null; then
    echo "Error: Neon CLI is not installed."
    echo ""
    echo "Install with: npm install -g neonctl"
    echo "Or: brew install neonctl"
    echo ""
    echo "Then authenticate with: neon auth"
    exit 1
fi

# Load .env file if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | grep -v '^$' | xargs)
fi

# Check for required config
if [ -z "$NEON_PROJECT_ID" ]; then
    echo "Error: NEON_PROJECT_ID not set in .env"
    exit 1
fi

if [ -z "$DATABASE_NAME" ]; then
    echo "Error: DATABASE_NAME not set in .env"
    exit 1
fi

# Get current git branch
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# If on main branch, use the default Neon branch
if [ "$GIT_BRANCH" = "main" ] || [ "$GIT_BRANCH" = "master" ]; then
    echo "On main branch - using default Neon branch" >&2
    CONNECTION_STRING=$(neon connection-string --project-id "$NEON_PROJECT_ID" --database-name "$DATABASE_NAME")
    echo "$CONNECTION_STRING"
    exit 0
fi

# Sanitize branch name for Neon (replace / with -)
NEON_BRANCH_NAME=$(echo "$GIT_BRANCH" | tr '/' '-')

# Check if branch exists
BRANCH_EXISTS=$(neon branches list --project-id "$NEON_PROJECT_ID" --output json | jq -r ".[] | select(.name == \"$NEON_BRANCH_NAME\") | .name")

if [ -z "$BRANCH_EXISTS" ]; then
    echo "Creating Neon branch: $NEON_BRANCH_NAME (expires in 7 days)" >&2
    EXPIRES_AT=$(date -v+7d -u +%Y-%m-%dT%H:%M:%SZ)
    neon branches create --project-id "$NEON_PROJECT_ID" --name "$NEON_BRANCH_NAME" --expires-at "$EXPIRES_AT" --cu 0.25-4 >&2
fi

# Get connection string for the branch
CONNECTION_STRING=$(neon connection-string --project-id "$NEON_PROJECT_ID" --branch "$NEON_BRANCH_NAME" --database-name "$DATABASE_NAME")

echo "$CONNECTION_STRING"
