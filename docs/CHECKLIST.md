# Canary CLI — Parity Checklist

| Requirement | TokenParse | EnumValidate | NormalizeREQ | StatusJSON | CSVExport | VerifyGate | Staleness30d | SelfCanary | CI | Perf50k<10s> |
|------------:|:----------:|:------------:|:------------:|:----------:|:---------:|:----------:|:------------:|:----------:|:--:|:------------:|
| CBIN-101    | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ |
| CBIN-102    | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ |
| CBIN-103    | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ | ◻ |
| Overall     | ◐ (basic parser present but non-conformant) | ◐ (STATUS validated only) | ◻ | ◻ (pretty, not minified & schema drift) | ◐ (explodes rows; needs deterministic order spec) | ◐ (regex + claim inference diverges) | ◻ (60d window; spec wants 30d) | ◻ | ◻ (no GH workflow) | ◻ (no perf evidence) |

Legend: ✅ = proven by tests/evidence; ◐ = partial; ◻ = missing.

Evidence Notes:
- Current implementation uses legacy `REQ-GQL-###` prefix; spec requires `CBIN-###`.
- No self-canary tokens (CBIN-101..103) exist yet.
- JSON output is indented not canonical/minified; lacks ordered zero counts for all STATUS values.
- Staleness threshold hard-coded at 60 days instead of required 30 days.
- Verify logic treats any line with ✅/Implemented as implemented; spec requires strict `^✅ CBIN-###` and exit code 2 on missing TESTED/BENCHED tokens.
- Performance, CI pipeline, and acceptance tests absent.
