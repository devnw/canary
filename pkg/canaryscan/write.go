package canaryscan

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devnw.dev/canary/pkg/safewrite"
)

// writeOut replaces path atomically. These are output files the user asked
// for by name (--out, --csv), so replacing one is the point; what must never
// happen is a truncated report left behind by a failure mid-write.
//
// The target directory must already exist: a report writer names a file in a
// directory, it does not invent a directory tree. Requiring the parent means a
// mistyped or unwritable destination is reported as an error instead of
// silently materializing an empty tree, and it keeps the write-failure path
// observable to callers (safewrite would otherwise MkdirAll the parent away).
func writeOut(path string, data []byte) error {
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("write %s: output directory %s does not exist", path, dir)
	}
	_, err := safewrite.Write(path, data, 0o640, safewrite.Options{
		Root:  dir,
		Force: true,
	})
	return err
}

// csvGuard defuses CSV formula injection. A spreadsheet treats a cell whose
// first character is =, +, -, or @ as a formula, so a token value that begins
// with one of those bytes could execute when the report is opened. Prefixing a
// single quote makes the cell a literal string without changing what a
// standards-compliant CSV reader decodes back (the quote is part of the value).
func csvGuard(s string) string {
	if len(s) > 0 && strings.ContainsRune("=+-@", rune(s[0])) {
		return "'" + s
	}
	return s
}

// WriteJSON writes rep to path as JSON, atomically. The encode targets an
// in-memory buffer (which cannot short-write), and the file is committed by
// writeOut, whose error -- a failed create, write, or close -- is returned:
// a partial or unwritable report never passes silently.
func WriteJSON(path string, rep Report) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(rep); err != nil {
		return err
	}
	return writeOut(path, buf.Bytes())
}

// WriteCSV writes rep to path as CSV, atomically, one row per
// (requirement, feature, file). A feature's tests and benches are joined with
// "|" inside single cells rather than zipped column-wise against the file
// list: the old zip paired unrelated files, tests, and benches by shared row
// index, asserting adjacencies that never existed. Every cell is passed
// through csvGuard so no value can be interpreted as a spreadsheet formula,
// and encoding/csv quotes any value containing a comma, quote, or newline so
// the output is decodable by an independent reader.
func WriteCSV(path string, rep Report) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{
		"req", "feature", "aspect", "status", "file", "tests", "benches", "owner", "updated",
	}); err != nil {
		return err
	}

	for _, r := range rep.Requirements {
		for _, ft := range r.Features {
			tests := strings.Join(ft.Tests, "|")
			benches := strings.Join(ft.Benches, "|")

			writeRow := func(file string) error {
				return w.Write([]string{
					csvGuard(r.ID),
					csvGuard(ft.Feature),
					csvGuard(ft.Aspect),
					csvGuard(ft.Status),
					csvGuard(file),
					csvGuard(tests),
					csvGuard(benches),
					csvGuard(ft.Owner),
					csvGuard(ft.Updated),
				})
			}

			if len(ft.Files) == 0 {
				if err := writeRow(""); err != nil {
					return err
				}
				continue
			}
			for _, file := range ft.Files {
				if err := writeRow(file); err != nil {
					return err
				}
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return writeOut(path, buf.Bytes())
}
