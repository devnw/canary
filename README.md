# CANARY

**Agentic-Coding-Friendly Requirement Tracking System**

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Build](https://github.com/devnw/canary/actions/workflows/build.yml/badge.svg)](https://github.com/devnw/canary/actions/workflows/build.yml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
[![Go Reference](https://pkg.go.dev/badge/devnw.dev/canary.svg)](https://pkg.go.dev/devnw.dev/canary)
[![Version](https://img.shields.io/github/v/tag/devnw/canary?sort=semver&style=plastic)](https://github.com/devnw/canary/releases)

CANARY is a requirement tracking system that embeds tokens directly into
source code, enabling precise tracking of features, tests, benchmarks, and
documentation. This bridges the gap between requirements and implementation,
ensuring that an agent coding system has not only the ability to be precise
in the specification and planning phases but the outputs of which can be
measured and verified automatically.

The CANARY system is designed with autonomous AI agents in mind, providing
slash commands and structured data to facilitate agent workflows. It also
enforces a test-first development approach through constitutional principles,
ensuring that quality is prioritized over speed.

## Quick Start

### Installation

#### APT (Debian/Ubuntu) – Recommended

Canary publishes `.deb` packages to the APT repository at `apt.codepros.org` (same pattern as [void](https://gitlab.com/devnw/codepros/oss/void)).

```bash
# Import the Codepros APT signing key
curl -fsSL https://apt.codepros.org/codepros-keyring.gpg | sudo tee /usr/share/keyrings/codepros-archive-keyring.gpg > /dev/null

# Add the Codepros APT repository
echo "deb [signed-by=/usr/share/keyrings/codepros-archive-keyring.gpg] https://apt.codepros.org/ stable main" | sudo tee /etc/apt/sources.list.d/codepros.list

# Update and install
sudo apt update
sudo apt install canary
```

#### From source

```bash
# Install from source
go install devnw.dev/canary/cmd/canary@latest

# Or clone and build
git clone https://github.com/devnw/canary.git
cd canary
make build
```

## Repository (for the pages site)

The code is hosted on GitHub:

- Repository: [Github](https://github.com/devnw/canary)
- Code Ref: [pkg.go.dev](https://pkg.go.dev/devnw.dev/canary)

### Initialize Your Project

```bash
# Create a new project
canary init my-project

# This creates:
# .canary/
#   ├── memory/constitution.md      # Project principles
#   ├── templates/                   # Spec and plan templates
#   ├── specs/                       # Individual requirements
#   └── canary.db                    # Token database
# GAP_ANALYSIS.md                    # Requirement tracking
```

### Your First Requirement

```bash
# Create a specification (AI agent)
/canary.specify Add user authentication with JWT tokens

/canary.plan CBIN-001

/canary.implement

# The primary functions like specify, plan, and implement can be run
# via the CLI but won't really do much for a user. They are designed
# for AI agents to call programmatically.

# User's can find all of the command options through the --help flag

# Build database and query progress
canary index
canary show CBIN-001
canary status CBIN-001
```

## How It Works

### CANARY Tokens

Tokens are structured comments that track requirements:

```go
// CANARY: REQ=CBIN-105; FEATURE="UserAuth"; ASPECT=Security; STATUS=TESTED; TEST=TestUserAuth; UPDATED=2026-08-29
func AuthenticateUser(creds *Credentials) (*Session, error) {
    // implementation
}
```

#### Legacy Tokens & Normalization

For backward compatibility the scanner accepts older token patterns and normalizes them:

| Legacy Form | Normalized    |
| ----------- | ------------- |
| `CBIN-5`    | `CBIN-005`    |
| `CBIN-42`   | `CBIN-042`    |
| `REQ-7`     | `REQ-007`     |
| `REQ-12`    | `REQ-012`     |
| `REQ-GQL-4` | `REQ-GQL-004` |

Rules:

1. Bare ID segments (e.g. `REQ-7`) inside a CANARY line are accepted.
2. When both a bare legacy ID and a canonical `REQ=` key appear, the `REQ=` value wins.
3. Only the final numeric segment is zero‑padded to three digits.

Use the canonical format in new code:

```go
// CANARY: REQ=CBIN-005; FEATURE="Parser"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-18
```

Legacy forms are supported for historical tokens but should not be added to new implementations.

**Token Lifecycle:**

```
STUB → IMPL → TESTED → BENCHED
```

- **STUB**: Placeholder, not yet implemented
- **IMPL**: Implementation exists, tests missing
- **TESTED**: Fully tested with passing tests
- **BENCHED**: Tested and performance benchmarked

### Architecture Aspects

CANARY organizes code by architectural concerns:

- **API** - Public interfaces, exported functions
- **CLI** - Command-line interfaces
- **Engine** - Core algorithms and business logic
- **Storage** - Databases, persistence, repositories
- **Security** - Authentication, authorization, encryption
- **Docs** - Documentation files
- **Wire** - Serialization, protocols, networking
- **Planner** - Planning and scheduling
- **Bench** - Performance benchmarks
- **FrontEnd** - User interface
- **Dist** - Distribution and deployment

### Dependency Management (CBIN-147)

Express dependencies between requirements:

```markdown
## Dependencies

### Full Dependencies (entire requirement needed)

- CBIN-146 (Multi-Project Support - required for token namespacing)

### Partial Dependencies (specific features/aspects)

- CBIN-140:GapRepository,GapService (only gap storage needed)
- CBIN-133:Engine (only Engine aspect required)
```

**Features:**

- Circular dependency detection using DFS algorithm
- Transitive dependency resolution
- Status-based satisfaction (only TESTED/BENCHED satisfy)
- Reverse dependency queries
- ASCII tree visualization

```bash
canary deps check CBIN-147        # Check if dependencies satisfied
canary deps graph CBIN-147 --status  # Visualize dependency tree
canary deps reverse CBIN-146      # What depends on this?
canary deps validate              # Check entire graph for cycles
```

### Verification Gates

Prevent overclaiming with automatic verification:

```bash
# Scan codebase for tokens
canary scan --out status.json --csv status.csv

# Verify claims in GAP_ANALYSIS.md
canary scan --verify GAP_ANALYSIS.md --strict
# Exits with code 2 if:
# - Claimed requirements lack TESTED/BENCHED status
# - Tokens are stale (>30 days old)
```

**GAP_ANALYSIS.md Format:**

```markdown
# Requirements Gap Analysis

## Claimed Requirements

✅ CBIN-101 - Scanner Core
✅ CBIN-102 - Verify Gate

## Gaps

- [ ] CBIN-103 - Status JSON (needs tests)
```

## Core Commands

### Query and Inspection

```bash
canary show CBIN-105          # Display all tokens for a requirement
canary files CBIN-105         # List implementation files
canary status CBIN-105        # Show progress summary
canary grep "Authentication"  # Search tokens by pattern
canary list --status TESTED --aspect API  # Filtered listing
```

### Workflow Automation

```bash
canary next                   # Get next priority requirement
canary next --prompt          # Generate AI agent prompt
canary implement CBIN-105     # Get implementation guidance
canary implement fuzzy        # Fuzzy match requirement
```

### Specification Management

```bash
canary specify                # Create new specification
canary specify update CBIN-105  # Modify existing spec
canary plan CBIN-105          # Generate implementation plan
```

### Documentation Tracking

```bash
canary doc status CBIN-105 UserAuth     # Check doc currency
canary doc update --req CBIN-105        # Update doc hashes
canary doc report --show-undocumented   # Coverage report
```

### Dependency Management

```bash
canary deps check CBIN-147         # Check dependency satisfaction
canary deps graph CBIN-147 --status  # Show dependency tree
canary deps reverse CBIN-146       # Show reverse dependencies
canary deps validate               # Detect circular dependencies
```

### Multi-Project Support (CBIN-146)

```bash
# Global mode (default)
canary index                  # Uses ~/.canary/canary.db

# Local mode (project-specific)
canary index --local          # Uses .canary/canary.db

# Project management
canary projects list          # List all projects
canary projects add my-app    # Register project
canary projects switch my-app # Change context
```

## Complete Workflow

### For AI Agents

```
1. Agent runs: /canary.next
2. System returns next priority requirement with:
   - Full specification
   - Implementation plan
   - Constitution principles
   - Test-first guidance
3. Agent implements following RED-GREEN-REFACTOR
4. Agent places CANARY tokens in code
5. Agent updates token STATUS as work progresses
6. Agent verifies with /canary.scan
7. Repeat from step 1
```

### For Human Developers

```bash
# Morning routine
canary next                   # See what's next

# Review requirement
cat .canary/specs/CBIN-105-fuzzy-search/spec.md
cat .canary/specs/CBIN-105-fuzzy-search/plan.md

# Implement with test-first
# 1. Write failing test (RED)
# 2. Implement minimum code to pass (GREEN)
# 3. Refactor (REFACTOR)
# 4. Add CANARY tokens
# 5. Update STATUS field

# Verify progress
canary status CBIN-105
canary scan --verify GAP_ANALYSIS.md

# Check what's next
canary next
```

## Key Features

### 🎯 Test-First Enforcement

Constitutional principles ensure tests before implementation:

```markdown
## Article IV: Test-First Imperative

All features SHALL be implemented using test-first development (TDD).
Tests MUST be written before implementation code.
```

### 📊 Real-Time Progress Tracking

```bash
canary status CBIN-105
# Output:
# Requirement: CBIN-105 (Fuzzy Search)
# Total tokens: 8
# Status breakdown:
#   TESTED: 6 (75%)
#   IMPL: 1 (12.5%)
#   STUB: 1 (12.5%)
# Incomplete work:
#   - FuzzyRanking (Engine): IMPL → needs tests
#   - FuzzyConfig (API): STUB → not implemented
```

### 🔍 Powerful Search

```bash
canary grep Authentication
# Searches across:
# - Requirement IDs
# - Feature names
# - Aspects
# - Owners
# - Files
# Returns tokens with file locations and line numbers
```

### 📚 Documentation Currency

Track documentation status with cryptographic hashes:

```go
// CANARY: REQ=CBIN-105; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=TESTED; TEST=TestFuzzySearch; DOC=user:docs/user/search-guide.md; DOC_HASH=a3f5b8c2e1d4a6f9; UPDATED=2026-08-29
```

```bash
canary doc status CBIN-105 FuzzySearch
# Status: DOC_CURRENT (hash matches)

# After editing docs/user/search-guide.md:
canary doc status CBIN-105 FuzzySearch
# Status: DOC_STALE (hash mismatch)

canary doc update --req CBIN-105 --feature FuzzySearch
# Recalculates and updates DOC_HASH
```

### 🔗 Dependency Tracking

Full dependency graph with cycle detection:

```bash
canary deps graph CBIN-147 --status
# Output:
# CBIN-147 (Specification Dependencies)
# ├── ✅ CBIN-146 (Multi-Project Support)
# │   └── ✅ CBIN-129 (Database Migrations)
# └── ✅ CBIN-140:GapRepository,GapService
#     ├── ✅ CBIN-133:Engine
#     └── ❌ CBIN-135:Storage (STATUS=IMPL, needs tests)
#
# Summary: 3 satisfied, 1 blocking
```

### 🤖 AI Agent Integration

Slash commands for autonomous workflows:

- `/canary.next` - Get next priority with full context
- `/canary.show <req-id>` - Display requirement tokens
- `/canary.status <req-id>` - Check progress
- `/canary.implement <req-id>` - Get implementation guidance
- `/canary.scan` - Verify token placement
- `/canary.specify` - Create new requirement
- `/canary.plan <req-id>` - Generate implementation plan

### 🚀 GitHub Copilot Integration

<!-- CANARY: REQ=CBIN-148; FEATURE="InitDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-19 -->

CANARY automatically configures GitHub Copilot with project-specific instructions:

```bash
canary init my-project
# Creates .github/instructions/ with CANARY workflow guidance
```

**What Gets Configured:**

- **Repository-wide instructions** - CANARY token format, test-first development, constitutional principles
- **Path-specific guidance** - Context-aware help for specs, tests, and .canary/ directory
- **Automatic discovery** - Works with both GitHub Copilot CLI and VS Code Copilot Chat

**Instruction Files Created:**

```
.github/instructions/
├── repository.md              # CANARY workflow fundamentals
├── .canary/
│   ├── instruction.md        # CANARY directory guidelines
│   └── specs/
│       └── instruction.md    # Specification writing (WHAT/WHY, not HOW)
└── tests/
    └── instruction.md        # Test-first development guidelines
```

**Verification:**

```bash
# Using GitHub Copilot CLI
gh copilot suggest "What is the CANARY token format?"

# Using VS Code Copilot Chat
# Ask: "@workspace What is the CANARY token format?"
```

**Features:**

- ✅ Zero manual configuration required
- ✅ Preserves custom instructions on re-init
- ✅ Project key substitution in templates
- ✅ Compatible with Copilot CLI and VS Code

**Re-initialization Safe:**

```bash
# Customize your instructions
echo "# Custom Rule" >> .github/instructions/repository.md

# Re-run init - your customizations are preserved
canary init --local
# ⏭️  Skipping existing instruction file: repository.md
```

## Documentation

### User Documentation

- **[Getting Started Guide](docs/user/getting-started.md)** - Installation, core concepts, complete walkthrough
- **[Query Commands Guide](docs/user/query-commands-guide.md)** - show, files, status, grep commands
- **[Next Priority Guide](docs/user/next-priority-guide.md)** - Automated workflow assistant
- **[Implement Command Guide](docs/user/implement-command-guide.md)** - Implementation guidance
- **[Specification Modification Guide](docs/user/spec-modification-guide.md)** - Creating and updating specs
- **[Documentation Tracking Guide](docs/user/documentation-tracking-guide.md)** - DOC= and DOC_HASH= fields

### Developer Documentation

- **[CLAUDE.md](CLAUDE.md)** - AI agent context and slash commands
- **[.canary/AGENT_CONTEXT.md](.canary/AGENT_CONTEXT.md)** - Complete agent reference
- **[REQUIREMENTS.md](docs/REQUIREMENTS.md)** - Original self-canary requirements
- **[CANARY_POLICY.md](docs/CANARY_POLICY.md)** - Token format and policy

### Architecture Documentation

- **[Architecture Decision Records](docs/architecture/)** - ADR-001: Documentation Tracking

## Project Structure

```
canary/
├── cmd/canary/              # Main CLI application
│   ├── main.go             # CLI entry point and command registration
│   ├── deps.go             # Dependency management commands (CBIN-147)
│   └── *_test.go           # Command tests
├── internal/
│   ├── specs/              # Specification and dependency engine
│   │   ├── types.go        # Data models (Token, Dependency, Graph)
│   │   ├── parser_dependency.go      # Dependency parser
│   │   ├── validator.go             # Circular dependency detection
│   │   ├── status_checker.go        # Dependency satisfaction
│   │   ├── graph_generator.go       # Tree visualization
│   │   └── *_test.go                # Comprehensive test suite
│   └── storage/            # SQLite database layer
│       ├── storage.go      # Database operations
│       └── migrations.go   # Schema migrations
├── .canary/
│   ├── memory/
│   │   └── constitution.md          # Project principles
│   ├── templates/
│   │   ├── spec-template.md         # Requirement template
│   │   └── plan-template.md         # Implementation plan template
│   └── specs/
│       └── CBIN-XXX-feature/        # Individual requirements
│           ├── spec.md
│           └── plan.md
├── docs/
│   ├── user/               # User-facing documentation
│   ├── architecture/       # Architecture decision records
│   └── *.md                # Various docs
├── tools/canary/           # Legacy scanner (CBIN-101, 102, 103)
├── GAP_ANALYSIS.md         # Requirement tracking
├── CLAUDE.md               # AI agent guide
└── README.md               # This file
```

## Performance

CANARY is designed for speed and efficiency:

- **Circular Detection**: O(V+E) using DFS, <209ms for 500 requirements
- **Database Queries**: SQLite with indexes, <50ms for typical queries
- **Scanning**: Streams file I/O, <10s for 50k files
- **Memory**: ≤512 MiB RSS for large repositories

## Development

### Build

```bash
make build          # Build binary
make test           # Run tests
make bench          # Run benchmarks
make verify         # Self-verify with CANARY
```

### Self-Canary

CANARY uses itself for requirement tracking:

```bash
# Scan the codebase
canary scan --root . --out status.json --csv status.csv

# Verify claims
canary scan --verify GAP_ANALYSIS.md --strict

# Check dependencies
canary deps validate
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test ./internal/specs -run Integration

# Benchmarks
go test ./internal/specs -bench=. -benchmem

# Acceptance tests
go test ./tools/canary/internal -run Acceptance -v
```

## Contributing

We welcome contributions! Please:

1. Check existing requirements: `canary list`
2. Create a specification: `canary specify`
3. Follow test-first development
4. Place CANARY tokens in your code
5. Update documentation with DOC= fields
6. Verify before submitting: `canary scan --verify GAP_ANALYSIS.md`

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for detailed guidelines.

## License

Licensed under the terms found in [LICENSE](https://github.com/devnw/canary/blob/main/LICENSE).

## Acknowledgments

CANARY was inspired by:

- **spec-kit** methodology for requirement-first development
- **Test-Driven Development (TDD)** principles
- **Evidence-based claims** from formal verification
- **Zero-trust verification** from security engineering

Built with love by [Developer Network](https://devnw.com).

## Related Projects

- [Claude Code](https://claude.com/claude-code) - AI coding assistant with CANARY integration
- [spec-kit](https://github.com/spec-kit) - Specification-driven development methodology

## Getting Help

- **Documentation**: Start with [Getting Started Guide](docs/user/getting-started.md)
- **Command Help**: `canary --help` or `canary <command> --help`
- **GitHub Issues**: [Report bugs and request features](https://github.com/devnw/canary/issues)
- **Discussions**: [Ask questions and share experiences](https://github.com/devnw/canary/discussions)

---

**Ready to start?** → [Getting Started Guide](docs/user/getting-started.md)

**For AI Agents** → [CLAUDE.md](CLAUDE.md)

**For API Documentation** → [pkg.go.dev](https://pkg.go.dev/devnw.dev/canary)

---

_CANARY: Making every feature claim searchable, verifiable, and traceable._
