package gate

// CANARY: REQ=ENG-4324; FEATURE="GenericScanner"; ASPECT=Engine; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-02
import (
	"bufio"
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

// ScanResult represents the generic scan output independent of root package types.
type ScanResult struct {
	Requirements []RequirementTokens
	Summary      SummaryTokens
}

// RequirementTokens groups feature tokens by requirement ID.
type RequirementTokens struct {
	ID       string
	Features []FeatureTokens
}

// FeatureTokens describes one feature/aspect/status token aggregation.
type FeatureTokens struct {
	Feature string
	Aspect  string
	Status  string
	Files   []string
	Tests   []string
	Benches []string
	Owner   string
	Updated string
}

// SummaryTokens holds aggregate counts.
type SummaryTokens struct {
	ByStatus map[string]int
	ByAspect map[string]int
}

// ScannerOption configures Scanner behavior.
type ScannerOption func(*scannerConfig)

type scannerConfig struct {
	skipDirs map[string]bool
}

// WithSkipDirs adds directories to skip.
func WithSkipDirs(dirs ...string) ScannerOption {
	return func(c *scannerConfig) {
		for _, d := range dirs {
			c.skipDirs[d] = true
		}
	}
}

// NewScanner returns a configured Scanner.
func NewScanner(opts ...ScannerOption) *Scanner {
	c := scannerConfig{skipDirs: map[string]bool{".git": true, "node_modules": true, "vendor": true, "bin": true, "zig-out": true, ".zig-cache": true}}
	for _, o := range opts {
		o(&c)
	}
	return &Scanner{cfg: c}
}

// Scanner performs repository scans for CANARY tokens.
type Scanner struct{ cfg scannerConfig }

// token line regex (duplicated from root for now; could be unified later)
var scannerCanaryRe = regexp.MustCompile(`^\s*(?://|#|--|\[//\]:\s*#|<!--|/\*+|\*)[\s\*]*CANARY:\s*(.*)$`)
var scannerKVRe = regexp.MustCompile(`\s*([^=;\s]+)\s*=\s*([^;]+)\s*`)

// ScanRepository scans root and returns ScanResult.
func (s *Scanner) ScanRepository(root string) (ScanResult, error) {
	type key struct{ id, feature, aspect, status, owner, updated string }
	type val struct{ files, tests, benches map[string]struct{} }
	agg := map[key]*val{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if s.cfg.skipDirs[filepath.Base(path)] {
				return filepath.SkipDir
			}
			return nil
		}
		if isBinaryExt(path) {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer func() { _ = f.Close() }()
		sc := bufio.NewScanner(f)
		const maxSize = 1024 * 1024
		buf := make([]byte, maxSize)
		sc.Buffer(buf, maxSize)
		ln := 0
		for sc.Scan() {
			ln++
			m := scannerCanaryRe.FindStringSubmatch(sc.Text())
			if len(m) < 2 {
				continue
			}
			// A MIGRATE line is free-text guidance, not a KV token, and must
			// never reach parseKV (a MIGRATE line must never abort a scan).
			if strings.HasPrefix(strings.TrimSpace(m[1]), "MIGRATE") {
				continue
			}
			segment := normalizeSegment(m[1])
			rec, perr := parseKV(segment)
			if perr != nil {
				return fmt.Errorf("parse %s:%d: %w", path, ln, perr)
			}
			// If parseKV returned empty map (placeholder skip), ignore this token
			if len(rec) == 0 {
				continue
			}
			k := key{id: rec["REQ"], feature: unquoteValue(rec["FEATURE"]), aspect: rec["ASPECT"], status: rec["STATUS"], owner: rec["OWNER"], updated: rec["UPDATED"]}
			if k.id == "" || k.aspect == "" || k.status == "" {
				return fmt.Errorf("missing required fields in %s:%d", path, ln)
			}
			if _, ok := agg[k]; !ok {
				agg[k] = &val{files: map[string]struct{}{}, tests: map[string]struct{}{}, benches: map[string]struct{}{}}
			}
			agg[k].files[path] = struct{}{}
			for _, t := range split(rec["TEST"]) {
				if t != "" {
					agg[k].tests[t] = struct{}{}
				}
			}
			for _, b := range split(rec["BENCH"]) {
				if b != "" {
					agg[k].benches[b] = struct{}{}
				}
			}
		}
		return nil
	})
	if err != nil {
		return ScanResult{}, err
	}

	reqMap := map[string][]FeatureTokens{}
	byStatus := map[string]int{}
	byAspect := map[string]int{}
	for k, v := range agg {
		// STATUS is a declaration: TEST=/BENCH= are recorded as evidence
		// references but never change it.
		status := k.status
		ft := FeatureTokens{Feature: k.feature, Aspect: k.aspect, Status: status, Files: keysList(v.files), Tests: keysList(v.tests), Benches: keysList(v.benches), Owner: k.owner, Updated: k.updated}
		id := normalizeRequirement(k.id)
		reqMap[id] = append(reqMap[id], ft)
		byStatus[status]++
		byAspect[k.aspect]++
	}
	var reqs []RequirementTokens
	for id, feats := range reqMap {
		sort.Slice(feats, func(i, j int) bool {
			if feats[i].Feature == feats[j].Feature {
				return feats[i].Aspect < feats[j].Aspect
			}
			return feats[i].Feature < feats[j].Feature
		})
		reqs = append(reqs, RequirementTokens{ID: id, Features: feats})
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })
	return ScanResult{Requirements: reqs, Summary: SummaryTokens{ByStatus: byStatus, ByAspect: byAspect}}, nil
}

// Helper / internal routines (duplicated minimal logic)
func parseKV(s string) (map[string]string, error) {
	s = strings.TrimSuffix(strings.TrimSpace(s), "-->")
	s = strings.TrimSpace(s)
	out := map[string]string{}
	// Skip placeholder/example tokens containing angle-bracket placeholders
	if strings.ContainsAny(s, "<>") || strings.Contains(s, "{{") || strings.Contains(s, "}}") || strings.Contains(s, "%s") {
		return map[string]string{}, nil
	}
	// Accept bare legacy ID-only segments for REQ, TASK, BUG forms (optionally namespaced like REQ-GQL-4)
	legacyReqRe := regexp.MustCompile(`^((?:REQ|TASK|BUG)(?:-[A-Z]+)?-?\d{1,4})$`)
	for _, seg := range strings.Split(s, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		// Legacy ID-only segment (e.g. REQ-1 or REQ-GQL-42)
		if !strings.Contains(seg, "=") && legacyReqRe.MatchString(seg) {
			// Treat TASK-/BUG- prefixed IDs as requirement identifiers for aggregation purposes
			out["REQ"] = normalizeRequirement(seg)
			continue
		}
		m := scannerKVRe.FindStringSubmatch(seg)
		if len(m) != 3 {
			return nil, fmt.Errorf("bad kv segment %q", seg)
		}
		out[strings.ToUpper(strings.TrimSpace(m[1]))] = strings.TrimSpace(m[2])
	}
	status := strings.ToUpper(out["STATUS"])
	switch status {
	case "", "MISSING", "STUB", "IMPL", "TESTED", "BENCHED", "REMOVED", "FIXED", "OPEN":
		// valid (empty means optional)
	default:
		return nil, fmt.Errorf("invalid STATUS %q", status)
	}
	return out, nil
}
func normalizeSegment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "*/")
	s = strings.TrimSuffix(s, "-->")
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "*") {
		s = strings.TrimLeft(s, "*")
		s = strings.TrimSpace(s)
	}
	return s
}
func unquoteValue(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
		return v[1 : len(v)-1]
	}
	return v
}
func split(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == ',' })
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
func keysList(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func isBinaryExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".pdf", ".zip", ".gz", ".tgz", ".jar", ".wasm", ".exe", ".dll", ".so", ".dylib", ".db", ".db-wal", ".db-shm":
		return true
	}
	if ext == "" && strings.Contains(path, "zig-out") {
		return true
	}
	return false
}
func normalizeRequirement(id string) string {
	id = strings.TrimSpace(id)
	id = strings.ReplaceAll(id, "‑", "-")
	id = strings.ReplaceAll(id, "–", "-")
	// Zero-pad common legacy forms: REQ-<n>, REQ-GQL-<n>, REQ-<ALPHA>-<n>
	legacyPad := regexp.MustCompile(`^(REQ(?:-[A-Z]+)?-)(\d{1,3})$`)
	if m := legacyPad.FindStringSubmatch(id); len(m) == 3 {
		num := m[2]
		for len(num) < 3 {
			num = "0" + num
		}
		return m[1] + num
	}
	cbinPad := regexp.MustCompile(`^(CBIN-)(\d{1,3})$`)
	if m := cbinPad.FindStringSubmatch(id); len(m) == 3 {
		num := m[2]
		for len(num) < 3 {
			num = "0" + num
		}
		return m[1] + num
	}
	// Pad TASK and BUG numeric suffixes
	genericPad := regexp.MustCompile(`^((?:TASK|BUG)-)(\d{1,3})$`)
	if m := genericPad.FindStringSubmatch(id); len(m) == 3 {
		num := m[2]
		for len(num) < 3 {
			num = "0" + num
		}
		return m[1] + num
	}
	return id
}

// CheckStalenessTokens validates staleness directly on ScanResult.
func CheckStalenessTokens(res ScanResult, durSeconds int64) error {
	cut := time.Now().UTC().Add(-time.Duration(durSeconds) * time.Second)
	var errs []string
	for _, r := range res.Requirements {
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				if f.Updated == "" {
					errs = append(errs, fmt.Sprintf("%s %s missing UPDATED", r.ID, f.Feature))
					continue
				}
				t, err := time.Parse("2006-01-02", f.Updated)
				if err != nil {
					errs = append(errs, fmt.Sprintf("%s %s bad UPDATED=%q", r.ID, f.Feature, f.Updated))
					continue
				}
				if t.Before(cut) {
					errs = append(errs, fmt.Sprintf("REQ=%s feature=%s UPDATED=%s stale", r.ID, f.Feature, f.Updated))
				}
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
