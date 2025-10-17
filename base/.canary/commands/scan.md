## ScanCmd (Scan Tokens & Generate Status Reports)

```yaml
---
description: Scan the codebase for CANARY tokens and produce verifiable status reports with summaries, action items, and trend deltas (strict, no-mock/no-simulate)
command: ScanCmd
version: 2.4
subcommands: [scan]
outputs:
  - human_text: STDOUT (concise operator report; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
  - reports: files written to disk (status.json, status.csv) by the real scanner
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  root: .
  out_json: status.json
  out_csv: status.csv
  skip: ".git|node_modules|vendor|bin|dist|build"
  strict: false                # when true, enforce 30-day staleness check
  verify_path: null            # GAP_ANALYSIS.md path (optional)
  trend_prev: null             # previous status.json for trend deltas (optional)
---
```

<!-- CANARY: REQ=CBIN-111; FEATURE="ScanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse flags: `--root <path>`, `--out <path>`, `--csv <path>`, `--skip <regex>`, `--strict`, `--verify <GAP_ANALYSIS.md>`, `--trend-prev <status.json>`.
  If not provided, use **defaults** above.

### 2) Preconditions & Resolution

1. **Root gate:** `root` must exist and be readable; else `ERROR_ROOT_NOT_FOUND(path)`.
2. **Output paths:** ensure parent dirs exist or create; else `ERROR_OUT_PATH(path)`.
3. **Skip pattern:** validate regex; else `ERROR_SKIP_REGEX_INVALID(expr)`.
4. **Verify path (optional):** if provided but missing → `ERROR_VERIFY_MISSING(path)`.
5. **Trend file (optional):** if provided but unreadable → ignore with warning in JSON (`trend_usable=false`).

### 3) Planning & Parallelism

Build a **Work DAG** and run independent steps concurrently; **join** before final aggregation. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Scan (I/O bound):**

  ```bash
  canary scan --root <root> --out <out_json> --csv <out_csv> --skip "<skip>" {--strict} {--verify <verify_path>}
  ```
* **CG‑2 Parse/Compute:** Read `<out_json>`; compute totals by `STATUS` (STUB/IMPL/TESTED/BENCHED), aspect coverage, unique REQs, stale set (if strict), coverage KPIs.
* **CG‑3 Trend (optional):** If `trend_prev`, read it and compute deltas (coverage %, totals).
* **CG‑4 Action Items:** derive lists: stale, IMPL‑needs‑tests, TESTED‑needs‑bench, missing OWNER.
* **CG‑5 Assemble Outputs:** HUMAN_TEXT (optional) + SUMMARY_JSON (strict).

### 4) Behavior (must do; never simulate)

* **Run the real scanner**; do **not** fabricate counts, paths, or lines.
* **Respect skip** pattern and defaults; show them in outputs for traceability.
* **Strict staleness:** default threshold = 30 days; mark stale tokens; include counts.
* **Verification:** if `--verify` is set, report pass/fail for GAP_ANALYSIS.md claims (from scanner output).
* **Links:** surface `out_json` and `out_csv` paths in outputs; do not “fake” files.
* **Trend:** only compute deltas when a valid `trend_prev` is provided; else indicate `trend=false`.
* **Size/clarity:** keep HUMAN_TEXT compact; put all details in JSON.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **scan**, and after **aggregate**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"scan|aggregate",
  "f":[["<root>",1,1],["<out_json>",1,1],["<trend_prev?>",1,1]],
  "k":["strict:<bool>","reqs:<N>","tokens:<T>","tested+benched:<KPI%>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["publish reports","suggest next actions"]
}'
```

*Compact keys minimize tokens while preserving filenames, key facts, false‑positives, issues, and next steps.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** Use structured outputs for reliable downstream automation. 

**A. HUMAN_TEXT (optional)**
Begin with `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Recommended contents: scan date, totals, percent bars or emojis, key action items, and links to `status.json` / `status.csv`.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "params": {
    "root": ".",
    "out_json": "status.json",
    "out_csv": "status.csv",
    "skip": ".git|node_modules|vendor|bin|dist|build",
    "strict": false,
    "verify_path": null,
    "trend_prev": null
  },
  "totals": {
    "requirements": 0,
    "tokens": 0,
    "by_status": { "BENCHED": 0, "TESTED": 0, "IMPL": 0, "STUB": 0, "MISSING": 0 }
  },
  "coverage": {
    "test_pct": 0.0,
    "bench_pct": 0.0
  },
  "by_aspect": [
    { "aspect": "API", "tokens": 0 },
    { "aspect": "CLI", "tokens": 0 }
  ],
  "stale": { "count": 0, "items": ["{{.ReqID}}-<ASPECT>-..."] },
  "unique_requirements": [],
  "verification": { "enabled": false, "gap_pass": null, "details": null },
  "action_items": {
    "update_stale": ["{{.ReqID}}-<ASPECT>-..."],
    "impl_needs_tests": ["{{.ReqID}}-<ASPECT>-..."],
    "tested_needs_bench": ["{{.ReqID}}-<ASPECT>-..."],
    "missing_owner": ["{{.ReqID}}-<ASPECT>-..."]
  },
  "reports": { "json": "status.json", "csv": "status.csv" },
  "trend": {
    "enabled": false,
    "prev_path": null,
    "deltas": {
      "test_pct": 0.0,
      "bench_pct": 0.0,
      "tokens": 0,
      "requirements": 0
    }
  },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **Root Gate:** path exists and readable.
* **Skip Gate:** regex valid and applied.
* **Strict Staleness Gate:** when `strict=true`, stale tokens identified with 30‑day policy.
* **Counting Gate:** `totals.tokens` equals sum of status buckets; aspect sums consistent.
* **Coverage Gate:** `test_pct = (TESTED+BENCHED)/total`; `bench_pct = BENCHED/total` (handle divide‑by‑zero).
* **Trend Gate:** if `trend.enabled`, deltas computed and numeric.
* **Schema Gate:** JSON conforms exactly; field names/types exact.
* **Verify Gate:** if `verify_path` provided, include pass/fail with details.

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_ROOT_NOT_FOUND(path)`
* `ERROR_OUT_PATH(path)`
* `ERROR_SKIP_REGEX_INVALID(expr)`
* `ERROR_VERIFY_MISSING(path)`
* `ERROR_SCANNER_UNAVAILABLE()`
* `ERROR_SCANNER_FAILED(exit_code,stderr_excerpt)`
* `ERROR_STATUS_JSON_MISSING(path)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real `canary scan` executed; **no simulated/mocked** results.
* Defaults respected unless flags override; skip patterns applied.
* HUMAN_TEXT (if produced) concise and consistent with JSON.
* Links to `status.json`/`status.csv` are real paths.
* CANARY snapshot(s) emitted when required.
* JSON returned **without** code fences; schema exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## CANARY Token Scan Results
**Scan Date:** 2025-10-16 • **Root:** . • **Strict:** false

### Status Distribution
- BENCHED: 3 (30%) ✅
- TESTED: 4 (40%) ✅
- IMPL: 2 (20%) ⚠️
- STUB: 1 (10%) ⚠️
- MISSING: 0 (0%)

### Aspect Coverage
- API: 4 • CLI: 3 • Engine: 2 • Storage: 1

### Quality
- Test Coverage: 70% (TESTED+BENCHED)
- Benchmark Coverage: 30% (BENCHED)
- Stale Tokens: 2 (strict mode only)

**Reports:** status.json • status.csv

### Action Items
1) Update stale: {{.ReqID}}-<ASPECT>-…-001, {{.ReqID}}-<ASPECT>-…-004 → `canary scan --update-stale`
2) Add tests for IMPL: {{.ReqID}}-<ASPECT>-…-003, {{.ReqID}}-<ASPECT>-…-007
3) Add benchmarks for TESTED: {{.ReqID}}-<ASPECT>-…-002, {{.ReqID}}-<ASPECT>-…-005
=== HUMAN_TEXT END ===
```

### 11) Operator Guidance (safe defaults)

* **Automatic Execution:** If no args, run with defaults (`root=.`; standard skip).
* **Clear Visualization:** Prefer concise text in HUMAN_TEXT; put full data in JSON.
* **Actionable Output:** Always include commands to remediate findings.
* **Trend Analysis:** If `--trend-prev` present, compute deltas for coverage and totals.
* **Performance:** Treat scan as I/O bound; parallelize safely; avoid blocking joins. 

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers enable reliable parsing and CI checks. 
* **Section delimiting & structure:** clearer inputs → gates → behavior → outputs improves maintainability and UX. 
* **Parallel pipeline & trend deltas:** explicit DAG + concurrency groups + optional previous‑report comparison. 
* **No‑mock/no‑simulate:** codified as runtime guarantees; only real scanner outputs are used.

### Assumptions & Risks

* `canary scan` emits `status.json`/`status.csv` with the fields required for counts/coverage; if fields differ, map them carefully and document.
* Staleness threshold is 30 days when `strict`; adjust if your constitution specifies differently.
* Very large repos may require chunked scanning or tighter skip patterns.

### Targeted questions (for fit)

1. Confirm the exact JSON shape of `status.json` (field names for status, aspect, owner).
2. Should **strict** staleness threshold be configurable (e.g., `--stale-days N`)?
3. Any additional KPIs desired (e.g., OWNER coverage %, UPDATED age buckets)?
4. Keep CANARY snapshot threshold at **70%** context usage?
