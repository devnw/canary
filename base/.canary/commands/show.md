## ShowCmd (Display Tokens for a Requirement)

```yaml
---
description: Display all CANARY tokens for a specific requirement ID, with grouping, analysis, navigation, and strict machine-readable output (no-mock/no-simulate)
command: ShowCmd
version: 2.2
subcommands: [show]
outputs:
  - human_text: STDOUT (concise operator view; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  group_by: aspect     # values: aspect | status
  include_hidden: false
  req_id_regex: '^[A-Z]{4,}-[A-Za-z]+-[0-9]+$'   # adjust to your canonical pattern
---
```

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into: `req_id` (required), flags: `--group-by <aspect|status>`, `--json`, `--no-color`, `--db <path>`.

### 2) Preconditions & Resolution

1. **REQ‑ID Gate:** `req_id` must match `req_id_regex`; else `ERROR_REQ_FORMAT(req_id)`.
2. **DB Gate:** ensure DB at `db_path` exists/readable; else `ERROR_DB_MISSING(path)` with remediation (`canary index`).
3. **Command Availability:** ensure `canary show` exists; else `ERROR_CANARY_SHOW_UNAVAILABLE()`.
4. **Scope Policy:** hidden/spec/template paths excluded by default; if the CLI includes them unavoidably, mark `included_hidden=true` in output.

### 3) Planning & Parallelism

Build a **Work DAG**; run independent steps concurrently; **join** before final aggregation. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Query (I/O):**

  ```bash
  canary show <REQ-ID> --db <path> [--group-by <aspect|status>] [--no-color] [--json]
  ```

  Capture raw rows (file, line, feature, aspect, status, owner, priority, test, bench, updated, doc?).
* **CG‑2 Normalize:** sanitize/standardize fields, coerce types, ensure `line` is integer or `null`, dedupe tokens.
* **CG‑3 Group & Analyze:** group by requested dimension; compute totals & distributions; detect **IMPL without TEST=** and **TESTED without BENCH=**; prepare navigation hints (top files by token count).

### 4) Behavior (must do; never simulate)

* **Run the real command**; do **not** fabricate tokens, file paths, or line numbers.
* **Grouping:** default `aspect`; if `--group-by status`, regroup and set `group_by="status"`.
* **Navigation:** include `file` and `line` for every token; if `line` unknown, set `null` (do **not** invent).
* **Analysis:** compute token totals and per‑status counts; identify missing tests/benches; surface OWNER/PRIORITY.
* **Operator UX:** suppress ANSI when parsing (`--no-color`) but allow pretty HUMAN_TEXT when requested.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **query**, and after **aggregate**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"show|query|aggregate",
  "f":[["<db_path>",1,1]],
  "k":["req:<REQ-ID>","group:<aspect|status>","tokens:<N>","files:<M>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["present results","offer recommendations"]
}'
```

*Compact keys capture filenames, key facts, false‑positives, and next steps with minimal tokens.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** Use structured outputs for reliable downstream automation. 

**A. HUMAN_TEXT (optional)**
Begin with: `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Recommended contents: title with REQ‑ID + feature, grouped lists with emoji/status, totals, short analysis, and concrete next steps.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "req_id": "{{.ReqID}}-<ASPECT>-API-133",
  "feature": "UserAuthentication",
  "group_by": "aspect",
  "db_path": ".canary/canary.db",
  "included_hidden": false,
  "totals": {
    "tokens": 0,
    "files": 0,
    "by_status": { "BENCHED": 0, "TESTED": 0, "IMPL": 0, "STUB": 0 }
  },
  "groups": [
    {
      "key": "API",                 // or "TESTED" when grouping by status
      "items": [
        {
          "feature": "ValidationMiddleware",
          "aspect": "API",
          "status": "IMPL",
          "priority": 2,
          "owner": "api-team",
          "file": "src/api/middleware.go",
          "line": 45,
          "test": null,
          "bench": null,
          "updated": "2025-10-16"
        }
      ],
      "counts": { "items": 0, "tokens": 0 }
    }
  ],
  "analysis": {
    "impl_without_tests": [
      { "feature": "ValidationMiddleware", "file": "src/api/middleware.go", "line": 45 }
    ],
    "tested_without_bench": [
      { "feature": "SessionStore", "file": "internal/storage/session.go", "line": 67 }
    ],
    "primary_files": ["src/api/auth.go","internal/storage/session.go"]
  },
  "recommendations": [
    "Add TEST= for IMPL tokens (e.g., src/api/middleware.go:45)",
    "Add BENCH= for TESTED tokens lacking benchmarks",
    "/canary.plan {{.ReqID}}-<ASPECT>-API-133  # if STUB tokens present"
  ],
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **REQ‑ID Format Gate:** matches regex.
* **DB Gate:** DB reachable at `db_path`.
* **Group‑By Gate:** `group_by ∈ {aspect,status}`.
* **Counting Gate:** `totals.tokens == Σ group.counts.tokens == len(all items)`.
* **Schema Gate:** JSON conforms; field names/types exact.
* **Navigation Gate:** paths are strings; `line` integer or `null` (never fabricated).

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_REQ_FORMAT(req_id)`
* `ERROR_DB_MISSING(path)` → suggest `canary index`
* `ERROR_CANARY_SHOW_UNAVAILABLE()`
* `ERROR_CANARY_SHOW_FAILED(exit_code,stderr_excerpt)`
* `ERROR_REQ_NOT_FOUND(req_id)`
* `ERROR_NO_TOKENS_FOUND(req_id)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real `canary show` executed; **no simulated/mocked** results.
* Grouping correct for requested dimension; OWNER/PRIORITY surfaced when available.
* Totals and distributions consistent; analysis lists accurate.
* HUMAN_TEXT (if produced) concise and matches JSON.
* CANARY snapshot(s) emitted when required.
* JSON returned **without** code fences; schema exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## Tokens for {{.ReqID}}-<ASPECT>-API-133 — UserAuthentication

### API
📌 {{.ReqID}}-<ASPECT>-API-133 — UserAuthentication
   Status: TESTED | Priority: 1 | Owner: api-team
   Location: src/api/auth.go:25
   Test: TestCANARY_CB..._UserAuthentication

📌 {{.ReqID}}-<ASPECT>-API-133 — ValidationMiddleware
   Status: IMPL | Priority: 2
   Location: src/api/middleware.go:45

### Storage
📌 {{.ReqID}}-<ASPECT>-Storage-133 — SessionStore
   Status: BENCHED | Priority: 1 | Owner: backend-team
   Location: internal/storage/session.go:67
   Test: TestCANARY_CB..._SessionStore
   Bench: BenchCANARY_CB..._SessionStore

**Summary:** 3 tokens • BENCHED 1 • TESTED 1 • IMPL 1  
**Recommendations:** Add tests for IMPL token (src/api/middleware.go:45); keep BENCH for SessionStore up to date.
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers ensure reliable parsing and CI checks. 
* **Section delimiting & structure:** clear inputs → gates → behavior → outputs → examples improve maintainability and UX. 
* **Parallel pipeline:** explicit DAG + concurrency groups + join points for fast query/aggregate over large token sets. 
* **No‑mock/no‑simulate:** runtime guarantees ensure you report actual file/line/token data only.

### Assumptions & Risks

* `canary show` returns fields listed; if some (e.g., `owner`, `priority`) are absent, set to `null`—never invent.
* REQ‑ID regex may differ across repos; adjust `req_id_regex` accordingly.
* Hidden/spec/template files are excluded by default; surface inclusion explicitly when the CLI cannot filter them.

### Targeted questions (for fit)

1. Confirm your canonical **REQ‑ID** pattern and whether aspects are fixed or free‑form.
2. Should HUMAN_TEXT always be returned, or only when `--json` isn’t passed?
3. Any extra analysis desired (e.g., **age since UPDATED**, **per‑aspect coverage %**, **owner load**)?
