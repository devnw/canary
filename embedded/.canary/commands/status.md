## StatusCmd (Implementation Progress for a Requirement)

```yaml
---
description: Show implementation progress for a specific CANARY requirement with verified counts, percentages, and actionable next steps (strict, no-mock/no-simulate)
command: StatusCmd
version: 2.3
subcommands: [status]
outputs:
  - human_text: STDOUT (concise operator view; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  req_id_regex: '^[A-Z]{4,}-[A-Za-z]+-[0-9]{3}$'   # adjust to your canonical pattern
  progress_bar_width: 30
  completion_status_threshold: 1.0   # 100% = complete
---
```

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into: `req_id` (required), flags: `--db <path>`, `--no-color`, `--json` (machine‑readable only), `--width <int>` (progress bar width; optional).

### 2) Preconditions & Resolution

1. **REQ‑ID Gate:** `req_id` must match `req_id_regex`; else `ERROR_REQ_FORMAT(req_id)`.
2. **DB Gate:** database must exist/readable at `db_path`; else `ERROR_DB_MISSING(path)` with remediation (`canary index`).
3. **Command Availability:** ensure `canary status` exists; else `ERROR_CANARY_STATUS_UNAVAILABLE()`.
4. **Zero‑Data Handling:** if **no tokens** found for `req_id` → `ERROR_NO_TOKENS_FOUND(req_id)` (advise `/canary.specify` or verify `req_id`).

### 3) Planning & Parallelism

Create a **Work DAG**; run independent steps concurrently; **join** before final aggregation. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Query (I/O):**

  ```bash
  canary status <REQ-ID> --db <path> [--no-color] [--json]
  ```

  Collect raw rows: `feature, aspect, status, file, line, test?, bench?, owner?, priority?, updated`.
* **CG‑2 Aggregate:** compute totals, distributions, percentages, and **progress bar** string.
* **CG‑3 Analyze:** list **IMPL needs tests**, **STUB not started**, **TESTED missing BENCH**; derive **next steps**.

### 4) Behavior (must do; never simulate)

* **Run the real command**; do **not** fabricate counts, files, or lines.
* **Math (authoritative):**

  * `total = BENCHED + TESTED + IMPL + STUB` (ignore MISSING unless tool returns it; if returned, include in totals bucket `MISSING` and adjust denominator accordingly).
  * `completed = BENCHED + TESTED`
  * `in_progress = IMPL`
  * `not_started = STUB`
  * `pct_completed = completed / max(total,1)`
  * `pct_in_progress = in_progress / max(total,1)`
  * `pct_not_started = not_started / max(total,1)`
  * **Completion flag:** `is_complete = (pct_completed >= completion_status_threshold)`
* **Progress bar (ASCII):** width = `--width || defaults.progress_bar_width`;
  `filled = round(pct_completed * width)`; bar = `'[' + '='*filled + '-'*(width-filled) + ']'`.
* **Navigation:** include `file` and `line` for incomplete items; if `line` unknown, set `null` (do **not** invent).
* **Operator UX:** suppress ANSI when parsing (`--no-color`); HUMAN_TEXT may include emojis.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit snapshot when **context ≥70%**, after **query**, and after **aggregate**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"status|query|aggregate",
  "f":[["<db_path>",1,1]],
  "k":["req:<REQ-ID>","total:<N>","completed_pct:<0..1>","width:<W>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["present results","offer next steps"]
}'
```

*Compact keys + file+line spans minimize tokens while capturing essentials; consistent section delimiting aids downstream parsing.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** Use structured outputs for reliable automation. 

**A. HUMAN_TEXT (optional)**
Begin with `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Recommended contents: title with REQ‑ID, **progress bar** + `%`, totals, status breakdown, incomplete items (with file:line), and **Next Steps**.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "req_id": "{{.ReqID}}-<ASPECT>-API-102",
  "feature": "UserList",
  "db_path": ".canary/canary.db",
  "progress": {
    "bar": "[====================----------]",
    "width": 30,
    "percent_completed": 0.75
  },
  "totals": {
    "tokens": 20,
    "by_status": { "BENCHED": 5, "TESTED": 10, "IMPL": 3, "STUB": 2, "MISSING": 0 }
  },
  "percentages": {
    "completed": 0.75,
    "in_progress": 0.15,
    "not_started": 0.10
  },
  "incomplete": {
    "impl_needs_tests": [
      { "feature": "UserList", "file": "src/api/users.go", "line": 45 },
      { "feature": "DataFilter", "file": "internal/filter/engine.go", "line": 102 },
      { "feature": "QueryBuilder", "file": "internal/db/queries.go", "line": 78 }
    ],
    "stub_not_started": [
      { "feature": "AdvancedFilters", "file": "internal/filter/advanced.go", "line": 23 },
      { "feature": "CacheLayer", "file": "internal/cache/layer.go", "line": 15 }
    ],
    "tested_without_bench": [
      { "feature": "X", "file": "path/to/file.go", "line": null }
    ]
  },
  "is_complete": false,
  "recommendations": [
    "Add tests for IMPL tokens",
    "Plan implementation for STUB tokens using /canary.plan {{.ReqID}}-<ASPECT>-API-102",
    "Consider adding benchmarks to performance-critical TESTED features"
  ],
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **REQ‑ID Gate:** matches `req_id_regex`.
* **DB Gate:** DB reachable at `db_path`.
* **Counting Gate:** `totals.tokens == Σ by_status` and equals the number of items across buckets.
* **Math Gate:** percentages recompute to 1.0 within rounding tolerance (`±0.01`).
* **Schema Gate:** JSON conforms; field names/types exact.
* **Navigation Gate:** `file` present; `line` integer or `null` (never fabricated).

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_REQ_FORMAT(req_id)`
* `ERROR_DB_MISSING(path)` → suggest `canary index`
* `ERROR_CANARY_STATUS_UNAVAILABLE()`
* `ERROR_CANARY_STATUS_FAILED(exit_code,stderr_excerpt)`
* `ERROR_NO_TOKENS_FOUND(req_id)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real `canary status` executed; **no simulated/mocked** results.
* Progress math correct; bar matches computed percent and width.
* Incomplete lists include concrete file:line refs when available.
* HUMAN_TEXT (if produced) concise and consistent with JSON.
* CANARY snapshot(s) emitted when required.
* JSON returned **without** code fences; schema exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## Implementation Status for {{.ReqID}}-<ASPECT>-API-102

Progress: [========================--------] 75%
**Total:** 20 tokens • **Completed:** 15 (75%)

**In Progress:** IMPL 3 (15%) • **Not Started:** STUB 2 (10%)

**Status Breakdown**
- BENCHED: 5 (25%) ✅
- TESTED: 10 (50%) ✅
- IMPL: 3 (15%) ⚠️
- STUB: 2 (10%) ⚠️

**Incomplete Work**
IMPL → tests needed:
- UserList (src/api/users.go:45)
- DataFilter (internal/filter/engine.go:102)
- QueryBuilder (internal/db/queries.go:78)

STUB → not started:
- AdvancedFilters (internal/filter/advanced.go:23)
- CacheLayer (internal/cache/layer.go:15)

**Next Steps**
1) Add tests for 3 IMPL features
2) /canary.plan {{.ReqID}}-<ASPECT>-API-102 for 2 STUB features
3) Consider adding benchmarks for performance‑critical features
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic, parseable outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers enable reliable parsing and CI checks. 
* **Section‑delimited structure:** clear inputs → gates → behavior → outputs improves maintainability and downstream automation. 
* **Parallel pipeline:** explicit DAG + concurrency groups + join points for fast query/aggregate flows over large sets. 
* **No‑mock/no‑simulate:** codified guarantees ensure the command surfaces **real** counts and file:line evidence only. 

### Assumptions & Risks

* `canary status --json` returns fields listed; if some (e.g., `owner`, `priority`) are missing, set `null`—never invent.
* Some tools include `MISSING` bucket; include and reflect in totals if present.
* Progress bars are visual aids only; JSON is the source of truth for automation.

### Targeted questions (for fit)

1. Confirm canonical **REQ‑ID** regex and whether `MISSING` should count toward `total`.
2. Do you want `--width` exposed to operators (default 30)?
3. Should we include **age since UPDATED** and **owner load** in `SUMMARY_JSON`?
4. Keep CANARY snapshot threshold at **70%** context usage?
