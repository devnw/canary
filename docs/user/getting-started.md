# CANARY CLI Getting Started Guide

**Last Updated:** 2025-10-17

## Overview

CANARY is a requirement tracking system that embeds tokens directly into source code, enabling precise tracking of features, tests, and documentation. This guide walks you through installing CANARY, understanding core concepts, and using the system effectively for both AI agents and human developers.

**What You'll Learn:**
- Installing and initializing CANARY
- Understanding CANARY tokens and their lifecycle
- Core workflow: specify → plan → implement → verify
- Querying and inspecting implementation progress
- Maintaining documentation currency

## Prerequisites

- Go 1.20 or later installed
- Git repository for your project (recommended)
- Text editor or IDE

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/canary.git
cd canary

# Install the CLI
go install ./cmd/canary

# Verify installation
canary version
```

### Quick Test

```bash
canary --help
```

You should see available commands including `init`, `specify`, `plan`, `show`, `next`, and more.

## Core Concepts

### CANARY Tokens

CANARY tokens are structured comments embedded in your code:

```go
// CANARY: REQ=CBIN-105; FEATURE="Authentication"; ASPECT=API; STATUS=TESTED; UPDATED=2025-10-17
```

**Token Fields:**
- `REQ` - Requirement ID (format: CBIN-###)
- `FEATURE` - Short feature name (CamelCase, quoted)
- `ASPECT` - Architecture aspect (API, CLI, Engine, Storage, Docs, etc.)
- `STATUS` - Implementation status (STUB, IMPL, TESTED, BENCHED)
- `TEST` - Test function name (when STATUS=TESTED)
- `BENCH` - Benchmark function name (when STATUS=BENCHED)
- `DOC` - Documentation reference (type:path, e.g., user:docs/user/guide.md)
- `DOC_HASH` - SHA256 hash of documentation (first 16 chars)
- `UPDATED` - Last modification date (YYYY-MM-DD)
- `OWNER` - Owner/team identifier

### Status Progression

Features follow a strict lifecycle:

```
STUB → IMPL → TESTED → BENCHED
```

- **STUB**: Not yet implemented (placeholder)
- **IMPL**: Implementation exists, tests missing
- **TESTED**: Implementation complete with passing tests
- **BENCHED**: Fully tested and performance benchmarked

### Valid Aspects

CANARY organizes code by architectural aspects:

- **API** - Public interfaces, exported functions
- **CLI** - Command-line interfaces
- **Engine** - Core logic and algorithms
- **Storage** - Database, persistence layer
- **Security** - Authentication, authorization, encryption
- **Docs** - Documentation files
- **Wire** - Serialization, network protocols
- **Planner** - Planning and scheduling logic
- **Bench** - Performance benchmarks
- **FrontEnd** - User interface components
- **Dist** - Distribution and deployment

## Quick Start: Your First CANARY Project

### Step 1: Initialize a New Project

```bash
# Create project directory
mkdir my-app
cd my-app

# Initialize Go module
go mod init github.com/yourusername/my-app

# Initialize CANARY
canary init my-app
```

**What this creates:**

```
my-app/
├── .canary/
│   ├── memory/
│   │   └── constitution.md     # Project principles (empty template)
│   ├── templates/
│   │   ├── spec-template.md    # Requirement specification template
│   │   └── plan-template.md    # Implementation plan template
│   └── specs/                  # Individual requirement directories
├── GAP_ANALYSIS.md              # Requirement tracking
└── README.md                    # Project documentation
```

### Step 2: Establish Project Principles

Before implementing features, define your project's governing principles:

```bash
# For AI agents (using slash commands)
/canary.constitution Create principles for test-first development and simplicity

# For human developers (manual creation)
# Edit .canary/memory/constitution.md with your team's principles
```

**Example constitution.md**:

```markdown
# Project Constitution

## Article I: Test-First Imperative
All features SHALL be implemented using test-first development.
Tests MUST be written before implementation.

## Article II: Simplicity First
Prefer simple, direct solutions over complex abstractions.
Use Go standard library when possible.

## Article III: Documentation Currency
All TESTED features MUST have current documentation.
Documentation MUST be updated when behavior changes.
```

### Step 3: Specify Your First Requirement

Use the `/canary.specify` command (AI agents) or `canary specify` CLI to create structured requirements:

```bash
# AI agent approach
/canary.specify Add user authentication with JWT tokens and password hashing

# Human developer approach (interactive)
canary specify
# Follow prompts to create requirement
```

This creates a specification at `.canary/specs/CBIN-001-user-authentication/spec.md` with:
- Requirement ID (auto-generated)
- User stories
- Functional requirements
- Acceptance criteria
- Success metrics
- Dependencies and constraints

See [Specification Modification Guide](./spec-modification-guide.md) for detailed information.

### Step 4: Create an Implementation Plan

Generate technical guidance for your requirement:

```bash
# AI agent approach
/canary.plan CBIN-001 Use bcrypt for password hashing, standard library JWT

# Human developer approach
canary plan CBIN-001
```

This creates `.canary/specs/CBIN-001-user-authentication/plan.md` with:
- Technical approach
- File structure
- Token placement strategy
- Test plan
- Implementation checklist

### Step 5: Implement the Feature

Follow test-first development:

#### A. Write Tests First (RED phase)

```go
// internal/auth/auth_test.go

// CANARY: REQ=CBIN-001; FEATURE="Authentication"; ASPECT=Security; STATUS=STUB; UPDATED=2025-10-17
func TestCANARY_CBIN_001_Security_PasswordHashing(t *testing.T) {
    password := "secure-password-123"

    // Hash password
    hashed, err := HashPassword(password)
    if err != nil {
        t.Fatalf("HashPassword failed: %v", err)
    }

    // Verify correct password
    if !VerifyPassword(hashed, password) {
        t.Error("VerifyPassword failed for correct password")
    }

    // Verify wrong password
    if VerifyPassword(hashed, "wrong-password") {
        t.Error("VerifyPassword succeeded for wrong password")
    }
}
```

Run tests (should fail):

```bash
go test ./internal/auth/
# Expected: FAIL (HashPassword and VerifyPassword don't exist yet)
```

#### B. Implement Feature (GREEN phase)

```go
// internal/auth/auth.go

// CANARY: REQ=CBIN-001; FEATURE="Authentication"; ASPECT=Security; STATUS=IMPL; UPDATED=2025-10-17
package auth

import (
    "golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash for the given password.
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// VerifyPassword checks if the password matches the hash.
func VerifyPassword(hash, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

Run tests (should pass):

```bash
go test ./internal/auth/
# Expected: PASS
```

#### C. Update Token Status to TESTED

```go
// internal/auth/auth_test.go

// CANARY: REQ=CBIN-001; FEATURE="Authentication"; ASPECT=Security; STATUS=TESTED; TEST=TestCANARY_CBIN_001_Security_PasswordHashing; UPDATED=2025-10-17
func TestCANARY_CBIN_001_Security_PasswordHashing(t *testing.T) {
    // ... test code ...
}
```

```go
// internal/auth/auth.go

// CANARY: REQ=CBIN-001; FEATURE="Authentication"; ASPECT=Security; STATUS=TESTED; TEST=TestCANARY_CBIN_001_Security_PasswordHashing; UPDATED=2025-10-17
package auth
// ... implementation ...
```

### Step 6: Index and Query Your Progress

Build the database for fast queries:

```bash
canary index
```

**Inspect your implementation:**

```bash
# Show all tokens for CBIN-001
canary show CBIN-001

# List implementation files
canary files CBIN-001

# Check completion status
canary status CBIN-001
```

See [Query Commands Guide](./query-commands-guide.md) for detailed query capabilities.

### Step 7: Scan and Verify

Generate progress reports:

```bash
# Scan codebase for all tokens
canary scan --out status.json --csv status.csv

# Verify GAP_ANALYSIS.md claims
canary scan --verify GAP_ANALYSIS.md
```

### Step 8: Document the Feature

Create user documentation:

```bash
# Create documentation file
vim docs/user/authentication-guide.md
# Write comprehensive user guide...

# Update token with DOC field
# In your source file:
// CANARY: REQ=CBIN-001; FEATURE="Authentication"; ASPECT=Security; STATUS=TESTED; TEST=TestCANARY_CBIN_001_Security_PasswordHashing; DOC=user:docs/user/authentication-guide.md; UPDATED=2025-10-17

# Calculate and add hash
canary doc update --req CBIN-001 --feature Authentication

# Verify documentation status
canary doc status CBIN-001 Authentication
# Expected: DOC_CURRENT
```

See [Documentation Tracking Guide](./documentation-tracking-guide.md) for details.

## Core Workflows

### Workflow 1: Autonomous AI Agent

AI agents can work independently using slash commands:

```
1. Agent completes current task
2. Agent runs: /canary.next
3. System returns next priority requirement with full guidance
4. Agent implements following test-first approach
5. Agent places CANARY tokens and updates status
6. Agent verifies with /canary.scan
7. Repeat from step 2
```

This creates a continuous implementation loop.

### Workflow 2: Human Developer Daily Work

```bash
# Morning: Check what's next
canary next

# Review requirement
cat .canary/specs/CBIN-105-fuzzy-search/spec.md

# Get implementation guidance
canary next --prompt > implementation-guidance.md

# Implement following guidance
# ... write tests ...
# ... write implementation ...
# ... place tokens ...

# Verify progress
canary status CBIN-105
canary scan --verify GAP_ANALYSIS.md

# Check what's next
canary next
```

See [Next Priority Guide](./next-priority-guide.md) for advanced usage.

### Workflow 3: Feature Discovery and Modification

Search and update existing requirements:

```bash
# Search for authentication-related features
canary grep Authentication

# Find specific requirement
canary list --status IMPL | grep "Auth"

# Update existing specification
canary specify update CBIN-001
# Interactive prompts guide you through modification
```

See [Specification Modification Guide](./spec-modification-guide.md) for details.

### Workflow 4: Implementation Guidance

Get detailed guidance for specific requirements:

```bash
# Get full implementation prompt
canary implement CBIN-105

# Or use fuzzy matching
canary implement fuzzy
# Selects best match for "fuzzy" keyword
```

See [Implement Command Guide](./implement-command-guide.md) for details.

## Querying and Inspecting Progress

CANARY provides four powerful query commands:

### Show All Tokens

```bash
canary show CBIN-105
```

Displays all tokens for a requirement with:
- Feature name
- Aspect
- Status
- File location and line number
- Test names
- Owner

### List Implementation Files

```bash
canary files CBIN-105
```

Shows which files contain implementations, grouped by aspect, with token counts.

### Check Progress Summary

```bash
canary status CBIN-105
```

Displays:
- Total tokens
- Count by status (STUB, IMPL, TESTED, BENCHED)
- Completion percentage
- List of incomplete work

### Search by Pattern

```bash
canary grep "FuzzyMatch"
```

Searches across:
- Requirement IDs
- Feature names
- Aspects
- Owners
- Keywords

See [Query Commands Guide](./query-commands-guide.md) for complete reference.

## Best Practices

### For AI Agents

1. **Always start with constitution** - Reference `.canary/memory/constitution.md` before implementing
2. **Use test-first workflow** - RED → GREEN → REFACTOR
3. **Update tokens immediately** - Change STATUS as you progress
4. **Run /canary.next after each completion** - Get next priority automatically
5. **Verify with /canary.scan** - Ensure tokens are correctly placed

### For Human Developers

1. **Morning routine** - Start with `canary next` to see priorities
2. **Write tests first** - Follow RED → GREEN → REFACTOR religiously
3. **Use query commands** - Inspect progress with show/files/status
4. **Update UPDATED field** - Change date when modifying implementations
5. **Document TESTED features** - Add DOC= fields for completed work
6. **Verify regularly** - Run `canary scan --verify GAP_ANALYSIS.md`

### For Teams

1. **Establish constitution early** - Define principles before implementing
2. **Use ASPECT consistently** - Agree on aspect taxonomy
3. **Review token placement** - Ensure tokens mark actual implementation locations
4. **Track documentation currency** - Use DOC= and DOC_HASH= fields
5. **Automate scanning** - Add `canary scan` to CI/CD pipeline
6. **Update priorities** - Use `canary prioritize` to reflect changing needs

## Common Pitfalls

### Pitfall 1: Skipping Tests

**Problem**: Marking STATUS=TESTED without TEST= field

**Solution**:
```go
// Wrong
// CANARY: REQ=CBIN-105; FEATURE="Search"; ASPECT=API; STATUS=TESTED; UPDATED=2025-10-17

// Correct
// CANARY: REQ=CBIN-105; FEATURE="Search"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_105_API_FuzzySearch; UPDATED=2025-10-17
```

### Pitfall 2: Stale Documentation

**Problem**: Documentation exists but hasn't been updated after code changes

**Solution**: Use documentation tracking:
```bash
canary doc status CBIN-105 Search
# Shows: DOC_STALE (hash mismatch)

# Update documentation, then:
canary doc update --req CBIN-105 --feature Search
```

### Pitfall 3: Missing Tokens in Implementation

**Problem**: Specification has tokens, implementation files don't

**Solution**: Place tokens at actual implementation locations:
```go
// internal/search/fuzzy.go

// CANARY: REQ=CBIN-105; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=TESTED; TEST=TestFuzzyMatch; UPDATED=2025-10-17
func FuzzyMatch(pattern, text string) bool {
    // Implementation...
}
```

### Pitfall 4: Inconsistent UPDATED Dates

**Problem**: Forgetting to update UPDATED field when modifying code

**Solution**: Use automated staleness detection:
```bash
canary scan --update-stale
# Automatically updates UPDATED field for tokens older than 30 days
```

## Troubleshooting

### Problem: "No tokens found"

```bash
canary show CBIN-999
# Error: no tokens found for CBIN-999
```

**Solutions:**
1. Verify requirement exists: `canary list | grep CBIN-999`
2. Check token format: `grep -r "CBIN-999" .`
3. Rebuild database: `canary index`

### Problem: Database queries slow

```bash
canary show CBIN-105
# Takes >1 second
```

**Solutions:**
1. Check database exists: `ls .canary/canary.db`
2. Rebuild index: `canary index`
3. Check file count: Database is optimized for <10,000 files

### Problem: Documentation shows as DOC_STALE

```bash
canary doc status CBIN-105 Search
# Status: DOC_STALE (hash mismatch)
```

**Solutions:**
1. Verify documentation was updated: `git diff docs/user/search-guide.md`
2. Recalculate hash: `canary doc update --req CBIN-105 --feature Search`
3. Check for typos in DOC path

## Advanced Topics

### Custom Aspects

Define project-specific aspects in your constitution:

```markdown
## Valid Aspects

- **CustomerAPI** - Customer-facing API endpoints
- **AdminAPI** - Administrative API endpoints
- **Reporting** - Report generation and analytics
```

Use consistently across all tokens.

### Priority Management

Control implementation order:

```go
// CANARY: REQ=CBIN-105; FEATURE="Search"; ASPECT=API; STATUS=STUB; PRIORITY=1; UPDATED=2025-10-17
```

Lower numbers = higher priority. Use `canary next` to automatically select highest priority.

### Dependency Tracking

Express dependencies between requirements:

```go
// CANARY: REQ=CBIN-106; FEATURE="AdvancedSearch"; ASPECT=API; STATUS=STUB; DEPENDS_ON=CBIN-105; UPDATED=2025-10-17
```

`canary next` will automatically select CBIN-105 before CBIN-106.

### CI/CD Integration

Add to your GitHub Actions workflow:

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
      - name: Install CANARY
        run: go install ./cmd/canary
      - name: Verify requirements
        run: canary scan --verify GAP_ANALYSIS.md --strict
      - name: Check documentation currency
        run: canary doc report
```

## Reference

### All CANARY Commands

```bash
canary init <project>          # Initialize new CANARY project
canary index                   # Build/rebuild token database
canary list [flags]            # List requirements with filtering
canary show <req-id>           # Display all tokens for requirement
canary files <req-id>          # List implementation files
canary status <req-id>         # Show progress summary
canary grep <pattern>          # Search tokens by pattern
canary next [flags]            # Get next priority requirement
canary implement <req-id>      # Get implementation guidance
canary specify [update]        # Create/modify specifications
canary plan <req-id>           # Generate implementation plan
canary scan [flags]            # Scan for tokens and verify
canary doc status <req> <feat> # Check documentation currency
canary doc update [flags]      # Update documentation hashes
canary doc report [flags]      # Generate documentation report
canary prioritize <req> <pri>  # Adjust requirement priority
```

### Documentation Suite

- [Getting Started](./getting-started.md) - This guide
- [Query Commands Guide](./query-commands-guide.md) - show, files, status, grep
- [Next Priority Guide](./next-priority-guide.md) - Automated workflow assistant
- [Implement Command Guide](./implement-command-guide.md) - Implementation guidance
- [Specification Modification Guide](./spec-modification-guide.md) - Creating and updating specs
- [Documentation Tracking Guide](./documentation-tracking-guide.md) - DOC= and DOC_HASH= fields

### Key Files

```
.canary/
├── memory/
│   └── constitution.md          # Project principles
├── templates/
│   ├── spec-template.md         # Requirement template
│   └── plan-template.md         # Implementation plan template
├── specs/
│   └── CBIN-XXX-feature/        # Individual requirements
│       ├── spec.md              # Specification
│       └── plan.md              # Implementation plan
└── canary.db                    # SQLite token database

GAP_ANALYSIS.md                   # Requirement tracking
CLAUDE.md                         # AI agent context
```

## Getting Help

### Command Help

```bash
canary --help                # General help
canary show --help           # Command-specific help
```

### Documentation

- [README.md](../../README.md) - Project overview
- [CLAUDE.md](../../CLAUDE.md) - AI agent guide
- [.canary/AGENT_CONTEXT.md](../../.canary/AGENT_CONTEXT.md) - Complete agent context
- [User Guides](./README.md) - All user documentation

### Community

- GitHub Issues: Report bugs and request features
- Discussions: Ask questions and share experiences

---

## Next Steps

Now that you understand the basics:

1. **Initialize your project**: `canary init my-project`
2. **Establish principles**: Edit `.canary/memory/constitution.md`
3. **Create your first requirement**: Use `/canary.specify` or `canary specify`
4. **Follow the workflow**: specify → plan → implement → verify
5. **Explore query commands**: `canary show`, `canary files`, `canary status`
6. **Use automated prioritization**: `canary next`

**For AI Agents**: Reference this guide alongside [CLAUDE.md](../../CLAUDE.md) for complete context.

**For Human Developers**: Bookmark the [Query Commands Guide](./query-commands-guide.md) and [Next Priority Guide](./next-priority-guide.md) for daily use.

Happy tracking!

---

*Last verified: 2025-10-17 with canary v0.1.0*
