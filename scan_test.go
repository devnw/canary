// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canary_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary"
)

func TestAcceptance_ParseAndSummarizeFixture_WithPromotion(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "file1.zig"), `// CANARY: REQ=REQ-GQL-042; FEATURE="CDC/Streaming"; ASPECT=API; STATUS=STUB; TEST=tests/e2e_cdc.zig:TestCANARY_REQ_GQL_042_StartStop; OWNER=streaming; UPDATED=2025-10-15`)
	mustWrite(t, filepath.Join(dir, "file2.go"), `// CANARY: REQ=REQ-GQL-046; FEATURE="TDE"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_REQ_GQL_046_KeyRotate; OWNER=security; UPDATED=2025-10-15`)

	rep, err := canary.Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	// STUB remains STUB
	if rep.Summary.ByStatus["STUB"] != 1 {
		t.Fatalf("expected STUB=1 got %+v", rep.Summary.ByStatus)
	}
	// IMPL with test is promoted to TESTED
	if rep.Summary.ByStatus["TESTED"] != 1 {
		t.Fatalf("expected TESTED=1 (promotion) got %+v", rep.Summary.ByStatus)
	}
	if rep.Summary.ByStatus["IMPL"] != 0 {
		t.Fatalf("expected IMPL=0 after promotion got %+v", rep.Summary.ByStatus)
	}
}

func TestAcceptance_PromotionToBenched(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "file3.zig"), `// CANARY: REQ=REQ-GQL-050; FEATURE="RecursiveQuery"; ASPECT=Planner; STATUS=IMPL; BENCH=BenchmarkCANARY_REQ_GQL_050_RecursivePerf; UPDATED=2025-10-15`)
	rep, err := canary.Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if rep.Summary.ByStatus["BENCHED"] != 1 {
		t.Fatalf("expected BENCHED=1 promotion got %+v", rep.Summary.ByStatus)
	}
}

func TestAcceptance_VerifyFailsOnOverclaim(t *testing.T) {
	// fake GAP line claiming Implemented but only STUB in repo
	claimsContent := `| REQ‑GQL‑042 | Streaming and CDC | Implemented | evidence |`
	p := filepath.Join(t.TempDir(), "GAP.md")
	mustWrite(t, p, claimsContent)

	// repo with only STUB marker
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "cdc.zig"), `// CANARY: REQ=REQ-GQL-042; FEATURE="CDC"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-15`)

	rep, _ := canary.Scan(dir)
	claims, _ := canary.ParseGAPClaims(p)
	if err := canary.VerifyClaims(rep, claims); err == nil {
		t.Fatalf("expected verify error, got nil")
	}
}

func TestScanMultiLanguageTokens(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"example.zig":   "// CANARY: REQ=CBIN-201; FEATURE=\"ZigCore\"; ASPECT=API; STATUS=IMPL; OWNER=zig; UPDATED=2025-11-02",
		"lib.rs":        "// CANARY: REQ=CBIN-202; FEATURE=\"RustCore\"; ASPECT=Engine; STATUS=IMPL; OWNER=rust; UPDATED=2025-11-02",
		"component.tsx": "// CANARY: REQ=CBIN-203; FEATURE=\"TSXComp\"; ASPECT=FrontEnd; STATUS=STUB; OWNER=web; UPDATED=2025-11-02",
		"view.jsx":      "/* CANARY: REQ=CBIN-204; FEATURE=\"JSXView\"; ASPECT=FrontEnd; STATUS=IMPL; OWNER=web; UPDATED=2025-11-02 */",
	}
	for name, line := range cases {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(line+"\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	rep, err := canary.Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := map[string]bool{"CBIN-201": true, "CBIN-202": true, "CBIN-203": true, "CBIN-204": true}
	for _, r := range rep.Requirements {
		delete(want, r.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing requirements parsed: %v", want)
	}
}

func TestScanDocCommentVariants(t *testing.T) {
	dir := t.TempDir()
	content := `/**
 * CANARY: REQ=CBIN-205; FEATURE="DocComment"; ASPECT=API; STATUS=IMPL; OWNER=docs; UPDATED=2025-11-02
 */
/** CANARY: REQ=CBIN-206; FEATURE="InlineDoc"; ASPECT=API; STATUS=IMPL; OWNER=docs; UPDATED=2025-11-02 */
/* * CANARY: REQ=CBIN-207; FEATURE="StarPrefixed"; ASPECT=API; STATUS=IMPL; OWNER=docs; UPDATED=2025-11-02 */
/* CANARY: REQ=CBIN-208; FEATURE="PlainBlock"; ASPECT=API; STATUS=IMPL; OWNER=docs; UPDATED=2025-11-02 */
`
	if err := os.WriteFile(filepath.Join(dir, "doc.js"), []byte(content), 0644); err != nil {
		t.Fatalf("write doc.js: %v", err)
	}
	rep, err := canary.Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := map[string]bool{"CBIN-205": true, "CBIN-206": true, "CBIN-207": true, "CBIN-208": true}
	for _, r := range rep.Requirements {
		delete(want, r.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing doc comment reqs: %v", want)
	}
}

func BenchmarkScanLarge(b *testing.B) {
	dir := b.TempDir()
	// Generate 5000 small files with CANARY tokens across mixed extensions
	total := 5000
	for i := 0; i < total; i++ {
		ext := ".go"
		switch i % 5 {
		case 1:
			ext = ".rs"
		case 2:
			ext = ".zig"
		case 3:
			ext = ".tsx"
		case 4:
			ext = ".md"
		}
		content := fmt.Sprintf("// CANARY: REQ=CBIN-%03d; FEATURE=\"Bench\"; ASPECT=API; STATUS=IMPL; OWNER=bench; UPDATED=2025-11-02\n", i)
		if ext == ".md" {
			content = fmt.Sprintf("<!-- CANARY: REQ=CBIN-%03d; FEATURE=\"Bench\"; ASPECT=API; STATUS=IMPL; OWNER=bench; UPDATED=2025-11-02 -->\n", i)
		}
		name := fmt.Sprintf("file_%d%s", i, ext)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			b.Fatalf("write %s: %v", name, err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := canary.Scan(dir)
		if err != nil {
			b.Fatalf("scan error: %v", err)
		}
	}
}

func TestAcceptance_StrictStaleness(t *testing.T) {
	dir := t.TempDir()
	// 90 days old
	mustWrite(t, filepath.Join(dir, "tde.go"), `// CANARY: REQ=REQ-GQL-046; FEATURE="TDE"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_REQ_GQL_046_KeyRotate; UPDATED=2025-06-01`)
	rep, _ := canary.Scan(dir)
	if err := canary.CheckStaleness(rep, 60*24*60*60*1e9); err == nil {
		t.Fatalf("expected staleness error")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
