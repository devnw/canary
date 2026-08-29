// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3960_Next_ExternalSatisfied_NotBlocking,TestCANARY_ENG_3960_Next_ExternalUnsatisfied_Blocking,TestCANARY_ENG_3960_Next_ExternalUnknown_NotBlockingByDefault,TestCANARY_ENG_3960_Next_ExternalUnknown_StrictBlocks,TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking,TestCANARY_ENG_3960_Next_ExternalUnknown_NoteOncePerRun; UPDATED=2026-08-29
package next

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

func engRegistry(t *testing.T) *sources.Registry {
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

func openTestDB(t *testing.T, dependsOn string) *storage.DB {
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
	return db
}

func tokenFor(db *storage.DB, t *testing.T, reqID string) *storage.Token {
	t.Helper()
	tokens, err := db.GetTokensByReqID(reqID)
	if err != nil || len(tokens) == 0 {
		t.Fatalf("GetTokensByReqID(%s): %v (tokens=%d)", reqID, err, len(tokens))
	}
	return tokens[0]
}

// TestCANARY_ENG_3960_Next_ExternalSatisfied_NotBlocking proves a dependency
// on an external id with zero local tokens does not block next-selection
// when the cached remote status is in the source's done-set.
func TestCANARY_ENG_3960_Next_ExternalSatisfied_NotBlocking(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	if err := external.SaveCache(root, map[string]string{"ENG-1": "Done"}, freshRefTime()); err != nil {
		t.Fatal(err)
	}
	reg := engRegistry(t)

	if hasUnresolvedDependencies(db, tok, reg, root, false, map[string]bool{}) {
		t.Error("satisfied external dependency must not block")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnsatisfied_Blocking proves a cached but
// not-done remote status blocks selection.
func TestCANARY_ENG_3960_Next_ExternalUnsatisfied_Blocking(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	if err := external.SaveCache(root, map[string]string{"ENG-1": "In Progress"}, freshRefTime()); err != nil {
		t.Fatal(err)
	}
	reg := engRegistry(t)

	if !hasUnresolvedDependencies(db, tok, reg, root, false, map[string]bool{}) {
		t.Error("unsatisfied external dependency must block")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnknown_NotBlockingByDefault proves that
// an external dependency with no cached status degrades to non-blocking
// unless --strict-external is set.
func TestCANARY_ENG_3960_Next_ExternalUnknown_NotBlockingByDefault(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir() // no cache file at all
	reg := engRegistry(t)

	if hasUnresolvedDependencies(db, tok, reg, root, false, map[string]bool{}) {
		t.Error("unknown external dependency must not block by default (degradation is sacred)")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnknown_StrictBlocks proves --strict-external
// flips unknown external dependencies to blocking.
func TestCANARY_ENG_3960_Next_ExternalUnknown_StrictBlocks(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	reg := engRegistry(t)

	if !hasUnresolvedDependencies(db, tok, reg, root, true, map[string]bool{}) {
		t.Error("unknown external dependency must block under --strict-external")
	}
}

// TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking proves that a
// dependency with zero local tokens that is NOT external (unresolvable
// prefix, or resolves to a flatfile source) keeps the legacy
// missing-dependency-blocks behavior.
func TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking(t *testing.T) {
	db := openTestDB(t, "CBIN-999")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	reg := engRegistry(t)

	if !hasUnresolvedDependencies(db, tok, reg, root, false, map[string]bool{}) {
		t.Error("missing local (flatfile) dependency must still block")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnknown_NoteOncePerRun proves the stderr
// note for an unknown external dependency is printed once per id per run
// (dedup via the shared warned map), not once per candidate evaluated.
func TestCANARY_ENG_3960_Next_ExternalUnknown_NoteOncePerRun(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	reg := engRegistry(t)
	warned := map[string]bool{}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	hasUnresolvedDependencies(db, tok, reg, root, false, warned)
	hasUnresolvedDependencies(db, tok, reg, root, false, warned)
	os.Stderr = origStderr
	_ = w.Close()

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	got := string(buf[:n])
	_ = r.Close()

	if count := strings.Count(got, "no cached status"); count != 1 {
		t.Errorf("expected exactly 1 stderr note across 2 calls, got %d: %q", count, got)
	}
	if !strings.Contains(got, "note: external dependency ENG-1 has no cached status") {
		t.Errorf("stderr note format unexpected: %q", got)
	}
}

// freshRefTime returns a timestamp guaranteed to be considered fresh by
// external.Resolve's staleness check.
func freshRefTime() time.Time { return time.Now().UTC() }
