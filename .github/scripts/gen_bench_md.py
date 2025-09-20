#!/usr/bin/env python3
"""Generate benchmark markdown page (bench.md) with optional charts.

Logic:
1. Attempt to fetch bench-data branch and add a worktree (bench_data_wt).
2. If summary.json exists there, copy summary + per-benchmark JSON into
   site_src/bench/, copy bench.js asset, and write a rich bench.md.
3. Otherwise create a placeholder bench.md explaining absence of history.

This keeps YAML in the workflow simple and mirrors coverage generation.
"""
from __future__ import annotations

import json
import pathlib
import shutil
import subprocess
import sys

ROOT = pathlib.Path.cwd()
SITE_SRC = ROOT / "site_src"
BENCH_SRC = SITE_SRC / "bench"
BENCH_MD = SITE_SRC / "bench.md"
ASSET_JS = ROOT / ".github" / "assets" / "bench.js"
WORKTREE = ROOT / "bench_data_wt"


def run(cmd: list[str]) -> subprocess.CompletedProcess:
    return subprocess.run(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True
    )


def ensure_dirs():
    SITE_SRC.mkdir(parents=True, exist_ok=True)


def fetch_worktree() -> bool:
    fetch = run(
        [
            "git",
            "fetch",
            "origin",
            "bench-data:refs/remotes/origin/bench-data",
        ]
    )
    if fetch.returncode != 0:
        return False
    # Add worktree (force if path exists but is stale)
    if WORKTREE.exists():
        # If it's a git dir already, reuse; else remove.
        if not (WORKTREE / ".git").exists():
            shutil.rmtree(WORKTREE)
    add = run(
        [
            "git",
            "worktree",
            "add",
            "-f",
            str(WORKTREE),
            "origin/bench-data",
        ]
    )
    return add.returncode == 0


def copy_history() -> bool:
    summary_src = WORKTREE / "bench" / "summary.json"
    if not summary_src.exists():
        return False
    BENCH_SRC.mkdir(parents=True, exist_ok=True)
    data_dir_src = summary_src.parent / "data"
    # Copy summary
    shutil.copy2(summary_src, BENCH_SRC / "summary.json")
    # Copy per-benchmark series
    dest_data = BENCH_SRC / "data"
    dest_data.mkdir(exist_ok=True)
    if data_dir_src.exists():
        for p in data_dir_src.glob("*.json"):
            shutil.copy2(p, dest_data / p.name)
    return True


def write_placeholder(msg: str):
    BENCH_MD.write_text(f"# Benchmarks\n\n{msg}\n", encoding="utf-8")


def write_rich_page():
    if not (BENCH_SRC / "summary.json").exists():
        write_placeholder("_No benchmark history yet._")
        return
    # Copy JS asset
    if ASSET_JS.exists():
        dest_js = BENCH_SRC / "bench.js"
        shutil.copy2(ASSET_JS, dest_js)
    content = (
        "# Benchmarks\n\n"
        "Benchmark performance over time.\n\n"
        "Raw summary: [summary.json](summary.json)\n\n"
        '<div id="bench-charts">Loading benchmark history...</div>\n'
        '<script src="bench.js"></script>\n'
    )
    BENCH_MD.write_text(content, encoding="utf-8")


def main():
    ensure_dirs()
    if not fetch_worktree():
        write_placeholder("_No benchmark history branch found yet._")
        return 0
    copied = copy_history()
    if not copied:
        write_placeholder("_No benchmark history yet._")
        return 0
    # Optionally validate JSON structure (best effort)
    try:
        summary_path = BENCH_SRC / "summary.json"
        summary = json.loads(summary_path.read_text(encoding="utf-8"))
        if not summary.get("benchmarks"):
            write_placeholder("_Benchmark summary empty._")
            return 0
    except Exception:
        write_placeholder("_Benchmark summary unreadable._")
        return 0
    write_rich_page()
    return 0


if __name__ == "__main__":  # pragma: no cover - simple script
    sys.exit(main())
