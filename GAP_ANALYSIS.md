# Requirements Gap Analysis

## Claimed Requirements

List requirements that are fully implemented and verified:

✅ CP-267 - TicketSources: JIRA/GitLab/GitHub requirement-ID sources with flatfile fallback
✅ CP-268 - MermaidRefs: requirement references extracted from mermaid diagrams
✅ CP-269 - MermaidGraph: dependency graphs rendered as mermaid with ticket click-through
✅ CP-270 - RequirementView: canary view / MCP view one-call requirement picture
✅ CP-271 - ContextCaps: small-by-default bounded output across CLI and MCP
✅ CP-272 - DiagramRefsIndex: refs table indexing diagram references
✅ CP-281 - MCPServerCommand: mcp subcommand construction (flags, Use) covered by TestMCPCommandCreation

## Gaps

List requirements that are planned or in progress:

- [ ] Follow-up: DB-side filtering for MCP bug-list/grep field queries (currently Go-side over full fetch)
- [ ] Follow-up: MCP next lacks the CLI's project id_pattern filter
- [ ] Follow-up: thread registry through Scan signature (remove activeRegistry global; fixes direct-Scan normalization bypass)
- [ ] Follow-up: nested four-backtick mermaid fence handling
- [ ] Follow-up: next-command candidate-window undershoot with id_pattern
- [ ] Follow-up: --order-by allowlist in storage
- [ ] Follow-up: --db flag parity for canary view
- [ ] Follow-up: unify index's ExtractField parser with canaryscan
- [ ] Follow-up: canaryscan parse <!-- --> markdown tokens (md-heading parity)
- [ ] Follow-up: sanitize ;/quotes in bug TITLE at buildBugToken
- [ ] Follow-up: remove transitional CBIN source + dual id_pattern after CBIN-CLI-001 token migrates
- [ ] Follow-up: delete dead embedded stub packages x3
- [ ] Follow-up: migrate FetchRemoteStatus to /search/jql (deprecated endpoint)

## Verification

Run verification with:

```bash
canary scan --root . --verify GAP_ANALYSIS.md
```

This will:
- ✅ Verify claimed requirements are TESTED or BENCHED
- ❌ Fail with exit code 2 if claims are overclaimed
