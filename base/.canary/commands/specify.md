## SpecifyCmd (Create New Requirement from Natural‑Language Description)

```yaml
---
description: Create a new CANARY requirement specification from a natural-language feature description, with strict validation and verifiable outputs (no-mock/no-simulate)
command: SpecifyCmd
version: 2.4
scripts:
  sh: .canary/scripts/create-new-requirement.sh
outputs:
  - spec_markdown: .canary/specs/<REQ-ID>-<slug>/spec.md
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  specs_root: .canary/specs
  template_spec: templates/spec-template.md
  requirements_index: .canary/requirements.md
  aspect_vocab: ["API","CLI","Engine","Storage","Security","Docs","Frontend","Data","Infra"]
  req_id_pattern: '^[A-Z]{4,}-[A-Za-z]+-[0-9]{3}$'   # e.g., {{.ReqID}}-<ASPECT>-API-105
  id_lockfile: .canary/.id.lock
  max_clarifications: 3
---
```

<!-- CANARY: REQ=CBIN-110; FEATURE="SpecifyCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS` (natural‑language description).
  **MANDATORY:** If non‑empty, you **must** use it to derive the feature name, aspect, and scope. If empty, return `ERROR_DESCRIPTION_REQUIRED()`.

### 2) Preconditions & Resolution

1. **Paths/Gate:** Ensure `specs_root` and `template_spec` exist; else `ERROR_PATH_MISSING(path)`.
2. **Script/Gate:** Ensure script is executable: `.canary/scripts/create-new-requirement.sh`; else `ERROR_SCRIPT_MISSING(path)`.
3. **Aspect detection:** Classify to one of `aspect_vocab` using keyword cues (e.g., *endpoint/api/oauth*→API; *flags/command/cli*→CLI; *db/cache/storage*→Storage; *auth/security/acl*→Security). If uncertain, default to **API** and note `aspect_confidence="low"`.
4. **ID generation (collision‑safe):**

   * Acquire `id_lockfile`.
   * Scan `specs_root` for existing `{{.ReqID}}-<ASPECT>-<ASPECT>-NNN` for **selected aspect**; pick next NNN (zero‑padded).
   * If directory already exists after selection, increment and retry.
   * Release lock.

### 3) Planning & Parallelism

Build a **Work DAG** and run independent steps concurrently; **join** before shared writes. If true parallelism isn’t available, interleave non‑blocking steps while preserving joins. 

* **CG‑1 Scan IDs:** list existing requirement folders; compute next ID.
* **CG‑2 Derive name/slug:** extract 2–4 word **concise feature name** (keep key terms like OAuth2/JWT); make `kebab-case` slug.
* **CG‑3 Script create:** run shell script to create folder and seed `spec.md`.
* **CG‑4 Author spec:** load `template_spec`; fill **WHAT/WHY** sections; insert ≤ `max_clarifications` `[NEEDS CLARIFICATION]` items. **Never include HOW/tech choices.**
* **CG‑5 Register:** append entry to `requirements_index` (create if missing) with `STATUS=STUB`.

### 4) Behavior (must do; never simulate)

* **Run the real script** (no mock). Command form:

  ```bash
  .canary/scripts/create-new-requirement.sh --req-id <REQ-ID> --feature "<FeatureName>" --aspect <ASPECT>
  ```

  Use the returned `SPEC_FILE` (parse script stdout/stderr safely).
* **Spec authoring rules (WHAT/WHY only):**

  * **Sections (exact headings):**

    1. *Feature Overview* • 2) *User Stories* • 3) *Functional Requirements* • 4) *Success Criteria* (measurable, tech‑agnostic) • 5) *Assumptions & Constraints* • 6) *Open Questions* (≤3 `[NEEDS CLARIFICATION]`).
  * Avoid implementation details (languages/frameworks/db).
  * Write crisp, testable acceptance statements (Given/When/Then allowed).
* **CANARY token proposal (ready to paste):**

  ```
  // CANARY: REQ=<REQ-ID>; FEATURE="<PascalCaseName>"; ASPECT=<ASPECT>; STATUS=STUB; UPDATED=<YYYY-MM-DD>
  ```

  Suggest a **logical file** for placement based on aspect (e.g., API→`src/api/<slug>.go`, CLI→`cmd/<slug>.go`, Storage→`internal/db/<slug>.go`), but **do not fabricate existing paths**; mark as suggestion.
* **Index update:** Append to `requirements_index`:
  `- [ ] <REQ-ID> - <FeatureName> (STATUS=STUB)`

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **ID selection**, and after **spec write**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"specify|id|write",
  "f":[[".canary/specs/<REQ-ID>-<slug>/spec.md",1,999],[".canary/requirements.md",1,999]],
  "k":["req:<REQ-ID>","aspect:<ASPECT>","name:<FeatureName>","slug:<slug>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["add token in code","/canary.plan <REQ-ID>"]
}'
```

*Compact keys capture filenames, line spans, key facts, false‑positives, and next steps while minimizing tokens.* 

### 6) Output Contract (strict)

Return artifacts in this order. **Do not wrap JSON in code fences.** Use structured outputs for reliable downstream automation. 

**A. SPEC_MARKDOWN**
Start with the exact line:
`=== SPEC_MARKDOWN BEGIN ===`
…then the full `spec.md` content you wrote (WHAT/WHY only)…
End with:
`=== SPEC_MARKDOWN END ===`

**B. SUMMARY_JSON (unwrapped JSON)** — schema:

```json
{
  "ok": true,
  "req_id": "{{.ReqID}}-<ASPECT>-API-106",
  "aspect": "API",
  "feature_name": "User Authentication",
  "slug": "user-authentication",
  "paths": {
    "spec_dir": ".canary/specs/{{.ReqID}}-<ASPECT>-API-106-user-authentication",
    "spec_file": ".canary/specs/{{.ReqID}}-<ASPECT>-API-106-user-authentication/spec.md",
    "requirements_index": ".canary/requirements.md"
  },
  "token_suggestion": "// CANARY: REQ={{.ReqID}}-<ASPECT>-API-106; FEATURE=\"UserAuthentication\"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-17",
  "suggested_code_location": "src/api/user-authentication.go",
  "gates": {
    "description_present": "pass|fail",
    "id_unique": "pass|fail",
    "template_loaded": "pass|fail",
    "what_why_only": "pass|fail",
    "clarifications_lte_max": "pass|fail",
    "index_updated": "pass|fail"
  },
  "script": {
    "cmd": ".canary/scripts/create-new-requirement.sh --req-id {{.ReqID}}-<ASPECT>-API-106 --feature \"User Authentication\" --aspect API",
    "exit_code": 0,
    "stdout_tail": "<last-200-bytes>",
    "stderr_tail": "<last-200-bytes>"
  },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **Description Gate:** non‑empty `$ARGUMENTS`.
* **ID Uniqueness Gate:** no existing folder for chosen `<REQ-ID>`.
* **Template Gate:** `template_spec` readable.
* **WHAT/WHY Gate:** forbid technology choices and step‑by‑step implementation details.
* **Clarification Gate:** ≤ `max_clarifications` occurrences of `[NEEDS CLARIFICATION]`.
* **Index Gate:** entry appended/updated in `requirements_index`.
* **Schema Gate:** JSON conforms exactly; field names/types exact.

### 8) Failure Modes (return one with reason + remediation)

* `ERROR_DESCRIPTION_REQUIRED()`
* `ERROR_SCRIPT_MISSING(path)`
* `ERROR_PATH_MISSING(path)`
* `ERROR_ID_COLLISION(req_id)`
* `ERROR_SPEC_WRITE(path,reason)`
* `ERROR_INDEX_UPDATE(path,reason)`
* `ERROR_PARSE_SCRIPT_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify before output)

* Real script executed; **no simulated/mocked** operations.
* Feature name 2–4 words; slug `kebab-case`; **PascalCase** for token `FEATURE`.
* Spec contains *Feature Overview, User Stories, Functional Requirements, Success Criteria, Assumptions & Constraints, Open Questions*.
* Success criteria measurable & tech‑agnostic; max 3 open questions.
* Token suggestion present; suggested code location logical for aspect.
* CANARY snapshots emitted when required. 

### 10) Example HUMAN_TEXT (operator‑friendly; optional)

```
=== HUMAN_TEXT BEGIN ===
Created {{.ReqID}}-<ASPECT>-API-106 — User Authentication
Spec: .canary/specs/{{.ReqID}}-<ASPECT>-API-106-user-authentication/spec.md

Next steps:
1) Paste token in src/api/user-authentication.go
2) /canary.plan {{.ReqID}}-<ASPECT>-API-106  (generate implementation plan)
3) /canary.doc create {{.ReqID}}-<ASPECT>-API-106 --type feature  (optional)

Reminder: keep spec WHAT/WHY only; limit to ≤3 [NEEDS CLARIFICATION].
=== HUMAN_TEXT END ===
```

---

### What changed & why (brief)

* **Deterministic outputs:** strict **BEGIN/END** markers + **SUMMARY_JSON** enable parsing and CI enforcement. 
* **Section‑delimited, structured design:** improves clarity, maintainability, and downstream automation. 
* **Parallel pipeline & collision‑safe ID assignment:** explicit DAG + lockfile reduces races; faster end‑to‑end. 
* **No‑mock/no‑simulate:** script must run; missing artifacts reported, not invented. 

### Assumptions & Risks

* The creation script outputs (or can be parsed for) `SPEC_FILE` path; if not, rely on deterministic path construction.
* Aspect heuristics may misclassify; expose `aspect_confidence` if you add it.
* Some repos may require custom ID namespaces; adjust `req_id_pattern` and lockfile path accordingly.

### Targeted questions (for fit)

1. Confirm canonical **aspect vocabulary** and **REQ‑ID pattern**.
2. Should we **always** create `requirements.md` if missing, or fail fast?
3. Do you want an **`aspect_confidence`** field in `SUMMARY_JSON`?
4. Keep CANARY snapshot threshold at **70%** context usage?
