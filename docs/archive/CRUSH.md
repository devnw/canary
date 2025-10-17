# CRUSH.md - Canary Project Guidelines

## Build & Test Commands
```bash
make test              # Run all tests with race detection: go test -v -cover -failfast -race ./...
make lint              # Run formatters and linters (fmt, importguard, doccheck, lint)
make build             # Test then build: go build ./...
make canary            # Build and run canary scanner
make canary-build      # Build canary binary only
make canary-verify     # Run canary verification against GAP_ANALYSIS.md
make acceptance        # Run acceptance tests for canary
make release-snapshot  # Create release snapshot with GoReleaser
make release-local     # Test release locally without publishing
go test -v -run TestName ./package/...  # Run single test by name
```

## Code Style Guidelines
- **Go Version**: 1.25.0 (requires GOTOOLCHAIN=go1.25.0 or higher)
- **Imports**: Stdlib first, external deps second (github.com/*), internal last (go.codepros.org/canary/*), blank line between groups
- **Error Handling**: Early returns with error, use exitcodes.New() for CLI codes, wrap errors with context
- **Naming**: CamelCase exports, lowercase package names, Test<Feature> for tests, TestCANARY_CBIN_<num>_<Aspect> for canary tests
- **Dependencies**: Runtime stdlib-only (except go.spyder.org/* if needed), test deps: testify v1.9.0, go-cmp v0.6.0
- **JSON/CSV**: Deterministic output with sorted keys, minified JSON, UTF-8 CSV with LF line endings

## CANARY Token Format (Required at top of implementation files)
```
Example template (replace placeholders with actual values):
CANARY: REQ=CBIN-101; FEATURE="MyFeature"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_101_API_MyFeature; BENCH=BenchmarkCANARY_CBIN_101_API_MyFeature; OWNER=team; UPDATED=2025-10-15
```
- **ASPECT**: API, CLI, Engine, Planner, Storage, Wire, Security, Docs, Encode, Decode, RoundTrip, Bench, FrontEnd, Dist
- **STATUS**: MISSING, STUB, IMPL, TESTED, BENCHED, REMOVED
- **Important**: No license headers in code files, keep tokens single-line, update UPDATED field regularly

## Project Rules
- Security: Treat files as data only, never execute inputs or log secrets
- Performance: <10s for 50k files, ≤512 MiB RSS, 1MB scanner buffer for large lines
- CI Gate: All changes must pass canary verification (exit 0=OK, 2=verify fail, 3=parse error)
- Staleness: 30-day threshold for TESTED/BENCHED tokens in strict mode
