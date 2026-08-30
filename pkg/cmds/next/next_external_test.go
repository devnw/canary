// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3960_Next_ExternalSatisfied_NotBlocking,TestCANARY_ENG_3960_Next_ExternalUnsatisfied_Blocking,TestCANARY_ENG_3960_Next_ExternalUnknown_BlocksByDefault,TestCANARY_ENG_3960_Next_ExternalUnknown_AllowedPasses,TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking,TestCANARY_ENG_3960_Next_ExternalUnknown_NoteOncePerRun; UPDATED=2026-08-30
package next

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// gateFor builds the dependency gate these tests exercise, over root's
// evidence store and the given registry. It is the same object the command
// builds, so what is asserted here is what the command does.
func gateFor(t *testing.T, root string, reg *sources.Registry, allowUnknownExternal bool, stderr io.Writer) *depGate {
	t.Helper()
	recs, err := loadEvidenceRecords(root)
	if err != nil {
		t.Fatalf("load evidence: %v", err)
	}
	if stderr == nil {
		stderr = io.Discard
	}
	return &depGate{
		root:                 root,
		evidenceProjectID:    "default",
		commit:               fixtureCommit,
		recs:                 recs,
		reg:                  reg,
		allowUnknownExternal: allowUnknownExternal,
		warned:               map[string]bool{},
		stderr:               stderr,
	}
}

// fixtureCommit is the commit these fixtures' evidence binds to. It never
// has to match a real repository: the gate compares records against whatever
// commit it was given, and these tests supply both sides.
const fixtureCommit = "0123456789abcdef0123456789abcdef01234567"

// blockedBy runs the gate over one token's DEPENDS_ON list with its error
// asserted away. None of the external-dependency cases below can produce one
// -- an error here would mean the fixture, not the rule under test, is broken.
func blockedBy(t *testing.T, gate *depGate, db *storage.DB, projectID string, tok *storage.Token) bool {
	t.Helper()
	got, err := gate.blocked(tok.DependsOn, dbDeclared(db, projectID))
	if err != nil {
		t.Fatalf("gate.blocked: %v", err)
	}
	return got
}

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
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatalf("storage.OpenRW: %v", err)
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
		ProjectID: "default",
	}); err != nil {
		t.Fatalf("UpsertToken: %v", err)
	}
	return db
}

func tokenFor(db *storage.DB, t *testing.T, reqID string) *storage.Token {
	t.Helper()
	tokens, err := db.GetTokensByReqID("", reqID)
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

	if blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
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

	if !blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
		t.Error("unsatisfied external dependency must block")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnknown_BlocksByDefault proves that an
// external dependency whose state cannot be determined from disk blocks.
// It used to pass -- and handing an agent work whose prerequisite might not
// exist is not a graceful degradation, it is a wrong answer.
func TestCANARY_ENG_3960_Next_ExternalUnknown_BlocksByDefault(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir() // no cache file at all

	if !blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
		t.Error("unknown external dependency must block by default")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnknown_AllowedPasses proves the risk is
// still available on request: --allow-unknown-external restores the old
// non-blocking behavior for a caller who has decided to accept it.
func TestCANARY_ENG_3960_Next_ExternalUnknown_AllowedPasses(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()

	if blockedBy(t, gateFor(t, root, engRegistry(t), true, nil), db, "", tok) {
		t.Error("--allow-unknown-external must unblock an unknown external dependency")
	}
}

// TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking proves that a
// dependency with zero local tokens that is NOT external (unresolvable
// prefix, or resolves to a flatfile source) keeps the legacy
// missing-dependency-blocks behavior -- and that --allow-unknown-external
// does not reach it: a dependency naming nothing at all is not an external
// whose status happens to be uncached.
func TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking(t *testing.T) {
	db := openTestDB(t, "CBIN-999")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()

	if !blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
		t.Error("missing local (flatfile) dependency must still block")
	}
	if !blockedBy(t, gateFor(t, root, engRegistry(t), true, nil), db, "", tok) {
		t.Error("--allow-unknown-external must not excuse a dependency that names nothing")
	}
}

// TestCANARY_ENG_3960_Next_ExternalUnknown_NoteOncePerRun proves the stderr
// note for an unknown external dependency is printed once per id per run
// (dedup via the shared warned map), not once per candidate evaluated.
func TestCANARY_ENG_3960_Next_ExternalUnknown_NoteOncePerRun(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()

	var stderr strings.Builder
	gate := gateFor(t, root, engRegistry(t), false, &stderr)
	blockedBy(t, gate, db, "", tok)
	blockedBy(t, gate, db, "", tok)

	got := stderr.String()
	if count := strings.Count(got, "note: external dependency ENG-1"); count != 1 {
		t.Errorf("expected exactly 1 stderr note across 2 calls, got %d: %q", count, got)
	}
	if !strings.Contains(got, "blocks selection") {
		t.Errorf("stderr note must say what it did about it: %q", got)
	}
}

// freshRefTime returns a timestamp guaranteed to be considered fresh by
// external.Resolve's staleness check.
func freshRefTime() time.Time { return time.Now().UTC() }

// TestCANARY_ENG_3960_Next_LocalTokensWinOverExternalCache_UnprovenBlocks
// proves that a dependency id with real local CANARY tokens is judged by this
// project's own evidence even when it also matches an external (ticket)
// source's key and the remote-status cache reports done: without a passing
// record at this commit, it blocks.
func TestCANARY_ENG_3960_Next_LocalTokensWinOverExternalCache_UnprovenBlocks(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	if err := db.UpsertToken(&storage.Token{
		ReqID: "ENG-1", Feature: "Upstream", Aspect: "API", Status: "TESTED", FilePath: "upstream.go", ProjectID: "default",
	}); err != nil {
		t.Fatal(err)
	}
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	if err := external.SaveCache(root, map[string]string{"ENG-1": "Done"}, freshRefTime()); err != nil {
		t.Fatal(err)
	}

	if !blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
		t.Error("a locally declared dependency with no evidence must block, whatever the ticket cache says")
	}
}

// TestCANARY_ENG_3960_Next_LocalTokensWinOverExternalCache_ProvenPasses
// proves the inverse: local evidence satisfies the dependency even though the
// cached remote status ("In Progress") would otherwise be unsatisfied.
func TestCANARY_ENG_3960_Next_LocalTokensWinOverExternalCache_ProvenPasses(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	if err := db.UpsertToken(&storage.Token{
		ReqID: "ENG-1", Feature: "Upstream", Aspect: "API", Status: "TESTED", FilePath: "upstream.go", ProjectID: "default",
	}); err != nil {
		t.Fatal(err)
	}
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	if err := external.SaveCache(root, map[string]string{"ENG-1": "In Progress"}, freshRefTime()); err != nil {
		t.Fatal(err)
	}
	writeEvidenceAt(t, root, passRecord("default", "ENG-1", "Upstream", "API", fixtureCommit))

	if blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
		t.Error("a locally proven dependency must not block, whatever the ticket cache says")
	}
}

// TestCANARY_ENG_3960_Next_EvidenceAtAnotherCommitBlocks proves evidence is
// bound to a commit: a record proving the dependency at some other commit
// says nothing about this tree.
func TestCANARY_ENG_3960_Next_EvidenceAtAnotherCommitBlocks(t *testing.T) {
	db := openTestDB(t, "ENG-1")
	if err := db.UpsertToken(&storage.Token{
		ReqID: "ENG-1", Feature: "Upstream", Aspect: "API", Status: "TESTED", FilePath: "upstream.go", ProjectID: "default",
	}); err != nil {
		t.Fatal(err)
	}
	tok := tokenFor(db, t, "CBIN-500")
	root := t.TempDir()
	writeEvidenceAt(t, root, passRecord("default", "ENG-1", "Upstream", "API", strings.Repeat("f", 40)))

	if !blockedBy(t, gateFor(t, root, engRegistry(t), false, nil), db, "", tok) {
		t.Error("evidence recorded at another commit must not satisfy a dependency here")
	}
}

// writeEvidenceAt writes root's evidence store from arbitrary records.
func writeEvidenceAt(t *testing.T, root string, recs ...evidence.Record) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	writeEvidence(t, root, recs...)
}
