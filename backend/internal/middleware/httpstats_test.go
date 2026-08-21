package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPStatsRecordAndSnapshot(t *testing.T) {
	s := NewHTTPStats()
	s.record("POST /graphql", 200, 40)
	s.record("POST /graphql", 200, 90)
	s.record("POST /graphql", 500, 900)
	s.record("GET /health", 200, 0.5)

	snap := s.Snapshot()
	require.Len(t, snap.Routes, 2)

	gql := snap.Routes[0] // sorted by count desc
	assert.Equal(t, "POST /graphql", gql.Route)
	assert.Equal(t, uint64(3), gql.Count)
	assert.Equal(t, uint64(2), gql.ByStatus["2xx"])
	assert.Equal(t, uint64(1), gql.ByStatus["5xx"])
	assert.InDelta(t, (40.0+90+900)/3, gql.AvgMs, 0.01)
	assert.Equal(t, 900.0, gql.MaxMs)
	// histogram upper-bound estimates: 40 -> 50, 90 -> 100, 900 -> 1000
	assert.Equal(t, 100.0, gql.P50Ms)
	assert.Equal(t, 1000.0, gql.P99Ms)
}

func TestBucketQuantile(t *testing.T) {
	s := NewHTTPStats()
	for range 97 {
		s.record("GET /x", 200, 5) // bucket bound 5
	}
	for range 3 {
		s.record("GET /x", 200, 4000) // bucket bound 5000
	}

	r := s.Snapshot().Routes[0]
	assert.Equal(t, 5.0, r.P50Ms)
	assert.Equal(t, 5.0, r.P95Ms)
	assert.Equal(t, 5000.0, r.P99Ms) // 3 slow outliers push the top 1% into the 5000 bucket
	assert.Equal(t, 4000.0, r.MaxMs)
}

func TestHTTPStatsOverflowBucketReportsMax(t *testing.T) {
	s := NewHTTPStats()
	s.record("GET /slow", 200, 60000) // beyond the last bound
	r := s.Snapshot().Routes[0]
	assert.Equal(t, 60000.0, r.P99Ms)
}

func TestHTTPStatsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := NewHTTPStats()
	s.SlowThreshold = time.Hour // don't trip slow logging in tests

	router := gin.New()
	router.Use(s.Middleware())
	router.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	router.GET("/boom", func(c *gin.Context) { c.String(http.StatusInternalServerError, "boom") })

	for _, path := range []string{"/ok", "/ok", "/boom", "/missing"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(w, req)
	}

	snap := s.Snapshot()
	byRoute := map[string]RouteSnapshot{}
	for _, r := range snap.Routes {
		byRoute[r.Route] = r
	}

	assert.Equal(t, uint64(2), byRoute["GET /ok"].Count)
	assert.Equal(t, uint64(1), byRoute["GET /boom"].ByStatus["5xx"])
	assert.Equal(t, uint64(1), byRoute["GET (unmatched)"].Count)
}

func TestStartDumperWritesJSONL(t *testing.T) {
	s := NewHTTPStats()
	path := t.TempDir() + "/stats.jsonl"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.record("GET /a", 200, 3)
	s.StartDumper(ctx, path, 20*time.Millisecond)

	require.Eventually(t, func() bool {
		b, err := os.ReadFile(path)
		return err == nil && len(b) > 0
	}, 2*time.Second, 10*time.Millisecond)

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	line := bytes.SplitN(b, []byte("\n"), 2)[0]
	var rec struct {
		At     time.Time       `json:"at"`
		Routes []RouteSnapshot `json:"routes"`
	}
	require.NoError(t, json.Unmarshal(line, &rec))
	assert.False(t, rec.At.IsZero())
	require.Len(t, rec.Routes, 1)
	assert.Equal(t, "GET /a", rec.Routes[0].Route)

	// idle ticks must not append more lines
	firstSize := len(b)
	time.Sleep(80 * time.Millisecond)
	b2, _ := os.ReadFile(path)
	assert.Equal(t, firstSize, len(b2))
}

func TestDrainWindowResetsWindowButNotCumulative(t *testing.T) {
	s := NewHTTPStats()
	s.record("GET /a", 200, 10)

	first := s.DrainWindow()
	require.Len(t, first.Routes, 1)
	assert.Equal(t, uint64(1), first.Routes[0].Count)

	// window is empty now, cumulative is not
	assert.Empty(t, s.DrainWindow().Routes)
	require.Len(t, s.Snapshot().Routes, 1)
	assert.Equal(t, uint64(1), s.Snapshot().Routes[0].Count)

	// a fresh window only sees new traffic
	s.record("GET /a", 200, 30)
	second := s.DrainWindow()
	require.Len(t, second.Routes, 1)
	assert.Equal(t, uint64(1), second.Routes[0].Count)
	assert.InDelta(t, 30.0, second.Routes[0].AvgMs, 0.01)
	assert.Equal(t, uint64(2), s.Snapshot().Routes[0].Count)
}
