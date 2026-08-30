package canaryscan

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const archMD = "# Architecture\n" +
	"CBIN-105 mentioned in prose is NOT a diagram ref.\n" +
	"```mermaid\n" +
	"flowchart TD\n" +
	"  A[CBIN-105 Scanner] --> B[CBIN-042 Storage]\n" +
	"  B --> C[PLAT-4521 Ingest]\n" +
	"```\n" +
	"```go\n" +
	"// CBIN-999 in a go fence is not a diagram ref\n" +
	"```\n" +
	"```mermaid\n" +
	"sequenceDiagram\n" +
	"  participant S as CBIN-105\n" +
	"```\n"

func TestCANARY_CBIN_202_ExtractDiagramRefs(t *testing.T) {
	reg := ticketRegistry(t) // from sources_test.go: CBIN flatfile + PLAT jira
	refs := ExtractDiagramRefs("docs/arch.md", archMD, reg)
	got := map[string][]int{}
	for _, r := range refs {
		if r.File != "docs/arch.md" {
			t.Errorf("File = %q", r.File)
		}
		got[r.ReqID] = append(got[r.ReqID], r.Line)
	}
	// CBIN-042 normalized (flatfile padding), PLAT verbatim, go-fence and prose excluded
	want := map[string][]int{"CBIN-105": {5, 13}, "CBIN-042": {5}, "PLAT-4521": {6}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("refs = %v, want %v", got, want)
	}
}

func TestCANARY_CBIN_202_ScanAttachesDiagrams(t *testing.T) {
	root := t.TempDir()
	code := "package x\n// CANARY: REQ=CBIN-105; FEATURE=\"Scanner\"; ASPECT=Engine; STATUS=IMPL; UPDATED=2026-08-28\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}
	md := "```mermaid\nflowchart TD\n  A[CBIN-105] --> B[other]\n```\n"
	if err := os.WriteFile(filepath.Join(root, "arch.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	AnnotateSources(&rep, nil)
	var found *Requirement
	for i := range rep.Requirements {
		if rep.Requirements[i].ID == "CBIN-105" {
			found = &rep.Requirements[i]
		}
	}
	if found == nil {
		t.Fatal("CBIN-105 not in report")
	}
	if !reflect.DeepEqual(found.Diagrams, []string{"arch.md:3"}) {
		t.Errorf("Diagrams = %v, want [arch.md:3]", found.Diagrams)
	}
}

// TestCANARY_CBIN_202_CanaryIgnoreExcludesDiagramRefs verifies that Scan
// honors .canaryignore for diagram-ref extraction the same way it does for
// token scanning: a mermaid reference under an ignored directory is dropped,
// while the same reference outside it is kept.
func TestCANARY_CBIN_202_CanaryIgnoreExcludesDiagramRefs(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, ".canaryignore"), []byte("docs/\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	code := "package x\n// CANARY: REQ=CBIN-105; FEATURE=\"Scanner\"; ASPECT=Engine; STATUS=IMPL; UPDATED=2026-08-28\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	md := "```mermaid\nflowchart TD\n  A[CBIN-105] --> B[other]\n```\n"
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "arch.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "arch2.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}

	ignorePatterns, err := LoadCanaryIgnore(root)
	if err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, ignorePatterns, nil)
	if err != nil {
		t.Fatal(err)
	}

	var found *Requirement
	for i := range rep.Requirements {
		if rep.Requirements[i].ID == "CBIN-105" {
			found = &rep.Requirements[i]
		}
	}
	if found == nil {
		t.Fatal("CBIN-105 not in report")
	}
	if !reflect.DeepEqual(found.Diagrams, []string{"arch2.md:3"}) {
		t.Errorf("Diagrams = %v, want [arch2.md:3] (docs/arch.md should be excluded by .canaryignore)", found.Diagrams)
	}
}

// TestCANARY_CBIN_202_ExtractDiagramRefsWholeFileMmd verifies that a .mmd
// file is treated as a whole-file mermaid diagram: IDs are extracted without
// requiring ```mermaid fences.
func TestCANARY_CBIN_202_ExtractDiagramRefsWholeFileMmd(t *testing.T) {
	reg := ticketRegistry(t)
	content := "flowchart TD\n  A[CBIN-105 Scanner] --> B[PLAT-4521 Ingest]\n"
	refs := ExtractDiagramRefs("flow.mmd", content, reg)
	got := map[string][]int{}
	for _, r := range refs {
		if r.File != "flow.mmd" {
			t.Errorf("File = %q", r.File)
		}
		got[r.ReqID] = append(got[r.ReqID], r.Line)
	}
	want := map[string][]int{"CBIN-105": {2}, "PLAT-4521": {2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("refs = %v, want %v", got, want)
	}
}

// TestCANARY_CBIN_202_ScanDiagramRefsSkipsOversizedFile proves ScanDiagramRefs
// checks a file's size against MaxFileBytes before reading it, the same way
// Scan's own token reader does, instead of doing an unbounded os.ReadFile. A
// file over the limit must be reported as a "file_too_large" ScanIssue and
// skipped rather than read into memory.
func TestCANARY_CBIN_202_ScanDiagramRefsSkipsOversizedFile(t *testing.T) {
	root := t.TempDir()

	pad := strings.Repeat("x", MaxFileBytes+1)
	big := "```mermaid\nflowchart TD\n  A[CBIN-105] --> B[other]\n```\n<!-- " + pad + " -->\n"
	bigPath := filepath.Join(root, "big.md")
	if err := os.WriteFile(bigPath, []byte(big), 0o600); err != nil {
		t.Fatal(err)
	}

	small := "```mermaid\nflowchart TD\n  A[CBIN-042] --> B[other]\n```\n"
	if err := os.WriteFile(filepath.Join(root, "small.md"), []byte(small), 0o600); err != nil {
		t.Fatal(err)
	}

	refs, issues, err := ScanDiagramRefs(root, DefaultSkipRegex(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range refs {
		if r.File == "big.md" {
			t.Errorf("refs from oversized file were read: %+v", r)
		}
	}
	got := map[string]bool{}
	for _, r := range refs {
		got[r.ReqID] = true
	}
	if !got["CBIN-042"] {
		t.Errorf("refs = %v, want small.md's CBIN-042 to still be scanned", refs)
	}

	if !hasIssue(issues, "big.md", IssueFileTooLarge) {
		t.Errorf("issues = %+v, want a file_too_large issue for big.md", issues)
	}
}
