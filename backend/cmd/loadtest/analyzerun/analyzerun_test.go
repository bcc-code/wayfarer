package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeStats(t *testing.T) {
	t.Run("empty has only a count", func(t *testing.T) {
		s := computeStats(nil)
		assert.Equal(t, Stats{}, s)
		data, err := json.Marshal(s)
		require.NoError(t, err)
		assert.JSONEq(t, `{"count":0}`, string(data))
	})

	t.Run("nearest rank percentiles", func(t *testing.T) {
		vals := make([]float64, 0, 101)
		for i := 100; i >= 0; i-- { // unsorted on purpose
			vals = append(vals, float64(i))
		}
		s := computeStats(vals)
		assert.Equal(t, 101, s.Count)
		assert.Equal(t, 50.0, *s.P50)
		assert.Equal(t, 95.0, *s.P95)
		assert.Equal(t, 99.0, *s.P99)
		assert.Equal(t, 100.0, *s.Max)
	})

	t.Run("rounds to two decimals", func(t *testing.T) {
		s := computeStats([]float64{1.23456})
		assert.Equal(t, 1.23, *s.P50)
		assert.Equal(t, 1.23, *s.Max)
	})
}

func TestRoundHalfToEven(t *testing.T) {
	assert.Equal(t, 3396.2, round(3396.25, 1))
	assert.Equal(t, 3396.4, round(3396.35, 1))
	assert.Equal(t, 2.0, round(2.5, 0))
}

const pidstatOut = `Linux 6.12.63+deb13-cloud-amd64 (interact-lt-server) 	2026-10-13 	_x86_64_	(8 CPU)

# Time        UID       PID    %usr %system  %guest   %wait    %CPU   CPU  minflt/s  majflt/s     VSZ     RSS   %MEM  Command
10:00:49        0     40494    1.00    0.00    0.00    0.00   10.50     3      0.00      0.00 1402280   98116   0.31  postgres
10:00:49        0     40495    1.00    0.00    0.00    0.00    4.50     3      0.00      0.00 1402280    1024   0.31  postgres
10:00:50        0     40494    1.00    0.00    0.00    0.00   20.00     3      0.00      0.00 1402280   98116   0.31  postgres
10:00:51        0     40494    1.00    0.00    0.00    0.00   20.00     3      0.00      0.00 1402280   98116   0.31
Average:        0     40494    1.00    0.00    0.00    0.00   15.00     -      0.00      0.00 1402280   98116   0.31  postgres
`

func TestParsePidstat(t *testing.T) {
	t.Run("one process per time", func(t *testing.T) {
		got, err := parsePidstat(strings.NewReader(pidstatOut), false)
		require.NoError(t, err)
		assert.Equal(t, map[string]cpuSample{
			"10:00:49": {cpu: 4.5, rssKB: 1024}, // last row wins
			"10:00:50": {cpu: 20, rssKB: 98116},
		}, got, "rows not matching the header (10:00:51 lacks Command) and the average are skipped")
	})

	t.Run("summed over processes", func(t *testing.T) {
		got, err := parsePidstat(strings.NewReader(pidstatOut), true)
		require.NoError(t, err)
		assert.Equal(t, cpuSample{cpu: 15, rssKB: 99140}, got["10:00:49"])
		assert.Equal(t, cpuSample{cpu: 20, rssKB: 98116}, got["10:00:50"])
	})
}

func TestParseMpstat(t *testing.T) {
	out := `Linux 6.12.63+deb13-cloud-amd64 (interact-lt-server) 	2026-10-13 	_x86_64_	(8 CPU)

10:00:48     CPU    %usr   %nice    %sys %iowait    %irq   %soft  %steal  %guest  %gnice   %idle
10:00:49     all    2.01    0.00    1.38    0.00    0.00    0.00    0.00    0.00    0.00   96.61
10:00:50     all   50.00    0.00    1.13    0.00    0.00    0.00    0.00    0.00    0.00   25.555
Average:     all    0.50    0.00    1.13    0.00    0.00    0.00    0.00    0.00    0.00   98.37
`
	got, err := parseMpstat(strings.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, map[string]float64{"10:00:49": 3.39, "10:00:50": 74.44}, got)
}

const k6Header = "metric_name,timestamp,metric_value,check,error,error_code,expected_response,group,method,name,proto,scenario,service,status,subproto,tls_version,url,extra_tags,metadata\n"

func k6Row(metric string, ts int, value, name, url string) string {
	return metric + "," + strconv.Itoa(ts) + "," + value + ",,,,true,,POST," + name + ",HTTP/2.0,s,,200,,tls1.3," + url + ",,\n"
}

func TestK6Read(t *testing.T) {
	csv := k6Header +
		k6Row("http_reqs", 100, "1", "Bet", "") +
		k6Row("http_req_duration", 100, "10", "Bet", "") +
		k6Row("http_req_duration", 100, "30", "Bet", "") +
		k6Row("http_req_duration", 101, "5", "", "https://x/health") + // no name tag: URL
		k6Row("http_req_failed", 100, "1", "Bet", "") +
		k6Row("http_req_failed", 100, "0", "Bet", "") +
		k6Row("http_req_tls_handshaking", 100, "12.5", "Bet", "") +
		k6Row("http_req_tls_handshaking", 100, "0", "Bet", "") + // reused connection
		k6Row("bets_accepted", 100, "1", "", "") +
		k6Row("bets_accepted", 101, "1", "", "") +
		k6Row("settlement_duration", 102, "11200", "", "") +
		k6Row("vus", 102, "300", "", "") // not summarized

	k := &k6Data{
		durations: map[string][]float64{}, failed: map[string]int{},
		perSec: map[int64]map[string][]float64{}, perSecFailed: map[int64]map[string]int{},
		tls: map[int64][]float64{}, counters: map[string]float64{}, trends: map[string][]float64{},
	}
	require.NoError(t, k.read(strings.NewReader(csv)))

	assert.Equal(t, map[string][]float64{"Bet": {10, 30}, "https://x/health": {5}}, k.durations)
	assert.Equal(t, []float64{10, 30}, k.perSec[100]["Bet"])
	assert.Equal(t, map[string]int{"Bet": 1}, k.failed)
	assert.Equal(t, 1, k.perSecFailed[100]["Bet"])
	assert.Equal(t, map[int64][]float64{100: {12.5}}, k.tls)
	assert.Equal(t, map[string]float64{"bets_accepted": 2}, k.counters)
	assert.Equal(t, map[string][]float64{"settlement_duration": {11200}}, k.trends)
}

func TestK6ReadRejectsBadInput(t *testing.T) {
	k := &k6Data{}
	assert.Error(t, k.read(strings.NewReader("a,b,c\n")), "missing columns")
	k = &k6Data{counters: map[string]float64{}}
	assert.Error(t, k.read(strings.NewReader(k6Header+k6Row("bets_accepted", 1, "x", "", ""))), "bad value")
}

func TestSettledPerSecond(t *testing.T) {
	csv := `id,user_id,response_id,points,kind,created_at
SJ1,US1,QR1,-139,stake,2026-10-13 10:02:01.067006+00
SJ2,US1,QR1,417,winnings,2026-10-13 10:02:01.067006+00
SJ3,US2,QR2,-24,stake,2026-10-13 10:02:01.5+00
SJ4,US2,QR2,0,winnings,2026-10-13 10:02:01.5+00
SJ5,US3,QR3,0,winnings,2026-10-13 10:02:04+00
`
	got, err := settledPerSecond(strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"2026-10-13 10:02:01": 2, "2026-10-13 10:02:04": 1}, got)

	_, err = settledPerSecond(strings.NewReader("id,points\nSJ1,1\n"))
	assert.Error(t, err)
}

// writeRunDir builds a minimal run directory like loadtest-remote.sh writes
func writeRunDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	writeGz := func(rel, content string) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write([]byte(content))
		require.NoError(t, err)
		require.NoError(t, gz.Close())
		write(rel, buf.String())
	}
	writeGz("generator0/shard0.csv.gz", k6Header+
		k6Row("http_req_duration", 1791885721, "20", "Bet", "")+
		k6Row("http_req_tls_handshaking", 1791885721, "8", "Bet", "")+
		k6Row("bets_accepted", 1791885721, "1", "", ""))
	writeGz("generator1/shard1.csv.gz", k6Header+
		k6Row("http_req_duration", 1791885722, "40", "Bet", "")+
		k6Row("http_req_failed", 1791885722, "1", "Bet", "")+
		k6Row("bets_accepted", 1791885722, "1", "", ""))
	write("server/server.pidstat", pidstatOut)
	write("server/server-box.mpstat", "10:00:49     all    2.01    0.00    1.38    0.00    0.00    0.00    0.00    0.00    0.00   96.61\n")
	write("generator0/shard0.box.mpstat", "10:00:50     all    2.01    0.00    1.38    0.00    0.00    0.00    0.00    0.00    0.00   50.00\n")
	write("server/bet_journal.csv", `id,user_id,response_id,points,kind,created_at
SJ2,US1,QR1,417,winnings,2026-10-13 10:02:01.067006+00
SJ4,US2,QR2,0,winnings,2026-10-13 10:02:04.5+00
`)
	return dir
}

func TestAnalyze(t *testing.T) {
	dir := writeRunDir(t)
	var out bytes.Buffer
	require.NoError(t, analyze(dir, &out))

	data, err := os.ReadFile(filepath.Join(dir, "summary.json"))
	require.NoError(t, err)
	var s Summary
	require.NoError(t, json.Unmarshal(data, &s))
	assert.Equal(t, []string{"generator0/shard0.csv.gz", "generator1/shard1.csv.gz"}, s.K6Files)
	assert.Equal(t, 2, s.Requests.Count)
	assert.Equal(t, 1, s.Requests.Failed)
	assert.Equal(t, 40.0, *s.Requests.Max)
	assert.Equal(t, 1, s.TLSHandshakeMs.Count)
	assert.Equal(t, 1, s.TLSHandshakeMs.PeakPerSecond)
	assert.Equal(t, map[string]int64{"bets_accepted": 2}, s.Counters)
	assert.Equal(t, map[string]float64{
		"server_cpu_pct": 20, "server_rss_mb": 95.8, "server_box_busy_pct": 3.39, "generator0_box_busy_pct": 50,
	}, s.CPUPeaksPct)
	require.NotNil(t, s.Settlement)
	assert.Equal(t, Settlement{
		Bets: 2, First: "2026-10-13 10:02:01", Last: "2026-10-13 10:02:04",
		Seconds: 4, BetsPerSecond: 0.5, PeakPerSecond: 1,
	}, *s.Settlement)

	files := map[string]string{
		"requests_per_second.csv": "time_utc,epoch,name,count,failed,p50_ms,p95_ms,p99_ms,max_ms\n" +
			"2026-10-13T10:02:01Z,1791885721,Bet,1,0,20,20,20,20\n" +
			"2026-10-13T10:02:02Z,1791885722,Bet,1,1,40,40,40,40\n",
		"tls_per_second.csv": "time_utc,epoch,handshakes,p50_ms,p95_ms,max_ms\n" +
			"2026-10-13T10:02:01Z,1791885721,1,8,8,8\n",
		"cpu_per_second.csv": "time,server_cpu_pct,server_rss_mb,postgres_cpu_pct,server_box_busy_pct,generator0_box_busy_pct\n" +
			"10:00:49,4.5,1,,3.39,\n" +
			"10:00:50,20,95.8,,,50\n",
		"settlement_per_second.csv": "time_utc,bets_settled,cumulative\n" +
			"2026-10-13 10:02:01,1,1\n" +
			"2026-10-13 10:02:04,1,2\n",
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err, name)
		assert.Equal(t, want, string(got), name)
	}

	assert.Contains(t, out.String(), "requests 2 failed 1")
	assert.Contains(t, out.String(), "settlement: 2 bets in 4s (0.5/s, peak 1/s)")
}

func TestAnalyzeEmptyRunDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, analyze(dir, &bytes.Buffer{}))
	data, err := os.ReadFile(filepath.Join(dir, "summary.json"))
	require.NoError(t, err)
	var s map[string]any
	require.NoError(t, json.Unmarshal(data, &s))
	assert.Nil(t, s["settlement"])
	assert.Equal(t, map[string]any{"count": 0.0, "failed": 0.0}, s["requests"])
	_, err = os.Stat(filepath.Join(dir, "settlement_per_second.csv"))
	assert.True(t, os.IsNotExist(err))
}
