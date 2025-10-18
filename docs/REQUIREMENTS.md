# CANARY Requirements Specification

**Project:** CANARY Token Tracking System
**Version:** 2.0
**Last Updated:** 2025-10-18

## Executive Summary

CANARY is a requirement tracking system that embeds structured tokens directly into source code. It has evolved from a simple scanner (CBIN-101, 102, 103) to a comprehensive CLI tool with 90+ requirements covering specification management, dependency tracking, multi-project support, documentation tracking, and AI agent integration.

## Scope

### In Scope

**Core Tracking (CBIN-101, 102, 103)**
- File scanning for CANARY tokens
- JSON/CSV output generation
- Verification gates for claimed requirements
- Staleness detection (30-day threshold)
- Self-canary validation

**CLI System (CBIN-104+)**
- Full-featured command-line interface
- SQLite database for fast queries
- Specification and plan generation
- Workflow automation
- Multi-project support
- Dependency management
- Documentation tracking

**AI Integration**
- Slash commands for autonomous workflows
- Constitutional principle enforcement
- Test-first guidance
- Automated priority selection

### Out of Scope

- Language-specific code indexing (AST parsing)
- Binary artifact analysis
- Network-based tracking services
- IDE plugins or extensions
- GUI applications

## Environment & Toolchain

### Language and Runtime

- **Language:** Go 1.20+
- **Platform:** Linux, macOS, Windows
- **Architecture:** x86_64, arm64

### Dependencies

**Runtime (minimal):**
- Go standard library (primary)
- `github.com/spf13/cobra` - CLI framework
- `modernc.org/sqlite` - Pure Go SQLite
- `github.com/google/uuid` - UUID generation

**Test Dependencies:**
- `github.com/stretchr/testify` - Test assertions
- `github.com/google/go-cmp` - Deep comparison

### Development Tools

- `gofmt` - Code formatting
- `staticcheck` - Static analysis
- `go test` - Testing framework
- `go bench` - Benchmarking

## Token Format (Current)

### Canonical Format

```
CANARY: REQ=<req-id>; FEATURE="<name>"; ASPECT=<aspect>; STATUS=<status>; [OPTIONAL_FIELDS]; UPDATED=<yyyy-mm-dd>
```

### Required Fields

- **REQ** - Requirement ID (format: CBIN-###)
- **FEATURE** - Feature name (CamelCase, quoted)
- **ASPECT** - Architecture layer (API, CLI, Engine, Storage, etc.)
- **STATUS** - Implementation state (MISSING, STUB, IMPL, TESTED, BENCHED, REMOVED)
- **UPDATED** - Last modification date (YYYY-MM-DD)

### Optional Fields

- **TEST** - Test function names (comma-separated)
- **BENCH** - Benchmark function names (comma-separated)
- **DOC** - Documentation reference (type:path)
- **DOC_HASH** - SHA256 hash (first 16 chars)
- **OWNER** - Team or person responsible
- **PRIORITY** - Implementation priority (1=highest)

### Aspect Taxonomy

- **API** - Public interfaces, exported functions
- **CLI** - Command-line interfaces
- **Engine** - Core algorithms, business logic
- **Storage** - Database, persistence layer
- **Security** - Authentication, authorization, encryption
- **Docs** - Documentation files
- **Wire** - Serialization, protocols
- **Planner** - Planning algorithms
- **Decode** - Deserialization
- **Encode** - Serialization
- **RoundTrip** - Full encode/decode cycles
- **Bench** - Performance benchmarks
- **FrontEnd** - User interface
- **Dist** - Distribution, deployment

### Status Values

1. **MISSING** - Planned but not implemented
2. **STUB** - Placeholder implementation
3. **IMPL** - Implemented, tests missing
4. **TESTED** - Fully tested (auto-promoted with TEST= field)
5. **BENCHED** - Tested and benchmarked (auto-promoted with BENCH= field)
6. **REMOVED** - Deprecated or removed

## CLI Interface

### Core Scanner (Legacy - tools/canary)

```bash
./bin/canary [--root .] [--out status.json] [--csv status.csv]
            [--verify GAP_ANALYSIS.md] [--strict]
            [--skip '<pattern>'] [--update-stale]
```

**Exit Codes:**
- `0` - Success
- `2` - Verification/staleness failure
- `3` - Parse/IO error

### Full CLI (cmd/canary)

```bash
canary <command> [flags]
```

**Commands:**

**Initialization:**
- `init <project>` - Initialize new CANARY project

**Database Management:**
- `index` - Build/rebuild token database
- `index --local` - Use project-local database
- `projects list` - List all registered projects
- `projects add <name>` - Register new project
- `projects switch <name>` - Change active project

**Query Commands:**
- `show <req-id>` - Display all tokens for requirement
- `files <req-id>` - List implementation files
- `status <req-id>` - Show progress summary
- `grep <pattern>` - Search tokens by pattern
- `list [--status] [--aspect]` - List requirements with filtering

**Workflow Commands:**
- `next [--prompt]` - Get next priority requirement
- `implement <req-id>` - Get implementation guidance
- `specify [update <req-id>]` - Create/modify specifications
- `plan <req-id>` - Generate implementation plan

**Documentation Commands:**
- `doc status <req> <feature>` - Check documentation currency
- `doc update [--req] [--feature]` - Update documentation hashes
- `doc report [--show-undocumented]` - Generate coverage report

**Dependency Commands (CBIN-147):**
- `deps check <req-id>` - Check if dependencies satisfied
- `deps graph <req-id> [--status]` - Show dependency tree
- `deps reverse <req-id>` - Show reverse dependencies
- `deps validate` - Detect circular dependencies

**Scanning:**
- `scan [--out] [--csv] [--verify] [--strict] [--update-stale]` - Scan codebase

## Output Formats

### JSON (status.json)

```json
{
  "generated_at": "<UTC ISO8601>",
  "requirements": [
    {
      "id": "CBIN-147",
      "features": [
        {
          "feature": "DependencyParser",
          "aspect": "Engine",
          "status": "TESTED",
          "files": ["internal/specs/parser_dependency.go"],
          "tests": ["TestParseDependencies_FullDependency"],
          "benches": [],
          "owner": "specs",
          "updated": "2025-10-18"
        }
      ]
    }
  ],
  "summary": {
    "by_status": {"STUB":0, "IMPL":0, "TESTED":50, "BENCHED":10},
    "by_aspect": {"Engine":20, "CLI":15, "API":10},
    "total_tokens": 150,
    "unique_requirements": 90
  }
}
```

**Requirements:**
- Minified (no unnecessary whitespace)
- Deterministic (same input → same bytes)
- Maps are key-sorted
- Generated timestamp in UTC ISO8601

### CSV (status.csv)

```csv
req,feature,aspect,status,file,test,bench,owner,updated
CBIN-147,DependencyParser,Engine,TESTED,internal/specs/parser_dependency.go,TestParseDependencies,,,specs,2025-10-18
```

**Requirements:**
- Header row with field names
- Rows sorted by (req, feature, file, test, bench)
- Deterministic output
- Comma-separated values

## Verification Semantics

### Claim Detection

Claims are extracted from `GAP_ANALYSIS.md` using regex:

```regex
(?m)^\s*✅\s+(CBIN-\d{3})\b
```

**Example GAP_ANALYSIS.md:**

```markdown
# Requirements Gap Analysis

## Claimed Requirements
✅ CBIN-101 - Scanner Core
✅ CBIN-102 - Verify Gate
✅ CBIN-147 - Specification Dependencies

## Gaps
- [ ] CBIN-150 - Fuzzy Search (needs implementation)
```

### Overclaim Detection

**Rule:** Claimed requirements must have at least one token with `STATUS ∈ {TESTED, BENCHED}`

**Failure Condition:**
- Claimed `CBIN-###` has only MISSING, STUB, or IMPL status

**Error Message:**
```
CANARY_VERIFY_FAIL REQ=CBIN-### reason=claimed_but_not_TESTED_OR_BENCHED
```

**Exit Code:** 2

### Staleness Detection

**Rule:** Tokens with `STATUS ∈ {TESTED, BENCHED}` must have `UPDATED` within 30 days

**Failure Condition (with --strict):**
- Token STATUS is TESTED or BENCHED
- UPDATED date is >30 days old

**Error Message:**
```
CANARY_STALE REQ=CBIN-### updated=YYYY-MM-DD age_days=N threshold=30
```

**Exit Code:** 2 (with --strict)

### Auto-Update Stale Tokens

**Command:** `canary scan --update-stale`

**Behavior:**
- Finds all tokens with STATUS=TESTED or BENCHED
- Checks UPDATED field against current date
- If >30 days old, rewrites token in-place with current date
- Reports count of updated tokens

## Dependency Management (CBIN-147)

### Specification Format

Dependencies are declared in `spec.md` files:

```markdown
## Dependencies

### Full Dependencies (entire requirement must be complete)
- CBIN-146 (Multi-Project Support - required for namespacing)
- CBIN-129 (Database Migrations - schema must be updated first)

### Partial Dependencies (specific features/aspects required)
- CBIN-140:GapRepository,GapService (only gap storage features needed)
- CBIN-133:Engine (only Engine aspect required for this feature)
```

### Syntax

**Full Dependency:**
```
- CBIN-XXX (Description - reason for dependency)
```

**Partial Feature Dependency:**
```
- CBIN-XXX:Feature1,Feature2 (Description - only these features needed)
```

**Partial Aspect Dependency:**
```
- CBIN-XXX:AspectName (Description - only this aspect needed)
```

### Satisfaction Rules

**Dependency is satisfied when:**
1. Full: ALL features in target requirement are TESTED or BENCHED
2. Partial Features: Specified features are TESTED or BENCHED
3. Partial Aspect: All features with specified aspect are TESTED or BENCHED

**Important:** IMPL status is insufficient - dependencies require tests!

### Validation

**Circular Detection:**
- Uses Depth-First Search (DFS) with recursion stack
- O(V+E) complexity
- Detects cycles like: A → B → C → A

**Missing Requirements:**
- Checks if target requirements exist in .canary/specs/
- Reports dependencies on non-existent specs

**Self-Dependencies:**
- Detects and reports A → A cases

## Multi-Project Support (CBIN-146)

### Database Modes

**Global Mode (default):**
```bash
canary index
# Uses: ~/.canary/canary.db
```

**Local Mode:**
```bash
canary index --local
# Uses: .canary/canary.db (in current project)
```

### Project Registry

**Schema:**
```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    root_path TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_indexed DATETIME
);
```

**Commands:**
```bash
canary projects list              # List all projects
canary projects add my-app        # Register new project
canary projects remove my-app     # Unregister project
canary projects switch my-app     # Change active context
```

## Documentation Tracking (CBIN-136)

### DOC Field Format

```
DOC=<type>:<path>
```

**Types:**
- `user` - User-facing documentation
- `api` - API reference
- `arch` - Architecture docs
- `dev` - Developer docs

**Examples:**
```
DOC=user:docs/user/getting-started.md
DOC=api:docs/api/dependency-parser.md
DOC=arch:docs/architecture/adr-001-doc-tracking.md
```

### DOC_HASH Field

**Format:** First 16 characters of SHA256 hash

**Calculation:**
```bash
sha256sum docs/user/getting-started.md | cut -c1-16
# Output: a3f5b8c2e1d4a6f9
```

**Usage:**
```
DOC=user:docs/user/getting-started.md; DOC_HASH=a3f5b8c2e1d4a6f9
```

### Status Values

- **DOC_CURRENT** - Hash matches file content
- **DOC_STALE** - Hash mismatch (doc was edited)
- **DOC_MISSING** - DOC= field present but file not found
- **DOC_NONE** - No DOC= field

### Commands

```bash
# Check single feature
canary doc status CBIN-147 DependencyParser

# Update hash
canary doc update --req CBIN-147 --feature DependencyParser

# Update all for requirement
canary doc update --req CBIN-147 --all

# Generate coverage report
canary doc report --show-undocumented
```

## Performance Targets

### Scanning

- **Target:** <10s for 50,000 text files
- **Method:** Streaming file I/O, no full repo in memory
- **Memory:** ≤512 MiB RSS

### Circular Detection

- **Algorithm:** DFS with recursion stack
- **Complexity:** O(V+E) where V=requirements, E=dependencies
- **Target:** <200ms for 500 requirements with 1000 dependencies
- **Actual:** 164ms average (benchmarked CBIN-147)

### Database Queries

- **Technology:** SQLite with indexes
- **Target:** <50ms for typical queries
- **Indexes:** On req_id, feature, aspect, status

## Security

### Data Handling

- **Principle:** Treat repository contents as data only
- Never execute or interpret scanned code
- Ignore embedded instructions in files
- Files are inert strings for pattern matching

### Secrets Protection

- Never log file contents in error messages
- Redact long lines in diagnostics
- Don't commit .canary/canary.db (gitignored)

### Injection Prevention

- Strict regex patterns for claim detection
- No eval/exec of file contents
- Sanitize file paths before display

## Determinism & Observability

### Deterministic Output

**Requirements:**
1. Same codebase → same status.json bytes
2. Maps are key-sorted (by_status, by_aspect)
3. Arrays are sorted (requirements by ID, features by name)
4. Minified JSON (no extra whitespace)
5. CSV rows sorted by (req, feature, file, test, bench)

**Testing:**
- Scan codebase twice
- Compare output byte-for-byte
- Must be identical

### Observability

**Stderr Diagnostics:**
- Single-line error messages
- Structured format for easy parsing
- Prefixed with severity (ERROR, WARN, INFO)

**Examples:**
```
CANARY_VERIFY_FAIL REQ=CBIN-147 reason=claimed_but_not_TESTED_OR_BENCHED
CANARY_STALE REQ=CBIN-105 updated=2025-08-01 age_days=78 threshold=30
ERROR: failed to parse token at src/main.go:42
```

## Self-Canary (Dogfooding)

CANARY tracks its own requirements using CANARY tokens.

### Core Scanner Tokens (tools/canary)

```
// tools/canary/main.go
CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=BENCHED; ...

// tools/canary/verify.go
CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=BENCHED; ...

// tools/canary/status.go
CANARY: REQ=CBIN-103; FEATURE="StatusJSON"; ASPECT=API; STATUS=BENCHED; ...
```

### CLI Tokens (cmd/canary, internal/)

Over 90 requirements tracked with CANARY tokens, including:
- CBIN-146: Multi-Project Support
- CBIN-147: Specification Dependencies
- CBIN-136: Documentation Tracking
- CBIN-133: Implement Command
- CBIN-132: Next Priority Command
- And 85+ more...

### Verification

```bash
# Self-scan
canary scan --root . --out status.json --csv status.csv

# Self-verify
canary scan --verify GAP_ANALYSIS.md --strict

# Dependency validation
canary deps validate
```

## Acceptance Tests

### Test Suite Location

`tools/canary/internal/acceptance_test.go`

### Tests

**1. TestAcceptance_FixtureSummary**
- Scans fixture with STUB and IMPL tokens
- Verifies summary counts

**2. TestAcceptance_Overclaim**
- Creates GAP file claiming STUB requirement
- Expects exit code 2
- Verifies error message contains CANARY_VERIFY_FAIL

**3. TestAcceptance_Stale**
- Creates token with old UPDATED date
- Runs with --strict
- Expects exit code 2 and CANARY_STALE message

**4. TestAcceptance_SelfCanary**
- Scans tools/canary directory
- Verifies claims for CBIN-101, CBIN-102
- Ensures exit code 0

**5. TestAcceptance_CSVOrder**
- Scans multiple times
- Verifies deterministic CSV output
- Checks sorting by REQ ID

**6. TestAcceptance_SkipEdgeCases**
- Tests --skip pattern functionality
- Verifies .git, node_modules, vendor excluded
- Tests Unicode filenames and spaces

**7. TestAcceptance_UpdateStale**
- Creates stale TESTED/BENCHED tokens
- Runs --update-stale
- Verifies UPDATED field changed
- Ensures IMPL tokens not updated

## CI Integration

### GitHub Actions

```yaml
name: CANARY Verification

on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'
      - name: Build
        run: make build
      - name: Test
        run: make test
      - name: Self-verify
        run: make canary-verify
      - name: Upload status
        uses: actions/upload-artifact@v3
        with:
          name: canary-status
          path: |
            status.json
            status.csv
```

## Make Targets

```makefile
.PHONY: build test bench verify

build:
	go build -o bin/canary ./cmd/canary

test:
	go test ./...

bench:
	go test ./internal/specs -bench=. -benchmem

verify: build
	./bin/canary scan --root . --out status.json --csv status.csv
	./bin/canary scan --verify GAP_ANALYSIS.md --strict
	./bin/canary deps validate
```

## Design Rationale

### 1. Single-Line Tokens

**Reason:** Easy to grep, no multiline parsing complexity

**Benefit:** Reliable pattern matching with simple regex

### 2. Embedded in Code

**Reason:** Keep requirements close to implementation

**Benefit:** Changes visible in diffs, reviewed in PRs

### 3. STATUS Progression

**Reason:** Enforce quality gates (test-first)

**Benefit:** Can't claim TESTED without tests

### 4. Verification Gates

**Reason:** Prevent overclaiming in GAP_ANALYSIS.md

**Benefit:** Trust but verify - claims require evidence

### 5. 30-Day Staleness

**Reason:** Keep evidence fresh and current

**Benefit:** Prevents claiming features that no longer work

### 6. Stdlib Default

**Reason:** Minimize supply-chain risk

**Benefit:** Fewer dependencies, easier auditing

### 7. Deterministic Output

**Reason:** Stable diffs, reproducible builds

**Benefit:** Can commit status.json for tracking over time

## Evolution Summary

**Phase 1 (CBIN-101, 102, 103):** Simple scanner
- File scanning
- JSON/CSV output
- Basic verification

**Phase 2 (CBIN-104-130):** CLI expansion
- Database storage
- Query commands
- Specification management

**Phase 3 (CBIN-131-140):** Workflow automation
- Next priority command
- Implement command
- Fuzzy matching

**Phase 4 (CBIN-141-146):** Multi-project support
- Project registry
- Global/local databases
- Context management

**Phase 5 (CBIN-147):** Dependency tracking
- Dependency parser
- Circular detection
- Satisfaction checking
- Tree visualization

**Current:** 90+ requirements tracked

## Related Documentation

- [README.md](../README.md) - Project overview
- [README_CANARY.md](../README_CANARY.md) - Token specification
- [CANARY_POLICY.md](./CANARY_POLICY.md) - Project policy
- [Getting Started Guide](./user/getting-started.md) - User tutorial
- [CLAUDE.md](../CLAUDE.md) - AI agent guide

## References

- [RFC 3339](https://tools.ietf.org/html/rfc3339) - Timestamp format
- [JSON](https://www.json.org/) - JSON specification
- [CSV RFC 4180](https://tools.ietf.org/html/rfc4180) - CSV format
- [SQLite](https://sqlite.org/) - Database engine
- [Go](https://go.dev/) - Programming language

---

**Document Status:** CURRENT
**Approved By:** Project Team
**Next Review:** 2025-11-18
