package canaryscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

// gapFile writes a claims file and returns its path.
func gapFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "GAP_ANALYSIS.md")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// oneClaim is a report with a single declared requirement.
var oneClaim = Report{Requirements: []Requirement{
	{ID: "CBIN-101", Features: []Feature{{Feature: "F", Aspect: "API", Status: "TESTED"}}},
}}

// TestVerifyClaims_NoEvidenceFails proves a STATUS=TESTED declaration is not
// evidence: with no records, the claim fails.
func TestVerifyClaims_NoEvidenceFails(t *testing.T) {
	diags := VerifyClaims(oneClaim, gapFile(t, "✅ CBIN-101\n"), nil, nil, "p", testCommit)
	want := "CANARY_VERIFY_FAIL REQ=CBIN-101 reason=no_current_evidence"
	if len(diags) != 1 || diags[0] != want {
		t.Fatalf("diags = %v, want [%q]", diags, want)
	}
}

// TestVerifyClaims_CurrentEvidencePasses proves a passing record at the
// current commit clears the claim.
func TestVerifyClaims_CurrentEvidencePasses(t *testing.T) {
	recs := []evidence.Record{evRec("CBIN-101", "F", "API")}
	if diags := VerifyClaims(oneClaim, gapFile(t, "✅ CBIN-101\n"), nil, recs, "p", testCommit); len(diags) != 0 {
		t.Fatalf("diags = %v, want none", diags)
	}
}

// TestVerifyClaims_UndeclaredClaimFails proves a claim for a requirement no
// token declares fails rather than passing vacuously.
func TestVerifyClaims_UndeclaredClaimFails(t *testing.T) {
	diags := VerifyClaims(Report{}, gapFile(t, "✅ CBIN-404\n"), nil, nil, "p", testCommit)
	want := "CANARY_VERIFY_FAIL REQ=CBIN-404 reason=no_current_evidence"
	if len(diags) != 1 || diags[0] != want {
		t.Fatalf("diags = %v, want [%q]", diags, want)
	}
}

// TestVerifyClaims_EmptyClaimsFileIsSilent proves `scan --verify` keeps its
// historical contract: a claims file with no claims produces no diagnostics.
// (`canary verify` is where an empty claims file is itself a failure.)
func TestVerifyClaims_EmptyClaimsFileIsSilent(t *testing.T) {
	if diags := VerifyClaims(oneClaim, gapFile(t, "# nothing claimed\n"), nil, nil, "p", testCommit); len(diags) != 0 {
		t.Fatalf("diags = %v, want none", diags)
	}
}

// TestVerifyClaims_UnreadableGapFileReportsParseError proves a claims file
// that cannot be read is reported, not treated as "no claims".
func TestVerifyClaims_UnreadableGapFileReportsParseError(t *testing.T) {
	diags := VerifyClaims(oneClaim, filepath.Join(t.TempDir(), "missing.md"), nil, nil, "p", testCommit)
	if len(diags) != 1 || !strings.Contains(diags[0], "CANARY_PARSE_ERROR") {
		t.Fatalf("diags = %v, want a parse error", diags)
	}
}
