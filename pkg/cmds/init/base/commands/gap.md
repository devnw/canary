---
description: Track and query implementation-gap entries so future planning learns from past mistakes
---

## User Input

```text
$ARGUMENTS
```

## Outline

`canary gap` records what went wrong during implementation and lets later `plan`/`implement` runs learn from it.

1. **mark** — Record a gap entry:
   ```bash
   canary gap mark <REQ-ID> <Feature> --category <cat> --description "<what went wrong>" [--action "<fix taken>"] [--aspect API]
   ```
   Categories: `logic_error`, `test_failure`, `performance`, `security`, `edge_case`, `integration`, `documentation`, `other`.
2. **query** — Filter gaps: `canary gap query --req-id CBIN-140 [--category ...] [--feature ...] [--limit N]`
3. **report `<REQ-ID>`** — Full report: counts, grouping, helpfulness.
4. **helpful `<gap-id>`** / **unhelpful `<gap-id>`** — Rate a gap so it's prioritized (or deprioritized) when injected into future planning prompts.
5. **config** — View/update injection tuning: `--max-gaps`, `--min-helpful`, `--ranking` (`helpful_desc`, `recency_desc`, `weighted`).
6. **categories** — List valid categories.

## Example

```bash
canary gap mark CBIN-140 GapTracking --category logic_error \
  --description "Incorrect query ordering in GetFilesByReqID" \
  --action "Added ORDER BY clause"
canary gap query --req-id CBIN-140
```

## Guidelines

- Mark gaps as soon as a mistake is found and fixed — the whole point is future `plan`/`implement` runs surface them.
- Use `report` before starting new work on a requirement that's had prior gaps.
