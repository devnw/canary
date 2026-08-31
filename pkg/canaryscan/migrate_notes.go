// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4325; FEATURE="MigrateNotes"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_301_ExtractMigrateNotes; UPDATED=2026-08-29
package canaryscan

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"devnw.dev/canary/pkg/sources"
)

// MigrateNote is one CANARY:MIGRATE guidance comment.
type MigrateNote struct {
	File   string   `json:"file"`
	Line   int      `json:"line"` // 1-based
	Text   string   `json:"text"`
	ReqIDs []string `json:"req_ids,omitempty"` // registry-pattern matches inside Text, normalized
}

// migrateLineRe matches a CANARY:MIGRATE guidance line across common comment
// styles (//, #, --, <!-- -->, /* */, and markdown [//]: #), capturing the
// free-text guidance in group 1.
var migrateLineRe = regexp.MustCompile(`^\s*(?://|#|--|<!--|/\*+|\*|\[//\]:\s*#)?\s*CANARY:MIGRATE\b[:\s]*(.*)$`)

// maxMigrateFileSize bounds how large a file ScanMigrateNotes will read.
const maxMigrateFileSize = 1 << 20 // 1MB

// ExtractMigrateNotes finds CANARY:MIGRATE guidance comments in content.
// relPath is recorded verbatim on each note. reg nil means sources.Default()
// (the legacy CBIN series). A MIGRATE line is free text, not a KV token, and
// is never fed to parseKV — this is what keeps a MIGRATE line from ever
// aborting a scan.
func ExtractMigrateNotes(relPath, content string, reg *sources.Registry) []MigrateNote {
	if reg == nil {
		reg = sources.Default()
	}
	var notes []MigrateNote
	for i, line := range strings.Split(content, "\n") {
		m := migrateLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := strings.TrimSpace(m[1])
		text = strings.TrimSuffix(text, "-->")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		var reqIDs []string
		seen := map[string]struct{}{}
		for _, hit := range reg.Pattern().FindAllString(text, -1) {
			id := reg.Normalize(hit)
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			reqIDs = append(reqIDs, id)
		}

		notes = append(notes, MigrateNote{File: relPath, Line: i + 1, Text: text, ReqIDs: reqIDs})
	}
	return notes
}

// ScanMigrateNotes walks root and extracts CANARY:MIGRATE guidance notes from
// every non-binary file the scanner would otherwise read. Paths in the
// result are root-relative with forward slashes. ignorePatterns (from
// LoadCanaryIgnore) is honored the same way Scan honors it: relative path
// matched first, dirs skipped via SkipDir; nil means no .canaryignore
// patterns apply. Files larger than 1MB are skipped.
//
// A file that cannot be read is reported as a ScanIssue rather than silently
// skipped, so callers can tell an empty result from an incomplete one.
func ScanMigrateNotes(root string, skip *regexp.Regexp, ignorePatterns *ignore.GitIgnore, reg *sources.Registry) ([]MigrateNote, []ScanIssue, error) {
	if skip == nil {
		skip = DefaultSkipRegex()
	}
	var out []MigrateNote
	var issues []ScanIssue
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
		if skip.MatchString(path) {
			return nil
		}
		if info, ierr := d.Info(); ierr == nil && info.Size() > maxMigrateFileSize {
			return nil
		}
		b, rerr := os.ReadFile(path) //nolint:gosec // path comes from the local scan walk
		if rerr != nil {
			// Unreadable file is not fatal to a scan, but it is not
			// invisible either.
			issues = append(issues, ScanIssue{Path: path, Reason: IssueReadError, Detail: rerr.Error()})
			return nil
		}
		rel, relerr := filepath.Rel(root, path)
		if relerr != nil {
			rel = path
		}
		out = append(out, ExtractMigrateNotes(filepath.ToSlash(rel), string(b), reg)...)
		return nil
	})
	return out, issues, err
}
