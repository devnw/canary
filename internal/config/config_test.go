// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProjectYAML(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCANARY_CBIN_201_LoadSources(t *testing.T) {
	root := writeProjectYAML(t, `
project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: platform
    type: jira
    key: "PLAT"
    url: "https://company.atlassian.net/browse/{id}"
  - name: app
    type: gitlab
    key: "GL"
    url: "https://gitlab.com/devnw/app/-/issues/{num}"
  - name: oss
    type: github
    key: "GH"
    url: "https://github.com/devnw/app/issues/{num}"
`)
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Project.Key != "CBIN" {
		t.Errorf("Project.Key = %q, want CBIN", cfg.Project.Key)
	}
	if len(cfg.Sources) != 4 {
		t.Fatalf("len(Sources) = %d, want 4", len(cfg.Sources))
	}
	want := SourceConfig{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"}
	if cfg.Sources[1] != want {
		t.Errorf("Sources[1] = %+v, want %+v", cfg.Sources[1], want)
	}
}

func TestCANARY_CBIN_201_LoadSources_AbsentIsEmpty(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\n")
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Sources) != 0 {
		t.Errorf("Sources should be empty when absent, got %d", len(cfg.Sources))
	}
}
