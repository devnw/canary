## DocCmd (Documentation Management)

```yaml
---
description: Manage documentation tracking, creation, and verification for CANARY requirements (strict, verifiable, no-mock/no-simulate)
command: DocCmd
version: 2.2
subcommands: [create, update, status, report]
outputs:
  - summary_json: STDOUT (unwrapped JSON; strict schema per subcommand below)
  - human_text: STDOUT (concise table/metrics for operator UX; optional unless report --format text)
runtime_guarantees:
  no_mock_data: true
  no_simulation_of_results: true
  canary_logging: required_when(context_usage>=0.7 || on_milestones)
---
```

<!-- CANARY: REQ=CBIN-136; FEATURE="DocCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

### 1) Inputs

* **Raw args:** subcommand + flags (e.g., `create <REQ-ID> --type user --output <path>`, `update [<REQ-ID>|--all [--stale-only]]`, `status [<REQ-ID>|--all [--stale-only]]`, `report [--format json] [--show-undocumented]`).
* **Repository layout (assumed):**

  * Templates: `.canary/templates/docs/`
  * Constitution: `.canary/memory/constitution.md`
  * Tokens live in source; requirement specs under `.canary/specs/`
  * Tracking DB (implementation-defined) storing DocPath/DocHash/DocType/DocCheckedAt/DocStatus.

### 2) Preconditions & Resolution

1. **Resolve REQ** (if required): exact `<REQ-ID>` or fail with `ERROR_REQ_NOT_FOUND`.
2. **Validate template by `--type`** ∈ {user, api, technical, feature, architecture}; else `ERROR_UNSUPPORTED_TYPE`.
3. **Line endings policy:** normalize CRLF→LF before hashing. Abbreviate SHA256 to first 16 hex chars (64 bits).
4. **Token linkage:** DOC= uses `type:path` pairs; DOC_HASH= is a comma‑aligned list matching DOC order. If counts mismatch → `ERROR_DOC_HASH_MISMATCH`.

### 3) Hashing & Status Model (authoritative)

* **Algorithm:** SHA256 over normalized bytes; **Abbrev:** 16 hex chars.
* **Statuses:** `DOC_CURRENT`, `DOC_STALE`, `DOC_MISSING`, `DOC_UNHASHED`.
* **Staleness detection:** content hash changes or missing DOC_HASH for a referenced doc ⇒ not current.
* **Performance target:** hashing ≈ `<0.01ms/KB` (guideline). Batch ops may process in parallel.

### 4) Planning & Parallelism

* For `--all` batches, build a **Work DAG** with **Concurrency Groups (CG)**:

  * **CG‑1 Enumerate**: Scan tokens/specs to collect `(req_id, type, path, prior_hash)`.
  * **CG‑2 Hash**: Compute hashes for all candidate files in parallel (I/O bound; no shared writes).
  * **CG‑3 Update**: Apply DB writes and token updates; **join** before any shared file edits.
* If platform lacks true parallelism, interleave non‑blocking tasks preserving join points. 

### 5) Subcommand Behaviors (must do; never simulate)

**create `<REQ-ID>` --type <type> [--output <path>]**

* Generate from template; write file; **do not** invent content beyond template placeholders.
* Compute initial hash; update DB; instruct to add/merge `DOC=` & `DOC_HASH=` into the **actual source token(s)** (paths + line numbers in evidence).
* Output both UX text and JSON summary.

**update [<REQ-ID>] [--all [--stale-only]]**

* Recompute hashes for the target set; update DB and token `DOC_HASH=` in place; mark `DOC_CURRENT`.
* With `--stale-only`, limit to files where computed hash ≠ stored.
* Output counts and per‑doc change details.

**status [<REQ-ID>] [--all [--stale-only]]**

* Evaluate each referenced doc; return status per file + roll‑up metrics.

**report [--format <text|json>] [--show-undocumented]**

* Compute coverage, breakdown by type, staleness stats, and recommendations.
* If `--show-undocumented`, list REQs lacking DOC fields.

### 6) CANARY Snapshot Protocol (compact; low‑token)

Emit when **context ≥70%**, after **batch enumeration**, and after **writes**:

```bash
canary log --kind state --data '{
  "t":"<ISO8601>","s":"doc|verify|update",
  "f":[["<docpath1>",1,999],["<tokenfile>",L1,L2]],
  "k":["req:<REQ-ID>","op:<create|update|status|report>","n:<affected-count>"],
  "fp":["<disproven assumption>"],
  "iss":["<tracker-ids-or-n/a>"],
  "nx":["<next actions>"]
}'
```

Keys: `t` time • `s` stage • `f` file+line spans • `k` key facts • `fp` false‑positives • `iss` issues • `nx` next steps.

### 7) Output Contract (strict; schema per subcommand)

Return artifacts in this order. **Do not wrap JSON in code fences.** 

**A. HUMAN_TEXT (optional unless report text)**
Begin with: `=== HUMAN_TEXT BEGIN ===` … `=== HUMAN_TEXT END ===`

**B. SUMMARY_JSON (unwrapped JSON)** — envelopes:

**Common envelope**

```json
{
  "cmd": "create|update|status|report",
  "ok": true,
  "req_id": "{{.ReqID}}-<ASPECT>-XXX",
  "metrics": { "processed": 0, "current": 0, "stale": 0, "missing": 0, "unhashed": 0 },
  "canary": { "emitted": true, "last_id": "<id-or-n/a>" },
  "items": []
}
```

**items.create** element

```json
{
  "type": "user|api|technical|feature|architecture",
  "path": "docs/.../file.md",
  "hash": "8f434346648f6b96",
  "token_updates": [{ "file": "path/to/code.go", "line": 123, "action": "add|merge" }]
}
```

**items.update/status** element

```json
{
  "type": "user|api|technical|feature|architecture",
  "path": "docs/.../file.md",
  "status": "DOC_CURRENT|DOC_STALE|DOC_MISSING|DOC_UNHASHED",
  "old_hash": "aaaaaaaaaaaaaaaa",
  "new_hash": "bbbbbbbbbbbbbbbb"
}
```

**report** adds:

```json
{
  "coverage_pct": 0.0,
  "by_type": { "user": {"current":0,"stale":0,"missing":0,"unhashed":0}, "api": {...} },
  "undocumented": ["{{.ReqID}}-<ASPECT>-123","{{.ReqID}}-<ASPECT>-200"]
}
```

### 8) Validation Gates (compute and report)

* **Article VII (Documentation Currency):** UPDATED fields and DOC_HASH reflect current content.
* **Security Gate:** No secrets in docs or tokens; paths are relative.
* **Integrity Gate:** DOC & DOC_HASH lists index‑aligned; counts match; hashes are 16‑hex.
* **Performance Gate:** Batch hashing parallelized or interleaved; avoid blocking joins.

### 9) Failure Modes (return one, with reason + remediation)

* `ERROR_REQ_NOT_FOUND(req_id)`
* `ERROR_UNSUPPORTED_TYPE(type)`
* `ERROR_TEMPLATE_MISSING(type)`
* `ERROR_DOC_HASH_MISMATCH(doc_count, hash_count)`
* `ERROR_FILE_IO(path,reason)`
* `ERROR_DB_WRITE(reason)`
* `ERROR_TOKEN_UPDATE(path,line,reason)`

### 10) Quality Checklist (auto‑verify before output)

* Hash normalization (CRLF→LF) applied; 16‑hex SHA256 abbrev computed.
* All referenced docs exist (or flagged MISSING).
* Token DOC/DOC_HASH updated where applicable; line numbers recorded.
* Batch ops produced **Work DAG & CGs**; joins prior to shared writes.
* JSON envelope valid; **no code‑fence wrapping**; human text (if any) concise. 
* CANARY snapshot emitted when required.

### 11) Example HUMAN_TEXT (report, text mode)

Begin/End markers as above; include: coverage %, counts per status/type, top recommendations (e.g., “Run `canary doc update --all --stale-only` to refresh stale docs”), and N undocumented REQs listing (if `--show-undocumented`).

---

### What changed & why (brief)

* **Deterministic outputs:** unified **SUMMARY_JSON** envelopes per subcommand enable tooling to parse and assert success. 
* **Section delimiting & structure:** clearer inputs → gates → behavior → outputs for maintainability. 
* **Parallel batch handling:** explicit Work DAG + CGs + join points for `--all` and `--stale-only`. 
* **No‑mock/no‑simulate:** runtime guarantees make it explicit that hashing/tokens/DB writes must be real operations. 

---

### Assumptions & Risks

* Assumes `canary` CLI is available and writable paths exist; DB interface is callable.
* If your platform lacks true concurrency, interleave I/O and compute safely at join barriers. 

### Targeted questions (for fit)

1. Confirm canonical storage for the docs DB (file vs. service) and write API.
2. Are additional doc types needed or type aliases expected?
3. Should `report --format json` include per‑REQ deltas since last check?
4. Default threshold for CANARY snapshots (keep 70%)?
