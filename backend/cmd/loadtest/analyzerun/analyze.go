package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RequestStats are request durations (ms) plus the failed request count
type RequestStats struct {
	Stats
	Failed int `json:"failed"`
}

// TLSStats are new TLS handshake durations (ms) plus the busiest second
type TLSStats struct {
	Stats
	PeakPerSecond int `json:"peak_per_second"`
}

// Summary is summary.json
type Summary struct {
	K6Files        []string                `json:"k6_files"`
	Requests       RequestStats            `json:"requests"`
	PerNameMs      map[string]RequestStats `json:"per_name_ms"`
	TLSHandshakeMs TLSStats                `json:"tls_handshake_ms"`
	Counters       map[string]int64        `json:"counters"`
	Trends         map[string]Stats        `json:"trends"`
	CPUPeaksPct    map[string]float64      `json:"cpu_peaks_pct"`
	Settlement     *Settlement             `json:"settlement"`
}

// analyze writes the derived files into runDir and prints the summary to out
func analyze(runDir string, out io.Writer) error {
	k, err := readK6(runDir)
	if err != nil {
		return err
	}
	if err := writeRequestsPerSecond(runDir, k); err != nil {
		return fmt.Errorf("requests_per_second.csv: %w", err)
	}
	if err := writeTLSPerSecond(runDir, k); err != nil {
		return fmt.Errorf("tls_per_second.csv: %w", err)
	}
	cpuPeaks, err := writeCPU(runDir)
	if err != nil {
		return fmt.Errorf("cpu_per_second.csv: %w", err)
	}
	settlement, err := writeSettlement(runDir)
	if err != nil {
		return fmt.Errorf("settlement_per_second.csv: %w", err)
	}

	s := buildSummary(k, cpuPeaks, settlement)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(runDir, "summary.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	printSummary(out, s)
	return nil
}

func buildSummary(k *k6Data, cpuPeaks map[string]float64, settlement *Settlement) Summary {
	s := Summary{
		K6Files:     k.files,
		PerNameMs:   map[string]RequestStats{},
		Counters:    map[string]int64{},
		Trends:      map[string]Stats{},
		CPUPeaksPct: cpuPeaks,
		Settlement:  settlement,
	}
	var all []float64
	failed := 0
	for name, vals := range k.durations {
		all = append(all, vals...)
		s.PerNameMs[name] = RequestStats{Stats: computeStats(vals), Failed: k.failed[name]}
	}
	for _, n := range k.failed {
		failed += n
	}
	s.Requests = RequestStats{Stats: computeStats(all), Failed: failed}

	peak := 0
	for _, v := range k.tls {
		peak = max(peak, len(v))
	}
	s.TLSHandshakeMs = TLSStats{Stats: computeStats(k.tlsAll), PeakPerSecond: peak}
	for name, v := range k.counters {
		s.Counters[name] = int64(v)
	}
	for name, v := range k.trends {
		s.Trends[name] = computeStats(v)
	}
	return s
}

func isoSecond(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("2006-01-02T15:04:05Z")
}

func sortedSeconds[V any](m map[int64]V) []int64 {
	secs := make([]int64, 0, len(m))
	for ts := range m {
		secs = append(secs, ts)
	}
	slices.Sort(secs)
	return secs
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// writeCSV creates path and writes the header and rows
func writeCSV(path string, header []string, rows func(w *csv.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return err
	}
	if err := rows(w); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return f.Close()
}

func writeRequestsPerSecond(runDir string, k *k6Data) error {
	return writeCSV(filepath.Join(runDir, "requests_per_second.csv"),
		[]string{"time_utc", "epoch", "name", "count", "failed", "p50_ms", "p95_ms", "p99_ms", "max_ms"},
		func(w *csv.Writer) error {
			for _, ts := range sortedSeconds(k.perSec) {
				for _, name := range sortedKeys(k.perSec[ts]) {
					s := computeStats(k.perSec[ts][name])
					if err := w.Write([]string{
						isoSecond(ts), strconv.FormatInt(ts, 10), name,
						strconv.Itoa(s.Count), strconv.Itoa(k.perSecFailed[ts][name]),
						fmtFloat(s.P50), fmtFloat(s.P95), fmtFloat(s.P99), fmtFloat(s.Max),
					}); err != nil {
						return err
					}
				}
			}
			return nil
		})
}

func writeTLSPerSecond(runDir string, k *k6Data) error {
	return writeCSV(filepath.Join(runDir, "tls_per_second.csv"),
		[]string{"time_utc", "epoch", "handshakes", "p50_ms", "p95_ms", "max_ms"},
		func(w *csv.Writer) error {
			for _, ts := range sortedSeconds(k.tls) {
				s := computeStats(k.tls[ts])
				if err := w.Write([]string{
					isoSecond(ts), strconv.FormatInt(ts, 10), strconv.Itoa(s.Count),
					fmtFloat(s.P50), fmtFloat(s.P95), fmtFloat(s.Max),
				}); err != nil {
					return err
				}
			}
			return nil
		})
}

func printSummary(out io.Writer, s Summary) {
	r := s.Requests
	fmt.Fprintf(out, "requests %d failed %d  p50 %sms p95 %sms p99 %sms\n",
		r.Count, r.Failed, fmtFloat(r.P50), fmtFloat(r.P95), fmtFloat(r.P99))
	fmt.Fprintf(out, "%-32s %7s %5s %9s %9s %9s %9s\n", "name", "count", "fail", "p50", "p95", "p99", "max")
	for _, name := range sortedKeys(s.PerNameMs) {
		n := s.PerNameMs[name]
		fmt.Fprintf(out, "%-32.32s %7d %5d %9s %9s %9s %9s\n", name, n.Count, n.Failed,
			fmtFloat(n.P50), fmtFloat(n.P95), fmtFloat(n.P99), fmtFloat(n.Max))
	}
	t := s.TLSHandshakeMs
	fmt.Fprintf(out, "TLS handshakes %d (peak %d/s) p50 %sms p95 %sms p99 %sms\n",
		t.Count, t.PeakPerSecond, fmtFloat(t.P50), fmtFloat(t.P95), fmtFloat(t.P99))
	counters := make([]string, 0, len(s.Counters))
	for _, name := range sortedKeys(s.Counters) {
		counters = append(counters, fmt.Sprintf("%s=%d", name, s.Counters[name]))
	}
	fmt.Fprintln(out, "counters:", strings.Join(counters, ", "))
	if d, ok := s.Trends["settlement_duration"]; ok {
		fmt.Fprintf(out, "settlement request: %s ms\n", fmtFloat(d.Max))
	}
	if st := s.Settlement; st != nil {
		fmt.Fprintf(out, "settlement: %d bets in %.0fs (%s/s, peak %d/s)\n",
			st.Bets, st.Seconds, fmtFloat(&st.BetsPerSecond), st.PeakPerSecond)
	}
	peaks := make([]string, 0, len(s.CPUPeaksPct))
	for _, name := range sortedKeys(s.CPUPeaksPct) {
		peaks = append(peaks, fmt.Sprintf("%s=%.0f", name, s.CPUPeaksPct[name]))
	}
	fmt.Fprintln(out, "CPU peaks %:", strings.Join(peaks, ", "))
}
