package canaryscan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"devnw.dev/canary/pkg/safewrite"
)

// writeOut replaces path atomically. These are output files the user asked
// for by name (--out, --csv), so replacing one is the point; what must never
// happen is a truncated report left behind by a failure mid-write.
func writeOut(path string, data []byte) error {
	_, err := safewrite.Write(path, data, 0o640, safewrite.Options{
		Root:  filepath.Dir(path),
		Force: true,
	})
	return err
}

func max3(a, b, c int) int {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}

// WriteJSON writes rep to path as JSON, atomically.
func WriteJSON(path string, rep Report) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(rep); err != nil {
		return err
	}
	return writeOut(path, buf.Bytes())
}

// WriteCSV writes rep to path as CSV, atomically.
func WriteCSV(path string, rep Report) error {
	f := &bytes.Buffer{}
	_, _ = fmt.Fprintln(f, "req,feature,aspect,status,file,test,bench,owner,updated")
	for _, r := range rep.Requirements {
		for _, ft := range r.Features {
			max := max3(len(ft.Files), len(ft.Tests), len(ft.Benches))
			if max == 0 {
				_, _ = fmt.Fprintf(f, "%s,%s,%s,%s,,,,%s,%s\n", r.ID, ft.Feature, ft.Aspect, ft.Status, ft.Owner, ft.Updated)
				continue
			}
			for i := 0; i < max; i++ {
				file, test, bench := "", "", ""
				if i < len(ft.Files) {
					file = ft.Files[i]
				}
				if i < len(ft.Tests) {
					test = ft.Tests[i]
				}
				if i < len(ft.Benches) {
					bench = ft.Benches[i]
				}
				_, _ = fmt.Fprintf(f, "%s,%s,%s,%s,%s,%s,%s,%s,%s\n", r.ID, ft.Feature, ft.Aspect, ft.Status, file, test, bench, ft.Owner, ft.Updated)
			}
		}
	}
	return writeOut(path, f.Bytes())
}

// MarshalSortedMap ensures deterministic JSON object key order for map[string]int.
func MarshalSortedMap(m map[string]int) ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		_, _ = fmt.Fprintf(&b, "%q:%d", k, m[k])
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}
