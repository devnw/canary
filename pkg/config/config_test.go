// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package config

import (
	"os"
	"path/filepath"
	"reflect"
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
	if !reflect.DeepEqual(cfg.Sources[1], want) {
		t.Errorf("Sources[1] = %+v, want %+v", cfg.Sources[1], want)
	}
}

func TestCANARY_CBIN_306_LoadSources_TicketSyncFields(t *testing.T) {
	root := writeProjectYAML(t, `
project:
  name: "demo"
  key: "CBIN"
sources:
  - name: platform
    type: jira
    key: "PLAT"
    url: "https://company.atlassian.net/browse/{id}"
    api: "https://company.atlassian.net"
    status_map:
      STUB: "Backlog"
      IMPL: "In Development"
`)
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("len(Sources) = %d, want 1", len(cfg.Sources))
	}
	got := cfg.Sources[0]
	if got.API != "https://company.atlassian.net" {
		t.Errorf("API = %q, want https://company.atlassian.net", got.API)
	}
	want := map[string]string{"STUB": "Backlog", "IMPL": "In Development"}
	if !reflect.DeepEqual(got.StatusMap, want) {
		t.Errorf("StatusMap = %+v, want %+v", got.StatusMap, want)
	}
}

func TestCANARY_ENG_3958_LoadSources_ProjectDestinationFields(t *testing.T) {
	root := writeProjectYAML(t, `
project:
  name: "demo"
  key: "CBIN"
sources:
  - name: platform
    type: jira
    key: "PLAT"
    project: "PLATPROJ"
    destination: true
`)
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("len(Sources) = %d, want 1", len(cfg.Sources))
	}
	got := cfg.Sources[0]
	if got.Project != "PLATPROJ" {
		t.Errorf("Project = %q, want PLATPROJ", got.Project)
	}
	if !got.Destination {
		t.Error("Destination = false, want true")
	}
}

func TestCANARY_ENG_3958_LoadSources_ProjectDestinationDefaultToZeroValue(t *testing.T) {
	// A source with no project/destination fields must parse to the zero
	// value — old configs behave exactly as before.
	root := writeProjectYAML(t, `
project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
`)
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := cfg.Sources[0]
	if got.Project != "" || got.Destination {
		t.Errorf("Project/Destination = %q/%v, want empty/false", got.Project, got.Destination)
	}
}

// TestCANARY_ENG_3961_LoadPeers is a passthrough test proving `peers:`
// parses into ProjectConfig.Peers with Name/Root intact.
func TestCANARY_ENG_3961_LoadPeers(t *testing.T) {
	root := writeProjectYAML(t, `
project:
  name: "demo"
  key: "CBIN"
peers:
  - name: upstream
    root: "../upstream-repo"
  - name: sibling
    root: "/abs/path/to/sibling"
`)
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Peers) != 2 {
		t.Fatalf("len(Peers) = %d, want 2", len(cfg.Peers))
	}
	want := []PeerConfig{
		{Name: "upstream", Root: "../upstream-repo"},
		{Name: "sibling", Root: "/abs/path/to/sibling"},
	}
	if !reflect.DeepEqual(cfg.Peers, want) {
		t.Errorf("Peers = %+v, want %+v", cfg.Peers, want)
	}
}

// TestCANARY_ENG_3961_LoadPeers_AbsentIsEmpty proves a config with no
// `peers:` key parses to an empty (nil) Peers slice.
func TestCANARY_ENG_3961_LoadPeers_AbsentIsEmpty(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\n")
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Peers) != 0 {
		t.Errorf("Peers should be empty when absent, got %d", len(cfg.Peers))
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
