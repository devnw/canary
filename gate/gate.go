package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"devnw.dev/canary/pkg/safewrite"
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

// CANARY: REQ=ENG-4321; FEATURE="GenericSectionGater"; ASPECT=API; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-02
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

// buildMarker builds one marker line: the configured comment style wrapped
// around "KEY:section:token", or "KEY:token" for the legacy unnamed section.
// Both markers of a section come from this one builder, so a start marker and
// its end marker can never drift apart in shape.
func buildMarker(o Options, section, token string) string {
	body := o.Key + ":" + token
	if section != "" {
		body = o.Key + ":" + section + ":" + token
	}
	switch {
	case o.Style.LinePrefix != "" && o.Style.BlockEnd != "":
		// HTML-like: a line prefix that also carries a closing delimiter,
		// e.g. <!-- KEY:section:START -->
		return o.Style.LinePrefix + " " + body + " " + o.Style.BlockEnd
	case o.Style.BlockStart != "" && o.Style.BlockEnd != "":
		// Multi-line block comment style, e.g. /* KEY:section:START */
		space := ""
		if o.Style.Space {
			space = " "
		}
		return o.Style.BlockStart + space + body + space + o.Style.BlockEnd
	default:
		// Single-line only, e.g. # KEY:section:START
		return o.Style.LinePrefix + " " + body
	}
}

// buildStartMarker builds a start marker based on options and section key.
func buildStartMarker(o Options, section string) string {
	return buildMarker(o, section, o.StartToken)
}

// buildEndMarker builds an end marker based on options and section key.
func buildEndMarker(o Options, section string) string {
	return buildMarker(o, section, o.EndToken)
}

// StartMarker returns the start marker string for a keyed or unnamed section given options.
func StartMarker(section string, opts ...Option) string {
	return buildStartMarker(resolveOptions(opts), section)
}

// EndMarker returns the end marker string for a keyed or unnamed section given options.
func EndMarker(section string, opts ...Option) string {
	return buildEndMarker(resolveOptions(opts), section)
}

// resolveOptions applies opts over the defaults.
func resolveOptions(opts []Option) Options {
	o := DefaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// BuildGatedBody returns a gated body snippet (unnamed section) with start/end markers
// surrounding the provided content. Newlines are normalized so the snippet always
// ends with a trailing newline. The content is placed exactly between markers.
// Example (markdown):
// <!-- CANARY:START -->\n<content>\n<!-- CANARY:END -->\n
func BuildGatedBody(content string, opts ...Option) string {
	o := resolveOptions(opts)
	start := buildStartMarker(o, "")
	end := buildEndMarker(o, "")
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
	o := resolveOptions(opts)
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

// markerSpan is one marker-delimited region located in a file: the line
// indices of its start and end markers, and the section key they name
// (empty for a legacy unnamed section).
type markerSpan struct {
	key        string
	start, end int
}

// markerKey reports whether a trimmed line is exactly the marker carrying
// token, and if so which section key it names. Matching is exact on the whole
// trimmed line -- a line that merely mentions a marker in prose, in a code
// fence, or as part of a longer sentence is not a marker, and is left alone.
func markerKey(trimmed string, o Options, token string) (string, bool) {
	if trimmed == buildMarker(o, "", token) {
		return "", true
	}
	// Derive the keyed marker's fixed prefix and suffix from the builder
	// itself (via a sentinel key), so recognition can never disagree with
	// generation for any configured comment style.
	const sentinel = "\x00key\x00"
	tmpl := buildMarker(o, sentinel, token)
	i := strings.Index(tmpl, sentinel)
	if i < 0 {
		return "", false
	}
	pre, suf := tmpl[:i], tmpl[i+len(sentinel):]
	if len(trimmed) <= len(pre)+len(suf) ||
		!strings.HasPrefix(trimmed, pre) || !strings.HasSuffix(trimmed, suf) {
		return "", false
	}
	key := trimmed[len(pre) : len(trimmed)-len(suf)]
	if key == "" || strings.ContainsAny(key, " \t:") {
		return "", false
	}
	return key, true
}

// scanMarkers locates every marker section in lines and validates the file's
// whole marker structure before any rewrite is contemplated. A start without
// a matching end, an end before any start, a nested start, a mismatched end
// key, or two sections sharing one key are all errors: the file's intent is
// ambiguous, and guessing is how a "repair" ends up appending a duplicate
// section. The caller reports the error and leaves the file byte-identical.
func scanMarkers(lines []string, o Options) ([]markerSpan, error) {
	var spans []markerSpan
	seen := map[string]int{}
	open, openKey := -1, ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if k, ok := markerKey(trimmed, o, o.StartToken); ok {
			if open >= 0 {
				return nil, fmt.Errorf("line %d: %s marker nested inside section %s opened at line %d",
					i+1, o.StartToken, describeKey(openKey), open+1)
			}
			open, openKey = i, k
			continue
		}
		k, ok := markerKey(trimmed, o, o.EndToken)
		if !ok {
			continue
		}
		if open < 0 {
			return nil, fmt.Errorf("line %d: %s marker with no matching %s", i+1, o.EndToken, o.StartToken)
		}
		if k != openKey {
			return nil, fmt.Errorf("line %d: %s marker for %s closes section %s opened at line %d",
				i+1, o.EndToken, describeKey(k), describeKey(openKey), open+1)
		}
		if prev, dup := seen[k]; dup {
			return nil, fmt.Errorf("line %d: duplicate section %s (already defined at line %d)",
				open+1, describeKey(k), prev)
		}
		seen[k] = open + 1
		spans = append(spans, markerSpan{key: k, start: open, end: i})
		open, openKey = -1, ""
	}

	if open >= 0 {
		return nil, fmt.Errorf("line %d: %s marker for %s has no matching %s",
			open+1, o.StartToken, describeKey(openKey), o.EndToken)
	}
	return spans, nil
}

// describeKey renders a section key for an error message, naming the legacy
// unnamed section rather than printing an empty string.
func describeKey(key string) string {
	if key == "" {
		return "(unnamed)"
	}
	return strconv.Quote(key)
}

// UpdateSection updates or inserts the gated section named key in filePath.
//
// key must be non-empty. An unkeyed section cannot be created any more: two
// tools writing unkeyed sections into the same file have no way to tell their
// sections apart, which is how a file ends up with duplicates. An unkeyed
// section that already exists is still adopted -- its body is replaced in
// place and its markers left as they are -- so files written by older
// versions keep updating rather than growing a second section beside them.
func UpdateSection(filePath, key, sectionContent string, opts ...Option) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("gate: a section key is required to update %s", filePath)
	}
	return update(filePath, map[string]string{key: sectionContent}, true, resolveOptions(opts))
}

// UpdateMultiple updates or inserts multiple keyed sections.
func UpdateMultiple(filePath string, sections map[string]string, opts ...Option) error {
	for k := range sections {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("gate: a section key is required to update %s", filePath)
		}
	}
	return update(filePath, sections, false, resolveOptions(opts))
}

// update is the single rewrite path behind UpdateSection and UpdateMultiple.
// It validates the file's marker structure first, replaces the body of each
// section it recognizes, appends the sections that are genuinely absent, and
// writes the result atomically. A file it cannot make sense of is never
// touched.
func update(filePath string, sections map[string]string, adoptUnnamed bool, o Options) error {
	existing, err := os.ReadFile(filePath) //nolint:gosec // caller-supplied managed file
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read file: %w", err)
	}
	lines := splitLines(string(existing))

	spans, serr := scanMarkers(lines, o)
	if serr != nil {
		return fmt.Errorf("%s: %w", filePath, serr)
	}
	spanByKey := make(map[string]markerSpan, len(spans))
	spanAt := make(map[int]markerSpan, len(spans))
	for _, s := range spans {
		spanByKey[s.key] = s
		spanAt[s.start] = s
	}

	keys := make([]string, 0, len(sections))
	for k := range sections {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	replaceAt := map[int]string{}
	var missing []string
	for _, k := range keys {
		if s, ok := spanByKey[k]; ok {
			replaceAt[s.start] = k
			continue
		}
		if s, ok := spanByKey[""]; adoptUnnamed && ok {
			replaceAt[s.start] = k
			continue
		}
		missing = append(missing, k)
	}

	var out []string
	for i := 0; i < len(lines); i++ {
		s, isStart := spanAt[i]
		key, replace := replaceAt[i]
		if isStart && replace {
			out = append(out, lines[s.start], sections[key], lines[s.end])
			i = s.end
			continue
		}
		out = append(out, lines[i])
	}
	for _, k := range missing {
		if o.EnsureBlankBefore && len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, buildStartMarker(o, k), sections[k], buildEndMarker(o, k))
	}

	return writeLines(filePath, out)
}

// RemoveSection removes a keyed section.
func RemoveSection(filePath, key string, opts ...Option) error {
	o := resolveOptions(opts)
	existing, err := os.ReadFile(filePath) //nolint:gosec // caller-supplied managed file
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	lines := splitLines(string(existing))

	spans, serr := scanMarkers(lines, o)
	if serr != nil {
		return fmt.Errorf("%s: %w", filePath, serr)
	}
	drop := map[int]markerSpan{}
	for _, s := range spans {
		if s.key == key {
			drop[s.start] = s
		}
	}

	var out []string
	for i := 0; i < len(lines); i++ {
		if s, ok := drop[i]; ok {
			i = s.end
			continue
		}
		out = append(out, lines[i])
	}
	return writeLines(filePath, out)
}

// splitLines splits a file's text into lines, treating a single trailing
// newline as a terminator rather than an extra empty line, so a rewrite
// round-trips a file's trailing whitespace unchanged.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// writeLines joins lines, terminates the file with exactly one newline, and
// replaces filePath atomically. Gating merges the caller's section into
// content the user owns, so replacement is always intended here -- what must
// never happen is a half-written file, which the staged write rules out.
func writeLines(filePath string, lines []string) error {
	data := strings.Join(lines, "\n") + "\n"
	_, err := safewrite.Write(filePath, []byte(data), 0o640, safewrite.Options{
		Root:  filepath.Dir(filePath),
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("write %s: %w", filePath, err)
	}
	return nil
}

// DetectStyleFromExtension returns a reasonable default style based on file extension.
func DetectStyleFromExtension(path string) CommentStyle {
	ext := strings.ToLower(filepath.Ext(path))
	if style, ok := DefaultCommentStyles[ext]; ok {
		return style
	}
	return CommentStyle{LinePrefix: "//"}
}
