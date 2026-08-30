// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canaryscan

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

// verifiedFixture writes a git repository holding one requirement and returns
// its root and HEAD commit. The export binds to a commit, so a fixture
// without one can prove nothing.
func verifiedFixture(t *testing.T) (root, commit string) {
	t.Helper()
	root = t.TempDir()
	steps := [][]string{
		{"init", "-q"},
		{"-c", "user.email=scan@example.com", "-c", "user.name=Scan", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-q", "-m", "init"},
	}
	for _, args := range steps {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...) //nolint:gosec // fixed argv
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	body := "package x\n" +
		"// CANARY: REQ=CBIN-050; FEATURE=\"X\"; ASPECT=API; STATUS=TESTED; UPDATED=2026-08-30\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output() //nolint:gosec // fixed argv
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return root, strings.TrimSpace(string(out))
}

// writeStore writes root's evidence store.
func writeStore(t *testing.T, root string, recs ...evidence.Record) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(evidence.File{SchemaVersion: 1, Records: recs})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidencePath(root), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// runScanTo runs a scan of root, writing status.json into a temp dir, and
// returns the decoded report as a raw map so field PRESENCE (not just value)
// can be asserted.
func runScanTo(t *testing.T, root string) map[string]any {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), "status.json")
	var stdout, stderr bytes.Buffer
	if code := Run(Config{Root: root, Out: outPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("scan exit %d: %s", code, stderr.String())
	}
	data, err := os.ReadFile(outPath) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

// TestScanReportCarriesVerifiedExport proves status.json publishes what is
// actually proven, and that "not checked" is distinguishable from "nothing
// verified": with no evidence store the key is absent entirely (a reader must
// treat that as unknown), and with one the key is present -- empty until a
// passing record at this commit exists, then naming the requirement.
func TestScanReportCarriesVerifiedExport(t *testing.T) {
	root, commit := verifiedFixture(t)

	noStore := runScanTo(t, root)
	if _, present := noStore["verified"]; present {
		t.Error("a scan with no evidence store must omit the verified key, not claim an empty one")
	}

	// A store that proves something else: the check ran, and verified nothing.
	writeStore(t, root, evidence.Record{
		ProjectID: "default", RequirementID: "CBIN-999", Feature: "Other", Aspect: "API",
		TestID: "TestOther", Command: "go test ./...", Result: "PASS", CommitSHA: commit,
		ObservedAt: "2026-08-30T00:00:00Z", Runner: "local",
		ArtifactDigest: "sha256:" + strings.Repeat("ab", 32),
	})
	empty := runScanTo(t, root)
	list, present := empty["verified"].([]any)
	if !present {
		t.Fatalf("verified key missing with an evidence store present: %v", empty["verified"])
	}
	if len(list) != 0 {
		t.Errorf("verified = %v, want empty (nothing proven for CBIN-050)", list)
	}

	// Now prove the declared feature at this commit.
	writeStore(t, root, evidence.Record{
		ProjectID: "default", RequirementID: "CBIN-050", Feature: "X", Aspect: "API",
		TestID: "TestX", Command: "go test ./...", Result: "PASS", CommitSHA: commit,
		ObservedAt: "2026-08-30T00:00:00Z", Runner: "local",
		ArtifactDigest: "sha256:" + strings.Repeat("ab", 32),
	})
	proven := runScanTo(t, root)
	got, _ := proven["verified"].([]any)
	if len(got) != 1 || got[0] != "CBIN-050" {
		t.Errorf("verified = %v, want [CBIN-050]", proven["verified"])
	}
}

// TestVerifiedRequirementsIsDeterministic proves the export is sorted, so two
// scans of an unchanged tree produce byte-identical status.json.
func TestVerifiedRequirementsIsDeterministic(t *testing.T) {
	rep := Report{Requirements: []Requirement{
		{ID: "CBIN-030", Features: []Feature{{Feature: "C", Aspect: "API"}}},
		{ID: "CBIN-010", Features: []Feature{{Feature: "A", Aspect: "API"}}},
		{ID: "CBIN-020", Features: []Feature{{Feature: "B", Aspect: "API"}}},
	}}
	commit := strings.Repeat("a", 40)
	rec := func(id, feature string) evidence.Record {
		return evidence.Record{
			ProjectID: "p", RequirementID: id, Feature: feature, Aspect: "API",
			TestID: "T", Command: "c", Result: "PASS", CommitSHA: commit,
			ObservedAt: "2026-08-30T00:00:00Z", Runner: "local",
			ArtifactDigest: "sha256:" + strings.Repeat("ab", 32),
		}
	}
	got := VerifiedRequirements(rep, []evidence.Record{rec("CBIN-030", "C"), rec("CBIN-010", "A")}, "p", commit)
	if strings.Join(got, ",") != "CBIN-010,CBIN-030" {
		t.Errorf("VerifiedRequirements = %v, want sorted [CBIN-010 CBIN-030]", got)
	}
}
