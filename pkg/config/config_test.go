// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

// TestLoadRejectsUnknownField proves the decoder is strict: a mistyped key
// (`projct:` for `project:`) is a hard error naming the offending field, not
// a silently-ignored block that leaves the setting unapplied.
// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=Storage; STATUS=TESTED; TEST=TestLoadRejectsUnknownField; UPDATED=2026-08-30
func TestLoadRejectsUnknownField(t *testing.T) {
	root := writeProjectYAML(t, "projct:\n  name: demo\n")
	_, err := Load(root)
	if err == nil {
		t.Fatal("unknown field accepted")
	}
	if !strings.Contains(err.Error(), "projct") {
		t.Errorf("error should name the unknown field, got %q", err.Error())
	}
}

// TestLoadRejectsDuplicateKey proves duplicate mapping keys are rejected:
// yaml.v3 silently keeps the last one, so a duplicated block would otherwise
// discard configuration the author believes is applied.
// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=Storage; STATUS=TESTED; TEST=TestLoadRejectsDuplicateKey; UPDATED=2026-08-30
func TestLoadRejectsDuplicateKey(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: first\nproject:\n  name: second\n")
	_, err := Load(root)
	if err == nil {
		t.Fatal("duplicate key accepted")
	}
	if !strings.Contains(err.Error(), "project") {
		t.Errorf("error should name the duplicated key, got %q", err.Error())
	}
}

// TestLoadRejectsDuplicateKeyNested proves the duplicate check walks nested
// mappings too, not just the document root.
func TestLoadRejectsDuplicateKeyNested(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: first\n  name: second\n")
	_, err := Load(root)
	if err == nil {
		t.Fatal("nested duplicate key accepted")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should name the duplicated key, got %q", err.Error())
	}
}

// TestLoadRejectsNegativeStaleness proves a negative staleness window is a
// config error rather than a silently-ignored value.
// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=Storage; STATUS=TESTED; TEST=TestLoadRejectsNegativeStaleness; UPDATED=2026-08-30
func TestLoadRejectsNegativeStaleness(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\nverification:\n  staleness_days: -1\n")
	if _, err := Load(root); err == nil {
		t.Fatal("negative staleness accepted")
	}
}

// TestLoadRejectsBadSourceType proves source validation runs at load time:
// an unknown source type fails the load instead of degrading to the default
// CBIN-only registry at scan time.
// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=Storage; STATUS=TESTED; TEST=TestLoadRejectsBadSourceType; UPDATED=2026-08-30
func TestLoadRejectsBadSourceType(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\nsources:\n  - name: bad\n    type: svn\n    key: BAD\n")
	_, err := Load(root)
	if err == nil {
		t.Fatal("unknown source type accepted")
	}
	if !strings.Contains(err.Error(), "svn") {
		t.Errorf("error should name the bad type, got %q", err.Error())
	}
}

func TestLoadRejectsBadSourceKey(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\nsources:\n  - name: bad\n    type: jira\n    key: bad-key\n")
	if _, err := Load(root); err == nil {
		t.Fatal("malformed source key accepted")
	}
}

func TestLoadRejectsFlatfileDestination(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\nsources:\n  - name: core\n    type: flatfile\n    key: CBIN\n    destination: true\n")
	if _, err := Load(root); err == nil {
		t.Fatal("flatfile destination accepted")
	}
}

func TestLoadRejectsTwoDestinations(t *testing.T) {
	root := writeProjectYAML(t, `project:
  name: demo
sources:
  - name: a
    type: jira
    key: AA
    destination: true
  - name: b
    type: jira
    key: BB
    destination: true
`)
	if _, err := Load(root); err == nil {
		t.Fatal("two destination sources accepted")
	}
}

func TestLoadRejectsEmptyPeerRoot(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\npeers:\n  - name: upstream\n    root: \"\"\n")
	if _, err := Load(root); err == nil {
		t.Fatal("peer with empty root accepted")
	}
}

// TestLoadRejectsMultiDocumentYAML proves a project.yaml containing a
// second `---`-separated YAML document is rejected rather than silently
// parsed from only its first document -- a second document (here holding an
// unknown field) must never be discarded without a word.
// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=Storage; STATUS=TESTED; TEST=TestLoadRejectsMultiDocumentYAML; UPDATED=2026-08-30
func TestLoadRejectsMultiDocumentYAML(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: first\n---\nproject:\n  name: second\nbogus_unknown: true\n")
	if _, err := Load(root); err == nil {
		t.Fatal("multi-document YAML accepted")
	}
}

func TestLoadRejectsBadProjectKey(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\n  key: bad-key\n")
	if _, err := Load(root); err == nil {
		t.Fatal("malformed project key accepted")
	}
}

// TestLoadMissingFileIsUnconfigured proves an absent project.yaml stays
// legal: unconfigured repos must keep scanning.
func TestLoadMissingFileIsUnconfigured(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load returned nil config for missing file")
	}
	if cfg.StalenessDays() != DefaultStaleDays {
		t.Errorf("StalenessDays() = %d, want %d", cfg.StalenessDays(), DefaultStaleDays)
	}
	if cfg.ProjectID() != "default" {
		t.Errorf("ProjectID() = %q, want default", cfg.ProjectID())
	}
}

// TestStalenessDaysAndProjectID proves the resolved accessors return the
// configured values when set.
func TestStalenessDaysAndProjectID(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\n  key: ACME\nverification:\n  staleness_days: 7\n")
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.StalenessDays() != 7 {
		t.Errorf("StalenessDays() = %d, want 7", cfg.StalenessDays())
	}
	if cfg.ProjectID() != "ACME" {
		t.Errorf("ProjectID() = %q, want ACME", cfg.ProjectID())
	}
}
