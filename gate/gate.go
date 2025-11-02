package gate

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Global registry for comment styles overrideable at runtime.
var DefaultCommentStyles = map[string]CommentStyle{
	".go":    {LinePrefix: "//"},
	".js":    {LinePrefix: "//"},
	".ts":    {LinePrefix: "//"},
	".tsx":   {LinePrefix: "//"},
	".jsx":   {LinePrefix: "//"},
	".java":  {LinePrefix: "//"},
	".c":     {LinePrefix: "//"},
	".h":     {LinePrefix: "//"},
	".cpp":   {LinePrefix: "//"},
	".rs":    {LinePrefix: "//"},
	".swift": {LinePrefix: "//"},
	".cs":    {LinePrefix: "//"},
	".zig":   {LinePrefix: "//"},
	".py":    {LinePrefix: "#"},
	".sh":    {LinePrefix: "#"},
	".rb":    {LinePrefix: "#"},
	".yml":   {LinePrefix: "#"},
	".yaml":  {LinePrefix: "#"},
	".toml":  {LinePrefix: "#"},
	".env":   {LinePrefix: "#"},
	".sql":   {LinePrefix: "--"},
	".html":  {LinePrefix: "<!--", BlockEnd: "-->", Space: true},
	".md":    {LinePrefix: "<!--", BlockEnd: "-->", Space: true},
	".xml":   {LinePrefix: "<!--", BlockEnd: "-->", Space: true},
}

// RegisterCommentStyle allows external code to override or add a comment style for extension.
func RegisterCommentStyle(ext string, style CommentStyle) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	DefaultCommentStyles[strings.ToLower(ext)] = style
}

// CANARY: REQ=CBIN-150; FEATURE="GenericSectionGater"; ASPECT=API; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-02
// Package gate provides generic file section gating utilities supporting
// multiple comment syntaxes (single-line and multi-line) via an Options pattern.

// CommentStyle describes how comments are represented for a language/file type.
// Either single-line (LinePrefix) or multi-line (BlockStart/BlockEnd) or both may be set.
type CommentStyle struct {
	// LinePrefix denotes a single-line comment marker (e.g. "//", "#", "--").
	LinePrefix string
	// BlockStart and BlockEnd denote multi-line comment delimiters (e.g. "/*" and "*/").
	BlockStart string
	BlockEnd   string
	// Optional space insertion between delimiters and content for readability.
	Space bool
}

// Options configures gating behavior.
type Options struct {
	Key        string // The logical gating namespace, e.g. "CANARY"
	Style      CommentStyle
	StartToken string // e.g. "START"
	EndToken   string // e.g. "END"
	// If true, enforce blank line preceding an appended section.
	EnsureBlankBefore bool
}

// Option functional option.
type Option func(*Options)

// WithStyle sets the comment style.
func WithStyle(cs CommentStyle) Option { return func(o *Options) { o.Style = cs } }

// WithKey sets gating key.
func WithKey(k string) Option { return func(o *Options) { o.Key = k } }

// WithTokens sets start/end tokens.
func WithTokens(start, end string) Option {
	return func(o *Options) { o.StartToken, o.EndToken = start, end }
}

// WithBlankLineBefore enables blank line insertion.
func WithBlankLineBefore() Option { return func(o *Options) { o.EnsureBlankBefore = true } }

// DefaultOptions returns a baseline configuration targeting HTML-style comments used in markdown.
func DefaultOptions() Options {
	return Options{
		Key: "CANARY",
		Style: CommentStyle{
			LinePrefix: "<!--",
			BlockEnd:   "-->",
			Space:      true,
		},
		StartToken:        "START",
		EndToken:          "END",
		EnsureBlankBefore: true,
	}
}

// buildStartMarker builds a start marker based on options and section key.
func buildStartMarker(o Options, section string) string {
	if o.Style.LinePrefix != "" && o.Style.BlockEnd != "" {
		// HTML-like or single-line with closing token semantics
		// Format: <!-- KEY:section:START --> or <!-- KEY:START --> if section empty
		// Ensure exactly one space after <!-- when Space true
		prefix := o.Style.LinePrefix
		if o.Style.Space && strings.HasSuffix(prefix, "<!--") {
			prefix = "<!--" // canonical
		}
		if section == "" {
			return fmt.Sprintf("%s %s:%s %s", prefix, o.Key, o.StartToken, o.Style.BlockEnd)
		}
		return fmt.Sprintf("%s %s:%s:%s %s", prefix, o.Key, section, o.StartToken, o.Style.BlockEnd)
	}
	if o.Style.BlockStart != "" && o.Style.BlockEnd != "" {
		// Multi-line block comment style (e.g., /* KEY:section:START */)
		space := ""
		if o.Style.Space {
			space = " "
		}
		if section == "" {
			return fmt.Sprintf("%s%s%s:%s%s%s", o.Style.BlockStart, space, o.Key, o.StartToken, space, o.Style.BlockEnd)
		}
		return fmt.Sprintf("%s%s%s:%s:%s%s%s", o.Style.BlockStart, space, o.Key, section, o.StartToken, space, o.Style.BlockEnd)
	}
	// Fallback single-line only: // KEY:section:START
	if section == "" {
		return fmt.Sprintf("%s %s:%s", o.Style.LinePrefix, o.Key, o.StartToken)
	}
	return fmt.Sprintf("%s %s:%s:%s", o.Style.LinePrefix, o.Key, section, o.StartToken)
}

// buildEndMarker builds an end marker based on options and section key.
func buildEndMarker(o Options, section string) string {
	if o.Style.LinePrefix != "" && o.Style.BlockEnd != "" {
		prefix := o.Style.LinePrefix
		if o.Style.Space && strings.HasSuffix(prefix, "<!--") {
			prefix = "<!--"
		}
		if section == "" {
			return fmt.Sprintf("%s %s:%s %s", prefix, o.Key, o.EndToken, o.Style.BlockEnd)
		}
		return fmt.Sprintf("%s %s:%s:%s %s", prefix, o.Key, section, o.EndToken, o.Style.BlockEnd)
	}
	if o.Style.BlockStart != "" && o.Style.BlockEnd != "" {
		space := ""
		if o.Style.Space {
			space = " "
		}
		if section == "" {
			return fmt.Sprintf("%s%s%s:%s%s%s", o.Style.BlockStart, space, o.Key, o.EndToken, space, o.Style.BlockEnd)
		}
		return fmt.Sprintf("%s%s%s:%s:%s%s%s", o.Style.BlockStart, space, o.Key, section, o.EndToken, space, o.Style.BlockEnd)
	}
	if section == "" {
		return fmt.Sprintf("%s %s:%s", o.Style.LinePrefix, o.Key, o.EndToken)
	}
	return fmt.Sprintf("%s %s:%s:%s", o.Style.LinePrefix, o.Key, section, o.EndToken)
}

// StartMarker returns the start marker string for a keyed or unnamed section given options.
func StartMarker(section string, opts ...Option) string {
	o := DefaultOptions()
	for _, fn := range opts { fn(&o) }
	m := buildStartMarker(o, section)
	if section == "" {
		// Normalize potential double colons defensively
		m = strings.ReplaceAll(m, "::", ":")
	}
	return m
}

// EndMarker returns the end marker string for a keyed or unnamed section given options.
func EndMarker(section string, opts ...Option) string {
	o := DefaultOptions()
	for _, fn := range opts { fn(&o) }
	m := buildEndMarker(o, section)
	if section == "" {
		m = strings.ReplaceAll(m, "::", ":")
	}
	return m
}

// BuildGatedBody returns a gated body snippet (unnamed section) with start/end markers
// surrounding the provided content. Newlines are normalized so the snippet always
// ends with a trailing newline. The content is placed exactly between markers.
// Example (markdown):
// <!-- CANARY:START -->\n<content>\n<!-- CANARY:END -->\n
func BuildGatedBody(content string, opts ...Option) string {
	o := DefaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	start := buildStartMarker(o, "")
	end := buildEndMarker(o, "")
	// Normalize double colons if any (defensive for unnamed)
	start = strings.ReplaceAll(start, "::", ":")
	end = strings.ReplaceAll(end, "::", ":")
	var b strings.Builder
	b.WriteString(start)
	b.WriteByte('\n')
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(end)
	b.WriteByte('\n')
	return b.String()
}

// BuildGatedBodyKey returns a gated body snippet for a keyed section.
// Example (markdown):
// <!-- CANARY:mykey:START -->\n<content>\n<!-- CANARY:mykey:END -->\n
func BuildGatedBodyKey(key, content string, opts ...Option) string {
	o := DefaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	start := buildStartMarker(o, key)
	end := buildEndMarker(o, key)
	var b strings.Builder
	b.WriteString(start)
	b.WriteByte('\n')
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(end)
	b.WriteByte('\n')
	return b.String()
}

// UpdateSingle updates or inserts one gated section in filePath.
func UpdateSingle(filePath, sectionContent string, opts ...Option) error {
	o := DefaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	start := buildStartMarker(o, "") // unnamed section uses empty key => KEY::START
	end := buildEndMarker(o, "")
	// Normalize markers for unnamed (remove double colons)
	start = strings.ReplaceAll(start, "::", ":")
	end = strings.ReplaceAll(end, "::", ":")
	return updateSectionInternal(filePath, start, end, sectionContent, o.EnsureBlankBefore)
}

// UpdateMultiple updates or inserts multiple keyed sections.
func UpdateMultiple(filePath string, sections map[string]string, opts ...Option) error {
	o := DefaultOptions()
	for _, fn := range opts {
		fn(&o)
	}

	existing, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read file: %w", err)
	}
	if os.IsNotExist(err) {
		// Create new file with all sections
		var buf bytes.Buffer
		for key, content := range sections {
			buf.WriteString(buildStartMarker(o, key) + "\n")
			buf.WriteString(content)
			buf.WriteString("\n" + buildEndMarker(o, key) + "\n\n")
		}
		return os.WriteFile(filePath, buf.Bytes(), 0644)
	}

	scanner := bufio.NewScanner(bytes.NewReader(existing))
	var result []string
	skipping := false
	found := map[string]bool{}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		// Start detection: look for KEY:<key>:START pattern inside line
		if k, isStart := matchStart(trimmed, o); isStart {
			found[k] = true
			result = append(result, line) // keep marker
			if newContent, ok := sections[k]; ok {
				result = append(result, newContent)
				skipping = true // skip original content until end marker
			} else {
				skipping = false
			}
			continue
		}
		if _, isEnd := matchEnd(trimmed, o); isEnd {
			// end of section
			skipping = false
			result = append(result, line)
			continue
		}
		if skipping {
			continue
		}
		result = append(result, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file: %w", err)
	}

	// Append missing sections
	for k, content := range sections {
		if !found[k] {
			if o.EnsureBlankBefore && len(result) > 0 && result[len(result)-1] != "" {
				result = append(result, "")
			}
			result = append(result, buildStartMarker(o, k))
			result = append(result, content)
			result = append(result, buildEndMarker(o, k))
		}
	}

	updated := strings.Join(result, "\n")
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return os.WriteFile(filePath, []byte(updated), 0644)
}

// RemoveSection removes a keyed section.
func RemoveSection(filePath, key string, opts ...Option) error {
	o := DefaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	existing, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(existing))
	startMarker := buildStartMarker(o, key)
	endMarker := buildEndMarker(o, key)
	in := false
	var result []string
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == startMarker {
			in = true
			continue
		}
		if trimmed == endMarker {
			in = false
			continue
		}
		if !in {
			result = append(result, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file: %w", err)
	}
	updated := strings.Join(result, "\n")
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return os.WriteFile(filePath, []byte(updated), 0644)
}

// internal helper for unnamed single section update
func updateSectionInternal(filePath, startMarker, endMarker, content string, ensureBlank bool) error {
	existing, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			out := fmt.Sprintf("%s\n%s\n%s\n", startMarker, content, endMarker)
			return os.WriteFile(filePath, []byte(out), 0644)
		}
		return fmt.Errorf("read file: %w", err)
	}
	lines := strings.Split(string(existing), "\n")
	var result []string
	in := false
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == startMarker {
			in = true
			found = true
			result = append(result, line)
			result = append(result, content)
			continue
		}
		if trimmed == endMarker {
			in = false
			result = append(result, line)
			continue
		}
		if in {
			continue
		}
		result = append(result, line)
	}
	if !found {
		if ensureBlank && len(result) > 0 && result[len(result)-1] != "" {
			result = append(result, "")
		}
		result = append(result, startMarker, content, endMarker)
	}
	updated := strings.Join(result, "\n")
	return os.WriteFile(filePath, []byte(updated), 0644)
}

// matchStart attempts to parse a start marker returning the section key.
func matchStart(line string, o Options) (string, bool) {
	// Accept patterns KEY:<key>:START inside supported comment style
	if !strings.Contains(line, o.Key+":") || !strings.Contains(line, ":"+o.StartToken) {
		return "", false
	}
	// Extract between KEY: and :START
	// Find KEY:
	idx := strings.Index(line, o.Key+":")
	if idx == -1 {
		return "", false
	}
	rest := line[idx+len(o.Key)+1:]
	endIdx := strings.Index(rest, ":"+o.StartToken)
	if endIdx == -1 {
		return "", false
	}
	key := rest[:endIdx]
	key = strings.TrimSpace(key)
	return key, true
}

// matchEnd attempts to parse an end marker returning the section key (ignored).
func matchEnd(line string, o Options) (string, bool) {
	if !strings.Contains(line, o.Key+":") || !strings.Contains(line, ":"+o.EndToken) {
		return "", false
	}
	idx := strings.Index(line, o.Key+":")
	if idx == -1 {
		return "", false
	}
	rest := line[idx+len(o.Key)+1:]
	endIdx := strings.Index(rest, ":"+o.EndToken)
	if endIdx == -1 {
		return "", false
	}
	key := rest[:endIdx]
	key = strings.TrimSpace(key)
	return key, true
}

// DetectStyleFromExtension returns a reasonable default style based on file extension.
func DetectStyleFromExtension(path string) CommentStyle {
	ext := strings.ToLower(filepath.Ext(path))
	if style, ok := DefaultCommentStyles[ext]; ok {
		return style
	}
	return CommentStyle{LinePrefix: "//"}
}
