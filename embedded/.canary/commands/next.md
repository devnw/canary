## NextCmd (Auto‑Select Next Priority & Generate Guidance)

```yaml
---
description: Identify and (optionally) generate full implementation guidance for the next highest‑priority CANARY requirement (strict outputs; no‑mock/no‑simulate)
command: NextCmd
version: 2.4
subcommands: [next]
outputs:
  - human_text: STDOUT (concise operator view; optional unless explicitly requested)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
  - implementation_prompt: STDOUT (only when --prompt; bounded markdown per contract)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  test_first_required: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  specs_root: .canary/specs
  constitution_path: .canary/memory/constitution.md
  limit_candidates: 5
  include_blocked: false      # if true, show blocked with reasons; still select first unblocked
  age_boost_days: 30          # aging bucket for priority boost
  scoring_weights:
    priority: 1.0             # lower numeric priority (1 is highest) is better
    status: 0.6               # STUB>IMPL>TESTED>BENCHED>REMOVED
    age: 0.3                  # older UPDATED gets boost
  max_prompt_kb: 200          # guardrail when emitting --prompt
filters_defaults:
  status: []                  # e.g., ["STUB","IMPL"]
  aspect: []                  # e.g., ["API","CLI",...]
  owner: []
  phase: []
  spec_status: []
---
```

<!-- CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse flags: `--prompt` (emit full guidance), `--json` (machine‑readable only), `--status`, `--aspect`, `--owner`, `--phase`, `--spec-status`, `--db <path>`.
  Default selection: highest‑ranked **unblocked** requirement among `status ∈ {STUB, IMPL}`.

### 2) Preconditions & Resolution

1. **DB gate:** ensure DB exists/readable at `db_path`; else `ERROR_DB_MISSING(path)` with remediation (`canary index`) and **filesystem fallback** scan of `.canary/specs/**` if available.
2. **Command availability:** ensure `canary next`/`canary list`/`canary scan` exist; else `ERROR_CANARY_NEXT_UNAVAILABLE()`.
3. **Spec presence:** selected REQ must have `spec.md`; else skip and promote next candidate; if none, `ERROR_NO_AVAILABLE_REQUIREMENTS(filters)`.
4. **Constitution (optional):** include only if `constitution.md` exists; never fabricate.
5. **Dependency gating:** if a candidate has `DEPENDS_ON` unmet, mark **blocked** (list missing REQ‑IDs) and skip unless `include_blocked=true`.

### 3) Priority & Ranking Model (deterministic; compute, don’t guess)

For each candidate i:

* **status_weight(i):** STUB=1.0, IMPL=0.6, TESTED=0.2, BENCHED=0.1, REMOVED=0.0
* **age_score(i):** `min(1.0, days_since_updated(i)/age_boost_days)`
* **priority_score(i):** `w_p*(1/priority_i) + w_s*status_weight + w_a*age_score` (use `scoring_weights`; configurable)
* **blocked:** if any `DEPENDS_ON` unmet ⇒ exclude from selection (unless `include_blocked=true`) and record in `blocked_candidates`.
* **tiebreakers:** (1) higher `status_weight`, (2) older `updated_at`, (3) lexicographic `req_id`.

### 4) Planning & Parallelism

Create a **Work DAG** and run independent steps concurrently; **join** before final assembly. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Query:**

  ```bash
  canary next --json [filters] || canary list --json [filters]
  ```

  (Fallback to scan when DB missing.)
* **CG‑2 Rank:** compute scores, dependency checks (read `DEPENDS_ON` from spec or DB).
* **CG‑3 Load (selected):** read `spec.md`, optional `plan.md`, optional `constitution.md`; compute token **Progress** via `canary scan --project-only`.
* **CG‑4 Assemble:** emit **HUMAN_TEXT** + **SUMMARY_JSON**; if `--prompt`, render **Implementation Prompt** using the same markers as ImplementCmd (see Output Contract).

### 5) Behavior (must do; never simulate)

* **Run real queries/scans**; do **not** fabricate rows, specs, or progress.
* **Test‑first bias:** guidance must show tests before code (Article IV).
* **Token discipline:** show canonical CANARY token examples; if real file:line is available, include it; otherwise leave `line=null` (do not invent).
* **Dependencies:** state unmet prerequisites explicitly; if all top candidates are blocked, select next unblocked and include a `blocked_candidates` list.
* **Size guard:** if `--prompt` exceeds `max_prompt_kb`, return `ERROR_PROMPT_TOO_LARGE(kb)` with reduction advice.

### 6) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **rank**, and after **assemble**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"next|rank|assemble",
  "f":[["<db_or_scan_source>",1,1],["<spec_path?>",1,999]],
  "k":["filters:<...>","candidates:<N>","blocked:<B>","selected:<REQ-ID>","prompt:<bool>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["open spec","follow TDD","/canary.plan if needed"]
}'
```

*Compact keys minimize tokens while preserving filenames, key facts, false‑positives, and next steps.* 

### 7) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** 

**A. HUMAN_TEXT (optional)**
Begin with `=== HUMAN_TEXT BEGIN ===` … end with `=== HUMAN_TEXT END ===`
Contents: title, selected REQ summary, dependency note, TDD reminder, and next‑steps.

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "mode": "summary|prompt",
  "filters": { "status": [], "aspect": [], "owner": [], "phase": [], "spec_status": [] },
  "db_path": ".canary/canary.db",
  "selected": {
    "req_id": "{{.ReqID}}-<ASPECT>-API-134",
    "feature": "UserOnboarding",
    "status": "STUB",
    "aspect": "API",
    "priority": 1,
    "updated_at": "2025-10-16T12:00:00Z",
    "depends_on": ["{{.ReqID}}-<ASPECT>-CORE-010"],
    "deps_satisfied": true,
    "score": 0.00,
    "score_components": { "priority": 0.00, "status": 0.00, "age": 0.00 }
  },
  "ranking": [
    {
      "req_id": "{{.ReqID}}-<ASPECT>-API-134",
      "status": "STUB",
      "deps_satisfied": true,
      "score": 0.00,
      "score_components": { "priority": 0.00, "status": 0.00, "age": 0.00 }
    }
  ],
  "blocked_candidates": [
    { "req_id": "{{.ReqID}}-<ASPECT>-ENGINE-140", "blocked_by": ["{{.ReqID}}-<ASPECT>-DB-020"] }
  ],
  "paths": {
    "spec": ".canary/specs/{{.ReqID}}-<ASPECT>-API-134-user-onboarding/spec.md",
    "plan": ".canary/specs/{{.ReqID}}-<ASPECT>-API-134-user-onboarding/plan.md",
    "constitution": ".canary/memory/constitution.md"
  },
  "progress": { "total": 0, "stub": 0, "impl": 0, "tested": 0, "benched": 0 },
  "recommendations": [
    "/canary.plan {{.ReqID}}-<ASPECT>-API-134",
    "/canary.implement {{.ReqID}}-<ASPECT>-API-134",
    "/canary.doc status --all --stale-only"
  ],
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

**C. IMPLEMENTATION_PROMPT (only when `--prompt`)**
Begin with: `=== IMPLEMENTATION_PROMPT BEGIN ===`
Include the same **delimited sections** used by ImplementCmd:

* `<<< SPEC BEGIN/END >>>` (verbatim)
* `<<< PLAN BEGIN/END >>>` (if present; else single‑line notice not to fabricate)
* `<<< CHECKLIST BEGIN/END >>>` (only if present)
* `<<< PROGRESS BEGIN/END >>>`
* `<<< CONSTITUTION BEGIN/END >>>` (if present)
* `<<< TDD BEGIN/END >>>` (concise test‑first steps)
* `<<< TOKENS BEGIN/END >>>` (examples or canonical template snippet)
  End with: `=== IMPLEMENTATION_PROMPT END ===`

> Section delimiting + structured outputs improve reliability and downstream parsing.  

### 8) Validation Gates (compute & report)

* **DB Gate:** reachable or valid fallback scan performed.
* **Dependency Gate:** `deps_satisfied=true` for selected item.
* **Article I Gate:** spec present before emitting guidance.
* **Article IV Gate:** TDD section present and first in the action list.
* **Article V Gate:** simplicity check (no unnecessary complexity in guidance).
* **Article VII Gate:** docs currency mentioned; UPDATED required in shown tokens.
* **Consistency Gate:** ranking totals coherent; selected ∈ ranking; progress tallies consistent.
* **Size Gate (when --prompt):** prompt ≤ `max_prompt_kb`.

### 9) Failure Modes (return one with reason + remediation)

* `ERROR_DB_MISSING(path)` → suggest `canary index` (fallback scan attempted).
* `ERROR_CANARY_NEXT_UNAVAILABLE()`
* `ERROR_NO_AVAILABLE_REQUIREMENTS(filters)`
* `ERROR_DEPENDENCY_BLOCK(req_id, blocked_by=[...])` (if user forced a specific REQ)
* `ERROR_SPEC_MISSING(path)`
* `ERROR_PROMPT_TOO_LARGE(kb)`
* `ERROR_FILE_IO(path,reason)`
* `ERROR_PARSE_OUTPUT(reason)`

### 10) Quality Checklist (auto‑verify before output)

* Real data (DB/scan) used; **no mocks/simulations**.
* Ranking uses deterministic formula; dependency gating applied.
* HUMAN_TEXT (if any) concise and consistent with JSON.
* If `--prompt`, all section markers present; nothing invented for missing artifacts.
* CANARY snapshot(s) emitted when required; JSON **not** wrapped in code fences. 

### 11) Example HUMAN_TEXT (operator‑friendly)

```
=== HUMAN_TEXT BEGIN ===
## Next Priority: {{.ReqID}}-<ASPECT>-API-134 — UserOnboarding
Status: STUB | Aspect: API | Priority: 1 | Score: 0.74
Dependencies: satisfied
Location: .canary/specs/{{.ReqID}}-<ASPECT>-API-134-user-onboarding/spec.md:1

**Why selected:** Highest priority, STUB status, aged 42d (boost).
**Next steps:** 1) /canary.plan {{.ReqID}}-<ASPECT>-API-134  2) /canary.implement {{.ReqID}}-<ASPECT>-API-134 -- or use --prompt to inline guidance
**TDD reminder (Art IV):** write failing tests first, then implement minimal pass, then refactor.

(Use --json for machine output or --prompt for full guidance.)
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs:** strict **SUMMARY_JSON** + optional **HUMAN_TEXT** + (on demand) **IMPLEMENTATION_PROMPT** with clear section markers, enabling robust parsing and CI checks. 
* **Structured, delimited design:** clearer inputs → gates → behavior → outputs improves maintainability and minimizes ambiguity. 
* **Parallel pipeline & fallback:** explicit DAG + concurrency groups; DB query with safe filesystem fallback; avoids blocking joins. 
* **No‑mock/no‑simulate:** codified as runtime guarantees; missing artifacts are reported—not fabricated.

### Assumptions & Risks

* `canary next --json` returns enough fields to compute scores; otherwise combine `canary list --json` + spec reads.
* `DEPENDS_ON` is discoverable from spec or DB; if not, dependency gate may be conservative.
* Very large specs may exceed `max_prompt_kb` when `--prompt`; prefer links or summaries when large.

### Targeted questions (for fit)

1. Confirm canonical enums for **status/aspect/phase/spec_status** and whether **priority** is always numeric.
2. Confirm exact field carrying dependencies (e.g., `DEPENDS_ON:` list in spec).
3. Should `include_blocked=true` ever select a blocked item (e.g., escalation mode)?
4. Keep CANARY snapshot threshold at **70%** context usage?
