#!/usr/bin/env python3
"""Update benchmark history on bench-data branch.

Reads bench.out (from `go test -run=^$ -bench=. -benchmem`), updates
per-benchmark JSON time‑series under bench/data/*.json plus an
aggregated bench/summary.json.

Environment:
  BENCH_OUT   Path to bench output file (default: bench.out)
  GITHUB_REPO owner/repo (github.repository)
  GITHUB_TOKEN token with contents write permission
  GITHUB_SERVER_URL e.g. https://github.com
  GITHUB_SHA commit being processed
  GITHUB_REF_NAME branch/ref name
"""
from __future__ import annotations

import datetime as _dt
import json
import os
import pathlib
import re
import subprocess
import tempfile

BENCH_RE = re.compile(
    r"^(Benchmark\S+)\s+(\d+)\s+(\d+) ns/op"
    r"(?:\s+(\d+) B/op)?(?:\s+(\d+) allocs/op)?"
)

# Canonical GitHub Actions bot noreply email
BOT_ACTIONS_EMAIL = "41898282+github-actions[bot]@users.noreply.github.com"


def run(cmd, cwd: pathlib.Path | None = None, check: bool = True):
    r = subprocess.run(
        cmd,
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if check and r.returncode != 0:
        msg = (
            "::error::Command failed: {cmd}\n"  # noqa: E501
            "STDOUT:\n{out}\nSTDERR:\n{err}".format(
                cmd=" ".join(cmd),
                out=r.stdout,
                err=r.stderr,
            )
        )
        print(msg)
        raise SystemExit(r.returncode)
    return r


def parse_bench(lines):
    for line in lines:
        m = BENCH_RE.match(line.strip())
        if not m:
            continue
        name, iters, ns, b_op, allocs = m.groups()
        rec = {
            # Use timezone-aware UTC timestamp (avoid deprecated utcnow)
            "timestamp": _dt.datetime.now(_dt.timezone.utc)
            .isoformat()
            .replace("+00:00", "Z"),
            "iterations": int(iters),
            "ns_per_op": int(ns),
        }
        if b_op:
            rec["bytes_per_op"] = int(b_op)
        if allocs:
            rec["allocs_per_op"] = int(allocs)
        yield name, rec


def sanitize(name: str) -> str:
    return re.sub(r"[^A-Za-z0-9_.-]+", "_", name)


def main():
    bench_out = pathlib.Path(os.environ.get("BENCH_OUT", "bench.out"))
    if not bench_out.exists():
        print("::warning::bench.out not found; skipping bench history update")
        return

    repo = os.environ.get("GITHUB_REPO")
    token = os.environ.get("GITHUB_TOKEN")
    server = os.environ.get("GITHUB_SERVER_URL", "https://github.com")
    sha = os.environ.get("GITHUB_SHA", "")
    ref_name = os.environ.get("GITHUB_REF_NAME", "")
    if not (repo and token):
        print("::warning::Missing GITHUB_REPO or GITHUB_TOKEN; skipping push")
        return

    lines = bench_out.read_text().splitlines()
    series_index = {}
    for name, rec in parse_bench(lines):
        series_index.setdefault(name, []).append(rec)

    if not series_index:
        print("No benchmarks parsed; nothing to do")
        return

    branch = "bench-data"
    with tempfile.TemporaryDirectory() as td:
        work = pathlib.Path(td) / "repo"
        run(["git", "init", work.as_posix()])
        # Configure git identity (needed in ephemeral CI environments)
        actor = os.environ.get("GITHUB_ACTOR", "github-actions[bot]")
        user_name = os.environ.get("GIT_AUTHOR_NAME", actor)
        # Resolve email (env override -> actor -> noreply)
        user_email = os.environ.get(
            "GIT_AUTHOR_EMAIL",
            os.environ.get(
                "GIT_COMMITTER_EMAIL",
                f"{actor}@users.noreply.github.com",
            ),
        )
        # Normalize if actor is actions bot
        if actor == "github-actions[bot]":
            user_email = BOT_ACTIONS_EMAIL
        run(["git", "config", "user.name", user_name], cwd=work)
        run(["git", "config", "user.email", user_email], cwd=work)
        origin_url = "https://x-access-token:{tok}@{host}/{repo}.git".format(
            tok=token,
            host=server.split("://", 1)[1],
            repo=repo,
        )
        run(["git", "remote", "add", "origin", origin_url], cwd=work)
        fetch = run(
            ["git", "fetch", "--depth", "1", "origin", branch],
            cwd=work,
            check=False,
        )
        if fetch.returncode == 0:
            run(
                ["git", "checkout", "-B", branch, f"origin/{branch}"],
                cwd=work,
            )
        else:
            run(["git", "checkout", "-B", branch], cwd=work)

        bench_dir = work / "bench" / "data"
        bench_dir.mkdir(parents=True, exist_ok=True)

        summary = {"benchmarks": []}
        for full_name, new_recs in series_index.items():
            fname = sanitize(full_name) + ".json"
            path = bench_dir / fname
            if path.exists():
                try:
                    data = json.loads(path.read_text())
                except Exception:
                    data = []
            else:
                data = []
            data.extend(new_recs)
            path.write_text(json.dumps(data, indent=2) + "\n")
            summary["benchmarks"].append({"name": full_name, "file": fname})

        summary["benchmarks"].sort(key=lambda x: x["name"].lower())
        summary_path = work / "bench" / "summary.json"
        summary_path.write_text(json.dumps(summary, indent=2) + "\n")
        run(["git", "add", "bench"], cwd=work)
        diff = run(
            ["git", "diff", "--cached", "--quiet"],
            cwd=work,
            check=False,
        )
        if diff.returncode != 0:
            msg = "bench: update for {sh} ({ref})".format(
                sh=sha[:12],
                ref=ref_name,
            )
            run(["git", "commit", "-m", msg], cwd=work)
            run(["git", "push", "origin", branch], cwd=work)
            print("Benchmark data updated & pushed")
        else:
            print("No benchmark changes to commit")


if __name__ == "__main__":
    main()
