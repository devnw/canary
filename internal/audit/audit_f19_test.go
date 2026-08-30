// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/config"
)

func TestAuditF19(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".canary"), 0o755)
	os.WriteFile(filepath.Join(root, ".canary", "project.yaml"),
		[]byte("project:\n  name: x\nverification:\n  staleness_days: -5\n"), 0o644)
	if _, err := config.Load(root); err == nil {
		t.Fatal("negative staleness accepted")
	}
	os.WriteFile(filepath.Join(root, ".canary", "project.yaml"),
		[]byte("project:\n  name: x\nunknown_top_level: true\n"), 0o644)
	if _, err := config.Load(root); err == nil {
		t.Fatal("unknown field accepted")
	}
}

// TestAuditF19_ScanFailsOnInvalidConfig proves an invalid project.yaml stops
// the scan with exit 3 instead of silently degrading to the default registry
// and the default staleness window.
func TestAuditF19_ScanFailsOnInvalidConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"),
		[]byte("project:\n  name: x\nsources:\n  - name: bad\n    type: svn\n    key: BAD\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := canaryscan.Run(canaryscan.Config{
		Root: root,
		Out:  filepath.Join(root, "status.json"),
	}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "svn") {
		t.Errorf("stderr should explain the config failure, got %q", stderr.String())
	}
}

// TestAuditF19_ScanWithoutConfigStillWorks proves an unconfigured repo (no
// .canary/project.yaml at all) keeps scanning against the default registry.
func TestAuditF19_ScanWithoutConfigStillWorks(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := canaryscan.Run(canaryscan.Config{
		Root: root,
		Out:  filepath.Join(root, "status.json"),
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "requirements=1") {
		t.Errorf("expected the CBIN requirement to be scanned, got %q", stdout.String())
	}
}

// TestAuditF19_ProjectOnlyBadPatternExits3 pins the existing contract: an
// uncompilable requirements.id_pattern under --project-only is fatal.
func TestAuditF19_ProjectOnlyBadPatternExits3(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"),
		[]byte("project:\n  name: x\nrequirements:\n  id_pattern: \"CBIN-([0-9\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := canaryscan.Run(canaryscan.Config{
		Root:        root,
		Out:         filepath.Join(root, "status.json"),
		ProjectOnly: true,
	}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3; stderr=%s", code, stderr.String())
	}
}
