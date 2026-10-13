// Command analyzerun summarizes a load-test run directory written by
// scripts/loadtest-remote.sh.
//
//	go run ./cmd/loadtest/analyzerun <run-dir>
//
// It reads the raw data (k6 CSV per shard, pidstat/mpstat samples, SQL
// exports) and writes derived, graph-ready files next to it:
//
//	requests_per_second.csv    per second x request name: count, failed,
//	                           p50/p95/p99/max http_req_duration (ms)
//	tls_per_second.csv         per second: new TLS handshakes, p50/p95 (ms)
//	cpu_per_second.csv         per second: server process / postgres / server
//	                           box busy %, server RSS, and k6 + box busy % per
//	                           generator
//	settlement_per_second.csv  per second: bets settled (from the score_journal
//	                           export's created_at), cumulative
//	summary.json               everything printed, machine-readable
//
// The raw files stay untouched; these are conveniences for graphing.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: analyzerun <run-dir>")
		os.Exit(1)
	}
	if err := analyze(os.Args[1], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "analyzerun:", err)
		os.Exit(1)
	}
}
