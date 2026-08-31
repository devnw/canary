// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/storage"
)

// TestAuditF27_ConstitutionAmends proves `canary constitution "<text>"` on a
// repo that already has a constitution appends a dated amendment inside the
// managed markers rather than doing nothing.
func TestAuditF27_ConstitutionAmends(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	memDir := filepath.Join(root, ".canary", "memory")
	if err := os.MkdirAll(memDir, 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(memDir, "constitution.md")
	if err := os.WriteFile(path, []byte("# Constitution\n\nPrinciples.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	run(t, root, bin, "constitution", "Amendment: prefer stdlib")

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "<!-- CANARY:amendments:START -->") ||
		!strings.Contains(s, "<!-- CANARY:amendments:END -->") {
		t.Fatalf("amendment markers missing:\n%s", s)
	}
	if !strings.Contains(s, "prefer stdlib") {
		t.Fatalf("amendment text missing:\n%s", s)
	}

	// A second amendment appends rather than clobbers.
	run(t, root, bin, "constitution", "Amendment: test first")
	b, _ = os.ReadFile(path)
	s = string(b)
	if !strings.Contains(s, "prefer stdlib") || !strings.Contains(s, "test first") {
		t.Fatalf("second amendment clobbered the first:\n%s", s)
	}
	if strings.Count(s, "<!-- CANARY:amendments:START -->") != 1 {
		t.Fatalf("amendments section was duplicated:\n%s", s)
	}
}

// TestAuditF27_ShowAmbiguous proves a prefix that matches several requirements
// prints the sorted candidate list and exits non-zero.
func TestAuditF27_ShowAmbiguous(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	seedShowTokens(t, root)

	out := runExpectFail(t, root, bin, "show", "CBIN-0")
	if !strings.Contains(out, "ambiguous requirement id") {
		t.Fatalf("expected ambiguity message, got:\n%s", out)
	}
	iOne := strings.Index(out, "CBIN-001")
	iTwo := strings.Index(out, "CBIN-002")
	if iOne < 0 || iTwo < 0 {
		t.Fatalf("candidates missing from output:\n%s", out)
	}
	if iOne > iTwo {
		t.Fatalf("candidates not sorted:\n%s", out)
	}
}

// TestAuditF27_ShowFilesystemFallback proves that with no index, show resolves
// a requirement by scanning the filesystem and marks the source.
func TestAuditF27_ShowFilesystemFallback(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	seedShowTokens(t, root)

	// No .canary/canary.db exists -> must fall back to filesystem.
	out := run(t, root, bin, "show", "CBIN-001")
	if !strings.Contains(out, "CBIN-001") {
		t.Fatalf("expected CBIN-001 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "source: filesystem") {
		t.Fatalf("expected filesystem source indicator, got:\n%s", out)
	}
}

// TestAuditF27_PluginVersionsMatch proves the two plugin manifests declare the
// same version. The released binary version is tag-derived; the manifests must
// at least agree with each other.
func TestAuditF27_PluginVersionsMatch(t *testing.T) {
	claude := readPluginVersion(t, repoPath(t, filepath.Join("claude-plugin", ".claude-plugin", "plugin.json")))
	cursor := readPluginVersion(t, repoPath(t, filepath.Join("cursor-plugin", ".cursor-plugin", "plugin.json")))
	if claude == "" || cursor == "" {
		t.Fatalf("empty plugin version(s): claude=%q cursor=%q", claude, cursor)
	}
	if claude != cursor {
		t.Fatalf("plugin versions disagree: claude=%q cursor=%q", claude, cursor)
	}
}

// TestAuditF27_SchemaDocMatches proves docs/DB_SCHEMA.md is exactly what
// `canary db schema` prints (storage.SchemaDDL). If they drift, regenerate the
// doc with `canary db schema > docs/DB_SCHEMA.md`.
func TestAuditF27_SchemaDocMatches(t *testing.T) {
	ddl, err := storage.SchemaDDL()
	if err != nil {
		t.Fatalf("SchemaDDL: %v", err)
	}
	doc, err := os.ReadFile(repoPath(t, filepath.Join("docs", "DB_SCHEMA.md")))
	if err != nil {
		t.Fatalf("read docs/DB_SCHEMA.md: %v", err)
	}
	if string(doc) != ddl {
		t.Fatalf("docs/DB_SCHEMA.md is out of date; regenerate with `canary db schema > docs/DB_SCHEMA.md`")
	}
}

func seedShowTokens(t *testing.T, root string) {
	t.Helper()
	content := "package a\n" +
		"// CANARY: REQ=CBIN-001; FEATURE=\"One\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-30\n" +
		"// CANARY: REQ=CBIN-002; FEATURE=\"Two\"; ASPECT=Engine; STATUS=TESTED; UPDATED=2026-08-30\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readPluginVersion(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return m.Version
}
