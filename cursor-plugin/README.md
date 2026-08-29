# CANARY Requirements – Cursor Plugin

Cursor plugin for [CANARY](https://github.com/devnw/canary) requirement tracking: adds a rule so the AI uses CANARY scan/verify and token format with minimal context.

## Install from this repo (local)

1. **Copy the plugin into Cursor’s plugin directory** (so Cursor sees it as an installed plugin):

   ```bash
   # From the canary repo root
   mkdir -p ~/.cursor/plugins
   cp -r cursor-plugin ~/.cursor/plugins/canary-requirements
   ```

   If Cursor uses a different path (e.g. per-workspace), copy `cursor-plugin` into that location instead.

2. **Restart Cursor** (or reload the window) so it picks up the new plugin.

3. **Enable the plugin** in Cursor: Settings → Plugins (or Extensions) and ensure “CANARY Requirements” is enabled.

## Project-level setup (recommended)

In a project that uses CANARY, run:

```bash
canary init --local
```

That creates:

- `.cursor/rules/canary-requirements.mdc` – same rule, scoped to the project
- `.cursor/mcp.json` – optional MCP config for `canary mcp` (list, show, scan tools)

Then start the MCP server when you want tool-based access:

```bash
canary mcp
```

## What this plugin does

- **Rule:** When you’re in `.canary/`, `GAP_ANALYSIS.md`, or spec/plan files, the AI is guided to:
  - Use only the relevant command file from `.canary/commands/`
  - Use the one-line stdout from `canary scan` / verify instead of loading full status.json or GAP_ANALYSIS
  - Apply the CANARY token format (REQ, FEATURE, ASPECT, STATUS, UPDATED)

No extra context is loaded unless the task needs it.
