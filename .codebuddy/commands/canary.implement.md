# /canary.implement - Generate Implementation Guidance

**Purpose:** Generate comprehensive implementation guidance for a CANARY requirement specification.

**Usage:**
```
/canary.implement <query>
/canary.implement --list
```

---

## Command Behavior

This command generates a complete implementation prompt that includes:

1. **Specification Details** - Full requirement specification
2. **Implementation Plan** - Technical plan with architecture and phases
3. **Implementation Checklist** - Step-by-step feature checklist
4. **Progress Tracking** - Current status of all tokens
5. **Constitutional Principles** - Project governance and best practices
6. **Test-First Guidance** - TDD workflow instructions
7. **CANARY Token Examples** - Proper token placement and format

---

## Query Modes

### Exact ID Match
```
/canary.implement CBIN-105
```
Finds specification by exact requirement ID.

### Fuzzy Search
```
/canary.implement "user auth"
/canary.implement authentication
```
Uses fuzzy matching with:
- Levenshtein distance for typo tolerance
- Substring matching
- Abbreviation matching (e.g., "ua" → "UserAuthentication")

**Auto-selection:** If top match scores >80% and is >20 points ahead of the second match, automatically selects it.

**Interactive selection:** For ambiguous matches, displays numbered list for user to select.

### Feature Name Match
```
/canary.implement UserAuthentication
```
Searches by feature name in spec directory names.

### List All Unimplemented
```
/canary.implement --list
```
Shows all requirements with incomplete implementation (STUB or IMPL status).

---

## Implementation Workflow

After running `/canary.implement <query>`:

### 1. Review Generated Prompt
The prompt contains everything you need:
- Complete specification
- Technical plan
- Implementation checklist
- Constitutional principles

### 2. Follow Test-First Approach (Article IV)
```
1. Create test file (*_test.go or equivalent)
2. Write tests (they should fail - TDD Red)
3. Implement features to pass tests (TDD Green)
4. Refactor while keeping tests green
```

### 3. Add CANARY Tokens
Place tokens above each implementation point:
```go
// CANARY: REQ=CBIN-XXX; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
func ImplementFeature() {
    // implementation
}
```

### 4. Update Token Status as You Progress
- `STATUS=STUB` - Placeholder, not yet implemented
- `STATUS=IMPL` - Implemented but needs tests
- `STATUS=TESTED; TEST=TestName` - Implemented with passing tests
- `STATUS=BENCHED; BENCH=BenchName` - Tested with benchmarks

### 5. Verify Progress
```bash
canary scan --root .
canary implement <REQ-ID>  # Check progress
```

---

## Examples

### Example 1: Implement by Exact ID
```
User: /canary.implement CBIN-105

Agent Response:
[Generates full implementation prompt with:
 - Spec for CBIN-105
 - Plan details
 - Checklist
 - Progress (0/8 features completed)
 - Constitutional principles
 - TDD instructions]

Agent: I'll implement this following the Test-First approach...
[Creates test files first, then implementation]
```

### Example 2: Fuzzy Search with Auto-Selection
```
User: /canary.implement "user auth"

Agent: Found strong match (score: 92%): CBIN-105 - UserAuthentication
[Auto-selects and generates prompt]
```

### Example 3: Ambiguous Search (Interactive Selection)
```
User: /canary.implement "auth"

Agent: Multiple matches found:
1. CBIN-105 - UserAuthentication (Score: 85%)
2. CBIN-110 - OAuthIntegration (Score: 80%)
3. CBIN-115 - AuthorizationRules (Score: 75%)

Select a requirement (1-3):
User: 1

[Generates prompt for CBIN-105]
```

### Example 4: List All Unimplemented
```
User: /canary.implement --list

Agent Response:
Requirements with incomplete implementation:

CBIN-105 - UserAuthentication [25% complete]
  Total: 8 | Tested: 2 | Impl: 3 | Stub: 3

CBIN-110 - OAuthIntegration [0% complete]
  Total: 5 | Tested: 0 | Impl: 0 | Stub: 5

CBIN-112 - DataValidation [60% complete]
  Total: 10 | Tested: 6 | Impl: 2 | Stub: 2
```

---

## When to Use This Command

✅ **Use when:**
- Starting implementation of a new requirement
- Resuming work on an incomplete requirement
- Need full context on what to implement
- Want to see progress on a requirement
- Need reminder of test-first workflow

❌ **Don't use when:**
- Just want to see token locations (use `canary implement <REQ-ID>` CLI instead)
- Scanning for all tokens (use `/canary.scan`)
- Creating new specifications (use `/canary.specify`)
- Planning architecture (use `/canary.plan`)

---

## Integration with Other Commands

### Workflow Sequence
```
1. /canary.specify "feature description"
   → Creates spec.md

2. /canary.plan CBIN-XXX
   → Creates plan.md

3. /canary.implement CBIN-XXX
   → Generates implementation prompt
   → Agent implements following TDD

4. /canary.scan
   → Verifies token status

5. /canary.verify
   → Confirms completion claims
```

---

## CLI Equivalent

This slash command wraps the CLI command:
```bash
canary implement <query>
canary implement --list
```

The slash command is designed for AI agents to automatically generate implementation prompts during development sessions.

---

## Technical Details

### Fuzzy Matching Algorithm
- **Levenshtein Distance** - Calculates edit distance between strings
- **Substring Matching** - Detects partial matches
- **Abbreviation Matching** - Matches first letters of words
- **Scoring:** 0-100 scale (100 = exact match)
- **Threshold:** Minimum 60% similarity to show in results
- **Auto-select:** >80% score AND >20 points ahead of second match

### Template Engine
Uses Go `text/template` for prompt generation. Template variables:
- `{{.ReqID}}` - Requirement ID
- `{{.FeatureName}}` - Feature name
- `{{.SpecContent}}` - Full specification markdown
- `{{.PlanContent}}` - Full plan markdown
- `{{.Checklist}}` - Extracted implementation checklist
- `{{.Progress.Total}}` - Total token count
- `{{.Progress.Tested}}` - Tested token count
- `{{.Progress.Impl}}` - Implemented (not tested) count
- `{{.Progress.Stub}}` - Stub token count
- `{{.Constitution}}` - Constitutional principles markdown

### Directory Structure
```
.canary/
├── specs/
│   └── CBIN-XXX-FeatureName/
│       ├── spec.md              # Required
│       └── plan.md              # Optional (included if exists)
├── templates/
│   └── implement-prompt-template.md  # Prompt template
└── memory/
    └── constitution.md          # Optional (included if exists)
```

---

## Performance

- **Lookup:** <1 second (fuzzy search across <1000 specs)
- **Prompt Generation:** <2 seconds (template rendering)
- **Context Size:** Varies by spec (typically 2-10KB)

---

## Error Handling

### No Match Found
```
Error: no matches found for query: "xyz"

Make sure:
1. Specification exists in .canary/specs/
2. Query matches requirement ID or feature name
3. Directory format: CBIN-XXX-FeatureName
```

### Missing Template
```
Error: failed to read template: implement-prompt-template.md not found

Fix: Run `canary init` to restore missing templates
```

### Missing Spec File
```
Error: failed to read spec.md

Fix: Create spec.md in the requirement directory or use /canary.specify
```

---

## Constitutional Compliance

This command enforces:
- **Article I** - Requirement-First (requires spec before implementation)
- **Article IV** - Test-First (includes TDD instructions in prompt)
- **Article VII** - Documentation Currency (requires UPDATED field)

---

**Ready to implement? Run `/canary.implement <query>` and follow the generated guidance!**
