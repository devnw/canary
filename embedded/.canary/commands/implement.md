## ImplementCmd (Generate Implementation Guidance)

```yaml
---
description: Generate comprehensive, test-first implementation guidance for a CANARY requirement with strict, verifiable outputs (no-mock/no-simulate)
command: ImplementCmd
version: 2.4
subcommands: [implement, list]
outputs:
  - implementation_prompt: STDOUT (bounded markdown, see Output Contract)
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  test_first_required: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  db_path: .canary/canary.db
  specs_root: .canary/specs
  constitution_path: .canary/memory/constitution.md
  template_path: .canary/templates/implement-prompt-template.md
  fuzzy: {min_score: 60, auto_select_score: 80, min_lead: 20}
  include_plan_if_missing: false    # if plan.md missing, flag; do not fabricate
  max_prompt_kb: 200                # guardrail for context size
---
```

<!-- CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Modes:

  * **Exact ID**: `<REQ-ID>` (e.g., `{{.ReqID}}-<ASPECT>-API-105`)
  * **Feature name**: `UserAuthentication`
  * **Fuzzy**: `"user auth"` (Levenshtein + substring + abbreviation)
  * **List**: `--list` (unimplemented only: STUB|IMPL)

### 2) Preconditions & Resolution

1. **DB gate:** ensure DB at `db_path` readable; else `ERROR_DB_MISSING(path)` → suggest `canary index`.
2. **Template gate:** `template_path` must exist; else `ERROR_TEMPLATE_MISSING(path)`.
3. **Spec/plan/constitution gates:**

   * Spec directory: `${specs_root}/<REQ-ID>-<slug>/spec.md` required; else `ERROR_SPEC_MISSING(path)`.
   * Plan optional: include only if file exists; **never invent**; else set `plan_included=false`.
   * Constitution optional: include only if file exists.
4. **Fuzzy selection policy:**

   * Compute score ∈ [0,100]. If `top.score ≥ 80` **and** `top.score - second.score ≥ 20` → auto‑select.
   * If `score ≥ 60` but ambiguous → return **CHOICE_SET** (non‑interactive) with ranked options.
   * If `< 60` → `ERROR_NO_MATCH(query)` with remediation.
5. **Exact & feature modes:** resolve to a single spec or fail with `ERROR_AMBIGUOUS(results)`.

### 3) Planning & Parallelism

Construct a **Work DAG** and run independent steps concurrently; **join** before assembly. If true parallelism isn’t available, interleave non‑blocking tasks while preserving joins. 

* **CG‑1 Resolve**: identify candidate REQ(s) (exact/feature/fuzzy) and rank.
* **CG‑2 Load**: read `spec.md`, `plan.md?`, `constitution.md?`, and compute **Progress** from tokens.
* **CG‑3 Assemble**: render implementation prompt from template with real contents (no placeholders), then size‑check.
* **CG‑4 Validate**: run gates (below) and produce machine‑readable `summary_json`.

### 4) Behavior (must do; never simulate)

* **Never fabricate** spec/plan/constitution/progress. If missing, **report** and set flags; do not “mock” content.
* **Test‑First bias:** include concrete TDD steps; ensure tests precede implementation.
* **Token discipline:** include exact token examples and update guidance; provide file:line for examples only if available.
* **Context guard:** if combined prompt > `max_prompt_kb`, return `ERROR_PROMPT_TOO_LARGE(kb)` with reduction advice (e.g., omit images, include plan summary only).

### 5) Implementation Prompt — Required Sections (exact headings)

When generating the **implementation prompt**, render these sections **in order** (delimited for reliability). 

1. **Specification Details** — verbatim `spec.md` (no edits).
2. **Implementation Plan** — full `plan.md` **if present**; else a one‑line notice: “Plan not found; run `/canary.plan <REQ-ID>`.”
3. **Implementation Checklist** — extract from plan/spec if present; else omit (do not fabricate).
4. **Progress Tracking** — computed token counts by `STATUS` (STUB/IMPL/TESTED/BENCHED), with file:line where available.
5. **Constitutional Principles** — verbatim `constitution.md` **if present**.
6. **Test‑First Guidance (Article IV)** — concise TDD steps (red→green→refactor) and CI hints.
7. **CANARY Token Examples** — real examples from repo or canonical template snippets (template‑based, clearly labeled).

> Section delimiting and explicit output formats improve reliability and downstream parsing. 

### 6) Validation Gates (compute & report)

* **Article I Gate:** Spec present and REQ token references valid.
* **Article IV Gate:** Test‑first steps present; checklist starts with tests.
* **Article VII Gate:** UPDATED fields present in shown tokens; docs currency mentioned.
* **Spec Clarity Gate:** Spec contains **no** `[NEEDS CLARIFICATION]`.
* **Consistency Gate:** Progress tallies equal totals; IDs consistent across sections.
* **Size Gate:** Output ≤ `max_prompt_kb`; else fail with guidance.

### 7) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **load**, and after **assembly**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"implement|load|assemble|verify",
  "f":[["<spec_path>",1,999],["<plan_path?>",1,999],["<constitution_path?>",1,999]],
  "k":["req:<REQ-ID>","mode:<exact|feature|fuzzy|list>","score:<N?>","prompt_kb:<N>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["emit prompt","follow TDD","update tokens"]
}'
```

*Compact keys minimize tokens while preserving filenames, line spans, key facts, false‑positives, issues, and next steps.* 

### 8) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** 

**A. IMPLEMENTATION_PROMPT (Markdown)**
Begin with: `=== IMPLEMENTATION_PROMPT BEGIN ===`
Required sub‑markers (include only when content exists):

* `<<< SPEC BEGIN >>>` … `<<< SPEC END >>>`
* `<<< PLAN BEGIN >>>` … `<<< PLAN END >>>`
* `<<< CHECKLIST BEGIN >>>` … `<<< CHECKLIST END >>>`
* `<<< PROGRESS BEGIN >>>` … `<<< PROGRESS END >>>`
* `<<< CONSTITUTION BEGIN >>>` … `<<< CONSTITUTION END >>>`
* `<<< TDD BEGIN >>>` … `<<< TDD END >>>`
* `<<< TOKENS BEGIN >>>` … `<<< TOKENS END >>>`
  End with: `=== IMPLEMENTATION_PROMPT END ===`

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "mode": "exact|feature|fuzzy|list",
  "selection": {
    "query": "<raw>",
    "req_id": "{{.ReqID}}-<ASPECT>-API-105",
    "feature": "UserAuthentication",
    "score": 92,
    "auto_selected": true,
    "choices": [
      {"req_id":"{{.ReqID}}-<ASPECT>-API-110","feature":"OAuthIntegration","score":80}
    ]
  },
  "paths": {
    "spec": ".canary/specs/{{.ReqID}}-<ASPECT>-API-105-user-auth/spec.md",
    "plan": ".canary/specs/{{.ReqID}}-<ASPECT>-API-105-user-auth/plan.md",
    "constitution": ".canary/memory/constitution.md"
  },
  "included": {"spec": true, "plan": true, "checklist": true, "constitution": false},
  "progress": {"total": 0, "stub": 0, "impl": 0, "tested": 0, "benched": 0},
  "gates": {
    "article_I": "pass|fail",
    "article_IV": "pass|fail",
    "article_VII": "pass|fail",
    "spec_clarity": "pass|fail",
    "consistency": "pass|fail",
    "size": "pass|fail"
  },
  "prompt_kb": 0,
  "canary": {"emitted": true, "last_id": "<id-or-n/a>"}
}
```

### 9) Subcommand: `--list` (unimplemented only)

* **Behavior:** List requirements with `status ∈ {STUB, IMPL}` (priority‑first, updated_at desc).
* **Output:** same **SUMMARY_JSON** envelope but with `mode="list"` and `items:[{req_id,feature,status,priority,updated_at}]`.
* **HUMAN_TEXT** (optional): present top N with next‑actions (`/canary.plan` for STUB; add tests for IMPL).
* **Never simulate** counts; query real DB.

### 10) Failure Modes (return one with reason + remediation)

* `ERROR_DB_MISSING(path)` → run `canary index`
* `ERROR_TEMPLATE_MISSING(path)`
* `ERROR_NO_MATCH(query)`
* `ERROR_AMBIGUOUS(results=[...])`
* `ERROR_SPEC_MISSING(path)`
* `ERROR_PROMPT_TOO_LARGE(kb)`
* `ERROR_FILE_IO(path,reason)`
* `ERROR_PARSE_OUTPUT(reason)`

### 11) Quality Checklist (auto‑verify before output)

* Real files read; **no mocks/simulations**.
* Fuzzy policy applied; ambiguity surfaced via `choices`.
* All section markers present/ordered; JSON schema exact.
* TDD steps present; token examples real or template‑labeled.
* Progress tallies consistent; gates computed.
* CANARY snapshots emitted when required. 

### 12) Example HUMAN_TEXT (operator‑friendly)

`````
=== IMPLEMENTATION_PROMPT BEGIN ===
<<< SPEC BEGIN >>>
# {{.ReqID}}-<ASPECT>-API-105 — UserAuthentication
…(verbatim spec.md)…
<<< SPEC END >>>
<<< PLAN BEGIN >>>
…(plan.md if present; otherwise single-line notice to create plan)…
<<< PLAN END >>>
<<< CHECKLIST BEGIN >>>
- [ ] Write TestUserAuthentication (red)
- [ ] Implement login flow (green)
- [ ] Refactor & add docs
<<< CHECKLIST END >>>
<<< PROGRESS BEGIN >>>
Total: 8 • TESTED: 2 • IMPL: 3 • STUB: 3
Top files: src/api/auth/user.go:45, src/api/auth/middleware.go:23
<<< PROGRESS END >>>
<<< TDD BEGIN >>>
1) Create failing tests → 2) Implement minimal code → 3) Refactor w/ tests green
<<< TDD END >>>
<<< TOKENS BEGIN >>>
```go
// CANARY: REQ={{.ReqID}}-<ASPECT>-API-105; FEATURE="UserAuthentication"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-17
```
<<< TOKENS END >>>
=== IMPLEMENTATION_PROMPT END ===
`````

---

### What changed & why (brief)
- **Deterministic outputs:** begin/end markers + strict **SUMMARY_JSON** enable robust parsing and CI checks. :contentReference[oaicite:9]{index=9}  
- **Parallel pipeline:** explicit **DAG + concurrency groups** for resolve/load/assemble/validate improves speed and reliability. :contentReference[oaicite:10]{index=10}  
- **No‑mock/no‑simulate:** codified as **runtime_guarantees**; missing artifacts are reported—not fabricated.  
- **Section delimiting & structure:** improves maintainability and minimizes ambiguity, consistent with prompt‑quality guidance. :contentReference[oaicite:11]{index=11}

### Assumptions & Risks
- Token progress comes from real scans/DB; line numbers may be absent—if so, set `null` (never invent).  
- Very large specs may exceed `max_prompt_kb`; caller should prefer linking or chunked flows.  
- Fuzzy thresholds can be tuned per repo; defaults provided.

### Targeted questions (for fit)
1) Confirm canonical **REQ‑ID** pattern and slug naming.  
2) Should we **always** inline `plan.md`, or just include a summary when prompt size is tight?  
3) Do we include **doc excerpts** from `/canary.doc report` into PROGRESS when docs are stale?  
4) Keep CANARY snapshot threshold at **70%** context usage?
