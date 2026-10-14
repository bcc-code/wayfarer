package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// cpuSample is one pidstat sample: %CPU and RSS (kB)
type cpuSample struct {
	cpu   float64
	rssKB float64
}

// isSampleTime tells a sample's time column ("10:00:49") from the summary
// rows ("Average:") sysstat prints when stopped cleanly
func isSampleTime(s string) bool {
	return strings.Contains(s, ":") && !strings.HasSuffix(s, ":")
}

// parsePidstat reads `pidstat -h` output: a "# Time UID PID ... %CPU ...
// Command" header, then one row per process and sample. Samples are keyed by
// their time column; with sum, processes sampled at the same time are added up
// (postgres backends, k6 workers).
func parsePidstat(r io.Reader, sum bool) (map[string]cpuSample, error) {
	out := map[string]cpuSample{}
	var header []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) == 0 {
			continue
		}
		if parts[0] == "#" {
			header = parts[1:]
			continue
		}
		if header == nil || len(parts) != len(header) || !isSampleTime(parts[0]) {
			continue
		}
		row := map[string]string{}
		for i, h := range header {
			row[h] = parts[i]
		}
		cpu, err := strconv.ParseFloat(row["%CPU"], 64)
		if err != nil {
			continue
		}
		rss, _ := strconv.ParseFloat(row["RSS"], 64)
		t := row["Time"]
		if prev, ok := out[t]; ok && sum {
			out[t] = cpuSample{cpu: prev.cpu + cpu, rssKB: prev.rssKB + rss}
		} else {
			out[t] = cpuSample{cpu: cpu, rssKB: rss}
		}
	}
	return out, sc.Err()
}

// parseMpstat reads `mpstat 1` output: rows "<time> all %usr ... %idle",
// returning the box's busy % (100 - idle) per time
func parseMpstat(r io.Reader) (map[string]float64, error) {
	out := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) > 3 && parts[1] == "all" && isSampleTime(parts[0]) {
			idle, err := strconv.ParseFloat(parts[len(parts)-1], 64)
			if err == nil {
				out[parts[0]] = round(100-idle, 2)
			}
		}
	}
	return out, sc.Err()
}

// readOptional opens path and hands it to parse; a missing file is no data
func readOptional(path string, parse func(io.Reader) error) error {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if err := parse(f); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// writeCPU writes cpu_per_second.csv and returns the peak of every column
func writeCPU(runDir string) (map[string]float64, error) {
	server := map[string]cpuSample{}
	postgres := map[string]cpuSample{}
	serverBox := map[string]float64{}
	if err := readOptional(filepath.Join(runDir, "server", "server.pidstat"),
		func(r io.Reader) (err error) { server, err = parsePidstat(r, false); return err }); err != nil {
		return nil, err
	}
	if err := readOptional(filepath.Join(runDir, "server", "postgres.pidstat"),
		func(r io.Reader) (err error) { postgres, err = parsePidstat(r, true); return err }); err != nil {
		return nil, err
	}
	if err := readOptional(filepath.Join(runDir, "server", "server-box.mpstat"),
		func(r io.Reader) (err error) { serverBox, err = parseMpstat(r); return err }); err != nil {
		return nil, err
	}

	// Per generator: k6 process CPU and box busy %
	gens := map[string]map[string]float64{}
	genDirs, err := filepath.Glob(filepath.Join(runDir, "generator*"))
	if err != nil {
		return nil, err
	}
	sort.Strings(genDirs)
	for _, dir := range genDirs {
		g := filepath.Base(dir)
		pidstats, _ := filepath.Glob(filepath.Join(dir, "shard*.k6.pidstat"))
		for _, p := range pidstats {
			if err := readOptional(p, func(r io.Reader) error {
				samples, err := parsePidstat(r, true)
				col := map[string]float64{}
				for t, s := range samples {
					col[t] = s.cpu
				}
				gens[g+"_k6_cpu_pct"] = col
				return err
			}); err != nil {
				return nil, err
			}
		}
		mpstats, _ := filepath.Glob(filepath.Join(dir, "shard*.box.mpstat"))
		for _, p := range mpstats {
			if err := readOptional(p, func(r io.Reader) (err error) {
				gens[g+"_box_busy_pct"], err = parseMpstat(r)
				return err
			}); err != nil {
				return nil, err
			}
		}
	}

	timeSet := map[string]bool{}
	for t := range server {
		timeSet[t] = true
	}
	for t := range postgres {
		timeSet[t] = true
	}
	for t := range serverBox {
		timeSet[t] = true
	}
	genCols := make([]string, 0, len(gens))
	for g, data := range gens {
		genCols = append(genCols, g)
		for t := range data {
			timeSet[t] = true
		}
	}
	sort.Strings(genCols)
	times := make([]string, 0, len(timeSet))
	for t := range timeSet {
		times = append(times, t)
	}
	sort.Strings(times)

	cols := append([]string{"server_cpu_pct", "server_rss_mb", "postgres_cpu_pct", "server_box_busy_pct"}, genCols...)
	peaks := map[string]float64{}
	f, err := os.Create(filepath.Join(runDir, "cpu_per_second.csv"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write(append([]string{"time"}, cols...)); err != nil {
		return nil, err
	}
	for _, t := range times {
		row := map[string]*float64{}
		if s, ok := server[t]; ok {
			row["server_cpu_pct"] = ptr(s.cpu)
			row["server_rss_mb"] = ptr(round(s.rssKB/1024, 1))
		}
		if s, ok := postgres[t]; ok {
			row["postgres_cpu_pct"] = ptr(round(s.cpu, 2))
		}
		if v, ok := serverBox[t]; ok {
			row["server_box_busy_pct"] = ptr(v)
		}
		for _, g := range genCols {
			if v, ok := gens[g][t]; ok {
				row[g] = ptr(v)
			}
		}
		rec := []string{t}
		for _, c := range cols {
			if v := row[c]; v != nil {
				peaks[c] = max(peaks[c], *v)
			}
			rec = append(rec, fmtFloat(row[c]))
		}
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return peaks, f.Close()
}
