# Spec-Kit Integration for Canary CLI

**Date:** 2025-10-16
**Status:** ✅ COMPLETED

## Overview

Implemented spec-kit-inspired CLI commands for the Canary requirement tracking system. The skeleton code in `cmd/`, `sub/`, and `internal/` that referenced non-existent packages has been replaced with a working implementation tailored for CANARY token management.

## Implementation

### New CLI Structure

Created `cmd/canary/main.go` with cobra-based subcommands:

```
canary
├── init      - Bootstrap projects with CANARY token templates
├── create    - Generate new CANARY token templates
└── scan      - Scan for tokens (wraps tools/canary)
```

### Commands

#### 1. `canary init <project-name>`

**Inspired by:** `/speckit.init` from spec-kit

Bootstrap a new project with CANARY token conventions:

```bash
canary init my-project
```

**Creates:**
- `README_CANARY.md` - Token format specification and usage guide
- `GAP_ANALYSIS.md` - Requirements tracking template with verification examples

**Benefits:**
- Onboards new projects quickly
- Provides documentation and examples
- Establishes consistent token conventions

#### 2. `canary create <req-id> <feature-name> [flags]`

**Inspired by:** Spec-kit's structured token generation

Generate properly formatted CANARY tokens:

```bash
canary create CBIN-105 "TokenGenerator" \
  --aspect API \
  --status IMPL \
  --owner canary \
  --test TestTokenGenerator
```

**Output:**
```go
// CANARY: REQ=CBIN-105; FEATURE="TokenGenerator"; ASPECT=API; STATUS=IMPL; TEST=TestTokenGenerator; OWNER=canary; UPDATED=2025-10-16

// Paste this above your implementation:
// func TokenGenerator() { ... }
```

**Benefits:**
- Eliminates manual token formatting errors
- Auto-populates UPDATED field with current date
- Ensures consistency across codebase
- Supports all token fields (TEST, BENCH, OWNER, etc.)

#### 3. `canary scan [flags]`

**Wraps:** Existing `tools/canary` scanner

The core scanning functionality remains at `tools/canary/main.go` for backward compatibility. The CLI command provides a cleaner interface:

```bash
# New CLI
canary scan --root . --out status.json --strict

# Legacy (still works)
go run ./tools/canary --root . --out status.json
```

**Available flags:**
- `--root` - Directory to scan
- `--out` - JSON output path
- `--csv` - Optional CSV output
- `--verify` - GAP_ANALYSIS.md verification
- `--strict` - Enforce 30-day staleness
- `--update-stale` - Auto-update stale tokens
- `--skip` - Path regex to exclude

## Spec-Kit Adaptation

The implementation adapts spec-kit principles for requirement tracking:

| Spec-Kit Feature | Canary Adaptation | Rationale |
|-----------------|-------------------|-----------|
| `/speckit.init` | `canary init` | Bootstrap projects with token conventions |
| `/speckit.specify` | *Not applicable* | CANARY tracks requirements in code, not separate specs |
| `/speckit.plan` | *Not applicable* | CANARY focuses on requirement tracking, not implementation planning |
| `/speckit.tasks` | *Not applicable* | CANARY doesn't manage task execution |
| `/speckit.create` | `canary create` | Generate properly formatted tokens |
| `/speckit.analyze` | Embedded in `scan` | Coverage analysis via status.json/csv reports |

## Technical Details

**Architecture:**
- `cmd/canary/main.go` - Cobra-based CLI with subcommands
- `tools/canary/main.go` - Core scanner (unchanged for compatibility)
- Clean separation between CLI interface and scanning logic

**Dependencies:**
- `github.com/spf13/cobra` - CLI framework (already in go.mod)
- Standard library (os, fmt, time, path/filepath, os/exec)

**Build:**
```bash
go build -o bin/canary ./cmd/canary
```

**Tests:**
- All existing tests pass (11/11)
- CLI functionality manually validated
- Backward compatibility maintained

## Removed Code

Deleted broken skeleton code that referenced non-existent packages:

```
cmd/canary/      - Skeleton CLI referencing go.devnw.com/canary/*
cmd/tmp/         - Empty placeholder
sub/             - Subcommand stubs (create, init, report, scan, update, verify, docs)
internal/        - Core logic stubs (core, acceptance, cli, gen, fixtures)
```

All referenced non-existent packages from `go.devnw.com/canary`, causing build failures.

## Benefits

1. **Developer Experience:** Clear, intuitive commands for common tasks
2. **Spec-Kit Alignment:** Familiar workflow for teams using spec-kit
3. **Reduced Errors:** Automated token generation eliminates formatting mistakes
4. **Onboarding:** `init` command provides instant documentation
5. **Backward Compatibility:** Existing `tools/canary` scanner still works
6. **Clean Build:** Repository now builds with `go build ./...` (100% success)

## Examples

### New Project Setup

```bash
# Initialize project
canary init my-service

# Create first requirement
canary create CBIN-001 "UserAuth" \
  --aspect API \
  --status IMPL \
  --owner backend

# Add token to code and scan
canary scan --root . --out status.json

# Verify implementation
canary scan --root . --verify GAP_ANALYSIS.md --strict
```

### Existing Project Migration

```bash
# Generate documentation
canary init .

# Scan existing tokens
canary scan --root . --out status.json --csv status.csv

# Find stale tokens
canary scan --root . --strict

# Auto-update stale tokens
canary scan --root . --update-stale
```

## Future Enhancements

Potential additions inspired by spec-kit:

1. **`canary docs`** - Generate HTML/Markdown reports from tokens
2. **`canary analyze`** - Detailed coverage gap analysis
3. **`canary export`** - Export to various formats (JIRA, GitHub Issues, etc.)
4. **`canary import`** - Import requirements from external sources
5. **`canary validate`** - Lint token formatting and consistency

## Conclusion

Successfully integrated spec-kit-inspired CLI commands into the Canary system while:
- ✅ Maintaining backward compatibility
- ✅ Fixing build issues (removed broken code)
- ✅ Providing immediate value (init, create commands)
- ✅ Keeping implementation simple and focused
- ✅ Aligning with spec-kit workflow principles

The canary CLI now provides a clean, professional interface for managing requirement tokens, inspired by spec-kit's specification-driven development methodology but tailored for code-embedded requirement tracking.
