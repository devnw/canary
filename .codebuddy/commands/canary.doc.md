# Documentation Management Command

**Command:** `/canary.doc`
**Purpose:** Manage documentation tracking, creation, and verification for CANARY requirements

## Description

Create, update, and verify documentation associated with CANARY requirements. The documentation tracking system uses SHA256 hashing to detect when documentation becomes stale and needs updating.

## Usage Patterns

### Create Documentation

When you need to create documentation for a requirement:

```
/canary.doc create CBIN-XXX --type user
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
/canary.doc update CBIN-XXX

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
/canary.doc status CBIN-XXX
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
// CANARY: REQ=CBIN-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL;
//         DOC=user:docs/user/auth.md; DOC_HASH=8f434346648f6b96;
//         UPDATED=2025-10-16
```

### Multiple Documentation Files

A single requirement can reference multiple documentation files:

```go
// CANARY: REQ=CBIN-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL;
//         DOC=user:docs/user/auth.md,api:docs/api/auth.md;
//         DOC_HASH=8f434346,a1b2c3d4;
//         UPDATED=2025-10-16
```

## Workflow Examples

### Example 1: Creating User Documentation

```
/canary.doc create CBIN-105 --type user
```

**Agent should:**
1. Run `canary doc create CBIN-105 --type user --output docs/user/authentication.md`
2. Edit the generated template with actual content
3. Add DOC= field to CANARY token in source code
4. Run `canary doc update CBIN-105` to register the hash

### Example 2: Checking Stale Documentation

```
/canary.doc status --all --stale-only
```

**Agent should:**
1. Run `canary doc status --all --stale-only`
2. Review list of stale documentation
3. For each stale doc:
   - Open and update the file
   - Run `canary doc update REQ-ID` to update hash
4. Verify all documentation is current

### Example 3: Updating After Code Changes

When implementing a feature:

```
# After code changes
/canary.doc update CBIN-105
```

**Agent should:**
1. Review what changed in the code
2. Update corresponding documentation files
3. Run `canary doc update CBIN-105` to recalculate hashes
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
- Line endings normalized (CRLF → LF)
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

## Integration with Other Commands

**After /canary.specify:**
```
/canary.specify "Add user authentication"
/canary.doc create CBIN-XXX --type feature
```

**After /canary.plan:**
```
/canary.plan CBIN-XXX
/canary.doc create CBIN-XXX --type technical
```

**Before /canary.verify:**
```
/canary.doc status --all
# Fix any stale documentation
/canary.verify
```

## Command Reference

### canary doc create

**Syntax:** `canary doc create <REQ-ID> --type <type> --output <path>`

**Arguments:**
- `<REQ-ID>`: Requirement identifier (e.g., CBIN-105)
- `--type`: Documentation type (user, api, technical, feature, architecture)
- `--output`: Output file path

**Example:**
```bash
canary doc create CBIN-105 --type user --output docs/user/auth.md
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
canary doc update CBIN-105

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
canary doc status CBIN-105
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

**Sample Output:**
```
📊 Documentation Report

Coverage: 6/125 requirements (4.8%)
Total Tokens: 971 (9 with docs, 119 without)

📋 Documentation Status:
  ✅ Current:  2 (50.0%)
  ⚠️  Stale:    1 (25.0%)
  ❌ Missing:  1 (25.0%)

💡 119 requirements without documentation (use --show-undocumented to list)

💡 Recommendations:
  Run 'canary doc update --all --stale-only' to update stale documentation
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

**Solution:** Line endings may differ. The hash normalizes CRLF→LF automatically, but git autocrlf settings can cause issues. Ensure consistent line endings.

**Problem:** Can't find documentation template

**Solution:** Templates are in `.canary/templates/docs/`. If missing, run `canary init` to regenerate templates.

**Problem:** Database doesn't have doc fields

**Solution:** Run database migration: `canary migrate all` or rebuild: `canary index`

## Agent Workflow Patterns

### Pattern 1: Proactive Documentation After Implementation

When you complete implementation of a feature with STATUS=IMPL or STATUS=TESTED:

1. **Check if documentation exists:**
   ```bash
   canary doc status REQ-ID
   ```

2. **If DOC_UNHASHED or no documentation:**
   - Create appropriate documentation type
   - For user-facing features: create user docs
   - For APIs: create API docs
   - For technical decisions: create architecture docs

3. **Example workflow:**
   ```bash
   # After implementing CBIN-105
   canary doc create CBIN-105 --type user --output docs/user/auth.md
   # Edit the documentation
   canary doc update CBIN-105
   # Verify
   canary doc status CBIN-105
   ```

### Pattern 2: Documentation-Driven Development

Before implementing a feature:

1. **Create feature specification:**
   ```bash
   /canary.specify "Add user authentication"
   # Creates CBIN-XXX
   ```

2. **Create feature documentation:**
   ```bash
   canary doc create CBIN-XXX --type feature --output docs/features/auth.md
   # Fill in user stories and acceptance criteria
   ```

3. **Create technical design:**
   ```bash
   canary doc create CBIN-XXX --type technical --output docs/technical/auth-impl.md
   # Document architecture and implementation approach
   ```

4. **Implement with references:**
   - Add DOC fields to CANARY tokens pointing to both docs
   - Keep documentation updated as implementation evolves

### Pattern 3: Regular Documentation Maintenance

Include in your regular workflow:

1. **Before starting new work:**
   ```bash
   canary doc status --all --stale-only
   ```
   - Fix any stale documentation before adding new features

2. **After completing work:**
   ```bash
   canary doc update --all --stale-only
   ```
   - Update hashes for any documentation you modified

3. **Before commits:**
   ```bash
   canary doc status --all
   ```
   - Ensure no stale or missing documentation

### Pattern 4: Documentation Coverage Improvement

When working to improve documentation coverage:

1. **Generate report:**
   ```bash
   canary doc report --show-undocumented
   ```

2. **For each undocumented requirement:**
   - Determine appropriate documentation type
   - Create documentation from template
   - Add DOC field to CANARY token
   - Update hash

3. **Track progress:**
   ```bash
   canary doc report
   # Monitor coverage percentage improvement
   ```

### Pattern 5: Multi-Document Features

For complex features requiring multiple documentation types:

1. **Create all documentation:**
   ```bash
   canary doc create CBIN-200 --type user --output docs/user/api-usage.md
   canary doc create CBIN-200 --type api --output docs/api/endpoints.md
   canary doc create CBIN-200 --type technical --output docs/technical/api-design.md
   ```

2. **Update CANARY token with multiple references:**
   ```go
   // CANARY: REQ=CBIN-200; FEATURE="RestAPI"; ASPECT=API; STATUS=IMPL;
   //         DOC=user:docs/user/api-usage.md,api:docs/api/endpoints.md,technical:docs/technical/api-design.md;
   //         DOC_HASH=hash1,hash2,hash3;
   //         UPDATED=2025-10-16
   ```

3. **Update all hashes together:**
   ```bash
   canary doc update CBIN-200
   ```

## Agent Decision Tree

```
┌─────────────────────────────────────┐
│ Completing feature implementation?  │
└──────────────┬──────────────────────┘
               │
               ▼
         ┌──────────┐
         │ Check    │
         │ doc      │
         │ status   │
         └────┬─────┘
              │
      ┌───────┴────────┐
      │                │
    DOC_CURRENT   DOC_STALE/MISSING
      │                │
      ▼                ▼
   Done         ┌──────────────┐
                │ What type    │
                │ of feature?  │
                └───┬──────────┘
                    │
      ┌─────────────┼─────────────┐
      │             │             │
  User-facing    API/Library  Design Decision
      │             │             │
      ▼             ▼             ▼
  Create user   Create API   Create arch
  docs          docs         docs (ADR)
      │             │             │
      └─────────────┼─────────────┘
                    │
                    ▼
              ┌──────────┐
              │ Update   │
              │ hash     │
              └──────────┘
```

## Proactive Agent Behavior

**When to automatically create documentation (without user asking):**

1. **After STATUS changes to TESTED:**
   - If no documentation exists
   - Feature is user-facing or API
   - Create user or API documentation

2. **After implementing major architectural decision:**
   - Create architecture documentation (ADR)
   - Document context, decision, alternatives

3. **When you notice stale documentation:**
   - Proactively update if changes are minor
   - Ask user if changes are significant

**When to ask user first:**

1. **Documentation type unclear:**
   - Multiple types might apply
   - User preference needed

2. **Large documentation changes:**
   - Significant rewrites needed
   - User should review scope

3. **Missing critical information:**
   - You don't have enough context
   - User input required for accuracy

## Common Mistakes to Avoid

1. **Forgetting to update hash after editing:**
   - Always run `canary doc update` after editing documentation
   - Hash must match file content

2. **Using wrong documentation type:**
   - User docs should be user-focused, not technical
   - API docs should document interfaces, not implementations
   - Keep types distinct

3. **Not linking documentation to code:**
   - Always add DOC= field to CANARY tokens
   - Missing link defeats the tracking system

4. **Letting documentation go stale:**
   - Check status regularly
   - Update when code changes
   - Don't accumulate stale docs

5. **Creating documentation without context:**
   - Read the code first
   - Understand the feature
   - Provide accurate, helpful content

## See Also

- `/canary.specify` - Create requirement specifications
- `/canary.plan` - Generate implementation plans
- `/canary.verify` - Verify GAP_ANALYSIS.md claims
- `/canary.scan` - Scan for CANARY tokens
