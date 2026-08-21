// Command loadtestui is a quick-and-dirty web UI for firing off-box load
// test runs (scripts/loadtest-remote.sh) and browsing the history of past
// runs from the results directory. Single stdlib-only binary, meant to run
// on the load-generator VM.
//
//	loadtestui -listen :8090 -backend /root/wayfarer/backend
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	backendDir string
	resultsDir string
	labelRe    = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)
	floatRe    = regexp.MustCompile(`^[0-9.]{1,8}$`)

	runMu   sync.Mutex
	running string // label of the in-flight run, "" if idle
)

type meta struct {
	Label      string `json:"label"`
	Started    string `json:"started"`
	Finished   string `json:"finished,omitempty"`
	RampScale  string `json:"ramp_scale"`
	ThinkScale string `json:"think_scale"`
	ExitCode   *int   `json:"exit_code,omitempty"`
	Note       string `json:"note,omitempty"` // what change was under test
	Status     string `json:"status"`         // running | passed | thresholds-crossed | failed
}

type runRow struct {
	meta
	Completions string `json:"completions"`
	Failures    string `json:"failures"`
	HTTPReqs    string `json:"http_reqs"`
	P95         string `json:"p95"`
	JournalD    string `json:"score_journal_delta"`
	ServerCPU   string `json:"server_peak_cpu"`
	PgCPU       string `json:"pg_peak_cpu"`
	MTime       int64  `json:"mtime"`
}

func metaPath(label string) string { return filepath.Join(resultsDir, label+".meta.json") }
func logPath(label string) string  { return filepath.Join(resultsDir, label+".run.log") }

func writeMeta(m meta) {
	b, _ := json.MarshalIndent(m, "", "  ")
	_ = os.WriteFile(metaPath(m.Label), b, 0o644)
}

// pull a "name...: value more" metric out of a k6/script log
func extract(log, name string) string {
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(name) + `\.*:?\s*(.+)$`)
	ms := re.FindAllStringSubmatch(log, -1)
	if len(ms) == 0 {
		return ""
	}
	return strings.TrimSpace(ms[len(ms)-1][1]) // last occurrence = final summary
}

func firstField(s string) string {
	if i := strings.IndexAny(s, " \t"); i > 0 {
		return s[:i]
	}
	return s
}

func p95Of(durLine string) string {
	if m := regexp.MustCompile(`p\(95\)=([^\s]+)`).FindStringSubmatch(durLine); m != nil {
		return m[1]
	}
	return ""
}

func rowFor(label string) runRow {
	row := runRow{meta: meta{Label: label, Status: "unknown"}}
	if b, err := os.ReadFile(metaPath(label)); err == nil {
		_ = json.Unmarshal(b, &row.meta)
	}
	src := logPath(label)
	if _, err := os.Stat(src); err != nil {
		src = filepath.Join(resultsDir, label+".k6.log") // legacy runs from before the UI
	}
	if st, err := os.Stat(src); err == nil {
		row.MTime = st.ModTime().Unix()
		if b, err := os.ReadFile(src); err == nil {
			logTxt := string(b)
			row.Completions = firstField(extract(logTxt, "quiz_completions"))
			row.Failures = firstField(extract(logTxt, "quiz_failures"))
			row.HTTPReqs = firstField(extract(logTxt, "http_reqs"))
			row.P95 = p95Of(extract(logTxt, "http_req_duration"))
			if m := regexp.MustCompile(`score_journal delta=(\d+)`).FindStringSubmatch(logTxt); m != nil {
				row.JournalD = m[1]
			}
			if m := regexp.MustCompile(`server peak %CPU: (\d+)`).FindStringSubmatch(logTxt); m != nil {
				row.ServerCPU = m[1]
			}
			if m := regexp.MustCompile(`postgres peak %CPU: (\d+)`).FindStringSubmatch(logTxt); m != nil {
				row.PgCPU = m[1]
			}
			if row.Status == "unknown" || row.Status == "" {
				switch {
				case strings.Contains(logTxt, "have been crossed"):
					row.Status = "thresholds-crossed"
				case row.Completions != "":
					row.Status = "passed"
				}
			}
		}
	}
	return row
}

func listRuns() []runRow {
	seen := map[string]bool{}
	var rows []runRow
	entries, _ := os.ReadDir(resultsDir)
	for _, e := range entries {
		n := e.Name()
		var label string
		switch {
		case strings.HasSuffix(n, ".meta.json"):
			label = strings.TrimSuffix(n, ".meta.json")
		case strings.HasSuffix(n, ".k6.log"):
			label = strings.TrimSuffix(n, ".k6.log")
		default:
			continue
		}
		if seen[label] {
			continue
		}
		seen[label] = true
		rows = append(rows, rowFor(label))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].MTime > rows[j].MTime })
	return rows
}

func startRun(label, ramp, think, note string) error {
	runMu.Lock()
	defer runMu.Unlock()
	if running != "" {
		return fmt.Errorf("a run is already in progress: %s", running)
	}
	logF, err := os.Create(logPath(label))
	if err != nil {
		return err
	}
	sh := fmt.Sprintf(
		"ulimit -n 65535; cd %s && ./scripts/loadtest-remote.sh run %s --env THINK_SCALE=%s --env RAMP_SCALE=%s",
		backendDir, label, think, ramp)
	cmd := exec.Command("bash", "-c", sh)
	cmd.Stdout, cmd.Stderr = logF, logF
	if err := cmd.Start(); err != nil {
		logF.Close()
		return err
	}
	running = label
	m := meta{Label: label, Started: time.Now().Format(time.RFC3339), RampScale: ramp, ThinkScale: think, Note: note, Status: "running"}
	writeMeta(m)
	go func() {
		err := cmd.Wait()
		logF.Close()
		code := 0
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else if err != nil {
			code = -1
		}
		m.Finished = time.Now().Format(time.RFC3339)
		m.ExitCode = &code
		switch code {
		case 0:
			m.Status = "passed"
		case 99:
			m.Status = "thresholds-crossed"
		default:
			m.Status = "failed"
		}
		writeMeta(m)
		runMu.Lock()
		running = ""
		runMu.Unlock()
	}()
	return nil
}

func main() {
	listen := flag.String("listen", ":8090", "listen address")
	flag.StringVar(&backendDir, "backend", "/root/wayfarer/backend", "backend dir containing scripts/ and cmd/loadtest/")
	flag.Parse()
	resultsDir = filepath.Join(backendDir, "cmd", "loadtest", "results")
	_ = os.MkdirAll(resultsDir, 0o755)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})
	http.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		runMu.Lock()
		cur := running
		runMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"running": cur, "runs": listRuns()})
	})
	http.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct{ Label, Ramp, Think, Note string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if req.Label == "" {
			req.Label = "run_" + time.Now().Format("0102_150405")
		}
		if req.Ramp == "" {
			req.Ramp = "1"
		}
		if req.Think == "" {
			req.Think = "1"
		}
		if !labelRe.MatchString(req.Label) || !floatRe.MatchString(req.Ramp) || !floatRe.MatchString(req.Think) {
			http.Error(w, "bad label/ramp/think", 400)
			return
		}
		if err := startRun(req.Label, req.Ramp, req.Think, req.Note); err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"label":%q}`, req.Label)
	})
	http.HandleFunc("/api/note", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct{ Label, Note string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !labelRe.MatchString(req.Label) {
			http.Error(w, "bad request", 400)
			return
		}
		row := rowFor(req.Label)
		if row.Started == "" && row.MTime == 0 {
			http.Error(w, "unknown run", 404)
			return
		}
		row.meta.Label = req.Label
		row.meta.Note = req.Note
		if row.meta.Status == "" {
			row.meta.Status = row.Status
		}
		writeMeta(row.meta)
		w.WriteHeader(204)
	})
	http.HandleFunc("/api/log", func(w http.ResponseWriter, r *http.Request) {
		label := r.URL.Query().Get("label")
		if !labelRe.MatchString(label) {
			http.Error(w, "bad label", 400)
			return
		}
		b, err := os.ReadFile(logPath(label))
		if err != nil {
			b, err = os.ReadFile(filepath.Join(resultsDir, label+".k6.log"))
		}
		if err != nil {
			http.Error(w, "no log", 404)
			return
		}
		const max = 64 * 1024
		if len(b) > max {
			b = b[len(b)-max:]
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write(b)
	})

	log.Printf("loadtestui listening on %s (backend %s)", *listen, backendDir)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

const indexHTML = `<!doctype html>
<meta charset="utf-8">
<title>Wayfarer Loadtest</title>
<style>
  body{font:14px/1.5 ui-monospace,Menlo,monospace;background:#14161a;color:#d6d8de;margin:0;padding:24px;max-width:1100px}
  h1{font-size:18px;margin:0 0 16px}
  form{display:flex;gap:8px;align-items:center;flex-wrap:wrap;margin-bottom:8px}
  input{background:#1d2026;border:1px solid #333;color:#d6d8de;padding:6px 8px;border-radius:4px;width:110px}
  button{background:#2f6feb;color:#fff;border:0;padding:7px 16px;border-radius:4px;cursor:pointer}
  button:disabled{background:#444}
  #status{margin:8px 0 16px;color:#9aa0aa}
  table{border-collapse:collapse;width:100%}
  th,td{text-align:left;padding:6px 10px;border-bottom:1px solid #2a2d33;white-space:nowrap}
  tr:hover td{background:#1b1e24;cursor:pointer}
  .passed{color:#4ec97a}.thresholds-crossed{color:#e0b13f}.failed{color:#e05252}.running{color:#5aa7ff}
  pre{background:#0e1013;border:1px solid #2a2d33;border-radius:6px;padding:12px;overflow:auto;max-height:50vh}
</style>
<h1>Wayfarer off-box loadtest</h1>
<form onsubmit="fire(event)">
  label <input id="label" placeholder="(auto)">
  ramp scale <input id="ramp" value="1">
  think scale <input id="think" value="0.06">
  note <input id="note" placeholder="what change is under test" style="width:320px">
  <button id="go">Run</button>
</form>
<div id="status">idle</div>
<table><thead><tr>
  <th>label</th><th>status</th><th>ramp</th><th>think</th><th>completions</th><th>failures</th>
  <th>http reqs</th><th>p95</th><th>journal Δ</th><th>srv CPU%</th><th>pg CPU%</th><th>when</th><th>note</th>
</tr></thead><tbody id="rows"></tbody></table>
<h1 style="margin-top:24px" id="logtitle"></h1>
<pre id="log" style="display:none"></pre>
<script>
let selected = null;
function esc(s){const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
async function editNote(label,el){
  const cur = el.parentElement.textContent.replace(' ✎','').trim();
  const note = prompt('Note for '+label, cur);
  if(note===null) return false;
  await fetch('/api/note',{method:'POST',body:JSON.stringify({label,note})});
  refresh(); return false;
}
async function refresh(){
  const r = await fetch('/api/runs'); const d = await r.json();
  document.getElementById('status').textContent = d.running ? 'RUNNING: '+d.running : 'idle';
  document.getElementById('go').disabled = !!d.running;
  const tb = document.getElementById('rows'); tb.innerHTML='';
  for(const run of d.runs||[]){
    const tr = document.createElement('tr');
    const when = run.mtime ? new Date(run.mtime*1000).toLocaleTimeString() : '';
    tr.innerHTML = '<td>'+run.label+'</td><td class="'+run.status+'">'+run.status+'</td>'+
      '<td>'+(run.ramp_scale||'')+'</td><td>'+(run.think_scale||'')+'</td>'+
      '<td>'+(run.completions||'')+'</td><td>'+(run.failures||'')+'</td>'+
      '<td>'+(run.http_reqs||'')+'</td><td>'+(run.p95||'')+'</td>'+
      '<td>'+(run.score_journal_delta||'')+'</td><td>'+(run.server_peak_cpu||'')+'</td>'+
      '<td>'+(run.pg_peak_cpu||'')+'</td><td>'+when+'</td>'+
      '<td style="white-space:normal;min-width:220px">'+esc(run.note||'')+' <a href="#" onclick="return editNote(\''+run.label+'\',this)">✎</a></td>';
    tr.onclick = () => { selected = run.label; showLog(); };
    tb.appendChild(tr);
  }
  if(d.running){ selected = selected || d.running; }
  if(selected) showLog();
}
async function showLog(){
  const r = await fetch('/api/log?label='+encodeURIComponent(selected));
  const el = document.getElementById('log');
  document.getElementById('logtitle').textContent = 'log: '+selected;
  el.style.display='block';
  const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 30;
  el.textContent = r.ok ? await r.text() : '(no log)';
  if(atBottom) el.scrollTop = el.scrollHeight;
}
async function fire(e){
  e.preventDefault();
  const body = JSON.stringify({label:document.getElementById('label').value.trim(),
    ramp:document.getElementById('ramp').value.trim(), think:document.getElementById('think').value.trim(),
    note:document.getElementById('note').value.trim()});
  const r = await fetch('/api/run',{method:'POST',body});
  if(!r.ok){ alert(await r.text()); return; }
  const d = await r.json(); selected = d.label;
  refresh();
}
refresh(); setInterval(refresh, 3000);
</script>`
