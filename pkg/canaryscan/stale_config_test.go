package canaryscan

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

func writeProjectYAML(t *testing.T, root string, stalenessDays int) {
	t.Helper()
	dir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "project:\n  name: test\nverification:\n  staleness_days: " + strconv.Itoa(stalenessDays) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCANARY_CBIN_304_StaleDaysFromConfig proves verification.staleness_days from
// .canary/project.yaml is actually consulted: a 10-day-old TESTED token trips
// --strict against a configured 7-day window, but would not against the 30-day
// default (simulated here via an explicit Config.StaleDays=30 override).
func TestCANARY_CBIN_304_StaleDaysFromConfig(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")
	root := t.TempDir()
	writeProjectYAML(t, root, 7)
	// 10 days before the fixed ref time.
	src := "package x\n" +
		"// CANARY: REQ=CBIN-777; FEATURE=\"Aging\"; ASPECT=API; STATUS=TESTED; TEST=TestAging; UPDATED=2026-08-19\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("7-day project config trips strict", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cfg := Config{Root: root, Out: filepath.Join(root, "status.json"), Strict: true}
		code := Run(cfg, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("expected exit 2 (stale), got %d; stderr=%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "CANARY_STALE REQ=CBIN-777") {
			t.Errorf("expected CANARY_STALE diag for CBIN-777, got: %s", stderr.String())
		}
		if !strings.Contains(stderr.String(), "threshold=7") {
			t.Errorf("expected threshold=7 (from project.yaml), got: %s", stderr.String())
		}
	})

	t.Run("explicit 30-day override does not trip strict", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cfg := Config{Root: root, Out: filepath.Join(root, "status2.json"), Strict: true, StaleDays: 30}
		code := Run(cfg, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("expected exit 0 (not stale under 30d), got %d; stderr=%s", code, stderr.String())
		}
		if strings.Contains(stderr.String(), "CANARY_STALE") {
			t.Errorf("did not expect staleness diag with 30-day override, got: %s", stderr.String())
		}
	})
}

// TestCANARY_CBIN_304_UpdateStaleReportsEvidenceCurrency proves --update-stale's
// new job: for each stale requirement it reports whether that requirement has
// passing evidence at the current commit -- "current" when every declared
// feature/aspect does, "missing" otherwise. It also keeps the v2 multi-segment
// ID coverage: the diag REQ regex must match IDs like CBIN-CLI-001, which the
// old `REQ=([A-Z][A-Z0-9]*-\d+)` pattern could never match.
func TestCANARY_CBIN_304_UpdateStaleReportsEvidenceCurrency(t *testing.T) {
	rep := Report{Requirements: []Requirement{
		{ID: "CBIN-CLI-001", Features: []Feature{{Feature: "V2ID", Aspect: "CLI", Status: "TESTED"}}},
		{ID: "CBIN-777", Features: []Feature{{Feature: "Aging", Aspect: "API", Status: "TESTED"}}},
	}}
	diags := []string{
		"CANARY_STALE REQ=CBIN-CLI-001 updated=2020-01-01 age_days=2431 threshold=30",
		"CANARY_STALE REQ=CBIN-777 updated=2020-01-01 age_days=2431 threshold=30",
	}
	recs := []evidence.Record{evRec("CBIN-CLI-001", "V2ID", "CLI")}

	got := ReportEvidenceCurrency(rep, diags, recs, "p", testCommit)
	want := []string{
		"CANARY_UPDATE_STALE req=CBIN-777 evidence=missing",
		"CANARY_UPDATE_STALE req=CBIN-CLI-001 evidence=current",
	}
	if len(got) != len(want) {
		t.Fatalf("lines = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestCANARY_CBIN_304_UpdateStaleMutatesNothing proves --update-stale no longer
// rewrites source. Rewriting UPDATED= made a stale claim look fresh without any
// new proof -- exactly the failure evidence-backed verification exists to
// prevent -- so the flag now only reports, and the file on disk is untouched
// byte for byte.
func TestCANARY_CBIN_304_UpdateStaleMutatesNothing(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")
	root := t.TempDir()
	src := "package x\n" +
		"// CANARY: REQ=CBIN-501; FEATURE=\"Dup\"; ASPECT=API; STATUS=TESTED; TEST=TestDup; UPDATED=2020-01-01\n"
	path := filepath.Join(root, "a.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cfg := Config{Root: root, Out: filepath.Join(root, "status.json"), UpdateStale: true}
	if code := Run(cfg, &stdout, &stderr); code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", code, stderr.String())
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != src {
		t.Errorf("--update-stale mutated source:\n%s", after)
	}
	if !strings.Contains(stderr.String(), "CANARY_UPDATE_STALE req=CBIN-501 evidence=missing") {
		t.Errorf("expected an evidence-currency report for CBIN-501, got: %s", stderr.String())
	}
	// root is not a git repository, so HeadCommit fails; a plain
	// --update-stale run (no --verify requested) must report that with the
	// neutral skip marker, never the CANARY_VERIFY_FAIL marker reserved for
	// actual --verify failures.
	if strings.Contains(stderr.String(), "CANARY_VERIFY_FAIL") {
		t.Errorf("plain --update-stale run must not emit CANARY_VERIFY_FAIL, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "CANARY_UPDATE_STALE_SKIP reason=no_commit") {
		t.Errorf("expected CANARY_UPDATE_STALE_SKIP reason=no_commit, got: %s", stderr.String())
	}
}
