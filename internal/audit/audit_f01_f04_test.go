// Package audit holds the F-01–F-27 audit regression tests: one test per
// audit finding, asserting the remediated behavior end to end.
package audit

import (
	"os"
	"path/filepath"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
)

func TestAuditF01(t *testing.T) { // lexical TEST= must not promote
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=IMPL; TEST=TestDoesNotExist; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := canaryscan.Scan(dir, canaryscan.DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := rep.Requirements[0].Features[0].Status
	if got != "IMPL" {
		t.Fatalf("status promoted lexically: got %q want IMPL", got)
	}
}

func TestAuditF04(t *testing.T) { // binary + oversize become issues, not silence/abort
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bin.dat"), append([]byte("CANARY: x"), 0x00, 0x01), 0o644); err != nil {
		t.Fatal(err)
	}
	big, err := os.Create(filepath.Join(dir, "big.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := big.Truncate(canaryscan.MaxFileBytes + 1); err != nil {
		_ = big.Close()
		t.Fatal(err)
	}
	if err := big.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ok.go"), []byte(
		"// CANARY: REQ=CBIN-002; FEATURE=\"G\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := canaryscan.Scan(dir, canaryscan.DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Issues) != 2 {
		t.Fatalf("want 2 issues, got %+v", rep.Issues)
	}
	if len(rep.Requirements) != 1 {
		t.Fatalf("good file must still be scanned")
	}
}

// TestAuditF15 covers the scan-level half of F-15: the canonical
// ParseTokenLine/SerializeToken pair is the single token grammar, and it
// round-trips values that the legacy ";"-split parser could not represent.
func TestAuditF15(t *testing.T) {
	fields := []canaryscan.Field{
		{Key: "REQ", Value: "CBIN-003"},
		{Key: "FEATURE", Value: `semi; and "quote"`},
		{Key: "ASPECT", Value: "Engine"},
		{Key: "STATUS", Value: "IMPL"},
		{Key: "UPDATED", Value: "2026-01-01"},
	}
	line, err := canaryscan.SerializeToken(fields)
	if err != nil {
		t.Fatalf("SerializeToken: %v", err)
	}
	got, ok, err := canaryscan.ParseTokenLine(line)
	if err != nil || !ok {
		t.Fatalf("ParseTokenLine(%q): ok=%v err=%v", line, ok, err)
	}
	if len(got) != len(fields) {
		t.Fatalf("round-trip field count = %d, want %d", len(got), len(fields))
	}
	for i := range fields {
		if got[i] != fields[i] {
			t.Errorf("field %d = %+v, want %+v", i, got[i], fields[i])
		}
	}
}
