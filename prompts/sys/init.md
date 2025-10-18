# FINAL CODING PROMPT — Seed {{PROJECT_NAME}} with CANARY Tokens, Scanner, CI Gate

## T — Task & Role
You are a senior **{{SCANNER_LANG}} {{SCANNER_LANG_VERSION}}** engineer. Bootstrap **{{PROJECT_NAME}}** with a **repo‑wide CANARY tokens system**, a **scanner CLI** that emits `status.json`/`status.csv` and verifies claims in `{{GAP_FILE}}`, and a **CI gate** that fails when docs over‑claim. Do **not** add or modify license text/headers.

## A — Action Steps

1) **Add CANARY Policy (doc snippet + examples)**
   - Create `{{DOCS_DIR}}/CANARY_POLICY.md` containing:
     - One‑line **CANARY token** (language‑agnostic) to place above implementing functions or at file headers:
       ```
       CANARY: REQ={{REQ_PREFIX}}-<###>; FEATURE="<name>"; ASPECT={{ASPECT}}; STATUS={{STATUS}};
               TEST=<TestCANARY_{{REQ_PREFIX}}_<###>_<Short>>; BENCH=<BenchmarkCANARY_{{REQ_PREFIX}}_<###>_<Short>>;
               OWNER=<team-or-alias>; UPDATED=<YYYY-MM-DD>
       ```
     - **Enums** (case‑sensitive):
       - `ASPECT ∈ {{ASPECT_ENUM}}`
       - `STATUS ∈ {{STATUS_ENUM}}`
     - **Test/bench naming**:
       - `TestCANARY_{{REQ_PREFIX}}_<###>_<Aspect>_<Short>`
       - `BenchmarkCANARY_{{REQ_PREFIX}}_<###>_<Aspect>_<Short>`
     - **Greps**:
       - `rg -n "CANARY:\s*REQ={{REQ_PREFIX}}-" {{SOURCE_DIRS}}`
       - `rg -n "TestCANARY_{{REQ_PREFIX}}_" {{TEST_DIRS}}`

2) **Implement Scanner CLI (`{{SCANNER_BIN}}`)**
   - Location: `{{SCANNER_DIR}}/`
   - CLI:
     ```
     {{SCANNER_BIN}} [--root {{REPO_ROOT}}] [--out status.json] [--csv status.csv]
                     [--verify {{GAP_FILE}}] [--strict] [--skip "{{SKIP_DIRS_REGEX}}"]
     ```
   - Behavior:
     - Parse `CANARY:` lines (regex: `\bCANARY:\s*(.*)$`), read `key=value;` pairs.
     - Validate enums; normalize `REQ` as `{{REQ_PREFIX}}-NNN`.
     - Emit **`status.json`** schema:
       ```json
       {
         "generated_at": "<UTC ISO8601>",
         "requirements": [
           { "id": "{{REQ_PREFIX}}-042", "features": [
               { "feature": "CDC", "aspect": "API", "status": "STUB",
                 "files": ["src/streaming/cdc.zig"], "tests": ["TestCANARY_{{REQ_PREFIX}}_042_CDC_StartStop"],
                 "benches": [], "owner": "streaming", "updated": "2025-09-20"
               }
           ]}
         ],
         "summary": { "by_status": {}, "by_aspect": {} }
       }
       ```
     - Optional **CSV** explosion of rows (`req,feature,aspect,status,file,test,bench,owner,updated`).
     - `--verify {{GAP_FILE}}`: fail (**exit 2**) if any `{{REQ_PREFIX}}-NNN` is **claimed Implemented/✅** in `{{GAP_FILE}}`
       **without** at least one CANARY entry with `STATUS ∈ {TESTED,BENCHED}`.
     - `--strict`: also fail on **stale** `UPDATED` for `STATUS ∈ {TESTED,BENCHED}` older than **{{STALE_DAYS}} days**.
     - Exit codes: **0=OK**, **2=verification/staleness failure**, **3=parse/IO error**.
     - **Performance**: stream files; complete <10s on 50k files (text‑only scan).

3) **Tests for Scanner**
   - Location: `{{SCANNER_DIR}}/{{TEST_SUBDIR}}`
   - Use **{{TEST_FRAMEWORK}}**.
   - **Acceptance tests (run verbatim)**
     1. **Fixture parse summary**
        - Given 2 files with `STATUS=STUB` and `STATUS=IMPL`, expect summary counts:
          ```json
          {"summary":{"by_status":{"STUB":1,"IMPL":1}}}
          ```
     2. **Verify over‑claim fails**
        - Given `{{GAP_FILE}}` claiming “Implemented/✅” for `{{REQ_PREFIX}}-042` but repo has only `STATUS=STUB`,
          `{{SCANNER_BIN}} --verify {{GAP_FILE}}` → **exit 2**, stderr contains `CANARY_VERIFY_FAIL REQ={{REQ_PREFIX}}-042`.
     3. **Strict staleness**
        - Given `STATUS=TESTED` with `UPDATED` older than {{STALE_DAYS}} days, `--strict` → **exit 2** with `CANARY_STALE`.

4) **CI Gate**
   - Add `{{CI_DIR}}/{{CI_FILE}}` to:
     - Build `{{SCANNER_BIN}}` with **{{SCANNER_LANG}} {{SCANNER_LANG_VERSION}}**.
     - Run:
       ```bash
       {{SCANNER_BIN}} --root . --out status.json --csv status.csv
       {{SCANNER_BIN}} --root . --verify {{GAP_FILE}} --strict
       ```
     - Upload artifacts: `status.json`, `status.csv`.

5) **Developer UX**
   - `Makefile` (or `justfile`) targets: `canary`, `canary-verify`.
   - `README.md` section “CANARY at a glance” (copy from policy).
   - Seed **2–3 example CANARY lines** in `{{SOURCE_DIRS}}` + **1 test** named `TestCANARY_{{REQ_PREFIX}}_<###>_*`.

## R — Result Format (strict)
1. **files:** repo tree of added/changed files (paths).
2. **code blocks** per file with complete contents.
3. **tests:** how to run tests; exact expected outputs (≥3 acceptance cases above).
4. **readme:** snippet with CANARY commands.
5. **rationale:** ≤7 bullets.
6. **notes:** approvals needed (should be “none”).

## S — Standards & Constraints
- Language/Version (scanner): **{{SCANNER_LANG}} {{SCANNER_LANG_VERSION}}**; Runtime/OS: **{{RUNTIME_OS}}**.
- Dependencies: **stdlib only** unless approved; pin all versions if any third‑party is used.
- Style/Lint: **{{STYLE_OR_LINT}}**.
- Security: placeholders (e.g., `{{API_KEY}}`); never log secrets; least privilege.
- License: **{{LICENSE}}**; do not add/modify license text/headers.
- Repro: deterministic outputs; stable sort; fixed seeds if any.

## Interfaces & I/O
- Inputs: repo path; optional `{{GAP_FILE}}` for verification.
- Outputs: `status.json`, optional `status.csv`; structured stderr diagnostics.

## Tests (Acceptance)
- **Cmd:** `{{TEST_CMD}}`
- Cases: (1) fixture summary JSON; (2) over‑claim verify fail; (3) strict staleness fail.

## Run Instructions
```bash
{{BUILD_SCANNER_CMD}}
{{SCANNER_BIN}} --root . --out status.json --csv status.csv
{{SCANNER_BIN}} --root . --verify {{GAP_FILE}} --strict
````

## Assumptions

* A1) Requirement IDs use **{{REQ\_PREFIX}}-###**.
* A2) Status/Aspect enums = **{{STATUS\_ENUM}}** / **{{ASPECT\_ENUM}}**.
* A3) CI provider = **{{CI\_PROVIDER}}**.

## Output Quality

* Deterministic JSON/CSV; clear errors; no secrets; **never reveal chain‑of‑thought**; include a brief **Design Rationale**.
