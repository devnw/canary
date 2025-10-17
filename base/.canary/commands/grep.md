## GrepCmd (Search CANARY Tokens)

```yaml
---
description: Search CANARY tokens by keyword or pattern across all fields (strict outputs, parallel processing, no-mock/no-simulate)
command: GrepCmd
version: 2.2
subcommands: [grep]
outputs:
  - human_text: STDOUT (concise operator view; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  group_by: none         # values: none | requirement
  case_insensitive: true
  max_results: 500       # truncation guard for operator UX; JSON includes truncation flag
---
```

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into: `pattern` (string; required), flags `--group-by <none|requirement>`, `--db <path>`.
* **Search scope:** All token fields: `REQ`, `FEATURE`, `ASPECT`, `STATUS`, `OWNER`, `UPDATED`, `DOC`, `DOC_HASH`, `TEST`, `BENCH`, file path, and any indexed metadata.

### 2) Preconditions & Resolution

1. **Pattern gate:** must be non‑empty after trim; default **case‑insensitive** substring. If user supplies a regex prefix (e.g., `re:`), treat the remainder as a safe regex, else plain substring.

   * Reject unsupported or catastrophic regex (e.g., backtracking bombs) → `ERROR_PATTERN_UNSAFE(details)`.
2. **DB gate:** ensure DB at `db_path` exists/readable; otherwise `ERROR_DB_MISSING(path)` with remediation (`canary index`).
3. **Command availability:** ensure `canary grep` exists; else `ERROR_CANARY_GREP_UNAVAILABLE()`.

### 3) Planning & Parallelism

Build a **Work DAG** with **Concurrency Groups (CG)**; **join** before aggregation. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Query:**

  ```bash
  canary grep <pattern> [--db <path>] [--group-by <mode>]
  ```

  Return raw matches with field identifiers and file:line.
* **CG‑2 Normalize:** Normalize field names, unify casing, compute highlight spans, and deduplicate identical hits.
* **CG‑3 Aggregate:** Group (if requested), compute distributions (by status/aspect/dir), truncation flags, and suggestions.

### 4) Behavior (must do; never simulate)

* **Run the real command**; do not fabricate matches or counts.
* **Case‑insensitive** by default; set engine flags accordingly or pre‑lowercase both haystack and needle.
* **Matched‑field evidence:** for each hit, report `matched.field`, `match_excerpt`, and `match_span` within that field’s value.
* **Navigation:** include `file` and `line` for each hit; if unavailable, set `line=null` (do **not** invent).
* **Truncation:** cap display at `max_results` (JSON includes `truncated=true` and `limit=max_results`).

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **query**, and after **aggregation**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"grep|query|aggregate",
  "f":[["<db_path>",1,1]],
  "k":["pattern:<p>","group:<none|requirement>","hits:<N>","truncated:<bool>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["present results","offer refinements"]
}'
```

*Compact keys capture filenames, key facts, false‑positives, and next steps with minimal tokens.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** 

**A. HUMAN_TEXT (optional)**
Begin with `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Recommended contents: title, total matches, top N hits with file:line, grouped summaries (if requested), status distribution, and refinement tips.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "pattern": "<original pattern>",
  "group_by": "none|requirement",
  "db_path": ".canary/canary.db",
  "totals": { "matches": 0, "requirements": 0, "files": 0 },
  "status_distribution": { "IMPL": 0, "TESTED": 0, "BENCHED": 0, "OTHER": 0 },
  "dir_distribution": [{ "dir": "src/api/auth", "matches": 0 }],
  "truncated": false,
  "limit": 500,
  "items": [
    {
      "req_id": "{{.ReqID}}-<ASPECT>-API-120",
      "feature": "UserAuthentication",
      "status": "TESTED",
      "aspect": "API",
      "priority": null,
      "file": "src/api/auth/user.go",
      "line": 45,
      "test": "TestCANARY_CB..._UserAuthentication",
      "bench": null,
      "matched": { "field": "feature|file|test|bench|req_id|aspect|status|owner|doc", "match_excerpt": "...Auth...", "match_span": [5, 9] }
    }
  ],
  "grouped": [
    {
      "req_id": "{{.ReqID}}-<ASPECT>-API-120",
      "matches": 0,
      "files": ["src/api/auth/user.go","src/api/auth/middleware.go"]
    }
  ],
  "suggestions": [
    "canary grep src/api/auth",
    "canary grep {{.ReqID}}-<ASPECT>-API-120",
    "canary show {{.ReqID}}-<ASPECT>-API-121"
  ],
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **Pattern Gate:** non‑empty, safe (reject catastrophic regex).
* **DB Gate:** DB reachable at `db_path`.
* **Schema Gate:** JSON conforms (field names/types present; `items[*].matched` present).
* **Counting Gate:** `totals.matches` equals the length of `items` (post‑truncation) and group sums are consistent.
* **Navigation Gate:** `file` exists; `line` may be `null` but never fabricated.
* **Case Gate:** Case‑insensitive matching verified (e.g., `Auth` equals `auth`).

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_PATTERN_EMPTY()`
* `ERROR_PATTERN_UNSAFE(details)`
* `ERROR_DB_MISSING(path)` → suggest `canary index`
* `ERROR_CANARY_GREP_UNAVAILABLE()`
* `ERROR_CANARY_GREP_FAILED(exit_code,stderr_excerpt)`
* `ERROR_NO_MATCHES(pattern)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real command executed; **no simulated/mocked** results.
* Case‑insensitive matching enforced; matched‑field evidence included.
* Grouping and distributions computed; truncation flags accurate.
* HUMAN_TEXT (if produced) concise and consistent with JSON.
* CANARY snapshot(s) emitted when required.
* JSON returned **without** code‑fence wrapping; field names exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## Search Results for "Auth"

Found 8 matches (group-by: none; truncated: false)

📌 {{.ReqID}}-<ASPECT>-API-120 — UserAuthentication
   Status: TESTED | Aspect: API
   Location: src/api/auth/user.go:45
   Match: feature name ("Auth")

📌 {{.ReqID}}-<ASPECT>-API-120 — AuthMiddleware
   Status: TESTED | Aspect: API
   Location: src/api/auth/middleware.go:23
   Match: feature name ("Auth")

📌 {{.ReqID}}-<ASPECT>-API-121 — OAuth2Integration
   Status: IMPL | Aspect: API
   Location: src/api/auth/oauth.go:67
   Match: file path ("auth")

**Summary**
- Total matches: 8
- Requirements: 3 ({{.ReqID}}-<ASPECT>-API-120, {{.ReqID}}-<ASPECT>-API-121, {{.ReqID}}-<ASPECT>-Security-134)
- Status: TESTED (5), IMPL (3)
- Primary location: src/api/auth/

**Refinements**
- canary grep src/api/auth
- canary grep {{.ReqID}}-<ASPECT>-API-120
- canary show {{.ReqID}}-<ASPECT>-API-121
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers enable reliable parsing and downstream checks. 
* **Section delimiting & structure:** clearer inputs → gates → behavior → outputs for maintainability. 
* **Parallel pipeline:** explicit Work DAG + concurrency groups + join points for fast searches over large datasets. 
* **No‑mock/no‑simulate:** made explicit as runtime guarantees to prevent fabricated hits or counts. 

### Assumptions & Risks

* `canary grep` returns field‑level identifiers and file:line; if not, GrepCmd performs safe post‑processing to map fields.
* Regex support may vary; we default to substring and opt‑in to regex via `re:` prefix to avoid catastrophic patterns.
* Directory distribution uses normalized paths; symlinked trees may double‑count unless deduped.

### Targeted questions (for fit)

1. Confirm canonical field set indexed by `canary grep` (exact names).
2. Do you want a hard cap for `max_results` different from 500?
3. Should we add an optional `--regex` flag instead of the `re:` prefix?
4. Any additional distributions desired (e.g., by OWNER or UPDATED month)?
