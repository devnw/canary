# CANARY Policy (Repo‑Wide)

**Purpose.** Make every feature claim searchable and verifiable by linking requirements → code → tests → docs.

## Token (single line, place at top of implementation files or relevant tests)
`CANARY: REQ=CBIN-<###>; FEATURE="<name>"; ASPECT=<ASPECT>; STATUS=<STATUS>; TEST=<TestCANARY_CBIN_<###>_<Aspect>_<Short>>; BENCH=<BenchmarkCANARY_CBIN_<###>_<Aspect>_<Short>>; OWNER=<team-or-alias>; UPDATED=<YYYY-MM-DD>`

- **ASPECT** ∈ ["API","CLI","Engine","Planner","Storage","Wire","Security","Docs","Decode","Encode","RoundTrip","Bench","FrontEnd","Dist"]
- **STATUS** ∈ ["MISSING","STUB","IMPL","TESTED","BENCHED","REMOVED"]

### Greps

- `rg -n "CANARY:\s*REQ=CBIN-" src internal cmd tools`
- `rg -n "TestCANARY_CBIN_" src internal cmd tools tests`
- `rg -n "BenchmarkCANARY_CBIN_" src internal cmd tools tests`
