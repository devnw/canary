#!/usr/bin/env python3
"""Generate a robust markdown coverage summary + link to full HTML report.

Creates site_src/coverage.md and copies cover.html (raw Go report).

Improvements over previous approach:
* Avoids brittle regex extraction + JS manipulation of cover.html body.
* Provides a concise per-file coverage table (computed from cover.out),
  with colored bars for quick visual scanning.
* Offers both a link and optionally an embedded iframe to the original
  detailed HTML (toggle via INLINE_IFRAME flag if desired later).
"""
from __future__ import annotations

import pathlib
import sys
from collections import defaultdict

cover_profile = pathlib.Path("cover.out")
cover_html_path = pathlib.Path("cover.html")
out_dir = pathlib.Path("site_src")
out_dir.mkdir(parents=True, exist_ok=True)
md_path = out_dir / "coverage.md"
raw_copy = out_dir / "cover.html"


if not cover_profile.exists():
    md_path.write_text(
        "# Coverage Report\n\nNo coverage profile (cover.out) produced.\n",
        encoding="utf-8",
    )
    sys.exit(0)

# Parse Go cover profile
# Format: file:line.col,line.col statements count
totals = defaultdict(lambda: {"stmts": 0, "covered": 0})
try:
    with cover_profile.open(encoding="utf-8") as f:
        first = f.readline()
        if not first.startswith("mode:"):
            # not a cover profile; fallback
            raise ValueError("unexpected cover profile header")
        for line in f:
            line = line.strip()
            if not line:
                continue
            # path:pos statements count
            # ex: github.com/me/proj/pkg/file.go:34.2,36.10 5 1
            try:
                loc_part, stmts_s, count_s = line.rsplit(" ", 2)
                file_path = loc_part.split(":", 1)[0]
                stmts = int(stmts_s)
                cnt = int(count_s)
            except Exception:
                continue
            rec = totals[file_path]
            rec["stmts"] += stmts
            if cnt > 0:
                rec["covered"] += stmts
except Exception as e:  # pragma: no cover - defensive fallback
    error_msg = f"# Coverage Report\n\nFailed to parse coverage profile: {e}\n"
    md_path.write_text(error_msg, encoding="utf-8")
    # Still copy HTML if present
    if cover_html_path.exists():
        raw_copy.write_text(
            cover_html_path.read_text(encoding="utf-8"), encoding="utf-8"
        )
    sys.exit(0)

rows = []
total_stmts = 0
total_cov = 0
for file_path, rec in sorted(totals.items()):
    stmts = rec["stmts"]
    cov = rec["covered"]
    pct = (cov / stmts * 100) if stmts else 0.0
    rows.append((file_path, stmts, cov, pct))
    total_stmts += stmts
    total_cov += cov

overall_pct = (total_cov / total_stmts * 100) if total_stmts else 0.0


def bar(p: float) -> str:
    # Simple classless bar using inline style
    # to avoid theme conflicts
    width = 120  # px
    color = (
        "#d9534f"
        if p < 50
        else "#f0ad4e" if p < 70 else "#5bc0de" if p < 80 else "#5cb85c"
    )
    outer = (
        '<div style="background:#eee;border:1px solid #ccc;'
        f'width:{width}px;height:10px;position:relative">'
    )
    inner = (
        f'<div style="background:{color};height:100%;width:{p:.2f}%;'
        'max-width:100%"></div>'
    )
    return outer + inner + "</div>"


table_lines = [
    "| File | Stmts | Covered | % | Graph |",
    "|------|-------|---------|----|-------|",
]
for file_path, stmts, cov, pct in rows:
    rel = file_path  # Already relative-ish from tool; we leave it as-is.
    bar_html = bar(pct)
    line = f"| `{rel}` | {stmts} | {cov} | {pct:.2f}% | {bar_html} |"
    table_lines.append(line)

body = [
    "# Coverage Report",
    "",
    (
        f"Overall statements coverage: **{overall_pct:.2f}%** "
        f"({total_cov}/{total_stmts})"
    ),
    "",
    "Generated from `cover.out`. See the full annotated HTML report:",
    "",
    "[Open full coverage report](cover.html)",
    "",
    "<details><summary>Embedded full report (may be large)</summary>",
    '<iframe src="cover.html" style="width:100%;height:70vh;border:0" '
    'title="Coverage HTML"></iframe>',
    "</details>",
    "",
    "## Per-file Coverage",
    "",
    *table_lines,
    "",
    "_Bar colors: <span style='color:#d9534f'>low</span>, "
    "<span style='color:#f0ad4e'>moderate</span>, "
    "<span style='color:#5bc0de'>good</span>, "
    "<span style='color:#5cb85c'>high</span>._",
]

md_path.write_text("\n".join(body), encoding="utf-8")

if cover_html_path.exists():
    raw_copy.write_text(
        cover_html_path.read_text(encoding="utf-8"),
        encoding="utf-8",
    )
else:  # In rare cases missing
    raw_copy.write_text(
        "<html><body><p>No cover.html generated.</p></body></html>",
        encoding="utf-8",
    )
