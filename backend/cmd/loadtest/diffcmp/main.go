// diffcmp replays an identical battery of GraphQL operations and plugin
// webhook calls against two Wayfarer servers (A = baseline, B = candidate) in
// lockstep and structurally diffs the responses. It is the differential
// harness for proving functional equivalence between two builds running on
// byte-identical database clones.
//
// Usage:
//
//	diffcmp -a http://127.0.0.1:8080 -b http://127.0.0.1:8081 \
//	        -run '^R' -out results/
//
// Exit codes: 0 = all verdicts as expected, 1 = unexpected divergence or an
// expected divergence that did NOT reproduce, 2 = infrastructure error.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type stepResult struct {
	Scenario string   `json:"scenario"`
	Step     string   `json:"step"`
	Verdict  string   `json:"verdict"` // MATCH | DIVERGE | EXPECTED-DIVERGE | UNEXPECTED-MATCH | SKIPPED
	Probe    string   `json:"probe,omitempty"`
	Diff     []string `json:"diff,omitempty"`
	CheckA   string   `json:"checkA,omitempty"`
	CheckB   string   `json:"checkB,omitempty"`
}

type runner struct {
	baseA, baseB  string
	jwtSecret     string
	webhookSecret string
	outDir        string
	tokens        map[string]string
	results       []stepResult
}

func main() {
	baseA := flag.String("a", "http://127.0.0.1:8080", "base URL of side A (baseline)")
	baseB := flag.String("b", "http://127.0.0.1:8081", "base URL of side B (candidate)")
	jwtSecret := flag.String("jwt-secret", "your-secret-key-for-signing-wayfarer-jwts", "HS256 JWT secret shared with both servers")
	webhookSecret := flag.String("webhook-secret", "diff-webhook-secret", "Ladder to Heaven plugin webhook HMAC secret")
	outDir := flag.String("out", "results", "output directory for raw bodies, diffs and the summary")
	runFilter := flag.String("run", "", "regexp selecting scenarios by name (empty = all)")
	list := flag.Bool("list", false, "list scenario names and exit")
	flag.Parse()

	scenarios := buildScenarios()
	if *list {
		for _, s := range scenarios {
			fmt.Printf("%-24s %s\n", s.Name, s.Description)
		}
		return
	}

	var filter *regexp.Regexp
	if *runFilter != "" {
		var err error
		filter, err = regexp.Compile(*runFilter)
		if err != nil {
			fatal("bad -run regexp: %v", err)
		}
	}

	r := &runner{
		baseA: strings.TrimRight(*baseA, "/"), baseB: strings.TrimRight(*baseB, "/"),
		jwtSecret: *jwtSecret, webhookSecret: *webhookSecret, outDir: *outDir,
		tokens: map[string]string{},
	}
	runStart := time.Now()

	for _, sc := range scenarios {
		if filter != nil && !filter.MatchString(sc.Name) {
			continue
		}
		if err := r.runScenario(sc, runStart); err != nil {
			fatal("scenario %s: %v", sc.Name, err)
		}
	}

	unexpected := r.summarize()
	if err := r.writeSummary(); err != nil {
		fatal("writing summary: %v", err)
	}
	if unexpected > 0 {
		os.Exit(1)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "diffcmp: "+format+"\n", args...)
	os.Exit(2)
}

// token returns (and caches) a JWT for a step's User spec.
func (r *runner) token(user string) (string, string, error) {
	if user == "" {
		return "", "", nil
	}
	if tok, ok := r.tokens[user]; ok {
		return tok, r.userID(user), nil
	}
	var tok string
	var err error
	switch {
	case user == "admin":
		tok, err = mintToken(r.jwtSecret, r.userID(user), []string{"user", "admin", "superadmin"})
	case strings.HasPrefix(user, "u"):
		tok, err = mintToken(r.jwtSecret, r.userID(user), []string{"user"})
	default:
		return "", "", fmt.Errorf("unknown user spec %q", user)
	}
	if err != nil {
		return "", "", err
	}
	r.tokens[user] = tok
	return tok, r.userID(user), nil
}

func (r *runner) userID(user string) string {
	if user == "admin" {
		return seededUserID(1)
	}
	var n int
	fmt.Sscanf(user, "u%d", &n)
	return seededUserID(n)
}

func (r *runner) runScenario(sc Scenario, runStart time.Time) error {
	fmt.Printf("=== %s — %s\n", sc.Name, sc.Description)
	dir := filepath.Join(r.outDir, sc.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	capturedA := map[string]string{}
	capturedB := map[string]string{}
	// Per-side ULID placeholder maps span the scenario so cross-step
	// references (e.g. a submission id) normalize consistently.
	optsA := &NormOpts{RunStart: runStart, RedactPaths: sc.RedactPaths, SortPaths: sc.SortPaths}
	optsB := &NormOpts{RunStart: runStart, RedactPaths: sc.RedactPaths, SortPaths: sc.SortPaths}

	for i, step := range sc.Steps {
		if step.SleepBefore > 0 {
			fmt.Printf("    (sleeping %s before %s)\n", step.SleepBefore, step.Name)
			time.Sleep(step.SleepBefore)
		}
		respA, err := r.fire(r.baseA, step, capturedA)
		if err != nil {
			return fmt.Errorf("step %s side A: %w", step.Name, err)
		}
		respB, err := r.fire(r.baseB, step, capturedB)
		if err != nil {
			return fmt.Errorf("step %s side B: %w", step.Name, err)
		}

		prefix := fmt.Sprintf("%02d-%s", i, step.Name)
		writeRaw(filepath.Join(dir, prefix+"-a.json"), respA)
		writeRaw(filepath.Join(dir, prefix+"-b.json"), respB)

		if err := capture(step, respA, capturedA); err != nil {
			return fmt.Errorf("step %s side A: %w", step.Name, err)
		}
		if err := capture(step, respB, capturedB); err != nil {
			// On the candidate side a failed capture after an expected
			// divergence is possible; fail hard only when no divergence was
			// expected, otherwise later steps would silently misfire anyway.
			return fmt.Errorf("step %s side B: %w", step.Name, err)
		}

		res := r.compareStep(sc, step, respA, respB, optsA, optsB, dir, prefix, capturedA)
		r.results = append(r.results, res)
		fmt.Printf("    %-28s %s\n", step.Name, res.Verdict)
	}
	return nil
}

// fire executes one step against one side, honoring the step's cache mode.
func (r *runner) fire(base string, step Step, captured map[string]string) (*gqlResponse, error) {
	tok, _, err := r.token(step.User)
	if err != nil {
		return nil, err
	}

	if step.Query == "" { // REST step
		var body []byte
		if step.Body != nil {
			resolved, err := applyVars(map[string]any(step.Body), captured)
			if err != nil {
				return nil, err
			}
			if body, err = json.Marshal(resolved); err != nil {
				return nil, err
			}
		}
		secret := ""
		if step.Signed {
			secret = r.webhookSecret
		}
		return restRequest(base, step.Method, step.Path, body, tok, secret)
	}

	text, ok := queries[step.Query]
	if !ok {
		return nil, fmt.Errorf("unknown query %q", step.Query)
	}
	mode := step.CacheMode
	if mode == "" {
		mode = cacheCold
	}
	if mode == cacheCold {
		// Same nonce for both sides of a step (stepNonce is memoized per
		// step pointer via the captured map is NOT possible — use the step
		// name + scenario-scoped nonce stored in captured).
		key := "__nonce__" + step.Name
		nonce, okN := captured[key]
		if !okN {
			// First side generates; but captured maps are per side. Nonce
			// must be identical across sides, so derive it deterministically
			// from the step name and the shared run nonce.
			nonce = deriveNonce(step.Name)
			captured[key] = nonce
		}
		text = text + "\n# nonce:" + nonce
	}

	var vars map[string]any
	if step.Vars != nil {
		resolved, err := applyVars(map[string]any(step.Vars), captured)
		if err != nil {
			return nil, err
		}
		vars = resolved.(map[string]any)
	}

	resp, err := gqlPost(base, tok, text, vars)
	if err != nil {
		return nil, err
	}
	if mode == cacheWarm {
		// Identical second request; compare the warm response.
		return gqlPost(base, tok, text, vars)
	}
	return resp, nil
}

// runNonce is generated once per process so cold-mode texts differ between
// diffcmp invocations but are identical across both sides within one run.
var runNonce = func() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}()

func deriveNonce(stepName string) string {
	return runNonce + "-" + stepName
}

func capture(step Step, resp *gqlResponse, captured map[string]string) error {
	for name, path := range step.Capture {
		v, err := capturePath(resp.Data, path)
		if err != nil {
			return fmt.Errorf("capture %q: %w", name, err)
		}
		captured[name] = v
	}
	return nil
}

func (r *runner) compareStep(sc Scenario, step Step, respA, respB *gqlResponse, optsA, optsB *NormOpts, dir, prefix string, vars map[string]string) stepResult {
	res := stepResult{Scenario: sc.Name, Step: step.Name, Probe: step.ExpectDiverge}

	_, callerID, _ := r.token(step.User)
	resolvedVars, _ := applyVars(map[string]any(step.Vars), vars)
	rv, _ := resolvedVars.(map[string]any)
	if err := runCheck(step.Check, respA, callerID, rv); err != nil {
		res.CheckA = err.Error()
	}
	if err := runCheck(step.Check, respB, callerID, rv); err != nil {
		res.CheckB = err.Error()
	}

	if step.SkipDiff {
		res.Verdict = "SKIPPED"
		if res.CheckA != "" || res.CheckB != "" {
			res.Verdict = "DIVERGE" // a failed check on a skipped-diff step still counts
		}
		return res
	}

	var diffs []string
	if respA.Status != respB.Status {
		diffs = append(diffs, fmt.Sprintf("$.status: %d vs %d", respA.Status, respB.Status))
	}
	if step.Check != "statusOnly" {
		normA := Normalize(respA.Data, optsA)
		normB := Normalize(respB.Data, optsB)
		diffs = append(diffs, Diff(normA, normB)...)
		diffs = append(diffs, diffErrors(respA.Errors, respB.Errors)...)
	}
	// Per-side check asymmetry is itself a divergence (e.g. B failed a check
	// that A passed).
	if res.CheckA != res.CheckB {
		diffs = append(diffs, fmt.Sprintf("$.check: A=%q vs B=%q", res.CheckA, res.CheckB))
	}

	if len(diffs) > 0 {
		writeLines(filepath.Join(dir, prefix+"-diff.txt"), diffs)
		res.Diff = diffs
		if step.ExpectDiverge != "" {
			res.Verdict = "EXPECTED-DIVERGE"
		} else {
			res.Verdict = "DIVERGE"
		}
	} else {
		if step.ExpectDiverge != "" {
			res.Verdict = "UNEXPECTED-MATCH"
		} else {
			res.Verdict = "MATCH"
		}
	}
	return res
}

func diffErrors(a, b []gqlError) []string {
	var out []string
	if len(a) != len(b) {
		out = append(out, fmt.Sprintf("$.errors: count %d vs %d (A=%s, B=%s)", len(a), len(b), errText(a), errText(b)))
		return out
	}
	for i := range a {
		if a[i].Message != b[i].Message {
			out = append(out, fmt.Sprintf("$.errors[%d].message: %q vs %q", i, a[i].Message, b[i].Message))
		}
	}
	return out
}

func errText(errs []gqlError) string {
	if len(errs) == 0 {
		return "none"
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Message
	}
	return strings.Join(msgs, "; ")
}

func writeRaw(path string, resp *gqlResponse) {
	body := resp.Body
	if len(body) == 0 {
		body = fmt.Appendf(nil, `{"__status": %d}`, resp.Status)
	}
	_ = os.WriteFile(path, body, 0o644)
}

func writeLines(path string, lines []string) {
	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// summarize prints the verdict table and returns the count of unexpected
// results (plain divergences plus probes that failed to reproduce).
func (r *runner) summarize() int {
	unexpected := 0
	fmt.Println("\n================ SUMMARY ================")
	for _, res := range r.results {
		flag := ""
		switch res.Verdict {
		case "DIVERGE":
			unexpected++
			flag = "  <-- UNEXPECTED"
		case "UNEXPECTED-MATCH":
			unexpected++
			flag = "  <-- probe " + res.Probe + " did NOT reproduce (suspicion disproved?)"
		}
		fmt.Printf("%-24s %-28s %-18s%s\n", res.Scenario, res.Step, res.Verdict, flag)
	}
	fmt.Printf("\n%d steps, %d unexpected results\n", len(r.results), unexpected)
	return unexpected
}

func (r *runner) writeSummary() error {
	b, err := json.MarshalIndent(r.results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.outDir, "summary.json"), b, 0o644)
}
