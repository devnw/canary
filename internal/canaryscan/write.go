package canaryscan

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func max3(a, b, c int) int {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}

// WriteJSON writes rep to path as JSON.
func WriteJSON(path string, rep Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	return enc.Encode(rep)
}

// WriteCSV writes rep to path as CSV.
func WriteCSV(path string, rep Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, "req,feature,aspect,status,file,test,bench,owner,updated")
	for _, r := range rep.Requirements {
		for _, ft := range r.Features {
			max := max3(len(ft.Files), len(ft.Tests), len(ft.Benches))
			if max == 0 {
				fmt.Fprintf(f, "%s,%s,%s,%s,,,,%s,%s\n", r.ID, ft.Feature, ft.Aspect, ft.Status, ft.Owner, ft.Updated)
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
				fmt.Fprintf(f, "%s,%s,%s,%s,%s,%s,%s,%s,%s\n", r.ID, ft.Feature, ft.Aspect, ft.Status, file, test, bench, ft.Owner, ft.Updated)
			}
		}
	}
	return nil
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
		b.WriteString(fmt.Sprintf("%q:%d", k, m[k]))
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}
