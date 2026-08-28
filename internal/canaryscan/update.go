package canaryscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// UpdateStaleTokens rewrites UPDATED in source files for tokens matching staleDiags REQ IDs.
// Only updates TESTED/BENCHED tokens. Returns files that were modified.
func UpdateStaleTokens(root string, skip *regexp.Regexp, staleDiags []string) (map[string]bool, error) {
	staleReqs := make(map[string]bool)
	reqRe := regexp.MustCompile(`REQ=([A-Z][A-Z0-9]*-\d+)`)
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
	refTime := time.Now().UTC()
	if t := RefTimeFromEnv(); !t.IsZero() {
		refTime = t
	}
	today := refTime.Format("2006-01-02")
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
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if !tokenLineRe.Match(content) {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		modified := false
		for i, line := range lines {
			if !tokenLineRe.MatchString(line) {
				continue
			}
			match := tokenLineRe.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}
			attrs, err := parseKV(match[1])
			if err != nil {
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
			updatedRe := regexp.MustCompile(`(UPDATED=)([0-9]{4}-[0-9]{2}-[0-9]{2})`)
			if updatedRe.MatchString(line) {
				lines[i] = updatedRe.ReplaceAllString(line, fmt.Sprintf("${1}%s", today))
				modified = true
			}
		}
		if modified {
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
