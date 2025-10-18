// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCANARY_CBIN_101_Engine_ScanBasic validates the core scanning functionality.
// This test ensures the scanner can:
// - Parse CANARY tokens from multiple files
// - Correctly categorize by STATUS (STUB, IMPL)
// - Correctly categorize by ASPECT (API, CLI, Engine)
// - Generate accurate summary counts
func TestCANARY_CBIN_101_Engine_ScanBasic(t *testing.T) {
	// Setup: temp dir with 3 fixture files
	dir := t.TempDir()
	fixtures := map[string]string{
		"file1.go": `package p
// CANARY: REQ=CBIN-200; FEATURE="Alpha"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20
`,
		"file2.go": `package p
// CANARY: REQ=CBIN-201; FEATURE="Bravo"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-09-20
`,
		"file3.go": `package p
// CANARY: REQ=CBIN-202; FEATURE="Charlie"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-09-20
`,
	}
	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Execute: scan directory
	rep, err := scan(dir, skipDefault, nil, nil)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Verify: status counts
	if rep.Summary.ByStatus["STUB"] != 1 {
		t.Errorf("expected STUB=1, got %d", rep.Summary.ByStatus["STUB"])
	}
	if rep.Summary.ByStatus["IMPL"] != 2 {
		t.Errorf("expected IMPL=2, got %d", rep.Summary.ByStatus["IMPL"])
	}

	// Verify: requirement count
	if len(rep.Requirements) != 3 {
		t.Errorf("expected 3 requirements, got %d", len(rep.Requirements))
	}

	// Verify: aspect diversity
	if rep.Summary.ByAspect["API"] != 1 {
		t.Errorf("expected API=1, got %d", rep.Summary.ByAspect["API"])
	}
	if rep.Summary.ByAspect["CLI"] != 1 {
		t.Errorf("expected CLI=1, got %d", rep.Summary.ByAspect["CLI"])
	}
	if rep.Summary.ByAspect["Engine"] != 1 {
		t.Errorf("expected Engine=1, got %d", rep.Summary.ByAspect["Engine"])
	}
}

// setupFixture creates a test directory with numFiles CANARY tokens.
// Used by benchmarks to create consistent test fixtures.
func setupFixture(tb testing.TB, numFiles int) string {
	tb.Helper()
	dir := tb.TempDir()
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf(`package p
// CANARY: REQ=CBIN-%03d; FEATURE="Feature%d"; ASPECT=API; STATUS=IMPL; UPDATED=2025-09-20
`, i, i)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%d.go", i)), []byte(content), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
	return dir
}

// BenchmarkCANARY_CBIN_101_Engine_Scan benchmarks the scanning engine.
// Measures performance of scanning 100 files with CANARY tokens.
// Baseline target: allocs/op ≤ 10
func BenchmarkCANARY_CBIN_101_Engine_Scan(b *testing.B) {
	dir := setupFixture(b, 100)
	skip := skipDefault
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scan(dir, skip, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCANARY_CBIN_101_Engine_Scan50k validates <10s requirement for 50k files.
// This benchmark establishes the performance baseline for large-scale scanning.
// Target: <10s per operation (10,000,000,000 ns/op)
// Runs only once (N=1) due to setup cost.
func BenchmarkCANARY_CBIN_101_Engine_Scan50k(b *testing.B) {
	// Create 50k file fixture (one-time setup)
	b.StopTimer()
	dir := setupFixture(b, 50000)
	skip := skipDefault
	b.StartTimer()

	// Run scan N times
	for i := 0; i < b.N; i++ {
		_, err := scan(dir, skip, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	// Validate <10s target
	if b.Elapsed() > 10*time.Second && b.N == 1 {
		b.Errorf("50k scan took %v, exceeds 10s target", b.Elapsed())
	}
}
