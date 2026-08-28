package main

import (
	"os"
	"testing"

	"go.devnw.com/canary/pkg/canaryscan"
)

// CANARY: REQ=CBIN-211; FEATURE="BugOpenStatusTest"; ASPECT=Engine; STATUS=TESTED; TEST=TestBUGOpenStatus; OWNER=canary; UPDATED=2025-11-02

func TestBUGOpenStatus(t *testing.T) {
	dir := t.TempDir()
	content := `// CANARY: BUG-7; FEATURE="OpenBug"; ASPECT=Engine; STATUS=OPEN; UPDATED=2025-10-18`
	path := dir + "/bug.go"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	rep, err := canaryscan.Scan(dir, canaryscan.DefaultSkipRegex(), nil, nil)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	found := false
	for _, r := range rep.Requirements {
		if r.ID == "BUG-007" { // normalized
			for _, f := range r.Features {
				if f.Feature == "OpenBug" && f.Status == "OPEN" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("BUG-007 OpenBug feature with OPEN status not found: %+v", rep.Requirements)
	}
	if rep.Summary.ByStatus["OPEN"] != 1 {
		t.Fatalf("expected OPEN count=1 got %d", rep.Summary.ByStatus["OPEN"])
	}
	// Basic sanity check of normalized ID length
	if len("BUG-007") != 7 {
		t.Fatalf("unexpected BUG ID length")
	}
}
