#!/usr/bin/env python3
"""Summarize a load-test run directory written by scripts/loadtest-remote.sh.

    python3 -I analyze_run.py <run-dir>

Reads the raw data (k6 CSV per shard, pidstat/mpstat samples, SQL exports)
and writes derived, graph-ready files next to it:

  requests_per_second.csv  per second x request name: count, failed,
                           p50/p95/p99/max http_req_duration (ms)
  tls_per_second.csv       per second: new TLS handshakes, p50/p95 (ms)
  cpu_per_second.csv       per second: server process / postgres / server
                           box busy %, server RSS, and k6 + box busy % per
                           generator
  settlement_per_second.csv  per second: bets settled (from the score_journal
                           export's created_at), cumulative
  summary.json             everything printed below, machine-readable

The raw files stay untouched; these are conveniences for graphing.
Standard library only.
"""

import csv
import glob
import gzip
import json
import os
import sys
from collections import defaultdict
from datetime import datetime, timezone


def pct(sorted_vals, p):
    if not sorted_vals:
        return None
    return sorted_vals[min(len(sorted_vals) - 1, int(round((len(sorted_vals) - 1) * p)))]


def stats(vals):
    vals = sorted(vals)
    if not vals:
        return {"count": 0}
    return {
        "count": len(vals),
        "p50": round(pct(vals, 0.50), 2),
        "p95": round(pct(vals, 0.95), 2),
        "p99": round(pct(vals, 0.99), 2),
        "max": round(vals[-1], 2),
    }


def iso(ts):
    return datetime.fromtimestamp(ts, timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def read_k6(run_dir):
    """Stream every k6 CSV (shards + settle) once, aggregating what we need."""
    durations = defaultdict(list)  # name -> [ms]
    failed = defaultdict(int)  # name -> failed requests
    per_sec = defaultdict(lambda: defaultdict(list))  # ts -> name -> [ms]
    per_sec_failed = defaultdict(lambda: defaultdict(int))
    tls = defaultdict(list)  # ts -> [ms]
    tls_all = []
    counters = defaultdict(float)
    trends = defaultdict(list)
    files = sorted(glob.glob(os.path.join(run_dir, "generator*", "*.csv.gz")))
    files += sorted(glob.glob(os.path.join(run_dir, "local.csv.gz")))
    custom_counters = {"bets_accepted", "bets_none", "bets_double_tap", "bets_rejected_expected",
                       "bets_wrong_outcome", "quiz_completions", "quiz_failures", "quiz_skipped",
                       "dropped_iterations", "iterations", "graphql_errors"}
    custom_trends = {"settlement_duration", "bet_amount", "iteration_duration"}
    for path in files:
        with gzip.open(path, "rt", newline="") as f:
            for row in csv.DictReader(f):
                metric = row["metric_name"]
                value = float(row["metric_value"])
                if metric == "http_req_duration":
                    name = row.get("name") or row.get("url")
                    ts = int(row["timestamp"])
                    durations[name].append(value)
                    per_sec[ts][name].append(value)
                elif metric == "http_req_failed":
                    if value > 0:
                        name = row.get("name") or row.get("url")
                        failed[name] += 1
                        per_sec_failed[int(row["timestamp"])][name] += 1
                elif metric == "http_req_tls_handshaking":
                    if value > 0:
                        tls[int(row["timestamp"])].append(value)
                        tls_all.append(value)
                elif metric in custom_counters:
                    counters[metric] += value
                elif metric in custom_trends:
                    trends[metric].append(value)
    return {
        "files": [os.path.relpath(p, run_dir) for p in files],
        "durations": durations, "failed": failed, "per_sec": per_sec,
        "per_sec_failed": per_sec_failed, "tls": tls, "tls_all": tls_all,
        "counters": counters, "trends": trends,
    }


def write_requests_per_second(run_dir, k6):
    path = os.path.join(run_dir, "requests_per_second.csv")
    with open(path, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["time_utc", "epoch", "name", "count", "failed", "p50_ms", "p95_ms", "p99_ms", "max_ms"])
        for ts in sorted(k6["per_sec"]):
            for name, vals in sorted(k6["per_sec"][ts].items()):
                s = stats(vals)
                w.writerow([iso(ts), ts, name, s["count"], k6["per_sec_failed"][ts].get(name, 0),
                            s["p50"], s["p95"], s["p99"], s["max"]])


def write_tls_per_second(run_dir, k6):
    path = os.path.join(run_dir, "tls_per_second.csv")
    with open(path, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["time_utc", "epoch", "handshakes", "p50_ms", "p95_ms", "max_ms"])
        for ts in sorted(k6["tls"]):
            s = stats(k6["tls"][ts])
            w.writerow([iso(ts), ts, s["count"], s["p50"], s["p95"], s["max"]])


def parse_pidstat(path, sum_per_time=False):
    """pidstat -h: '# Time UID PID %usr ... %CPU ... Command' header, then rows.
    Returns {time: {'cpu': %, 'rss_kb': kB}} (summed over PIDs if asked)."""
    out = {}
    if not os.path.exists(path):
        return out
    header = None
    with open(path) as f:
        for line in f:
            parts = line.split()
            if not parts:
                continue
            if parts[0] == "#":
                header = parts[1:]
                continue
            if header is None or len(parts) != len(header) or ":" not in parts[0]:
                continue
            row = dict(zip(header, parts))
            try:
                cpu = float(row["%CPU"])
            except (KeyError, ValueError):
                continue
            rss = float(row.get("RSS", 0) or 0)
            t = row["Time"]
            if sum_per_time and t in out:
                out[t]["cpu"] += cpu
                out[t]["rss_kb"] += rss
            else:
                out[t] = {"cpu": cpu, "rss_kb": rss}
    return out


def parse_mpstat(path):
    """mpstat 1: rows '<time> all %usr ... %idle' -> {time: busy %}"""
    out = {}
    if not os.path.exists(path):
        return out
    with open(path) as f:
        for line in f:
            parts = line.split()
            if len(parts) > 3 and parts[1] == "all" and ":" in parts[0]:
                try:
                    out[parts[0]] = round(100 - float(parts[-1]), 2)
                except ValueError:
                    pass
    return out


def write_cpu(run_dir):
    series = {
        "server_cpu_pct": parse_pidstat(os.path.join(run_dir, "server", "server.pidstat")),
        "postgres_cpu_pct": parse_pidstat(os.path.join(run_dir, "server", "postgres.pidstat"), sum_per_time=True),
    }
    server_box = parse_mpstat(os.path.join(run_dir, "server", "server-box.mpstat"))
    gens = {}
    for gdir in sorted(glob.glob(os.path.join(run_dir, "generator*"))):
        g = os.path.basename(gdir)
        for p in glob.glob(os.path.join(gdir, "shard*.k6.pidstat")):
            gens[f"{g}_k6_cpu_pct"] = parse_pidstat(p, sum_per_time=True)
        for p in glob.glob(os.path.join(gdir, "shard*.box.mpstat")):
            gens[f"{g}_box_busy_pct"] = parse_mpstat(p)
    times = set(server_box)
    for data in list(series.values()) + list(gens.values()):
        times |= set(data)
    times = sorted(times)
    cols = ["server_cpu_pct", "server_rss_mb", "postgres_cpu_pct", "server_box_busy_pct"] + sorted(gens)
    peaks = defaultdict(float)
    with open(os.path.join(run_dir, "cpu_per_second.csv"), "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["time"] + cols)
        for t in times:
            srv = series["server_cpu_pct"].get(t)
            row = {
                "server_cpu_pct": srv["cpu"] if srv else "",
                "server_rss_mb": round(srv["rss_kb"] / 1024, 1) if srv else "",
                "postgres_cpu_pct": round(series["postgres_cpu_pct"][t]["cpu"], 2) if t in series["postgres_cpu_pct"] else "",
                "server_box_busy_pct": server_box.get(t, ""),
            }
            for g, data in gens.items():
                v = data.get(t, "")
                row[g] = v["cpu"] if isinstance(v, dict) else v
            for c in cols:
                if row.get(c, "") != "":
                    peaks[c] = max(peaks[c], float(row[c]))
            w.writerow([t] + [row.get(c, "") for c in cols])
    return dict(peaks)


def write_settlement(run_dir):
    """Settlement throughput from the exported BET journal rows (exact)."""
    path = os.path.join(run_dir, "server", "bet_journal.csv")
    if not os.path.exists(path):
        return None
    per_sec = defaultdict(int)
    with open(path, newline="") as f:
        for row in csv.DictReader(f):
            # Two journal rows per bet (stake < 0, winnings >= 0): count one
            if row.get("kind") != "winnings":
                continue
            t = row["created_at"][:19]
            per_sec[t] += 1
    if not per_sec:
        return None
    total = 0
    with open(os.path.join(run_dir, "settlement_per_second.csv"), "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["time_utc", "bets_settled", "cumulative"])
        for t in sorted(per_sec):
            total += per_sec[t]
            w.writerow([t, per_sec[t], total])
    first, last = min(per_sec), max(per_sec)
    span = (datetime.fromisoformat(last) - datetime.fromisoformat(first)).total_seconds() + 1
    return {"bets": total, "first": first, "last": last, "seconds": span,
            "bets_per_second": round(total / span, 1), "peak_per_second": max(per_sec.values())}


def main():
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(1)
    run_dir = sys.argv[1]
    k6 = read_k6(run_dir)
    write_requests_per_second(run_dir, k6)
    write_tls_per_second(run_dir, k6)
    cpu_peaks = write_cpu(run_dir)
    settlement = write_settlement(run_dir)

    per_name = {}
    for name, vals in sorted(k6["durations"].items()):
        s = stats(vals)
        s["failed"] = k6["failed"].get(name, 0)
        per_name[name] = s
    all_ms = [v for vals in k6["durations"].values() for v in vals]
    tls_per_sec = [len(v) for v in k6["tls"].values()]
    summary = {
        "k6_files": k6["files"],
        "requests": dict(stats(all_ms), failed=sum(k6["failed"].values())),
        "per_name_ms": per_name,
        "tls_handshake_ms": dict(stats(k6["tls_all"]), peak_per_second=max(tls_per_sec) if tls_per_sec else 0),
        "counters": {k: int(v) for k, v in sorted(k6["counters"].items())},
        "trends": {k: stats(v) for k, v in sorted(k6["trends"].items())},
        "cpu_peaks_pct": cpu_peaks,
        "settlement": settlement,
    }
    with open(os.path.join(run_dir, "summary.json"), "w") as f:
        json.dump(summary, f, indent=2)

    r = summary["requests"]
    print(f"requests {r['count']} failed {r['failed']}  p50 {r.get('p50')}ms p95 {r.get('p95')}ms p99 {r.get('p99')}ms")
    print(f"{'name':32} {'count':>7} {'fail':>5} {'p50':>9} {'p95':>9} {'p99':>9} {'max':>9}")
    for name, s in per_name.items():
        print(f"{name[:32]:32} {s['count']:>7} {s['failed']:>5} {s['p50']:>9} {s['p95']:>9} {s['p99']:>9} {s['max']:>9}")
    t = summary["tls_handshake_ms"]
    print(f"TLS handshakes {t['count']} (peak {t['peak_per_second']}/s) p50 {t.get('p50')}ms p95 {t.get('p95')}ms p99 {t.get('p99')}ms")
    print("counters:", ", ".join(f"{k}={v}" for k, v in summary["counters"].items()))
    if "settlement_duration" in summary["trends"]:
        print(f"settlement request: {summary['trends']['settlement_duration'].get('max')} ms")
    if settlement:
        print(f"settlement: {settlement['bets']} bets in {settlement['seconds']:.0f}s "
              f"({settlement['bets_per_second']}/s, peak {settlement['peak_per_second']}/s)")
    print("CPU peaks %:", ", ".join(f"{k}={v:.0f}" for k, v in sorted(cpu_peaks.items())))


if __name__ == "__main__":
    main()
