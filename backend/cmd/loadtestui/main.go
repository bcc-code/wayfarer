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
	Completions string            `json:"completions"`
	Failures    string            `json:"failures"`
	HTTPReqs    string            `json:"http_reqs"`
	ReqRate     string            `json:"req_rate"`
	Duration    map[string]string `json:"duration,omitempty"` // avg/min/med/max/p90/p95
	Ops         []opStats         `json:"ops,omitempty"`      // per-operation breakdown
	JournalD    string            `json:"score_journal_delta"`
	ServerCPU   string            `json:"server_peak_cpu"`
	PgCPU       string            `json:"pg_peak_cpu"`
	BoxBusy     string            `json:"box_peak_busy"`
	MTime       int64             `json:"mtime"`
}

type opStats struct {
	Name  string            `json:"name"`
	Stats map[string]string `json:"stats"`
}

// statsOf parses "avg=1.2ms min=... med=... max=... p(90)=... p(95)=..." pairs
func statsOf(line string) map[string]string {
	out := map[string]string{}
	for _, m := range regexp.MustCompile(`([a-z]+|p\(\d+\))=([^\s]+)`).FindAllStringSubmatch(line, -1) {
		k := strings.NewReplacer("p(", "p", ")", "").Replace(m[1])
		out[k] = m[2]
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
			reqs := extract(logTxt, "http_reqs")
			row.HTTPReqs = firstField(reqs)
			if f := strings.Fields(reqs); len(f) > 1 {
				row.ReqRate = f[1]
			}
			row.Duration = statsOf(extract(logTxt, "http_req_duration"))
			for _, m := range regexp.MustCompile(`(?m)^\s*\{ name:([A-Za-z]+) \}\.*:\s*(.+)$`).FindAllStringSubmatch(logTxt, -1) {
				replaced := false
				for i := range row.Ops {
					if row.Ops[i].Name == m[1] {
						row.Ops[i].Stats = statsOf(m[2]) // keep last occurrence (final summary)
						replaced = true
						break
					}
				}
				if !replaced {
					row.Ops = append(row.Ops, opStats{Name: m[1], Stats: statsOf(m[2])})
				}
			}
			if m := regexp.MustCompile(`score_journal delta=(\d+)`).FindStringSubmatch(logTxt); m != nil {
				row.JournalD = m[1]
			}
			if m := regexp.MustCompile(`server peak %CPU: (\d+)`).FindStringSubmatch(logTxt); m != nil {
				row.ServerCPU = m[1]
			}
			if m := regexp.MustCompile(`postgres peak %CPU: (\d+)`).FindStringSubmatch(logTxt); m != nil {
				row.PgCPU = m[1]
			}
			if m := regexp.MustCompile(`box peak busy: ([0-9.]+%)`).FindStringSubmatch(logTxt); m != nil {
				row.BoxBusy = m[1]
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
<div id="pctchart"></div>
<table><thead><tr>
  <th>label</th><th>status</th><th>ramp</th><th>think</th><th>done</th><th>fail</th>
  <th>reqs</th><th>req/s</th><th>avg</th><th>med</th><th>p90</th><th>p95</th><th>max</th>
  <th>journal Δ</th><th>srv%</th><th>pg%</th><th>box</th><th>when</th><th>note</th>
</tr></thead><tbody id="rows"></tbody></table>
<h1 style="margin-top:24px" id="opstitle"></h1>
<div id="rampchart"></div>
<table id="opstable" style="display:none"><thead><tr>
  <th>operation</th><th>avg</th><th>min</th><th>med</th><th>p90</th><th>p95</th><th>max</th>
</tr></thead><tbody id="oprows"></tbody></table>
<h1 style="margin-top:24px" id="logtitle"></h1>
<pre id="log" style="display:none"></pre>
<script>
let selected = null;
function toMs(v){
  if(!v) return null;
  if(v.endsWith('\u00b5s')) return parseFloat(v)/1000;
  if(v.endsWith('ms')) return parseFloat(v);
  if(v.endsWith('s')) return parseFloat(v)*1000;
  return parseFloat(v);
}
function fmtMs(ms){ return ms>=1000 ? (ms/1000).toFixed(1)+'s' : ms>=1 ? Math.round(ms)+'ms' : ms.toFixed(2)+'ms'; }
function pctChart(runs){
  const data = runs.filter(r=>r.duration&&r.duration.p95).map(r=>({
    label:r.label, p50:toMs(r.duration.med), p90:toMs(r.duration.p90), p95:toMs(r.duration.p95)})).reverse();
  const el = document.getElementById('pctchart');
  if(data.length<1){ el.innerHTML=''; return; }
  const W=1000,H=260,L=64,B=84,T=16,R=16,n=data.length;
  const vals=data.flatMap(d=>[d.p50,d.p90,d.p95]).filter(v=>v>0);
  let lo=Math.min.apply(null,vals), hi=Math.max.apply(null,vals);
  if(lo===hi){ lo/=2; hi*=2; }
  const ly=v=>T+(1-(Math.log10(v)-Math.log10(lo))/(Math.log10(hi)-Math.log10(lo)))*(H-T-B);
  const x=i=>n>1 ? L+i*(W-L-R)/(n-1) : (L+W-R)/2;
  const series=[['p50','#4ec97a'],['p90','#e0b13f'],['p95','#e05252']];
  let svg='<svg viewBox="0 0 '+W+' '+H+'" style="width:100%;max-width:'+W+'px">';
  for(const tv of [lo, Math.sqrt(lo*hi), hi]){
    const y=ly(tv);
    svg+='<line x1="'+L+'" y1="'+y+'" x2="'+(W-R)+'" y2="'+y+'" stroke="#2a2d33"/>'+
         '<text x="'+(L-6)+'" y="'+(y+4)+'" fill="#9aa0aa" font-size="11" text-anchor="end">'+fmtMs(tv)+'</text>';
  }
  for(const [key,color] of series){
    let pts='';
    data.forEach((d,i)=>{ if(d[key]>0) pts+=x(i)+','+ly(d[key])+' '; });
    svg+='<polyline points="'+pts+'" fill="none" stroke="'+color+'" stroke-width="2"/>';
    data.forEach((d,i)=>{ if(d[key]>0) svg+='<circle cx="'+x(i)+'" cy="'+ly(d[key])+'" r="3.5" fill="'+color+'"><title>'+esc(d.label)+' '+key+': '+fmtMs(d[key])+'</title></circle>'; });
  }
  data.forEach((d,i)=>{
    svg+='<text x="'+x(i)+'" y="'+(H-B+16)+'" fill="#9aa0aa" font-size="11" text-anchor="end" transform="rotate(-30 '+x(i)+' '+(H-B+16)+')">'+esc(d.label)+'</text>';
  });
  series.forEach(([key,color],i)=>{
    const lx=L+i*70;
    svg+='<rect x="'+lx+'" y="'+(H-16)+'" width="12" height="12" fill="'+color+'"/>'+
         '<text x="'+(lx+18)+'" y="'+(H-6)+'" fill="#d6d8de" font-size="12">'+key+'</text>';
  });
  el.innerHTML = svg+'</svg>';
}
function rampChart(run){
  const el = document.getElementById('rampchart');
  const scale = parseFloat(run.ramp_scale)||1;
  const pts=[[0,0],[1,1000],[2,5000],[4,8000],[10,10000]].map(p=>[p[0],p[1]*scale]);
  const W=1000,H=200,L=64,B=32,T=14,R=16,xmax=10,ymax=pts[pts.length-1][1];
  const gx=t=>L+t*(W-L-R)/xmax, gy=u=>T+(1-u/ymax)*(H-T-B);
  let svg='<svg viewBox="0 0 '+W+' '+H+'" style="width:100%;max-width:'+W+'px">';
  for(let t=0;t<=xmax;t+=2)
    svg+='<line x1="'+gx(t)+'" y1="'+T+'" x2="'+gx(t)+'" y2="'+(H-B)+'" stroke="#2a2d33"/>'+
         '<text x="'+gx(t)+'" y="'+(H-B+16)+'" fill="#9aa0aa" font-size="11" text-anchor="middle">'+t+'s</text>';
  for(const f of [0.25,0.5,0.75,1]){
    const u=ymax*f;
    svg+='<line x1="'+L+'" y1="'+gy(u)+'" x2="'+(W-R)+'" y2="'+gy(u)+'" stroke="#2a2d33"/>'+
         '<text x="'+(L-6)+'" y="'+(gy(u)+4)+'" fill="#9aa0aa" font-size="11" text-anchor="end">'+Math.round(u).toLocaleString()+'</text>';
  }
  const line=pts.map(p=>gx(p[0])+','+gy(p[1])).join(' ');
  svg+='<polygon points="'+gx(0)+','+gy(0)+' '+line+' '+gx(xmax)+','+gy(0)+'" fill="#2f6feb22"/>';
  svg+='<polyline points="'+line+'" fill="none" stroke="#5aa7ff" stroke-width="2.5"/>';
  pts.forEach(p=>{ svg+='<circle cx="'+gx(p[0])+'" cy="'+gy(p[1])+'" r="3.5" fill="#5aa7ff"><title>'+p[0]+'s: '+Math.round(p[1]).toLocaleString()+' users</title></circle>'; });
  svg+='<text x="'+(W-R)+'" y="'+(T+2)+'" fill="#9aa0aa" font-size="12" text-anchor="end">cumulative arrivals (ramp scale '+scale+')</text>';
  el.innerHTML = svg+'</svg>';
}
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
    const d = run.duration || {};
    tr.innerHTML = '<td>'+run.label+'</td><td class="'+run.status+'">'+run.status+'</td>'+
      '<td>'+(run.ramp_scale||'')+'</td><td>'+(run.think_scale||'')+'</td>'+
      '<td>'+(run.completions||'')+'</td><td>'+(run.failures||'')+'</td>'+
      '<td>'+(run.http_reqs||'')+'</td><td>'+(run.req_rate||'')+'</td>'+
      '<td>'+(d.avg||'')+'</td><td>'+(d.med||'')+'</td><td>'+(d.p90||'')+'</td>'+
      '<td>'+(d.p95||'')+'</td><td>'+(d.max||'')+'</td>'+
      '<td>'+(run.score_journal_delta||'')+'</td><td>'+(run.server_peak_cpu||'')+'</td>'+
      '<td>'+(run.pg_peak_cpu||'')+'</td><td>'+(run.box_peak_busy||'')+'</td><td>'+when+'</td>'+
      '<td style="white-space:normal;min-width:220px">'+esc(run.note||'')+' <a href="#" onclick="return editNote(\''+run.label+'\',this)">✎</a></td>';
    tr.onclick = () => { selected = run.label; showOps(run); showLog(); };
    if(run.label === selected) showOps(run);
    tb.appendChild(tr);
  }
  pctChart(d.runs||[]);
  if(d.running){ selected = selected || d.running; }
  if(selected) showLog();
}
function showOps(run){
  const tbl = document.getElementById('opstable'), tb = document.getElementById('oprows');
  document.getElementById('opstitle').textContent = 'per-operation: '+run.label;
  rampChart(run);
  tb.innerHTML='';
  for(const op of run.ops||[]){
    const s = op.stats||{};
    const tr = document.createElement('tr');
    tr.innerHTML = '<td>'+op.name+'</td><td>'+(s.avg||'')+'</td><td>'+(s.min||'')+'</td>'+
      '<td>'+(s.med||'')+'</td><td>'+(s.p90||'')+'</td><td>'+(s.p95||'')+'</td><td>'+(s.max||'')+'</td>';
    tb.appendChild(tr);
  }
  tbl.style.display = (run.ops&&run.ops.length) ? 'table' : 'none';
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
