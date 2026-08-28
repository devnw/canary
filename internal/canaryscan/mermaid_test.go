package canaryscan

import (
	"os"
	"path/filepath"
	"reflect"
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
	rep, err := Scan(root, DefaultSkipRegex(), nil, nil)
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
