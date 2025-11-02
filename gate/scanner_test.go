package gate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenericScanner(t *testing.T) {
    dir := t.TempDir()
    // create sample files
    write := func(name, content string) {
        if err := os.WriteFile(filepath.Join(dir, name), []byte(content+"\n"), 0644); err != nil { t.Fatalf("write %s: %v", name, err) }
    }
    write("a.go", `// CANARY: REQ=CBIN-301; FEATURE="GenScan"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_301_API_Gen; UPDATED=2025-11-02`)
    write("b.rs", `// CANARY: REQ=CBIN-301; FEATURE="GenScan"; ASPECT=API; STATUS=IMPL; BENCH=BenchmarkCANARY_CBIN_301_API_Gen; UPDATED=2025-11-02`)
    sc := NewScanner()
    res, err := sc.ScanRepository(dir)
    if err != nil { t.Fatalf("scan error: %v", err) }
    if len(res.Requirements) != 1 { t.Fatalf("expected 1 requirement, got %d", len(res.Requirements)) }
    // Promotion: IMPL + test -> TESTED; bench stronger -> BENCHED
    status := res.Requirements[0].Features[0].Status
    foundBenched := false
    for _, f := range res.Requirements[0].Features { if f.Status == "BENCHED" { foundBenched = true } }
    if !foundBenched { t.Fatalf("expected BENCHED promotion, got status=%s", status) }
}
