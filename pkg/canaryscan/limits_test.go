package canaryscan

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hasIssue reports whether issues contains one for path with the given reason.
func hasIssue(issues []ScanIssue, path, reason string) bool {
	for _, i := range issues {
		if i.Reason == reason && filepath.Base(i.Path) == path {
			return true
		}
	}
	return false
}

// TestScanSkipsBinary proves a file carrying a NUL byte in its first 8 KiB is
// recorded as a "binary" issue and skipped rather than parsed or fatal.
func TestScanSkipsBinary(t *testing.T) {
	root := t.TempDir()
	bin := append([]byte("CANARY: not-a-real-token"), 0x00, 0x01, 0x02)
	if err := os.WriteFile(filepath.Join(root, "bin.dat"), bin, 0o600); err != nil {
		t.Fatal(err)
	}
	good := "// CANARY: REQ=CBIN-002; FEATURE=\"G\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte(good), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !hasIssue(rep.Issues, "bin.dat", "binary") {
		t.Errorf("no binary issue for bin.dat, got %+v", rep.Issues)
	}
	if len(rep.Requirements) != 1 {
		t.Fatalf("good file must still be scanned, got %d requirements", len(rep.Requirements))
	}
}

// TestScanSkipsOversizedFile proves a file larger than MaxFileBytes is
// recorded as a "file_too_large" issue and skipped.
func TestScanSkipsOversizedFile(t *testing.T) {
	root := t.TempDir()
	f, err := os.Create(filepath.Join(root, "big.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(MaxFileBytes + 1); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !hasIssue(rep.Issues, "big.go", "file_too_large") {
		t.Errorf("no file_too_large issue for big.go, got %+v", rep.Issues)
	}
}

// TestScanSkipsOversizedLine proves a single logical line longer than
// MaxLineBytes is recorded as a "line_too_large" issue rather than buffered.
func TestScanSkipsOversizedLine(t *testing.T) {
	root := t.TempDir()
	content := strings.Repeat("a", MaxLineBytes+1) + "\n"
	if err := os.WriteFile(filepath.Join(root, "long.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !hasIssue(rep.Issues, "long.go", "line_too_large") {
		t.Errorf("no line_too_large issue for long.go, got %+v", rep.Issues)
	}
}

// TestScanReportsReadError proves an unreadable file becomes a "read_error"
// issue instead of aborting the whole scan.
func TestScanReportsReadError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not deny reads")
	}
	root := t.TempDir()
	secret := filepath.Join(root, "secret.go")
	if err := os.WriteFile(secret, []byte("// nothing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o600) })
	good := "// CANARY: REQ=CBIN-003; FEATURE=\"H\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte(good), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !hasIssue(rep.Issues, "secret.go", "read_error") {
		t.Errorf("no read_error issue for secret.go, got %+v", rep.Issues)
	}
	if len(rep.Requirements) != 1 {
		t.Fatalf("good file must still be scanned, got %d requirements", len(rep.Requirements))
	}
}

// TestScanReportsParseError proves a malformed token becomes a "parse_error"
// issue; the rest of the file — and the rest of the tree — still scans.
func TestScanReportsParseError(t *testing.T) {
	root := t.TempDir()
	bad := "// CANARY: REQ=CBIN-004; FEATURE=\"unterminated; ASPECT=API\n" +
		"// CANARY: REQ=CBIN-005; FEATURE=\"After\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !hasIssue(rep.Issues, "bad.go", "parse_error") {
		t.Errorf("no parse_error issue for bad.go, got %+v", rep.Issues)
	}
	if len(rep.Requirements) != 1 || rep.Requirements[0].ID != "CBIN-005" {
		t.Fatalf("token after the bad one must still scan, got %+v", rep.Requirements)
	}
}

// TestScanNoPromotion proves STATUS is a pure declaration: a TEST= field no
// longer promotes IMPL to TESTED, and a BENCH= field no longer promotes to
// BENCHED. Both are still recorded as evidence references.
func TestScanNoPromotion(t *testing.T) {
	cases := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "test does not promote IMPL",
			token: "// CANARY: REQ=CBIN-010; FEATURE=\"A\"; ASPECT=API; STATUS=IMPL; TEST=TestX; UPDATED=2026-01-01\n",
			want:  "IMPL",
		},
		{
			name:  "bench does not promote IMPL",
			token: "// CANARY: REQ=CBIN-011; FEATURE=\"B\"; ASPECT=API; STATUS=IMPL; BENCH=BenchX; UPDATED=2026-01-01\n",
			want:  "IMPL",
		},
		{
			name:  "bench does not promote TESTED",
			token: "// CANARY: REQ=CBIN-012; FEATURE=\"C\"; ASPECT=API; STATUS=TESTED; BENCH=BenchX; UPDATED=2026-01-01\n",
			want:  "TESTED",
		},
		{
			name:  "STUB stays STUB",
			token: "// CANARY: REQ=CBIN-013; FEATURE=\"D\"; ASPECT=API; STATUS=STUB; TEST=TestX; UPDATED=2026-01-01\n",
			want:  "STUB",
		},
		{
			name:  "FIXED is not rewritten to REMOVED",
			token: "// CANARY: REQ=CBIN-014; FEATURE=\"E\"; ASPECT=API; STATUS=FIXED; UPDATED=2026-01-01\n",
			want:  "FIXED",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(tc.token), 0o600); err != nil {
				t.Fatal(err)
			}
			rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if len(rep.Requirements) != 1 || len(rep.Requirements[0].Features) != 1 {
				t.Fatalf("want exactly one requirement/feature, got %+v", rep.Requirements)
			}
			if got := rep.Requirements[0].Features[0].Status; got != tc.want {
				t.Errorf("status = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestScanRecordsEvidenceWithoutPromoting proves TEST=/BENCH= are still
// captured on the Feature even though they no longer affect STATUS.
func TestScanRecordsEvidenceWithoutPromoting(t *testing.T) {
	root := t.TempDir()
	token := "// CANARY: REQ=CBIN-020; FEATURE=\"E\"; ASPECT=API; STATUS=IMPL; TEST=TestA,TestB; BENCH=BenchA; UPDATED=2026-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(token), 0o600); err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	f := rep.Requirements[0].Features[0]
	if f.Status != "IMPL" {
		t.Errorf("status = %q, want IMPL", f.Status)
	}
	if len(f.Tests) != 2 {
		t.Errorf("tests = %v, want 2 entries", f.Tests)
	}
	if len(f.Benches) != 1 {
		t.Errorf("benches = %v, want 1 entry", f.Benches)
	}
}

// TestRunReportsIssues proves Run surfaces every scan issue on stderr, and
// that --strict turns an incomplete scan into a failing exit.
func TestRunReportsIssues(t *testing.T) {
	newRoot := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		bin := append([]byte("CANARY: not-a-real-token"), 0x00)
		if err := os.WriteFile(filepath.Join(root, "bin.dat"), bin, 0o600); err != nil {
			t.Fatal(err)
		}
		token := "// CANARY: REQ=CBIN-030; FEATURE=\"R\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"
		if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte(token), 0o600); err != nil {
			t.Fatal(err)
		}
		return root
	}

	t.Run("issues reported but not fatal", func(t *testing.T) {
		root := newRoot(t)
		var stdout, stderr bytes.Buffer
		code := Run(Config{Root: root, Out: filepath.Join(root, "status.json")}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (issues alone are not fatal)", code)
		}
		if !strings.Contains(stderr.String(), "CANARY_SCAN_ISSUE path=") ||
			!strings.Contains(stderr.String(), "reason=binary") {
			t.Errorf("stderr missing CANARY_SCAN_ISSUE line: %q", stderr.String())
		}
	})

	t.Run("strict fails on incomplete scan", func(t *testing.T) {
		root := newRoot(t)
		var stdout, stderr bytes.Buffer
		code := Run(Config{Root: root, Out: filepath.Join(root, "status.json"), Strict: true}, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("exit code = %d, want 2", code)
		}
		if !strings.Contains(stderr.String(), "SCAN_INCOMPLETE path=") {
			t.Errorf("stderr missing SCAN_INCOMPLETE diag: %q", stderr.String())
		}
		if !strings.Contains(stdout.String(), "CANARY_VERIFY_FAIL count=1") {
			t.Errorf("stdout missing CANARY_VERIFY_FAIL count=1: %q", stdout.String())
		}
	})
}
