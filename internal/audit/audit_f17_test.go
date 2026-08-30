// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
)

// runStdout runs the canary binary in root and returns STDOUT ONLY. stderr is
// captured separately and surfaced only on failure: the whole point of a
// machine format is that a program can read stdout without filtering, so a
// test for that property must never fold stderr into what it decodes.
func runStdout(t *testing.T, root, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...) //nolint:gosec // test-built binary
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("canary %s: %v\nstdout: %s\nstderr: %s", strings.Join(args, " "), err, out, stderr.String())
	}
	return string(out)
}

// reportWithFeature builds a one-feature scan report whose feature name and
// owner are caller-supplied, so a test can inject a formula-shaped value and a
// comma-and-quote value into the cells WriteCSV must neutralize and quote.
func reportWithFeature(t *testing.T, feature, owner string) canaryscan.Report {
	t.Helper()
	return canaryscan.Report{
		Requirements: []canaryscan.Requirement{
			{
				ID: "CBIN-001",
				Features: []canaryscan.Feature{
					{
						Feature: feature,
						Aspect:  "API",
						Status:  "IMPL",
						Files:   []string{"a.go"},
						Tests:   []string{"TestA"},
						Owner:   owner,
						Updated: "2026-01-01",
					},
				},
			},
		},
	}
}

// indexedEmptyRepo returns a git repository with a built but empty token index,
// so `doc report` has a database to open yet zero requirements to divide by --
// the exact condition that used to make coverage_percent NaN.
func indexedEmptyRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	run(t, root, buildCanary(t), "index")
	return root
}

// specsRepo returns a repository holding one specification directory, so
// `specs --json` has something to marshal.
func specsRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".canary", "specs", "CBIN-001-example-feature")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestAuditF17 covers F-17: output must be produced by encoding/csv and
// encoding/json rather than hand-rolled formatting, must defuse CSV formula
// injection, must fail loudly on a write it cannot complete, and must not emit
// NaN for a zero-requirement coverage ratio.
func TestAuditF17(t *testing.T) {
	// 1. CSV: injection guard + independent decoder round-trip
	rep := reportWithFeature(t, "=HYPERLINK(\"x\")", "va,l\"ue")
	p := filepath.Join(t.TempDir(), "out.csv")
	if err := canaryscan.WriteCSV(p, rep); err != nil {
		t.Fatal(err)
	}
	f, _ := os.Open(p)                      //nolint:gosec // test-controlled path
	rows, err := csv.NewReader(f).ReadAll() // independent decoder accepts output
	if err != nil {
		t.Fatalf("csv not standards-decodable: %v", err)
	}
	for _, row := range rows[1:] {
		for _, cell := range row {
			if len(cell) > 0 && strings.ContainsRune("=+-@", rune(cell[0])) {
				t.Fatalf("unguarded formula cell %q", cell)
			}
		}
	}
	// 2. short write fails: write to a closed file / full pipe
	if err := canaryscan.WriteCSV(filepath.Join(t.TempDir(), "no", "such", "dir.csv"), rep); err == nil {
		t.Fatal("write error swallowed")
	}
	// 3. zero-doc metrics equal numeric 0
	out := runStdout(t, indexedEmptyRepo(t), buildCanary(t), "doc", "report", "--format", "json")
	var rpt map[string]any
	if err := json.Unmarshal([]byte(out), &rpt); err != nil {
		t.Fatalf("doc report json invalid: %v", err)
	}
	if v, ok := rpt["coverage_percent"].(float64); !ok || v != 0 {
		t.Fatalf("coverage_percent = %v", rpt["coverage_percent"])
	}
	// 4. specs --json decodable
	out2 := runStdout(t, specsRepo(t), buildCanary(t), "specs", "--json")
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out2), &arr); err != nil {
		t.Fatalf("specs json invalid: %v", err)
	}
}
