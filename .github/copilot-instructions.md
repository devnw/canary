<!-- CANARY:START -->
# CANARY Development Guide for GitHub Copilot

This project uses CANARY requirement tracking. Load only the command file you need from `.canary/commands/`.

## Token format
`// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD`
Status: STUB → IMPL → TESTED → BENCHED.

## Scan & verify (low context)
- **Scan:** `canary scan --root . --out status.json` — use the one-line stdout (`CANARY_SCAN tokens=...`) for metrics.
- **Verify:** `canary scan --root . --verify GAP_ANALYSIS.md --strict` — use stdout `CANARY_VERIFY_OK` or stderr for failures.

## Principles
1. Requirement-First 2. Test-First 3. Evidence-Based (TEST=/BENCH=) 4. Simplicity 5. Keep UPDATED current.

See [.canary/AGENT_CONTEXT.md](./.canary/AGENT_CONTEXT.md) for full reference.

<!-- CANARY:END -->
