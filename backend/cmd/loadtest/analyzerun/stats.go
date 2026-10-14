package main

import (
	"math"
	"sort"
	"strconv"
)

// Stats are the percentiles of a set of samples; empty sets only have Count
type Stats struct {
	Count int      `json:"count"`
	P50   *float64 `json:"p50,omitempty"`
	P95   *float64 `json:"p95,omitempty"`
	P99   *float64 `json:"p99,omitempty"`
	Max   *float64 `json:"max,omitempty"`
}

// percentile picks the nearest-rank sample of sorted values (index rounded
// half to even)
func percentile(sorted []float64, p float64) float64 {
	i := int(math.RoundToEven(float64(len(sorted)-1) * p))
	return sorted[min(len(sorted)-1, i)]
}

// round to digits decimals, half to even
func round(v float64, digits int) float64 {
	f := math.Pow(10, float64(digits))
	return math.RoundToEven(v*f) / f
}

func ptr(v float64) *float64 { return &v }

// computeStats sorts vals in place
func computeStats(vals []float64) Stats {
	if len(vals) == 0 {
		return Stats{}
	}
	sort.Float64s(vals)
	return Stats{
		Count: len(vals),
		P50:   ptr(round(percentile(vals, 0.50), 2)),
		P95:   ptr(round(percentile(vals, 0.95), 2)),
		P99:   ptr(round(percentile(vals, 0.99), 2)),
		Max:   ptr(round(vals[len(vals)-1], 2)),
	}
}

// fmtFloat formats a number for CSV and text output, "" for nil
func fmtFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}
