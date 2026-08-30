// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2026-08-30
package next

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/storage"
)

// gitInit turns root into a git repository with one empty commit, so
// `git rev-parse HEAD` resolves. `next` binds both its index-freshness check
// and its evidence lookups to that commit, so a fixture without one is a
// fixture where nothing can ever be current.
func gitInit(tb testing.TB, root string) {
	tb.Helper()
	steps := [][]string{
		{"init", "-q"},
		{"-c", "user.email=next@example.com", "-c", "user.name=Next", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-q", "-m", "init"},
	}
	for _, args := range steps {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...) //nolint:gosec // fixed argv
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			tb.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// headSHA returns root's HEAD commit.
func headSHA(tb testing.TB, root string) string {
	tb.Helper()
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output() //nolint:gosec // fixed argv
	if err != nil {
		tb.Fatalf("git rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// nextRepo creates a git repository and makes it the working directory for
// the duration of the test. `next` answers about the tree it is standing in,
// so every fixture here stands somewhere disposable -- never in this
// repository.
func nextRepo(tb testing.TB) string {
	tb.Helper()
	root := tb.TempDir()
	gitInit(tb, root)
	tb.Chdir(root)
	return root
}

// seedIndex writes tokens into a fresh index at root and records the metadata
// that makes it current for root's HEAD -- exactly what `canary index` does,
// and what `next` requires before it will believe an index describes the tree
// in front of it.
func seedIndex(tb testing.TB, root string, tokens ...*storage.Token) string {
	tb.Helper()
	dbPath := filepath.Join(root, ".canary", "canary.db")
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		tb.Fatalf("open index: %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, tok := range tokens {
		if tok.ProjectID == "" {
			tok.ProjectID = "default"
		}
		if err := db.UpsertToken(tok); err != nil {
			tb.Fatalf("seed %s: %v", tok.ReqID, err)
		}
	}
	meta := storage.IndexMeta{
		Root:         root,
		ProjectID:    "default",
		CommitSHA:    headSHA(tb, root),
		ParserSchema: canaryscan.ParserSchemaVersion,
		ScanDigest:   "fixture",
		IndexedAt:    "2026-01-01T00:00:00Z",
	}
	if err := db.PutIndexMeta(meta); err != nil {
		tb.Fatalf("record index metadata: %v", err)
	}
	return dbPath
}

// passRecord builds one PASS evidence record binding a requirement's
// feature/aspect to commit.
func passRecord(projectID, reqID, feature, aspect, commit string) evidence.Record {
	return evidence.Record{
		ProjectID:      projectID,
		RequirementID:  reqID,
		Feature:        feature,
		Aspect:         aspect,
		TestID:         "TestFixture",
		Command:        "go test ./...",
		Result:         "PASS",
		CommitSHA:      commit,
		ObservedAt:     "2026-01-01T00:00:00Z",
		Runner:         "local",
		ArtifactDigest: "sha256:" + strings.Repeat("ab", 32),
	}
}

// writeEvidence writes root's evidence store.
func writeEvidence(tb testing.TB, root string, recs ...evidence.Record) {
	tb.Helper()
	path := filepath.Join(root, ".canary", "evidence.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		tb.Fatal(err)
	}
	data, err := json.Marshal(evidence.File{SchemaVersion: 1, Records: recs})
	if err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		tb.Fatal(err)
	}
}

// TestCANARY_CBIN_132_CLI_NextPrioritySelection verifies priority-based selection logic
func TestCANARY_CBIN_132_CLI_NextPrioritySelection(t *testing.T) {
	root := nextRepo(t)
	dbPath := seedIndex(t, root,
		&storage.Token{ReqID: "CBIN-201", Feature: "LowPriority", Aspect: "API", Status: "STUB", Priority: 10, FilePath: "test1.go"},
		&storage.Token{ReqID: "CBIN-202", Feature: "HighPriority", Aspect: "CLI", Status: "STUB", Priority: 1, FilePath: "test2.go"},
		&storage.Token{ReqID: "CBIN-203", Feature: "MediumPriority", Aspect: "API", Status: "STUB", Priority: 5, FilePath: "test3.go"},
		&storage.Token{ReqID: "CBIN-204", Feature: "AlreadyTested", Aspect: "API", Status: "TESTED", Priority: 2, FilePath: "test4.go"},
	)

	res, err := selectNext(selection{DBPath: dbPath, Root: root, Stderr: io.Discard})
	if err != nil {
		t.Fatalf("selectNext failed: %v", err)
	}

	if res.Source != SourceDatabase {
		t.Errorf("Source = %q, want %q for a current index", res.Source, SourceDatabase)
	}
	if res.Token == nil {
		t.Fatal("expected a token to be selected, got nil")
	}
	if res.Token.ReqID != "CBIN-202" {
		t.Errorf("expected CBIN-202 (highest priority STUB), got %s", res.Token.ReqID)
	}
	if res.Token.Priority != 1 {
		t.Errorf("expected priority 1, got %d", res.Token.Priority)
	}
	if res.Token.Status != "STUB" {
		t.Errorf("expected STATUS=STUB, got %s", res.Token.Status)
	}
}

// TestCANARY_CBIN_132_CLI_DependencyBlocking verifies dependency resolution
func TestCANARY_CBIN_132_CLI_DependencyBlocking(t *testing.T) {
	root := nextRepo(t)
	dbPath := seedIndex(t, root,
		&storage.Token{ReqID: "CBIN-301", Feature: "BlockedHighPriority", Aspect: "API", Status: "STUB", Priority: 1, FilePath: "test1.go", DependsOn: "CBIN-302"},
		&storage.Token{ReqID: "CBIN-302", Feature: "DependencyRequirement", Aspect: "API", Status: "STUB", Priority: 5, FilePath: "test2.go"},
		&storage.Token{ReqID: "CBIN-303", Feature: "IndependentLowPriority", Aspect: "CLI", Status: "STUB", Priority: 8, FilePath: "test3.go"},
	)

	selected, err := selectNextPriority(dbPath, "", nil)
	if err != nil {
		t.Fatalf("selectNextPriority failed: %v", err)
	}

	// Verify: Skips blocked CBIN-301, returns CBIN-302 (the dependency)
	if selected == nil {
		t.Fatal("expected a token to be selected, got nil")
	}
	if selected.ReqID != "CBIN-302" {
		t.Errorf("expected CBIN-302 (dependency), got %s", selected.ReqID)
	}
}

// TestCANARY_CBIN_132_CLI_DeclaredCompletionIsNotProof proves the one
// completion definition reaches the index path: a dependency whose every
// token says TESTED still blocks until evidence at this commit says so, and
// clears once it does. STATUS is a claim; only the evidence store proves it.
func TestCANARY_CBIN_132_CLI_DeclaredCompletionIsNotProof(t *testing.T) {
	root := nextRepo(t)
	dbPath := seedIndex(t, root,
		&storage.Token{ReqID: "CBIN-401", Feature: "Consumer", Aspect: "API", Status: "STUB", Priority: 1, FilePath: "consumer.go", DependsOn: "CBIN-402"},
		&storage.Token{ReqID: "CBIN-402", Feature: "Upstream", Aspect: "API", Status: "TESTED", Priority: 5, FilePath: "upstream.go"},
	)

	selected, err := selectNextPriority(dbPath, "", nil)
	if err != nil {
		t.Fatalf("selectNextPriority failed: %v", err)
	}
	// CBIN-402 declares itself TESTED, so it is not work either -- with its
	// dependent blocked there is nothing to hand out at all.
	if selected != nil {
		t.Fatalf("a TESTED declaration with no evidence must still block its dependent; got %s", selected.ReqID)
	}

	writeEvidence(t, root, passRecord("default", "CBIN-402", "Upstream", "API", headSHA(t, root)))

	selected, err = selectNextPriority(dbPath, "", nil)
	if err != nil {
		t.Fatalf("selectNextPriority failed: %v", err)
	}
	if selected == nil || selected.ReqID != "CBIN-401" {
		t.Fatalf("proven dependency must unblock its dependent; got %+v", selected)
	}
}

// TestCANARY_CBIN_132_CLI_TemplateRendering verifies prompt generation from template
func TestCANARY_CBIN_132_CLI_TemplateRendering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .canary structure
	canaryDir := filepath.Join(tmpDir, ".canary")
	specsDir := filepath.Join(canaryDir, "specs", "CBIN-401-test-feature")
	memoryDir := filepath.Join(canaryDir, "memory")
	templatesDir := filepath.Join(canaryDir, "templates")

	for _, dir := range []string{specsDir, memoryDir, templatesDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	specContent := `# Feature Specification: Test Feature

**Requirement ID:** CBIN-401
**Feature Name:** TestFeature
**Status:** STUB

## Success Criteria
- Requirement must be verified
- All tests must pass
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	constitutionContent := `# Constitution

## Article IV: Test-First Imperative
Tests must be written before implementation.
`
	if err := os.WriteFile(filepath.Join(memoryDir, "constitution.md"), []byte(constitutionContent), 0o644); err != nil {
		t.Fatalf("failed to write constitution: %v", err)
	}

	templateContent := `# Implementation Guidance: {{.Feature}}

**Requirement:** {{.ReqID}}
**Priority:** {{.Priority}}

## Constitution
{{.Constitution}}

## Specification
{{.SpecContent}}

## Extra Prompt ({{.PromptArg}})
{{.PromptContent}}
`
	if err := os.WriteFile(filepath.Join(templatesDir, "next-prompt-template.md"), []byte(templateContent), 0o644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	extraPrompt := filepath.Join(tmpDir, "extra-prompt.md")
	if err := os.WriteFile(extraPrompt, []byte("EXTRA GUIDANCE BODY"), 0o644); err != nil {
		t.Fatalf("failed to write extra prompt: %v", err)
	}

	token := &storage.Token{
		ReqID:    "CBIN-401",
		Feature:  "TestFeature",
		Aspect:   "API",
		Status:   "STUB",
		Priority: 3,
		FilePath: "test.go",
	}

	t.Chdir(tmpDir)

	prompt, err := renderPrompt(token, "", true, extraPrompt)
	if err != nil {
		t.Fatalf("renderPrompt failed: %v", err)
	}

	for _, want := range []string{"CBIN-401", "TestFeature", "Test-First Imperative", "Success Criteria", "EXTRA GUIDANCE BODY"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

// TestCANARY_CBIN_132_CLI_PromptArgOnlyReadWhenRendering proves --prompt-arg
// is resolved on the path that renders a template and nowhere else: a summary
// run never reads it, so an unreadable prompt cannot fail a command that was
// never going to use it.
func TestCANARY_CBIN_132_CLI_PromptArgOnlyReadWhenRendering(t *testing.T) {
	token := &storage.Token{ReqID: "CBIN-402", Feature: "Summary", Aspect: "API", Status: "STUB", Priority: 1}

	if _, err := renderPrompt(token, "", false, "./does-not-exist.md"); err != nil {
		t.Fatalf("summary output must not read the prompt argument: %v", err)
	}
	t.Chdir(t.TempDir())
	if _, err := renderPrompt(token, "", true, "./does-not-exist.md"); err == nil {
		t.Error("a prompt run with an unreadable --prompt-arg must fail")
	}
}

// TestCANARY_CBIN_132_CLI_NoWorkAvailable verifies behavior when all requirements are complete
func TestCANARY_CBIN_132_CLI_NoWorkAvailable(t *testing.T) {
	root := nextRepo(t)
	dbPath := seedIndex(t, root,
		&storage.Token{ReqID: "CBIN-501", Feature: "CompleteFeature1", Aspect: "API", Status: "TESTED", Priority: 1, FilePath: "test1.go"},
		&storage.Token{ReqID: "CBIN-502", Feature: "CompleteFeature2", Aspect: "CLI", Status: "BENCHED", Priority: 2, FilePath: "test2.go"},
	)

	res, err := selectNext(selection{DBPath: dbPath, Root: root, Stderr: io.Discard})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if res.Token != nil {
		t.Errorf("expected nil (no work available), got token: %s", res.Token.ReqID)
	}
	// The completion claim is only ever made from a current index, so this is
	// the one shape of "nothing to do" allowed to say so.
	if res.Source != SourceDatabase {
		t.Errorf("Source = %q, want %q", res.Source, SourceDatabase)
	}
	if got := noWorkMessage(res.Source, res.Blocked); got != "all requirements completed" {
		t.Errorf("noWorkMessage = %q", got)
	}
}

// TestCANARY_CBIN_132_CLI_FilesystemFallback verifies filesystem scan when database unavailable
func TestCANARY_CBIN_132_CLI_FilesystemFallback(t *testing.T) {
	root := nextRepo(t)
	dbPath := filepath.Join(root, "nonexistent.db") // Database doesn't exist

	fileContent := `package test

// CANARY: REQ=CBIN-999; FEATURE="FilesystemFallbackTest"; ASPECT=API; STATUS=STUB; PRIORITY=1; UPDATED=2025-10-16
func FilesystemFallbackTest() {
	// This is test data to verify filesystem scanning works correctly when database is unavailable.
}
`
	if err := os.WriteFile(filepath.Join(root, "test.go"), []byte(fileContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	res, err := selectNext(selection{DBPath: dbPath, Root: root, Stderr: io.Discard})
	if err != nil {
		t.Fatalf("filesystem fallback failed: %v", err)
	}
	if res.Source != SourceFilesystem {
		t.Errorf("Source = %q, want %q", res.Source, SourceFilesystem)
	}
	if res.Token == nil {
		t.Fatal("expected token from filesystem, got nil")
	}
	if res.Token.ReqID != "CBIN-999" {
		t.Errorf("expected CBIN-999 from filesystem, got %s", res.Token.ReqID)
	}
}

// TestCANARY_CBIN_132_CLI_StaleIndexIsNotUsed proves an index built at
// another commit is not consulted: it describes a tree that has moved, and
// answering from it is how `next` came to report work that no longer exists
// (and completion that was never true).
func TestCANARY_CBIN_132_CLI_StaleIndexIsNotUsed(t *testing.T) {
	root := nextRepo(t)
	dbPath := seedIndex(t, root,
		&storage.Token{ReqID: "CBIN-601", Feature: "Indexed", Aspect: "API", Status: "STUB", Priority: 1, FilePath: "indexed.go"},
	)

	// Rewrite the metadata as if the index had been built elsewhere/earlier.
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.PutIndexMeta(storage.IndexMeta{
		Root:         root,
		ProjectID:    "default",
		CommitSHA:    strings.Repeat("0", 40),
		ParserSchema: canaryscan.ParserSchemaVersion,
		ScanDigest:   "fixture",
		IndexedAt:    "2026-01-01T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	res, err := selectNext(selection{DBPath: dbPath, Root: root, Stderr: io.Discard})
	if err != nil {
		t.Fatalf("selectNext failed: %v", err)
	}
	if res.Source != SourceFilesystem {
		t.Errorf("Source = %q, want %q for an index built at another commit", res.Source, SourceFilesystem)
	}
	if res.Token != nil {
		t.Errorf("the tree holds no tokens; got %s from the stale index", res.Token.ReqID)
	}
	if got := noWorkMessage(res.Source, res.Blocked); got != "no actionable requirements found" {
		t.Errorf("noWorkMessage = %q, want the filesystem wording (never a completion claim)", got)
	}
}

// TestCANARY_CBIN_132_CLI_StatusFiltering verifies filtering by status
func TestCANARY_CBIN_132_CLI_StatusFiltering(t *testing.T) {
	root := nextRepo(t)
	dbPath := seedIndex(t, root,
		&storage.Token{ReqID: "CBIN-701", Feature: "StubFeature", Aspect: "API", Status: "STUB", Priority: 2, FilePath: "test1.go"},
		&storage.Token{ReqID: "CBIN-702", Feature: "ImplFeature", Aspect: "API", Status: "IMPL", Priority: 1, FilePath: "test2.go"},
	)

	selected, err := selectNextPriority(dbPath, "", map[string]string{"status": "IMPL"})
	if err != nil {
		t.Fatalf("selectNextPriority with filter failed: %v", err)
	}

	if selected == nil {
		t.Fatal("expected a token to be selected")
	}
	if selected.Status != "IMPL" {
		t.Errorf("expected STATUS=IMPL, got %s", selected.Status)
	}
	if selected.ReqID != "CBIN-702" {
		t.Errorf("expected CBIN-702, got %s", selected.ReqID)
	}
}

// TestCANARY_CBIN_132_CLI_ScanCandidateOrder proves the filesystem path is
// deterministic and prefers unstarted work: same priority, STUB before IMPL,
// then requirement id.
func TestCANARY_CBIN_132_CLI_ScanCandidateOrder(t *testing.T) {
	rep := canaryscan.Report{Requirements: []canaryscan.Requirement{
		{ID: "CBIN-020", Features: []canaryscan.Feature{{Feature: "B", Aspect: "API", Status: "IMPL", Files: []string{"b.go"}}}},
		{ID: "CBIN-030", Features: []canaryscan.Feature{{Feature: "C", Aspect: "API", Status: "STUB", Files: []string{"c.go"}}}},
		{ID: "CBIN-010", Features: []canaryscan.Feature{{Feature: "A", Aspect: "API", Status: "STUB", Files: []string{"a.go"}}}},
		{ID: "CBIN-040", Features: []canaryscan.Feature{{Feature: "D", Aspect: "API", Status: "TESTED", Files: []string{"d.go"}}}},
	}}

	got := scanCandidates(rep, map[string]string{})
	var ids []string
	for _, c := range got {
		ids = append(ids, c.ReqID)
	}
	want := []string{"CBIN-010", "CBIN-030", "CBIN-020"}
	if strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Errorf("candidate order = %v, want %v (TESTED is not actionable)", ids, want)
	}
}

// TestCANARY_CBIN_601_API_FilesystemTest verifies the FilesystemTest token is properly tracked
// This is a meta-test that verifies the CANARY tracking system itself.
func TestCANARY_CBIN_601_API_FilesystemTest(t *testing.T) {
	root := nextRepo(t)

	fileContent := `package test

// CANARY: REQ=CBIN-601; FEATURE="FilesystemTest"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_601_API_FilesystemTest; PRIORITY=1; UPDATED=2025-10-16
func FilesystemTest() {
	// This is a meta-test fixture used to verify filesystem scanning works correctly.
}
`
	if err := os.WriteFile(filepath.Join(root, "filesystemtest.go"), []byte(fileContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Force the filesystem path with a database that does not exist.
	selected, err := selectNextPriority(filepath.Join(root, "nonexistent.db"), "", map[string]string{"status": "TESTED"})
	if err != nil {
		t.Fatalf("filesystem scan failed: %v", err)
	}
	if selected == nil {
		t.Fatal("expected to find FilesystemTest token, got nil")
	}
	if selected.ReqID != "CBIN-601" {
		t.Errorf("expected CBIN-601, got %s", selected.ReqID)
	}
	if selected.Feature != "FilesystemTest" {
		t.Errorf("expected feature FilesystemTest, got %s", selected.Feature)
	}
	if selected.Aspect != "API" {
		t.Errorf("expected aspect API, got %s", selected.Aspect)
	}
	if selected.Status != "TESTED" {
		t.Errorf("expected status TESTED, got %s", selected.Status)
	}
}

// TestNextJSONCarriesSource proves the machine-readable output always names
// its source, and that "nothing to do" is a message rather than a silently
// empty object.
func TestNextJSONCarriesSource(t *testing.T) {
	var buf strings.Builder
	if err := emitJSON(&buf, result{Source: SourceFilesystem}); err != nil {
		t.Fatal(err)
	}
	var empty map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &empty); err != nil {
		t.Fatalf("output is not JSON: %q", buf.String())
	}
	if empty["source"] != SourceFilesystem {
		t.Errorf("source = %v, want %q", empty["source"], SourceFilesystem)
	}
	if empty["message"] != "no actionable requirements found" {
		t.Errorf("message = %v", empty["message"])
	}
	if _, present := empty["req_id"]; present {
		t.Error("req_id must be absent when nothing was selected")
	}

	buf.Reset()
	tok := &storage.Token{ReqID: "CBIN-1", Feature: "F", Aspect: "API", Status: "STUB", Priority: 3, FilePath: "a.go", UpdatedAt: "2026-01-01"}
	if err := emitJSON(&buf, result{Token: tok, Source: SourceDatabase}); err != nil {
		t.Fatal(err)
	}
	var chosen map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &chosen); err != nil {
		t.Fatalf("output is not JSON: %q", buf.String())
	}
	if chosen["req_id"] != "CBIN-1" || chosen["source"] != SourceDatabase {
		t.Errorf("unexpected output: %v", chosen)
	}
	if _, present := chosen["message"]; present {
		t.Error("message must be absent when a requirement was selected")
	}
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2026-08-30
// BenchmarkCANARY_CBIN_132_CLI_PriorityQuery measures priority query performance
// Target: <100ms per operation (for <10,000 requirements)
func BenchmarkCANARY_CBIN_132_CLI_PriorityQuery(b *testing.B) {
	root := nextRepo(b)

	tokens := make([]*storage.Token, 0, 1000)
	for i := 0; i < 1000; i++ {
		tokens = append(tokens, &storage.Token{
			ReqID:    fmt.Sprintf("CBIN-%04d", i),
			Feature:  fmt.Sprintf("Feature%d", i),
			Aspect:   "API",
			Status:   "STUB",
			Priority: (i % 10) + 1,
			FilePath: fmt.Sprintf("test%d.go", i),
		})
	}
	dbPath := seedIndex(b, root, tokens...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := selectNextPriority(dbPath, "", nil); err != nil {
			b.Fatalf("selectNextPriority failed: %v", err)
		}
	}
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; BENCH=BenchmarkCANARY_CBIN_132_CLI_PromptGeneration; OWNER=canary; UPDATED=2026-08-30
// BenchmarkCANARY_CBIN_132_CLI_PromptGeneration measures prompt rendering performance
// Target: <500ms per operation
func BenchmarkCANARY_CBIN_132_CLI_PromptGeneration(b *testing.B) {
	tmpDir := b.TempDir()

	canaryDir := filepath.Join(tmpDir, ".canary")
	specsDir := filepath.Join(canaryDir, "specs", "CBIN-999-bench-feature")
	memoryDir := filepath.Join(canaryDir, "memory")
	templatesDir := filepath.Join(canaryDir, "templates")

	for _, dir := range []string{specsDir, memoryDir, templatesDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	specContent := `# Feature Specification: Benchmark Feature

**Requirement ID:** CBIN-999
**Feature Name:** BenchmarkFeature
**Status:** STUB

## Success Criteria
1. Performance meets target
2. All sections rendered correctly
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		b.Fatalf("failed to write spec file: %v", err)
	}

	constitutionContent := `# CANARY Development Constitution

## Article IV: Test-First Imperative
All implementation MUST follow Test-Driven Development.
`
	if err := os.WriteFile(filepath.Join(memoryDir, "constitution.md"), []byte(constitutionContent), 0o644); err != nil {
		b.Fatalf("failed to write constitution: %v", err)
	}

	templateContent := `# Implementation Guidance: {{.Feature}}

**Requirement ID:** {{.ReqID}}
**Aspect:** {{.Aspect}}
**Priority:** {{.Priority}}

## Constitutional Principles
{{.Constitution}}

## Full Specification
{{.SpecContent}}

## Test Guidance
{{.TestGuidance}}
`
	if err := os.WriteFile(filepath.Join(templatesDir, "next-prompt-template.md"), []byte(templateContent), 0o644); err != nil {
		b.Fatalf("failed to write template: %v", err)
	}

	token := &storage.Token{
		ReqID:    "CBIN-999",
		Feature:  "BenchmarkFeature",
		Aspect:   "API",
		Status:   "STUB",
		Priority: 1,
		FilePath: "test.go",
	}

	b.Chdir(tmpDir)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := renderPrompt(token, "", true, ""); err != nil {
			b.Fatalf("renderPrompt failed: %v", err)
		}
	}
}
