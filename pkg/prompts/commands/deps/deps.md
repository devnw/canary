# Deps Command Prompt

## Purpose
Check, visualize, and validate requirement dependencies declared in spec files.

## Task
`canary deps` is a parent command with four real subcommands. There is no `show`, `list`, or `add` subcommand -- dependencies are declared inside spec.md files (via `DEPENDS_ON=`), not added through the CLI.

## Subcommands

### check <req-id>
Is this requirement's dependencies satisfied? Only `TESTED`/`BENCHED` status counts; `IMPL` is insufficient.
```bash
canary deps check CBIN-147 [--show-satisfied]
```

### graph <req-id>
Visual dependency tree (direct + transitive), from `.canary/specs/`.
```bash
canary deps graph CBIN-147 [--status] [--format ascii|mermaid]
```
- `--format mermaid` renders a flowchart with click-through ticket links where a source is configured (default: `ascii`, a Unicode box-drawing tree).
- `--status` marks each dependency satisfied (✅) or blocking (❌).

### reverse <req-id>
What would be blocked if this requirement changes?
```bash
canary deps reverse CBIN-146
```

### validate
Check the whole graph for circular dependencies, self-dependencies, and dependencies on missing requirements.
```bash
canary deps validate
```

## Output Format (graph, ascii)
```
CBIN-147
├── CBIN-005 [TESTED]
└── CBIN-010 [BENCHED]

Dependency Status: ✅ All dependencies satisfied
```

## Standards
- Dependencies are parsed from `.canary/specs/<REQ-ID>-*/spec.md`; requirements without a spec file have no dependency graph.
- `check`/`graph --status` classify a dependency as satisfied only when ALL of its tokens are `TESTED` or `BENCHED`.
- `validate` should be run in CI alongside `canary scan --verify` to catch cycles before they block `canary next`.
