package acceptance

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// helper to build binary once per test package.
var binPath string

func build(t *testing.T) string {
	if binPath != "" {
		return binPath
	}
	t.Helper()
	root := findRepoRoot(t)
	binDir := filepath.Join(root, "bin")
	_ = os.MkdirAll(binDir, 0o755)
	outPath := filepath.Join(binDir, "canary")
	cmd := exec.Command("go", "build", "-o", outPath, "./cmd/canary")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(out))
	}
	binPath = outPath
	return binPath
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("pwd: %v", err)
	}
	for i := 0; i < 10; i++ { // walk up to 10 levels
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	t.Fatalf("could not locate repo root with go.mod")
	return ""
}

func run(t *testing.T, args []string, dir string) (stdout, stderr string, exit int) {
	t.Helper()
	cmd := exec.Command(build(t), args...)
	cmd.Dir = dir
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB
	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			exit = 900
		}
	}
	return outB.String(), errB.String(), exit
}

// 1. Fixture parse summary
func TestAcceptance_FixtureSummary(t *testing.T) {
	dir := t.TempDir()
	// create fixture directory with two tokens
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package p\n// CANARY: REQ=CBIN-200; FEATURE=\"Alpha\"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package p\n// CANARY: REQ=CBIN-201; FEATURE=\"Bravo\"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_201_API_Bravo; UPDATED=2025-09-20\n"), 0o644)
	_, _, exit := run(t, []string{"scan", "--root", dir, "--out", filepath.Join(dir, "status.json")}, dir)
	if exit != 0 {
		t.Fatalf("expected exit 0 got %d", exit)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "status.json"))
	// Expect one STUB and one TESTED (promotion occurred)
	if !strings.Contains(string(b), `"STUB":1`) || !strings.Contains(string(b), `"TESTED":1`) {
		t.Fatalf("unexpected status json: %s", string(b))
	}
}

// 2. Overclaim verify fails
func TestAcceptance_Overclaim(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "c.go"), []byte("package p\n// CANARY: REQ=CBIN-042; FEATURE=\"OC\"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20\n"), 0o644)
	gap := "# Gap\n✅ CBIN-042\n"
	os.WriteFile(filepath.Join(dir, "GAP_ANALYSIS.md"), []byte(gap), 0o644)
	_, stderr, exit := run(t, []string{"verify", "--root", dir, "--gap", "GAP_ANALYSIS.md"}, dir)
	if exit != 2 {
		t.Fatalf("expected exit 2 got %d stderr=%s", exit, stderr)
	}
	if !strings.Contains(stderr, "CANARY_VERIFY_FAIL REQ=CBIN-042") {
		t.Fatalf("missing verify fail diagnostic: %s", stderr)
	}
}

// 3. Stale strict
func TestAcceptance_Stale(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "s.go"), []byte("package p\n// CANARY: REQ=CBIN-051; FEATURE=\"Stale\"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_051_API_Stale; UPDATED=2024-01-01\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "GAP_ANALYSIS.md"), []byte("# gap\n"), 0o644)
	// strict verify should flag stale
	_, stderr, exit := run(t, []string{"verify", "--root", dir, "--gap", "GAP_ANALYSIS.md", "--strict"}, dir)
	if exit != 2 {
		t.Fatalf("expected exit 2 got %d stderr=%s", exit, stderr)
	}
	if !strings.Contains(stderr, "CANARY_STALE REQ=CBIN-051") {
		t.Fatalf("missing stale diagnostic: %s", stderr)
	}
}

// 4. Self-scan & self-verify minimal
func TestAcceptance_SelfCanary(t *testing.T) {
	root := findRepoRoot(t)
	gap := []byte("# GAP\n✅ CBIN-101\n✅ CBIN-102\n")
	os.WriteFile(filepath.Join(root, "GAP_SELF.md"), gap, 0o644)
	_, stderr, exit := run(t, []string{"verify", "--root", root, "--gap", "GAP_SELF.md"}, root)
	if exit != 0 {
		t.Fatalf("expected exit 0 got %d stderr=%s", exit, stderr)
	}
}
