package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGenericScanner(t *testing.T) {
	dir := t.TempDir()
	// create sample files
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content+"\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("a.go", `// CANARY: REQ=CBIN-301; FEATURE="GenScan"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_301_API_Gen; UPDATED=2025-11-02`)
	write("b.rs", `// CANARY: REQ=CBIN-301; FEATURE="GenScan"; ASPECT=API; STATUS=IMPL; BENCH=BenchmarkCANARY_CBIN_301_API_Gen; UPDATED=2025-11-02`)
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(res.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(res.Requirements))
	}
	// Promotion: IMPL + test -> TESTED; bench stronger -> BENCHED
	status := res.Requirements[0].Features[0].Status
	foundBenched := false
	for _, f := range res.Requirements[0].Features {
		if f.Status == "BENCHED" {
			foundBenched = true
		}
	}
	if !foundBenched {
		t.Fatalf("expected BENCHED promotion, got status=%s", status)
	}
}

func TestLegacyRequirementID(t *testing.T) {
	dir := t.TempDir()
	content := `// CANARY: REQ-1; FEATURE="Legacy"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-02`
	if err := os.WriteFile(filepath.Join(dir, "legacy.go"), []byte(content+"\n"), 0644); err != nil {
		t.Fatalf("write legacy: %v", err)
	}
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(res.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(res.Requirements))
	}
	if res.Requirements[0].ID != "REQ-001" {
		t.Fatalf("expected padded ID REQ-001, got %s", res.Requirements[0].ID)
	}
}

func TestLegacyPaddedVariants(t *testing.T) {
	dir := t.TempDir()
	lines := []string{
		`// CANARY: REQ-GQL-4; FEATURE="Graph"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-02`,
		`// CANARY: REQ-12; FEATURE="Core"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-02`,
	}
	for i, l := range lines {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%d.go", i)), []byte(l+"\n"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := map[string]bool{"REQ-GQL-004": true, "REQ-012": true}
	for _, r := range res.Requirements {
		delete(want, r.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing padded IDs: %v", want)
	}
}

func TestMixedLegacyModernREQ(t *testing.T) {
	dir := t.TempDir()
	// Same line contains legacy REQ-7 and modern REQ=CBIN-005; the parser should treat key-based REQ as authoritative and ignore ID-only segment (which still sets REQ first but gets overridden by key-value later in parse loop logic).
	line := `// CANARY: REQ-7; REQ=CBIN-5; FEATURE="Mixed"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-02`
	if err := os.WriteFile(filepath.Join(dir, "mixed.go"), []byte(line+"\n"), 0644); err != nil {
		t.Fatalf("write mixed: %v", err)
	}
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(res.Requirements) != 1 {
		t.Fatalf("expected 1 requirement got %d", len(res.Requirements))
	}
	// CBIN-5 should be padded to CBIN-005
	if res.Requirements[0].ID != "CBIN-005" {
		t.Fatalf("expected CBIN-005 got %s", res.Requirements[0].ID)
	}
}

func TestGenericLegacyPrefixes_TASK_BUG(t *testing.T) {
	dir := t.TempDir()
	lines := []string{
		`// CANARY: TASK-2; FEATURE="TaskWork"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-02`,
		`// CANARY: BUG-7; FEATURE="BugFix"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-02`,
	}
	for i, l := range lines {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("x%d.go", i)), []byte(l+"\n"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	// Expect two distinct requirements padded
	want := map[string]bool{"TASK-002": true, "BUG-007": true}
	for _, r := range res.Requirements {
		delete(want, r.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing padded TASK/BUG IDs: %v", want)
	}
}

func TestMigrateLineSkippedByGateScanner(t *testing.T) {
	dir := t.TempDir()
	content := "package x\n// CANARY:MIGRATE free text with = signs; and semicolons\n"
	if err := os.WriteFile(filepath.Join(dir, "migrate.go"), []byte(content), 0644); err != nil {
		t.Fatalf("write migrate.go: %v", err)
	}
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("MIGRATE line must never abort the gate scanner: %v", err)
	}
	if len(res.Requirements) != 0 {
		t.Fatalf("expected 0 requirements for a MIGRATE-only file, got %d: %+v", len(res.Requirements), res.Requirements)
	}
}

func TestPlaceholderTokenSkipped(t *testing.T) {
	dir := t.TempDir()
	line := `// CANARY: <ID>; FEATURE="<name>"; ASPECT=<aspect>; STATUS=<status>; UPDATED=2025-11-02`
	if err := os.WriteFile(filepath.Join(dir, "placeholder.go"), []byte(line+"\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	sc := NewScanner()
	res, err := sc.ScanRepository(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(res.Requirements) != 0 {
		t.Fatalf("expected 0 requirements for placeholder token, got %d", len(res.Requirements))
	}
}
