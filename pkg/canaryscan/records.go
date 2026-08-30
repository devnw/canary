package canaryscan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"

	"devnw.dev/canary/pkg/sources"
)

// CANARY: REQ=CBIN-305; FEATURE="TokenRecordScan"; ASPECT=Engine; STATUS=TESTED; TEST=TestScanTokenRecordsLineNumbers,TestScanTokenRecordsReportsParseIssues,TestAuditF12; UPDATED=2026-08-30

// TokenRecord is one CANARY token exactly where it was found. Scan folds
// tokens into per-requirement aggregates, which is the right shape for a
// report and the wrong shape for an index: the index needs the line each
// token sits on, its verbatim text, and every field it declared. Both views
// come out of the same parser, so the index can never disagree with the scan
// about what a token says.
type TokenRecord struct {
	// File is the path the token was read from, relative to the scan root.
	File string
	// Line is the 1-based line the token sits on.
	Line int
	// Raw is the token body verbatim, without the "CANARY:" marker.
	Raw string
	// ReqID is the registry-normalized requirement ID.
	ReqID string
	// Fields holds every declared key, upper-cased, values decoded.
	Fields map[string]string
	// ContentHash is the hex SHA-256 of the whole file this token came from.
	ContentHash string
}

// Field returns the named field, or "" when the token did not declare it.
// It never invents a value: an absent UPDATED reads as "" so a caller cannot
// mistake today's date for something the author wrote.
func (r TokenRecord) Field(key string) string {
	if r.Fields == nil {
		return ""
	}
	return r.Fields[key]
}

// FileDigest pairs a scanned file with the digest of its contents.
type FileDigest struct {
	Path string
	Hash string
}

// ScanTokenRecords walks root and returns every CANARY token it finds, one
// record per token, alongside the digest of every file that contributed at
// least one token and the issues the scan hit. skip, projectFilter,
// ignorePatterns and reg may be nil.
//
// Issues are returned rather than swallowed: a caller that is building an
// index (as opposed to printing a report) can refuse to commit anything when
// the tree contains a token the parser rejects.
func ScanTokenRecords(root string, skip *regexp.Regexp, projectFilter *regexp.Regexp, ignorePatterns *ignore.GitIgnore, reg *sources.Registry) ([]TokenRecord, []FileDigest, []ScanIssue, error) {
	if root == "" {
		root = "."
	}
	if skip == nil {
		skip = DefaultSkipRegex()
	}

	var records []TokenRecord
	var issues []ScanIssue
	digests := map[string]string{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, relErr := filepath.Rel(root, path)
		if relErr != nil || relPath == "" {
			relPath = path
		}
		relPath = filepath.ToSlash(relPath)
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

		captures, issue := readTokenCapturesLocated(path, d)
		if issue != nil {
			issues = append(issues, *issue)
		}
		if len(captures) == 0 {
			return nil
		}

		hash, herr := hashFile(path)
		if herr != nil {
			issues = append(issues, ScanIssue{Path: path, Reason: IssueReadError, Detail: herr.Error()})
			return nil
		}

		contributed := false
		for _, cap := range captures {
			if isMigrateCapture(cap.Body) {
				continue
			}
			fields, perr := parseKV(cap.Body, reg)
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
					Detail: fmt.Sprintf("missing %s in token: CANARY: %s", missing, cap.Body)})
				continue
			}
			req := normalizeREQWithRegistry(fields["REQ"], reg)
			if projectFilter != nil && !projectFilter.MatchString(req) {
				continue
			}
			if _, ok := aspects[fields["ASPECT"]]; !ok {
				issues = append(issues, ScanIssue{Path: path, Reason: IssueParseError,
					Detail: fmt.Sprintf("invalid ASPECT %s", fields["ASPECT"])})
				continue
			}
			if _, ok := statusSet[fields["STATUS"]]; !ok {
				issues = append(issues, ScanIssue{Path: path, Reason: IssueParseError,
					Detail: fmt.Sprintf("invalid STATUS %s", fields["STATUS"])})
				continue
			}
			fields["FEATURE"] = unquote(fields["FEATURE"])
			fields["REQ"] = req

			records = append(records, TokenRecord{
				File:        path,
				Line:        cap.Line,
				Raw:         strings.TrimSpace(cap.Body),
				ReqID:       req,
				Fields:      fields,
				ContentHash: hash,
			})
			contributed = true
		}
		if contributed {
			digests[relPath] = hash
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].File != records[j].File {
			return records[i].File < records[j].File
		}
		return records[i].Line < records[j].Line
	})

	files := make([]FileDigest, 0, len(digests))
	for p, h := range digests {
		files = append(files, FileDigest{Path: p, Hash: h})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	return records, files, normalizeIssues(issues), nil
}

// ScanDigest folds a file-digest set into one digest over the sorted
// "path:hash" lines, so two scans of the same tree produce the same value and
// any change to any indexed file changes it.
func ScanDigest(files []FileDigest) string {
	sorted := make([]FileDigest, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })

	h := sha256.New()
	for _, f := range sorted {
		fmt.Fprintf(h, "%s:%s\n", f.Path, f.Hash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// hashFile returns the hex SHA-256 of a file's contents.
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from the scan walk
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
