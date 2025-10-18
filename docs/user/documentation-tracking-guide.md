# Documentation Tracking User Guide

**Feature:** CBIN-136 - Documentation Tracking and Consistency
**Status:** Production Ready
**Last Updated:** 2025-10-16

## Overview

The CANARY documentation tracking system helps you keep documentation synchronized with code. It uses SHA256 hashing to automatically detect when documentation becomes stale, provides coverage metrics, and enables batch operations for maintaining large documentation sets.

## Quick Start

### 1. Create Documentation for a Requirement

When you have a new requirement (e.g., CBIN-105), create its documentation:

```bash
canary doc create CBIN-105 --type user --output docs/user/authentication.md
```

This will:
- Create `docs/user/authentication.md` from a template
- Calculate the initial SHA256 hash
- Show you how to add the DOC field to your CANARY token

**Output:**
```
✅ Created documentation: docs/user/authentication.md
   Requirement: CBIN-105
   Type: user
   Hash: 8f434346648f6b96

Next steps:
  1. Edit the documentation file: docs/user/authentication.md
  2. Add DOC= field to your CANARY token:
     DOC=user:docs/user/authentication.md; DOC_HASH=8f434346648f6b96
  3. After editing, run: canary doc update CBIN-105
```

### 2. Add DOC Field to CANARY Token

In your source code, add the DOC and DOC_HASH fields:

```go
// CANARY: REQ=CBIN-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL;
//         DOC=user:docs/user/authentication.md;
//         DOC_HASH=8f434346648f6b96;
//         UPDATED=2025-10-16
func Authenticate(username, password string) (*User, error) {
    // Implementation...
}
```

### 3. Update Documentation

After editing the documentation file:

```bash
canary doc update CBIN-105
```

**Output:**
```
✅ Updated: docs/user/authentication.md (hash: a1b2c3d4e5f6g7h8)

✅ Updated 1 documentation file(s)
```

### 4. Check Documentation Status

Verify all documentation is current:

```bash
canary doc status --all
```

**Output:**
```
✅ CBIN-105 (DOC_CURRENT): user:docs/user/authentication.md
⚠️  CBIN-200 (DOC_STALE): api:docs/api/rest.md
❌ CBIN-300 (DOC_MISSING): technical:docs/tech/design.md

Summary: 3 total
  ✅ Current:  1
  ⚠️  Stale:    1
  ❌ Missing:  1
```

### 5. Generate Coverage Report

Get project-wide documentation metrics:

```bash
canary doc report
```

**Output:**
```
📊 Documentation Report

Coverage: 6/125 requirements (4.8%)
Total Tokens: 971 (9 with docs, 119 without)

📋 Documentation Status:
  ✅ Current:  5 (55.6%)
  ⚠️  Stale:    3 (33.3%)
  ❌ Missing:  1 (11.1%)

💡 119 requirements without documentation (use --show-undocumented to list)

💡 Recommendations:
  Run 'canary doc update --all --stale-only' to update stale documentation
```

## Documentation Types

The system supports five documentation types, each serving a different purpose:

### User Documentation (`user`)

**Purpose:** Help end-users understand and use features

**When to Create:**
- After implementing user-facing features
- When writing tutorials or how-to guides
- For troubleshooting guides

**Example:**
```bash
canary doc create CBIN-105 --type user --output docs/user/authentication.md
```

**Template Includes:**
- Overview
- Usage instructions
- Examples
- Common issues

### API Documentation (`api`)

**Purpose:** Document function signatures and API contracts

**When to Create:**
- When creating public APIs
- For library interfaces
- When defining REST/GraphQL endpoints

**Example:**
```bash
canary doc create CBIN-200 --type api --output docs/api/auth-endpoints.md
```

**Template Includes:**
- Function signatures
- Parameter descriptions
- Return values
- Code examples

### Technical Documentation (`technical`)

**Purpose:** Explain implementation details and architecture

**When to Create:**
- During technical design phase
- For complex algorithms
- When documenting system internals

**Example:**
```bash
canary doc create CBIN-300 --type technical --output docs/technical/auth-flow.md
```

**Template Includes:**
- Architecture overview
- Implementation details
- Performance considerations
- Technical constraints

### Feature Documentation (`feature`)

**Purpose:** Capture feature specifications and requirements

**When to Create:**
- During requirement specification
- For user stories
- When defining acceptance criteria

**Example:**
```bash
canary doc create CBIN-400 --type feature --output docs/features/oauth2-support.md
```

**Template Includes:**
- Feature description
- User stories
- Acceptance criteria
- Functional requirements

### Architecture Documentation (`architecture`)

**Purpose:** Record architecture decisions (ADRs)

**When to Create:**
- When making significant architectural decisions
- For technology choices
- When establishing patterns/conventions

**Example:**
```bash
canary doc create CBIN-500 --type architecture --output docs/architecture/adr-002-auth-provider.md
```

**Template Includes:**
- Context
- Decision
- Alternatives considered
- Consequences

## Common Workflows

### Workflow 1: New Feature Development

1. **Specify Requirement:**
   ```bash
   /canary.specify "Add OAuth2 authentication support"
   # Creates CBIN-XXX
   ```

2. **Create Feature Documentation:**
   ```bash
   canary doc create CBIN-XXX --type feature --output docs/features/oauth2.md
   ```

3. **Edit Feature Docs:**
   - Open `docs/features/oauth2.md`
   - Fill in user stories, acceptance criteria

4. **Plan Implementation:**
   ```bash
   /canary.plan CBIN-XXX
   ```

5. **Create Technical Documentation:**
   ```bash
   canary doc create CBIN-XXX --type technical --output docs/technical/oauth2-impl.md
   ```

6. **Implement Feature:**
   - Write code with CANARY tokens including DOC fields

7. **Update Hashes:**
   ```bash
   canary doc update CBIN-XXX
   ```

8. **Verify Status:**
   ```bash
   canary doc status CBIN-XXX
   ```

### Workflow 2: Updating Stale Documentation

1. **Check for Stale Docs:**
   ```bash
   canary doc status --all --stale-only
   ```

2. **Review Each Stale Document:**
   ```bash
   # For each stale doc, open and review changes
   vim docs/user/authentication.md
   ```

3. **Update Documentation:**
   - Make necessary changes to reflect current implementation

4. **Update Hashes:**
   ```bash
   canary doc update --all --stale-only
   ```

5. **Verify All Current:**
   ```bash
   canary doc status --all
   ```

### Workflow 3: Batch Documentation Update

When you've edited multiple documentation files:

1. **Update All Documentation:**
   ```bash
   canary doc update --all
   ```

2. **Review Results:**
   ```
   ✅ CBIN-105: user:docs/user/auth.md (hash: 8f434346)
   ✅ CBIN-200: api:docs/api/rest.md (hash: a1b2c3d4)
   ✅ CBIN-300: technical:docs/tech/design.md (hash: e5f6g7h8)

   ✅ Updated 3 requirement(s)
   ```

### Workflow 4: Documentation Coverage Audit

1. **Generate Report:**
   ```bash
   canary doc report --show-undocumented
   ```

2. **Review Undocumented Requirements:**
   ```
   📝 Undocumented Requirements (119):
     - CBIN-001
     - CBIN-002
     - ...
   ```

3. **Create Missing Documentation:**
   ```bash
   # For each undocumented requirement
   canary doc create CBIN-001 --type user --output docs/user/feature-001.md
   ```

4. **Update GAP_ANALYSIS.md:**
   - Add documentation tracking to claimed requirements

5. **Verify Improvement:**
   ```bash
   canary doc report
   # Check coverage percentage increase
   ```

## Multiple Documentation Files

A single requirement can have multiple documentation files of different types:

### Example: Authentication Feature

```go
// CANARY: REQ=CBIN-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL;
//         DOC=user:docs/user/auth.md,api:docs/api/auth-api.md,technical:docs/tech/auth-design.md;
//         DOC_HASH=8f434346,a1b2c3d4,e5f6g7h8;
//         UPDATED=2025-10-16
```

**Benefits:**
- User docs for end-users
- API docs for developers
- Technical docs for maintainers
- Each tracked independently

**Checking Status:**
```bash
canary doc status CBIN-105
```

**Output:**
```
✅ CBIN-105 (DOC_CURRENT): docs/user/auth.md
✅ CBIN-105 (DOC_CURRENT): docs/api/auth-api.md
⚠️  CBIN-105 (DOC_STALE): docs/tech/auth-design.md

Summary: 3 total
  ✅ Current:  2
  ⚠️  Stale:    1
```

## Batch Operations

### Update All Documentation

Update hashes for all documentation in the database:

```bash
canary doc update --all
```

**Use When:**
- After restructuring documentation
- When verifying all docs are current
- During release preparation

### Update Only Stale Documentation

Selectively update only documentation that has changed:

```bash
canary doc update --all --stale-only
```

**Use When:**
- Updating only modified docs
- Skipping already-current documentation
- Optimizing update time for large projects

**Output:**
```
✅ CBIN-200: api:docs/api/rest.md (hash: new123)
✅ CBIN-300: technical:docs/tech/design.md (hash: new456)

✅ Updated 2 requirement(s) (skipped 15 current)
```

## Status Values

### DOC_CURRENT ✅

**Meaning:** Documentation hash matches file content

**Action Required:** None - documentation is up to date

**Example:**
```
✅ CBIN-105 (DOC_CURRENT): user:docs/user/auth.md
```

### DOC_STALE ⚠️

**Meaning:** Documentation has been modified since last hash update

**Action Required:**
1. Review what changed in the documentation
2. Ensure changes are intentional
3. Run `canary doc update CBIN-XXX` to update hash

**Example:**
```
⚠️  CBIN-200 (DOC_STALE): api:docs/api/rest.md
```

### DOC_MISSING ❌

**Meaning:** Documentation file does not exist at specified path

**Action Required:**
1. Check if file was moved or deleted
2. Either create the documentation or remove DOC field from token
3. Update hash if documentation is recreated

**Example:**
```
❌ CBIN-300 (DOC_MISSING): technical:docs/tech/design.md
```

### DOC_UNHASHED ℹ️

**Meaning:** Documentation path specified but no hash stored

**Action Required:**
1. Run `canary doc update CBIN-XXX` to calculate and store hash
2. Or add DOC_HASH field manually

**Example:**
```
ℹ️  CBIN-400 (DOC_UNHASHED): feature:docs/features/oauth2.md
```

## JSON Output for Automation

All commands support JSON output for scripting and CI/CD integration:

### Report in JSON

```bash
canary doc report --format json
```

**Output:**
```json
{
  "total_tokens": 971,
  "tokens_with_docs": 9,
  "tokens_without_docs": 119,
  "coverage_percent": 4.8,
  "by_type": {
    "user": 3,
    "api": 2,
    "technical": 4
  },
  "by_status": {
    "DOC_CURRENT": 5,
    "DOC_STALE": 3,
    "DOC_MISSING": 1
  },
  "undocumented_count": 119
}
```

**Use Cases:**
- CI/CD pipeline checks (fail if coverage < threshold)
- Dashboard visualizations
- Automated reporting
- Metrics tracking over time

### Example: CI/CD Integration

```bash
#!/bin/bash
# Fail if documentation coverage < 80%

REPORT=$(canary doc report --format json)
COVERAGE=$(echo $REPORT | jq '.coverage_percent')

if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "❌ Documentation coverage too low: $COVERAGE%"
    exit 1
fi

echo "✅ Documentation coverage acceptable: $COVERAGE%"
```

## Best Practices

### 1. Create Documentation Early

**Do:**
- Create feature docs during specification phase
- Create technical docs during planning
- Create API docs before implementation starts

**Don't:**
- Wait until after implementation to document
- Skip documentation for "simple" features
- Delay documentation until release time

### 2. Keep Documentation Current

**Do:**
- Update docs when code changes
- Run `canary doc status --all` regularly
- Fix stale documentation immediately
- Include doc updates in code reviews

**Don't:**
- Let documentation drift
- Ignore staleness warnings
- Update code without updating docs
- Assume docs stay current automatically

### 3. Use Appropriate Types

**Do:**
- User docs for end-user features
- API docs for developer-facing interfaces
- Technical docs for implementation details
- Architecture docs for design decisions

**Don't:**
- Mix different documentation types
- Use vague or generic types
- Create docs without clear purpose
- Over-document internal details in user docs

### 4. Link Documentation to Code

**Do:**
- Always include DOC= in CANARY tokens
- Reference multiple docs when appropriate
- Keep paths relative to project root
- Use type prefixes (user:, api:, technical:)

**Don't:**
- Create orphaned documentation
- Use absolute file paths
- Forget DOC_HASH field
- Mix documentation types in single file

### 5. Verify Before Release

**Do:**
- Check documentation status before commits
- Include documentation in code review
- Verify all documentation is DOC_CURRENT
- Run coverage report before releases

**Don't:**
- Release with stale documentation
- Skip documentation in reviews
- Ignore missing documentation
- Assume documentation is current

## Troubleshooting

### Problem: Documentation shows as DOC_STALE but I didn't change it

**Possible Causes:**
- Line ending differences (CRLF vs LF)
- Git autocrlf settings changed line endings
- Whitespace modifications

**Solution:**
```bash
# Check line endings
file docs/user/auth.md

# If mixed, normalize:
dos2unix docs/user/auth.md

# Update hash:
canary doc update CBIN-105
```

### Problem: Can't find documentation template

**Error:**
```
failed to read template: open .canary/templates/docs/user-template.md: no such file or directory
```

**Solution:**
```bash
# Regenerate templates
canary init

# Or create custom template manually
mkdir -p .canary/templates/docs
vim .canary/templates/docs/user-template.md
```

### Problem: Database doesn't have doc fields

**Error:**
```
SQL logic error: table tokens has no column named doc_path (1)
```

**Solution:**
```bash
# Run database migration
canary migrate all

# Or rebuild database
rm .canary/canary.db
canary index --root .
```

### Problem: Hash mismatch after normalizing line endings

**Scenario:** Updated documentation but hash still shows stale

**Solution:**
```bash
# Ensure consistent line endings
git config core.autocrlf input

# Normalize existing files
find docs -type f -name "*.md" -exec dos2unix {} \;

# Update all hashes
canary doc update --all
```

## Integration with Other Commands

### After /canary.specify

```bash
/canary.specify "Add user authentication"
# Creates requirement CBIN-XXX

# Create feature documentation
canary doc create CBIN-XXX --type feature --output docs/features/auth.md
```

### After /canary.plan

```bash
/canary.plan CBIN-XXX
# Creates implementation plan

# Create technical documentation
canary doc create CBIN-XXX --type technical --output docs/technical/auth-impl.md
```

### Before /canary.verify

```bash
# Check all documentation is current
canary doc status --all

# Fix any stale documentation
canary doc update --all --stale-only

# Verify GAP_ANALYSIS claims
/canary.verify
```

## Advanced Usage

### Custom Templates

Create custom documentation templates:

```bash
# Create template directory
mkdir -p .canary/templates/docs

# Create custom user template
cat > .canary/templates/docs/user-template.md <<'EOF'
# {{.Feature}} User Guide

**Requirement:** {{.ReqID}}
**Version:** 1.0
**Last Updated:** {{.Date}}

## Quick Start

TODO: Provide quick start instructions

## Features

TODO: List key features

## Usage

TODO: Detailed usage instructions

## Examples

TODO: Code examples

## FAQ

TODO: Common questions
EOF
```

### Documentation Metrics Over Time

Track documentation coverage trends:

```bash
# Generate daily report
canary doc report --format json > metrics/doc-coverage-$(date +%Y-%m-%d).json

# Compare over time with jq
jq -s '.[0].coverage_percent - .[1].coverage_percent' \
    metrics/doc-coverage-2025-10-16.json \
    metrics/doc-coverage-2025-10-15.json
```

### Pre-Commit Hook

Prevent commits with stale documentation:

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check documentation status
STALE=$(canary doc status --all --stale-only 2>/dev/null | grep "DOC_STALE" | wc -l)

if [ "$STALE" -gt 0 ]; then
    echo "❌ Cannot commit: $STALE stale documentation file(s)"
    echo "Run: canary doc update --all --stale-only"
    exit 1
fi

echo "✅ All documentation current"
exit 0
```

## Summary

The CANARY documentation tracking system provides:

✅ **Automated staleness detection** using SHA256 hashing
✅ **Coverage metrics** to identify documentation gaps
✅ **Batch operations** for efficient bulk updates
✅ **Type categorization** for different documentation purposes
✅ **JSON output** for CI/CD integration
✅ **Multiple docs per requirement** for comprehensive coverage

Use it to keep your documentation synchronized with code and prevent documentation drift.
