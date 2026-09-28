package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPStats replaces the per-request access log in production: it aggregates
// request counts and latency histograms per route in memory (a mutex-guarded
// map update per request instead of a synchronous journald write), and only
// emits log lines for requests worth looking at — 5xx responses and slow
// requests. Snapshot() serves the aggregates for post-factum analysis.
//
// Note: all GraphQL operations aggregate under "POST /graphql" here; use the
// OTel/Jaeger instrumentation for per-operation depth.
type HTTPStats struct {
	mu          sync.Mutex
	routes      map[string]*routeStats // cumulative since boot (serves /metrics/http)
	window      map[string]*routeStats // current dump window, drained by the dumper
	start       time.Time
	windowStart time.Time

	// SlowThreshold above which a request is logged as slow (default 2s).
	SlowThreshold time.Duration
}

// bucketBoundsMs are cumulative-histogram upper bounds in milliseconds.
var bucketBoundsMs = []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

type routeStats struct {
	count   uint64
	status  map[int]uint64 // by status class: 2 -> 2xx, 3 -> 3xx, ...
	sumMs   float64
	maxMs   float64
	buckets []uint64 // len(bucketBoundsMs)+1, last = overflow
}

func NewHTTPStats() *HTTPStats {
	now := time.Now()
	return &HTTPStats{
		routes:        make(map[string]*routeStats),
		window:        make(map[string]*routeStats),
		start:         now,
		windowStart:   now,
		SlowThreshold: 2 * time.Second,
	}
}

// Middleware records every request and logs only errors and slow requests.
func (s *HTTPStats) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		elapsed := time.Since(startedAt)

		route := c.FullPath()
		if route == "" {
			route = "(unmatched)"
		}
		status := c.Writer.Status()
		s.record(c.Request.Method+" "+route, status, float64(elapsed)/float64(time.Millisecond))

		switch {
		case status >= http.StatusInternalServerError:
			slog.Error("request failed",
				"method", c.Request.Method, "path", c.Request.URL.Path,
				"status", status, "duration", elapsed, "client", c.ClientIP(),
				"errors", c.Errors.String())
		case elapsed >= s.SlowThreshold:
			slog.Warn("slow request",
				"method", c.Request.Method, "path", c.Request.URL.Path,
				"status", status, "duration", elapsed)
		}
	}
}

func (s *HTTPStats) record(key string, status int, durMs float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	recordInto(s.routes, key, status, durMs)
	recordInto(s.window, key, status, durMs)
}

func recordInto(m map[string]*routeStats, key string, status int, durMs float64) {
	rs, ok := m[key]
	if !ok {
		rs = &routeStats{
			status:  make(map[int]uint64),
			buckets: make([]uint64, len(bucketBoundsMs)+1),
		}
		m[key] = rs
	}
	rs.count++
	rs.status[status/100]++
	rs.sumMs += durMs
	if durMs > rs.maxMs {
		rs.maxMs = durMs
	}
	idx := sort.SearchFloat64s(bucketBoundsMs, durMs)
	rs.buckets[idx]++
}

// RouteSnapshot is the externally served view of one route's aggregates.
// Percentiles are histogram upper-bound estimates (conservative: the true
// value is at or below the reported bucket bound).
type RouteSnapshot struct {
	Route    string            `json:"route"`
	Count    uint64            `json:"count"`
	ByStatus map[string]uint64 `json:"by_status"`
	AvgMs    float64           `json:"avg_ms"`
	MaxMs    float64           `json:"max_ms"`
	P50Ms    float64           `json:"p50_ms"`
	P90Ms    float64           `json:"p90_ms"`
	P95Ms    float64           `json:"p95_ms"`
	P99Ms    float64           `json:"p99_ms"`
}

type StatsSnapshot struct {
	Since  time.Time       `json:"since"`
	Routes []RouteSnapshot `json:"routes"`
}

func (s *HTTPStats) Snapshot() StatsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return snapshotOf(s.routes, s.start)
}

// DrainWindow returns the stats accumulated since the previous drain and
// starts a fresh window, so every dumped line covers exactly one interval —
// counts, averages and percentiles all describe that window alone.
func (s *HTTPStats) DrainWindow() StatsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := snapshotOf(s.window, s.windowStart)
	s.window = make(map[string]*routeStats)
	s.windowStart = time.Now()
	return snap
}

func snapshotOf(routes map[string]*routeStats, since time.Time) StatsSnapshot {
	out := StatsSnapshot{Since: since}
	for key, rs := range routes {
		snap := RouteSnapshot{
			Route:    key,
			Count:    rs.count,
			ByStatus: map[string]uint64{},
			MaxMs:    rs.maxMs,
		}
		for class, n := range rs.status {
			snap.ByStatus[statusClassLabel(class)] = n
		}
		if rs.count > 0 {
			snap.AvgMs = rs.sumMs / float64(rs.count)
		}
		snap.P50Ms = bucketQuantile(rs.buckets, rs.count, 0.50, rs.maxMs)
		snap.P90Ms = bucketQuantile(rs.buckets, rs.count, 0.90, rs.maxMs)
		snap.P95Ms = bucketQuantile(rs.buckets, rs.count, 0.95, rs.maxMs)
		snap.P99Ms = bucketQuantile(rs.buckets, rs.count, 0.99, rs.maxMs)
		out.Routes = append(out.Routes, snap)
	}
	sort.Slice(out.Routes, func(i, j int) bool { return out.Routes[i].Count > out.Routes[j].Count })
	return out
}

func statusClassLabel(class int) string {
	switch class {
	case 1:
		return "1xx"
	case 2:
		return "2xx"
	case 3:
		return "3xx"
	case 4:
		return "4xx"
	case 5:
		return "5xx"
	default:
		return "other"
	}
}

// dumpRecord is one JSONL line written by StartDumper.
type dumpRecord struct {
	At time.Time `json:"at"`
	StatsSnapshot
}

// StartDumper appends one JSON line per interval to path, until ctx is
// cancelled. Each line covers ONLY that interval (the window is drained and
// reset), so counts, averages and percentiles describe the window itself —
// no diffing needed. Ticks with no requests are skipped, so an idle server
// does not grow the file. Since/At on the line bound the window.
func (s *HTTPStats) StartDumper(ctx context.Context, path string, interval time.Duration) {
	if path == "" {
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				snap := s.DrainWindow()
				if len(snap.Routes) == 0 {
					continue
				}
				line, err := json.Marshal(dumpRecord{At: time.Now(), StatsSnapshot: snap})
				if err != nil {
					continue
				}
				f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					slog.Error("http stats dump: open failed", "path", path, "error", err)
					continue
				}
				if _, err := f.Write(append(line, '\n')); err != nil {
					slog.Error("http stats dump: write failed", "path", path, "error", err)
				}
				_ = f.Close()
			}
		}
	}()
}

// bucketQuantile returns the upper bound of the bucket containing quantile q;
// the overflow bucket reports the observed max.
func bucketQuantile(buckets []uint64, count uint64, q float64, maxMs float64) float64 {
	if count == 0 {
		return 0
	}
	target := uint64(q*float64(count) + 0.999999)
	var cum uint64
	for i, n := range buckets {
		cum += n
		if cum >= target {
			if i < len(bucketBoundsMs) {
				return bucketBoundsMs[i]
			}
			return maxMs
		}
	}
	return maxMs
}
