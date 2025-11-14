#!/bin/bash

# Profiling helper script for Wayfarer backend
# Usage: ./scripts/profile.sh [cpu|heap|allocs|goroutine|all] [duration] [server_url]

set -e

PROFILE_TYPE="${1:-all}"
DURATION="${2:-30}"
SERVER_URL="${3:-http://localhost:8080}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PROFILE_DIR="profiles/${TIMESTAMP}"

mkdir -p "${PROFILE_DIR}"

echo "Starting profiling session..."
echo "Profile type: ${PROFILE_TYPE}"
echo "Duration: ${DURATION}s"
echo "Server: ${SERVER_URL}"
echo "Output directory: ${PROFILE_DIR}"
echo ""

# Function to capture CPU profile
capture_cpu() {
    echo "Capturing CPU profile for ${DURATION}s..."
    curl -s "${SERVER_URL}/debug/pprof/profile?seconds=${DURATION}" -o "${PROFILE_DIR}/cpu.prof"
    echo "CPU profile saved to ${PROFILE_DIR}/cpu.prof"
}

# Function to capture heap profile
capture_heap() {
    echo "Capturing heap profile..."
    curl -s "${SERVER_URL}/debug/pprof/heap" -o "${PROFILE_DIR}/heap.prof"
    echo "Heap profile saved to ${PROFILE_DIR}/heap.prof"
}

# Function to capture allocs profile
capture_allocs() {
    echo "Capturing allocations profile..."
    curl -s "${SERVER_URL}/debug/pprof/allocs" -o "${PROFILE_DIR}/allocs.prof"
    echo "Allocations profile saved to ${PROFILE_DIR}/allocs.prof"
}

# Function to capture goroutine profile
capture_goroutine() {
    echo "Capturing goroutine profile..."
    curl -s "${SERVER_URL}/debug/pprof/goroutine" -o "${PROFILE_DIR}/goroutine.prof"
    echo "Goroutine profile saved to ${PROFILE_DIR}/goroutine.prof"
}

# Capture based on profile type
case "${PROFILE_TYPE}" in
    cpu)
        capture_cpu
        ;;
    heap)
        capture_heap
        ;;
    allocs)
        capture_allocs
        ;;
    goroutine)
        capture_goroutine
        ;;
    all)
        capture_cpu &
        CPU_PID=$!
        sleep 1  # Small delay to stagger requests
        capture_heap
        capture_allocs
        capture_goroutine
        wait $CPU_PID
        ;;
    *)
        echo "Unknown profile type: ${PROFILE_TYPE}"
        echo "Valid types: cpu, heap, allocs, goroutine, all"
        exit 1
        ;;
esac

echo ""
echo "Profiling complete!"
echo ""
echo "Analyze profiles with:"
echo "  go tool pprof ${PROFILE_DIR}/cpu.prof"
echo "  go tool pprof ${PROFILE_DIR}/heap.prof"
echo "  go tool pprof ${PROFILE_DIR}/allocs.prof"
echo "  go tool pprof ${PROFILE_DIR}/goroutine.prof"
echo ""
echo "Generate visual reports:"
echo "  go tool pprof -http=:8081 ${PROFILE_DIR}/cpu.prof"
echo "  go tool pprof -http=:8081 ${PROFILE_DIR}/heap.prof"
echo ""
echo "Generate text reports:"
echo "  go tool pprof -text ${PROFILE_DIR}/cpu.prof"
echo "  go tool pprof -text ${PROFILE_DIR}/heap.prof"
echo ""
echo "Generate flame graphs (SVG):"
echo "  go tool pprof -svg ${PROFILE_DIR}/cpu.prof > ${PROFILE_DIR}/cpu_flamegraph.svg"
echo "  go tool pprof -svg ${PROFILE_DIR}/heap.prof > ${PROFILE_DIR}/heap_flamegraph.svg"
