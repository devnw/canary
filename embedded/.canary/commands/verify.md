## VerifyCmd (Verify GAP Analysis Claims vs. Reality)

```yaml
---
description: Verify GAP_ANALYSIS.md claims against actual CANARY token status with strict rules and auditable, machine-readable outputs (no-mock/no-simulate)
command: VerifyCmd
version: 2.4
subcommands: [verify]
outputs:
  - human_text: STDOUT (concise operator report; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  root: .
  gap_path: ./GAP_ANALYSIS.md
  db_path: .canary/canary.db
  strict: true                  # enforce staleness (Article VII)
  stale_days: 30
  pass_exit_code: 0
  fail_exit_code: 2
  claims_regex: '^[\s]*[✅✔☑]\s+({{.ReqID}}-<ASPECT>-(?:[A-Za-z]+-)?[0-9]{3,})\b'   # supports {{.ReqID}}-<ASPECT>-001 and {{.ReqID}}-<ASPECT>-API-105
---
```

<!-- CANARY: REQ=CBIN-112; FEATURE="VerifyCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse flags: `--path <GAP_ANALYSIS.md>`, `--root <path>`, `--strict/--no-strict`, `--db <path>`, `--json`, `--no-color`.

### 2) Preconditions & Resolution

1. **GAP file gate:** `gap_path` must exist/readable → else `ERROR_GAP_FILE_NOT_FOUND(path)`.
2. **DB gate:** database exists/readable at `db_path` → else `ERROR_DB_MISSING(path)` (remediation: `canary index`).
3. **Scanner gate:** `canary scan` must be available → else `ERROR_SCANNER_UNAVAILABLE()`.
4. **Claims parse:** extract unique REQ IDs by `claims_regex`; empty set → `ERROR_NO_CLAIMS()`.
5. **Strict policy:** when `strict=true`, treat **TESTED/BENCHED with UPDATED > stale_days** as **stale** failures.

### 3) Planning & Parallelism

Create a **Work DAG** and run independent steps in **Concurrency Groups (CG)**; **join** before aggregation. If true parallelism isn’t available, interleave safely at join points. 

* **CG‑1 Parse Claims (CPU):** scan `gap_path`, collect `{req_id, line}`.
* **CG‑2 Verify (I/O):**

  ```bash
  canary scan --root <root> --verify <gap_path> {--strict}
  ```

  Capture `exit_code`, `stdout`, `stderr`.
* **CG‑3 Cross‑check (CPU):** For each claimed `req_id`, locate actual tokens/status/UPDATED from scan/DB; label as:

  * **valid**: STATUS ∈ {TESTED, BENCHED} and (if strict) not stale
  * **overclaim**: STATUS ∈ {STUB, IMPL, MISSING}
  * **stale**: valid but UPDATED > stale_days
  * **missing**: not present in codebase
* **CG‑4 Assemble:** HUMAN_TEXT (optional) + strict **SUMMARY_JSON**; set `result.status = PASS|FAIL` and `result.exit_code = pass_exit_code|fail_exit_code` per outcomes.

### 4) Behavior (must do; never simulate)

* **Run the real command**; do **not fabricate** claims, tokens, or counts.
* **Respect exit codes:** `0=PASS`, `2=FAIL`; any other non‑zero → `ERROR_SCAN_FAILED(exit_code,stderr_excerpt)`.
* **Evidence:** include file:line when available; if unknown set `line=null` (never invent).
* **Operator UX:** when `--no-color`, ensure parse‑safe output; keep detailed data in JSON.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **parse**, and after **verify**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"verify|parse|scan",
  "f":[["<gap_path>",1,999],["<root>",1,1]],
  "k":["claims:<N>","strict:<bool>","stale_days:<N>","pass_exit_code:0","fail_exit_code:2"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["present results","remediate failures"]
}'
```

*Compact snapshots + stable delimiters support reliable context handoff and re‑entry.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** Use structured outputs for reliable automation. 

**A. HUMAN_TEXT (optional)**
Begin with `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Include: verification date; PASS/FAIL; counts; per‑claim bullets with reason; remediation checklist.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "params": {
    "root": ".",
    "gap_path": "GAP_ANALYSIS.md",
    "db_path": ".canary/canary.db",
    "strict": true,
    "stale_days": 30
  },
  "claims": [
    { "req_id": "{{.ReqID}}-<ASPECT>-001", "line": 12 },
    { "req_id": "{{.ReqID}}-<ASPECT>-API-105", "line": 21 }
  ],
  "results": {
    "status": "PASS|FAIL",
    "exit_code": 0,
    "counts": { "claims": 0, "valid": 0, "overclaims": 0, "stale": 0, "missing": 0 }
  },
  "details": [
    {
      "req_id": "{{.ReqID}}-<ASPECT>-003",
      "claim": { "line": 33 },
      "found": true,
      "status": "IMPL",
      "updated": "2025-10-10",
      "stale": false,
      "verdict": "overclaim",
      "reason": "Status IMPL (needs TEST=)"
    }
  ],
  "remediation": [
    "Add TEST= for IMPL items: {{.ReqID}}-<ASPECT>-003",
    "Run `canary scan --update-stale` for stale: {{.ReqID}}-<ASPECT>-004",
    "Remove ✅ for missing: {{.ReqID}}-<ASPECT>-099 or implement before re‑adding"
  ],
  "scanner": {
    "cmd": "canary scan --root . --verify GAP_ANALYSIS.md --strict",
    "exit_code": 0,
    "stderr_tail": "<last-200-bytes>"
  },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Verification Rules (authoritative)

* **Valid claim:** STATUS ∈ {TESTED, BENCHED}, and if `strict=true`, **not stale** (`UPDATED ≤ stale_days`).
* **Overclaim:** claimed but STATUS ∈ {STUB, IMPL, MISSING}.
* **Stale:** valid claim with UPDATED older than `stale_days`.
  *(Rules are enforced consistently and mirrored in `results.counts` and `details[*].verdict`.)*

### 8) Validation Gates (compute & report)

* **GAP file Gate:** file exists/readable.
* **Claims Gate:** ≥1 parsed claim; duplicates removed.
* **DB Gate:** DB reachable; else remediation provided.
* **Scanner Gate:** tool available; exit‑code semantics respected.
* **Counting Gate:** `results.counts.claims == len(claims)` and equals `valid+overclaims+stale+missing`.
* **Schema Gate:** JSON conforms exactly; field names/types exact.
* **Strict Gate:** when strict, staleness computed using `stale_days` and `UPDATED` dates.

### 9) Failure Modes (return one with reason + remediation)

* `ERROR_GAP_FILE_NOT_FOUND(path)`
* `ERROR_DB_MISSING(path)` → suggest `canary index`
* `ERROR_SCANNER_UNAVAILABLE()`
* `ERROR_NO_CLAIMS()`
* `ERROR_SCAN_FAILED(exit_code,stderr_excerpt)`
* `ERROR_PARSE_GAP_FILE(reason)`
* `ERROR_PARSE_OUTPUT(reason)`

### 10) Quality Checklist (auto‑verify before output)

* Real `canary scan --verify` executed; **no simulated/mocked** results.
* Claims parsed by regex; IDs de‑duplicated and normalized.
* Exit codes handled (0 pass; 2 fail).
* HUMAN_TEXT concise and consistent with JSON.
* CANARY snapshots emitted at parse + verify phases.
* JSON returned **without** code fences; schema exact. 

### 11) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## GAP Analysis Verification Results
**Verification Date:** 2025-10-17 • **Strict:** true (30‑day staleness)

**Status:** ❌ FAIL
**Claims:** 4 • **Valid:** 2 • **Overclaims:** 1 • **Stale:** 1 • **Missing:** 0

- ✅ {{.ReqID}}-<ASPECT>-001 — UserAuth → BENCHED (verified)
- ✅ {{.ReqID}}-<ASPECT>-002 — DataValidation → TESTED (verified)
- ❌ {{.ReqID}}-<ASPECT>-003 — ReportGen → IMPL only (overclaim; add TEST=TestReportGeneration)
- ⚠️ {{.ReqID}}-<ASPECT>-004 — Cache → TESTED but UPDATED 288d ago (stale; run `canary scan --update-stale`)

**Action Required**
1) Add tests & TEST= for {{.ReqID}}-<ASPECT>-003; re‑run verify
2) Refresh UPDATED for {{.ReqID}}-<ASPECT>-004 with `canary scan --update-stale`
3) Commit changes; enforce in CI (exit code 2 on failure)
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs**: strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers for CI parsing and dashboards. 
* **Section delimiting & structure**: explicit inputs → gates → behavior → outputs for maintainability and reliability. 
* **Parallel pipeline**: DAG with concurrency groups for parse/scan/aggregate reduces latency and clarifies join points. 
* **No‑mock/no‑simulate**: runtime guarantees enforce real scans and real exit‑code handling.

### Assumptions & Risks

* `canary scan --verify` returns fields sufficient to map `req_id → status, updated`. If not, supplement via DB read.
* GAP file may use different checkmarks; regex includes common variants.
* Mixed ID formats (with/without ASPECT) are normalized before cross‑checking.

### Targeted questions (for fit)

1. Confirm canonical **REQ‑ID** patterns accepted in GAP files (with/without ASPECT).
2. Should **underclaims** (verified items not listed) be flagged as informational?
3. Do you want `--stale-days <N>` exposed or fixed by constitution (Article VII)?
