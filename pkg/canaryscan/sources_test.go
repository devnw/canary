package canaryscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/sources"
)

// testCommit is the commit fixture evidence records bind to.
const testCommit = "0123456789abcdef0123456789abcdef01234567"

// evRec builds one PASS evidence record at testCommit for project "p".
func evRec(req, feature, aspect string) evidence.Record {
	return evidence.Record{
		ProjectID: "p", RequirementID: req, Feature: feature, Aspect: aspect,
		TestID: "TestX", Command: "go test ./...", Result: "PASS", CommitSHA: testCommit,
		ObservedAt: "2026-08-30T00:00:00Z", Runner: "local",
		ArtifactDigest: "sha256:" + strings.Repeat("ab", 32),
	}
}

func ticketRegistry(t *testing.T) *sources.Registry {
	t.Helper()
	r, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestCANARY_CBIN_201_VerifyClaimsTicketSource(t *testing.T) {
	reg := ticketRegistry(t)
	gap := filepath.Join(t.TempDir(), "GAP_ANALYSIS.md")
	if err := os.WriteFile(gap, []byte("✅ PLAT-4521\n✅ CBIN-105\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rep := Report{Requirements: []Requirement{
		{ID: "PLAT-4521", Features: []Feature{{Feature: "F", Aspect: "API", Status: "TESTED"}}},
		{ID: "CBIN-105", Features: []Feature{{Feature: "G", Aspect: "API", Status: "IMPL"}}},
	}}
	// Only PLAT-4521 has passing evidence at the current commit; CBIN-105
	// is claimed with none, so it (and only it) is diagnosed.
	recs := []evidence.Record{evRec("PLAT-4521", "F", "API")}
	diags := VerifyClaims(rep, gap, reg, recs, "p", testCommit)
	if len(diags) != 1 {
		t.Fatalf("diags = %v, want exactly 1 (CBIN-105 overclaim)", diags)
	}
	if !strings.Contains(diags[0], "REQ=CBIN-105") {
		t.Errorf("diag should name CBIN-105: %s", diags[0])
	}
}

func TestCANARY_CBIN_201_VerifyClaimsNilRegistryDefaultsToCBIN(t *testing.T) {
	gap := filepath.Join(t.TempDir(), "GAP.md")
	if err := os.WriteFile(gap, []byte("✅ CBIN-101\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rep := Report{Requirements: []Requirement{
		{ID: "CBIN-101", Features: []Feature{{Feature: "F", Aspect: "API", Status: "TESTED"}}},
	}}
	recs := []evidence.Record{evRec("CBIN-101", "F", "API")}
	if diags := VerifyClaims(rep, gap, nil, recs, "p", testCommit); len(diags) != 0 {
		t.Errorf("nil registry must keep legacy CBIN behavior, got %v", diags)
	}
}

func TestCANARY_CBIN_201_NormalizeREQTicketIDsVerbatim(t *testing.T) {
	reg := ticketRegistry(t)
	if got := normalizeREQWithRegistry("PLAT-42", reg); got != "PLAT-42" {
		t.Errorf("jira ID must not be padded: got %q", got)
	}
	if got := normalizeREQWithRegistry("CBIN-42", reg); got != "CBIN-042" {
		t.Errorf("flatfile ID must be padded: got %q", got)
	}
	// legacy prefixes keep padding even without registry entry
	if got := normalizeREQWithRegistry("TASK-7", reg); got != "TASK-007" {
		t.Errorf("legacy TASK padding lost: got %q", got)
	}
}

func TestCANARY_CBIN_201_AnnotateSources(t *testing.T) {
	reg := ticketRegistry(t)
	rep := Report{Requirements: []Requirement{
		{ID: "PLAT-4521"}, {ID: "CBIN-105"}, {ID: "OTHER-1"},
	}}
	AnnotateSources(&rep, reg)
	if rep.Requirements[0].Source != "platform" ||
		rep.Requirements[0].TicketURL != "https://company.atlassian.net/browse/PLAT-4521" {
		t.Errorf("PLAT annotation wrong: %+v", rep.Requirements[0])
	}
	if rep.Requirements[1].Source != "core" || rep.Requirements[1].TicketURL != "" {
		t.Errorf("CBIN annotation wrong: %+v", rep.Requirements[1])
	}
	if rep.Requirements[2].Source != "" {
		t.Errorf("unknown prefix must stay unannotated: %+v", rep.Requirements[2])
	}
}
