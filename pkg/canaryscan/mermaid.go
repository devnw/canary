// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CP-268; FEATURE="MermaidRefs"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_202_ExtractDiagramRefs; UPDATED=2026-08-28
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

// DiagramRef records one requirement-ID mention inside a mermaid diagram.
type DiagramRef struct {
	ReqID string
	File  string
	Line  int // 1-based
}

// fence state machine states for ExtractDiagramRefs.
const (
	outside = iota
	inMermaidBlock
	inOtherBlock
)

// ExtractDiagramRefs finds requirement IDs inside ```mermaid fenced blocks.
// relPath ending in .mmd is treated as a whole-file mermaid diagram (no
// fences required). IDs are normalized through reg (flatfile padding);
// reg nil means sources.Default() (the legacy CBIN series).
func ExtractDiagramRefs(relPath, content string, reg *sources.Registry) []DiagramRef {
	if reg == nil {
		reg = sources.Default()
	}
	wholeFile := strings.HasSuffix(relPath, ".mmd")
	var refs []DiagramRef

	state := outside
	if wholeFile {
		state = inMermaidBlock
	}
	for i, line := range strings.Split(content, "\n") {
		if !wholeFile {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") {
				switch state {
				case outside:
					if strings.HasPrefix(trimmed, "```mermaid") {
						state = inMermaidBlock
					} else {
						state = inOtherBlock
					}
				default: // inMermaidBlock or inOtherBlock: this fence closes it
					state = outside
				}
				continue
			}
		}
		if state != inMermaidBlock {
			continue
		}
		for _, m := range reg.Pattern().FindAllString(line, -1) {
			refs = append(refs, DiagramRef{ReqID: reg.Normalize(m), File: relPath, Line: i + 1})
		}
	}
	return refs
}

// ScanDiagramRefs walks root for .md/.markdown/.mmd files and extracts all
// requirement-ID mentions inside mermaid diagrams. Paths in the result are
// root-relative with forward slashes. ignorePatterns (from LoadCanaryIgnore)
// is honored the same way Scan honors it: relative path matched first, dirs
// skipped via SkipDir; nil means no .canaryignore patterns apply.
func ScanDiagramRefs(root string, skip *regexp.Regexp, reg *sources.Registry, ignorePatterns *ignore.GitIgnore) ([]DiagramRef, error) {
	if skip == nil {
		skip = DefaultSkipRegex()
	}
	var out []DiagramRef
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
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" && ext != ".mmd" {
			return nil
		}
		if skip.MatchString(path) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil // unreadable markdown is not fatal to a scan
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			rel = path
		}
		out = append(out, ExtractDiagramRefs(filepath.ToSlash(rel), string(b), reg)...)
		return nil
	})
	return out, err
}
