// CANARY: REQ=CBIN-304; FEATURE="StalenessConfig"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_304_StaleDaysFromConfig,TestCANARY_CBIN_304_UpdateStaleV2IDs,TestCANARY_CBIN_304_UpdateStaleAddsMissingUpdated,TestCANARY_CBIN_304_RunReportsActualRewriteCount; UPDATED=2026-08-29
package canaryscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// diagReqRe extracts the REQ ID from a CANARY_STALE diagnostic line
// (format: "CANARY_STALE REQ=<id> updated=<date> age_days=<n> threshold=<n>").
// REQ IDs are opaque, source-defined strings (e.g. "CBIN-304", legacy
// "PLAT-4521", or v2 multi-segment IDs like "CBIN-CLI-001"), so match
// anything up to the next whitespace rather than assuming a shape.
var diagReqRe = regexp.MustCompile(`REQ=(\S+)`)

// updatedDateRe matches an existing UPDATED=YYYY-MM-DD attribute within a token line.
var updatedDateRe = regexp.MustCompile(`(UPDATED=)([0-9]{4}-[0-9]{2}-[0-9]{2})`)

// UpdateStaleTokens rewrites UPDATED in source files for tokens matching staleDiags REQ IDs.
// Only updates TESTED/BENCHED tokens. Existing UPDATED=YYYY-MM-DD values are rewritten to
// today; a token line missing UPDATED= entirely gains one. Returns the set of files that
// were modified and the number of individual token lines actually rewritten (which may
// differ from len(staleDiags): several diags can share one REQ, or a REQ can span multiple
// physical token lines/files).
func UpdateStaleTokens(root string, skip *regexp.Regexp, staleDiags []string) (updatedFiles map[string]bool, tokenCount int, err error) {
	staleReqs := make(map[string]bool)
	for _, diag := range staleDiags {
		matches := diagReqRe.FindStringSubmatch(diag)
		if len(matches) > 1 {
			staleReqs[matches[1]] = true
		}
	}
	if len(staleReqs) == 0 {
		return nil, 0, nil
	}
	updatedFiles = make(map[string]bool)
	refTime := time.Now().UTC()
	if t := RefTimeFromEnv(); !t.IsZero() {
		refTime = t
	}
	today := refTime.Format("2006-01-02")
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if !tokenLineRe.Match(content) {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		modified := false
		fileTokenCount := 0
		for i, line := range lines {
			match := tokenLineRe.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}
			attrs, perr := parseKV(match[1])
			if perr != nil {
				continue
			}
			reqID, hasReq := attrs["REQ"]
			if !hasReq || !staleReqs[reqID] {
				continue
			}
			status, hasStatus := attrs["STATUS"]
			if !hasStatus || (status != "TESTED" && status != "BENCHED") {
				continue
			}
			if updatedDateRe.MatchString(line) {
				lines[i] = updatedDateRe.ReplaceAllString(line, fmt.Sprintf("${1}%s", today))
				modified = true
				fileTokenCount++
				continue
			}
			// Token line is missing UPDATED= entirely: add it rather than
			// silently skipping (the previous behavior could only rewrite
			// an existing date and left dateless tokens untouched forever).
			prefix := line[:len(line)-len(match[1])]
			newContent, changed := addMissingUpdated(match[1], today)
			if changed {
				lines[i] = prefix + newContent
				modified = true
				fileTokenCount++
			}
		}
		if modified {
			newContent := strings.Join(lines, "\n")
			if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			updatedFiles[path] = true
			tokenCount += fileTokenCount
		}
		return nil
	})
	if walkErr != nil {
		return nil, 0, walkErr
	}
	return updatedFiles, tokenCount, nil
}

// addMissingUpdated appends "; UPDATED=<today>" to a CANARY token's content
// (the text captured after "CANARY:" by tokenLineRe) when it lacks an
// UPDATED= attribute. Mirrors pkg/upgrade's addUpdatedContent: CRLF-safe,
// and inserts before a trailing "*/" or "-->" comment closer rather than
// after it. Callers must have already confirmed UPDATED= is absent.
func addMissingUpdated(content, today string) (string, bool) {
	cr := ""
	body := content
	if strings.HasSuffix(body, "\r") {
		cr = "\r"
		body = strings.TrimSuffix(body, "\r")
	}
	if strings.TrimSpace(body) == "" {
		return content, false
	}
	suffix := ""
	trimmedRight := strings.TrimRight(body, " \t")
	for _, suf := range []string{"*/", "-->"} {
		if strings.HasSuffix(trimmedRight, suf) {
			suffix = suf
			body = strings.TrimSpace(strings.TrimSuffix(trimmedRight, suf))
			break
		}
	}
	if suffix == "" {
		body = trimmedRight
	}
	if !strings.HasSuffix(strings.TrimSpace(body), ";") {
		body += ";"
	}
	body += " UPDATED=" + today
	if suffix != "" {
		body += " " + suffix
	}
	return body + cr, true
}
