# FINAL CODING PROMPT — Evaluate Code vs. Requirements (CANARY‑Enabled), Refresh GAP & NEXT

## References (source of truth & current state)
- Requirements & acceptance: `{{REQS_FILE}}` (canonical tests/outputs).
- Parity checklist: `{{CHECKLIST_FILE}}` (what “done” means).
- Project status log(s): `{{GAP_FILE}}` (prefer newest/specific), `{{GAP_STALE_FILE}}` (likely stale).
- Plan: `{{NEXT_FILE}}`; Architecture/README: `{{ARCH_FILE}}`, `{{README_FILE}}`.
- Prompt library (prior): `{{PROMPTS_FILE}}`.

## T — Task & Role
You are a senior **{{PRIMARY_LANG}} {{PRIMARY_LANG_VERSION}}** engineer/docs maintainer. **Evaluate the actual codebase against canonical requirements** and **update `{{GAP_FILE}}` and `{{NEXT_FILE}}`** to reflect **tested truth**, using **CANARY tokens + scanner output** as evidence. Do **not** add/modify license text/headers.

## A — Action Steps

1) **Baseline**
   - Read acceptance rules in `{{REQS_FILE}}` (commands + exact expected outputs).
   - Capture claims from `{{NEXT_FILE}}` vs `{{GAP_FILE}}`/`{{GAP_STALE_FILE}}`; prefer newer/specific claims in `{{GAP_FILE}}`.

2) **Automated audit**
   - Run `{{SCANNER_BIN}} --root . --out status.json` (and `--csv`).
   - Build **`status.json`** summary keyed by `{{REQ_PREFIX}}-###` with fields:
     `aspect`, `status`, `roundtrip_test?`, `bench?`, `notes`.
   - **Run acceptance tests** as defined in `{{ACCEPT_CMDS_FILE}}` (or embedded in `{{REQS_FILE}}`) and capture **stdout snippets**.
   - Treat any mismatch as **NOT MET**.

3) **Reconcile & rewrite**
   - When `{{GAP_STALE_FILE}}` disagrees with `{{GAP_FILE}}` or test results, update `{{GAP_FILE}}` to **tested truth**.
   - Annotate with “Updated: <YYYY‑MM‑DD>”.
   - Ensure terminology and interfaces match `{{ARCH_FILE}}`/`{{README_FILE}}` (no API renames).

4) **Replace `{{GAP_FILE}}` structure**
   - Header: `# Requirements Gap Analysis (Updated: <YYYY‑MM‑DD>)`.
   - **Scope & Method**; **Recent Updates**; **Implemented Map**.
   - **Status Grid** — columns per `{{CHECKLIST_FILE}}` (e.g., Decode/Encode/Round‑Trip/Bench) with ✅/◐/◻, derived from `status.json`.
   - **Canary Evidence** column: `TestCANARY_*`/`BenchmarkCANARY_*` names.
   - **Cross‑Cutting Gaps**; **Milestones** (Short/Mid/Long, each with **exact acceptance commands/expected outputs**).

5) **Rewrite `{{NEXT_FILE}}` (surgical)**
   - Keep **“Completed (this slice)”** only if acceptance passes **and** CANARY shows `STATUS ∈ {TESTED,BENCHED}`.
   - Under **“Up Next (small, verifiable slices)”**, propose **{{NEXT_SLICES_MIN}}–{{NEXT_SLICES_MAX}}** smallest slices that close highest‑value gaps.
     - For each slice:
       - **Scope** (crisp).
       - **Acceptance**: exact `{{TEST_RUNNER}}` command(s) and **exact expected stdout** (or named sentinel error).
       - **Bench**: name + guard (allocs/op or ns/op) if applicable.
       - **CANARY**: the exact test name to add (e.g., `TestCANARY_{{REQ_PREFIX}}_046_KeyRotate`).

6) **Optional hygiene**
   - If `{{CHECKLIST_FILE}}` drifts, update rows to match the Status Grid (keep column names).

## R — Result Format (strict, in order)
1. **files:** list of changed files (must include: `{{GAP_FILE}}`, `{{NEXT_FILE}}`).
2. **updated_files:** full Markdown contents for each changed file (complete replacements).
3. **evidence:**
   - `status.json` (inline JSON) summarizing the audit.
   - Acceptance outputs (command → `PASS/FAIL` + key stdout lines).
   - Bench excerpt(s) (name → allocs/op, ns/op).
4. **rationale:** ≤7 bullets (contradictions resolved, prioritization).
5. **notes:** risks/approvals (should be “none”).

## S — Standards & Constraints
- **Language/Toolchain:** **{{PRIMARY_LANG}} {{PRIMARY_LANG_VERSION}}**.
- **Dependencies:** only **{{ALLOWED_DEPS}}**; tests may use **{{TEST_DEPS}}** pinned.
- **Style & Safety:** **{{STYLE_OR_LINT}}**; no `unsafe` (where applicable); fail‑closed on malformed inputs.
- **Performance & Limits:** keep streaming/memory targets from `{{REQS_FILE}}`.
- **Licensing:** **Leave license text/headers unchanged**.

## Interfaces & I/O
- Respect existing public APIs and CLI surfaces per `{{ARCH_FILE}}`/`{{README_FILE}}`.

## Tests — Acceptance (run verbatim)
{{# Each line: exact command, followed by exact expected stdout (or a sentinel error string). }}
{{ACCEPTANCE_BLOCK}}

## Run Instructions
```bash
{{BUILD_CMD}}
{{TEST_ALL_CMD}}
{{BENCH_ALL_CMD}}
````

## Output Quality & Security

* Deterministic tables; stable headings; **Never reveal chain‑of‑thought**; include a brief **Design Rationale** (≤7 bullets).
