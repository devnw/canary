<!-- CANARY: REQ=CBIN-141; FEATURE="EmbeddedPromptFlag"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
# Feature Specification: Embedded Prompt Flag

**Requirement ID:** CBIN-141
**Status:** STUB
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

## Overview

**Purpose:** Enable AI agents to programmatically retrieve the system prompts for slash commands directly from the CLI binary, eliminating dependency on template files in the project directory. This allows agents to understand slash command behavior even when template files are missing, moved, or in projects that haven't been fully initialized.

**Scope:** This feature adds a `--prompt` flag to each CLI subcommand that corresponds to a slash command (e.g., `/canary.plan`, `/canary.specify`). When invoked, the flag outputs the embedded system prompt that defines how the slash command should be executed. The prompts are compiled into the binary, not read from `.canary/templates/commands/`. Excludes modification of slash command execution logic or prompt content.

## User Stories

### Primary User Stories

**US-1: Retrieve Slash Command Instructions**
As an AI agent,
I want to retrieve the system prompt for a slash command using a CLI flag,
So that I can understand how to execute the command correctly without accessing template files.

**Acceptance Criteria:**
- [ ] Agent can run `canary specify --prompt` and receive the full system prompt for `/canary.specify`
- [ ] Prompt output is identical to what would be expanded when running the slash command
- [ ] Command executes in under 1 second
- [ ] Works even if `.canary/templates/commands/` directory doesn't exist

**US-2: Discover Available Slash Commands**
As an AI agent,
I want to list all available slash commands and their purposes,
So that I can determine which command to use for a given task.

**Acceptance Criteria:**
- [ ] Agent can run `canary --list-slash-commands` to see all available slash commands
- [ ] Output includes command name, brief description, and whether `--prompt` is available
- [ ] Output is formatted for both human reading and machine parsing

**US-3: Verify Prompt Consistency**
As a project maintainer,
I want to ensure embedded prompts match template files,
So that I can detect drift between compiled prompts and project templates.

**Acceptance Criteria:**
- [ ] Agent can compare embedded prompt with template file
- [ ] System warns if embedded prompt differs from template
- [ ] Comparison is byte-for-byte accurate

### Secondary User Stories (if applicable)

**US-4: Debug Slash Command Behavior**
As a developer,
I want to view the exact prompt that will be sent to the AI agent,
So that I can debug unexpected slash command behavior or understand why a command produces certain outputs.

**Acceptance Criteria:**
- [ ] `--prompt` flag works for all slash commands
- [ ] Output includes metadata (version, last updated, command name)
- [ ] Can be combined with `--format json` for structured output

## Functional Requirements

### FR-1: Embed System Prompts in Binary
**Priority:** High
**Description:** System must compile slash command system prompts into the CLI binary at build time, making them available without filesystem access to template files.
**Acceptance:** Binary contains all slash command prompts; prompts are retrievable via CLI flags; binary size increase is less than 100KB.

### FR-2: Prompt Retrieval Flag
**Priority:** High
**Description:** Each CLI subcommand corresponding to a slash command must support a `--prompt` flag that outputs the embedded system prompt and exits without executing the command.
**Acceptance:** Running `canary <subcommand> --prompt` outputs the prompt text to stdout; command exits with code 0; no side effects occur.

### FR-3: Prompt Format Options
**Priority:** Medium
**Description:** System must support outputting prompts in multiple formats (plain text, JSON, markdown) for different consumption scenarios.
**Acceptance:** `--format` flag accepts `text`, `json`, `markdown`; JSON includes metadata fields; default is plain text.

### FR-4: Slash Command Discovery
**Priority:** Medium
**Description:** System must provide a mechanism to list all available slash commands with their embedded prompts.
**Acceptance:** `canary --list-slash-commands` outputs all commands; includes brief description for each; indicates which commands have `--prompt` support.

### FR-5: Prompt Versioning
**Priority:** Low
**Description:** System must track version information for embedded prompts to detect when prompts have been updated across CLI versions.
**Acceptance:** Embedded prompts include version hash or timestamp; version is visible in `--prompt` output; can be queried separately.

### FR-6: Backwards Compatibility
**Priority:** High
**Description:** Adding `--prompt` flag must not break existing CLI usage or interfere with normal command execution.
**Acceptance:** All existing commands work identically; `--prompt` is recognized and handled before command execution; no conflicts with existing flags.

## Success Criteria

**Quantitative Metrics:**
- [ ] `--prompt` flag returns results in under 1 second for all commands
- [ ] Binary size increase is less than 100KB with all prompts embedded
- [ ] 100% of slash commands have corresponding `--prompt` support
- [ ] Zero breaking changes to existing CLI behavior

**Qualitative Measures:**
- [ ] AI agents can successfully retrieve and use prompts without template file access
- [ ] Prompt output is complete and immediately usable by agents
- [ ] Embedded prompts remain synchronized with official prompt content
- [ ] Feature is discoverable through `--help` documentation

## User Scenarios & Testing

### Scenario 1: Agent Retrieves Plan Prompt
**Given:** Agent needs to understand how to execute `/canary.plan` command
**When:** Agent runs `canary plan --prompt`
**Then:** System outputs the complete system prompt for the plan command, including all instructions, formatting requirements, and expected outputs; agent can use this to understand command behavior

### Scenario 2: Project Without Template Files
**Given:** Project was initialized with older CANARY version without template files
**When:** Agent runs `canary specify --prompt` in this project
**Then:** System outputs embedded prompt from binary; agent successfully uses prompt to execute specification task; no errors about missing template files

### Scenario 3: Prompt Comparison for Drift Detection
**Given:** Project has customized `.canary/templates/commands/plan.md` template
**When:** Maintainer runs `canary plan --prompt --compare`
**Then:** System compares embedded prompt with template file; outputs differences if any; warns that agent behavior may differ from expectations

### Scenario 4: Structured Prompt Retrieval
**Given:** Agent needs to parse prompt programmatically
**When:** Agent runs `canary verify --prompt --format json`
**Then:** System outputs JSON with fields: `command_name`, `prompt_text`, `version`, `last_updated`, `description`; JSON is valid and parseable

### Scenario 5: Discovery of All Slash Commands
**Given:** Agent is new to a CANARY project
**When:** Agent runs `canary --list-slash-commands`
**Then:** System outputs table of all slash commands with columns: name, description, has_prompt; agent can identify which commands are available for use

## Key Entities (if data-driven feature)

### Entity 1: EmbeddedPrompt
**Attributes:**
- command_name: Name of the slash command (e.g., "canary.plan")
- subcommand_name: CLI subcommand name (e.g., "plan")
- prompt_text: Full system prompt content
- version: Hash or version identifier
- last_updated: Timestamp of prompt content
- description: Brief description of command purpose

**Relationships:**
- One-to-one mapping between slash commands and embedded prompts
- Links to CLI subcommand registration

### Entity 2: PromptMetadata
**Attributes:**
- format: Output format (text, json, markdown)
- include_metadata: Whether to include version/timestamp in output
- compare_with_template: Path to template file for comparison

**Relationships:**
- Used by `--prompt` flag to control output format

## Assumptions

- Slash command prompts are relatively stable and don't change frequently during development
- Embedded prompts will be updated when CLI is rebuilt from source
- AI agents can parse and understand system prompts when provided as text
- Template files in `.canary/templates/commands/` remain the source of truth for prompt content
- Binary size increase from embedded prompts is acceptable

## Constraints

**Technical Constraints:**
- Embedded prompts must be compiled at build time (not runtime)
- Prompts must not contain sensitive information (since they're in the binary)
- Must support all existing slash commands without breaking changes

**Business Constraints:**
- Implementation should not require major CLI refactoring
- Feature should work across all supported platforms (Linux, macOS, Windows)

**Regulatory Constraints:**
- None identified

## Out of Scope

- Dynamic modification of embedded prompts after compilation
- User-customizable embedded prompts (templates remain customizable)
- Automatic synchronization between template files and embedded prompts
- Execution of slash commands via the `--prompt` flag (retrieval only)
- Embedding of non-slash-command prompts or documentation
- Network-based prompt retrieval or remote prompt storage

## Dependencies

- Existing slash command implementation (CBIN-110 and related)
- CLI framework that supports flag parsing
- Build system that can embed static content
- Template files in `.canary/templates/commands/` as source of truth

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Embedded prompts drift from templates | High | Medium | Add CI check to compare embedded vs. template; document rebuild requirements |
| Binary size becomes too large | Medium | Low | Compress prompts; monitor binary size in CI |
| Prompt content leaks sensitive info | High | Low | Audit all prompts before embedding; add security review step |
| `--prompt` flag conflicts with existing flags | Medium | Low | Test all subcommands; use unique flag name if needed |
| Agents rely on outdated embedded prompts | Medium | Medium | Include version info; warn when CLI version is old |

## Clarifications Needed

[NEEDS CLARIFICATION: Should embedded prompts be the source of truth, or should template files remain authoritative?]
**Options:** A) Embedded prompts are canonical - templates are just for reference, B) Templates are canonical - embedded prompts are convenience copies, C) Both are kept synchronized with CI enforcement
**Impact:** Determines which version takes precedence when there's a mismatch and how updates are managed

[NEEDS CLARIFICATION: Should `--prompt` output include only the prompt text, or also metadata and usage examples?]
**Options:** A) Minimal - just the prompt text, B) Standard - prompt + version metadata, C) Verbose - prompt + metadata + examples + related docs
**Impact:** Affects output size, agent parsing complexity, and usefulness for debugging

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [ ] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable and technology-agnostic
- [x] All acceptance scenarios defined
- [x] Edge cases identified
- [x] Scope clearly bounded
- [x] Dependencies and assumptions identified

**Readiness:**
- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [ ] Ready for technical planning (`/canary.plan`) - after clarifications resolved

---

## Implementation Checklist

Break down this requirement into specific implementation points. Each point gets its own CANARY token to help agents locate where to implement changes.

### Core Features

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptEmbedding"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 1: Prompt Embedding System**
- [ ] Implement build-time embedding of slash command prompts into binary
- **Location hint:** "internal/prompts/embed.go" or build scripts
- **Dependencies:** Access to template files at build time

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptFlag"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 2: --prompt Flag Implementation**
- [ ] Add `--prompt` flag to all slash command subcommands
- **Location hint:** "cmd/plan.go", "cmd/specify.go", etc. - each subcommand file
- **Dependencies:** Embedded prompt system

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptRetrieval"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 3: Prompt Retrieval Engine**
- [ ] Implement system to retrieve embedded prompts by command name
- **Location hint:** "internal/prompts/retrieval.go"
- **Dependencies:** Embedded prompts

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptFormatting"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 4: Multi-Format Output**
- [ ] Implement text, JSON, and markdown output formats for prompts
- **Location hint:** "internal/prompts/formatter.go"
- **Dependencies:** Prompt retrieval system

<!-- CANARY: REQ=CBIN-141; FEATURE="SlashCmdList"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 5: Slash Command Listing**
- [ ] Implement `--list-slash-commands` global flag
- **Location hint:** "cmd/root.go"
- **Dependencies:** Embedded prompt metadata

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptComparison"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 6: Template Comparison**
- [ ] Implement `--compare` flag to detect drift between embedded and template prompts
- **Location hint:** "internal/prompts/compare.go"
- **Dependencies:** Prompt retrieval, file system access to templates

### Build System

<!-- CANARY: REQ=CBIN-141; FEATURE="BuildEmbedding"; ASPECT=Dist; STATUS=STUB; UPDATED=2025-10-17 -->
**Build-Time Embedding:**
- [ ] Configure build system to read template files and embed into binary
- **Location hint:** "Makefile", "build.sh", or Go embed directives
- **Dependencies:** Template files in known locations

### Testing Requirements

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptUnitTests"; ASPECT=CLI; STATUS=STUB; TEST=TestPromptEmbedding; UPDATED=2025-10-17 -->
**Unit Tests:**
- [ ] Test prompt retrieval, formatting, comparison logic
- **Location hint:** "internal/prompts/*_test.go"

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptIntegrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestPromptFlags; UPDATED=2025-10-17 -->
**Integration Tests:**
- [ ] Test `--prompt` flag on all subcommands, verify output correctness
- **Location hint:** "test/integration/prompt_test.go"

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptConsistencyTests"; ASPECT=CLI; STATUS=STUB; TEST=TestEmbeddedVsTemplate; UPDATED=2025-10-17 -->
**Consistency Tests:**
- [ ] Test that embedded prompts match template files during CI
- **Location hint:** "test/ci/prompt_consistency_test.go"

### Documentation

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptFlagDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**CLI Documentation:**
- [ ] Document `--prompt` flag in all subcommand help text
- **Location hint:** "cmd/*/flags.go", "docs/cli/"

<!-- CANARY: REQ=CBIN-141; FEATURE="PromptUsageGuide"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Usage Guide:**
- [ ] Create guide for agents on using `--prompt` to understand slash commands
- **Location hint:** ".canary/AGENT_CONTEXT.md", "docs/agent-guide.md"

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-141` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-141; FEATURE="EmbeddedPromptFlag"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```
// CANARY: REQ=CBIN-141; FEATURE="PromptFlag"; ASPECT=CLI; STATUS=IMPL; TEST=TestPromptFlag; UPDATED=2025-10-17
```

**Use `canary implement CBIN-141` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
