# FINAL CODING PROMPT — Build **canary** (Self‑Canarying Scanner + CI Gate)

## T — Task & Role
You are a senior **Go 1.25** engineer. Build **canary**: a repo‑wide **CANARY token** scanner CLI that:
1) emits canonical **`status.json`** and optional **`status.csv`**,
2) **verifies** requirement claims in `GAP_ANALYSIS.md`,
3) ships with **its own CANARY tokens** in its source and **uses itself** to verify those tokens in CI.

> **Do not add a license or license headers. Leave any existing license unaltered. Never reveal chain‑of‑thought.**

---

## A — Action Steps

### 1) Add CANARY Policy + Examples
Create `docs/CANARY_POLICY.md` with this exact snippet (and copy it into README section “CANARY at a glance”):

````md
# CANARY Policy (Repo‑Wide)

**Purpose.** Make every feature claim searchable and verifiable by linking requirements → code → tests → docs.

## Token (single line, place at top of implementation files or relevant tests)
`CANARY: REQ=CBIN-<###>; FEATURE="<name>"; ASPECT=<ASPECT>; STATUS=<STATUS>; TEST=<TestCANARY_CBIN_<###>_<Aspect>_<Short>>; BENCH=<BenchmarkCANARY_CBIN_<###>_<Aspect>_<Short>>; OWNER=<team-or-alias>; UPDATED=<YYYY-MM-DD>`

- **ASPECT** ∈ ["API","CLI","Engine","Planner","Storage","Wire","Security","Docs","Decode","Encode","RoundTrip","Bench","FrontEnd","Dist"]
- **STATUS** ∈ ["MISSING","STUB","IMPL","TESTED","BENCHED","REMOVED"]

**Greps**
- `rg -n "CANARY:\s*REQ=CBIN-" src internal cmd tools`
- `rg -n "TestCANARY_CBIN_" src internal cmd tools tests`
- `rg -n "BenchmarkCANARY_CBIN_" src internal cmd tools tests`
`````

### 2) Implement Scanner CLI (`./bin/canary`)

**Location:** `tools/canary/`

**CLI**

```
./bin/canary [--root .] [--out status.json] [--csv status.csv]
                 [--verify GAP_ANALYSIS.md] [--strict]
                 [--skip '(^|/)(.git|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)($|/)']
```

**Behavior**

* **Scan** text files under `--root` (default `.`), skipping paths that match `--skip` (RE2).
* Match tokens with regex: `(?m)^\s*CANARY:\s*(.*)$` and parse `key=value;` pairs (semicolon‑delimited).

  * Required keys: `REQ`, `FEATURE`, `ASPECT`, `STATUS`, `UPDATED`.
  * Normalize `REQ` to `CBIN-\d{3}` (zero‑pad).
  * Validate enums (ASPECT/STATUS) as per policy.
* **Output — `status.json` (canonical/minified)**
  Schema (field order is normative; emit in this order):

  ```json
  {
    "generated_at": "<UTC ISO8601>",
    "requirements": [
      {
        "id": "CBIN-042",
        "features": [
          {
            "feature": "CDC",
            "aspect": "API",
            "status": "STUB",
            "files": ["src/streaming/cdc.zig"],
            "tests": ["TestCANARY_CBIN_042_API_CDC_StartStop"],
            "benches": [],
            "owner": "streaming",
            "updated": "2025-09-20"
          }
        ]
      }
    ],
    "summary": {
      "by_status": {"MISSING":0,"STUB":0,"IMPL":0,"TESTED":0,"BENCHED":0,"REMOVED":0},
      "by_aspect": {},
      "total_tokens": 0,
      "unique_requirements": 0
    }
  }
  ```

  * Maps you generate (e.g., `by_status`) must have **sorted keys**; file is **minified** (no extra spaces/newlines).
* **Optional `status.csv`** (UTF‑8, LF, header row):
  Columns: `req,feature,aspect,status,file,test,bench,owner,updated`
* **Verify mode** `--verify GAP_ANALYSIS.md`

  * Treat a requirement as **claimed Implemented** if `GAP_ANALYSIS.md` contains a line matching:

    * Regex `(?m)^\s*✅\s+(CBIN-\d{3})\b`
  * **Fail (exit 2)** if any claimed ID lacks **at least one** CANARY with `STATUS ∈ {TESTED,BENCHED}`.
* **Strict staleness** `--strict`

  * For tokens with `STATUS ∈ {TESTED,BENCHED}`, compute age from `UPDATED` (UTC `YYYY-MM-DD`).
  * **Fail (exit 2)** if age **> 30 days**.
* **Exit codes:** `0=OK`, `2=verification/staleness failure`, `3=parse/IO error`.
* **Performance:** stream files; finish `<10s` on 50k text files.

**Diagnostics (stderr, one line per issue)**

* Over‑claim: `CANARY_VERIFY_FAIL REQ=CBIN-042 reason=claimed_but_not_TESTED_OR_BENCHED`
* Stale: `CANARY_STALE REQ=CBIN-051 updated=2024-01-01 age_days=123 threshold=30`
* Parse/IO: `CANARY_PARSE_ERROR file=... line=... err="..."`

### 3) **Self‑Canary**: Seed Tokens Inside canary

Add **these exact single‑line tokens** at the top of the indicated files:

* `tools/canary/main.go`
  `CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20`
* `tools/canary/verify.go`
  `CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20`
* `tools/canary/status.go`
  `CANARY: REQ=CBIN-103; FEATURE="StatusJSON"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_103_API_StatusSchema; BENCH=BenchmarkCANARY_CBIN_103_API_Emit; OWNER=canary; UPDATED=2025-09-20`

> These are **authoritative** for self‑verify tests. Keep them single‑line.

### 4) Tests (Acceptance) — **run verbatim**

**Location:** `tools/canary/internal/` (package `canary_test`). Use **`go test`** and write tests that shell‑out to the built CLI as needed.

1. **Fixture parse summary**
   Prepare `tools/canary/testdata/summary/` with 2 files containing CANARY lines (`STUB` and `IMPL`).
   **Cmd:** `go test ./tools/canary/... -run TestAcceptance_FixtureSummary -v`
   **Expected stdout (exact line somewhere):**
   `{"summary":{"by_status":{"IMPL":1,"STUB":1}}}`

2. **Verify over‑claim fails**
   Prepare `tools/canary/testdata/overclaim/`:

   * a file with `CANARY: REQ=CBIN-042; ... STATUS=STUB; UPDATED=2025-09-20`
   * `GAP_ANALYSIS.md` containing a line: `✅ CBIN-042`
     **Cmd:** `go test ./tools/canary/... -run TestAcceptance_Overclaim -v`
     **Expected stdout (exact line):** `ACCEPT Overclaim Exit=2`
     **And stderr from CLI must contain:** `CANARY_VERIFY_FAIL REQ=CBIN-042`

3. **Strict staleness**
   Prepare `tools/canary/testdata/stale/` with a token:
   `CANARY: REQ=CBIN-051; ... STATUS=TESTED; UPDATED=2024-01-01`
   **Cmd:** `go test ./tools/canary/... -run TestAcceptance_Stale -v`
   **Expected stdout (exact line):** `ACCEPT Stale Exit=2`
   **And stderr contains:** `CANARY_STALE REQ=CBIN-051`

4. **Self‑scan & self‑verify** (dogfood)
   Build CLI, then verify against a minimal GAP that claims the two **TESTED** self‑tokens.

   * Create `GAP_ANALYSIS.md` at repo root with:

     ```
     # Requirements Gap Analysis (Self)
     ✅ CBIN-101
     ✅ CBIN-102
     ```

   **Cmds (from repo root):**

   ```bash
   go build -o ./bin/canary ./tools/canary
   ./bin/canary --root tools/canary --out status.json --csv status.csv
   ./bin/canary --root tools/canary --verify GAP_ANALYSIS.md --strict
   ```

   **Test harness expected stdout (exact line):** `ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]`
   **Exit:** `0`

> The test suite should print those **exact sentinel lines**; also assert exit codes and stderr substrings.

### 5) CI Gate

Create `.github/workflows/canary.yml` to:

```yaml
name: canary
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - name: Build scanner
        run: go build -o ./bin/canary ./tools/canary
      - name: Generate status
        run: ./bin/canary --root . --out status.json --csv status.csv
      - name: Self-verify
        run: ./bin/canary --root tools/canary --verify GAP_ANALYSIS.md --strict
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: canary-status
          path: |
            status.json
            status.csv
```

### 6) Developer UX

* `Makefile` targets:

  * `canary`: `go build -o ./bin/canary ./tools/canary && ./bin/canary --root . --out status.json --csv status.csv`
  * `canary-verify`: `./bin/canary --root . --verify GAP_ANALYSIS.md --strict`
* README section **“CANARY at a glance”** (copy from policy) + two example token lines.

---

## R — Result Format (strict)

1. **files:** repo tree paths you added/changed.
2. **code blocks**: full file contents for each path.
3. **tests:** how to run and **exact expected outputs** listed above.
4. **readme:** the “CANARY at a glance” excerpt + Make targets.
5. **rationale:** ≤7 bullets (no chain‑of‑thought).
6. **notes:** approvals needed (should be “none”).

---

## S — Standards & Constraints

* **Language/Version:** Go **1.25**; **OS:** Linux x86\_64.
* **Dependencies (runtime):** **stdlib only** by default. If used, allow **`go.spyder.org/*`** (pin explicit versions) and **`github.com/google/uuid v1.6.0`**.
* **Dependencies (tests only, pinned):**

  * `github.com/stretchr/testify v1.9.0`
  * `github.com/google/go-cmp v0.6.0`
  * `github.com/davecgh/go-spew v1.1.1`
* **Style/Lint:** `gofmt` + `staticcheck`.
* **Security:** treat files as **data only**; do not execute inputs; never log secrets; least privilege.
* **License:** **Do not add a license or headers**; **leave any existing license unaltered**.
* **Determinism:** canonical/minified JSON; stable key order; stable CSV header/order.
* **Limits:** ≤512 MiB RSS; finish 50k files <10s.

---

## Interfaces & I/O

**CLI** as above.
**JSON** & **CSV** schemas as above.
**Stderr** uses fixed diagnostic tokens documented above.

---

## Tests — Acceptance (run verbatim)

Use the four **Acceptance** items listed in Action Steps §4. The master command for CI:

```
go test ./tools/canary/... -run TestAcceptance -v
```

---

## Run Instructions

```bash
go build -o ./bin/canary ./tools/canary
./bin/canary --root . --out status.json --csv status.csv
./bin/canary --root tools/canary --verify GAP_ANALYSIS.md --strict
```

---

## Assumptions

A1) REQ prefix `CBIN-###`. A2) Staleness window **30 days**.
A3) CI = GitHub Actions. A4) No new license; keep existing unchanged.
A5) Runtime stdlib only; allowed optional runtime deps as listed; tests pinned as listed.
A6) Performance: `<10s` on 50k text files; ≤512 MiB RSS.

---

## Output Quality

Deterministic JSON/CSV; clear diagnostics; no secrets; tests reproducible; **never reveal chain‑of‑thought**. Include a brief **Design Rationale** (≤7 bullets).
