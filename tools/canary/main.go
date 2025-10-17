// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


package main

// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=BENCHED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-10-15

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
	"gopkg.in/yaml.v3"
)

type StatusCounts map[string]int

func (m StatusCounts) MarshalJSON() ([]byte, error) { return marshalSortedMap(m) }

type AspectCounts map[string]int

func (m AspectCounts) MarshalJSON() ([]byte, error) { return marshalSortedMap(m) }

// ProjectConfig defines project-specific configuration
type ProjectConfig struct {
	Project struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"project"`
	Requirements struct {
		IDPattern string `yaml:"id_pattern"`
	} `yaml:"requirements"`
	Scanner struct {
		ExcludePaths []string `yaml:"exclude_paths"`
	} `yaml:"scanner"`
}

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
	Updated string   `json:"updated"`
}
type Summary struct {
	ByStatus           StatusCounts `json:"by_status"`
	ByAspect           AspectCounts `json:"by_aspect"`
	TotalTokens        int          `json:"total_tokens"`
	UniqueRequirements int          `json:"unique_requirements"`
}

var (
	// Allow optional comment markers before CANARY (//, #, or /*) to permit embedding in source files.
	tokenLineRe = regexp.MustCompile(`(?m)^[ \t]*(?:\/\/|#|\/\*)?[ \t]*CANARY:\s*(.*)$`)
	kvRe        = regexp.MustCompile(`\s*([A-Za-z_]+)\s*=\s*([^;]+)\s*`)
	claimRe     = regexp.MustCompile(`(?m)^\s*✅\s+(CBIN-\d{3})\b`)
	skipDefault = regexp.MustCompile(`(^|/)(.git|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)(/|$)`)
	aspects     = map[string]struct{}{"API": {}, "CLI": {}, "Engine": {}, "Planner": {}, "Storage": {}, "Wire": {}, "Security": {}, "Docs": {}, "Decode": {}, "Encode": {}, "RoundTrip": {}, "Bench": {}, "FrontEnd": {}, "Dist": {}}
	statuses    = []string{"MISSING", "STUB", "IMPL", "TESTED", "BENCHED", "REMOVED"}
	statusSet   = func() map[string]struct{} {
		m := map[string]struct{}{}
		for _, s := range statuses {
			m[s] = struct{}{}
		}
		return m
	}()
)

func main() {
	root := flag.String("root", ".", "root directory to scan")
	outJSON := flag.String("out", "status.json", "output status.json path")
	outCSV := flag.String("csv", "", "optional status.csv path")
	verifyPath := flag.String("verify", "", "GAP_ANALYSIS file to verify claims")
	strict := flag.Bool("strict", false, "enforce staleness on TESTED/BENCHED (30d)")
	skipExpr := flag.String("skip", skipDefault.String(), "skip path regex (RE2)")
	updateStale := flag.Bool("update-stale", false, "rewrite UPDATED field for stale TESTED/BENCHED tokens")
	projectOnly := flag.Bool("project-only", false, "filter by project requirement ID pattern from .canary/project.yaml")
	flag.Parse()
	skip, err := regexp.Compile(*skipExpr)
	if err != nil {
		failParse(fmt.Errorf("bad --skip regex: %w", err))
	}

	// Load project config if --project-only is set
	var projectFilter *regexp.Regexp
	if *projectOnly {
		cfg, err := loadProjectConfig(*root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: --project-only specified but failed to load .canary/project.yaml: %v\n", err)
			fmt.Fprintf(os.Stderr, "Scanning all requirements. Run 'canary init' to create project config.\n")
		} else if cfg.Requirements.IDPattern != "" {
			projectFilter, err = regexp.Compile(cfg.Requirements.IDPattern)
			if err != nil {
				failParse(fmt.Errorf("invalid project id_pattern %q: %w", cfg.Requirements.IDPattern, err))
			}
			fmt.Fprintf(os.Stderr, "Filtering by project pattern: %s\n", cfg.Requirements.IDPattern)
		}
	}

	// Load .canaryignore if it exists
	ignorePatterns, err := loadCanaryIgnore(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load .canaryignore: %v\n", err)
	}
	if ignorePatterns != nil {
		fmt.Fprintf(os.Stderr, "Loaded .canaryignore patterns\n")
	}

	rep, err := scan(*root, skip, projectFilter, ignorePatterns)
	if err != nil {
		failParse(err)
	}

	// Handle --update-stale before writing output files
	if *updateStale {
		staleTokens := stale(rep, 30*24*time.Hour)
		if len(staleTokens) > 0 {
			updatedFiles, err := updateStaleTokens(*root, skip, staleTokens)
			if err != nil {
				fmt.Fprintf(os.Stderr, "CANARY_UPDATE_ERROR: %v\n", err)
				os.Exit(3)
			}
			fmt.Fprintf(os.Stderr, "Updated %d stale tokens in %d files\n", len(staleTokens), len(updatedFiles))
			// Re-scan after updates
			rep, err = scan(*root, skip, projectFilter, ignorePatterns)
			if err != nil {
				failParse(err)
			}
		} else {
			fmt.Fprintln(os.Stderr, "No stale tokens found")
		}
	}

	if err := writeJSON(*outJSON, rep); err != nil {
		failParse(err)
	}
	if *outCSV != "" {
		if err := writeCSV(*outCSV, rep); err != nil {
			failParse(err)
		}
	}
	var diags []string
	if *verifyPath != "" {
		diags = append(diags, verifyClaims(rep, *verifyPath)...)
	}
	if *strict && !*updateStale {
		diags = append(diags, stale(rep, 30*24*time.Hour)...)
	}
	if len(diags) > 0 {
		for _, d := range diags {
			fmt.Fprintln(os.Stderr, d)
		}
		os.Exit(2)
	}
}

type aggregateKey struct{ req, feature, aspect, owner, updated string }
type aggregateVal struct {
	status                string
	files, tests, benches map[string]struct{}
}

// loadProjectConfig loads .canary/project.yaml from the given root
func loadProjectConfig(root string) (*ProjectConfig, error) {
	configPath := filepath.Join(root, ".canary", "project.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse project.yaml: %w", err)
	}
	return &cfg, nil
}

// loadCanaryIgnore loads .canaryignore patterns from the given root
// Returns nil if .canaryignore doesn't exist (not an error)
func loadCanaryIgnore(root string) (*ignore.GitIgnore, error) {
	ignorePath := filepath.Join(root, ".canaryignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return nil, nil // No .canaryignore file, not an error
	}

	gi, err := ignore.CompileIgnoreFile(ignorePath)
	if err != nil {
		return nil, fmt.Errorf("parse .canaryignore: %w", err)
	}
	return gi, nil
}

func scan(root string, skip *regexp.Regexp, projectFilter *regexp.Regexp, ignorePatterns *ignore.GitIgnore) (Report, error) {
	if root == "" {
		root = "."
	}
	if skip == nil {
		skip = skipDefault
	}
	agg := map[aggregateKey]*aggregateVal{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for .canaryignore matching
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			relPath = path
		}

		// Check .canaryignore patterns
		if ignorePatterns != nil && ignorePatterns.MatchesPath(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if skip.MatchString(path) {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip acceptance fixture testdata only when scanning the canary tool itself (root path equals the tools/canary dir)
		if strings.Contains(path, string(filepath.Separator)+"testdata"+string(filepath.Separator)) {
			// Only skip when the scan root itself is the canary tool directory (self-scan)
			base := filepath.Clean(root)
			if strings.HasSuffix(base, string(filepath.Join("tools", "canary"))) {
				return nil
			}
		}
		if skip.MatchString(path) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		matches := tokenLineRe.FindAllStringSubmatch(string(b), -1)
		for _, m := range matches {
			fields, perr := parseKV(m[1])
			if perr != nil {
				return fmt.Errorf("%s: %w", path, perr)
			}
			for _, k := range []string{"REQ", "FEATURE", "ASPECT", "STATUS", "UPDATED"} {
				if fields[k] == "" {
					return fmt.Errorf("%s: missing %s", path, k)
				}
			}
			req := normalizeREQ(fields["REQ"])

			// Apply project filter if specified
			if projectFilter != nil && !projectFilter.MatchString(req) {
				continue // Skip requirements that don't match project pattern
			}

			aspect := fields["ASPECT"]
			if _, ok := aspects[aspect]; !ok {
				return fmt.Errorf("%s: invalid ASPECT %s", path, aspect)
			}
			if _, ok := statusSet[fields["STATUS"]]; !ok {
				return fmt.Errorf("%s: invalid STATUS %s", path, fields["STATUS"])
			}
			k := aggregateKey{req: req, feature: unquote(fields["FEATURE"]), aspect: aspect, owner: fields["OWNER"], updated: fields["UPDATED"]}
			a := agg[k]
			if a == nil {
				a = &aggregateVal{status: fields["STATUS"], files: map[string]struct{}{}, tests: map[string]struct{}{}, benches: map[string]struct{}{}}
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
		return nil
	})
	if err != nil {
		return Report{}, err
	}
	byReq := map[string][]Feature{}
	byStatus := StatusCounts{"MISSING": 0, "STUB": 0, "IMPL": 0, "TESTED": 0, "BENCHED": 0, "REMOVED": 0}
	byAspect := AspectCounts{}
	total := 0
	for k, v := range agg {
		status := promote(v.status, len(v.tests) > 0, len(v.benches) > 0)
		f := Feature{Feature: k.feature, Aspect: k.aspect, Status: status, Files: keys(v.files), Tests: keys(v.tests), Benches: keys(v.benches), Owner: k.owner, Updated: k.updated}
		byReq[k.req] = append(byReq[k.req], f)
		byStatus[status]++
		byAspect[k.aspect]++
		total++
	}
	var reqs []Requirement
	for id, feats := range byReq {
		sort.Slice(feats, func(i, j int) bool { return feats[i].Feature+feats[i].Aspect < feats[j].Feature+feats[j].Aspect })
		reqs = append(reqs, Requirement{ID: id, Features: feats})
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })
	rep := Report{GeneratedAt: getTimestamp(), Requirements: reqs, Summary: Summary{ByStatus: byStatus, ByAspect: byAspect, TotalTokens: total, UniqueRequirements: len(reqs)}}
	return rep, nil
}

// getTimestamp returns current UTC timestamp in RFC3339 format, or a fixed timestamp if CANARY_TEST_TIMESTAMP is set
func getTimestamp() string {
	if testTS := os.Getenv("CANARY_TEST_TIMESTAMP"); testTS != "" {
		return testTS
	}
	return time.Now().UTC().Format(time.RFC3339)
}
func promote(status string, hasTests, hasBenches bool) string {
	if status == "IMPL" && hasTests {
		status = "TESTED"
	}
	if (status == "IMPL" || status == "TESTED") && hasBenches {
		status = "BENCHED"
	}
	return status
}
func verifyClaims(rep Report, gapPath string) []string {
	b, err := os.ReadFile(gapPath)
	if err != nil {
		return []string{fmt.Sprintf("CANARY_PARSE_ERROR file=%s err=%q", gapPath, err)}
	}
	matches := claimRe.FindAllStringSubmatch(string(b), -1)
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
	var diags []string
	for id := range claimed {
		if !evidence[id] {
			diags = append(diags, fmt.Sprintf("CANARY_VERIFY_FAIL REQ=%s reason=claimed_but_not_TESTED_OR_BENCHED", id))
		}
	}
	return diags
}
func stale(rep Report, maxAge time.Duration) []string {
	cut := time.Now().UTC().Add(-maxAge)
	var diags []string
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				t, err := time.Parse("2006-01-02", f.Updated)
				if err != nil {
					diags = append(diags, fmt.Sprintf("CANARY_PARSE_ERROR file=%s err=%q", strings.Join(f.Files, ","), err))
					continue
				}
				if t.Before(cut) {
					age := int(time.Since(t).Hours() / 24)
					diags = append(diags, fmt.Sprintf("CANARY_STALE REQ=%s updated=%s age_days=%d threshold=%d", r.ID, f.Updated, age, int(maxAge.Hours()/24)))
				}
			}
		}
	}
	return diags
}

// updateStaleTokens rewrites UPDATED field for stale tokens in source files.
// Returns map of file paths that were updated.
func updateStaleTokens(root string, skip *regexp.Regexp, staleDiags []string) (map[string]bool, error) {
	// Parse stale diagnostics to get REQ IDs that need updating
	staleReqs := make(map[string]bool)
	reqRe := regexp.MustCompile(`REQ=([A-Z]+-\d{3})`)
	for _, diag := range staleDiags {
		matches := reqRe.FindStringSubmatch(diag)
		if len(matches) > 1 {
			staleReqs[matches[1]] = true
		}
	}

	if len(staleReqs) == 0 {
		return nil, nil
	}

	updatedFiles := make(map[string]bool)
	today := time.Now().UTC().Format("2006-01-02")

	// Walk directory and update files
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			rel, _ := filepath.Rel(root, path)
			if rel != "." && skip.MatchString(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if skip.MatchString(rel) {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		// Check if file contains CANARY tokens
		if !tokenLineRe.Match(content) {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		modified := false

		for i, line := range lines {
			if !tokenLineRe.MatchString(line) {
				continue
			}

			// Parse the token to get REQ ID
			match := tokenLineRe.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}

			attrs, err := parseKV(match[1])
			if err != nil {
				continue // Skip malformed tokens
			}

			reqID, hasReq := attrs["REQ"]
			if !hasReq || !staleReqs[reqID] {
				continue
			}

			// Check if token is TESTED or BENCHED
			status, hasStatus := attrs["STATUS"]
			if !hasStatus || (status != "TESTED" && status != "BENCHED") {
				continue
			}

			// Update UPDATED field
			updatedRe := regexp.MustCompile(`(UPDATED=)([0-9]{4}-[0-9]{2}-[0-9]{2})`)
			if updatedRe.MatchString(line) {
				lines[i] = updatedRe.ReplaceAllString(line, fmt.Sprintf("${1}%s", today))
				modified = true
			}
		}

		if modified {
			// Write back to file
			newContent := strings.Join(lines, "\n")
			if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			updatedFiles[path] = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedFiles, nil
}

func writeJSON(path string, rep Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	return enc.Encode(rep)
}
func writeCSV(path string, rep Report) error {
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
		out[strings.ToUpper(m[1])] = strings.TrimSpace(m[2])
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
func max3(a, b, c int) int {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	return a
}
func normalizeREQ(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "‑", "-")
	v = strings.ReplaceAll(v, "–", "-")
	if m := regexp.MustCompile(`^(CBIN-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		n := m[2]
		for len(n) < 3 {
			n = "0" + n
		}
		return m[1] + n
	}
	return v
}
func marshalSortedMap(m map[string]int) ([]byte, error) {
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
func failParse(err error) { fmt.Fprintf(os.Stderr, "CANARY_PARSE_ERROR err=%q\n", err); os.Exit(3) }
