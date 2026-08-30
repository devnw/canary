---
description: Report which stale CANARY tokens still have current passing evidence
---


## User Input

```text
$ARGUMENTS
```

## Outline

Report the evidence currency of stale CANARY tokens (TESTED/BENCHED with
UPDATED older than the configured staleness window, default 30 days).

`--update-stale` does **not** rewrite `UPDATED=` dates. Rewriting a date made
a stale claim look fresh without any new proof — exactly the failure that
evidence-backed verification exists to prevent. It reports, and changes
nothing on disk.

1. **Read project configuration**:
   - Load `.canary/project.yaml` to determine the project key

2. **Run the staleness report**:
   ```bash
   canary scan --root . --update-stale
   ```

   For every stale requirement, one line on stderr:

   ```text
   CANARY_UPDATE_STALE req=<PROJECT_KEY>-<NNN> evidence=current
   CANARY_UPDATE_STALE req=<PROJECT_KEY>-<NNN> evidence=missing
   ```

   - `evidence=current` — every feature/aspect the requirement declares has a
     PASS record at the current commit in `.canary/evidence.json`. The date is
     old; the proof is not. Re-run the tests and re-record evidence to refresh
     the date honestly, or leave it.
   - `evidence=missing` — the requirement has no current proof. This is real
     staleness: the work must be re-verified, not re-dated.

3. **Refresh the evidence** (this is what actually makes a claim current):
   ```bash
   go test -count=1 -json ./... > gotest.json
   canary evidence from-go-test --project <PROJECT_KEY> --commit "$(git rev-parse HEAD)" < gotest.json > evidence.json
   canary evidence ingest --in evidence.json --out .canary/evidence.json
   ```

4. **Re-verify**:
   ```bash
   canary verify --root . --claims GAP_ANALYSIS.md --format text
   ```

5. **Generate the report**:
   ```markdown
   ## Stale Token Evidence Report

   **Report Date:** YYYY-MM-DD
   **Stale Requirements:** N

   ### Current evidence (date is old, proof is not)
   - <PROJECT_KEY>-<NNN>: src/api/auth.go

   ### Missing evidence (must be re-verified)
   - <PROJECT_KEY>-<NNN>: internal/cache/cache.go

   **Next Steps:**
   - Re-run the tests for every requirement listed as `evidence=missing`
   - Re-record evidence, then run `canary verify` to confirm the claims
   ```

## Guidelines

- **Never re-date without re-proving**: an `UPDATED=` bump is not evidence
- **Only TESTED/BENCHED tokens are stale-checked** (not STUB/IMPL)
- **Nothing is mutated**: this command only reads and reports
- **Evidence is commit-bound**: it must be re-recorded after every commit
