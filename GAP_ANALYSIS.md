# Requirements Gap Analysis

## Claimed Requirements

List requirements that are fully implemented and verified:

✅ CBIN-201 - TicketSources: JIRA/GitLab/GitHub requirement-ID sources with flatfile fallback
✅ CBIN-202 - MermaidRefs: requirement references extracted from mermaid diagrams
✅ CBIN-203 - MermaidGraph: dependency graphs rendered as mermaid with ticket click-through
✅ CBIN-204 - RequirementView: canary view / MCP view one-call requirement picture
✅ CBIN-205 - ContextCaps: small-by-default bounded output across CLI and MCP
✅ CBIN-206 - DiagramRefsIndex: refs table indexing diagram references

## Gaps

List requirements that are planned or in progress:

- [ ] Follow-up: DB-side filtering for MCP bug-list/grep field queries (currently Go-side over full fetch)
- [ ] Follow-up: MCP next lacks the CLI's project id_pattern filter

## Verification

Run verification with:

```bash
canary scan --root . --verify GAP_ANALYSIS.md
```

This will:
- ✅ Verify claimed requirements are TESTED or BENCHED
- ❌ Fail with exit code 2 if claims are overclaimed
