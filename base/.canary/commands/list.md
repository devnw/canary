## ListCmd (Priority Listing with Filters & Sorting)

```yaml
---
description: List CANARY requirements with priority-first ordering, filtering, safe sorting, and machine-readable summaries (strict, verifiable, no-mock/no-simulate)
command: ListCmd
version: 2.3
outputs:
  - human_text: STDOUT (concise operator view; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  limit: 10                # 0 = unlimited
  order_by: "priority ASC, updated_at DESC"
  include_hidden: false    # hide tests/templates/examples/agent dirs
  hidden_globs: ["**/test/**","**/templates/**",".canary/agents/**","**/examples/**"]
  format: text             # or json (mirrors --json)
safe_order_by_columns: ["priority","updated_at","status","aspect","phase","owner","req_id"]
safe_order_by_directions: ["ASC","DESC"]
---
```

<!-- CANARY: REQ=CBIN-135; FEATURE="ListCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into filters/controls:

  * Filters: `--status <STUB|IMPL|TESTED|BENCHED|REMOVED>`, `--aspect <CLI|API|Engine|Storage|Security|Docs|…>`, `--phase <Phase0|Phase1|Phase2|Phase3>`, `--owner <name>`, `--spec-status <draft|approved|in-progress|completed|archived>`.
  * Output control: `--limit N` (int; 0=unlimited), `--order-by <clause>`, `--json` (sets `format=json`), `--include-hidden`, `--db <path>`.

### 2) Preconditions & Resolution

1. **DB gate:** `.canary/canary.db` (or `--db`) must exist/readable; else `ERROR_DB_MISSING(path)` with remediation `canary index`.
2. **Flag validation:** unknown flags or invalid enums → `ERROR_FLAG_INVALID(flag,value)`.
3. **Limit gate:** `--limit` must be `N ≥ 0`; non‑numeric → `ERROR_LIMIT_INVALID(value)`.
4. **Order‑by safety:** allow only comma‑separated clauses of **`<column> <direction>`** where `<column>` ∈ `safe_order_by_columns` and `<direction>` ∈ `safe_order_by_directions`; otherwise `ERROR_ORDER_BY_UNSAFE(clause)`.
5. **Hidden scope:** default excludes `hidden_globs`; `--include-hidden` flips inclusion and marks it in output.

### 3) Planning & Parallelism

Create a **Work DAG** with **Concurrency Groups (CG)**; **join** before final aggregation. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Query:**

  ```bash
  canary list [resolved flags]
  ```

  Return rows: `req_id, feature, status, aspect, phase, owner, priority, updated_at, file, line, spec_status`.
* **CG‑2 Enrich:** (optional) fetch first test name or doc path if present (non‑blocking lookups).
* **CG‑3 Aggregate:** apply grouping/filters, compute counts, produce analysis/recommendations.

### 4) Behavior (must do; never simulate)

* **Run the real command**; do **not** fabricate rows or totals.
* Apply default ordering `priority ASC, updated_at DESC` unless `--order-by` passes the safety gate.
* Respect `--limit` at the **row level** (post‑filter, post‑sort).
* Exclude hidden paths unless `--include-hidden`.
* Compute **analysis**:

  * `highest_priority_stub` (first STUB by order)
  * `needs_tests` = items with `status=IMPL`
  * `stale` = items with lowest `updated_at` within result set
* Emit **recommendations**: `/canary.plan <req>` for STUB, `/canary.next` for auto‑selection, and test‑adding notes for IMPL.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **query**, and after **aggregation**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"list|query|aggregate",
  "f":[["<db_path>",1,1]],
  "k":["filters:<status/aspect/phase/owner>","limit:<N>","order:\"<clause>\"","rows:<M>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["present results","offer next-actions"]
}'
```

*Compact keys capture filenames, key facts, false‑positives, and next steps with minimal tokens.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** 

**A. HUMAN_TEXT (optional)**
Begin with: `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Recommended contents: title, total count, top N rows (with file:line), short analysis, recommendations.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "filters": {
    "status": ["STUB"],
    "aspect": ["API"],
    "phase": [],
    "owner": [],
    "spec_status": []
  },
  "db_path": ".canary/canary.db",
  "include_hidden": false,
  "order_by": "priority ASC, updated_at DESC",
  "limit": 10,
  "totals": { "returned": 0, "matched": 0, "stub": 0, "impl": 0, "tested": 0, "benched": 0, "removed": 0 },
  "items": [
    {
      "req_id": "{{.ReqID}}-<ASPECT>-API-134",
      "feature": "UserOnboarding",
      "status": "STUB",
      "aspect": "API",
      "phase": "Phase0",
      "owner": "team-core",
      "priority": 1,
      "updated_at": "2025-10-16T12:00:00Z",
      "spec_status": "approved",
      "location": { "file": ".canary/specs/{{.ReqID}}-<ASPECT>-API-134-user-onboarding/spec.md", "line": 1 }
    }
  ],
  "analysis": {
    "highest_priority_stub": "{{.ReqID}}-<ASPECT>-API-134",
    "needs_tests": ["{{.ReqID}}-<ASPECT>-API-105","{{.ReqID}}-<ASPECT>-Engine-142"],
    "stale_candidates": ["{{.ReqID}}-<ASPECT>-API-099","{{.ReqID}}-<ASPECT>-CLI-088"]
  },
  "recommendations": [
    "/canary.plan {{.ReqID}}-<ASPECT>-API-134",
    "Add tests for {{.ReqID}}-<ASPECT>-API-105, {{.ReqID}}-<ASPECT>-Engine-142",
    "/canary.next"
  ],
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **DB Gate:** DB reachable at `db_path`.
* **Filter Gate:** all filter values within allowed enums.
* **Order‑by Safety Gate:** clause strictly within `safe_order_by_*` allowlists.
* **Limit Gate:** numeric and ≥0.
* **Counting Gate:** `totals.returned == len(items)`; status tallies equal counts in `items`.
* **Schema Gate:** JSON conforms; field names/types exact.
* **Hidden Gate:** hidden items excluded unless `include_hidden=true`.

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_DB_MISSING(path)` → suggest `canary index`
* `ERROR_FLAG_INVALID(flag,value)`
* `ERROR_LIMIT_INVALID(value)`
* `ERROR_ORDER_BY_UNSAFE(clause)`
* `ERROR_CANARY_LIST_UNAVAILABLE()`
* `ERROR_CANARY_LIST_FAILED(exit_code,stderr_excerpt)`
* `ERROR_NO_MATCHES(filters)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real command executed; **no simulated/mocked** results.
* Filters applied; ordering safe; limit enforced.
* Hidden files/templates/tests excluded by default.
* HUMAN_TEXT (if produced) is concise and consistent with JSON.
* CANARY snapshot(s) emitted when required.
* JSON returned **without** code fences; schema exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## Top Priority Requirements (filters: status=STUB; limit=5)

Found 12 matching • Showing top 5 (priority ASC, updated_at DESC)

📌 {{.ReqID}}-<ASPECT>-API-134 — UserOnboarding
   Status: STUB | Aspect: API | Priority: 1
   Location: .canary/specs/{{.ReqID}}-<ASPECT>-API-134-user-onboarding/spec.md:1

📌 {{.ReqID}}-<ASPECT>-Engine-140 — ValidationRules
   Status: STUB | Aspect: Engine | Priority: 1
   Location: .canary/specs/{{.ReqID}}-<ASPECT>-Engine-140-validation-rules/spec.md:1

📌 {{.ReqID}}-<ASPECT>-CLI-155 — FlagUX
   Status: STUB | Aspect: CLI | Priority: 2
   Location: cmd/canin/flags.go:77

**Analysis**
- Highest Priority STUB: {{.ReqID}}-<ASPECT>-API-134
- Items needing tests (IMPL): {{.ReqID}}-<ASPECT>-API-105, {{.ReqID}}-<ASPECT>-Engine-142
- Stale candidates: {{.ReqID}}-<ASPECT>-API-099, {{.ReqID}}-<ASPECT>-CLI-088

**Recommendations**
1) /canary.plan {{.ReqID}}-<ASPECT>-API-134
2) Add tests for {{.ReqID}}-<ASPECT>-API-105, {{.ReqID}}-<ASPECT>-Engine-142
3) /canary.next
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers enable reliable parsing and downstream checks. 
* **Section delimiting & structure:** clearer inputs → gates → behavior → outputs improve maintainability and UX. 
* **Parallel pipeline:** explicit Work DAG + concurrency groups + join points for fast list/aggregate flows. 
* **Security:** **order‑by allowlist** prevents SQL‑injection via sort clauses while preserving flexibility.
* **No‑mock/no‑simulate:** runtime guarantees ensure real DB queries and counts.

### Assumptions & Risks

* `canary list` returns row fields listed in **CG‑1**; if some fields are unavailable (e.g., `phase`), set to `null` (do **not** invent).
* Hidden glob patterns may need tuning per repo; expose via `hidden_globs` in defaults.
* Very large `--limit 0` results may be truncated by the model context; prefer `/canary.next` for agent‑driven selection.

### Targeted questions (for fit)

1. Confirm canonical enums for **status**, **aspect**, **phase**, and **spec_status**.
2. Any additional **order‑by** columns to allowlist (e.g., `priority_bucket`, `owner`)?
3. Should HUMAN_TEXT always be returned, or only when `--json` isn’t used?
4. Include **age buckets** (e.g., days since `updated_at`) in `SUMMARY_JSON`?
