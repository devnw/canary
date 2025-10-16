// Copyright (c) 2024 by CodePros.
//
// This software is proprietary information of CodePros.
// Unauthorized use, copying, modification, distribution, and/or
// disclosure is strictly prohibited, except as provided under the terms
// of the commercial license agreement you have entered into with
// CodePros.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact CodePros at info@codepros.org.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func build(t *testing.T) string {
	t.Helper()
	exe := filepath.Join("bin", "canary_test_build")
	// Build from repo root (tests run inside tools/canary/internal package directory)
	cmd := exec.Command("go", "build", "-o", exe, "./tools/canary")
	cmd.Dir = filepath.Clean("../../..") // go up from tools/canary/internal to repo root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	if _, statErr := os.Stat(filepath.Join(cmd.Dir, exe)); statErr != nil {
		t.Fatalf("built binary missing: %v", statErr)
	}
	return filepath.Join(cmd.Dir, exe)
}

type runResult struct {
	code           int
	stdout, stderr string
}

func run(exe string, args ...string) runResult {
	c := exec.Command(exe, args...)
	var stdout, stderr strings.Builder
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = -1
		}
	}
	return runResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func TestAcceptance_FixtureSummary(t *testing.T) {
	exe := build(t)
	root := filepath.Join("tools", "canary", "testdata", "summary")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(root, "a.txt"), []byte("CANARY: REQ=CBIN-10; FEATURE=\"Foo\"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20\n"), 0o644)

	os.WriteFile(filepath.Join(root, "b.txt"), []byte("CANARY: REQ=CBIN-11; FEATURE=\"Bar\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-09-20\n"), 0o644)
	res := run(exe, "--root", root, "--out", filepath.Join(root, "status.json"))
	if res.code != 0 {
		t.Fatalf("exit=%d stderr=%s", res.code, res.stderr)
	}
	b, _ := os.ReadFile(filepath.Join(root, "status.json"))
	var parsed struct {
		Summary struct {
			ByStatus map[string]int `json:"by_status"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Summary.ByStatus["IMPL"] != 1 || parsed.Summary.ByStatus["STUB"] != 1 {
		t.Fatalf("bad counts: %+v", parsed.Summary.ByStatus)
	}
	fmt.Println("{\"summary\":{\"by_status\":{\"IMPL\":1,\"STUB\":1}}}")
}

func TestAcceptance_Overclaim(t *testing.T) {
	exe := build(t)
	root := filepath.Join("tools", "canary", "testdata", "overclaim")

	os.MkdirAll(root, 0o755)

	os.WriteFile(filepath.Join(root, "t.txt"), []byte("CANARY: REQ=CBIN-042; FEATURE=\"X\"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20\n"), 0o644)
	gap := filepath.Join(root, "GAP_ANALYSIS.md")

	os.WriteFile(gap, []byte("✅ CBIN-042\n"), 0o644)
	res := run(exe, "--root", root, "--verify", gap, "--out", filepath.Join(root, "status.json"))
	if res.code != 2 {
		t.Fatalf("expected 2 got %d stderr=%s", res.code, res.stderr)
	}
	if !strings.Contains(res.stderr, "CANARY_VERIFY_FAIL REQ=CBIN-042") {
		t.Fatalf("missing diag: %s", res.stderr)
	}
	fmt.Println("ACCEPT Overclaim Exit=2")
}

func TestAcceptance_Stale(t *testing.T) {
	exe := build(t)
	root := filepath.Join("tools", "canary", "testdata", "stale")

	os.MkdirAll(root, 0o755)

	os.WriteFile(filepath.Join(root, "t.txt"), []byte("CANARY: REQ=CBIN-051; FEATURE=\"Y\"; ASPECT=API; STATUS=TESTED; UPDATED=2024-01-01\n"), 0o644)
	res := run(exe, "--root", root, "--strict", "--out", filepath.Join(root, "status.json"))
	if res.code != 2 {
		t.Fatalf("expected 2 got %d stderr=%s", res.code, res.stderr)
	}
	if !strings.Contains(res.stderr, "CANARY_STALE REQ=CBIN-051") {
		t.Fatalf("missing stale diag: %s", res.stderr)
	}
	fmt.Println("ACCEPT Stale Exit=2")
}

func TestAcceptance_SelfCanary(t *testing.T) {
	exe := build(t)
	// derive repo root via caller file path for robustness regardless of test working dir
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../.."))
	gap := filepath.Join(repoRoot, "GAP_ANALYSIS.md")
	if _, err := os.Stat(gap); err != nil {
		if err := os.WriteFile(gap, []byte("# Requirements Gap Analysis (Self)\n✅ CBIN-101\n✅ CBIN-102\n"), 0o644); err != nil {
			t.Fatalf("create GAP_ANALYSIS.md: %v", err)
		}
	}
	canaryRoot := filepath.Join(repoRoot, "tools", "canary")
	skipPattern := `(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|testdata|internal)(/|$)`
	res1 := run(exe, "--root", canaryRoot, "--out", "status.json", "--csv", "status.csv", "--skip", skipPattern)
	if res1.code != 0 {
		t.Fatalf("scan exit=%d stderr=%s root=%s", res1.code, res1.stderr, canaryRoot)
	}
	res2 := run(exe, "--root", canaryRoot, "--verify", gap, "--strict", "--out", "status.json", "--skip", skipPattern)
	if res2.code != 0 {
		t.Fatalf("verify exit=%d stderr=%s", res2.code, res2.stderr)
	}
	fmt.Println("ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]")
	_ = os.Remove("status.json")
	_ = os.Remove("status.csv")
}

func TestAcceptance_CSVOrder(t *testing.T) {
	exe := build(t)
	root := filepath.Join("tools", "canary", "testdata", "csvorder")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create fixture with multiple requirements and features in non-alphabetical order
	fixtures := map[string]string{
		"z.txt": "CANARY: REQ=CBIN-999; FEATURE=\"Zulu\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",
		"a.txt": "CANARY: REQ=CBIN-101; FEATURE=\"Alpha\"; ASPECT=CLI; STATUS=TESTED; TEST=TestAlpha; UPDATED=2025-10-15\n",
		"m.txt": "CANARY: REQ=CBIN-500; FEATURE=\"Mike\"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-15\n",
		"b.txt": "CANARY: REQ=CBIN-102; FEATURE=\"Bravo\"; ASPECT=API; STATUS=BENCHED; TEST=TestBravo; BENCH=BenchBravo; UPDATED=2025-10-15\n",
		"c.txt": "CANARY: REQ=CBIN-102; FEATURE=\"Charlie\"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-15\n", // Same REQ as Bravo
	}

	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Run scanner first time
	csv1Path := filepath.Join(root, "status1.csv")
	res1 := run(exe, "--root", root, "--out", filepath.Join(root, "status1.json"), "--csv", csv1Path)
	if res1.code != 0 {
		t.Fatalf("first scan exit=%d stderr=%s", res1.code, res1.stderr)
	}

	csv1, err := os.ReadFile(csv1Path)
	if err != nil {
		t.Fatalf("read csv1: %v", err)
	}

	// Run scanner second time
	csv2Path := filepath.Join(root, "status2.csv")
	res2 := run(exe, "--root", root, "--out", filepath.Join(root, "status2.json"), "--csv", csv2Path)
	if res2.code != 0 {
		t.Fatalf("second scan exit=%d stderr=%s", res2.code, res2.stderr)
	}

	csv2, err := os.ReadFile(csv2Path)
	if err != nil {
		t.Fatalf("read csv2: %v", err)
	}

	// Verify outputs are identical
	if string(csv1) != string(csv2) {
		t.Fatalf("CSV output not deterministic:\nRun1:\n%s\n\nRun2:\n%s", csv1, csv2)
	}

	// Verify rows are sorted by REQ ID
	lines := strings.Split(strings.TrimSpace(string(csv1)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least header + 1 row, got %d lines", len(lines))
	}

	// Skip header
	var prevReq string
	for i, line := range lines[1:] {
		fields := strings.Split(line, ",")
		if len(fields) < 1 {
			continue
		}
		req := fields[0]
		if prevReq != "" && req < prevReq {
			t.Errorf("CSV not sorted by REQ: line %d has %s after %s", i+2, req, prevReq)
		}
		prevReq = req
	}

	fmt.Println("ACCEPT CSVOrder deterministic and sorted")
}

func TestAcceptance_SkipEdgeCases(t *testing.T) {
	exe := build(t)
	root := filepath.Join("tools", "canary", "testdata", "skipedges")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create various edge case files
	fixtures := map[string]string{
		// Normal files that should be scanned
		"normal.go":      "// CANARY: REQ=CBIN-001; FEATURE=\"Normal\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",
		"subdir/file.go": "// CANARY: REQ=CBIN-002; FEATURE=\"Subdir\"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-15\n",

		// Files that should be skipped
		".hidden":             "// CANARY: REQ=CBIN-099; FEATURE=\"Hidden\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",
		"node_modules/pkg.js": "// CANARY: REQ=CBIN-098; FEATURE=\"NodeModules\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",
		"vendor/lib.go":       "// CANARY: REQ=CBIN-097; FEATURE=\"Vendor\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",
		".git/config":         "// CANARY: REQ=CBIN-096; FEATURE=\"Git\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",

		// Edge cases
		"file with spaces.go": "// CANARY: REQ=CBIN-003; FEATURE=\"Spaces\"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-15\n",
		// Unicode filename (using Chinese characters)
		"测试.go": "// CANARY: REQ=CBIN-004; FEATURE=\"Unicode\"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-15\n",
	}

	// Create fixture files
	for name, content := range fixtures {
		path := filepath.Join(root, name)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Test 1: Scan with default skip pattern (should skip .git, node_modules, vendor, hidden files)
	skipPattern := `(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build)(/|$)|^\.|/\.`
	res1 := run(exe, "--root", root, "--out", filepath.Join(root, "status1.json"), "--skip", skipPattern)
	if res1.code != 0 {
		t.Fatalf("scan with skip pattern exit=%d stderr=%s", res1.code, res1.stderr)
	}

	// Parse output to verify expected tokens found
	statusBytes, err := os.ReadFile(filepath.Join(root, "status1.json"))
	if err != nil {
		t.Fatalf("read status.json: %v", err)
	}

	var status struct {
		Requirements []struct {
			ID string `json:"id"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal(statusBytes, &status); err != nil {
		t.Fatalf("unmarshal status: %v", err)
	}

	// Verify we found the expected tokens (CBIN-001, CBIN-002, CBIN-003, CBIN-004)
	// and NOT the skipped ones (CBIN-096, CBIN-097, CBIN-098, CBIN-099)
	foundIDs := make(map[string]bool)
	for _, req := range status.Requirements {
		foundIDs[req.ID] = true
	}

	expectedFound := []string{"CBIN-001", "CBIN-002", "CBIN-003", "CBIN-004"}
	expectedSkipped := []string{"CBIN-096", "CBIN-097", "CBIN-098", "CBIN-099"}

	for _, id := range expectedFound {
		if !foundIDs[id] {
			t.Errorf("expected to find %s, but it was not scanned", id)
		}
	}

	for _, id := range expectedSkipped {
		if foundIDs[id] {
			t.Errorf("expected to skip %s, but it was scanned", id)
		}
	}

	// Test 2: Scan with no skip pattern (should find all tokens)
	res2 := run(exe, "--root", root, "--out", filepath.Join(root, "status2.json"))
	if res2.code != 0 {
		t.Fatalf("scan without skip exit=%d stderr=%s", res2.code, res2.stderr)
	}

	statusBytes2, err := os.ReadFile(filepath.Join(root, "status2.json"))
	if err != nil {
		t.Fatalf("read status2.json: %v", err)
	}

	var status2 struct {
		Requirements []struct {
			ID string `json:"id"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal(statusBytes2, &status2); err != nil {
		t.Fatalf("unmarshal status2: %v", err)
	}

	// Should find at least the normal files (hidden files might be skipped by filesystem walk)
	// Expecting at least: CBIN-001, CBIN-002, CBIN-003, CBIN-004, CBIN-096, CBIN-097, CBIN-098, CBIN-099
	foundIDs2 := make(map[string]bool)
	for _, req := range status2.Requirements {
		foundIDs2[req.ID] = true
	}

	// Without skip pattern, should find more than with skip pattern
	if len(status2.Requirements) <= len(status.Requirements) {
		t.Errorf("without skip pattern, expected more requirements than with skip (%d), got %d",
			len(status.Requirements), len(status2.Requirements))
	}

	// Verify that at least the normal files are found
	normalFiles := []string{"CBIN-001", "CBIN-002", "CBIN-003", "CBIN-004"}
	for _, id := range normalFiles {
		if !foundIDs2[id] {
			t.Errorf("expected to find %s without skip pattern", id)
		}
	}

	fmt.Println("ACCEPT SkipEdgeCases patterns work correctly")
}

func TestAcceptance_UpdateStale(t *testing.T) {
	exe := build(t)
	root := filepath.Join("tools", "canary", "testdata", "updatestale")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create test file with stale and fresh tokens
	testFile := filepath.Join(root, "test.go")
	content := `package test
// CANARY: REQ=CBIN-001; FEATURE="StaleToken"; ASPECT=API; STATUS=TESTED; TEST=Test1; UPDATED=2024-01-01
// CANARY: REQ=CBIN-002; FEATURE="FreshToken"; ASPECT=CLI; STATUS=TESTED; TEST=Test2; UPDATED=2025-10-15
// CANARY: REQ=CBIN-003; FEATURE="StaleImplNotUpdated"; ASPECT=Engine; STATUS=IMPL; UPDATED=2024-01-01
// CANARY: REQ=CBIN-004; FEATURE="StaleBenchedToken"; ASPECT=Storage; STATUS=BENCHED; BENCH=Bench4; UPDATED=2024-01-01
`
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run with --update-stale
	res := run(exe, "--root", root, "--update-stale", "--out", filepath.Join(root, "status.json"))
	if res.code != 0 {
		t.Fatalf("update-stale exit=%d stderr=%s", res.code, res.stderr)
	}

	// Verify file was updated
	updatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("read updated file: %v", err)
	}

	// Parse lines to verify each token
	lines := strings.Split(string(updatedContent), "\n")

	// Helper to find line containing REQ ID
	findLine := func(reqID string) string {
		for _, line := range lines {
			if strings.Contains(line, "REQ="+reqID) {
				return line
			}
		}
		return ""
	}

	// CBIN-001 (TESTED, stale) should be updated
	line001 := findLine("CBIN-001")
	if line001 == "" {
		t.Error("CBIN-001 missing from updated file")
	} else if strings.Contains(line001, "UPDATED=2024-01-01") {
		t.Errorf("CBIN-001 should have UPDATED field changed from 2024-01-01, got: %s", line001)
	}

	// CBIN-002 (TESTED, fresh) should NOT be updated
	line002 := findLine("CBIN-002")
	if line002 == "" {
		t.Error("CBIN-002 missing from updated file")
	} else if !strings.Contains(line002, "UPDATED=2025-10-15") {
		t.Errorf("CBIN-002 fresh token should remain UPDATED=2025-10-15, got: %s", line002)
	}

	// CBIN-003 (IMPL, stale) should NOT be updated (only TESTED/BENCHED are updated)
	line003 := findLine("CBIN-003")
	if line003 == "" {
		t.Error("CBIN-003 missing from updated file")
	} else if !strings.Contains(line003, "UPDATED=2024-01-01") {
		t.Errorf("CBIN-003 IMPL token should remain UPDATED=2024-01-01, got: %s", line003)
	}

	// CBIN-004 (BENCHED, stale) should be updated
	line004 := findLine("CBIN-004")
	if line004 == "" {
		t.Error("CBIN-004 missing from updated file")
	} else if strings.Contains(line004, "UPDATED=2024-01-01") {
		t.Errorf("CBIN-004 should have UPDATED field changed from 2024-01-01, got: %s", line004)
	}

	// Verify stderr message
	if !strings.Contains(res.stderr, "Updated") || !strings.Contains(res.stderr, "stale tokens") {
		t.Errorf("expected update message in stderr, got: %s", res.stderr)
	}

	fmt.Println("ACCEPT UpdateStale rewrites stale TESTED/BENCHED tokens")
}

func TestMetadata(t *testing.T) {
	t.Logf("go=%s os=%s arch=%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
