// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package drift

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/drift"
)

func runGit(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	if env != nil {
		cmd.Env = env
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", full, err, out)
	}
}

// writeToken writes a single-requirement Go file with a CANARY token whose
// UPDATED date is in the past, then commits it at commitDate — so the file's
// last-commit date is after the token's UPDATED and code-drift fires.
func writeDriftingRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, nil, "init", "-q")

	content := "package foo\n\n" +
		"// CANARY: REQ=CBIN-950; FEATURE=\"Foo\"; ASPECT=API; STATUS=IMPL; UPDATED=2020-01-01\n" +
		"func Foo() {}\n"
	full := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, nil, "add", "foo.go")
	env := append(os.Environ(), "GIT_AUTHOR_DATE=2026-08-20T12:00:00+00:00", "GIT_COMMITTER_DATE=2026-08-20T12:00:00+00:00")
	runGit(t, dir, env, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "msg")
	return dir
}

func TestCANARY_CBIN_305_Execute_CodeDrift(t *testing.T) {
	dir := writeDriftingRepo(t)
	findings, err := scanAndDetect(dir, 0)
	if err != nil {
		t.Fatalf("scanAndDetect: %v", err)
	}
	found := false
	for _, f := range findings {
		if f.ReqID == "CBIN-950" && f.Kind == "code-drift" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a code-drift finding for CBIN-950, got %+v", findings)
	}
}

func TestCANARY_CBIN_305_SummaryLine_Format(t *testing.T) {
	dir := writeDriftingRepo(t)
	findings, err := scanAndDetect(dir, 0)
	if err != nil {
		t.Fatalf("scanAndDetect: %v", err)
	}
	line := summaryLine(findings)
	if !strings.HasPrefix(line, "CANARY_DRIFT requirements=") {
		t.Fatalf("summary line = %q", line)
	}
	for _, want := range []string{"requirements=", "code_drift=", "stale=", "doc_drift="} {
		if !strings.Contains(line, want) {
			t.Errorf("summary line %q missing %q", line, want)
		}
	}
}

func TestCANARY_CBIN_305_Cmd_JSON(t *testing.T) {
	dir := writeDriftingRepo(t)
	cmd := CreateDriftCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--root", dir, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var payload struct {
		Findings []struct {
			ReqID string `json:"req_id"`
			Kind  string `json:"kind"`
		} `json:"findings"`
		Summary struct {
			Requirements int `json:"requirements"`
			CodeDrift    int `json:"code_drift"`
			Stale        int `json:"stale"`
			DocDrift     int `json:"doc_drift"`
		} `json:"summary"`
	}
	// The JSON payload is the first line; the summary line follows on stdout too.
	firstLine := strings.SplitN(out.String(), "\n", 2)[0]
	if err := json.Unmarshal([]byte(firstLine), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", firstLine, err)
	}
	if payload.Summary.CodeDrift < 1 {
		t.Errorf("summary.code_drift = %d, want >= 1", payload.Summary.CodeDrift)
	}
	found := false
	for _, f := range payload.Findings {
		if f.ReqID == "CBIN-950" && f.Kind == "code-drift" {
			found = true
		}
	}
	if !found {
		t.Errorf("findings missing CBIN-950 code-drift: %+v", payload.Findings)
	}
}

// TestCANARY_CBIN_305_StrictShouldFail unit-tests the pure --strict
// decision. The os.Exit(2) side effect it gates (mirroring
// pkg/canaryscan.Run's exit-code pattern) is not exercised in-process —
// see the transcript in .superpowers/sdd/task-7-report.md for an empirical
// `canary drift --strict; echo $?` run against a built binary.
func TestCANARY_CBIN_305_StrictShouldFail(t *testing.T) {
	if strictShouldFail(nil) {
		t.Error("strictShouldFail(nil) = true, want false")
	}
	if strictShouldFail([]drift.Finding{}) {
		t.Error("strictShouldFail(empty) = true, want false")
	}
	findings := []drift.Finding{{ReqID: "CBIN-950", Kind: drift.KindCodeDrift}}
	if !strictShouldFail(findings) {
		t.Error("strictShouldFail(non-empty) = false, want true")
	}
}

func TestCANARY_CBIN_305_Cmd_Strict_NoFindings_NoError(t *testing.T) {
	dir := t.TempDir() // empty root: no tokens, no drift
	runGit(t, dir, nil, "init", "-q")
	cmd := CreateDriftCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--root", dir, "--strict"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func TestCANARY_CBIN_305_Cmd_Table_NoDrift(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, nil, "init", "-q")
	cmd := CreateDriftCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "No drift detected") {
		t.Errorf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "CANARY_DRIFT requirements=0") {
		t.Errorf("output missing zeroed summary line: %q", out.String())
	}
}

// TestCANARY_CBIN_305_Cmd_TimestampEnv_StaleFlips is an end-to-end test at
// the CLI layer: it sets CANARY_TEST_TIMESTAMP (which scanAndDetect reads
// via canaryscan.RefTimeFromEnv) so a fixed reference time, not the wall
// clock, drives staleness, and asserts the summary counts flip as refTime
// crosses the 30-day default staleness window.
func TestCANARY_CBIN_305_Cmd_TimestampEnv_StaleFlips(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, nil, "init", "-q")

	content := "package foo\n\n" +
		"// CANARY: REQ=CBIN-951; FEATURE=\"Foo\"; ASPECT=API; STATUS=TESTED; TEST=TestFoo; UPDATED=2026-08-01\n" +
		"func Foo() {}\n"
	full := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, nil, "add", "foo.go")
	env := append(os.Environ(), "GIT_AUTHOR_DATE=2026-08-01T12:00:00+00:00", "GIT_COMMITTER_DATE=2026-08-01T12:00:00+00:00")
	runGit(t, dir, env, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "msg")

	// refTime 14 days after UPDATED: inside the 30-day window, not stale.
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-15T00:00:00Z")
	cmd := CreateDriftCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "CANARY_DRIFT requirements=0 code_drift=0 stale=0 doc_drift=0") {
		t.Errorf("expected no findings within the staleness window, output = %q", out.String())
	}

	// refTime 45 days after UPDATED: past the 30-day window, stale.
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-09-15T00:00:00Z")
	cmd2 := CreateDriftCommand()
	var out2 bytes.Buffer
	cmd2.SetOut(&out2)
	cmd2.SetErr(&out2)
	cmd2.SetArgs([]string{"--root", dir})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out2.String(), "CANARY_DRIFT requirements=1 code_drift=0 stale=1 doc_drift=0") {
		t.Errorf("expected exactly one stale finding past the staleness window, output = %q", out2.String())
	}
}
