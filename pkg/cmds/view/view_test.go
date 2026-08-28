// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package view

import (
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/pkg/storage"
)

func seedDB(t *testing.T) (dbPath, root string) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	dbPath = filepath.Join(root, ".canary", "canary.db")

	// storage.Open does not run migrations itself (see storage.go:82
	// "Migrations are handled automatically by the CLI's PersistentPreRunE");
	// tests must migrate explicitly, mirroring the convention used in
	// internal/storage/refs_test.go and cmd/canary/main.go's PersistentPreRunE.
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	toks := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED",
			FilePath: "scan.go", LineNumber: 10, Test: "TestScan", DependsOn: "CBIN-101"},
		{ReqID: "CBIN-105", Feature: "ScannerCLI", Aspect: "CLI", Status: "IMPL",
			FilePath: "cli.go", LineNumber: 5, RelatedTo: "PLAT-4521"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.ReplaceRefs("diagram", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 7},
	}); err != nil {
		t.Fatal(err)
	}
	return dbPath, root
}

func TestCANARY_CBIN_204_BuildView(t *testing.T) {
	dbPath, root := seedDB(t)
	v, err := BuildView(dbPath, root, "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if v.ReqID != "CBIN-105" {
		t.Errorf("ReqID = %q", v.ReqID)
	}
	if v.Statuses["TESTED"] != 1 || v.Statuses["IMPL"] != 1 {
		t.Errorf("Statuses = %v", v.Statuses)
	}
	if v.Completion != 50 {
		t.Errorf("Completion = %d, want 50", v.Completion)
	}
	if len(v.Files) != 2 || v.FilesTotal != 2 {
		t.Errorf("Files = %v (total %d)", v.Files, v.FilesTotal)
	}
	if len(v.Tests) != 1 || v.Tests[0] != "TestScan" {
		t.Errorf("Tests = %v", v.Tests)
	}
	if len(v.DependsOn) != 1 || v.DependsOn[0] != "CBIN-101" {
		t.Errorf("DependsOn = %v", v.DependsOn)
	}
	if len(v.RelatedTo) != 1 || v.RelatedTo[0] != "PLAT-4521" {
		t.Errorf("RelatedTo = %v", v.RelatedTo)
	}
	if len(v.Diagrams) != 1 || v.Diagrams[0] != "docs/arch.md:7" {
		t.Errorf("Diagrams = %v", v.Diagrams)
	}
}

func TestCANARY_CBIN_204_BuildView_FileCap(t *testing.T) {
	dbPath, root := seedDB(t)
	v, err := BuildView(dbPath, root, "CBIN-105", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Files) != 1 || v.FilesTotal != 2 {
		t.Errorf("cap not applied: files=%v total=%d", v.Files, v.FilesTotal)
	}
}

func TestCANARY_CBIN_204_BuildView_DiagramCap(t *testing.T) {
	dbPath, root := seedDB(t)

	// Add extra diagram refs to test the cap
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.ReplaceRefs("diagram", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 7},
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/design.md", LineNumber: 12},
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/flow.md", LineNumber: 5},
	}); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "CBIN-105", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Diagrams) != 1 || v.DiagramsTotal != 3 {
		t.Errorf("cap not applied: diagrams=%v total=%d", v.Diagrams, v.DiagramsTotal)
	}
}

func TestCANARY_CBIN_204_BuildView_NotFound(t *testing.T) {
	dbPath, root := seedDB(t)
	if _, err := BuildView(dbPath, root, "CBIN-999", 10); err == nil {
		t.Error("unknown requirement must return an error")
	}
}
