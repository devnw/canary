---
description: Manage documentation tracking, creation, and verification for CANARY requirements
---

# Documentation Management Command

**Command:** `/canary.doc`
**Purpose:** Manage documentation tracking, creation, and verification for CANARY requirements

## Description

Create, update, and verify documentation associated with CANARY requirements. The documentation tracking system uses SHA256 hashing to detect when documentation becomes stale and needs updating.

## Project Configuration

Read `.canary/project.yaml` to determine the project key. Use the project key for all requirement ID prefixes (e.g., `<PROJECT_KEY>-<NNN>`).

## Usage Patterns

### Create Documentation

When you need to create documentation for a requirement:

```
/canary.doc create <REQ-ID> --type user
```

This will:
1. Identify the requirement specification
2. Create documentation from the appropriate template
3. Calculate the initial SHA256 hash
4. Provide instructions for adding DOC= field to CANARY tokens

### Update Documentation Hash

After editing documentation files:

```
# Update specific requirement
/canary.doc update <REQ-ID>

# Update all documentation (batch operation)
/canary.doc update --all

# Update only stale documentation (batch operation)
/canary.doc update --all --stale-only
```

This will:
1. Recalculate SHA256 hashes for all documentation files
2. Update the database with new hashes
3. Mark documentation as DOC_CURRENT

Batch operations allow updating multiple requirements at once:
- `--all`: Update all documentation in the database
- `--stale-only`: Only update documentation that has changed (requires `--all`)

### Check Documentation Status

To verify documentation freshness:

```
/canary.doc status <REQ-ID>
/canary.doc status --all
/canary.doc status --all --stale-only
```

This will:
1. Check all documentation files for staleness
2. Report DOC_CURRENT, DOC_STALE, DOC_MISSING, or DOC_UNHASHED status
3. Provide summary statistics

### Generate Documentation Report

To get comprehensive coverage and health metrics:

```
/canary.doc report
/canary.doc report --format json
/canary.doc report --show-undocumented
```

This will:
1. Calculate documentation coverage percentage
2. Breakdown documentation by type (user, api, technical, etc.)
3. Show staleness statistics (current, stale, missing, unhashed)
4. List undocumented requirements (with --show-undocumented flag)
5. Provide recommendations for improvement

## Documentation Types

The system supports five documentation types:

1. **user** - User-facing documentation
   - How to use features
   - Quick start guides
   - Troubleshooting

2. **api** - API reference documentation
   - Function signatures
   - Parameters and return values
   - Code examples

3. **technical** - Technical design documentation
   - Architecture
   - Implementation details
   - Performance considerations

4. **feature** - Feature specifications
   - User stories
   - Acceptance criteria
   - Functional requirements

5. **architecture** - Architecture Decision Records (ADR)
   - Context and decision
   - Alternatives considered
   - Consequences

## CANARY Token Integration

Documentation is linked to CANARY tokens using DOC= and DOC_HASH= fields:

```go
// CANARY: REQ=<PROJECT_KEY>-<NNN>; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL;
//         DOC=user:docs/user/auth.md; DOC_HASH=8f434346648f6b96;
//         UPDATED=2025-10-16
```

### Multiple Documentation Files

A single requirement can reference multiple documentation files:

```go
// CANARY: REQ=<PROJECT_KEY>-<NNN>; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL;
//         DOC=user:docs/user/auth.md,api:docs/api/auth.md;
//         DOC_HASH=8f434346,a1b2c3d4;
//         UPDATED=2025-10-16
```

## Workflow Examples

### Example 1: Creating User Documentation

```
/canary.doc create <REQ-ID> --type user
```

**Agent should:**
1. Run `canary doc create <REQ-ID> --type user --output docs/user/authentication.md`
2. Edit the generated template with actual content
3. Add DOC= field to CANARY token in source code
4. Run `canary doc update <REQ-ID>` to register the hash

### Example 2: Checking Stale Documentation

```
/canary.doc status --all --stale-only
```

**Agent should:**
1. Run `canary doc status --all --stale-only`
2. Review list of stale documentation
3. For each stale doc:
   - Open and update the file
   - Run `canary doc update <REQ-ID>` to update hash
4. Verify all documentation is current

### Example 3: Updating After Code Changes

When implementing a feature:

```
# After code changes
/canary.doc update <REQ-ID>
```

**Agent should:**
1. Review what changed in the code
2. Update corresponding documentation files
3. Run `canary doc update <REQ-ID>` to recalculate hashes
4. Verify documentation status shows DOC_CURRENT

## Constitutional Compliance

**Article VII - Documentation Currency:**

> "CANARY tokens must maintain current UPDATED fields when implementation
> changes. Stale tokens (>30 days) should be flagged and updated."

Documentation tracking extends this principle:
- Documentation files are hashed when created
- Hashes are stored in DOC_HASH= field
- Staleness is automatically detected when file changes
- Regular verification prevents documentation drift

## Implementation Notes

**Hash Calculation:**
- SHA256 algorithm
- Line endings normalized (CRLF to LF)
- Abbreviated to first 16 characters (64 bits)
- Sufficient collision resistance for documentation tracking

**Performance:**
- Hash calculation: <0.01ms per KB
- Suitable for large documentation sets
- Batch operations supported

**Database Storage:**
- DocPath: Comma-separated paths with type prefixes
- DocHash: Comma-separated abbreviated SHA256 hashes
- DocType: Documentation type (user, api, technical, etc.)
- DocCheckedAt: ISO 8601 timestamp of last check
- DocStatus: Current status (CURRENT, STALE, MISSING, UNHASHED)

## Error Handling

**Missing Files:**
- Status: DOC_MISSING
- Agent should: Create documentation or update token to remove reference

**Unhashed Documentation:**
- Status: DOC_UNHASHED
- Agent should: Calculate hash and add DOC_HASH= field

**Stale Documentation:**
- Status: DOC_STALE
- Agent should: Update documentation content, then run `canary doc update`

## Command Reference

### canary doc create

**Syntax:** `canary doc create <REQ-ID> --type <type> --output <path>`

**Arguments:**
- `<REQ-ID>`: Requirement identifier (e.g., `<PROJECT_KEY>-<NNN>`)
- `--type`: Documentation type (user, api, technical, feature, architecture)
- `--output`: Output file path

**Example:**
```bash
canary doc create <REQ-ID> --type user --output docs/user/auth.md
```

### canary doc update

**Syntax:** `canary doc update [REQ-ID] [--all] [--stale-only]`

**Arguments:**
- `[REQ-ID]`: Optional requirement identifier (required if not using --all)
- `--all`: Update all documentation in database
- `--stale-only`: Only update stale documentation (requires --all)

**Examples:**
```bash
# Update specific requirement
canary doc update <REQ-ID>

# Update all documentation
canary doc update --all

# Update only stale documentation
canary doc update --all --stale-only
```

### canary doc status

**Syntax:** `canary doc status [REQ-ID] [--all] [--stale-only]`

**Arguments:**
- `[REQ-ID]`: Optional requirement identifier
- `--all`: Check all requirements
- `--stale-only`: Show only stale documentation

**Examples:**
```bash
canary doc status <REQ-ID>
canary doc status --all
canary doc status --all --stale-only
```

### canary doc report

**Syntax:** `canary doc report [--format <format>] [--show-undocumented]`

**Arguments:**
- `--format`: Output format (text or json), defaults to text
- `--show-undocumented`: Show list of undocumented requirements

**Examples:**
```bash
# Generate text report
canary doc report

# Generate JSON report for scripting
canary doc report --format json

# Show undocumented requirements
canary doc report --show-undocumented
```

## Integration with Other Commands

**After /canary.specify:**
```
/canary.specify "Add user authentication"
/canary.doc create <REQ-ID> --type feature
```

**After /canary.plan:**
```
/canary.plan <REQ-ID>
/canary.doc create <REQ-ID> --type technical
```

**Before /canary.verify:**
```
/canary.doc status --all
# Fix any stale documentation
/canary.verify
```

## Best Practices

1. **Create Documentation Early**
   - Create feature docs during specification
   - Create technical docs during planning
   - Create API docs before implementation

2. **Keep Documentation Current**
   - Update docs when code changes
   - Run `canary doc status --all` regularly
   - Fix stale documentation immediately

3. **Use Appropriate Types**
   - User documentation for end-users
   - API documentation for developers
   - Technical documentation for maintainers
   - Architecture documentation for decisions

4. **Link Documentation to Code**
   - Always include DOC= in CANARY tokens
   - Reference multiple docs when needed
   - Keep paths relative to project root

5. **Verify Before Release**
   - Check documentation status before commits
   - Include documentation in code review
   - Verify all documentation is DOC_CURRENT

## Troubleshooting

**Problem:** Documentation shows as DOC_STALE but hasn't changed

**Solution:** Line endings may differ. The hash normalizes CRLF to LF automatically, but git autocrlf settings can cause issues. Ensure consistent line endings.

**Problem:** Can't find documentation template

**Solution:** Templates are in `.canary/templates/docs/`. If missing, run `canary init` to regenerate templates.

**Problem:** Database doesn't have doc fields

**Solution:** Run database migration: `canary migrate all` or rebuild: `canary index`

## See Also

- `/canary.specify` - Create requirement specifications
- `/canary.plan` - Generate implementation plans
- `/canary.verify` - Verify GAP_ANALYSIS.md claims
- `/canary.scan` - Scan for CANARY tokens
