# Files Command Prompt

## Purpose
Find files containing tokens for a requirement.

## Task
Implement `canary files <REQ-ID>` to list all files implementing a requirement.

## Expected Behavior
```bash
canary files CBIN-001
```

## Output Format
```
Files for CBIN-001:

Implementation Files:
  src/auth/api.go           (API, TESTED)
  src/auth/session.go       (API, IMPL)
  cli/auth/login.go         (CLI, IMPL)

Test Files:
  tests/auth_api_test.go    (TestCANARY_CBIN_001_Login)
  tests/auth_session_test.go (TestCANARY_CBIN_001_Session)

Benchmark Files:
  bench/auth_bench_test.go  (BenchmarkCANARY_CBIN_001_Login)

Total: 6 files (3 impl, 2 test, 1 bench)
```

## File Categories
- Implementation files (contains CANARY token)
- Test files (contains TestCANARY_*)
- Benchmark files (contains BenchmarkCANARY_*)

## Standards
- Query database for requirement ID
- Collect unique file paths
- Categorize by file type
- Show aspect and status for each
- Link to test/bench functions
