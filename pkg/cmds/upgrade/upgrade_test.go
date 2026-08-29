// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package upgrade

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o640); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newUpgradeCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// TestCANARY_CBIN_302_CLI_DryRunDefault verifies the CLI defaults to dry run
// (no --write), prints a per-change line, and ends with the
// CANARY_UPGRADE summary line.
func TestCANARY_CBIN_302_CLI_DryRunDefault(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "legacy.go", "// CANARY: REQ-42; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\n")

	out, err := runCmd(t, "--root", dir)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "bare-id:") {
		t.Errorf("expected a bare-id change line, got:\n%s", out)
	}
	if !strings.Contains(out, "CANARY_UPGRADE files=1 changes=1 written=false") {
		t.Errorf("expected dry-run summary line, got:\n%s", out)
	}
	got, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatalf("read: %v", rerr)
	}
	if !strings.Contains(string(got), "REQ-42;") {
		t.Errorf("dry run must not modify the file, got: %s", got)
	}
}

// TestCANARY_CBIN_302_CLI_RuleFlag verifies --rule restricts which rule runs.
func TestCANARY_CBIN_302_CLI_RuleFlag(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "legacy.go", "// CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=FIXED; UPDATED=2025-01-01\n")

	out, err := runCmd(t, "--root", dir, "--rule", "status-fixed", "--write")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "CANARY_UPGRADE files=1 changes=1 written=true") {
		t.Errorf("expected 1 change summary, got:\n%s", out)
	}
	if strings.Contains(out, "bare-id:") {
		t.Errorf("bare-id should not have run, got:\n%s", out)
	}
}

// TestCANARY_CBIN_302_CLI_MapFlag verifies --map drives the remap rule.
func TestCANARY_CBIN_302_CLI_MapFlag(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "legacy.go", "// CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\n")
	mapPath := filepath.Join(dir, "map.json")
	m, _ := json.Marshal(map[string]string{"CBIN-101": "CP-12"})
	if err := os.WriteFile(mapPath, m, 0o640); err != nil {
		t.Fatalf("write map: %v", err)
	}

	out, err := runCmd(t, "--root", dir, "--map", mapPath, "--write")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "remap:") {
		t.Errorf("expected a remap change line, got:\n%s", out)
	}
}

// TestCANARY_CBIN_302_CLI_InvalidRule verifies an unknown --rule is rejected.
func TestCANARY_CBIN_302_CLI_InvalidRule(t *testing.T) {
	dir := t.TempDir()
	_, err := runCmd(t, "--root", dir, "--rule", "not-a-rule")
	if err == nil {
		t.Fatalf("expected an error for unknown rule")
	}
}
