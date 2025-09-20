package core

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Token line regex (comment styles optional at start)
var tokenRe = regexp.MustCompile(`(?m)^\s*(?://|#|--)\s*CANARY:\s*(.*)$`)
var kvRe = regexp.MustCompile(`\s*([^=;\s]+)\s*=\s*([^;]+)\s*`)

// DefaultSkipPattern is the canonical skip regex for directories.
const DefaultSkipPattern = `(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)(/|$)`

// Report represents canonical output.
type Report struct {
	GeneratedAt  string        `json:"generated_at"`
	Requirements []Requirement `json:"requirements"`
	Summary      Summary       `json:"summary"`
}

type Requirement struct {
	ID       string    `json:"id"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Feature string   `json:"feature"`
	Aspect  string   `json:"aspect"`
	Status  string   `json:"status"`
	Files   []string `json:"files"`
	Tests   []string `json:"tests"`
	Benches []string `json:"benches"`
	Owner   string   `json:"owner,omitempty"`
	Updated string   `json:"updated,omitempty"`
}

type Summary struct {
	ByStatus           map[string]int `json:"by_status"`
	ByAspect           map[string]int `json:"by_aspect"`
	TotalTokens        int            `json:"total_tokens"`
	UniqueRequirements int            `json:"unique_requirements"`
}

// ScanOptions define scanning parameters.
type ScanOptions struct {
	Root string
	Skip *regexp.Regexp
}

// Scan walks files and extracts tokens.
func Scan(opts ScanOptions) (Report, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Skip == nil {
		opts.Skip = regexp.MustCompile(DefaultSkipPattern)
	}

	type key struct{ req, feature, aspect, status, owner, updated string }
	type aggVal struct{ files, tests, benches map[string]struct{} }
	agg := map[key]*aggVal{}

	err := filepath.WalkDir(opts.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := path
		if opts.Skip.MatchString(rel) && d.IsDir() {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if opts.Skip.MatchString(rel) {
			return nil
		}
		// simple binary skip heuristic
		if isBinary(path) {
			return nil
		}
		// debug logging removed
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 1024*1024)
		sc.Buffer(buf, 1024*1024) // allow up to 1MB lines
		for sc.Scan() {
			line := sc.Text()
			m := tokenRe.FindStringSubmatch(line)
			if len(m) != 2 {
				continue
			}
			// token matched
			fields, perr := parseKV(m[1])
			if perr != nil {
				return fmt.Errorf("parse %s: %w", path, perr)
			}
			k := key{
				req:     normalizeREQ(fields["REQ"]),
				feature: unquote(fields["FEATURE"]),
				aspect:  fields["ASPECT"],
				status:  fields["STATUS"],
				owner:   fields["OWNER"],
				updated: fields["UPDATED"],
			}
			if k.req == "" || k.feature == "" || k.aspect == "" || k.status == "" || k.updated == "" {
				return fmt.Errorf("missing required fields in token at %s", path)
			}
			a := agg[k]
			if a == nil {
				a = &aggVal{files: map[string]struct{}{}, tests: map[string]struct{}{}, benches: map[string]struct{}{}}
				agg[k] = a
			}
			a.files[path] = struct{}{}
			for _, t := range splitList(fields["TEST"]) {
				if t != "" {
					a.tests[t] = struct{}{}
				}
			}
			for _, b := range splitList(fields["BENCH"]) {
				if b != "" {
					a.benches[b] = struct{}{}
				}
			}
		}
		return sc.Err()
	})
	if err != nil {
		return Report{}, err
	}

	// aggregate
	byReq := map[string][]Feature{}
	sumStatus := map[string]int{"MISSING": 0, "STUB": 0, "IMPL": 0, "TESTED": 0, "BENCHED": 0, "REMOVED": 0}
	sumAspect := map[string]int{}
	totalTokens := 0

	for k, v := range agg {
		status := promoteStatus(k.status, len(v.tests) > 0, len(v.benches) > 0)
		feat := Feature{
			Feature: k.feature,
			Aspect:  k.aspect,
			Status:  status,
			Files:   keys(v.files),
			Tests:   keys(v.tests),
			Benches: keys(v.benches),
			Owner:   k.owner,
			Updated: k.updated,
		}
		byReq[k.req] = append(byReq[k.req], feat)
		sumStatus[status]++
		sumAspect[k.aspect]++
		totalTokens++
	}
	var reqs []Requirement
	for id, feats := range byReq {
		sort.Slice(feats, func(i, j int) bool { return feats[i].Feature+feats[i].Aspect < feats[j].Feature+feats[j].Aspect })
		reqs = append(reqs, Requirement{ID: id, Features: feats})
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })

	rep := Report{GeneratedAt: time.Now().UTC().Format(time.RFC3339), Requirements: reqs, Summary: Summary{ByStatus: sumStatus, ByAspect: sumAspect, TotalTokens: totalTokens, UniqueRequirements: len(reqs)}}
	return rep, nil
}

func promoteStatus(status string, hasTests, hasBenches bool) string {
	if status == "IMPL" && hasTests {
		status = "TESTED"
	}
	if (status == "IMPL" || status == "TESTED") && hasBenches {
		status = "BENCHED"
	}
	return status
}

func parseKV(s string) (map[string]string, error) {
	out := map[string]string{}
	for _, seg := range strings.Split(s, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		m := kvRe.FindStringSubmatch(seg)
		if len(m) != 3 {
			return nil, fmt.Errorf("bad kv segment %q", seg)
		}
		k := strings.ToUpper(m[1])
		v := strings.TrimSpace(m[2])
		out[k] = v
	}
	// enum validation (ASPECT not strictly enumerated here, STATUS validated)
	if st := out["STATUS"]; st != "MISSING" && st != "STUB" && st != "IMPL" && st != "TESTED" && st != "BENCHED" && st != "REMOVED" {
		return nil, fmt.Errorf("invalid STATUS %q", st)
	}
	return out, nil
}

func splitList(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == ',' })
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
		return v[1 : len(v)-1]
	}
	return v
}
func isBinary(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".pdf", ".zip", ".gz", ".tgz", ".jar", ".wasm", ".exe", ".dll", ".so", ".dylib":
		return true
	}
	return false
}
func normalizeREQ(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "‑", "-") // normalize odd hyphen
	v = strings.ReplaceAll(v, "–", "-")
	// Zero pad if pattern like CBIN-<n>
	if m := regexp.MustCompile(`^(CBIN-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		n := m[2]
		for len(n) < 3 {
			n = "0" + n
		}
		return m[1] + n
	}
	return v
}

// WriteJSON writes minified canonical JSON (key order relies on struct order & maps pre-filled).
func WriteJSON(path string, rep Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	// minified => no Indent
	return enc.Encode(rep)
}

// WriteCSV emits exploded rows.
func WriteCSV(path string, rep Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"req", "feature", "aspect", "status", "file", "test", "bench", "owner", "updated"})
	for _, r := range rep.Requirements {
		for _, ft := range r.Features {
			base := []string{r.ID, ft.Feature, ft.Aspect, ft.Status, "", "", "", ft.Owner, ft.Updated}
			max := max3(len(ft.Files), len(ft.Tests), len(ft.Benches))
			if max == 0 {
				_ = w.Write(base)
				continue
			}
			for i := 0; i < max; i++ {
				row := append([]string{}, base...)
				if i < len(ft.Files) {
					row[4] = ft.Files[i]
				}
				if i < len(ft.Tests) {
					row[5] = ft.Tests[i]
				}
				if i < len(ft.Benches) {
					row[6] = ft.Benches[i]
				}
				if err := w.Write(row); err != nil {
					return err
				}
			}
		}
	}
	return nil
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

// Staleness check (30d window default supplied by caller).
func CheckStaleness(rep Report, maxAge time.Duration) ([]string, error) {
	var stale []string
	cut := time.Now().UTC().Add(-maxAge)
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				t, err := time.Parse("2006-01-02", f.Updated)
				if err != nil {
					return nil, fmt.Errorf("bad UPDATED %s:%s", r.ID, f.Updated)
				}
				if t.Before(cut) {
					age := int(time.Since(t).Hours() / 24)
					stale = append(stale, fmt.Sprintf("CANARY_STALE REQ=%s updated=%s age_days=%d threshold=%d", r.ID, f.Updated, age, int(maxAge.Hours()/24)))
				}
			}
		}
	}
	return stale, nil
}

// Verify claims: lines like '✅ CBIN-101'.
var claimRe = regexp.MustCompile(`(?m)^\s*✅\s+(CBIN-\d{3})\b`)

func VerifyClaims(rep Report, gapContent string) ([]string, error) {
	matches := claimRe.FindAllStringSubmatch(gapContent, -1)
	claimed := map[string]struct{}{}
	for _, m := range matches {
		claimed[m[1]] = struct{}{}
	}
	evidence := map[string]bool{}
	for _, r := range rep.Requirements {
		ok := false
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				ok = true
				break
			}
		}
		evidence[r.ID] = ok
	}
	var errs []string
	for id := range claimed {
		if !evidence[id] {
			errs = append(errs, fmt.Sprintf("CANARY_VERIFY_FAIL REQ=%s reason=claimed_but_not_TESTED_OR_BENCHED", id))
		}
	}
	return errs, nil
}

// LoadFile helper.
func LoadFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComposeAndWrite writes outputs returning diagnostics (verify + stale).
func ComposeAndWrite(rep Report, outJSON, outCSV string) error {
	if err := WriteJSON(outJSON, rep); err != nil {
		return err
	}
	if outCSV != "" {
		if err := WriteCSV(outCSV, rep); err != nil {
			return err
		}
	}
	return nil
}

// High level convenience: Scan + write outputs.
func ScanAndWrite(opts ScanOptions, outJSON, outCSV string) (Report, error) {
	rep, err := Scan(opts)
	if err != nil {
		return Report{}, err
	}
	if err := ComposeAndWrite(rep, outJSON, outCSV); err != nil {
		return Report{}, err
	}
	return rep, nil
}

// Error aggregation helper.
func AggregateErrors(diags ...[]string) error {
	var all []string
	for _, d := range diags {
		all = append(all, d...)
	}
	if len(all) == 0 {
		return nil
	}
	return errors.New(strings.Join(all, "\n"))
}
