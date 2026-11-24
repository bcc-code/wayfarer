package main

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// Custom metrics of the k6 scripts that make it into the summary
var (
	customCounters = map[string]bool{
		"bets_accepted": true, "bets_none": true, "bets_double_tap": true,
		"bets_rejected_expected": true, "bets_wrong_outcome": true,
		"quiz_completions": true, "quiz_failures": true, "quiz_skipped": true,
		"dropped_iterations": true, "iterations": true, "graphql_errors": true,
		"sessions_created": true, "sessions_finished": true, "admin_failures": true,
		"session_wait_timeouts": true, "refetches": true, "refetch_cancelled": true,
		"refetch_errors": true,
	}
	customTrends = map[string]bool{
		"settlement_duration": true, "bet_amount": true, "iteration_duration": true,
		"finish_duration": true, "time_to_session_visible": true, "refetch_completed_latency": true,
		"refetch_send_lag": true,
	}
)

// k6Data is everything aggregated from the k6 CSV outputs of a run
type k6Data struct {
	files        []string
	durations    map[string][]float64           // request name -> ms
	failed       map[string]int                 // request name -> failed requests
	perSec       map[int64]map[string][]float64 // second -> request name -> ms
	perSecFailed map[int64]map[string]int       // second -> request name -> failed
	tls          map[int64][]float64            // second -> new handshakes (ms)
	tlsAll       []float64
	counters     map[string]float64
	trends       map[string][]float64
}

// k6Files are the CSVs of every generator shard (plus settle) and of a local run
func k6Files(runDir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(runDir, "generator*", "*.csv.gz"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	local := filepath.Join(runDir, "local.csv.gz")
	if _, err := os.Stat(local); err == nil {
		files = append(files, local)
	}
	return files, nil
}

// readK6 streams every k6 CSV of a run once, aggregating what the outputs need
func readK6(runDir string) (*k6Data, error) {
	k := &k6Data{
		durations:    map[string][]float64{},
		failed:       map[string]int{},
		perSec:       map[int64]map[string][]float64{},
		perSecFailed: map[int64]map[string]int{},
		tls:          map[int64][]float64{},
		counters:     map[string]float64{},
		trends:       map[string][]float64{},
	}
	files, err := k6Files(runDir)
	if err != nil {
		return nil, err
	}
	for _, path := range files {
		if err := k.readFile(path); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		rel, err := filepath.Rel(runDir, path)
		if err != nil {
			rel = path
		}
		k.files = append(k.files, rel)
	}
	return k, nil
}

func (k *k6Data) readFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	return k.read(gz)
}

// read aggregates one k6 CSV (--out csv format, with a header row)
func (k *k6Data) read(r io.Reader) error {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.ReuseRecord = true
	header, err := cr.Read()
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	for _, c := range []string{"metric_name", "timestamp", "metric_value"} {
		if _, ok := col[c]; !ok {
			return fmt.Errorf("no %s column", c)
		}
	}
	field := func(rec []string, name string) string {
		if i, ok := col[name]; ok && i < len(rec) {
			return rec[i]
		}
		return ""
	}

	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		metric := field(rec, "metric_name")
		value, err := strconv.ParseFloat(field(rec, "metric_value"), 64)
		if err != nil {
			return fmt.Errorf("metric_value: %w", err)
		}
		// Requests are named by their name tag, falling back to the URL
		name := field(rec, "name")
		if name == "" {
			name = field(rec, "url")
		}

		switch {
		case metric == "http_req_duration":
			ts, err := timestamp(field(rec, "timestamp"))
			if err != nil {
				return err
			}
			k.durations[name] = append(k.durations[name], value)
			if k.perSec[ts] == nil {
				k.perSec[ts] = map[string][]float64{}
			}
			k.perSec[ts][name] = append(k.perSec[ts][name], value)
		case metric == "http_req_failed":
			if value <= 0 {
				continue
			}
			ts, err := timestamp(field(rec, "timestamp"))
			if err != nil {
				return err
			}
			k.failed[name]++
			if k.perSecFailed[ts] == nil {
				k.perSecFailed[ts] = map[string]int{}
			}
			k.perSecFailed[ts][name]++
		case metric == "http_req_tls_handshaking":
			// Only requests that opened a connection handshake; the rest are 0
			if value <= 0 {
				continue
			}
			ts, err := timestamp(field(rec, "timestamp"))
			if err != nil {
				return err
			}
			k.tls[ts] = append(k.tls[ts], value)
			k.tlsAll = append(k.tlsAll, value)
		case customCounters[metric]:
			k.counters[metric] += value
		case customTrends[metric]:
			k.trends[metric] = append(k.trends[metric], value)
		}
	}
}

func timestamp(s string) (int64, error) {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("timestamp: %w", err)
	}
	return ts, nil
}
