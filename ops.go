// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canary

// Core reusable operations extracted from internal/cmds command implementations.
// These functions provide programmatic access for multiple commands and external
// integrations while keeping CLI layer focused on flag parsing & I/O.

import (
	"fmt"
	"sort"
	"strings"

	"go.devnw.com/canary/internal/storage"
)

// GrepTokens returns tokens whose feature/file/test/bench/reqID match pattern (case-insensitive substring).
func GrepTokens(db *storage.DB, pattern string) ([]*storage.Token, error) {
	if pattern == "" {
		return []*storage.Token{}, nil
	}

	// Primary keyword search (feature/req_id/keywords)
	tokens, err := db.SearchTokens(pattern)
	if err != nil {
		return nil, err
	}

	// Secondary scan of all tokens for path/test/bench matches
	all, err := db.ListTokens(nil, "", "", 0)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(pattern)
	index := map[string]*storage.Token{}
	key := func(t *storage.Token) string {
		return fmt.Sprintf("%s:%s:%s:%d", t.ReqID, t.Feature, t.FilePath, t.LineNumber)
	}
	for _, t := range tokens {
		index[key(t)] = t
	}
	for _, t := range all {
		if strings.Contains(strings.ToLower(t.FilePath), lower) ||
			strings.Contains(strings.ToLower(t.Test), lower) ||
			strings.Contains(strings.ToLower(t.Bench), lower) {
			index[key(t)] = t
		}
	}
	out := make([]*storage.Token, 0, len(index))
	for _, t := range index {
		out = append(out, t)
	}
	return out, nil
}

// FormatGrepResults returns human readable list output for grep tokens.
func FormatGrepResults(tokens []*storage.Token) string {
	var b strings.Builder
	for _, t := range tokens {
		fmt.Fprintf(&b, "📌 %s - %s\n", t.ReqID, t.Feature)
		fmt.Fprintf(&b, "   Status: %s | Aspect: %s\n", t.Status, t.Aspect)
		fmt.Fprintf(&b, "   Location: %s:%d\n", t.FilePath, t.LineNumber)
		if t.Test != "" {
			fmt.Fprintf(&b, "   Test: %s\n", t.Test)
		}
		if t.Bench != "" {
			fmt.Fprintf(&b, "   Bench: %s\n", t.Bench)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// FormatGrepResultsByRequirement groups grep tokens by requirement.
func FormatGrepResultsByRequirement(tokens []*storage.Token) string {
	reqMap := map[string][]*storage.Token{}
	for _, t := range tokens {
		reqMap[t.ReqID] = append(reqMap[t.ReqID], t)
	}
	// Stable order
	var reqs []string
	for r := range reqMap {
		reqs = append(reqs, r)
	}
	sort.Strings(reqs)
	var b strings.Builder
	for _, r := range reqs {
		group := reqMap[r]
		fmt.Fprintf(&b, "## %s (%d tokens)\n\n", r, len(group))
		for _, t := range group {
			fmt.Fprintf(&b, "  📌 %s\n", t.Feature)
			fmt.Fprintf(&b, "     Status: %s | Aspect: %s | %s:%d\n", t.Status, t.Aspect, t.FilePath, t.LineNumber)
			if t.Test != "" {
				fmt.Fprintf(&b, "     Test: %s\n", t.Test)
			}
			if t.Bench != "" {
				fmt.Fprintf(&b, "     Bench: %s\n", t.Bench)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// GroupTokens groups tokens by aspect/status (default aspect).
func GroupTokens(tokens []*storage.Token, groupBy string) map[string][]*storage.Token {
	groups := make(map[string][]*storage.Token)
	for _, t := range tokens {
		var k string
		switch groupBy {
		case "status":
			k = t.Status
		case "aspect":
			k = t.Aspect
		default:
			k = t.Aspect
		}
		if k == "" {
			k = "Ungrouped"
		}
		groups[k] = append(groups[k], t)
	}
	return groups
}

// FormatTokensTable renders grouped tokens (used by show command).
func FormatTokensTable(tokens []*storage.Token, groupBy string) string {
	groups := GroupTokens(tokens, groupBy)
	var groupNames []string
	for k := range groups {
		groupNames = append(groupNames, k)
	}
	sort.Strings(groupNames)
	var b strings.Builder
	for _, g := range groupNames {
		b.WriteString(fmt.Sprintf("## %s\n\n", g))
		for _, t := range groups[g] {
			fmt.Fprintf(&b, "📌 %s - %s\n", t.ReqID, t.Feature)
			line := fmt.Sprintf("   Status: %s | Aspect: %s", t.Status, t.Aspect)
			if t.Priority > 0 {
				line += fmt.Sprintf(" | Priority: %d", t.Priority)
			}
			b.WriteString(line + "\n")
			fmt.Fprintf(&b, "   Location: %s:%d\n", t.FilePath, t.LineNumber)
			if t.Test != "" {
				fmt.Fprintf(&b, "   Test: %s\n", t.Test)
			}
			if t.Bench != "" {
				fmt.Fprintf(&b, "   Bench: %s\n", t.Bench)
			}
			if t.Owner != "" {
				fmt.Fprintf(&b, "   Owner: %s\n", t.Owner)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

// FormatFilesList converts fileGroups (map[file][]tokens) into a human-readable summary grouped by aspect.
func FormatFilesList(fileGroups map[string][]*storage.Token) string {
	aspectFiles := map[string][]string{}
	fileCounts := map[string]int{}
	for file, toks := range fileGroups {
		aspects := map[string]struct{}{}
		for _, t := range toks {
			aspects[t.Aspect] = struct{}{}
		}
		for a := range aspects {
			aspectFiles[a] = append(aspectFiles[a], file)
		}
		fileCounts[file] = len(toks)
	}
	var aspectsSorted []string
	for a := range aspectFiles {
		aspectsSorted = append(aspectsSorted, a)
	}
	sort.Strings(aspectsSorted)
	var b strings.Builder
	for _, a := range aspectsSorted {
		files := aspectFiles[a]
		sort.Strings(files)
		fmt.Fprintf(&b, "**%s:**\n", a)
		for _, f := range files {
			c := fileCounts[f]
			plural := "token"
			if c != 1 {
				plural = "tokens"
			}
			fmt.Fprintf(&b, "  %s (%d %s)\n", f, c, plural)
		}
		b.WriteString("\n")
	}
	// Summary
	totalFiles := len(fileGroups)
	totalTokens := 0
	for _, toks := range fileGroups {
		totalTokens += len(toks)
	}
	fmt.Fprintf(&b, "Total: %d files, %d tokens\n", totalFiles, totalTokens)
	return b.String()
}
