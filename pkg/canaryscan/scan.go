package canaryscan

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"devnw.dev/canary/pkg/sources"
)

const (
	// binarySniffBytes is how much of a file's head is inspected for a NUL
	// byte before deciding it is binary.
	binarySniffBytes = 8 << 10
	// scanBufBytes is the read buffer used for the streaming line loop.
	scanBufBytes = 64 << 10
)

type aggregateKey struct{ req, feature, aspect, owner, updated string }
type aggregateVal struct {
	status                string
	files, tests, benches map[string]struct{}
	deps                  map[string]struct{}
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

// Scan walks root and returns a Report. skip, projectFilter, ignorePatterns
// and reg may be nil. Files the scanner cannot fully process are recorded in
// Report.Issues and skipped; only walk errors abort.
func Scan(root string, skip *regexp.Regexp, projectFilter *regexp.Regexp, ignorePatterns *ignore.GitIgnore, reg *sources.Registry) (Report, error) {
	if root == "" {
		root = "."
	}
	if skip == nil {
		skip = DefaultSkipRegex()
	}
	agg := map[aggregateKey]*aggregateVal{}
	var issues []ScanIssue
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
		captures, issue := readTokenCaptures(path, d)
		if issue != nil {
			issues = append(issues, *issue)
		}
		for _, cap := range captures {
			if isMigrateCapture(cap) {
				continue
			}
			fields, perr := parseKV(cap, reg)
			if perr != nil {
				issues = append(issues, ScanIssue{Path: path, Reason: IssueParseError, Detail: perr.Error()})
				continue
			}
			if len(fields) == 0 {
				continue
			}
			missing := ""
			for _, k := range []string{"REQ", "FEATURE", "ASPECT", "STATUS", "UPDATED"} {
				if fields[k] == "" {
					missing = k
					break
				}
			}
			if missing != "" {
				issues = append(issues, ScanIssue{Path: path, Reason: IssueParseError,
					Detail: fmt.Sprintf("missing %s in token: CANARY: %s", missing, cap)})
				continue
			}
			req := normalizeREQWithRegistry(fields["REQ"], reg)
			if projectFilter != nil && !projectFilter.MatchString(req) {
				continue
			}
			aspect := fields["ASPECT"]
			if _, ok := aspects[aspect]; !ok {
				issues = append(issues, ScanIssue{Path: path, Reason: IssueParseError,
					Detail: fmt.Sprintf("invalid ASPECT %s", aspect)})
				continue
			}
			if _, ok := statusSet[fields["STATUS"]]; !ok {
				issues = append(issues, ScanIssue{Path: path, Reason: IssueParseError,
					Detail: fmt.Sprintf("invalid STATUS %s", fields["STATUS"])})
				continue
			}
			k := aggregateKey{req: req, feature: unquote(fields["FEATURE"]), aspect: aspect, owner: fields["OWNER"], updated: fields["UPDATED"]}
			a := agg[k]
			if a == nil {
				a = &aggregateVal{status: fields["STATUS"], files: map[string]struct{}{}, tests: map[string]struct{}{}, benches: map[string]struct{}{}, deps: map[string]struct{}{}}
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
			for _, dep := range splitList(fields["DEPENDS_ON"]) {
				if dep != "" {
					a.deps[normalizeREQWithRegistry(dep, reg)] = struct{}{}
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
		// STATUS is a declaration: it passes through verbatim. TEST=/BENCH=
		// are recorded as evidence references but never change it.
		status := v.status
		f := Feature{Feature: k.feature, Aspect: k.aspect, Status: status, Files: mapKeys(v.files), Tests: mapKeys(v.tests), Benches: mapKeys(v.benches), Owner: k.owner, Updated: k.updated, DependsOn: mapKeys(v.deps)}
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
	effReg := reg
	if effReg == nil {
		effReg = sources.Default()
	}
	diagRefs, diagIssues, derr := ScanDiagramRefs(root, skip, effReg, ignorePatterns)
	issues = append(issues, diagIssues...)
	if derr != nil {
		fmt.Fprintf(os.Stderr, "Warning: diagram ref scan failed: %v\n", derr)
	}
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

	// Extract CANARY:MIGRATE guidance notes (Task CBIN-301).
	notes, noteIssues, nerr := ScanMigrateNotes(root, skip, ignorePatterns, effReg)
	issues = append(issues, noteIssues...)
	if nerr != nil {
		fmt.Fprintf(os.Stderr, "Warning: migrate notes scan failed: %v\n", nerr)
	}
	sort.Slice(notes, func(i, j int) bool {
		if notes[i].File != notes[j].File {
			return notes[i].File < notes[j].File
		}
		return notes[i].Line < notes[j].Line
	})

	return Report{
		GeneratedAt:    getTimestamp(),
		Requirements:   reqs,
		Summary:        Summary{ByStatus: byStatus, ByAspect: byAspect, TotalTokens: total, UniqueRequirements: len(reqs)},
		MigrationNotes: notes,
		Issues:         normalizeIssues(issues),
	}, nil
}

// normalizeIssues sorts issues by path then reason and drops duplicates, so
// the same unreadable file reported by two sub-scans appears once and the
// JSON report is stable across runs.
func normalizeIssues(issues []ScanIssue) []ScanIssue {
	if len(issues) == 0 {
		return nil
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		if issues[i].Reason != issues[j].Reason {
			return issues[i].Reason < issues[j].Reason
		}
		return issues[i].Detail < issues[j].Detail
	})
	out := issues[:0]
	seen := map[[2]string]struct{}{}
	for _, is := range issues {
		key := [2]string{is.Path, is.Reason}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, is)
	}
	return out
}

// readTokenCaptures streams path line by line and returns every CANARY token
// body it finds. A non-nil issue means the file was skipped (wholly or from
// the offending line on); it is never fatal to the surrounding walk.
func readTokenCaptures(path string, d os.DirEntry) ([]string, *ScanIssue) {
	info, ierr := d.Info()
	if ierr != nil {
		return nil, &ScanIssue{Path: path, Reason: IssueReadError, Detail: ierr.Error()}
	}
	if info.Size() > MaxFileBytes {
		return nil, &ScanIssue{Path: path, Reason: IssueFileTooLarge,
			Detail: fmt.Sprintf("%d bytes exceeds limit %d", info.Size(), MaxFileBytes)}
	}
	f, oerr := os.Open(path)
	if oerr != nil {
		return nil, &ScanIssue{Path: path, Reason: IssueReadError, Detail: oerr.Error()}
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReaderSize(f, scanBufBytes)
	head, perr := r.Peek(binarySniffBytes)
	if perr != nil && perr != io.EOF && perr != bufio.ErrBufferFull {
		return nil, &ScanIssue{Path: path, Reason: IssueReadError, Detail: perr.Error()}
	}
	if bytes.IndexByte(head, 0x00) >= 0 {
		return nil, &ScanIssue{Path: path, Reason: IssueBinary, Detail: "NUL byte in first 8 KiB"}
	}

	var captures []string
	var line []byte
	for {
		chunk, rerr := r.ReadSlice('\n')
		if len(line)+len(chunk) > MaxLineBytes {
			return captures, &ScanIssue{Path: path, Reason: IssueLineTooLarge,
				Detail: fmt.Sprintf("logical line exceeds limit %d", MaxLineBytes)}
		}
		line = append(line, chunk...)
		if rerr == bufio.ErrBufferFull {
			continue
		}
		if rerr != nil && rerr != io.EOF {
			return captures, &ScanIssue{Path: path, Reason: IssueReadError, Detail: rerr.Error()}
		}
		if len(line) > 0 {
			if m := tokenLineRe.FindSubmatch(bytes.TrimRight(line, "\r\n")); m != nil {
				captures = append(captures, string(m[1]))
			}
		}
		if rerr == io.EOF {
			return captures, nil
		}
		line = line[:0]
	}
}
