// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package view

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/storage"
)

// seedDBWithSources is like seedDB but writes a .canary/project.yaml
// declaring a CBIN flatfile source and an ENG jira source, and seeds a
// single CBIN-105 token whose DependsOn mixes a local id (CBIN-101) and an
// external id (ENG-12).
func seedDBWithSources(t *testing.T) (dbPath, root string) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}

	projectYAML := `project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: eng
    type: jira
    key: "ENG"
`
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(projectYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	dbPath = filepath.Join(root, ".canary", "canary.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}

	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.UpsertToken(&storage.Token{
		ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED",
		FilePath: "scan.go", LineNumber: 10, DependsOn: "CBIN-101,ENG-12",
	}); err != nil {
		t.Fatal(err)
	}
	return dbPath, root
}

// TestCANARY_ENG_3960_View_DependsOn_ExternalAnnotated proves BuildView
// annotates an external DependsOn entry with its cached remote status.
func TestCANARY_ENG_3960_View_DependsOn_ExternalAnnotated(t *testing.T) {
	dbPath, root := seedDBWithSources(t)
	if err := external.SaveCache(root, map[string]string{"ENG-12": "Done"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	want := []string{"CBIN-101", "ENG-12 (external: Done)"}
	if len(v.DependsOn) != len(want) {
		t.Fatalf("DependsOn = %v, want %v", v.DependsOn, want)
	}
	for i := range want {
		if v.DependsOn[i] != want[i] {
			t.Errorf("DependsOn[%d] = %q, want %q", i, v.DependsOn[i], want[i])
		}
	}
}

// TestCANARY_ENG_3960_View_DependsOn_LocalUnchanged proves a local
// (flatfile) DependsOn entry is left verbatim even when the project also
// configures external sources.
func TestCANARY_ENG_3960_View_DependsOn_LocalUnchanged(t *testing.T) {
	dbPath, root := seedDBWithSources(t) // no cache written -> ENG-12 unknown

	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	if v.DependsOn[0] != "CBIN-101" {
		t.Errorf("local DependsOn entry changed: %v", v.DependsOn)
	}
	if v.DependsOn[1] != "ENG-12 (external: no cached ticket status)" {
		t.Errorf("unknown external DependsOn entry = %q", v.DependsOn[1])
	}
}

// TestCANARY_ENG_3961_View_DependsOn_PeerAnnotated proves BuildView
// annotates a DependsOn entry resolved by a configured peer project with
// "(external: peer:<name>)", through the same annotateExternal path used
// for ticket-cache resolutions.
func TestCANARY_ENG_3961_View_DependsOn_PeerAnnotated(t *testing.T) {
	dbPath, root := seedDBWithSources(t)

	peerYAML := `project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: eng
    type: jira
    key: "ENG"
peers:
  - name: upstream-repo
    root: "peer"
`
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(peerYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	peerStatus := `{"requirements":[{"id":"ENG-12","features":[{"feature":"X","aspect":"Engine","status":"TESTED"}]}]}`
	if err := os.MkdirAll(filepath.Join(root, "peer"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "peer", "status.json"), []byte(peerStatus), 0o644); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	want := []string{"CBIN-101", "ENG-12 (external: peer:upstream-repo)"}
	if len(v.DependsOn) != len(want) {
		t.Fatalf("DependsOn = %v, want %v", v.DependsOn, want)
	}
	for i := range want {
		if v.DependsOn[i] != want[i] {
			t.Errorf("DependsOn[%d] = %q, want %q", i, v.DependsOn[i], want[i])
		}
	}
}

// TestCANARY_ENG_3960_View_DependsOn_LocalTokensWinOverExternalCache proves
// that a DependsOn id with real local CANARY tokens is rendered plain --
// never annotated with "(external: ...)" -- even when it also matches a
// configured external (ticket-source) prefix AND the remote-status cache
// holds a "done" entry for it. Local tokens always win over
// external.Resolve; see the comment on annotateExternal.
func TestCANARY_ENG_3960_View_DependsOn_LocalTokensWinOverExternalCache(t *testing.T) {
	dbPath, root := seedDBWithSources(t)

	// ENG-12 matches the "eng" jira source AND has a local IMPL token.
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertToken(&storage.Token{
		ReqID: "ENG-12", Feature: "Upstream", Aspect: "API", Status: "IMPL",
		FilePath: "upstream.go", LineNumber: 5,
	}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	if err := external.SaveCache(root, map[string]string{"ENG-12": "Done"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	want := []string{"CBIN-101", "ENG-12"}
	if len(v.DependsOn) != len(want) {
		t.Fatalf("DependsOn = %v, want %v", v.DependsOn, want)
	}
	for i := range want {
		if v.DependsOn[i] != want[i] {
			t.Errorf("DependsOn[%d] = %q, want %q (local tokens must suppress external annotation)", i, v.DependsOn[i], want[i])
		}
	}
}
