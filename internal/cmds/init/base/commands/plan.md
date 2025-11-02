## PlanCmd (Technical Implementation Plan Generator)

```yaml
---
description: Generate a technical implementation plan from a CANARY requirement specification (strict, verifiable, no-mock/no-simulate)
command: PlanCmd
version: 2.1
outputs:
  - plan_markdown: .canary/specs/<REQ_ID>-SECURITY_REVIEW-<slug>/plan.md
  - summary_json: STDOUT (unwrapped JSON, strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  test_first_required: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
---
```

<!-- CANARY: REQ=CBIN-113; FEATURE="PlanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS`
  Parse into:
  `req_id?` (e.g., `{{.ReqID}}-SECURITY_REVIEW-113`) • `preferences?` (tech stack, perf/security constraints) • `notes?`.
* **Repository Layout (assumed):**

  * Specs: `.canary/specs/<REQ_ID>-<slug>/spec.md`
  * Constitution: `.canary/memory/constitution.md`
  * Tracking: `.canary/requirements.md`

### 2) Preconditions & Resolution

1. **Identify requirement:**

   * If `req_id` provided: use it.
   * If empty: select most recent spec with `STATUS=STUB`.
   * **Fail fast** with `ERROR_NOT_FOUND(req_id|stub)` if no spec is resolvable.
2. **Load & validate spec:** must exist and contain **no** `"[NEEDS CLARIFICATION]"`.

   * On failure, return `ERROR_UNCLARIFIED(markers=[...])` listing exact line numbers.
3. **Load constitution** and cache applicable principles (e.g., Test‑First, Simplicity, Reliability, Security).

### 3) Planning Policy (what you must do)

* **No mock, no simulate:** Do not propose placeholder data or feigned results. Every step must be implementable by `/canary.implement`.
* **Parallelism awareness:** Include a **Work DAG** and **Concurrency Groups** (CG‑1, CG‑2, …) showing tasks that can proceed in parallel (e.g., test scaffolding vs. schema draft), and explicit **Join Points** before shared writes or integration.
* **Token discipline:** Specify **exact CANARY token(s)** to insert/update (paths + line numbers once implemented).
* **Security & Performance:** Always include a minimal threat model + perf targets (or `n/a` with rationale).
* **Observability:** Define metrics/logging needed to verify acceptance criteria.

### 4) Plan File Creation

* **Write** the plan to: `.canary/specs/<REQ_ID>-<slug>/plan.md` (create folder if absent).
* **Return** both: (A) the full Markdown plan content and (B) a strict **JSON summary** (schema below).

### 5) Required Plan Sections (exact headings)

1. **Title:** `# Implementation Plan: <REQ_ID> <FeatureName>`
2. **Tech Stack Decision** — chosen tech with rationale; pin versions where applicable.
3. **Architecture Overview** — components, interfaces, data flow (bulleted diagram text is fine).
4. **Work DAG & Concurrency Groups** — CG‑N lists parallel tasks; specify join barriers.
5. **CANARY Token Placement** — file paths + **line ranges TBD**; token evolution: `STUB → IMPL → TESTED → BENCHED`.
6. **Implementation Phases (Test‑First)**

   * *Phase 0: Pre‑Implementation Gates* (constitution checks).
   * *Phase 1: Tests* (red) — enumerate unit/integration/acceptance tests and expected failing messages.
   * *Phase 2: Implementation* (green) — minimal increments to pass tests.
   * *Phase 3: Benchmarks* (if perf‑critical) — target metrics and harness.
   * *Phase 4: Hardening* — security, resiliency, observability.
7. **Testing Strategy** — coverage targets, data fixtures (realistic, no mocks unless explicitly allowed by spec), CI notes.
8. **Risk & Mitigations** — top 3 risks, rollback plan, feature flags.
9. **Constitutional Gates** — map plan to each applicable article with pass/fail criteria.
10. **Ready‑to‑Implement Checklist** — short, binary items that `/canary.implement` can execute.

### 6) Validation Gates (must compute and report)

* **Article I Gate:** CANARY token(s) precisely specified (path + intent).
* **Article IV Gate:** Tests precede implementation; acceptance criteria explicit.
* **Article V Gate:** Simplicity kept; justify any additional dependency.
* **Article VI Gate:** Integration testing approach documented.
* **Security Gate:** Threats + mitigations listed or justified `n/a`.
* **Performance Gate:** Targets + how to measure (or justified `n/a`).

### 7) CANARY Snapshot Protocol (compact; low‑token)

Emit a snapshot when **context ≥70%**, after **spec load**, and post‑**plan write**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"plan|verify",
  "f":[[".canary/specs/<REQ_ID>-<slug>/spec.md",1,999],[".canary/specs/<REQ_ID>-<slug>/plan.md",1,999]],
  "k":["req:<REQ_ID>","feature:<FeatureName>","tests:first","parallel:CG-1..N"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["write plan.md","validate gates","update tracking"]
}'
```

**Fields:** `t` time • `s` stage • `f` file+line spans • `k` key facts • `fp` false‑positives to avoid retry failures • `iss` issues • `nx` next actions.
*(Compact keys minimize tokens.)*

### 8) Tracking Update

Append or edit in `.canary/requirements.md`:

* `- [ ] <REQ_ID> - <FeatureName> (STATUS=STUB → ready for implementation)`

### 9) Output Contract (strict)

Return **both** artifacts in this order:

**A. PLAN_MARKDOWN**
Start with the exact line:
`=== PLAN_MARKDOWN BEGIN ===`
Then the complete Markdown plan.
End with:
`=== PLAN_MARKDOWN END ===`

**B. SUMMARY_JSON** *(raw JSON; no code fences)* — must conform to schema:

```json
{
  "req_id": "{{.ReqID}}-SECURITY_REVIEW-XXX",
  "feature_name": "PlanCmd",
  "plan_path": ".canary/specs/{{.ReqID}}-SECURITY_REVIEW-XXX-<slug>/plan.md",
  "status": "ready-for-implementation",
  "gates": {
    "article_I": "pass|fail",
    "article_IV": "pass|fail",
    "article_V": "pass|fail",
    "article_VI": "pass|fail",
    "security": "pass|n/a|fail",
    "performance": "pass|n/a|fail"
  },
  "parallelism": { "groups": ["CG-1", "CG-2", "CG-3"], "has_joins": true },
  "risks": ["..."],
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 10) Failure Modes (return one, with reason + remediation)

* `ERROR_NOT_FOUND(req_id|stub)`
* `ERROR_UNCLARIFIED(markers=[{line,excerpt}])`
* `ERROR_WRITE_FAILED(path,reason)`
* `ERROR_CONSTITUTION(load|violation)`
* `ERROR_UNSUPPORTED_ARG(arg_name)`

### 11) Quality Checklist (auto‑verify before output)

* Tech decisions justified • Work DAG present with parallel groups • Token placement explicit
* Test‑first phases complete • Gates evaluated • Security/Perf addressed (or `n/a` with rationale)
* Plan is implementable by `/canary.implement` with no mocks or simulated steps
* CANARY snapshot emitted (when required)

```

---

### What changed & why (brief)
- **Deterministic outputs**: explicit **BEGIN/END** markers + **strict JSON** summary enable downstream parsing. :contentReference[oaicite:3]{index=3}  
- **Section delimiting + structure** tightened for clarity and maintainability. :contentReference[oaicite:4]{index=4}  
- **Parallelism requirement** made first‑class (Work DAG + CGs + joins), aligning with complex‑task prompting guidance. :contentReference[oaicite:5]{index=5}  
- **No‑mock/no‑simulate** guarantees moved to **runtime_guarantees** to avoid ambiguity at execution time. :contentReference[oaicite:6]{index=6}  
- **CANARY snapshots** standardized with a **compact JSON** schema to minimize token use while capturing filenames, line ranges, key facts, false‑positives, and next steps.

---

### Assumptions & Risks
- Assumes `.canary` layout is authoritative and the `canary` CLI is available on PATH.  
- If strict parallel tool execution isn’t supported, the agent should **interleave** non‑blocking steps while preserving join points. :contentReference[oaicite:7]{index=7}

---

### Targeted questions (for final fit)
1) Confirm the **context threshold** for CANARY snapshots (default 70%).  
2) Are **constitution article names** canonical and stable? If additional gates exist, list them.  
3) Should **SUMMARY_JSON** include links to CI jobs or issue IDs?  
4) Any **default tech stack** (language/framework) we should assume when the spec is silent?
