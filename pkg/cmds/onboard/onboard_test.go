// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package onboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mustWrite creates path (and its parent dirs) with content, failing the
// test on any error.
func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// polyglotTree builds a synthetic, non-canary-adopted tree: a Go entry point,
// a Markdown file carrying a pre-seeded CANARY:MIGRATE note, and a Python
// file — deliberately with no .canary/ directory, so NextID must fall back.
func polyglotTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "cmd", "app", "main.go"),
		"package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n")
	mustWrite(t, filepath.Join(dir, "docs", "NOTES.md"),
		"# Notes\n\n<!-- CANARY:MIGRATE legacy auth lives in pkg/auth; see CBIN-105 -->\n")
	mustWrite(t, filepath.Join(dir, "scripts", "tool.py"),
		"print(\"hi\")\n")
	return dir
}

func TestCANARY_CBIN_303_Analyze(t *testing.T) {
	dir := polyglotTree(t)

	rep, err := Analyze(dir, DefaultLimit)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	langs := map[string]int{}
	for _, l := range rep.Languages {
		langs[l.Ext] = l.Count
	}
	if langs["go"] != 1 {
		t.Errorf("languages[go] = %d, want 1: %+v", langs["go"], rep.Languages)
	}
	if langs["md"] != 1 {
		t.Errorf("languages[md] = %d, want 1: %+v", langs["md"], rep.Languages)
	}
	if langs["py"] != 1 {
		t.Errorf("languages[py] = %d, want 1: %+v", langs["py"], rep.Languages)
	}

	wantEntry := "cmd/app"
	found := false
	for _, e := range rep.EntryPoints {
		if e == wantEntry {
			found = true
		}
	}
	if !found {
		t.Errorf("EntryPoints = %v, want to contain %q", rep.EntryPoints, wantEntry)
	}

	if len(rep.MigrateNotes) != 1 {
		t.Fatalf("MigrateNotes = %d, want 1: %+v", len(rep.MigrateNotes), rep.MigrateNotes)
	}
	if rep.MigrateNotes[0].Text != "legacy auth lives in pkg/auth; see CBIN-105" {
		t.Errorf("MigrateNotes[0].Text = %q", rep.MigrateNotes[0].Text)
	}
	if rep.MigrateNotes[0].File != "docs/NOTES.md" {
		t.Errorf("MigrateNotes[0].File = %q, want docs/NOTES.md", rep.MigrateNotes[0].File)
	}
	if rep.MigrateNotesTotal != 1 {
		t.Errorf("MigrateNotesTotal = %d, want 1", rep.MigrateNotesTotal)
	}

	// No .canary/specs anywhere under dir: GenerateNextID must fail
	// gracefully and onboard must fall back to the flat "<KEY>-001" form.
	if rep.ProjectKey != "CBIN" {
		t.Errorf("ProjectKey = %q, want CBIN (no .canary/project.yaml present)", rep.ProjectKey)
	}
	if rep.NextID != "CBIN-001" {
		t.Errorf("NextID = %q, want CBIN-001", rep.NextID)
	}

	if len(rep.Sources) != 1 || rep.Sources[0].Key != "CBIN" {
		t.Errorf("Sources = %+v, want single CBIN flatfile source", rep.Sources)
	}

	if len(rep.NextSteps) == 0 {
		t.Error("NextSteps is empty")
	}
	joined := strings.Join(rep.NextSteps, " | ")
	if !strings.Contains(joined, "canary init") {
		t.Errorf("NextSteps = %v, want a step mentioning `canary init` (no .canary/ present)", rep.NextSteps)
	}
}

func TestCANARY_CBIN_303_Analyze_LimitBounds(t *testing.T) {
	dir := polyglotTree(t)

	rep, err := Analyze(dir, 1)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(rep.Languages) != 1 {
		t.Errorf("Languages len = %d, want 1 (limit=1)", len(rep.Languages))
	}
	if rep.LanguagesTotal < len(rep.Languages) {
		t.Errorf("LanguagesTotal = %d < len(Languages) = %d", rep.LanguagesTotal, len(rep.Languages))
	}
}

func TestCANARY_CBIN_303_Analyze_DefaultRootAndLimit(t *testing.T) {
	dir := polyglotTree(t)
	rep, err := Analyze(dir, 0) // 0 must fall back to DefaultLimit
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(rep.Languages) == 0 {
		t.Error("Languages is empty")
	}
}

func TestCANARY_CBIN_303_CreateOnboardCommand_JSON(t *testing.T) {
	dir := polyglotTree(t)
	cmd := CreateOnboardCommand()
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--root", dir, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var rep OnboardReport
	if err := json.Unmarshal([]byte(buf.String()), &rep); err != nil {
		t.Fatalf("json.Unmarshal: %v\noutput: %s", err, buf.String())
	}
	if rep.NextID != "CBIN-001" {
		t.Errorf("NextID = %q, want CBIN-001", rep.NextID)
	}
	if strings.Contains(buf.String(), "\n\n") {
		t.Errorf("--json output should be a single compact line, got: %q", buf.String())
	}
}

func TestCANARY_CBIN_303_CreateOnboardCommand_Human(t *testing.T) {
	dir := polyglotTree(t)
	cmd := CreateOnboardCommand()
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "cmd/app") {
		t.Errorf("human output missing entry point, got:\n%s", out)
	}
	if !strings.Contains(out, "CBIN-001") {
		t.Errorf("human output missing NextID, got:\n%s", out)
	}
	lines := strings.Count(out, "\n")
	if lines > 30 {
		t.Errorf("human output has %d lines, want <= 30 for this small tree:\n%s", lines, out)
	}
}
