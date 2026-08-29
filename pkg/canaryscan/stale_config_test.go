package canaryscan

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func writeProjectYAML(t *testing.T, root string, stalenessDays int) {
	t.Helper()
	dir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "project:\n  name: test\nverification:\n  staleness_days: " + strconv.Itoa(stalenessDays) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCANARY_CBIN_304_StaleDaysFromConfig proves verification.staleness_days from
// .canary/project.yaml is actually consulted: a 10-day-old TESTED token trips
// --strict against a configured 7-day window, but would not against the 30-day
// default (simulated here via an explicit Config.StaleDays=30 override).
func TestCANARY_CBIN_304_StaleDaysFromConfig(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")
	root := t.TempDir()
	writeProjectYAML(t, root, 7)
	// 10 days before the fixed ref time.
	src := "package x\n" +
		"// CANARY: REQ=CBIN-777; FEATURE=\"Aging\"; ASPECT=API; STATUS=TESTED; TEST=TestAging; UPDATED=2026-08-19\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("7-day project config trips strict", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cfg := Config{Root: root, Out: filepath.Join(root, "status.json"), Strict: true}
		code := Run(cfg, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("expected exit 2 (stale), got %d; stderr=%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "CANARY_STALE REQ=CBIN-777") {
			t.Errorf("expected CANARY_STALE diag for CBIN-777, got: %s", stderr.String())
		}
		if !strings.Contains(stderr.String(), "threshold=7") {
			t.Errorf("expected threshold=7 (from project.yaml), got: %s", stderr.String())
		}
	})

	t.Run("explicit 30-day override does not trip strict", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cfg := Config{Root: root, Out: filepath.Join(root, "status2.json"), Strict: true, StaleDays: 30}
		code := Run(cfg, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("expected exit 0 (not stale under 30d), got %d; stderr=%s", code, stderr.String())
		}
		if strings.Contains(stderr.String(), "CANARY_STALE") {
			t.Errorf("did not expect staleness diag with 30-day override, got: %s", stderr.String())
		}
	})
}

// TestCANARY_CBIN_304_UpdateStaleV2IDs proves the diag REQ regex now matches
// v2-style multi-segment IDs (e.g. CBIN-CLI-001), which the old
// `REQ=([A-Z][A-Z0-9]*-\d+)` pattern could never match, silently no-opping.
func TestCANARY_CBIN_304_UpdateStaleV2IDs(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")
	root := t.TempDir()
	src := "package x\n" +
		"// CANARY: REQ=CBIN-CLI-001; FEATURE=\"V2ID\"; ASPECT=CLI; STATUS=TESTED; TEST=TestV2; UPDATED=2020-01-01\n"
	path := filepath.Join(root, "x.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	diags := []string{"CANARY_STALE REQ=CBIN-CLI-001 updated=2020-01-01 age_days=2431 threshold=30"}

	updatedFiles, tokenCount, err := UpdateStaleTokens(root, DefaultSkipRegex(), diags)
	if err != nil {
		t.Fatal(err)
	}
	if len(updatedFiles) != 1 {
		t.Fatalf("expected 1 file updated (regression: v2 ID silently no-ops), got %d: %v", len(updatedFiles), updatedFiles)
	}
	if tokenCount != 1 {
		t.Fatalf("expected 1 token rewritten, got %d", tokenCount)
	}
	b, _ := os.ReadFile(path)
	if strings.Contains(string(b), "UPDATED=2020-01-01") {
		t.Errorf("stale v2-ID token was not updated:\n%s", b)
	}
	if !strings.Contains(string(b), "UPDATED=2026-08-29") {
		t.Errorf("UPDATED not rewritten to test timestamp:\n%s", b)
	}
}

// TestCANARY_CBIN_304_UpdateStaleAddsMissingUpdated proves UpdateStaleTokens can now
// ADD an UPDATED= attribute to a token line that lacks one entirely, rather than only
// rewriting an existing date. canaryscan.Scan hard-errors on a token missing UPDATED,
// so this drives UpdateStaleTokens directly with a synthetic diag rather than going
// through Run/Scan.
func TestCANARY_CBIN_304_UpdateStaleAddsMissingUpdated(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")

	t.Run("line comment token", func(t *testing.T) {
		root := t.TempDir()
		src := "package x\n" +
			"// CANARY: REQ=CBIN-305; FEATURE=\"NoUpdated\"; ASPECT=API; STATUS=TESTED; TEST=TestNoUpdated\n"
		path := filepath.Join(root, "x.go")
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		diags := []string{"CANARY_STALE REQ=CBIN-305 updated=MISSING age_days=99999 threshold=30"}

		updatedFiles, tokenCount, err := UpdateStaleTokens(root, DefaultSkipRegex(), diags)
		if err != nil {
			t.Fatal(err)
		}
		if len(updatedFiles) != 1 || tokenCount != 1 {
			t.Fatalf("expected 1 file/1 token updated, got files=%d tokens=%d", len(updatedFiles), tokenCount)
		}
		b, _ := os.ReadFile(path)
		if !strings.Contains(string(b), "UPDATED=2026-08-29") {
			t.Errorf("expected UPDATED= to be added:\n%s", b)
		}
		if !strings.Contains(string(b), `TEST=TestNoUpdated; UPDATED=2026-08-29`) {
			t.Errorf("expected UPDATED to be appended after existing content:\n%s", b)
		}
	})

	t.Run("block comment token is CRLF-safe and preserves closer", func(t *testing.T) {
		root := t.TempDir()
		src := "package x\r\n" +
			"/* CANARY: REQ=CBIN-306; FEATURE=\"Block\"; ASPECT=API; STATUS=BENCHED; BENCH=Bench6 */\r\n"
		path := filepath.Join(root, "x.go")
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		diags := []string{"CANARY_STALE REQ=CBIN-306 updated=MISSING age_days=99999 threshold=30"}

		updatedFiles, tokenCount, err := UpdateStaleTokens(root, DefaultSkipRegex(), diags)
		if err != nil {
			t.Fatal(err)
		}
		if len(updatedFiles) != 1 || tokenCount != 1 {
			t.Fatalf("expected 1 file/1 token updated, got files=%d tokens=%d", len(updatedFiles), tokenCount)
		}
		b, _ := os.ReadFile(path)
		content := string(b)
		if !strings.Contains(content, "UPDATED=2026-08-29") {
			t.Errorf("expected UPDATED= to be added:\n%s", content)
		}
		if !strings.Contains(content, "BENCH=Bench6; UPDATED=2026-08-29 */") {
			t.Errorf("expected UPDATED inserted before the */ closer:\n%s", content)
		}
		if !strings.Contains(content, "*/\r\n") {
			t.Errorf("expected CRLF line endings preserved:\n%q", content)
		}
	})
}

// TestCANARY_CBIN_304_RunReportsActualRewriteCount proves run.go's "Updated N stale
// tokens" message reports the count of tokens actually rewritten by UpdateStaleTokens,
// not len(staleDiags). Two files carry an identical stale TESTED token (same REQ,
// FEATURE, ASPECT, OWNER, UPDATED), so Scan aggregates them into a single Feature and
// Stale() emits exactly one diag -- but both physical files/lines get rewritten, so the
// honest count is 2, not 1.
func TestCANARY_CBIN_304_RunReportsActualRewriteCount(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")
	root := t.TempDir()
	src := "package x\n" +
		"// CANARY: REQ=CBIN-501; FEATURE=\"Dup\"; ASPECT=API; STATUS=TESTED; TEST=TestDup; UPDATED=2020-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cfg := Config{Root: root, Out: filepath.Join(root, "status.json"), UpdateStale: true}
	code := Run(cfg, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Updated 2 stale tokens in 2 files") {
		t.Errorf("expected honest count 'Updated 2 stale tokens in 2 files' (not the diag count of 1), got: %s", stderr.String())
	}
}
