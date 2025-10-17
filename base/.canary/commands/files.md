## FilesCmd (List Implementation Files by Aspect)

```yaml
---
description: List all implementation files containing CANARY tokens for a requirement, grouped by aspect with token counts, with strict outputs and analysis
command: FilesCmd
version: 2.1
subcommands: [list]
outputs:
  - human_text: STDOUT (concise table for operators; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  include_specs_and_templates: false
---
```

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into: `req_id` (e.g., `CBIN-123-API-105`), flags `--all` (include specs/templates), `--db <path>`.
* **Repository assumptions:**

  * DB at `.canary/canary.db` unless `--db` provided.
  * Specs/templates under `.canary/specs/**` and `.canary/templates/**`.

### 2) Preconditions & Resolution

1. **REQ‑ID format gate:** must match `CBIN-[A-Z]+-[0-9]+` (or your canonical pattern). On mismatch → `ERROR_REQ_FORMAT(req_id)`.
2. **Database availability:** if DB missing/unreadable, return `ERROR_DB_MISSING(path)` with remediation: `canary index`.
3. **Scope selection:**

   * Default excludes specs/templates.
   * If `--all`, include them and mark `included_non_impl: true` in output.
4. **Command availability:** ensure `canary files` exists; else `ERROR_CANARY_FILES_UNAVAILABLE()`.

### 3) Planning & Parallelism

Build a **Work DAG** with **Concurrency Groups** (CG); **join** before shared writes/aggregation. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Enumerate:** `canary files <REQ-ID> [--db <path>] [--all]` → raw list of file paths and token counts.
* **CG‑2 Enrich (optional):** obtain **line numbers** for navigation:

  * Prefer `canary tokens <REQ-ID> --loc` if available; else safe grep on `^//\s*CANARY:.*REQ=<REQ-ID>` (no edits).
* **CG‑3 Group & Analyze:** group by **ASPECT**, compute totals/flags (missing aspects, scatter metric), and navigation suggestions.

### 4) Behavior (must do; never simulate)

* **Run** the real command:

  ```bash
  canary files <REQ-ID> [--db <path>] [--all]
  ```
* **Exclude** spec/template files unless `--all`.
* **Group** by aspect (CLI, API, Engine, Storage, Docs, Frontend, etc. — use token `ASPECT=` when present; else infer from path heuristics and **label as inferred**).
* **Compute metrics:** per‑file token counts, totals, “scatter” (files with exactly 1 token), and missing aspects (expected but absent).
* **Navigation hints:** return top candidates to open (highest token counts; include first token line if available).
* **Do not fabricate** file paths, counts, or lines; if any source is unavailable, return a failure mode with precise remediation.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **enumeration**, and after **aggregation**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"files|enumerate|analyze",
  "f":[["<db_path>",1,1],["<output_tmp>",1,999]],
  "k":["req:<REQ-ID>","files:<N>","tokens:<T>","include_specs:<bool>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["present results","offer navigation"]
}'
```

*Compact keys capture filenames, line spans, key facts, false‑positives, and next steps with minimal tokens.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** 

**A. HUMAN_TEXT (optional)**
Begin with: `=== HUMAN_TEXT BEGIN ===`
Recommended contents: title, per‑aspect bullet lists `path (N tokens)`, totals, short analysis, navigation tips.
End with: `=== HUMAN_TEXT END ===`

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "req_id": "CBIN-API-105",
  "db_path": ".canary/canary.db",
  "included_non_impl": false,
  "totals": { "files": 0, "tokens": 0, "aspects": 0, "one_token_files": 0 },
  "groups": [
    {
      "aspect": "API",
      "source": "token|inferred",
      "files": [
        { "path": "src/api/user.go", "token_count": 4, "first_token_line": 12 }
      ],
      "files_count": 0,
      "tokens_count": 0
    }
  ],
  "missing_aspects": ["CLI","Storage"],
  "top_files": [
    { "path": "src/api/user.go", "token_count": 4 },
    { "path": "cmd/app/user.go", "token_count": 2 }
  ],
  "scatter": {
    "files_with_one_token": 0,
    "ratio": 0.0,
    "flag": false,
    "threshold_ratio": 0.5
  },
  "navigation": {
    "impl": ["src/api/user.go","internal/db/user.go"],
    "tests": ["src/api/user_test.go"],
    "docs": ["docs/api/user-endpoints.md"]
  },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **REQ‑ID Format Gate:** valid pattern?
* **DB Gate:** DB reachable at `db_path`? else **FAIL** with remediation.
* **Exclusion Gate:** specs/templates excluded unless `--all` set.
* **Token Consistency Gate:** group/token totals equal sum of per‑file counts.
* **Schema Gate:** `SUMMARY_JSON` conforms exactly (field names, types).
* **Navigation Gate:** if line numbers unavailable, set `first_token_line: null` and **do not invent**.

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_REQ_FORMAT(req_id)`
* `ERROR_REQ_NOT_FOUND(req_id)`
* `ERROR_DB_MISSING(path)` → suggest `canary index`
* `ERROR_CANARY_FILES_UNAVAILABLE()`
* `ERROR_CANARY_FILES_FAILED(exit_code,stderr_excerpt)`
* `ERROR_NO_TOKENS_FOUND(req_id)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real command executed; no simulated/mocked results.
* Grouping by aspect correct; **source** marks `token` vs `inferred`.
* Totals and scatter computed; missing aspects identified with rationale.
* HUMAN_TEXT (if produced) concise and matches JSON.
* CANARY snapshot(s) emitted when required.
* JSON returned **without** code fences; field names exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

Start/end markers as specified:

```
=== HUMAN_TEXT BEGIN ===
## Implementation Files for CBIN-API-105

### API
- src/api/user.go (4 tokens)
- src/api/user_test.go (3 tokens)

### CLI
- cmd/app/user.go (2 tokens)
- cmd/app/commands.go (1 token)

### Storage
- internal/db/user.go (1 token)

### Docs
- docs/api/user-endpoints.md (1 token)

**Totals:** 6 files • 12 tokens  
**Analysis:** Primary impl in `src/api/user.go`; CLI support present; Storage layer minimal.  
**Navigation:** impl → src/api/user.go • tests → src/api/user_test.go • CLI → cmd/app/user.go
=== HUMAN_TEXT END ===
```

### 11) Operator Guidance (safe defaults)

* **Automatic Execution:** If `req_id` is present and format‑valid, run immediately.
* **Focus on Implementation:** Keep specs/templates excluded unless `--all`.
* **Navigation:** Prefer file:line when available; otherwise path‑only hints.
* **Missing DB:** Prompt to run `canary index` with the resolved `db_path`.
* **Respect user intent:** If user passes `--db`, do not fall back to defaults silently.

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers enable reliable parsing and downstream checks. 
* **Section delimiting & structure:** clear inputs → gates → behavior → outputs improve maintainability and UX. 
* **Parallel batch handling:** explicit Work DAG + concurrency groups + join points accelerate `--all` scans without race risks. 

### Assumptions & Risks

* `canary files` returns accurate per‑file token counts; line‑level locations may require a secondary command (or safe grep).
* Aspect inference from paths is marked as `inferred` to avoid misrepresentation.
* DB path defaults to `.canary/canary.db` unless overridden.

### Targeted questions (for fit)

1. Confirm the canonical **REQ‑ID regex** and the **aspect vocabulary** (authoritative list).
2. Does `canary tokens --loc` exist? If not, which locator should FilesCmd call for line numbers?
3. Should scatter thresholds (e.g., `threshold_ratio=0.5`) be project‑specific?
4. Any additional fields desired in `SUMMARY_JSON` (e.g., per‑aspect coverage %)?
