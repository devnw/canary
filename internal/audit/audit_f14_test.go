// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/cmds/next"
	"devnw.dev/canary/pkg/storage"
)

// tuple is the canonical, source-independent identity of a scanned token: the
// five fields every public parse path must agree on.
type tuple struct{ req, feature, aspect, status, updated string }

// requiredKeys is the field set a token must declare to be counted, matching
// canaryscan.Scan / ScanTokenRecords.
var requiredKeys = []string{"REQ", "FEATURE", "ASPECT", "STATUS", "UPDATED"}

// TestAuditF14 is the F-14 acceptance test: every public path that reads CANARY
// tokens must route through the one canaryscan parser and therefore produce an
// identical set of (req, feature, aspect, status, updated) tuples over the same
// corpus. The corpus in internal/audit/corpus/ covers every documented token
// form -- //, #, /*, -- prefixes; a bare CANARY:; a legacy bare requirement id;
// the BUG= alias; quoted values carrying ; " and unicode; a canonical
// upgrade-output token; a multi-token file; and a CRLF file.
//
// Paths exercised:
//   - canaryscan.Scan (the report) -- the canonical set.
//   - canaryscan.ParseTokenLine, applied per source line. This is also the
//     exact primitive pkg/upgrade's pre-parse uses (upgrade.tokenShapeOf calls
//     canaryscan.ParseTokenLine), so proving it agrees with Scan proves the
//     upgrade pre-parse agrees too.
//   - `canary index` -> database rows, read back through pkg/storage.
//   - `canary next`'s filesystem source (next.SelectNext with no index), whose
//     selected token must be a member of the canonical set.
func TestAuditF14(t *testing.T) {
	corpus := filepath.Join(repoRoot(), "internal", "audit", "corpus")

	// Canonical set: canaryscan.Scan.
	scanSet := scanTuples(t, corpus)
	if len(scanSet) == 0 {
		t.Fatal("corpus produced no tokens via canaryscan.Scan")
	}

	// Path 2: ParseTokenLine per source line (also the upgrade pre-parse
	// primitive).
	lineSet := parseLineTuples(t, corpus)
	assertSameSet(t, "canaryscan.Scan", scanSet, "ParseTokenLine-per-line", lineSet)

	// Path 3: `canary index` -> DB rows.
	dbSet := indexTuples(t, corpus)
	assertSameSet(t, "canaryscan.Scan", scanSet, "canary index (DB rows)", dbSet)

	// Path 4: `canary next` filesystem source -- the selected token must be a
	// member of the canonical set (next scans the same tree with the same
	// parser).
	tmp := t.TempDir()
	copyCorpus(t, corpus, tmp)
	tok, _, _, err := next.SelectNext(filepath.Join(tmp, "does-not-exist.db"), tmp, "", nil, true, io.Discard)
	if err != nil {
		t.Fatalf("next.SelectNext (filesystem): %v", err)
	}
	if tok != nil {
		got := tuple{tok.ReqID, tok.Feature, tok.Aspect, tok.Status, tok.UpdatedAt}
		if _, ok := scanSet[got]; !ok {
			t.Errorf("next filesystem source selected %v, which is not in the canonical set", got)
		}
	}
}

// scanTuples returns the tuple set produced by canaryscan.Scan.
func scanTuples(t *testing.T, root string) map[tuple]struct{} {
	t.Helper()
	rep := scanDir(t, root)
	set := map[tuple]struct{}{}
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			set[tuple{r.ID, f.Feature, f.Aspect, f.Status, f.Updated}] = struct{}{}
		}
	}
	return set
}

// parseLineTuples reads every corpus file and derives tuples by feeding each
// source line to canaryscan.ParseTokenLine, then reproducing the same field
// rules Scan applies (upper-cased keys, BUG= aliased to REQ, all five required
// fields present). CRLF is stripped, matching the scanner.
func parseLineTuples(t *testing.T, root string) map[tuple]struct{} {
	t.Helper()
	set := map[tuple]struct{}{}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read corpus dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(root, e.Name()))
		if rerr != nil {
			t.Fatalf("read %s: %v", e.Name(), rerr)
		}
		for _, raw := range strings.Split(string(b), "\n") {
			line := strings.TrimRight(raw, "\r")
			fields, ok, perr := canaryscan.ParseTokenLine(line)
			if !ok || perr != nil {
				continue
			}
			m := map[string]string{}
			for _, f := range fields {
				m[strings.ToUpper(f.Key)] = f.Value
			}
			if m["REQ"] == "" && m["BUG"] != "" {
				m["REQ"] = m["BUG"]
			}
			complete := true
			for _, k := range requiredKeys {
				if m[k] == "" {
					complete = false
					break
				}
			}
			if !complete {
				continue
			}
			set[tuple{m["REQ"], m["FEATURE"], m["ASPECT"], m["STATUS"], m["UPDATED"]}] = struct{}{}
		}
	}
	return set
}

// indexTuples runs `canary index` over a copy of the corpus and reads the
// resulting rows back through pkg/storage.
func indexTuples(t *testing.T, corpus string) map[tuple]struct{} {
	t.Helper()
	bin := buildCanary(t)
	tmp := t.TempDir()
	copyCorpus(t, corpus, tmp)
	dbPath := filepath.Join(tmp, "canary.db")
	run(t, tmp, bin, "index", "--root", tmp, "--db", dbPath)

	db, err := storage.OpenRO(dbPath)
	if err != nil {
		t.Fatalf("open indexed db: %v", err)
	}
	defer db.Close()
	toks, err := db.ListTokens("", map[string]any{"include_hidden": "true"}, "", "", 0)
	if err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	set := map[tuple]struct{}{}
	for _, tok := range toks {
		set[tuple{tok.ReqID, tok.Feature, tok.Aspect, tok.Status, tok.UpdatedAt}] = struct{}{}
	}
	return set
}

// copyCorpus copies the corpus files (flat, no subdirs) into dst.
func copyCorpus(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(src, e.Name()))
		if rerr != nil {
			t.Fatalf("read %s: %v", e.Name(), rerr)
		}
		if werr := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o600); werr != nil {
			t.Fatalf("write %s: %v", e.Name(), werr)
		}
	}
}

// assertSameSet fails with a readable diff when two tuple sets disagree.
func assertSameSet(t *testing.T, aName string, a map[tuple]struct{}, bName string, b map[tuple]struct{}) {
	t.Helper()
	if len(a) != len(b) {
		t.Errorf("%s has %d tuples, %s has %d", aName, len(a), bName, len(b))
	}
	var onlyA, onlyB []string
	for k := range a {
		if _, ok := b[k]; !ok {
			onlyA = append(onlyA, fmtTuple(k))
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			onlyB = append(onlyB, fmtTuple(k))
		}
	}
	sort.Strings(onlyA)
	sort.Strings(onlyB)
	if len(onlyA) > 0 {
		t.Errorf("tuples in %s but not %s:\n  %s", aName, bName, strings.Join(onlyA, "\n  "))
	}
	if len(onlyB) > 0 {
		t.Errorf("tuples in %s but not %s:\n  %s", bName, aName, strings.Join(onlyB, "\n  "))
	}
}

func fmtTuple(x tuple) string {
	return strings.Join([]string{x.req, x.feature, x.aspect, x.status, x.updated}, " | ")
}
