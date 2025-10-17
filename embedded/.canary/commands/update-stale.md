## UpdateStaleCmd (Auto‑update `UPDATED` for Stale Tokens)

```yaml
---
description: Automatically update UPDATED for stale CANARY tokens (TESTED/BENCHED older than threshold) with strict verification and auditable outputs
command: UpdateStaleCmd
version: 2.4
subcommands: [update-stale]
outputs:
  - human_text: STDOUT (concise operator view; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  root: .
  stale_days: 30                # Article VII default
  eligible_statuses: ["TESTED","BENCHED"]
  out_json: status.json         # scanner output (pre & post)
  out_csv: status.csv
  confirm_required: true        # bypass with --force
  verify_after_update: true
  git_suggest_commit: true
---
```

<!-- CANARY: REQ=CBIN-114; FEATURE="UpdateStaleCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse flags: `--root <path>`, `--stale-days <N>`, `--force`, `--json`, `--out <status.json>`, `--csv <status.csv>`, `--no-color`.

### 2) Preconditions & Resolution

1. **Root Gate:** path exists and is readable; else `ERROR_ROOT_NOT_FOUND(path)`.
2. **Scanner Gate:** `canary scan` must be available; else `ERROR_SCANNER_UNAVAILABLE()`.
3. **Stale‑days Gate:** integer `N ≥ 1`; else `ERROR_STALE_DAYS_INVALID(value)`.
4. **Eligibility Policy:** **only** tokens with `status ∈ {TESTED,BENCHED}` are updatable; `STUB/IMPL` **must be excluded**.
5. **Clock/Date Source:** use system UTC date `YYYY‑MM‑DD` for all updates; never fabricate.

### 3) Planning & Parallelism

Create a **Work DAG** and run independent steps in **Concurrency Groups (CG)**; **join** before shared writes. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Scan (I/O bound):**

  ```bash
  canary scan --root <root> --strict --out <out_json> --csv <out_csv>
  ```

  Parse `<out_json>`; compute stale set = tokens with `status ∈ eligible_statuses` and `UPDATED < today - stale_days`.
* **CG‑2 Preview:** build **HUMAN_TEXT** preview + machine list (files, lines, ages). If `--force` **absent**, stop for confirmation (see Output Contract).
* **CG‑3 Apply (idempotent):**

  ```bash
  canary scan --root <root> --update-stale
  ```

  Update only `UPDATED=` field; preserve other fields verbatim (order/spacing).
* **CG‑4 Verify:** re‑run strict scan; assert **no stale** remain; compute before/after deltas and modified files.
* **CG‑5 Assemble Outputs:** produce **SUMMARY_JSON** and optional **HUMAN_TEXT**; include suggested git commands.

### 4) Behavior (must do; never simulate)

* **Run real commands;** do **not** invent file paths, lines, or counts.
* **Safety:** Only touch tokens that (a) are `TESTED` or `BENCHED`, (b) are older than `stale_days`.
* **Minimal patching:** change `UPDATED=` value only; keep whitespace, ordering, and all other fields identical.
* **Atomicity:** per‑file writes should be atomic (temp file + move) or fail with clear error; do **not** partially rewrite on error.
* **Post‑verify:** if any stale remain, mark `verification.passed=false` and list residuals; do not claim success.

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **scan**, and after **apply+verify**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"update-stale|scan|apply|verify",
  "f":[["<root>",1,1],["<out_json>",1,1]],
  "k":["stale_days:<N>","found:<F>","updated:<U>","skipped:<S>","residual:<R>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["commit changes","run verify with GAP_ANALYSIS.md"]
}'
```

*Compact keys minimize tokens while preserving filenames, counts, false‑positives, and next steps.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** Use structured outputs for reliable automation. 

**A. HUMAN_TEXT (optional)**
Begin with `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`

* **Preview mode (no `--force`):** show “Stale Tokens Found: N” + per‑item details (req_id, file:line, age, current→new date) and ask for confirm.
* **Result mode (`--force` or after confirm):** show update date, counts, per‑file bullets, and verification outcome.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "params": {
    "root": ".",
    "stale_days": 30,
    "forced": false,
    "out_json": "status.json",
    "out_csv": "status.csv"
  },
  "totals": {
    "found": 0,
    "eligible": 0,
    "updated": 0,
    "skipped": 0
  },
  "items": [
    {
      "req_id": "{{.ReqID}}-<ASPECT>-API-001",
      "feature": "UserAuth",
      "status": "TESTED",
      "file": "src/api/auth.go",
      "line": 10,
      "age_days": 288,
      "updated_before": "2024-01-01",
      "updated_after": "2025-10-17",
      "action": "updated|skipped"
    }
  ],
  "files_modified": ["src/api/auth.go","internal/cache/cache.go"],
  "verification": {
    "passed": true,
    "residual_stale": [],
    "post_scan_json": "status.json"
  },
  "git": {
    "suggest_commit": true,
    "cmd": "git add <files>; git commit -m \"chore: update stale CANARY tokens (N files)\""
  },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **Eligibility Gate:** updated **only** tokens with `status ∈ eligible_statuses`.
* **Staleness Gate:** each updated item had `age_days ≥ stale_days`.
* **Diff Integrity Gate:** before/after comparison shows **only** the `UPDATED=` field changed.
* **Counting Gate:** `totals.updated + totals.skipped == totals.eligible == len(items)`.
* **Verification Gate:** if `verify_after_update`, strict re‑scan confirms `verification.passed=true`.
* **Schema Gate:** JSON conforms; field names/types exact.

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_ROOT_NOT_FOUND(path)`
* `ERROR_SCANNER_UNAVAILABLE()`
* `ERROR_SCAN_FAILED(exit_code,stderr_excerpt)`
* `ERROR_NO_STALE_TOKENS()`
* `ERROR_STALE_DAYS_INVALID(value)`
* `ERROR_WRITE_FAILED(path,reason)`
* `ERROR_VERIFY_FAILED(residual=[...])`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real `canary scan` executed (strict + update); **no mocked/simulated** results.
* Only `UPDATED` field changed; statuses and other fields preserved exactly.
* HUMAN_TEXT aligns with JSON; includes explicit file lists and counts.
* CANARY snapshots emitted when required.
* JSON returned **without** code fences; schema exact. 

### 10) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## Stale Token Update Results
**Update Date:** 2025-10-17 • **Threshold:** 30 days

**Tokens Updated:** 2
- ✅ {{.ReqID}}-<ASPECT>-API-001 — src/api/auth.go (2024-01-01 → 2025-10-17)
- ✅ {{.ReqID}}-<ASPECT>-Engine-004 — internal/cache/cache.go (2024-01-01 → 2025-10-17)

**Files Modified:** src/api/auth.go • internal/cache/cache.go

**Verification:** Passed — no stale tokens remaining in TESTED/BENCHED.

**Next Steps**
1) git add src/api/auth.go internal/cache/cache.go
2) git commit -m "chore: update stale CANARY tokens (2 files)"
3) canary scan --verify GAP_ANALYSIS.md --strict
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional HUMAN_TEXT with begin/end markers ensures parseable, auditable results for CI. 
* **Section‑delimited structure:** explicit inputs → gates → behavior → outputs → examples increases reliability and maintainability. 
* **Parallel pipeline:** explicit **DAG + concurrency groups** for scan/preview/apply/verify improves speed and reduces ambiguity about when writes occur. 
* **No‑mock/no‑simulate:** runtime guarantees make clear that only real scanner outputs drive updates.

### Assumptions & Risks

* `canary scan --update-stale` edits files in‑place and restricts mutations to `UPDATED=`; if not, the Diff Integrity Gate will fail.
* Some repos may keep `UPDATED` in non‑UTC format; this prompt standardizes to `YYYY‑MM‑DD`.
* If the repo is large, consider chunking or tighter skip patterns on the scanner invocation to avoid long I/O.

### Targeted questions (for fit)

1. Confirm whether `stale_days` is always **30** (Article VII) or should be a configurable constitutional constant.
2. Should we require a **clean git working tree** or auto‑create a branch before updates?
3. Do you want a `--dry-run` flag that prints the preview and exits with non‑zero if stale tokens exist?
4. Keep CANARY snapshot threshold at **70%** context usage?
