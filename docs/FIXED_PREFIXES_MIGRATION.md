# Fixed Prefixes Migration Guide

## Overview

Canary is migrating from variable prefixes (e.g., `CBIN-001`, `PROJ-042`) to **fixed prefixes** for improved consistency and clarity.

### New ID System

| Type | Prefix | Purpose | Example |
|------|--------|---------|---------|
| **REQ** | `REQ` | Requirement specifications | `REQ-1`, `REQ-123` |
| **TASK** | `TASK` | Tasks within requirement plans | `TASK-1`, `TASK-456` |
| **BUG** | `BUG` | Bugs related to requirements/tasks | `BUG-1`, `BUG-789` |

## Key Changes

### 1. ID Format

**Before:**
```
// CANARY: REQ=CBIN-001; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15
// CANARY: REQ=PROJ-042; FEATURE="DataSync"; ASPECT=Storage; STATUS=TESTED; UPDATED=2025-10-20
```

**After:**
```
// CANARY: REQ-1; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15
// CANARY: TASK-1; FEATURE="SetupAuth"; ASPECT=API; STATUS=IMPL; PARENT=REQ-1; UPDATED=2025-10-16
// CANARY: BUG-1; FEATURE="AuthTimeout"; ASPECT=API; STATUS=OPEN; PARENT=REQ-1; UPDATED=2025-10-17
```

### 2. Hierarchical Relationships

#### REQ (Requirements)
- Top-level specifications
- Contains overall feature description
- Can have multiple TASKs
- Can have related BUGs

#### TASK (Tasks)
- Implementation tasks for a REQ
- Part of a requirement's implementation plan
- References parent REQ with `PARENT=REQ-###`
- Can have sub-TASKs (PARENT=TASK-###)
- Can have related BUGs

#### BUG (Bugs)
- Issues related to REQ or TASK
- References parent with `PARENT=REQ-###` or `PARENT=TASK-###`
- Tracks bug lifecycle independently

### 3. Token Format Updates

#### Full Format
```
// CANARY: <ID>; FEATURE="<name>"; ASPECT=<aspect>; STATUS=<status>; [PARENT=<parent-id>]; [TEST=<test>]; [BENCH=<bench>]; [OWNER=<owner>]; UPDATED=<date>
```

#### Examples

**Requirement:**
```
// CANARY: REQ-1; FEATURE="UserAuthentication"; ASPECT=API; STATUS=TESTED; TEST=TestREQ_1_Login; OWNER=backend; UPDATED=2025-11-01
```

**Task:**
```
// CANARY: TASK-1; FEATURE="ImplementPasswordHashing"; ASPECT=API; STATUS=IMPL; PARENT=REQ-1; OWNER=backend; UPDATED=2025-11-01
```

**Bug:**
```
// CANARY: BUG-1; FEATURE="SessionTimeout"; ASPECT=API; STATUS=OPEN; PARENT=REQ-1; SEVERITY=HIGH; OWNER=backend; UPDATED=2025-11-01
```

## Migration Steps

### For Existing Projects

#### 1. Backup Your Database
```bash
cp .canary/canary.db .canary/canary.db.backup
```

#### 2. Run Migration Tool
```bash
canary migrate up
```

This will:
- Convert all `CBIN-###`, `PROJ-###`, etc. to `REQ-###`
- Maintain numeric IDs where possible
- Create mapping for reference

#### 3. Update Token Comments
Use the provided migration script:
```bash
canary migrate tokens --dry-run    # Preview changes
canary migrate tokens               # Apply changes
```

#### 4. Update Documentation
- Update `GAP_ANALYSIS.md` references
- Update specification files
- Update README references

### For New Projects

Simply use the new format from the start:
```bash
canary init myproject
canary create REQ-1 "MyFeature" --aspect API
```

## Benefits

### 1. Clarity
- **REQ-1** immediately identifies as a requirement
- **TASK-42** clearly indicates a task
- **BUG-7** obviously represents a bug

### 2. Hierarchy
- Clear parent-child relationships
- `PARENT=REQ-1` links task to requirement
- Better dependency tracking

### 3. Simplicity
- No project-specific prefixes to remember
- Universal understanding across all Canary projects
- Easier for AI assistants to parse

### 4. Scalability
- Unlimited numeric space (REQ-1 to REQ-999999+)
- Clear categorization
- Better for large projects

## API Changes

### reqid Package

**Old API:**
```go
// Parse with variable prefix
id, err := reqid.ParseRequirementID("CBIN-001")  // RequirementID{Key: "CBIN", ID: "001"}
```

**New API:**
```go
// Parse with fixed prefix
id, err := reqid.Parse("REQ-1")    // ID{Type: "REQ", Number: 1}
id, err := reqid.Parse("TASK-42")  // ID{Type: "TASK", Number: 42}
id, err := reqid.Parse("BUG-7")    // ID{Type: "BUG", Number: 7}

// Type checking
if id.IsREQ() { ... }
if id.IsTASK() { ... }
if id.IsBUG() { ... }

// Create new IDs
req := reqid.NewREQ(1)
task := reqid.NewTASK(42)
bug := reqid.NewBUG(7)
```

### Storage Layer

**Token struct updated:**
```go
type Token struct {
    // ... existing fields ...
    ReqID    string // Now stores "REQ-1", "TASK-42", or "BUG-7"
    ParentID string // NEW: Parent REQ or TASK
    // ... existing fields ...
}
```

### CLI Commands

All commands now accept the new format:

```bash
# List requirements
canary list REQ

# List tasks
canary list TASK

# List bugs  
canary list BUG

# Show specific item
canary show REQ-1
canary show TASK-42
canary show BUG-7

# Create new items
canary create REQ "FeatureName" --aspect API
canary create TASK "TaskName" --aspect API --parent REQ-1
canary create BUG "BugTitle" --parent REQ-1 --severity HIGH
```

## Compatibility

### Backward Compatibility

The migration includes a compatibility layer:

1. **Reading Old Format**: Scanner can still read old `CBIN-###` format
2. **Automatic Conversion**: Old IDs automatically mapped to new format
3. **Warning Mode**: Shows deprecation warnings for old format
4. **Migration Reports**: Generates report of all conversions

### Breaking Changes

- **Project Configuration**: `project.yaml` no longer needs `key:` field
- **Gap Analysis**: Must use new format (`REQ-1` instead of `CBIN-001`)
- **Spec Directories**: Now named `.canary/specs/REQ-001-feature/` instead of `.canary/specs/CBIN-001-feature/`

## Test Naming

### Before
```go
func TestCANARY_CBIN_001_UserAuth(t *testing.T) { ... }
func BenchmarkCANARY_CBIN_001_UserAuth(b *testing.B) { ... }
```

### After
```go
func TestREQ_1_UserAuth(t *testing.T) { ... }
func TestTASK_42_PasswordHash(t *testing.T) { ... }
func BenchmarkREQ_1_UserAuth(b *testing.B) { ... }
```

## Example Workflow

### 1. Create Requirement
```bash
canary specify "Add user authentication with OAuth2"
```
Creates: `REQ-1` with specification in `.canary/specs/REQ-001-user-auth/spec.md`

### 2. Create Implementation Plan
```bash
canary plan REQ-1
```
Creates: `.canary/specs/REQ-001-user-auth/plan.md` with tasks:
- `TASK-1`: Setup OAuth2 library
- `TASK-2`: Implement login endpoint
- `TASK-3`: Add session management

### 3. Implement Tasks
```go
// CANARY: TASK-1; FEATURE="OAuth2Setup"; ASPECT=API; STATUS=IMPL; PARENT=REQ-1; UPDATED=2025-11-01
func setupOAuth2() { ... }

// CANARY: TASK-2; FEATURE="LoginEndpoint"; ASPECT=API; STATUS=TESTED; PARENT=REQ-1; TEST=TestTASK_2_Login; UPDATED=2025-11-01
func handleLogin() { ... }
```

### 4. Track Bugs
```bash
canary bug create "Session timeout too short" --parent REQ-1 --severity MEDIUM
```
Creates: `BUG-1` linked to `REQ-1`

```go
// CANARY: BUG-1; FEATURE="SessionTimeout"; ASPECT=API; STATUS=FIXED; PARENT=REQ-1; UPDATED=2025-11-02
const sessionTimeout = 30 * time.Minute
```

## File Structure

### Spec Directories
```
.canary/specs/
??? REQ-001-user-auth/
?   ??? spec.md
?   ??? plan.md
??? REQ-002-data-sync/
?   ??? spec.md
?   ??? plan.md
??? REQ-003-reporting/
    ??? spec.md
```

### Task Tracking
Tasks are tracked in `plan.md` within the requirement directory:

```markdown
# Implementation Plan: REQ-1

## Tasks
- [ ] TASK-1: Setup OAuth2 library (STATUS=IMPL)
- [ ] TASK-2: Implement login endpoint (STATUS=TESTED)
- [ ] TASK-3: Add session management (STATUS=STUB)
```

## Migration Timeline

### Phase 1: Compatibility (Current)
- Both old and new formats supported
- Warnings for old format usage
- Migration tools available

### Phase 2: New Format Preferred (3 months)
- New format is default
- Old format deprecated but still works
- Documentation updated

### Phase 3: Old Format Removed (6 months)
- Only new format supported
- Old format generates errors
- All tooling updated

## FAQ

### Q: Can I still use my old CBIN-### IDs?
A: During the migration period, yes. The migration tool will convert them to REQ-###.

### Q: What happens to my existing spec directories?
A: They're renamed during migration: `CBIN-001-feature/` ? `REQ-001-feature/`

### Q: How do I link a TASK to a REQ?
A: Use the `PARENT=` field: `CANARY: TASK-1; ...; PARENT=REQ-1; ...`

### Q: Can a TASK have sub-TASKs?
A: Yes! Use `PARENT=TASK-1` to create a hierarchy.

### Q: What if I have a large codebase?
A: Use the migration tool with `--dry-run` first to preview changes, then apply incrementally.

### Q: Will this break my CI?
A: No, the compatibility layer ensures old IDs still work during migration.

## Support

For migration issues:
1. Check `.canary/migration.log` for detailed logs
2. Run `canary migrate status` to see migration state
3. Use `canary migrate rollback` if needed
4. File an issue on GitHub with migration log

## Summary

The fixed prefix system (`REQ`, `TASK`, `BUG`) provides:
- ? Clearer identification
- ? Better hierarchy
- ? Simpler mental model
- ? Universal understanding
- ? Easier AI integration
- ? Scalable to any project size

**Migration is straightforward and well-supported!**
