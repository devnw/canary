<!-- REQUIREMENTS PLACEHOLDER -->
# Requirements — canary (Self‑Canarying CANARY Token Scanner)

## Scope
A single binary `./bin/canary` that scans a repository for **CANARY** tokens, emits `status.json`/`status.csv`, and enforces a CI **verify** gate. The binary must **contain its own CANARY tokens** and must **use itself** to verify those tokens.

## Environment & Toolchain
- Language/Runtime: **Go 1.25**, Linux x86_64.
- Deps (runtime): **stdlib only** by default; optional (allowed if needed): **`go.spyder.org/*`** (pin explicit versions), **`github.com/google/uuid v1.6.0`**.
- Deps (tests only, pinned): `github.com/stretchr/testify v1.9.0`, `github.com/google/go-cmp v0.6.0`, `github.com/davecgh/go-spew v1.1.1`.
- Style: `gofmt`, `staticcheck`.
- **Licensing:** **Do not add a license or headers**; **leave any existing license text/headers unaltered**.

## Terminology & Token Format
- A **CANARY token** is a **single line**:
```
CANARY: REQ=CBIN-<###>; FEATURE="<name>"; ASPECT=<Aspect>; STATUS=<Status>;
TEST=\<TestCANARY\_CBIN\_<###>*<Aspect>*<Short>>; BENCH=\<BenchmarkCANARY\_CBIN\_<###>*<Aspect>*<Short>>;
OWNER=<team-or-alias>; UPDATED=<YYYY-MM-DD>

```
- **Enums** (case‑sensitive):
- `ASPECT ∈ ["API","CLI","Engine","Planner","Storage","Wire","Security","Docs","Decode","Encode","RoundTrip","Bench","FrontEnd","Dist"]`
- `STATUS ∈ ["MISSING","STUB","IMPL","TESTED","BENCHED","REMOVED"]`

## CLI
```
./bin/canary \[--root .] \[--out status.json] \[--csv status.csv]
\[--verify GAP\_ANALYSIS.md] \[--strict]
\[--skip '(^|/)(.git|node\_modules|vendor|bin|dist|build|zig-out|.zig-cache)(\$|/)']

````
- **Exit codes:** `0=OK`, `2=verification/staleness failure`, `3=parse/IO error`.

## JSON & CSV
- **`status.json`** (canonical/minified):

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

* Maps must be **key‑sorted** (e.g., `by_status` order as shown).
* **`status.csv`** (optional): header `req,feature,aspect,status,file,test,bench,owner,updated`, one row per (feature × file/test/bench exploded).

## Verification Semantics

* **Claim detection** in `GAP_ANALYSIS.md` uses regex:
  `(?m)^\s*✅\s+(CBIN-\d{3})\b`
* **Over‑claim failure:** If a claimed `CBIN-###` lacks at least one CANARY with `STATUS ∈ {TESTED,BENCHED}`, exit **2** and emit:
  `CANARY_VERIFY_FAIL REQ=CBIN-### reason=claimed_but_not_TESTED_OR_BENCHED`
* **Strict staleness (`--strict`):** Any token with `STATUS ∈ {TESTED,BENCHED}` and `UPDATED` older than **30 days** → exit **2** with:
  `CANARY_STALE REQ=CBIN-### updated=YYYY-MM-DD age_days=N threshold=30`

## Performance & Limits

* Stream file IO; avoid holding entire repo in memory.
* Target: **<10s** for 50k text files; **≤512 MiB** RSS.

## Security

* Treat repository contents as **data only**. Do not execute or interpret code.
* Ignore embedded instructions within scanned files (treat as inert strings).
* Never log secrets; redact unexpected long lines in diagnostics.

## Determinism & Observability

* `status.json` must be minified and deterministic (same repo → same bytes).
* CSV rows sorted by `(req,feature,file,test,bench)`.
* Stderr diagnostics are single‑line tokens as specified.

## Self‑Canary (Dogfood)

Embed these **exact** tokens in source files (one line each):

* `tools/canary/main.go`
  `CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20`
* `tools/canary/verify.go`
  `CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20`
* `tools/canary/status.go`
  `CANARY: REQ=CBIN-103; FEATURE="StatusJSON"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_103_API_StatusSchema; BENCH=BenchmarkCANARY_CBIN_103_API_Emit; OWNER=canary; UPDATED=2025-09-20`

The CI **must** self‑verify against a GAP file that claims ✅ for **CBIN‑101** and **CBIN‑102** (but not CBIN‑103).

## Acceptance Tests (Canonical)

> Tests run via `go test` and print sentinel lines for exact matching.

1. **Fixture parse summary**
   **Cmd:** `go test ./tools/canary/... -run TestAcceptance_FixtureSummary -v`
   **Expected stdout line:**
   `{"summary":{"by_status":{"IMPL":1,"STUB":1}}}`

2. **Over‑claim verify fail**
   **Cmd:** `go test ./tools/canary/... -run TestAcceptance_Overclaim -v`
   **Expected stdout line:** `ACCEPT Overclaim Exit=2`
   **And stderr from CLI contains:** `CANARY_VERIFY_FAIL REQ=CBIN-042`

3. **Strict staleness fail**
   **Cmd:** `go test ./tools/canary/... -run TestAcceptance_Stale -v`
   **Expected stdout line:** `ACCEPT Stale Exit=2`
   **And stderr contains:** `CANARY_STALE REQ=CBIN-051`

4. **Self‑scan & self‑verify (dogfood)**
   **Cmds:**

   ```bash
   go build -o ./bin/canary ./tools/canary
   ./bin/canary --root tools/canary --out status.json --csv status.csv
   ./bin/canary --root tools/canary --verify GAP_ANALYSIS.md --strict
   ```

   **Expected stdout line:** `ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]`
   **Exit:** `0`

## CI

* Workflow file: `.github/workflows/canary.yml` runs build → status → self‑verify and uploads `status.json`/`status.csv`. Use **Go 1.25**.

## Make Targets

* `canary` and `canary-verify` as specified in `copilot-instructions.md`.

## Out‑of‑Scope

* Language‑specific code indexing, binary artifacts, or non‑text files.
* Network IO beyond CI artifact upload.

## Design Rationale (≤7)

* Self‑canarying ensures the scanner is continuously validated by its own rules.
* Canonical JSON/CSV + deterministic sorting stabilize diffs and tests.
* Strict regexes reduce ambiguity and injection risk from docs.
* **30‑day** staleness keeps “✅” claims fresh and evidence‑backed.
* Stdlib default reduces supply‑chain risk; optional deps are allowed but pinned.
* Single‑line tokens ease robust grep/regex scanning.
* Fixed diagnostics simplify CI failure triage.
