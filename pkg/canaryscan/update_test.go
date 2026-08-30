package canaryscan

import (
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

// TestCANARY_CBIN_201_UpdateStaleTicketIDs proves the stale-diag REQ parser
// still handles ticket-sourced IDs (PLAT-4521, not just the local CBIN
// series) now that --update-stale reports evidence currency instead of
// rewriting UPDATED= dates.
func TestCANARY_CBIN_201_UpdateStaleTicketIDs(t *testing.T) {
	rep := Report{Requirements: []Requirement{
		{ID: "PLAT-4521", Features: []Feature{{Feature: "Ingest", Aspect: "API", Status: "TESTED"}}},
	}}
	diags := []string{"CANARY_STALE REQ=PLAT-4521 updated=2020-01-01 age_days=2431 threshold=30"}

	got := ReportEvidenceCurrency(rep, diags, nil, "p", testCommit)
	want := "CANARY_UPDATE_STALE req=PLAT-4521 evidence=missing"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("lines = %v, want [%q]", got, want)
	}

	recs := []evidence.Record{evRec("PLAT-4521", "Ingest", "API")}
	got = ReportEvidenceCurrency(rep, diags, recs, "p", testCommit)
	want = "CANARY_UPDATE_STALE req=PLAT-4521 evidence=current"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("lines = %v, want [%q]", got, want)
	}
}

// TestUpdateStaleEvidenceCurrencyIgnoresUnparseableDiags proves a diagnostic
// with no REQ= field contributes nothing rather than producing a line for an
// empty requirement ID.
func TestUpdateStaleEvidenceCurrencyIgnoresUnparseableDiags(t *testing.T) {
	if got := ReportEvidenceCurrency(Report{}, []string{"CANARY_PARSE_ERROR file=x err=\"boom\""}, nil, "p", testCommit); got != nil {
		t.Fatalf("lines = %v, want none", got)
	}
}

// TestUpdateStaleEvidenceCurrencyUndeclaredClaim proves a stale requirement
// the report knows nothing about is reported as missing evidence, never as
// vacuously current.
func TestUpdateStaleEvidenceCurrencyUndeclaredClaim(t *testing.T) {
	diags := []string{"CANARY_STALE REQ=CBIN-999 updated=2020-01-01 age_days=1 threshold=30"}
	got := ReportEvidenceCurrency(Report{}, diags, nil, "p", testCommit)
	want := "CANARY_UPDATE_STALE req=CBIN-999 evidence=missing"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("lines = %v, want [%q]", got, want)
	}
}
