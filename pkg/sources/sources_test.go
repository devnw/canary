// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package sources

import (
	"os"
	"path/filepath"
	"testing"

	"devnw.dev/canary/pkg/config"
)

func testRegistry(t *testing.T) *Registry {
	t.Helper()
	r, err := NewRegistry([]Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"},
		{Name: "app", Type: "gitlab", Key: "GL", URL: "https://gitlab.com/devnw/app/-/issues/{num}"},
		{Name: "oss", Type: "github", Key: "GH", URL: "https://github.com/devnw/app/issues/{num}"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return r
}

func TestCANARY_CBIN_201_RegistryPattern(t *testing.T) {
	r := testRegistry(t)
	for _, id := range []string{"CBIN-105", "PLAT-4521", "GL-88", "GH-7"} {
		if !r.Pattern().MatchString(id) {
			t.Errorf("Pattern should match %s", id)
		}
	}
	if r.Pattern().MatchString("OTHER-123") {
		t.Error("Pattern should not match unconfigured prefix OTHER-123")
	}
}

func TestCANARY_CBIN_201_RegistryResolve(t *testing.T) {
	r := testRegistry(t)
	src, ok := r.Resolve("PLAT-4521")
	if !ok || src.Type != "jira" {
		t.Errorf("Resolve(PLAT-4521) = %+v, %v; want jira source", src, ok)
	}
	if _, ok := r.Resolve("OTHER-1"); ok {
		t.Error("Resolve should fail for unconfigured prefix")
	}
}

func TestCANARY_CBIN_201_TicketURL(t *testing.T) {
	r := testRegistry(t)
	cases := map[string]string{
		"PLAT-4521": "https://company.atlassian.net/browse/PLAT-4521",
		"GL-88":     "https://gitlab.com/devnw/app/-/issues/88",
		"GH-7":      "https://github.com/devnw/app/issues/7",
		"CBIN-105":  "",
		"OTHER-9":   "",
	}
	for id, want := range cases {
		if got := r.TicketURL(id); got != want {
			t.Errorf("TicketURL(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestCANARY_CBIN_201_Normalize(t *testing.T) {
	r := testRegistry(t)
	cases := map[string]string{
		"CBIN-42":  "CBIN-042", // flatfile: padded
		"CBIN-105": "CBIN-105",
		"PLAT-42":  "PLAT-42", // jira: verbatim, never padded
		"GL-8":     "GL-8",
		"OTHER-7":  "OTHER-7", // unknown: verbatim
	}
	for id, want := range cases {
		if got := r.Normalize(id); got != want {
			t.Errorf("Normalize(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestCANARY_CBIN_201_ClaimPattern(t *testing.T) {
	r := testRegistry(t)
	gap := "✅ CBIN-105\n  ✅ PLAT-4521\n- [ ] GL-88\n✅ OTHER-1\n"
	got := map[string]bool{}
	for _, m := range r.ClaimPattern().FindAllStringSubmatch(gap, -1) {
		got[m[1]] = true
	}
	if !got["CBIN-105"] || !got["PLAT-4521"] {
		t.Errorf("claims missing: %v", got)
	}
	if got["GL-88"] || got["OTHER-1"] {
		t.Errorf("false claims matched: %v", got)
	}
}

func TestCANARY_CBIN_201_DefaultRegistry(t *testing.T) {
	r := Default()
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("Default registry must match CBIN IDs")
	}
	if got := r.Normalize("CBIN-42"); got != "CBIN-042" {
		t.Errorf("Default Normalize(CBIN-42) = %q, want CBIN-042", got)
	}
}

func TestCANARY_CBIN_201_NewRegistryValidation(t *testing.T) {
	if _, err := NewRegistry([]Source{{Name: "a", Type: "jira", Key: "bad-key"}}); err == nil {
		t.Error("lowercase/hyphen key must be rejected")
	}
	if _, err := NewRegistry([]Source{
		{Name: "a", Type: "flatfile", Key: "X"},
		{Name: "b", Type: "jira", Key: "X"},
	}); err == nil {
		t.Error("duplicate keys must be rejected")
	}
	if _, err := NewRegistry([]Source{{Name: "a", Type: "svn", Key: "X"}}); err == nil {
		t.Error("unknown type must be rejected")
	}
	if _, err := NewRegistry(nil); err == nil {
		t.Error("empty registry must be rejected")
	}
}

func TestCANARY_CBIN_201_FromProjectConfigNil(t *testing.T) {
	// FromProjectConfig(nil) → registry that matches CBIN-105 (Default fallback).
	r := FromProjectConfig(nil)
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("nil config should fall back to Default registry matching CBIN-105")
	}
}

func TestCANARY_CBIN_201_FromProjectConfigSynthesizedFlatfile(t *testing.T) {
	// Config with Sources empty and Project.Key = "ACME" → registry matches ACME-7 and Normalize("ACME-7") == "ACME-007"
	cfg := &config.ProjectConfig{}
	cfg.Project.Key = "ACME"
	cfg.Sources = []config.SourceConfig{}

	r := FromProjectConfig(cfg)
	if !r.Pattern().MatchString("ACME-7") {
		t.Error("synthesized ACME source should match ACME-7")
	}
	if got := r.Normalize("ACME-7"); got != "ACME-007" {
		t.Errorf("Normalize(ACME-7) = %q, want ACME-007", got)
	}
}

func TestCANARY_CBIN_201_FromProjectConfigInvalidKeyFallback(t *testing.T) {
	// Config with Sources empty and Project.Key = "bad-key" (invalid) → falls back to Default (matches CBIN-105, does NOT match BAD-1).
	cfg := &config.ProjectConfig{}
	cfg.Project.Key = "bad-key"
	cfg.Sources = []config.SourceConfig{}

	r := FromProjectConfig(cfg)
	if r.Pattern().MatchString("BAD-1") {
		t.Error("invalid key should fall back to Default, should not match BAD-1")
	}
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("invalid key should fall back to Default, should match CBIN-105")
	}
}

func TestCANARY_CBIN_201_FromProjectConfigMalformedSourceFallback(t *testing.T) {
	// Config with a malformed source entry (e.g. Type: "svn") → falls back to Default.
	cfg := &config.ProjectConfig{}
	cfg.Project.Key = "TEST"
	cfg.Sources = []config.SourceConfig{
		{Name: "bad", Type: "svn", Key: "BAD"},
	}

	r := FromProjectConfig(cfg)
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("malformed source should fall back to Default, should match CBIN-105")
	}
	if r.Pattern().MatchString("BAD-1") {
		t.Error("malformed source should fall back to Default, should not match BAD-1")
	}
}

func TestCANARY_CBIN_201_FromProjectConfigValidSources(t *testing.T) {
	// Config with valid sources (one flatfile CBIN + one jira PLAT with URL) → Resolve("PLAT-1") returns the jira source.
	cfg := &config.ProjectConfig{}
	cfg.Project.Key = "CBIN"
	cfg.Sources = []config.SourceConfig{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"},
	}

	r := FromProjectConfig(cfg)
	src, ok := r.Resolve("PLAT-1")
	if !ok || src.Type != "jira" || src.Key != "PLAT" {
		t.Errorf("Resolve(PLAT-1) should return jira source, got %+v ok=%v", src, ok)
	}
	if !r.Pattern().MatchString("CBIN-105") || !r.Pattern().MatchString("PLAT-1") {
		t.Error("should match both CBIN-105 and PLAT-1")
	}
}

func TestCANARY_CBIN_201_LoadFromRootNoCanaryDir(t *testing.T) {
	// LoadFromRoot(t.TempDir()) (no .canary dir) → Default fallback (matches CBIN-105).
	tmpdir := t.TempDir()

	r := LoadFromRoot(tmpdir)
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("no .canary dir should fall back to Default, should match CBIN-105")
	}
}

func TestCANARY_CBIN_306_FromProjectConfig_TicketSyncFieldsPassThrough(t *testing.T) {
	// SourceConfig.API and SourceConfig.StatusMap must reach the built
	// Source unchanged, so pkg/ticket's ComputePlan can read StatusMap
	// overrides through the registry.
	cfg := &config.ProjectConfig{}
	cfg.Project.Key = "CBIN"
	cfg.Sources = []config.SourceConfig{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{
			Name: "platform", Type: "jira", Key: "PLAT",
			URL:       "https://company.atlassian.net/browse/{id}",
			API:       "https://company.atlassian.net",
			StatusMap: map[string]string{"STUB": "Backlog", "TESTED": "Closed"},
		},
	}

	r := FromProjectConfig(cfg)
	src, ok := r.Resolve("PLAT-1")
	if !ok {
		t.Fatal("Resolve(PLAT-1) failed")
	}
	if src.API != "https://company.atlassian.net" {
		t.Errorf("API = %q, want https://company.atlassian.net", src.API)
	}
	want := map[string]string{"STUB": "Backlog", "TESTED": "Closed"}
	if len(src.StatusMap) != len(want) || src.StatusMap["STUB"] != want["STUB"] || src.StatusMap["TESTED"] != want["TESTED"] {
		t.Errorf("StatusMap = %+v, want %+v", src.StatusMap, want)
	}
}

func TestCANARY_CBIN_201_LoadFromRootWithProjectYAML(t *testing.T) {
	// LoadFromRoot with a real .canary/project.yaml written to a temp dir declaring a jira source PLAT
	// → TicketURL("PLAT-42") expands correctly.
	tmpdir := t.TempDir()
	canaryDir := filepath.Join(tmpdir, ".canary")
	if err := os.MkdirAll(canaryDir, 0755); err != nil {
		t.Fatalf("mkdir .canary: %v", err)
	}

	yaml := `project:
  name: Test Project
  key: TEST
sources:
  - name: platform
    type: jira
    key: PLAT
    url: https://company.atlassian.net/browse/{id}
`
	projectPath := filepath.Join(canaryDir, "project.yaml")
	if err := os.WriteFile(projectPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("write project.yaml: %v", err)
	}

	r := LoadFromRoot(tmpdir)
	url := r.TicketURL("PLAT-42")
	want := "https://company.atlassian.net/browse/PLAT-42"
	if url != want {
		t.Errorf("TicketURL(PLAT-42) = %q, want %q", url, want)
	}
}
