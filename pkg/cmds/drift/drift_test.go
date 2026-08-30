// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package drift

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/drift"
	"devnw.dev/canary/pkg/storage"
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

func hashOf(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// writeIndexedRepo builds a git repo with one committed file carrying a
// CANARY token, and an index whose baseline content hash matches that file on
// disk — the CURRENT starting point. Return the dir so a test can then mutate
// the file to force DRIFTED.
func writeIndexedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, nil, "init", "-q")

	content := "package foo\n\n" +
		"// CANARY: REQ=CBIN-950; FEATURE=\"Foo\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-01\n" +
		"func Foo() {}\n"
	full := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, nil, "add", "foo.go")
	runGit(t, dir, nil, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "msg")

	dbPath := filepath.Join(dir, ".canary", "canary.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertToken(&storage.Token{
		ReqID: "CBIN-950", Feature: "Foo", Aspect: "API", Status: "IMPL",
		FilePath: "foo.go", ContentHash: hashOf(t, full), ProjectID: "default", UpdatedAt: "2026-08-01",
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.PutIndexMeta(storage.IndexMeta{
		Root: dir, ProjectID: "default", CommitSHA: "0000000000000000000000000000000000000000",
		ParserSchema: canaryscan.ParserSchemaVersion, ScanDigest: "digest", IndexedAt: "2026-01-01T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	return dir
}

func runDriftJSON(t *testing.T, args ...string) []drift.ReqState {
	t.Helper()
	cmd := CreateDriftCommand()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute %v: %v\nstderr: %s", args, err, errBuf.String())
	}
	var states []drift.ReqState
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &states); err != nil {
		t.Fatalf("decode drift json: %v\nstdout: %q", err, out.String())
	}
	return states
}

// TestCANARY_CP_278_Cmd_JSON_Clean: an unchanged indexed repo reports every
// requirement CURRENT, and stdout is a clean JSON array (no prose banner).
func TestCANARY_CP_278_Cmd_JSON_Clean(t *testing.T) {
	dir := writeIndexedRepo(t)
	states := runDriftJSON(t, "--root", dir, "--format", "json")
	if len(states) != 1 || states[0].RequirementID != "CBIN-950" || states[0].State != drift.StateCurrent {
		t.Fatalf("want [CBIN-950 CURRENT], got %+v", states)
	}
}

// TestCANARY_CP_278_Cmd_JSON_Drifted: mutating the file after indexing flips
// the requirement to DRIFTED.
func TestCANARY_CP_278_Cmd_JSON_Drifted(t *testing.T) {
	dir := writeIndexedRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "foo.go"), []byte("package foo\n// changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	states := runDriftJSON(t, "--root", dir, "--format", "json")
	if len(states) != 1 || states[0].State != drift.StateDrifted {
		t.Fatalf("want CBIN-950 DRIFTED, got %+v", states)
	}
}

// TestCANARY_CP_278_Cmd_Table_NoDrift: table mode over a clean repo names the
// current requirement and prints the zeroed-drift summary line.
func TestCANARY_CP_278_Cmd_Table_NoDrift(t *testing.T) {
	dir := writeIndexedRepo(t)
	cmd := CreateDriftCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "CBIN-950: CURRENT") {
		t.Errorf("output missing current requirement: %q", out.String())
	}
	if !strings.Contains(out.String(), "CANARY_DRIFT requirements=1 current=1 drifted=0 unknown=0") {
		t.Errorf("output missing summary line: %q", out.String())
	}
}

// TestCANARY_CP_278_Cmd_NoIndex: a repository that was never indexed is told
// to build one; nothing is created.
func TestCANARY_CP_278_Cmd_NoIndex(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, nil, "init", "-q")
	cmd := CreateDriftCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--root", dir, "--format", "json"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected an error for an unindexed repo, got success:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".canary", "canary.db")); err == nil {
		t.Error("drift must not create a database as a side effect")
	}
}

func TestCANARY_CP_278_SummaryLine_Format(t *testing.T) {
	states := []drift.ReqState{
		{RequirementID: "CBIN-1", State: drift.StateCurrent},
		{RequirementID: "CBIN-2", State: drift.StateDrifted},
		{RequirementID: "CBIN-3", State: drift.StateUnknown},
	}
	line := summaryLine(states)
	want := "CANARY_DRIFT requirements=3 current=1 drifted=1 unknown=1"
	if line != want {
		t.Fatalf("summary line = %q, want %q", line, want)
	}
}

// TestCANARY_CP_278_StrictShouldFail unit-tests the pure --strict decision.
// The os.Exit(2) side effect it gates is not exercised in-process.
func TestCANARY_CP_278_StrictShouldFail(t *testing.T) {
	if strictShouldFail(nil) {
		t.Error("strictShouldFail(nil) = true, want false")
	}
	if strictShouldFail([]drift.ReqState{{State: drift.StateCurrent}}) {
		t.Error("strictShouldFail(all current) = true, want false")
	}
	if !strictShouldFail([]drift.ReqState{{State: drift.StateDrifted}}) {
		t.Error("strictShouldFail(drifted) = false, want true")
	}
	if !strictShouldFail([]drift.ReqState{{State: drift.StateUnknown}}) {
		t.Error("strictShouldFail(unknown) = false, want true")
	}
}
