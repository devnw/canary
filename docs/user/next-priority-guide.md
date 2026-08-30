# Next Priority Command User Guide

**Requirement:** CP-252 (originally CBIN-132)
**Status:** Complete
**Last Updated:** 2026-08-30

## Overview

`canary next` picks the highest-priority requirement that is actually ready to
be worked on, and — with `--prompt` — generates the implementation guidance for
it. It answers one question, "what should be started now?", and it names the
evidence it answered from.

**Key properties:**

- **Automated prioritization**: declared `PRIORITY` first, then unstarted work
  before work in progress.
- **Evidence-gated dependencies**: a dependency is only complete when passing
  evidence at the current commit says so. `STATUS=TESTED` is a claim, not proof.
- **Honest about its source**: every answer says whether it came from the token
  index (`database`) or from a direct scan of the tree (`filesystem`).
- **Read-only**: `next` never creates or updates the index, the evidence store,
  or your tokens.

## Getting Started

### Prerequisites

- CANARY CLI installed.
- A git repository: freshness and evidence are both bound to the current
  commit, so a tree with no readable `HEAD` can prove nothing.
- (Recommended) a current index: `canary index`.
- (Optional) `.canary/memory/constitution.md` and `.canary/templates/next-prompt-template.md`
  for `--prompt` output.

### Quick Start

```bash
canary next            # what to work on
canary next --prompt   # full implementation guidance for it
```

## Where the Answer Comes From

`next` uses the token index **only while that index still describes the tree in
front of it**. The index is used when it exists, was built from this root, at
this commit, by this binary's token grammar. Otherwise the tree is scanned
directly with the same scanner, ignore rules, and token grammar `canary scan`
uses, so the two can never disagree about what the repository contains.

| `"source"` | Meaning |
|------------|---------|
| `database` | Answered from a current index. |
| `filesystem` | The index was missing, unbuilt, stale, or built elsewhere, so the tree was scanned. |

Two consequences worth knowing:

- **"All requirements completed" is only ever claimed from a current index.** A
  filesystem scan that finds no candidate says exactly that and no more — it
  cannot see the whole project's history, so it never congratulates you.
- **Nothing is written.** An absent index is not created to answer a read.

## Usage

### Summary Mode

```bash
canary next
```

**Output:**

```
Next: CBIN-101 - ScannerCore (Priority: 1, Status: STUB)
Run with --prompt for full implementation guidance.
```

### Prompt Generation Mode

```bash
canary next --prompt
```

Renders `.canary/templates/next-prompt-template.md` with the requirement's
specification, the constitution, test-first guidance, token placement examples,
success criteria, and dependency information. `--prompt-arg <file|name>` makes
one extra prompt available to the template as `{{.PromptArg}}` /
`{{.PromptContent}}`.

The template is read only on this path: a `--format json` or `--dry-run` run
never fails on a template it was not going to use.

### Filtering

```bash
canary next --status STUB     # only unstarted work
canary next --status IMPL     # only work that still needs tests
canary next --aspect Engine   # only Engine work
canary next --aspect CLI
```

### Dry Run Mode

```bash
canary next --dry-run
```

**Output:**

```
Next priority (dry run): CBIN-101 - ScannerCore
Priority: 1 | Status: STUB | Aspect: Engine
Location: a.go
Source: filesystem
```

### Machine-Readable Output

```bash
canary next --format json
```

`--format` takes `json` or `text` (default `text`). In JSON mode stdout carries
one JSON object and nothing else; notes and warnings go to stderr.

```json
{
  "req_id": "CBIN-101",
  "feature": "ScannerCore",
  "aspect": "Engine",
  "status": "STUB",
  "priority": 1,
  "file_path": "a.go",
  "updated": "2026-08-30",
  "source": "filesystem"
}
```

`source` is always present. The requirement fields are absent when there was
nothing to select, and `message` says why instead:

```json
{
  "source": "filesystem",
  "message": "no actionable requirements found"
}
```

`priority` is the requirement's declared `PRIORITY`, or `5` when its token
declared none — the same neutral default the index applies, and the same number
selection ranked it at.

JSON is a machine contract and outranks the human-facing modes: `--format json
--dry-run` returns the JSON object, not prose.

> **Deprecated:** `--json` is still accepted as an alias for `--format json`
> and now prints a deprecation warning **on stderr** (stdout stays pure JSON).
> An explicit `--format` always wins over the alias. Use `--format json`.

### External and Peer Dependencies

A dependency owned by a configured ticket source or `peers:` project resolves to
`satisfied`, `unsatisfied`, or `unknown`. Only `satisfied` clears it: an
unresolvable dependency blocks selection, because handing an agent a requirement
whose prerequisite might not exist is a wrong answer rather than a graceful
degradation. The block is explained on stderr:

```
note: external dependency ENG-9 blocks selection: no cached ticket status
```

Refresh the cache (`canary ticket status --refresh`) or, to start anyway:

```bash
canary next --allow-unknown-external
```

```
note: external dependency ENG-9 is unresolved (allowed): no cached ticket status
Next: CBIN-200 - Consumer (Priority: 5, Status: STUB)
```

A dependency that names nothing at all — an unconfigured prefix that no peer
knows either — always blocks; there is no flag for that, because there is
nothing to resolve.

## Common Workflows

### Workflow 1: Agent Autonomous Work

```bash
# 1. Agent completes the current task and records evidence for it
# 2. Agent asks for the next one
canary next --prompt

# 3. Agent implements the requirement following the generated guidance
# 4. Agent places/updates CANARY tokens
# 5. Agent runs the tests and records evidence (canary evidence ...)
# 6. Repeat from step 2
```

Recording evidence is part of the loop, not an afterthought: until a
requirement has passing evidence at the current commit, everything that depends
on it stays blocked.

### Workflow 2: Human Developer Daily Planning

```bash
canary index                 # make the index current for this commit
canary next                  # what's next
cat .canary/specs/CBIN-138-multiline-tokens/spec.md
canary next --prompt > implementation-guidance.md
```

### Workflow 3: Filtered Sprint Planning

```bash
canary list --aspect API --status STUB
canary next --aspect API --prompt
```

### Workflow 4: Dependency-Driven Development

```bash
# Current state:
#   CBIN-105 (PRIORITY=1, DEPENDS_ON=CBIN-104)
#   CBIN-104 (PRIORITY=3, STATUS=STUB)

canary next
# CBIN-104 — CBIN-105 is blocked until CBIN-104 is proven complete

# Implement CBIN-104, run its tests, record the result, then:
canary verify
canary next
# CBIN-105 — now unblocked
```

Marking CBIN-104 `STATUS=TESTED` is not what unblocks CBIN-105. A PASS evidence
record for every feature/aspect CBIN-104 declares, at this commit, is.

## Examples

### Example 1: First Implementation

```bash
$ canary next --prompt
```

**Output (abridged):**

````markdown
# Implementation Guidance: CBIN-101 - ScannerCore

## Priority Information
- Requirement ID: CBIN-101
- Priority: 1 (Highest)
- Status: STUB
- Aspect: Engine

## Specification

[Full spec.md content loaded here...]

## Constitutional Principles

From .canary/memory/constitution.md:

**Article IV: Test-First Imperative**
All features SHALL be implemented using test-first development...

## Implementation Guidance

### Step 1: Write Tests (RED phase)

```go
// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_101_Engine_BasicScan; UPDATED=2026-08-30
func TestCANARY_CBIN_101_Engine_BasicScan(t *testing.T) {
    // Test implementation...
}
```
````

### Example 2: No Work Available

**Scenario:** a current index, and every requirement complete.

```bash
$ canary next
```

**Output:**

```
🎉 All requirements completed! No work available.

Suggestions:
  • Run: canary verify
  • Review completed requirements
  • Consider creating new specifications
```

This wording is reachable **only** from a current index. Scanned directly, the
same empty answer reads:

```
no actionable requirements found (source=filesystem)

The tree was scanned directly because no current index was available.
  • Run: canary index   (then re-run canary next)
```

### Example 3: Everything Left Is Blocked

Work that exists but is blocked is never completion, whatever the source:

```bash
$ canary next
```

**Output:**

```
no unblocked requirements (1 blocked by unmet dependencies)

Every remaining candidate is waiting on a dependency:
  • Run: canary deps check <REQ-ID>   (which dependency, and in what state)
  • A local dependency blocks until evidence at this commit proves it:
    run its tests, record the result, then: canary verify
  • Unresolved external/peer dependencies are noted on stderr above
```

### Example 4: No Index Yet

```bash
$ canary next --format json
```

**Output:**

```json
{
  "req_id": "CBIN-101",
  "feature": "ScannerCore",
  "aspect": "Engine",
  "status": "STUB",
  "priority": 1,
  "file_path": "a.go",
  "updated": "2026-08-30",
  "source": "filesystem"
}
```

The answer is the same shape whichever source produced it; only `"source"`
differs. Run `canary index` to make later runs answer from the index (and to
let a genuinely finished project say so).

## Best Practices

### For AI Agents

1. **Record evidence, then ask again** — dependencies clear on evidence, not on
   status edits.
2. **Always use `--prompt`** for full context.
3. **Follow test-first guidance** — RED → GREEN → REFACTOR.
4. **Read `"source"`** before believing a completion claim.
5. **Do not paper over a block** — `--allow-unknown-external` is a deliberate
   risk, not a default.

### For Human Developers

1. Start the day with `canary next`.
2. Preview with `--dry-run` before generating a prompt.
3. Filter by `--aspect` to match what you are working in.
4. Re-run `canary index` after committing, so `next` can use the fast path.
5. Adjust `PRIORITY` fields (or `canary prioritize`) when plans change.

### For CI/CD Systems

1. **Use `--format json`** — stdout is JSON only.
2. **Check exit codes** — 0 = success (including "no work"), non-zero = error.
3. **Branch on `"source"`** — treat a `filesystem` answer as "index not current".
4. **Handle `"message"`** — the requirement fields are absent when nothing was
   selected.

## Priority Determination

### Factor 1: Explicit PRIORITY Field

```
PRIORITY=1  (Highest priority)
...
PRIORITY=5  (The default applied to a token that declares none)
...
PRIORITY=10 (Lowest priority)
```

Lower numbers are selected first. Both sources honor the declaration: the index
stores it, and a filesystem scan carries it out of the token. A token that
declares no `PRIORITY` is ranked — and reported — at 5.

### Factor 2: STATUS Value

Among candidates of equal priority:

1. **STUB** — not yet implemented (selected first)
2. **IMPL** — implemented, still needs tests
3. **TESTED** / **BENCHED** — not actionable, skipped unless `--status` asks for them

### Factor 3: Dependencies (DEPENDS_ON)

A candidate whose dependencies are not all complete is passed over, and the
blocking dependency is normally selected instead. "Complete" means:

- **Local dependency**: every feature/aspect it declares has a PASS evidence
  record for this project at the current commit.
- **External/peer dependency**: it resolved `satisfied`. `unsatisfied` blocks;
  `unknown` blocks unless `--allow-unknown-external` is passed.

### Factor 4: Deterministic Tie-Breaks

Remaining ties are broken so that two runs over an unchanged tree pick the same
requirement: the index orders by `priority ASC, updated_at DESC` (most recently
touched first), and a filesystem scan orders by priority, then status, then
requirement id, feature, and aspect.

## Troubleshooting

### "no actionable requirements found (source=filesystem)"

The tree was scanned and held nothing selectable. Check, in order:

1. Tokens exist and parse: `canary scan --root . --out status.json`
2. `.canaryignore` is not excluding them.
3. You are in the project root.
4. Hidden paths (tests, templates, agent directories) are excluded by design.

### "All requirements completed" but work remains

That claim comes from a current index, so the index and the tree have diverged
in a way `next` could not detect, or the work is filtered out:

1. `canary list --status STUB` — is it there at all?
2. `canary list --include-hidden --status STUB` — is it in a hidden path?
3. Commit your changes and re-run `canary index` — an index is current for the
   commit it was built at.

### Everything is blocked

1. `canary deps check <REQ-ID>` — which dependency, in what state.
2. For a local dependency: run its tests, record evidence, then `canary verify`.
3. For an external/peer dependency: `canary ticket status --refresh`, or scan
   the peer so its `status.json` carries a current verification export.

### Prompt Generation Fails

```
Error: render prompt: read template: ...
```

1. `.canary/templates/next-prompt-template.md` must exist.
2. The specification path referenced by the requirement must exist.
3. Run without `--prompt` to test selection independently.

## FAQ

**Q: Can I override the selection?**

A: Adjust `PRIORITY` in the tokens (or `canary prioritize <REQ-ID> <feature> <priority>`),
narrow with `--status`/`--aspect`, or pick one explicitly with `canary implement <REQ-ID>`.

**Q: Does the command modify anything?**

A: No. `canary next` is read-only — it does not even create an index it finds
missing.

**Q: Does it work without an index?**

A: Yes. It scans the tree with the same scanner `canary scan` uses and reports
`"source": "filesystem"`. The scan is slower on a large tree, and a filesystem
answer never claims the project is finished.

**Q: Why is a dependency still blocking after I marked it TESTED?**

A: Because a declaration is not proof. Run its tests, record the result in
`.canary/evidence.json`, and confirm with `canary verify`.

**Q: How does this integrate with slash commands?**

A: `/canary.next` runs `canary next --prompt` and feeds the result to the agent
as implementation guidance.

## Related Documentation

- [canary list](./list-command-guide.md) — viewing all requirements
- [canary implement](./implement-command-guide.md) — manual requirement selection
- [canary prioritize](./prioritize-command-guide.md) — adjusting priorities
- [Ticket sources & peers](./ticket-sources-guide.md) — external dependency resolution
- [Constitutional Principles](../../.canary/memory/constitution.md) — project governance

## Integration Examples

### Integration 1: GitHub Actions Workflow

```yaml
name: Auto-implement next requirement

on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  auto-implement:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Install canary
        run: go install ./cmd/canary
      - name: Get next requirement
        id: next
        run: |
          canary index
          canary next --format json > next.json
          echo "req_id=$(jq -r '.req_id // empty' next.json)" >> $GITHUB_OUTPUT
          echo "source=$(jq -r '.source' next.json)" >> $GITHUB_OUTPUT
      - name: Implement with AI agent
        if: steps.next.outputs.req_id != ''
        run: |
          # Call your AI agent here with steps.next.outputs.req_id
```

### Integration 2: Post-commit Hook

```bash
#!/bin/bash
# .git/hooks/post-commit

canary next --dry-run
```

### Integration 3: Claude Code Workflow

```
User: /canary.next
Claude: [Reads generated prompt]
Claude: I'll implement CBIN-138 (MultilineTokens) following the specification...
[Claude implements, tests, records evidence, and updates tokens]
User: continue
Claude: [Runs /canary.next again for the next task]
```

---

*Implementation status: BENCHED (tested and benchmarked)*
