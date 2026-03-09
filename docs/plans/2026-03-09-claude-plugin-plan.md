# Claude Code Plugin for CANARY — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a distributable Claude Code plugin that provides full CANARY requirement tracking integration via CLI commands, skills, agents, hooks, and optional MCP.

**Architecture:** Hybrid CLI + optional MCP. All 15 slash commands wrap `canary` CLI subcommands. 3 skills provide workflow/reference guidance. 5 agents handle specialized tasks. A SessionStart hook verifies CLI availability. `.mcp.json` configures optional MCP server.

**Tech Stack:** Markdown (commands, skills, agents), JSON (plugin.json, hooks.json, .mcp.json), Bash (hook script)

---

### Task 1: Scaffold Plugin Directory and Manifest

**Files:**
- Create: `claude-plugin/.claude-plugin/plugin.json`

**Step 1: Create plugin directory structure**

```bash
mkdir -p claude-plugin/.claude-plugin
mkdir -p claude-plugin/commands
mkdir -p claude-plugin/skills/canary-workflow
mkdir -p claude-plugin/skills/canary-token-format
mkdir -p claude-plugin/skills/canary-examples
mkdir -p claude-plugin/agents
mkdir -p claude-plugin/hooks/scripts
```

**Step 2: Write plugin.json**

```json
{
  "name": "canary",
  "version": "0.1.0",
  "description": "CANARY requirement tracking for Claude Code. Embeds requirement tokens in source code with full lifecycle management: specify, plan, implement, scan, verify.",
  "author": {
    "name": "Developer Network",
    "email": "info@devnw.com"
  },
  "license": "Apache-2.0",
  "repository": "https://github.com/devnw/canary",
  "keywords": ["canary", "requirements", "tracking", "tdd", "tokens"]
}
```

**Step 3: Commit**

```bash
git add claude-plugin/
git commit -m "feat(plugin): scaffold Claude Code plugin directory structure"
```

---

### Task 2: SessionStart Hook

**Files:**
- Create: `claude-plugin/hooks/hooks.json`
- Create: `claude-plugin/hooks/scripts/session-start.sh`

**Step 1: Write hooks.json**

```json
{
  "SessionStart": [
    {
      "hooks": [
        {
          "type": "command",
          "command": "bash ${CLAUDE_PLUGIN_ROOT}/hooks/scripts/session-start.sh"
        }
      ]
    }
  ]
}
```

**Step 2: Write session-start.sh**

The script checks for `canary` on PATH, prints version, and checks for `.canary/project.yaml` in the working directory.

```bash
#!/usr/bin/env bash
set -euo pipefail

# Check canary CLI availability
if ! command -v canary &>/dev/null; then
  echo "canary CLI not found on PATH. Install from: https://github.com/devnw/canary"
  echo "Some /canary.* commands will not work without the CLI."
  exit 0
fi

echo "canary is available: $(command -v canary)"

# Print version if available
if canary version &>/dev/null 2>&1; then
  echo ""
  echo "Version:"
  canary version 2>/dev/null || true
fi

# Check for project configuration
if [ -f ".canary/project.yaml" ]; then
  echo ""
  echo "CANARY project detected."
fi
```

**Step 3: Make script executable and commit**

```bash
chmod +x claude-plugin/hooks/scripts/session-start.sh
git add claude-plugin/hooks/
git commit -m "feat(plugin): add SessionStart hook to verify canary CLI"
```

---

### Task 3: Plugin CLAUDE.md

**Files:**
- Create: `claude-plugin/CLAUDE.md`

**Step 1: Write CLAUDE.md**

This is the plugin-level instructions file that Claude loads automatically. It provides:
- CANARY token format reference
- Status progression (STUB -> IMPL -> TESTED -> BENCHED)
- Valid aspects list
- Constitutional principles summary
- Minimal-context philosophy
- CLI command quick reference
- MCP server usage (optional)

Content should be adapted from `.canary/AGENT_CONTEXT.md` but made project-agnostic — no hardcoded project keys. Reference `.canary/project.yaml` for runtime project key discovery.

**Step 2: Commit**

```bash
git add claude-plugin/CLAUDE.md
git commit -m "feat(plugin): add CLAUDE.md with CANARY reference guide"
```

---

### Task 4: MCP Configuration

**Files:**
- Create: `claude-plugin/.mcp.json`

**Step 1: Write .mcp.json**

```json
{
  "mcpServers": {
    "canary": {
      "type": "url",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

**Step 2: Commit**

```bash
git add claude-plugin/.mcp.json
git commit -m "feat(plugin): add optional MCP server configuration"
```

---

### Task 5: Core Workflow Commands (scan, verify, specify, plan, implement, next)

**Files:**
- Create: `claude-plugin/commands/canary.scan.md`
- Create: `claude-plugin/commands/canary.verify.md`
- Create: `claude-plugin/commands/canary.specify.md`
- Create: `claude-plugin/commands/canary.plan.md`
- Create: `claude-plugin/commands/canary.implement.md`
- Create: `claude-plugin/commands/canary.next.md`

**Step 1: Write all 6 command files**

Each command file has YAML frontmatter with `description` field, then the command body. Adapt content from `.canary/commands/<name>.md` with these changes:
- Remove `{{.ReqID}}` and `{{.ProjectKey}}` template variables — replace with generic `<PROJECT_KEY>-<NNN>` or instruct to read from `.canary/project.yaml`
- Remove `SECURITY_REVIEW` hardcoded aspect references
- Keep the same structure: User Input, Outline, steps, Example Output, Guidelines
- For `specify.md`: keep the `scripts.sh` reference but use `${CLAUDE_PLUGIN_ROOT}` — actually no, the script lives in the user's `.canary/scripts/`, so reference that directly
- Ensure all bash commands use `canary` CLI (not MCP)

**Step 2: Commit**

```bash
git add claude-plugin/commands/canary.scan.md claude-plugin/commands/canary.verify.md claude-plugin/commands/canary.specify.md claude-plugin/commands/canary.plan.md claude-plugin/commands/canary.implement.md claude-plugin/commands/canary.next.md
git commit -m "feat(plugin): add core workflow commands (scan, verify, specify, plan, implement, next)"
```

---

### Task 6: Management Commands (constitution, update-stale, list, show, status)

**Files:**
- Create: `claude-plugin/commands/canary.constitution.md`
- Create: `claude-plugin/commands/canary.update-stale.md`
- Create: `claude-plugin/commands/canary.list.md`
- Create: `claude-plugin/commands/canary.show.md`
- Create: `claude-plugin/commands/canary.status.md`

**Step 1: Write all 5 command files**

Same adaptation rules as Task 5. These are simpler commands that mostly wrap a single CLI invocation.

**Step 2: Commit**

```bash
git add claude-plugin/commands/canary.constitution.md claude-plugin/commands/canary.update-stale.md claude-plugin/commands/canary.list.md claude-plugin/commands/canary.show.md claude-plugin/commands/canary.status.md
git commit -m "feat(plugin): add management commands (constitution, update-stale, list, show, status)"
```

---

### Task 7: Query and Utility Commands (grep, files, bug, doc)

**Files:**
- Create: `claude-plugin/commands/canary.grep.md`
- Create: `claude-plugin/commands/canary.files.md`
- Create: `claude-plugin/commands/canary.bug.md`
- Create: `claude-plugin/commands/canary.doc.md`

**Step 1: Write all 4 command files**

Same adaptation rules. `bug.md` and `doc.md` are more complex — preserve the full workflow structure from `.canary/commands/`.

**Step 2: Commit**

```bash
git add claude-plugin/commands/canary.grep.md claude-plugin/commands/canary.files.md claude-plugin/commands/canary.bug.md claude-plugin/commands/canary.doc.md
git commit -m "feat(plugin): add query and utility commands (grep, files, bug, doc)"
```

---

### Task 8: Skills

**Files:**
- Create: `claude-plugin/skills/canary-workflow/SKILL.md`
- Create: `claude-plugin/skills/canary-token-format/SKILL.md`
- Create: `claude-plugin/skills/canary-examples/SKILL.md`

**Step 1: Write canary-workflow skill**

Trigger: user asks "how do I use canary", starts a new feature, or asks about the development workflow. Content: the full specify -> plan -> implement -> scan -> verify pipeline with examples.

**Step 2: Write canary-token-format skill**

Trigger: user is placing or editing CANARY tokens, asks about token format/fields/aspects. Content: complete token field reference, aspect list, status progression rules, auto-promotion logic (TEST= -> TESTED, BENCH= -> BENCHED).

**Step 3: Write canary-examples skill**

Trigger: user asks for canary examples or common patterns. Content: common CLI usage patterns grouped by workflow (getting started, daily development, querying, maintenance).

**Step 4: Commit**

```bash
git add claude-plugin/skills/
git commit -m "feat(plugin): add skills (workflow, token-format, examples)"
```

---

### Task 9: Agents

**Files:**
- Create: `claude-plugin/agents/canary-docs-writer.md`
- Create: `claude-plugin/agents/canary-tests-writer.md`
- Create: `claude-plugin/agents/canary-integration-tests.md`
- Create: `claude-plugin/agents/canary-benchmarks.md`
- Create: `claude-plugin/agents/canary-security-reviewer.md`

**Step 1: Adapt all 5 agent files from `.canary/agents/`**

Each agent needs Claude Code plugin frontmatter:
- `name` — agent identifier
- `description` — when to use (with trigger examples in `<example>` tags)
- `model` — (optional, e.g. sonnet)
- `color` — (optional)

Body: the full system prompt from the existing `.canary/agents/` files, plus `canary_cli_context` with relevant CLI commands for that agent's domain.

Make project-agnostic: remove hardcoded project keys.

**Step 2: Commit**

```bash
git add claude-plugin/agents/
git commit -m "feat(plugin): add specialized agents (docs, tests, integration, benchmarks, security)"
```

---

### Task 10: Verify Plugin Structure

**Step 1: List all files and verify structure matches design**

```bash
find claude-plugin/ -type f | sort
```

Expected output should show:
- `.claude-plugin/plugin.json`
- `.mcp.json`
- `CLAUDE.md`
- 15 files in `commands/`
- 3 `SKILL.md` files in `skills/`
- 5 files in `agents/`
- `hooks/hooks.json` + `hooks/scripts/session-start.sh`

**Step 2: Validate plugin.json is valid JSON**

```bash
python3 -c "import json; json.load(open('claude-plugin/.claude-plugin/plugin.json'))"
```

**Step 3: Validate hooks.json is valid JSON**

```bash
python3 -c "import json; json.load(open('claude-plugin/hooks/hooks.json'))"
```

**Step 4: Validate .mcp.json is valid JSON**

```bash
python3 -c "import json; json.load(open('claude-plugin/.mcp.json'))"
```

**Step 5: Final commit if any fixes needed**

---

### Task 11: Final Commit

**Step 1: Verify all changes**

```bash
git status
git diff --stat HEAD
```

**Step 2: Create final commit if there are uncommitted changes**

```bash
git add claude-plugin/
git commit -m "feat: complete Claude Code plugin for CANARY requirement tracking

Adds distributable plugin with:
- 15 slash commands wrapping canary CLI
- 3 skills (workflow, token-format, examples)
- 5 specialized agents (docs, tests, integration, benchmarks, security)
- SessionStart hook for CLI verification
- Optional MCP server configuration
- CLAUDE.md reference guide"
```
