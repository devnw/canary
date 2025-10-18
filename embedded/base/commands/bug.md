## BugCmd (Create Bug Report)

```yaml
---
description: Create a new BUG-<ASPECT>-XXX CANARY + actionable bug report from natural-language input; enforce best-practice bug reporting; no-mock/no-simulate
command: BugCmd
version: 1.0
subcommands: [bug]
scripts:
  sh: .canary/scripts/create-bug.sh
outputs:
  - bug_report_markdown: .canary/bugs/<BUG-ID>-<slug>/report.md
  - summary_json: STDOUT (unwrapped JSON; strict schema below)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
defaults:
  bugs_root: .canary/bugs
  template_report: .canary/templates/bug-report-template.md
  id_lockfile: .canary/.bug.id.lock
  bug_id_regex: '^BUG-[A-Za-z]+-[0-9]{3}$'
  aspect_vocab: ["API","CLI","Engine","Storage","Security","Docs","Frontend","Data","Infra"]
  severity_vocab: ["S1-Critical","S2-High","S3-Medium","S4-Low"]
  priority_vocab: ["P0","P1","P2","P3"]
  status_vocab: ["OPEN","IN_PROGRESS","FIXED","VERIFIED","WONTFIX","DUPLICATE","NOTABUG","CANNOT_REPRO","BLOCKED"]
  repro_target_samples: 5
---
```

### 1) Inputs

* **User Arguments (raw):** `$ARGUMENTS` — natural‑language defect report from user/agent.
  Optional flags: `--title "..."`, `--aspect <ASPECT>`, `--severity S1..S4`, `--priority P0..P3`, `--labels <csv>`, `--related <BUG-ID|REQ-ID>...`.
  **MANDATORY:** If `$ARGUMENTS` is empty → `ERROR_DESCRIPTION_REQUIRED()`.

### 2) Preconditions & Resolution

1. **Paths/Gate:** Ensure `bugs_root` & `template_report` exist → else `ERROR_PATH_MISSING(path)`.
2. **Script/Gate:** Ensure `.canary/scripts/create-bug.sh` is executable → else `ERROR_SCRIPT_MISSING(path)`.
3. **Aspect detection:** Classify to `aspect_vocab` using cues in input; if uncertain default **API** and record `aspect_confidence:"low"`.
4. **Duplicate check (must run):**

   ```bash
   canary grep "<key terms from title/description>"
   ```

   Identify candidate duplicates (same summary keywords; same files/stack; same component). If a high‑similarity existing BUG is found, return as `DUPLICATE_CANDIDATE` with links; **do not create a new ID** unless user insists. Bugzilla guidance: check for existing reports; one issue per bug. ([Bugzilla][1])
5. **ID generation (collision‑safe):** Acquire `id_lockfile`; scan `bugs_root` for `BUG-<ASPECT>-NNN`; pick next zero‑padded NNN; if dir exists, increment & retry; release lock.

### 3) Planning & Parallelism

Use a Work‑DAG with concurrency groups; **join** before shared writes (no speculative steps).

* **CG‑1 Parse:** title, 5W’s (who/what/when/where/why), severity/priority hints, environment, logs. MoT recommends clear, audience‑appropriate storytelling with key elements. ([Ministry of Testing][2])
* **CG‑2 Deduplicate:** run `canary grep` and compute similarity.
* **CG‑3 Create:** run script to scaffold folder + `report.md`.
* **CG‑4 Author:** fill **report.md** sections (see §4) strictly from facts; **no speculation**; separate *Expected*/*Actual*; provide minimal reproducible steps and environment. Bugzilla emphasizes clear summary, precise steps, expected vs actual, and additional info. ([Bugzilla][1])
* **CG‑5 Register:** write CANARY token; update index if used (e.g., `.canary/bugs/index.md`).

### 4) Behavior (must do; never simulate)

* **Run the real script**; do **not** fabricate paths, line numbers, or IDs.
* **One issue per bug**; split unrelated issues into separate BUG IDs. ([Bugzilla][1])
* **Required sections in `report.md` (exact headings):**

  1. **Bug Summary** (≤ ~10 words, problem‑focused; not a solution). ([Bugzilla][1])
  2. **Environment** (app version/commit, OS & version, device, browser/version, locale, data/profile flags).
  3. **Preconditions** (state/setup needed).
  4. **Steps to Reproduce** (numbered, minimal but complete; include key interactions). ([Bugzilla][1])
  5. **Expected Result** / **Actual Result** (exact message, stack, screenshots/logs refs; facts ≠ speculation). ([Bugzilla][1])
  6. **Reproducibility** (e.g., `3/5` attempts; note “intermittent” & conditions). ([Bugzilla][1])
  7. **Scope & Impact** (user/business impact; data loss, security, perf).
  8. **Regression/First‑Seen** (version/build range; suspected change / pushlog if known). ([Bugzilla][1])
  9. **Workarounds** (if any).
  10. **Attachments** (sanitized logs, screenshots, reduced testcase URL/path). Bugzilla suggests reduced testcases for web issues. ([Bugzilla][1])
  11. **Related/Dependencies** (BUG/REQ links).
  12. **Next Actions (Triage)** (owner/team, first fix ideas, test to add).
* **Tone & brevity:** Be factual, avoid blame; right‑sized detail for dev audience; include visuals where useful. ([Ministry of Testing][2])

**CANARY token (paste into nearest code touchpoint or failing test):**

```go
// CANARY: BUG=BUG-<ASPECT>-NNN; TITLE="<ShortProblemSummary>";
//         ASPECT=<ASPECT>; STATUS=OPEN;
//         SEVERITY=<S1|S2|S3|S4>; PRIORITY=<P0|P1|P2|P3>;
//         REPRO=<k>/<n>; UPDATED=<YYYY-MM-DD>
```

### 5) CANARY Snapshot Protocol (compact; low‑token)

Emit after **ID selection** and after **report write**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"bug|id|write",
  "f":[[".canary/bugs/<BUG-ID>-<slug>/report.md",1,999]],
  "k":["bug:<BUG-ID>","title:<slug>","aspect:<ASPECT>","sev:<S#>","prio:<P#>","repro:<k>/<n>"],
  "fp":["<disproven assumption>"],   # prior false positives to avoid retrying failures
  "iss":["<linked-REQ-IDs-or-n/a>"],
  "nx":["triage","assign","create failing test"]
}'
```

### 6) Output Contract (strict)

Return artifacts **in order** (JSON **not** in code fences).
**A. BUG_REPORT_MARKDOWN** — delimiters:
`=== BUG_REPORT_MARKDOWN BEGIN ===` … `=== BUG_REPORT_MARKDOWN END ===`

**B. SUMMARY_JSON (unwrapped) — schema**

```json
{
  "ok": true,
  "bug_id": "BUG-API-123",
  "title": "Login fails on first attempt",
  "aspect": "API",
  "status": "OPEN",
  "severity": "S2-High",
  "priority": "P1",
  "repro": {"success": 3, "attempts": 5, "rate": 0.6},
  "env": {
    "app_version": "v1.9.0+abcd123",
    "os": "macOS 14.5",
    "browser": "Chrome 128",
    "device": "MBP 14 M2",
    "locale": "en-US"
  },
  "paths": {
    "dir": ".canary/bugs/BUG-API-123-login-fails",
    "report": ".canary/bugs/BUG-API-123-login-fails/report.md"
  },
  "duplicates": [{"bug_id":"BUG-API-097","score":0.84,"reason":"same stack signature"}],
  "token_suggestion": "// CANARY: BUG=BUG-API-123; TITLE=\"Login fails on first attempt\"; ASPECT=API; STATUS=OPEN; SEVERITY=S2; PRIORITY=P1; REPRO=3/5; UPDATED=2025-10-18",
  "gates": {
    "one_issue_only": "pass|fail",
    "summary_short_problem_focused": "pass|fail",
    "steps_present": "pass|fail",
    "expected_actual_present": "pass|fail",
    "env_present": "pass|fail",
    "repro_present": "pass|fail",
    "privacy_sanitized": "pass|fail",
    "duplicate_checked": "pass|fail"
  },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" }
}
```

### 7) Validation Gates (compute & report)

* **One‑Issue Gate:** report addresses exactly one problem. ([Bugzilla][1])
* **Summary Gate:** ~10‑word, problem‑focused title. ([Bugzilla][1])
* **Steps Gate:** numbered, precise, minimal repro steps. ([Bugzilla][1])
* **Expected/Actual Gate:** both sections present; facts separated from speculation. ([Bugzilla][1])
* **Env Gate:** OS/browser/app version captured. ([Bugzilla][1])
* **Repro Gate:** `k/n` provided; note intermittent if needed. ([Bugzilla][1])
* **Attachments Gate:** logs/screenshots or reduced testcase linked when useful. ([Bugzilla][1])
* **Duplicate Gate:** similarity search run & recorded.
* **Privacy Gate:** secrets scrubbed from logs/screens.
* **Schema Gate:** `SUMMARY_JSON` conforms exactly.

### 8) Failure Modes (return one + remediation)

* `ERROR_DESCRIPTION_REQUIRED()`
* `ERROR_PATH_MISSING(path)`
* `ERROR_SCRIPT_MISSING(path)`
* `ERROR_ID_COLLISION(bug_id)`
* `ERROR_DUPLICATE_CANDIDATE(existing_bug_id,reason)` (suggest linking instead of creating)
* `ERROR_REPORT_WRITE(path,reason)`
* `ERROR_PARSE_OUTPUT(reason)`

### 9) Quality Checklist (auto‑verify)

* Real script executed (**no mocked creation**).
* Title concise; steps reproducible; *Expected vs Actual* included; environment complete. ([Bugzilla][1])
* Visuals/logs attached when helpful; tone factual and audience‑appropriate. ([Ministry of Testing][2])
* Reduced testcase or regression window provided when applicable. ([Bugzilla][1])
* CANARY snapshots emitted as specified.

### 10) Example HUMAN_TEXT (operator‑friendly; optional)

```
=== HUMAN_TEXT BEGIN ===
Created BUG-API-123 — Login fails on first attempt
Report: .canary/bugs/BUG-API-123-login-fails/report.md
Next: assign owner, add failing test, link fix MR.
=== HUMAN_TEXT END ===
```

### 11) Example `report.md` skeleton (WHAT happened; not HOW to fix)

```
=== BUG_REPORT_MARKDOWN BEGIN ===
# BUG-API-123 — Login fails on first attempt

## Bug Summary
Login fails on first attempt

## Environment
App v1.9.0+abcd123 • macOS 14.5 • Chrome 128 • en-US

## Preconditions
Fresh account; logged out; cookies enabled.

## Steps to Reproduce
1) Open /login
2) Enter valid email/password
3) Click “Sign in” once

## Expected Result
User is authenticated and redirected to /home.

## Actual Result
HTTP 401 with message “invalid session token”. See logs: logs/api-2025-10-18T10:02Z.txt

## Reproducibility
3/5 attempts; fails on first attempt only.

## Scope & Impact
Blocks onboarding; revenue impact for first‑time users.

## Regression / First Seen
First seen in v1.9.0; v1.8.3 unaffected.

## Workarounds
Retry succeeds on second attempt.

## Attachments
- Screenshot: artifacts/login-first-attempt-401.png
- Logs: artifacts/api-401-excerpt.txt
- Reduced testcase: artifacts/login-retry.html

## Related / Dependencies
REQ: CBIN-API-134 (UserOnboarding)

## Next Actions (Triage)
Assign API team • Add failing test • Investigate session initialization race.
=== BUG_REPORT_MARKDOWN END ===
```
