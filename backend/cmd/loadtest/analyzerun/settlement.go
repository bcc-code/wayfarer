package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Settlement is the bet settlement throughput of a run
type Settlement struct {
	Bets          int     `json:"bets"`
	First         string  `json:"first"`
	Last          string  `json:"last"`
	Seconds       float64 `json:"seconds"`
	BetsPerSecond float64 `json:"bets_per_second"`
	PeakPerSecond int     `json:"peak_per_second"`
}

// settledPerSecond counts settled bets per second of the exported BET journal
// rows (server/bet_journal.csv). Every bet has two rows, a stake (< 0) and a
// winnings entry (>= 0); the winnings row counts the bet once.
func settledPerSecond(r io.Reader) (map[string]int, error) {
	cr := csv.NewReader(r)
	header, err := cr.Read()
	if errors.Is(err, io.EOF) {
		return map[string]int{}, nil
	}
	if err != nil {
		return nil, err
	}
	kindCol, createdCol := -1, -1
	for i, h := range header {
		switch h {
		case "kind":
			kindCol = i
		case "created_at":
			createdCol = i
		}
	}
	if kindCol < 0 || createdCol < 0 {
		return nil, fmt.Errorf("bet journal export needs kind and created_at columns")
	}
	perSec := map[string]int{}
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return perSec, nil
		}
		if err != nil {
			return nil, err
		}
		if rec[kindCol] != "winnings" {
			continue
		}
		// "2026-10-13 10:02:01.067006+00" -> the second, UTC
		t := rec[createdCol]
		if len(t) > 19 {
			t = t[:19]
		}
		perSec[t]++
	}
}

func parseSecond(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time %q", s)
}

// writeSettlement writes settlement_per_second.csv; nil when the run has no
// bet journal export or no settled bets
func writeSettlement(runDir string) (*Settlement, error) {
	f, err := os.Open(filepath.Join(runDir, "server", "bet_journal.csv"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	perSec, err := settledPerSecond(f)
	if err != nil {
		return nil, fmt.Errorf("bet_journal.csv: %w", err)
	}
	if len(perSec) == 0 {
		return nil, nil
	}
	secs := make([]string, 0, len(perSec))
	peak := 0
	for t, n := range perSec {
		secs = append(secs, t)
		peak = max(peak, n)
	}
	sort.Strings(secs)

	out, err := os.Create(filepath.Join(runDir, "settlement_per_second.csv"))
	if err != nil {
		return nil, err
	}
	defer out.Close()
	w := csv.NewWriter(out)
	if err := w.Write([]string{"time_utc", "bets_settled", "cumulative"}); err != nil {
		return nil, err
	}
	total := 0
	for _, t := range secs {
		total += perSec[t]
		if err := w.Write([]string{t, strconv.Itoa(perSec[t]), strconv.Itoa(total)}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	if err := out.Close(); err != nil {
		return nil, err
	}

	first, last := secs[0], secs[len(secs)-1]
	ft, err := parseSecond(first)
	if err != nil {
		return nil, err
	}
	lt, err := parseSecond(last)
	if err != nil {
		return nil, err
	}
	span := lt.Sub(ft).Seconds() + 1
	return &Settlement{
		Bets:          total,
		First:         first,
		Last:          last,
		Seconds:       span,
		BetsPerSecond: round(float64(total)/span, 1),
		PeakPerSecond: peak,
	}, nil
}
