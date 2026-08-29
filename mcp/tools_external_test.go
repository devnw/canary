// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"path/filepath"
	"testing"
	"time"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

func mcpEngRegistry(t *testing.T) *sources.Registry {
	t.Helper()
	reg, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "eng", Type: "jira", Key: "ENG"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return reg
}

func mcpTestDB(t *testing.T, dependsOn string) (*storage.DB, *storage.Token) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := storage.AutoMigrate(dbPath); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.UpsertToken(&storage.Token{
		ReqID:     "CBIN-500",
		Feature:   "Consumer",
		Aspect:    "API",
		Status:    "STUB",
		Priority:  1,
		FilePath:  "consumer.go",
		DependsOn: dependsOn,
	}); err != nil {
		t.Fatalf("UpsertToken: %v", err)
	}
	tokens, err := db.GetTokensByReqID("CBIN-500")
	if err != nil || len(tokens) == 0 {
		t.Fatalf("GetTokensByReqID: %v (n=%d)", err, len(tokens))
	}
	return db, tokens[0]
}

// TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking proves the MCP
// next mirror does not block on an external dependency whose cached remote
// status is in the source's done-set.
func TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := external.SaveCache(".", map[string]string{"ENG-1": "Done"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	db, tok := mcpTestDB(t, "ENG-1")
	reg := mcpEngRegistry(t)

	if hasUnresolvedDependencies(db, tok, reg, map[string]bool{}) {
		t.Error("satisfied external dependency must not block")
	}
}

// TestCANARY_ENG_3960_MCP_Next_ExternalUnsatisfied_Blocking proves a cached
// but not-done remote status blocks the MCP next mirror.
func TestCANARY_ENG_3960_MCP_Next_ExternalUnsatisfied_Blocking(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := external.SaveCache(".", map[string]string{"ENG-1": "In Progress"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	db, tok := mcpTestDB(t, "ENG-1")
	reg := mcpEngRegistry(t)

	if !hasUnresolvedDependencies(db, tok, reg, map[string]bool{}) {
		t.Error("unsatisfied external dependency must block")
	}
}

// TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_NotBlocking proves the MCP
// mirror always uses the non-strict default: an external dependency with no
// cached status does not block (MCP has no --strict-external flag).
func TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_NotBlocking(t *testing.T) {
	t.Chdir(t.TempDir()) // no cache file
	db, tok := mcpTestDB(t, "ENG-1")
	reg := mcpEngRegistry(t)

	if hasUnresolvedDependencies(db, tok, reg, map[string]bool{}) {
		t.Error("unknown external dependency must not block (degradation is sacred)")
	}
}

// TestCANARY_ENG_3960_MCP_Next_LocalMissingDep_StillBlocking proves a
// missing LOCAL (non-external) dependency keeps the legacy blocking
// behavior in the MCP mirror.
func TestCANARY_ENG_3960_MCP_Next_LocalMissingDep_StillBlocking(t *testing.T) {
	t.Chdir(t.TempDir())
	db, tok := mcpTestDB(t, "CBIN-999")
	reg := mcpEngRegistry(t)

	if !hasUnresolvedDependencies(db, tok, reg, map[string]bool{}) {
		t.Error("missing local (flatfile) dependency must still block")
	}
}

// TestCANARY_ENG_3960_MCP_Next_LocalTokensWinOverExternalCache_IMPLBlocks
// proves the MCP next mirror uses local TESTED/BENCHED logic for a
// dependency id that has real local CANARY tokens, even when the id also
// matches an external (ticket) source's key and the cache reports a
// done-equivalent status: a local IMPL token still blocks.
func TestCANARY_ENG_3960_MCP_Next_LocalTokensWinOverExternalCache_IMPLBlocks(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := external.SaveCache(".", map[string]string{"ENG-1": "Done"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	db, tok := mcpTestDB(t, "ENG-1")
	if err := db.UpsertToken(&storage.Token{
		ReqID: "ENG-1", Feature: "Upstream", Aspect: "API", Status: "IMPL", FilePath: "upstream.go",
	}); err != nil {
		t.Fatal(err)
	}
	reg := mcpEngRegistry(t)

	if !hasUnresolvedDependencies(db, tok, reg, map[string]bool{}) {
		t.Error("local IMPL token must block even though cached remote status is Done")
	}
}

// TestCANARY_ENG_3960_MCP_Next_LocalTokensWinOverExternalCache_TESTEDPasses
// proves the inverse in the MCP mirror: a local TESTED token satisfies the
// dependency even though the cached remote status ("In Progress") would
// otherwise be unsatisfied.
func TestCANARY_ENG_3960_MCP_Next_LocalTokensWinOverExternalCache_TESTEDPasses(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := external.SaveCache(".", map[string]string{"ENG-1": "In Progress"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	db, tok := mcpTestDB(t, "ENG-1")
	if err := db.UpsertToken(&storage.Token{
		ReqID: "ENG-1", Feature: "Upstream", Aspect: "API", Status: "TESTED", FilePath: "upstream.go",
	}); err != nil {
		t.Fatal(err)
	}
	reg := mcpEngRegistry(t)

	if hasUnresolvedDependencies(db, tok, reg, map[string]bool{}) {
		t.Error("local TESTED token must satisfy the dependency even though cached remote status is In Progress")
	}
}
