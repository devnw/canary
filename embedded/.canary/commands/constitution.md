## ConstitutionCmd (Project Governance & CANARY Guidelines)

```yaml
---
description: Create or update the project constitution (governing principles + CANARY development guidelines) with strict validation and verifiable outputs
command: ConstitutionCmd
version: 2.0
outputs:
  - constitution_markdown: .canary/memory/constitution.md
  - summary_json: STDOUT (unwrapped JSON; schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
---
```

<!-- CANARY: REQ=CBIN-109; FEATURE="ConstitutionCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into: `mode` (`create|update|amend` inferred), `edits?` (target articles/sections), `project_principles?` (bullets), `rationale?`, `effectivity? (ISO date)`.
* **Repo layout (assumed):** `.canary/memory/constitution.md` (authoritative), `.canary/specs/**` (cross‑refs).

### 2) Preconditions & Resolution

1. **Detect mode:**

   * **Empty args ⇒ `create`** (bootstrap from template).
   * **Args with article refs ⇒ `update`** (edit in place).
   * **Args with version/effective date ⇒ `amend`** (append Amendment entry and apply edits).
2. **Load existing constitution** if present; else `create` from template.
3. **Fail fast**:

   * Missing/unknown article IDs ⇒ `ERROR_ARTICLE_NOT_FOUND(id)`.
   * Edits that conflict or duplicate IDs ⇒ `ERROR_CONFLICT(details)`.
   * Non‑actionable text (no measurable criteria) ⇒ `ERROR_NON_ACTIONABLE(section)`.

### 3) Authoring Policy (what you must do)

* **No mock / no simulate:** Write the **real** constitution file; do not produce placeholders that pretend to exist elsewhere.
* **Actionable & testable:** Each article MUST include **Enforcement** and **Measurables** (objective criteria).
* **Compatibility:** Preserve core articles and numbering; project‑specific articles start at **Article X** upward.
* **Traceability:** Cross‑reference enforcement to `/canary.plan`, `/canary.doc`, `/canary.verify` gates.
* **Parallel edit safety:** If multiple articles are updated, prepare edits per‑article in parallel and **join** before write to avoid merge races. (Interleave if true parallelism is unavailable.) 

### 4) Constitution File — Required Structure (exact headings)

Write to `.canary/memory/constitution.md` with these sections:

1. `# Project Constitution` — Title + short **Preamble** (purpose, scope).
2. `## Articles` — ordered list. **Core (must exist):**

   * **Article I — Requirement‑First** (requirements precede work; tokens required).
   * **Article IV — Test‑First** (tests precede implementation; red/green discipline).
   * **Article VII — Documentation Currency** (DOC/DOC_HASH must be current; UPDATED ≤ 30 days).
     *(Include additional core project articles you already use, e.g., Simplicity, Security, Performance, Observability, Change Control. Keep numbering stable.)*
3. `## Project‑Specific Articles (Article X, XI, …)` — add per user/domain needs; each article MUST include:

   * **Policy** (normative text)
   * **Enforcement** (which command/gate verifies it)
   * **Measurables** (KPIs/SLOs or binary checks)
   * **Rationale** (why this exists; brief)
4. `## Enforcement Matrix` — table mapping **Article → Command/Gate** (e.g., Art IV → `/canary.plan` Phase 1 gate; Art VII → `/canary.doc status`).
5. `## Versioning & Amendments` — version number, effective date, changelog of article deltas.
6. `## Glossary & IDs` — canonical article IDs; keep stable for tool references.

> **Section delimiting & structured outputs** increase reliability and downstream parsing. Keep headings stable and concise.  

### 5) Creation/Update Behavior

* **Create:** Render full template with required core articles (I, IV, VII) + empty “Project‑Specific Articles” section + initial **Enforcement Matrix** and **Amendments v1.0** (effective today).
* **Update:** For each targeted article/section:

  * Validate: no contradictions; measurable criteria present; numbering preserved.
  * Apply edits; update Enforcement Matrix if needed.
  * Append an **Amendment** entry with version bump and rationale.
* **Amend:** Shortcut that both appends an Amendment and applies its edits atomically.

### 6) Validation Gates (compute & report)

* **Actionability Gate:** Every article has **Enforcement** + **Measurables**.
* **Non‑Contradiction Gate:** No article conflicts with another; if found, list pairs (A↔B).
* **Enforcement Coverage Gate:** Every article appears in the **Enforcement Matrix** (1+: `/canary.plan`, `/canary.doc`, `/canary.verify`).
* **Measurability Gate:** KPIs/SLOs present or explicit `n/a` with rationale.
* **Core Presence Gate:** Articles I, IV, VII present and non‑empty.
* **Versioning Gate:** Version increased when substantive changes occur; effective date in ISO 8601.

### 7) CANARY Snapshot Protocol (compact; low‑token)

Emit a snapshot when **context ≥70%**, after **load**, and after **write**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"constitution|plan|verify",
  "f":[[".canary/memory/constitution.md",1,999]],
  "k":["mode:<create|update|amend>","version:<x.y>","core:I,IV,VII","matrix:present"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["publish","notify","enforce in plan/doc/verify"]
}'
```

*Compact keys minimize tokens while preserving filenames, line spans, key facts, false‑positives, issues, next steps.* 

### 8) Output Contract (strict)

Return artifacts in this order:

**A. CONSTITUTION_MARKDOWN**
Start with the exact line:
`=== CONSTITUTION_MARKDOWN BEGIN ===`
…then the complete Markdown file content…
End with:
`=== CONSTITUTION_MARKDOWN END ===`

**B. SUMMARY_JSON** *(raw JSON; no code fences)* — schema:

```json
{
  "ok": true,
  "mode": "create|update|amend",
  "path": ".canary/memory/constitution.md",
  "version": "1.0",
  "effective": "2025-10-17",
  "core_articles": {"I": true, "IV": true, "VII": true},
  "articles_total": 0,
  "gates": {
    "actionability": "pass|fail",
    "non_contradiction": "pass|fail",
    "enforcement_coverage": "pass|fail",
    "measurability": "pass|n/a|fail",
    "core_presence": "pass|fail",
    "versioning": "pass|fail"
  },
  "enforcement_matrix_present": true,
  "amendment": {"added": true, "id": "Amend-2025-10-17", "summary": "<text>"},
  "canary": {"emitted": true, "last_id": "<id-or-n/a>"}
}
```

> Use **structured, strict JSON** for reliable automation; do not wrap in code fences. 

### 9) Failure Modes (return one with reason + remediation)

* `ERROR_ARTICLE_NOT_FOUND(id)`
* `ERROR_CONFLICT(details)`
* `ERROR_NON_ACTIONABLE(section)`
* `ERROR_FILE_IO(path,reason)`
* `ERROR_VERSIONING(required_state)`
* `ERROR_TEMPLATE_MISSING()`

### 10) Quality Checklist (auto‑verify before output)

* Core articles present; numbering stable; **Enforcement** + **Measurables** exist per article.
* Enforcement Matrix covers all articles; cross‑refs to `/canary.plan`, `/canary.doc`, `/canary.verify`.
* Version bumped if substantive change; effective date set; **Amendments** updated.
* JSON summary conforms to schema; **no code‑fence wrapping**. 
* CANARY snapshot emitted at required milestones. 

### 11) Example Article Snippets (for authoring clarity)

```markdown
## Article I: Requirement‑First
**Policy:** All work must originate from a CANARY requirement with a valid token.  
**Enforcement:** `/canary.plan` must reference REQ and refuse planning if missing.  
**Measurables:** 100% of merged PRs reference a valid REQ token.

## Article IV: Test‑First
**Policy:** Write failing tests before implementation; keep red→green loop.  
**Enforcement:** `/canary.plan` Phase 1 gate; CI blocks merges without failing‑then‑passing sequence.  
**Measurables:** Coverage ≥ ‹threshold›; each new feature adds at least one failing test initially.

## Article VII: Documentation Currency
**Policy:** Documentation must remain current; stale tokens (>30 days) flagged.  
**Enforcement:** `/canary.doc status --all` before release; block if DOC_STALE/MISSING.  
**Measurables:** 0 stale docs at release; DOC/DOC_HASH aligned.
```

> Section delimiting and explicit output formats improve reliability and downstream parsing.  

---

### What changed & why (brief)

* **Deterministic outputs**: file + strict **SUMMARY_JSON** enables CI enforcement and auditing. 
* **Validation gates**: actionability, contradictions, coverage, measurability enforce governance quality.
* **Compact CANARY snapshots**: low‑token state capture with filenames, line spans, key facts, false‑positives, next steps. 
* **Parallel edit safety**: concurrency groups + join points for multi‑article updates; interleave when parallelism unavailable. 

### Assumptions & Risks

* Constitution path `.canary/memory/constitution.md` is authoritative; tools have write access.
* Article IDs are stable; downstream tools rely on headings.
* If the project already defines articles II/III/V/VI/etc., retain numbering and update Enforcement Matrix accordingly.

### Targeted questions (for fit)

1. Confirm canonical list/numbers of **core articles** beyond I, IV, VII.
2. Confirm **default KPIs/SLOs** for measurables (coverage thresholds, P95 latency, etc.).
3. Do you require **sign‑off roles** (e.g., Architect, QA) recorded in Amendments?
4. Keep CANARY snapshot threshold at **70%** context usage?
