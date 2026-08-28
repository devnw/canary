package canaryscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCANARY_CBIN_201_UpdateStaleTicketIDs(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-28T00:00:00Z")
	root := t.TempDir()
	src := "package x\n" +
		"// CANARY: REQ=PLAT-4521; FEATURE=\"Ingest\"; ASPECT=API; STATUS=TESTED; TEST=TestIngest; UPDATED=2020-01-01\n"
	path := filepath.Join(root, "x.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	diags := []string{"CANARY_STALE REQ=PLAT-4521 updated=2020-01-01 age_days=2431 threshold=30"}
	if _, err := UpdateStaleTokens(root, DefaultSkipRegex(), diags); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if strings.Contains(string(b), "UPDATED=2020-01-01") {
		t.Errorf("stale ticket-sourced token was not updated:\n%s", b)
	}
	if !strings.Contains(string(b), "UPDATED=2026-08-28") {
		t.Errorf("UPDATED not rewritten to test timestamp:\n%s", b)
	}
}
