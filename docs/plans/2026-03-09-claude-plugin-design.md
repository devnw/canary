# Claude Code Plugin for CANARY — Design Document

**Date:** 2026-03-09
**Status:** Approved

## Overview

A distributable Claude Code plugin that provides full CANARY requirement tracking integration. Hybrid architecture: CLI-centric by default, with optional MCP server support when `canary mcp` is running.

## Plugin Location

- **Distributable:** `claude-plugin/` in the canary repo
- **Project-local:** Thin `.claude/` config for canary development itself

## Structure

```
claude-plugin/
├── plugin.json
├── .mcp.json
├── CLAUDE.md
├── commands/
│   ├── canary.scan.md
│   ├── canary.verify.md
│   ├── canary.specify.md
│   ├── canary.plan.md
│   ├── canary.implement.md
│   ├── canary.next.md
│   ├── canary.constitution.md
│   ├── canary.update-stale.md
│   ├── canary.list.md
│   ├── canary.show.md
│   ├── canary.status.md
│   ├── canary.grep.md
│   ├── canary.files.md
│   ├── canary.bug.md
│   └── canary.doc.md
├── skills/
│   ├── canary-workflow.md
│   ├── canary-token-format.md
│   └── canary-examples.md
├── agents/
│   ├── canary-docs-writer.md
│   ├── canary-tests-writer.md
│   ├── canary-integration-tests.md
│   ├── canary-benchmarks.md
│   └── canary-security-reviewer.md
└── hooks/
    └── session-start.sh
```

## Components

### Commands (15)

Each wraps the corresponding `canary` CLI subcommand. Adapted from `.canary/commands/*.md` but made project-agnostic — reads `.canary/project.yaml` at runtime for project key. Follows minimal-context philosophy: parse stdout one-liners, don't load full JSON/MD unless needed.

Commands: scan, verify, specify, plan, implement, next, constitution, update-stale, list, show, status, grep, files, bug, doc.

### Skills (3)

- **canary-workflow** — End-to-end workflow guide (specify -> plan -> implement -> scan -> verify)
- **canary-token-format** — Token format reference (fields, aspects, status progression)
- **canary-examples** — Common CLI usage patterns

### Agents (5)

Adapted from `.canary/agents/` with Claude Code agent frontmatter:
- docs-writer, tests-writer, integration-tests, benchmarks, security-reviewer

### Hooks (1)

- **SessionStart** — Verifies `canary` CLI is on PATH, prints version

### MCP (Optional)

`.mcp.json` configures `canary mcp` as streamable HTTP on `localhost:8080/mcp`. When server is running, Claude auto-discovers 18 MCP tools as supplements to CLI commands.

## Design Decisions

1. CLI primary, MCP supplementary — CLI covers all 33+ commands; MCP covers 18
2. Project-agnostic — reads project.yaml at runtime
3. Minimal context — stdout one-liners, not full file reads
4. Content reuse — adapt existing .canary/commands/ and .canary/agents/
