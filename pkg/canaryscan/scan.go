package canaryscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"gopkg.in/yaml.v3"

	"devnw.dev/canary/pkg/sources"
)

// ProjectConfig is the .canary/project.yaml shape.
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

type aggregateKey struct{ req, feature, aspect, owner, updated string }
type aggregateVal struct {
	status                string
	files, tests, benches map[string]struct{}
}

// LoadProjectConfig loads .canary/project.yaml from root. Returns nil if file missing.
func LoadProjectConfig(root string) (*ProjectConfig, error) {
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

// LoadCanaryIgnore loads .canaryignore from root. Returns nil if file missing.
func LoadCanaryIgnore(root string) (*ignore.GitIgnore, error) {
	ignorePath := filepath.Join(root, ".canaryignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return nil, nil
	}
	gi, err := ignore.CompileIgnoreFile(ignorePath)
	if err != nil {
		return nil, fmt.Errorf("parse .canaryignore: %w", err)
	}
	return gi, nil
}

// Scan walks root and returns a Report. skip, projectFilter, and ignorePatterns may be nil.
func Scan(root string, skip *regexp.Regexp, projectFilter *regexp.Regexp, ignorePatterns *ignore.GitIgnore) (Report, error) {
	if root == "" {
		root = "."
	}
	if skip == nil {
		skip = DefaultSkipRegex()
	}
	agg := map[aggregateKey]*aggregateVal{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(root, path)
		if relPath == "" {
			relPath = "."
		}
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
		if strings.Contains(path, string(filepath.Separator)+"testdata"+string(filepath.Separator)) {
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
			if len(fields) == 0 {
				continue
			}
			for _, k := range []string{"REQ", "FEATURE", "ASPECT", "STATUS", "UPDATED"} {
				if fields[k] == "" {
					absPath, _ := filepath.Abs(path)
					return fmt.Errorf("%s (abs: %s): missing %s in token: %s", path, absPath, k, m[0])
				}
			}
			req := normalizeREQ(fields["REQ"])
			if projectFilter != nil && !projectFilter.MatchString(req) {
				continue
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
	byStatus := map[string]int{"MISSING": 0, "STUB": 0, "IMPL": 0, "TESTED": 0, "BENCHED": 0, "REMOVED": 0}
	byAspect := map[string]int{}
	total := 0
	for k, v := range agg {
		if v.status == "FIXED" {
			v.status = "REMOVED"
		}
		status := promote(v.status, len(v.tests) > 0, len(v.benches) > 0)
		f := Feature{Feature: k.feature, Aspect: k.aspect, Status: status, Files: mapKeys(v.files), Tests: mapKeys(v.tests), Benches: mapKeys(v.benches), Owner: k.owner, Updated: k.updated}
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

	// Attach mermaid diagram references (Task CBIN-202).
	reg := activeRegistry
	if reg == nil {
		reg = sources.Default()
	}
	diagRefs, _ := ScanDiagramRefs(root, skip, reg)
	byID := map[string]map[string]struct{}{}
	for _, r := range diagRefs {
		key := fmt.Sprintf("%s:%d", r.File, r.Line)
		if byID[r.ReqID] == nil {
			byID[r.ReqID] = map[string]struct{}{}
		}
		byID[r.ReqID][key] = struct{}{}
	}
	for i := range reqs {
		if set, ok := byID[reqs[i].ID]; ok {
			reqs[i].Diagrams = mapKeys(set)
		}
	}

	return Report{
		GeneratedAt:  getTimestamp(),
		Requirements: reqs,
		Summary:      Summary{ByStatus: byStatus, ByAspect: byAspect, TotalTokens: total, UniqueRequirements: len(reqs)},
	}, nil
}
