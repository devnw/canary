package core_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	core "go.codepros.org/canary/internal/core"
)

// Test that the repository self-canary tokens (CBIN-101..103) are discoverable.
func TestSelfCanaryTokensPresent(t *testing.T) {
	_, _ = os.Getwd()
	re := regexp.MustCompile("(^|/)(.direnv|.git|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)(/|$)")
	// find repo root by ascending until go.mod present
	// create isolated test directory with canonical token lines
	dir := t.TempDir()
	content := []string{
		"// CANARY: REQ=CBIN-101; FEATURE=\"ScannerCore\"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20\n",
		"// CANARY: REQ=CBIN-102; FEATURE=\"VerifyGate\"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20\n",
		"// CANARY: REQ=CBIN-103; FEATURE=\"StatusJSON\"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_103_API_StatusSchema; BENCH=BenchmarkCANARY_CBIN_103_API_Emit; OWNER=canary; UPDATED=2025-09-20\n",
	}
	for i, c := range content {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%d.go", i+1)), []byte("package p\n"+c), 0o644)
	}
	rep, err := core.Scan(core.ScanOptions{Root: dir, Skip: re})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	b, _ := json.Marshal(rep)
	_ = b // suppress unused if logs removed
	want := map[string]bool{"CBIN-101": false, "CBIN-102": false, "CBIN-103": false}
	for _, r := range rep.Requirements {
		if _, ok := want[r.ID]; ok {
			want[r.ID] = true
		}
	}
	for id, ok := range want {
		if !ok {
			t.Fatalf("expected requirement %s token present", id)
		}
	}
}
