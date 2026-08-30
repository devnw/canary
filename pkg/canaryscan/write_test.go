package canaryscan

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleReport builds a Report whose cells exercise the two properties CSV
// output must hold: a formula-shaped value (leading '=') that a spreadsheet
// would execute, and an embedded comma and quote that a hand-rolled writer
// would split into the wrong number of columns.
func sampleReport() Report {
	return Report{
		Requirements: []Requirement{
			{
				ID: "CBIN-001",
				Features: []Feature{
					{
						Feature: "=HYPERLINK(\"http://evil\")",
						Aspect:  "API",
						Status:  "IMPL",
						Files:   []string{"a.go", "b.go"},
						Tests:   []string{"TestA", "TestB"},
						Benches: []string{"BenchA"},
						Owner:   "va,l\"ue",
						Updated: "2026-01-01",
					},
				},
			},
		},
	}
}

// TestWriteCSV_InjectionGuardAndRoundTrip proves every value that could be read
// as a spreadsheet formula is neutralized, and that the output is decodable by
// an independent csv.Reader without error -- the hand-rolled Fprintf writer it
// replaces produced neither guarantee.
func TestWriteCSV_InjectionGuardAndRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "out.csv")
	if err := WriteCSV(p, sampleReport()); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	f, err := os.Open(p) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = f.Close() }()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("csv not standards-decodable: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("want header + at least one data row, got %d rows", len(rows))
	}

	for _, row := range rows[1:] {
		for _, cell := range row {
			if len(cell) > 0 && strings.ContainsRune("=+-@", rune(cell[0])) {
				t.Fatalf("unguarded formula cell %q", cell)
			}
		}
	}

	// One row per file: the two-file feature yields two rows, and the tests
	// travel together in one cell rather than being zipped against files.
	if len(rows) != 3 {
		t.Fatalf("want 3 rows (header + 2 files), got %d: %v", len(rows), rows)
	}
	// tests column (index 5) carries both tests joined, decoded back intact.
	if got := rows[1][5]; got != "TestA|TestB" {
		t.Fatalf("tests cell = %q, want TestA|TestB", got)
	}
	// The comma-and-quote owner value survives the round trip.
	if got := rows[1][7]; got != "va,l\"ue" {
		t.Fatalf("owner cell = %q, want va,l\"ue", got)
	}
}

// TestWriteCSV_PropagatesWriteError proves a write to an unwritable path is
// reported, not swallowed: the target directory does not exist.
func TestWriteCSV_PropagatesWriteError(t *testing.T) {
	p := filepath.Join(t.TempDir(), "no", "such", "dir.csv")
	if err := WriteCSV(p, sampleReport()); err == nil {
		t.Fatal("WriteCSV to a nonexistent directory must fail")
	}
}

// TestWriteJSON_PropagatesWriteError proves the same for JSON: a short or
// failed write is an error the caller sees.
func TestWriteJSON_PropagatesWriteError(t *testing.T) {
	p := filepath.Join(t.TempDir(), "no", "such", "dir.json")
	if err := WriteJSON(p, sampleReport()); err == nil {
		t.Fatal("WriteJSON to a nonexistent directory must fail")
	}
}
