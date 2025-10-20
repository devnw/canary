// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCANARY_CBIN_102_CLI_Verify validates the verify gate functionality.
// This test ensures the verifier can:
// - Parse GAP_ANALYSIS.md claims (✅ CBIN-XXX format)
// - Detect overclaims (claimed but not implemented)
// - Generate CANARY_VERIFY_FAIL diagnostics with correct REQ IDs
func TestCANARY_CBIN_102_CLI_Verify(t *testing.T) {
	// Setup: temp GAP file with overclaim
	dir := t.TempDir()
	gapFile := filepath.Join(dir, "GAP.md")
	gapContent := `# Gap Analysis

## Implemented
✅ CBIN-999
✅ CBIN-888
`
	if err := os.WriteFile(gapFile, []byte(gapContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Setup: repo with only CBIN-888 token (CBIN-999 missing)
	// Note: Must use STATUS=TESTED or BENCHED for verify gate to consider it valid
	repoDir := t.TempDir()
	repoFile := filepath.Join(repoDir, "code.go")
	repoContent := `package p
// CANARY: REQ=CBIN-888; FEATURE="Present"; ASPECT=API; STATUS=TESTED; TEST=TestFoo; UPDATED=2025-10-15
`
	if err := os.WriteFile(repoFile, []byte(repoContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Execute: scan repo
	rep, err := scan(repoDir, skipDefault, nil, nil)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Execute: verify claims
	diags := verifyClaims(rep, gapFile)

	// Verify: overclaim detected
	if len(diags) == 0 {
		t.Fatal("expected verification failures, got none")
	}

	found := false
	for _, d := range diags {
		if strings.Contains(d, "CANARY_VERIFY_FAIL") && strings.Contains(d, "CBIN-999") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected CANARY_VERIFY_FAIL for CBIN-999, got: %v", diags)
	}

	// Verify: CBIN-888 should NOT be in diagnostics (it exists)
	for _, d := range diags {
		if strings.Contains(d, "CBIN-888") {
			t.Errorf("CBIN-888 should not be flagged as overclaim, but found: %s", d)
		}
	}
}

// setupGAPFixture creates a GAP file and matching report for benchmarking.
// Returns the GAP file path and a report with numClaims matching requirements.
func setupGAPFixture(tb testing.TB, numClaims int) (string, *Report) {
	tb.Helper()
	dir := tb.TempDir()

	// Create GAP file with N claims
	gapFile := filepath.Join(dir, "GAP.md")
	var gapContent strings.Builder
	gapContent.WriteString("# Gap Analysis\n\n")
	for i := 0; i < numClaims; i++ {
		gapContent.WriteString("✅ CBIN-")
		if i < 10 {
			gapContent.WriteString("00")
		} else if i < 100 {
			gapContent.WriteString("0")
		}
		gapContent.WriteString(strings.TrimPrefix(filepath.Base(filepath.Join(dir, fmt.Sprintf("%d", i))), dir))
		gapContent.WriteString(fmt.Sprintf("%d\n", i))
	}
	if err := os.WriteFile(gapFile, []byte(gapContent.String()), 0o644); err != nil {
		tb.Fatal(err)
	}

	// Create matching report with N requirements
	rep := &Report{
		GeneratedAt:  "2025-10-15T00:00:00Z",
		Requirements: make([]Requirement, numClaims),
		Summary: Summary{
			ByStatus:           StatusCounts{"TESTED": numClaims},
			ByAspect:           AspectCounts{"API": numClaims},
			TotalTokens:        numClaims,
			UniqueRequirements: numClaims,
		},
	}
	for i := 0; i < numClaims; i++ {
		rep.Requirements[i] = Requirement{
			ID: fmt.Sprintf("CBIN-%03d", i),
			Features: []Feature{
				{Feature: fmt.Sprintf("Feature%d", i), Aspect: "API", Status: "TESTED", Tests: []string{"TestFoo"}, Updated: "2025-10-15"},
			},
		}
	}

	return gapFile, rep
}

// BenchmarkCANARY_CBIN_102_CLI_Verify benchmarks the verify gate.
// Measures performance of verifying 50 claims against a matching report.
// Baseline target: allocs/op ≤ 10
func BenchmarkCANARY_CBIN_102_CLI_Verify(b *testing.B) {
	gapFile, rep := setupGAPFixture(b, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifyClaims(*rep, gapFile)
	}
}
