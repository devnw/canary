<!-- CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Feature Specification: Multi-Line CANARY Tokens

**Requirement ID:** CBIN-138
**Status:** STUB
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

## Overview

**Purpose:** Allow CANARY tokens to span multiple lines to improve code readability and prevent excessively long single-line comments that extend far into files. This addresses the problem where tokens with many fields (REQ, FEATURE, ASPECT, STATUS, OWNER, TEST, BENCH, DOC, DOC_HASH, etc.) create lines that are difficult to read and maintain.

**Scope:**
- Included: Multi-line token syntax with continuation indicator
- Included: Parsing logic to recognize and combine multi-line tokens
- Included: Backward compatibility with existing single-line tokens
- Included: Support for all comment styles (// for Go, # for Python/Shell, <!-- --> for HTML/Markdown)
- Included: Validation of multi-line token structure
- Excluded: Automatic reformatting of existing single-line tokens (manual migration)
- Excluded: Mixed single-line and multi-line for same token (must be all single or all multi)
- Excluded: Arbitrary continuation indicators (standardize on one format)

## User Stories

### Primary User Stories

**US-1: Write Readable Multi-Line Tokens**
As a developer,
I want to split long CANARY tokens across multiple lines,
So that my code remains readable and doesn't require horizontal scrolling.

**Acceptance Criteria:**
- [ ] Can write token across 2-5 lines with continuation indicator
- [ ] Continuation lines start with comment prefix + continuation marker
- [ ] All fields are recognized and parsed correctly
- [ ] Token semantics are identical to single-line equivalent

**US-2: Parse Multi-Line Tokens**
As the CANARY scanner,
I want to recognize and parse multi-line tokens,
So that I can extract all fields regardless of line breaks.

**Acceptance Criteria:**
- [ ] Scanner detects continuation indicator on next line
- [ ] Scanner combines lines into complete token before parsing
- [ ] Scanner handles various comment styles (// # <!-- -->)
- [ ] Scanner reports parse errors with line numbers

**US-3: Maintain Backward Compatibility**
As a project maintainer,
I want existing single-line tokens to continue working,
So that I don't need to rewrite all tokens immediately.

**Acceptance Criteria:**
- [ ] Single-line tokens parse exactly as before
- [ ] Mixed single-line and multi-line tokens coexist in same file
- [ ] No breaking changes to token field syntax
- [ ] Existing scans produce same results for single-line tokens

### Secondary User Stories

**US-4: Validate Multi-Line Structure**
As a developer,
I want the scanner to detect malformed multi-line tokens,
So that I get clear error messages when I make mistakes.

**Acceptance Criteria:**
- [ ] Error if continuation line missing continuation marker
- [ ] Error if continuation line appears without initial CANARY line
- [ ] Warning if continuation line is indented inconsistently
- [ ] Helpful error messages with line numbers

**US-5: Format Multi-Line Tokens Consistently**
As a team,
I want multi-line tokens to follow consistent formatting rules,
So that code reviews are easier and diffs are clean.

**Acceptance Criteria:**
- [ ] Documentation provides formatting guidelines
- [ ] Examples show recommended indentation
- [ ] Guidelines specify field grouping (related fields on same line)

## Functional Requirements

### FR-1: Continuation Indicator Syntax
**Priority:** High
**Description:** Define continuation marker that indicates a CANARY token continues on next line. Use `CANARY:>` as the continuation indicator.
**Acceptance:** Scanner recognizes `CANARY:>` at start of continuation lines and combines them with initial token line

### FR-2: Multi-Line Token Parser
**Priority:** High
**Description:** Extend token parser to accumulate lines with continuation indicator before parsing fields
**Acceptance:** Parser correctly extracts all fields from multi-line tokens, producing identical results to single-line equivalent

### FR-3: Comment Style Support
**Priority:** High
**Description:** Support multi-line tokens in all comment styles (// # <!-- -->)
**Acceptance:** Multi-line tokens work in Go (.go), Python (.py), Shell (.sh), Markdown (.md), HTML (.html) files

### FR-4: Backward Compatibility
**Priority:** High
**Description:** Existing single-line tokens continue to parse correctly without modifications
**Acceptance:** Scanner produces identical output for single-line tokens before and after multi-line support

### FR-5: Validation and Error Reporting
**Priority:** Medium
**Description:** Detect and report malformed multi-line tokens with helpful error messages
**Acceptance:** Errors include line numbers, describe problem, suggest fix

### FR-6: Field Grouping Recommendations
**Priority:** Low
**Description:** Document recommended field grouping for multi-line tokens (e.g., REQ+FEATURE on line 1, STATUS+OWNER+UPDATED on line 2)
**Acceptance:** Documentation provides examples and rationale for grouping

## Success Criteria

**Quantitative Metrics:**
- [ ] Multi-line tokens reduce line length by 40-60% for tokens with 5+ fields
- [ ] Parsing performance degradation < 5% for files with multi-line tokens
- [ ] Zero breaking changes for existing single-line tokens
- [ ] Scanner handles files with 100% multi-line tokens in < 1 second

**Qualitative Measures:**
- [ ] Developers find multi-line tokens easier to read
- [ ] Code reviews show fewer comments about token readability
- [ ] Team adopts multi-line format for new tokens with 4+ fields

## User Scenarios & Testing

### Scenario 1: Write Multi-Line Token (Happy Path)
**Given:** Developer has a token with 6 fields that creates a 150-character line
**When:** They split it across 3 lines using `CANARY:>` continuation
**Then:** Scanner parses all fields correctly and produces same output as single-line version

**Example:**
```go
// CANARY: REQ=CBIN-138; FEATURE="MultilineTokens"; ASPECT=Engine;
// CANARY:> STATUS=IMPL; OWNER=alice; TEST=TestMultiline;
// CANARY:> UPDATED=2025-10-16
```

### Scenario 2: Parse Mixed Single and Multi-Line Tokens
**Given:** File contains both single-line and multi-line CANARY tokens
**When:** Scanner processes the file
**Then:** Both token types are parsed correctly and appear in output

### Scenario 3: Detect Malformed Continuation (Error Case)
**Given:** Developer forgets `CANARY:>` marker on continuation line
**When:** Scanner processes the file
**Then:** Scanner reports error: "Line 23: Expected 'CANARY:>' continuation marker, found '// STATUS=IMPL'"

**Example Error:**
```go
// CANARY: REQ=CBIN-138; FEATURE="Test";
// STATUS=IMPL; UPDATED=2025-10-16    // ERROR: Missing CANARY:>
```

### Scenario 4: Python Multi-Line Token
**Given:** Python file with CANARY token using # comments
**When:** Token is split across 2 lines with `# CANARY:>` continuation
**Then:** Scanner parses token correctly

**Example:**
```python
<!-- CANARY: REQ=CBIN-138; FEATURE="PythonToken"; ASPECT=API; -->
<!-- CANARY:> STATUS=IMPL; UPDATED=2025-10-16 -->
```

### Scenario 5: Markdown Multi-Line Token
**Given:** Markdown file with CANARY token using <!-- --> comments
**When:** Token is split across multiple lines
**Then:** Scanner parses token correctly

**Example:**
```markdown
<!-- CANARY: REQ=CBIN-138; FEATURE="MarkdownToken";
     CANARY:> ASPECT=Docs; STATUS=IMPL;
     CANARY:> UPDATED=2025-10-16 -->
```

### Scenario 6: Excessive Line Breaks (Edge Case)
**Given:** Developer splits token across 10 lines (excessive)
**When:** Scanner processes the token
**Then:** Scanner parses correctly but may warn about excessive continuation (optional)

### Scenario 7: Backward Compatibility Test
**Given:** Project has 500 existing single-line CANARY tokens
**When:** Multi-line support is deployed
**Then:** All 500 tokens parse identically to before deployment

## Key Entities

### Entity 1: MultiLineToken
**Attributes:**
- initial_line: First line containing `CANARY:` prefix
- continuation_lines: List of lines starting with `CANARY:>`
- comment_style: Comment prefix used (// # <!-- -->)
- combined_text: Full token text after joining lines
- start_line_num: Line number where token starts
- end_line_num: Line number where token ends

**Relationships:**
- Parsed into Token entity (existing)
- Multiple MultiLineToken instances in a file

### Entity 2: ContinuationMarker
**Attributes:**
- marker_text: "CANARY:>" string
- comment_prefix: Comment style prefix (e.g., "//")
- full_pattern: Complete pattern to match (e.g., "// CANARY:>")

**Relationships:**
- Used to identify continuation lines
- Varies by file type / comment style

## Assumptions

- CANARY tokens always start with `CANARY:` prefix (no change)
- Continuation lines immediately follow previous line (no blank lines between)
- Comment style is consistent within a multi-line token
- Field syntax remains unchanged (key-value pairs with semicolons)
- Maximum 10 continuation lines per token (reasonable limit)
- Indentation is allowed but not required for continuation lines

## Constraints

**Technical Constraints:**
- Must maintain backward compatibility with single-line tokens
- Continuation marker must be unique enough to avoid false positives
- Parser performance should not degrade significantly
- Must work with all existing comment styles

**Business Constraints:**
- No breaking changes to existing tools or workflows
- Migration from single-line to multi-line is optional
- Documentation must be updated to explain new syntax

**Regulatory Constraints:**
- None

## Out of Scope

- **Automatic Reformatting:** No tool to automatically convert single-line to multi-line tokens (manual migration)
- **Custom Continuation Markers:** Only `CANARY:>` is supported (no configurable markers)
- **Mixed Style Per Token:** A single token must be entirely single-line OR entirely multi-line (no mixing)
- **Continuation Across Blank Lines:** Continuation lines must immediately follow previous line
- **Arbitrary Field Order:** Field order requirements remain unchanged
- **Pretty-Printing Tool:** No formatter to reflow tokens across lines
- **IDE Integration:** No syntax highlighting or auto-completion (future enhancement)

## Dependencies

- Existing CANARY token parser (internal/matcher/token.go)
- Comment detection logic
- File reading infrastructure

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| False positive continuation detection | Medium | Low | Use unique marker `CANARY:>` that's unlikely to appear naturally |
| Performance degradation on large files | Medium | Low | Optimize line buffering, early termination on non-CANARY lines |
| Breaking existing tokens | High | Low | Extensive backward compatibility tests, no changes to single-line syntax |
| Developer confusion about syntax | Medium | Medium | Clear documentation with many examples, error messages with suggestions |
| Inconsistent formatting across team | Low | Medium | Provide guidelines and examples, optional lint rules |

## Clarifications Needed

[NEEDS CLARIFICATION: Should continuation lines require specific indentation?]
**Options:**
A) Require continuation lines to align with initial line (strict indentation)
B) Allow any indentation as long as marker is present (flexible)
C) Recommend alignment but don't enforce (best practice guideline)
**Impact:** Option C provides best balance - keeps parser simple while encouraging readability

[NEEDS CLARIFICATION: Should there be a maximum number of continuation lines?]
**Options:**
A) No limit - allow any number of continuation lines
B) Hard limit of 10 continuation lines
C) Warning at 5 lines, error at 10 lines
**Impact:** Option C helps catch mistakes (missing semicolon causing runaway continuation) while allowing flexibility

[NEEDS CLARIFICATION: Should continuation work across comment blocks in HTML/Markdown?]
**Options:**
A) Continuation must stay within single comment block (<!-- ... -->)
B) Allow continuation across adjacent comment blocks
C) Support both but prefer single block
**Impact:** Option A is simplest and most predictable, avoids edge cases

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] Only 3 [NEEDS CLARIFICATION] markers remaining
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable and technology-agnostic
- [x] All acceptance scenarios defined
- [x] Edge cases identified
- [x] Scope clearly bounded
- [x] Dependencies and assumptions identified

**Readiness:**
- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Ready for technical planning (`/canary.plan`)

---

## Implementation Checklist

### Core Features

<!-- CANARY: REQ=CBIN-138; FEATURE="ContinuationParser"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 1: Continuation Line Parser**
- [ ] Detect `CANARY:>` continuation marker
- [ ] Accumulate continuation lines into buffer
- [ ] Combine lines removing comment prefixes and markers
- [ ] Pass combined text to existing field parser
- **Location hint:** `internal/matcher/token.go` (extend existing parser)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-138; FEATURE="CommentStyleDetection"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 2: Comment Style Detection**
- [ ] Identify comment prefix from first line (// # <!--)
- [ ] Match same prefix on continuation lines
- [ ] Handle edge case: <!-- --> blocks
- [ ] Validate consistency across continuation lines
- **Location hint:** `internal/matcher/token.go`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-138; FEATURE="BackwardCompatibility"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 3: Backward Compatibility Layer**
- [ ] Single-line tokens parse unchanged
- [ ] No performance regression for single-line tokens
- [ ] Mixed single/multi-line files supported
- [ ] Existing tests continue to pass
- **Location hint:** `internal/matcher/token.go`
- **Dependencies:** ContinuationParser

<!-- CANARY: REQ=CBIN-138; FEATURE="ErrorReporting"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 4: Error Reporting**
- [ ] Detect malformed continuation lines
- [ ] Report line numbers in error messages
- [ ] Suggest fixes for common mistakes
- [ ] Warn on excessive continuation (>5 lines)
- **Location hint:** `internal/matcher/token.go`
- **Dependencies:** ContinuationParser

### Testing Requirements

<!-- CANARY: REQ=CBIN-138; FEATURE="UnitTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN138_MultilineParser; UPDATED=2025-10-16 -->
**Unit Tests:**
- [ ] Test 2-line, 3-line, 5-line tokens
- [ ] Test all comment styles (// # <!--)
- [ ] Test malformed continuation lines
- [ ] Test backward compatibility (single-line tokens)
- [ ] Test mixed single/multi-line files
- **Location hint:** `internal/matcher/token_test.go`

<!-- CANARY: REQ=CBIN-138; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN138_Integration; UPDATED=2025-10-16 -->
**Integration Tests:**
- [ ] Test real Go files with multi-line tokens
- [ ] Test real Python files with multi-line tokens
- [ ] Test markdown specs with multi-line tokens
- [ ] Test large files (100+ tokens) for performance
- **Location hint:** `internal/matcher/token_integration_test.go`

### Documentation

<!-- CANARY: REQ=CBIN-138; FEATURE="SyntaxDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**Syntax Documentation:**
- [ ] Document `CANARY:>` continuation syntax
- [ ] Provide examples for all comment styles
- [ ] Show field grouping recommendations
- [ ] Explain when to use multi-line vs single-line
- **Location hint:** `README.md` or `.canary/docs/multiline-tokens.md`

<!-- CANARY: REQ=CBIN-138; FEATURE="MigrationGuide"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**Migration Guide:**
- [ ] Explain backward compatibility
- [ ] Show before/after examples
- [ ] Recommend which tokens to migrate first
- [ ] Provide manual conversion steps
- **Location hint:** `.canary/docs/migration-multiline.md`

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-138` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-138; FEATURE="MultilineTokens"; ASPECT=Engine;
// CANARY:> STATUS=IMPL; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-138; FEATURE="ContinuationParser"; ASPECT=Engine;
// CANARY:> STATUS=IMPL; TEST=TestContinuationParser;
// CANARY:> UPDATED=2025-10-16
```

**Use `canary implement CBIN-138` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
