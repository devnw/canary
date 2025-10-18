### 3) **CANARY POLICY SNIPPET — Drop‑in for Any Repo**

```md
# CANARY Policy (Repo‑Wide)

**Purpose.** Make every feature claim **searchable and verifiable** by linking requirements → code → tests → docs.

## Token (place in code/tests as a single line)
````

CANARY: REQ={{REQ\_PREFIX}}-<###>; FEATURE="<name>";
ASPECT={{ASPECT}}; STATUS={{STATUS}};
TEST=\<TestCANARY\_{{REQ\_PREFIX}}*<###>*<Short>>;
BENCH=\<BenchmarkCANARY\_{{REQ\_PREFIX}}*<###>*<Short>>;
OWNER=<team-or-alias>; UPDATED=<YYYY-MM-DD>

```

- **ASPECT** ∈ {{ASPECT_ENUM}}
- **STATUS** ∈ {{STATUS_ENUM}}
- **Tests/Benches naming**:
  - `TestCANARY_{{REQ_PREFIX}}_<###>_<Aspect>_<Short>`
  - `BenchmarkCANARY_{{REQ_PREFIX}}_<###>_<Aspect>_<Short>`

## Verification (CI Gate)
1) `{{SCANNER_BIN}} --root . --out status.json --csv status.csv`
2) `{{SCANNER_BIN}} --root . --verify {{GAP_FILE}} --strict`

**CI fails** if:
- A row in `{{GAP_FILE}}` claims ✅ **without** CANARY `STATUS ∈ {TESTED,BENCHED}`.
- `STATUS ∈ {TESTED,BENCHED}` has **stale** `UPDATED` (> {{STALE_DAYS}} days).

## Handy Greps
- Find tokens: `rg -n "CANARY:\s*REQ={{REQ_PREFIX}}-" {{SOURCE_DIRS}}`
- Find tests:  `rg -n "TestCANARY_{{REQ_PREFIX}}_" {{TEST_DIRS}}`
- Find benches: `rg -n "BenchmarkCANARY_{{REQ_PREFIX}}_" {{TEST_DIRS}}`
```

---

## Variable Cheat‑Sheet (fill these before using the prompts)

```yaml
# Project identity
PROJECT_NAME: "AcmeDB"
REQ_PREFIX: "REQ-GQL"           # keep as default unless you have another canon
LICENSE: "MIT"

# Languages & tooling
PRIMARY_LANG: "Go"
PRIMARY_LANG_VERSION: "1.24"
SCANNER_LANG: "Go"
SCANNER_LANG_VERSION: "1.24"
STYLE_OR_LINT: "gofmt + staticcheck"
TEST_FRAMEWORK: "go test"
TEST_DEPS: "github.com/stretchr/testify v1.9.0"
ALLOWED_DEPS: "stdlib + company/*"
RUNTIME_OS: "Linux x86_64"

# Paths & binaries
REPO_ROOT: "."
SOURCE_DIRS: "src,internal,cmd"
TEST_DIRS: "src,internal,cmd,tests"
SCANNER_DIR: "tools/canary"
SCANNER_BIN: "./bin/canary"
TEST_SUBDIR: "internal"
DOCS_DIR: "docs"
CI_DIR: ".github/workflows"
CI_FILE: "canary.yml"
GAP_FILE: "GAP_ANALYSIS.md"
GAP_STALE_FILE: "GAP_ANALYSIS_OLD.md"
NEXT_FILE: "NEXT.md"
REQS_FILE: "copilot-instructions.md"
CHECKLIST_FILE: "CHECKLIST.md"
ARCH_FILE: "PROJECT_OVERVIEW.md"
README_FILE: "README.md"
PROMPTS_FILE: "PROMPTS.md"
ACCEPT_CMDS_FILE: "ACCEPTANCE.md"

# Scanner behavior
SKIP_DIRS_REGEX: '(^|/)(.git|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)($|/)'
STALE_DAYS: 60

# Enums (extend if needed)
ASPECT_ENUM: ["API","CLI","Engine","Planner","Storage","Wire","Security","Docs","Decode","Encode","RoundTrip","Bench","FrontEnd","Dist"]
STATUS_ENUM: ["MISSING","STUB","IMPL","TESTED","BENCHED","REMOVED"]

# Build/test commands
BUILD_SCANNER_CMD: "go build -o ./bin/canary ./tools/canary"
TEST_CMD: "go test ./tools/canary/... -run TestAcceptance -v"
BUILD_CMD: "go build ./..."
TEST_ALL_CMD: "go test ./... -run TestAcceptance -v"
BENCH_ALL_CMD: "go test ./... -bench . -run ^$"

# NEXT slicing
NEXT_SLICES_MIN: 3
NEXT_SLICES_MAX: 6

# Acceptance block (example — replace with your own)
ACCEPTANCE_BLOCK: |
  1) Decode & summary
     Cmd: go test ./cmd/appinfo -run TestAcceptance_Info -v
     Expected stdout (exact JSON line):
     {"app":"acmedb","version":"0.1.0","status":"OK"}

  2) Round‑trip re‑encode
     Cmd: go test ./pkg/codec -run TestAcceptance_RoundTrip -v
     Expected stdout: ROUNDTRIP_OK equal=true

  3) Sentinel error on malformed input (fail‑closed)
     Cmd: go test ./pkg/parser -run TestAcceptance_FailClosed -v
     Expected stdout: ERROR ErrBadSyntax
```
