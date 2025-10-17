// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcceptance_ParseAndSummarizeFixture_WithPromotion(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "file1.zig"), `// CANARY: REQ=REQ-GQL-042; FEATURE="CDC/Streaming"; ASPECT=API; STATUS=STUB; TEST=tests/e2e_cdc.zig:TestCANARY_REQ_GQL_042_StartStop; OWNER=streaming; UPDATED=2025-09-20`)
	mustWrite(t, filepath.Join(dir, "file2.go"), `// CANARY: REQ=REQ-GQL-046; FEATURE="TDE"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_REQ_GQL_046_KeyRotate; OWNER=security; UPDATED=2025-09-20`)

	rep, err := Scan(dir)
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
	mustWrite(t, filepath.Join(dir, "file3.zig"), `// CANARY: REQ=REQ-GQL-050; FEATURE="RecursiveQuery"; ASPECT=Planner; STATUS=IMPL; BENCH=BenchmarkCANARY_REQ_GQL_050_RecursivePerf; UPDATED=2025-09-20`)
	rep, err := Scan(dir)
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
	mustWrite(t, filepath.Join(dir, "cdc.zig"), `// CANARY: REQ=REQ-GQL-042; FEATURE="CDC"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20`)

	rep, _ := Scan(dir)
	claims, _ := ParseGAPClaims(p)
	if err := VerifyClaims(rep, claims); err == nil {
		t.Fatalf("expected verify error, got nil")
	}
}

func TestAcceptance_StrictStaleness(t *testing.T) {
	dir := t.TempDir()
	// 90 days old
	mustWrite(t, filepath.Join(dir, "tde.go"), `// CANARY: REQ=REQ-GQL-046; FEATURE="TDE"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_REQ_GQL_046_KeyRotate; UPDATED=2025-06-01`)
	rep, _ := Scan(dir)
	if err := CheckStaleness(rep, 60*24*60*60*1e9); err == nil {
		t.Fatalf("expected staleness error")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
