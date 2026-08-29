// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package onboard analyzes a codebase that has few or no CANARY tokens and
// produces the agent hand-off needed to begin adoption: a language
// histogram, top-level directory layout, best-effort entry-point detection,
// any existing CANARY tokens, pre-seeded CANARY:MIGRATE guidance, the
// configured requirement-ID sources, the next available flatfile ID, and a
// next_steps checklist. Read-only — onboard never writes to the target repo.
// CANARY: REQ=CP-276; FEATURE="Onboard"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_303_Analyze,TestCANARY_CBIN_303_Analyze_LimitBounds,TestCANARY_CBIN_303_Analyze_DefaultRootAndLimit,TestCANARY_CBIN_303_CreateOnboardCommand_JSON,TestCANARY_CBIN_303_CreateOnboardCommand_Human; UPDATED=2026-08-29
package onboard

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/reqid"
	"devnw.dev/canary/pkg/sources"
)

// DefaultLimit bounds every list section by default; agents raise it with
// --limit when they need more.
const DefaultLimit = 20

// maxReadSize bounds how large a file onboard will read for entry-point
// detection and token counting (mirrors canaryscan's own file-size guard).
const maxReadSize = 1 << 20 // 1MB

// LanguageCount is one file extension's share of the histogram.
type LanguageCount struct {
	Ext   string `json:"ext"`
	Count int    `json:"count"`
}

// DirCount is one top-level directory's recursive file count.
type DirCount struct {
	Path  string `json:"path"`
	Files int    `json:"files"`
}

// MigrateNote is one pre-seeded CANARY:MIGRATE guidance line.
type MigrateNote struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

// SourceInfo is one configured requirement-ID source.
type SourceInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Key  string `json:"key"`
}

// OnboardReport is the full adoption-analysis answer for one codebase.
type OnboardReport struct {
	Root              string          `json:"root"`
	Languages         []LanguageCount `json:"languages"`
	LanguagesTotal    int             `json:"languages_total"`
	Directories       []DirCount      `json:"directories"`
	DirectoriesTotal  int             `json:"directories_total"`
	EntryPoints       []string        `json:"entry_points"`
	EntryPointsTotal  int             `json:"entry_points_total"`
	TokenCount        int             `json:"token_count"`
	MigrateNotes      []MigrateNote   `json:"migrate_notes"`
	MigrateNotesTotal int             `json:"migrate_notes_total"`
	ProjectKey        string          `json:"project_key"`
	Sources           []SourceInfo    `json:"sources"`
	NextID            string          `json:"next_id"`
	NextSteps         []string        `json:"next_steps"`
}

// onboardTokenLineRe mirrors canaryscan's own (unexported) token-line
// pattern closely enough to spot a CANARY requirement token, without ever
// running the line through KV validation. canaryscan.Scan hard-errors on
// malformed or foreign CANARY content (missing REQ/FEATURE/ASPECT/STATUS,
// an unknown ASPECT or STATUS value) — exactly the kind of noise expected in
// a codebase that hasn't adopted CANARY yet, so a full Scan is not safe to
// call here. Instead onboard counts tolerantly: any non-MIGRATE "CANARY:"
// line that mentions REQ counts as one token. This is a heuristic upper
// bound, not an exact parse — `canary scan` remains the authoritative count
// once the repo's tokens are well-formed.
var onboardTokenLineRe = regexp.MustCompile(`(?m)^[ \t]*(?://|#|/\*)?[ \t]*CANARY:\s*(.*)$`)

func countTokens(content string) int {
	n := 0
	for _, m := range onboardTokenLineRe.FindAllStringSubmatch(content, -1) {
		capture := strings.TrimSpace(m[1])
		if strings.HasPrefix(capture, "MIGRATE") {
			continue
		}
		if strings.Contains(capture, "REQ") {
			n++
		}
	}
	return n
}

// hasMainOrBin reports whether a package.json's content declares a non-empty
// "main" or "bin" field. Best-effort: an unparsable package.json falls back
// to a plain substring check rather than being silently dropped.
func hasMainOrBin(content string) bool {
	var v map[string]any
	if err := json.Unmarshal([]byte(content), &v); err == nil {
		if s, ok := v["main"].(string); ok && s != "" {
			return true
		}
		switch b := v["bin"].(type) {
		case string:
			if b != "" {
				return true
			}
		case map[string]any:
			if len(b) > 0 {
				return true
			}
		}
		return false
	}
	return strings.Contains(content, `"main"`) || strings.Contains(content, `"bin"`)
}

// topLevelDir returns the first path segment of a slash-separated relative
// path and true, or ("", false) when the file sits directly at root.
func topLevelDir(relSlash string) (string, bool) {
	i := strings.IndexByte(relSlash, '/')
	if i < 0 {
		return "", false
	}
	return relSlash[:i], true
}

// entryDir returns the directory portion of a slash-separated relative path,
// using "." for files directly at root.
func entryDir(relSlash string) string {
	d := stdpath.Dir(relSlash)
	if d == "." || d == "" {
		return "."
	}
	return d
}

// Analyze walks root read-only and builds the adoption-analysis report.
// limit bounds every list section; <=0 uses DefaultLimit.
func Analyze(root string, limit int) (*OnboardReport, error) {
	if root == "" {
		root = "."
	}
	if limit <= 0 {
		limit = DefaultLimit
	}

	skip := canaryscan.DefaultSkipRegex()
	ignorePatterns, _ := canaryscan.LoadCanaryIgnore(root) // a malformed .canaryignore just means "no filter"

	langCounts := map[string]int{}
	dirCounts := map[string]int{}
	entrySet := map[string]struct{}{}
	tokenCount := 0

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			relPath = ""
		}
		relSlash := filepath.ToSlash(relPath)

		if ignorePatterns != nil && relSlash != "" && ignorePatterns.MatchesPath(relSlash) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if path != root && skip.MatchString(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if skip.MatchString(path) {
			return nil
		}

		if top, ok := topLevelDir(relSlash); ok {
			dirCounts[top]++
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			langCounts["(no ext)"]++
		} else {
			langCounts[strings.TrimPrefix(ext, ".")]++
		}

		info, ierr := d.Info()
		if ierr != nil || info.Size() > maxReadSize {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil // unreadable file is not fatal to onboard analysis
		}
		content := string(b)

		base := filepath.Base(path)
		switch {
		case strings.HasSuffix(base, ".go") && strings.Contains(content, "func main("):
			entrySet[entryDir(relSlash)] = struct{}{}
		case base == "package.json" && hasMainOrBin(content):
			entrySet[entryDir(relSlash)] = struct{}{}
		case base == "pyproject.toml":
			entrySet[entryDir(relSlash)] = struct{}{}
		}

		tokenCount += countTokens(content)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	reg := sources.LoadFromRoot(root)
	notes, _ := canaryscan.ScanMigrateNotes(root, skip, ignorePatterns, reg) // unreadable files are skipped, not fatal

	projectKey := "CBIN"
	srcInfos := make([]SourceInfo, 0, len(reg.Sources()))
	for _, s := range reg.Sources() {
		srcInfos = append(srcInfos, SourceInfo{Name: s.Name, Type: s.Type, Key: s.Key})
	}
	if len(srcInfos) > 0 {
		projectKey = srcInfos[0].Key
	}

	rep := &OnboardReport{
		Root:       root,
		TokenCount: tokenCount,
		ProjectKey: projectKey,
		Sources:    srcInfos,
		NextID:     nextFlatfileID(root, projectKey),
	}
	rep.Languages, rep.LanguagesTotal = topLanguages(langCounts, limit)
	rep.Directories, rep.DirectoriesTotal = topDirectories(dirCounts, limit)
	rep.EntryPoints, rep.EntryPointsTotal = topEntryPoints(entrySet, limit)
	rep.MigrateNotes, rep.MigrateNotesTotal = boundMigrateNotes(notes, limit)
	rep.NextSteps = buildNextSteps(root, rep)

	return rep, nil
}

func topLanguages(counts map[string]int, limit int) ([]LanguageCount, int) {
	all := make([]LanguageCount, 0, len(counts))
	for ext, n := range counts {
		all = append(all, LanguageCount{Ext: ext, Count: n})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Count != all[j].Count {
			return all[i].Count > all[j].Count
		}
		return all[i].Ext < all[j].Ext
	})
	total := len(all)
	if len(all) > limit {
		all = all[:limit]
	}
	return all, total
}

func topDirectories(counts map[string]int, limit int) ([]DirCount, int) {
	all := make([]DirCount, 0, len(counts))
	for p, n := range counts {
		all = append(all, DirCount{Path: p, Files: n})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Files != all[j].Files {
			return all[i].Files > all[j].Files
		}
		return all[i].Path < all[j].Path
	})
	total := len(all)
	if len(all) > limit {
		all = all[:limit]
	}
	return all, total
}

func topEntryPoints(set map[string]struct{}, limit int) ([]string, int) {
	all := make([]string, 0, len(set))
	for p := range set {
		all = append(all, p)
	}
	sort.Strings(all)
	total := len(all)
	if len(all) > limit {
		all = all[:limit]
	}
	return all, total
}

func boundMigrateNotes(notes []canaryscan.MigrateNote, limit int) ([]MigrateNote, int) {
	out := make([]MigrateNote, 0, len(notes))
	for _, n := range notes {
		out = append(out, MigrateNote{File: n.File, Line: n.Line, Text: n.Text})
	}
	total := len(out)
	if len(out) > limit {
		out = out[:limit]
	}
	return out, total
}

// nextFlatfileID resolves the next available flatfile requirement ID for
// key — what `canary specify` would hand out next. "CLI" is used as a
// representative aspect purely to probe whether .canary/specs exists and
// has entries — onboard has no way to know which aspect the first real
// requirement will use. Any failure (no .canary/specs yet, unreadable)
// falls back to the flat "<KEY>-001" form, which is the common case for a
// codebase that hasn't run `canary init` yet.
func nextFlatfileID(root, key string) string {
	fallback := key + "-001"

	id, err := reqid.GenerateNextIDIn(root, key, "CLI")
	if err != nil {
		return fallback
	}
	return id
}

func buildNextSteps(root string, rep *OnboardReport) []string {
	var steps []string
	if _, err := os.Stat(filepath.Join(root, ".canary")); os.IsNotExist(err) {
		steps = append(steps, "Run `canary init` — no .canary/ directory found in this codebase yet.")
	}
	if rep.MigrateNotesTotal > 0 {
		steps = append(steps, fmt.Sprintf(
			"Read the %d CANARY:MIGRATE note(s) above first — pre-seeded author guidance for adoption.",
			rep.MigrateNotesTotal))
	}
	steps = append(steps,
		"Propose one CANARY requirement per subsystem (see Directories/Entry points above for candidates).",
		fmt.Sprintf("Run `canary create` per requirement, starting at %s.", rep.NextID),
		"Run `canary index && canary scan` to build the index and verify tokens.",
	)
	return steps
}

// CreateOnboardCommand returns the `canary onboard` command.
func CreateOnboardCommand() *cobra.Command {
	var jsonOut bool
	var limit int
	var root string
	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Fresh-codebase adoption analysis: languages, entry points, existing tokens, MIGRATE notes, next steps",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := Analyze(root, limit)
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				return enc.Encode(rep) // compact: no SetIndent
			}
			printReport(cmd, rep)
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "root directory to analyze")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "compact JSON output")
	cmd.Flags().IntVar(&limit, "limit", DefaultLimit, "max entries per list section (raise when you need more)")
	return cmd
}

func printReport(cmd *cobra.Command, r *OnboardReport) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Onboard analysis: %s\n", r.Root)

	langs := make([]string, 0, len(r.Languages))
	for _, l := range r.Languages {
		langs = append(langs, fmt.Sprintf("%s=%d", l.Ext, l.Count))
	}
	fmt.Fprintf(out, "Languages:    %s", strings.Join(langs, ", "))
	if r.LanguagesTotal > len(r.Languages) {
		fmt.Fprintf(out, " … +%d more (use --limit %d)", r.LanguagesTotal-len(r.Languages), r.LanguagesTotal)
	}
	fmt.Fprintln(out)

	dirs := make([]string, 0, len(r.Directories))
	for _, d := range r.Directories {
		dirs = append(dirs, fmt.Sprintf("%s(%d)", d.Path, d.Files))
	}
	fmt.Fprintf(out, "Directories:  %s", strings.Join(dirs, ", "))
	if r.DirectoriesTotal > len(r.Directories) {
		fmt.Fprintf(out, " … +%d more (use --limit %d)", r.DirectoriesTotal-len(r.Directories), r.DirectoriesTotal)
	}
	fmt.Fprintln(out)

	fmt.Fprintf(out, "Entry points: %s", strings.Join(r.EntryPoints, ", "))
	if r.EntryPointsTotal > len(r.EntryPoints) {
		fmt.Fprintf(out, " … +%d more (use --limit %d)", r.EntryPointsTotal-len(r.EntryPoints), r.EntryPointsTotal)
	}
	fmt.Fprintln(out)

	fmt.Fprintf(out, "CANARY tokens: %d (heuristic count; run `canary scan` for exact)\n", r.TokenCount)

	srcs := make([]string, 0, len(r.Sources))
	for _, s := range r.Sources {
		srcs = append(srcs, fmt.Sprintf("%s(%s:%s)", s.Name, s.Type, s.Key))
	}
	fmt.Fprintf(out, "Sources:      %s\n", strings.Join(srcs, ", "))
	fmt.Fprintf(out, "Project key:  %s   Next ID: %s\n", r.ProjectKey, r.NextID)

	if len(r.MigrateNotes) > 0 {
		fmt.Fprintln(out, "MIGRATE notes:")
		for _, n := range r.MigrateNotes {
			fmt.Fprintf(out, "  %s:%d: %s\n", n.File, n.Line, n.Text)
		}
		if r.MigrateNotesTotal > len(r.MigrateNotes) {
			fmt.Fprintf(out, "  … +%d more (use --limit %d)\n", r.MigrateNotesTotal-len(r.MigrateNotes), r.MigrateNotesTotal)
		}
	} else {
		fmt.Fprintln(out, "MIGRATE notes: none")
	}

	fmt.Fprintln(out, "Next steps:")
	for _, s := range r.NextSteps {
		fmt.Fprintf(out, "  - %s\n", s)
	}
}
