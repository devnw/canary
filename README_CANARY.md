# CANARY Token Specification

**Version:** 2.0
**Last Updated:** 2025-10-18

## Purpose

CANARY tokens are structured comments embedded in source code that make requirements traceable, searchable, and verifiable. They bridge the gap between specifications and implementation by placing requirement metadata directly at the point of implementation.

## Token Format

A CANARY token is a single-line structured comment:

```
CANARY: REQ=<req-id>; FEATURE="<name>"; ASPECT=<aspect>; STATUS=<status>; [OPTIONAL_FIELDS]; UPDATED=<yyyy-mm-dd>
```

### Complete Example

```go
// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_FullDependency,TestParseDependencies_PartialFeatures; OWNER=specs; UPDATED=2025-10-18
func ParseDependencies(sourceReqID string, reader io.Reader) ([]Dependency, error) {
    // implementation
}
```

## Required Fields

### REQ (Required)

**Format:** `CBIN-###` where ### is a 3-digit number

**Purpose:** Unique requirement identifier

**Examples:**
- `REQ=CBIN-001` - First requirement
- `REQ=CBIN-147` - Specification Dependencies feature
- `REQ=CBIN-105` - Fuzzy search feature

**Rules:**
- Must be uppercase
- Must follow `CBIN-` prefix pattern
- Number must be zero-padded to 3 digits

### FEATURE (Required)

**Format:** `FEATURE="<CamelCaseName>"`

**Purpose:** Short, descriptive feature name

**Examples:**
- `FEATURE="UserAuth"`
- `FEATURE="DependencyParser"`
- `FEATURE="FuzzySearch"`

**Rules:**
- Must be quoted
- Use CamelCase (no spaces, underscores, or hyphens)
- Keep under 40 characters
- Should be unique within a requirement

### ASPECT (Required)

**Format:** `ASPECT=<AspectName>`

**Purpose:** Architectural layer/concern

**Valid Values:**
- `API` - Public interfaces, exported functions
- `CLI` - Command-line interfaces, terminal I/O
- `Engine` - Core algorithms, business logic
- `Storage` - Database, persistence, repositories
- `Security` - Authentication, authorization, encryption
- `Docs` - Documentation files
- `Wire` - Serialization, protocols, networking
- `Planner` - Planning, scheduling, algorithms
- `Decode` - Deserialization, unmarshaling
- `Encode` - Serialization, marshaling
- `RoundTrip` - Full encode/decode cycles
- `Bench` - Performance benchmarks
- `FrontEnd` - User interface components
- `Dist` - Distribution, deployment, packaging

**Examples:**
- `ASPECT=Engine` - Core implementation
- `ASPECT=CLI` - Command-line tool
- `ASPECT=Storage` - Database layer

### STATUS (Required)

**Format:** `STATUS=<StatusValue>`

**Purpose:** Implementation progress tracking

**Valid Values (in progression order):**

1. **MISSING** - Planned but not implemented
   *Use when:* Token placed in spec before implementation exists

2. **STUB** - Placeholder implementation
   *Use when:* Basic structure exists but not functional

3. **IMPL** - Implemented but untested
   *Use when:* Code exists but TEST= field is missing

4. **TESTED** - Implemented with passing tests
   *Use when:* TEST= field present and tests pass

5. **BENCHED** - Tested and benchmarked
   *Use when:* BENCH= field present and benchmarks run

6. **REMOVED** - Deprecated/removed feature
   *Use when:* Feature was removed but token preserved for history

**Automatic Promotion:**
- Adding `TEST=` field automatically promotes `IMPL` → `TESTED`
- Adding `BENCH=` field automatically promotes `TESTED` → `BENCHED`

**Examples:**
- `STATUS=STUB` - Not yet functional
- `STATUS=TESTED` - Fully tested
- `STATUS=BENCHED` - Tested and benchmarked

### UPDATED (Required)

**Format:** `UPDATED=YYYY-MM-DD`

**Purpose:** Track currency of implementation

**Examples:**
- `UPDATED=2025-10-18`
- `UPDATED=2025-01-01`

**Rules:**
- Must be ISO 8601 date format
- Update when modifying implementation
- Used for staleness detection (>30 days)

**Staleness:**
- Tokens with STATUS=TESTED or BENCHED and UPDATED >30 days old trigger warnings
- Use `canary scan --update-stale` to auto-update

## Optional Fields

### TEST (Optional, promotes to TESTED)

**Format:** `TEST=<TestFunctionName>[,<TestFunctionName>...]`

**Purpose:** Link implementation to test functions

**Naming Convention:** `TestCANARY_CBIN_<###>_<Short>`

**Examples:**
- `TEST=TestCANARY_CBIN_147_Engine_ParseFull`
- `TEST=TestParseDependencies_FullDependency,TestParseDependencies_PartialFeatures`
- `TEST=TestUserAuth`

**Rules:**
- Adding TEST= automatically changes STATUS from IMPL → TESTED
- Multiple tests separated by commas (no spaces)
- Tests must exist and pass

### BENCH (Optional, promotes to BENCHED)

**Format:** `BENCH=<BenchmarkFunctionName>[,<BenchmarkFunctionName>...]`

**Purpose:** Link implementation to benchmark functions

**Naming Convention:** `BenchmarkCANARY_CBIN_<###>_<Short>`

**Examples:**
- `BENCH=BenchmarkCANARY_CBIN_147_Engine_CircularDetection`
- `BENCH=BenchmarkParseDependencies,BenchmarkValidateDependencies`

**Rules:**
- Adding BENCH= automatically changes STATUS from TESTED → BENCHED
- Multiple benchmarks separated by commas (no spaces)
- Requires TEST= field to be present

### OWNER (Optional)

**Format:** `OWNER=<team-or-person>`

**Purpose:** Track ownership/responsibility

**Examples:**
- `OWNER=backend`
- `OWNER=specs`
- `OWNER=canary`
- `OWNER=alice`

**Rules:**
- Lowercase recommended
- No spaces or special characters

### DOC (Optional)

**Format:** `DOC=<type>:<path>`

**Purpose:** Link implementation to documentation

**Types:**
- `user` - User-facing documentation
- `api` - API reference documentation
- `arch` - Architecture documentation
- `dev` - Developer documentation

**Examples:**
- `DOC=user:docs/user/getting-started.md`
- `DOC=api:docs/api/dependency-parser.md`
- `DOC=arch:docs/architecture/adr-001.md`

**Rules:**
- Path must be relative to repository root
- Use with DOC_HASH for currency tracking

### DOC_HASH (Optional, requires DOC)

**Format:** `DOC_HASH=<first-16-chars-of-sha256>`

**Purpose:** Cryptographic verification of documentation currency

**Example:**
- `DOC_HASH=a3f5b8c2e1d4a6f9`

**Generation:**
```bash
# Calculate hash
sha256sum docs/user/getting-started.md | cut -c1-16

# Or use CANARY
canary doc update --req CBIN-105 --feature UserAuth
```

**Status:**
- `DOC_CURRENT` - Hash matches file content
- `DOC_STALE` - Hash mismatch (doc was edited)
- `DOC_MISSING` - DOC= field present but file not found

### PRIORITY (Optional)

**Format:** `PRIORITY=<number>`

**Purpose:** Control implementation order

**Examples:**
- `PRIORITY=1` - Highest priority
- `PRIORITY=5` - Medium priority
- `PRIORITY=10` - Lower priority

**Rules:**
- Lower numbers = higher priority
- Used by `canary next` command
- Default is 999 if not specified

### DEPENDS_ON (Optional, deprecated)

**Format:** `DEPENDS_ON=<req-id>[,<req-id>...]`

**Status:** Deprecated in favor of specification-level dependencies

**Purpose:** Express dependencies between requirements

**Migration:** Use the Dependencies section in spec.md instead:

```markdown
## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support - required for namespacing)

### Partial Dependencies
- CBIN-140:GapRepository,GapService (only these features needed)
```

See [CBIN-147 Specification](/.canary/specs/CBIN-147-specification-dependencies/spec.md) for details.

## Token Placement

### Where to Place Tokens

**Primary Implementation File:**
```go
// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies; UPDATED=2025-10-18
func ParseDependencies(sourceReqID string, reader io.Reader) ([]Dependency, error) {
    // implementation
}
```

**Test Files:**
```go
// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_FullDependency; UPDATED=2025-10-18
func TestParseDependencies_FullDependency(t *testing.T) {
    // test implementation
}
```

**Specification Files:**
```markdown
<!-- CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-18 -->
**Feature 1: Dependency Model**
```

### Placement Guidelines

1. **One token per feature** - Each distinct feature gets its own token
2. **At point of implementation** - Place near the actual code, not just at file top
3. **In test files too** - Tests should have matching tokens
4. **Update STATUS as you go** - Change from STUB → IMPL → TESTED as work progresses
5. **Keep UPDATED current** - Change date when modifying implementation

### Anti-Patterns

❌ **Too many tokens in one file**
```go
// CANARY: REQ=CBIN-147; ...
// CANARY: REQ=CBIN-147; ...
// CANARY: REQ=CBIN-147; ...
// All in same file for unrelated features
```

✅ **One token per feature**
```go
// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ...
func ParseDependencies() { }

// Different file:
// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ...
func ValidateDependencies() { }
```

❌ **Stale STATUS**
```go
// CANARY: STATUS=IMPL; ...  (but tests exist!)
func UserAuth() { }
```

✅ **Accurate STATUS**
```go
// CANARY: STATUS=TESTED; TEST=TestUserAuth; ...
func UserAuth() { }
```

## Status Progression Examples

### Example 1: New Feature Implementation

**Step 1: Specification (STUB)**
```markdown
<!-- CANARY: REQ=CBIN-150; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-18 -->
```

**Step 2: Write Test (STUB, before implementation)**
```go
// CANARY: REQ=CBIN-150; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-18
func TestFuzzySearch(t *testing.T) {
    // Test fails - no implementation yet
}
```

**Step 3: Implement (IMPL)**
```go
// CANARY: REQ=CBIN-150; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-18
func FuzzySearch(pattern, text string) bool {
    // Implementation that makes test pass
}
```

**Step 4: Link Test (TESTED)**
```go
// CANARY: REQ=CBIN-150; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=TESTED; TEST=TestFuzzySearch; UPDATED=2025-10-18
func FuzzySearch(pattern, text string) bool {
    // Same implementation
}
```

**Step 5: Add Benchmarks (BENCHED)**
```go
// CANARY: REQ=CBIN-150; FEATURE="FuzzySearch"; ASPECT=Engine; STATUS=BENCHED; TEST=TestFuzzySearch; BENCH=BenchmarkFuzzySearch; UPDATED=2025-10-18
func FuzzySearch(pattern, text string) bool {
    // Same implementation
}
```

### Example 2: Multi-Feature Requirement

A single requirement (CBIN-147) with multiple features:

```go
// File: internal/specs/types.go
// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyCreation; UPDATED=2025-10-18
type Dependency struct { ... }

// File: internal/specs/parser_dependency.go
// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies; UPDATED=2025-10-18
func ParseDependencies() { ... }

// File: internal/specs/validator.go
// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies; UPDATED=2025-10-18
func ValidateDependencies() { ... }

// File: cmd/canary/deps.go
// CANARY: REQ=CBIN-147; FEATURE="DepsCheckCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsCheckCommand; UPDATED=2025-10-18
func createDepsCheckCommand() { ... }
```

## Querying Tokens

### Command-Line Queries

```bash
# Show all tokens for a requirement
canary show CBIN-147

# List implementation files
canary files CBIN-147

# Check progress
canary status CBIN-147

# Search by pattern
canary grep "DependencyParser"

# List by status and aspect
canary list --status TESTED --aspect Engine
```

### Grep Patterns

```bash
# Find all CANARY tokens
rg -n "CANARY:\s*REQ=" .

# Find specific requirement
rg -n "REQ=CBIN-147" .

# Find by status
rg -n "STATUS=TESTED" .

# Find by aspect
rg -n "ASPECT=Engine" .

# Find test functions
rg -n "TestCANARY_CBIN_" .

# Find benchmark functions
rg -n "BenchmarkCANARY_CBIN_" .
```

## Verification

### Scan and Verify

```bash
# Scan codebase
canary scan --out status.json --csv status.csv

# Verify claims
canary scan --verify GAP_ANALYSIS.md --strict

# Check for stale tokens
canary scan --strict

# Auto-update stale tokens
canary scan --update-stale
```

### Verification Rules

1. **Overclaim Detection**
   - Claims in GAP_ANALYSIS.md must have STATUS=TESTED or BENCHED
   - Exit code 2 if claimed but only IMPL or STUB

2. **Staleness Detection**
   - TESTED/BENCHED tokens with UPDATED >30 days trigger warnings
   - Use --strict to fail builds on stale tokens

3. **Dependency Satisfaction**
   - Dependencies only satisfied by TESTED or BENCHED status
   - IMPL is insufficient for dependency satisfaction

## Best Practices

### 1. Update UPDATED Field

Always change UPDATED when modifying implementation:

```go
// Before modification
// CANARY: REQ=CBIN-105; FEATURE="Auth"; ASPECT=API; STATUS=TESTED; TEST=TestAuth; UPDATED=2025-09-01

// After modification
// CANARY: REQ=CBIN-105; FEATURE="Auth"; ASPECT=API; STATUS=TESTED; TEST=TestAuth; UPDATED=2025-10-18
```

### 2. Use Meaningful Feature Names

```go
// ❌ Bad: Generic names
// CANARY: FEATURE="Function1"
// CANARY: FEATURE="Utils"
// CANARY: FEATURE="Helper"

// ✅ Good: Descriptive names
// CANARY: FEATURE="PasswordHasher"
// CANARY: FEATURE="JWTValidator"
// CANARY: FEATURE="SessionManager"
```

### 3. Match Test Names to Features

```go
// CANARY: REQ=CBIN-105; FEATURE="PasswordHasher"; ASPECT=Security; STATUS=TESTED; TEST=TestPasswordHasher; UPDATED=2025-10-18
func HashPassword(password string) (string, error) { }

func TestPasswordHasher(t *testing.T) {
    // Test implementation
}
```

### 4. Document TESTED Features

```go
// CANARY: REQ=CBIN-105; FEATURE="PasswordHasher"; ASPECT=Security; STATUS=TESTED;
// TEST=TestPasswordHasher; DOC=user:docs/user/auth-guide.md;
// DOC_HASH=a3f5b8c2e1d4a6f9; UPDATED=2025-10-18
```

### 5. Use Appropriate Aspects

```go
// ✅ Correct
// CANARY: ASPECT=API    (for public interfaces)
// CANARY: ASPECT=Engine (for core logic)
// CANARY: ASPECT=Storage (for database code)

// ❌ Wrong
// CANARY: ASPECT=API    (for internal helper functions)
// CANARY: ASPECT=Engine (for CLI command handlers)
```

## Token Format Reference

### Minimal Token
```
CANARY: REQ=CBIN-001; FEATURE="FeatureName"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-18
```

### Tested Token
```
CANARY: REQ=CBIN-001; FEATURE="FeatureName"; ASPECT=API; STATUS=TESTED; TEST=TestFeatureName; UPDATED=2025-10-18
```

### Fully Benchmarked Token
```
CANARY: REQ=CBIN-001; FEATURE="FeatureName"; ASPECT=API; STATUS=BENCHED; TEST=TestFeatureName; BENCH=BenchmarkFeatureName; UPDATED=2025-10-18
```

### Documented Token
```
CANARY: REQ=CBIN-001; FEATURE="FeatureName"; ASPECT=API; STATUS=TESTED; TEST=TestFeatureName; DOC=user:docs/user/feature-guide.md; DOC_HASH=a3f5b8c2e1d4a6f9; UPDATED=2025-10-18
```

### Complete Token (All Fields)
```
CANARY: REQ=CBIN-001; FEATURE="FeatureName"; ASPECT=API; STATUS=BENCHED; TEST=TestFeatureName; BENCH=BenchmarkFeatureName; DOC=user:docs/user/feature-guide.md; DOC_HASH=a3f5b8c2e1d4a6f9; OWNER=team; PRIORITY=1; UPDATED=2025-10-18
```

## Related Documentation

- [README.md](README.md) - Project overview and quick start
- [Getting Started Guide](docs/user/getting-started.md) - Complete tutorial
- [CLAUDE.md](CLAUDE.md) - AI agent integration
- [CANARY_POLICY.md](docs/CANARY_POLICY.md) - Project-wide policy
- [REQUIREMENTS.md](docs/REQUIREMENTS.md) - Original requirements

## Grammar (BNF-style)

```bnf
<token>         ::= "CANARY:" <required> <optional>* "UPDATED=" <date>
<required>      ::= "REQ=" <req-id> "; FEATURE=\"" <feature> "\"; ASPECT=" <aspect> "; STATUS=" <status> ";"
<optional>      ::= <test> | <bench> | <doc> | <doc_hash> | <owner> | <priority>
<req-id>        ::= "CBIN-" <digit><digit><digit>
<feature>       ::= <camelcase>
<aspect>        ::= "API" | "CLI" | "Engine" | "Storage" | "Security" | ...
<status>        ::= "MISSING" | "STUB" | "IMPL" | "TESTED" | "BENCHED" | "REMOVED"
<test>          ::= "TEST=" <test-name> ["," <test-name>]* ";"
<bench>         ::= "BENCH=" <bench-name> ["," <bench-name>]* ";"
<doc>           ::= "DOC=" <doc-type> ":" <path> ";"
<doc_hash>      ::= "DOC_HASH=" <hex16> ";"
<owner>         ::= "OWNER=" <identifier> ";"
<priority>      ::= "PRIORITY=" <number> ";"
<date>          ::= <yyyy> "-" <mm> "-" <dd>
```

---

**Version History:**
- 2.0 (2025-10-18): Added dependency tracking, multi-project support, documentation tracking
- 1.0 (2025-09-20): Initial specification

**See Also:**
- [CBIN-147 Specification](/.canary/specs/CBIN-147-specification-dependencies/spec.md) - Dependency tracking
- [CBIN-146 Specification](/.canary/specs/CBIN-146-multi-project-support/spec.md) - Multi-project support
- [CBIN-136 Specification](/.canary/specs/CBIN-136-documentation-tracking/spec.md) - Documentation tracking
