// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package upgrade rewrites legacy on-disk CANARY token shapes into the
// current parseable form: markdown `# CANARY:` headings into HTML comments,
// unicode hyphens inside IDs into ASCII hyphens, unpadded flatfile IDs into
// zero-padded IDs, bare legacy ID segments into keyed REQ= tokens, bug
// tokens missing FEATURE= into scan-parseable single lines, STATUS=FIXED
// into STATUS=REMOVED, missing UPDATED= into stamped tokens, the old
// multi-line bug-create continuation shape into one line, and (when an ID
// map is supplied) old requirement IDs into new ones across both CANARY
// tokens and GAP_ANALYSIS.md "✅ <ID>" claim lines.
//
// Every rule is independently selectable (Options.Rules) and independently
// safe: rules never touch CANARY:MIGRATE / CANARY:START / CANARY:END lines,
// and — including the md-heading rule — never touch lines inside fenced
// code blocks in markdown files: a fenced `# CANARY:` heading is a
// documentation example, not a live token, so it is left alone like every
// other rule's fenced content.
// CANARY: REQ=CP-275; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_AtomicWrite,TestCANARY_CBIN_302_CRLF,TestCANARY_CBIN_302_Combined,TestCANARY_CBIN_302_FenceProtection,TestCANARY_CBIN_302_Idempotent,TestCANARY_CBIN_302_MDHeadingFenceProtection,TestCANARY_CBIN_302_MigrateGuard,TestCANARY_CBIN_302_PlaceholderGuard,TestCANARY_CBIN_302_Remap,TestCANARY_CBIN_302_RemapCollision,TestCANARY_CBIN_302_Rules,TestCANARY_CBIN_302_RuleFiltering,TestCANARY_CBIN_302_UnicodeHyphenProse,TestCANARY_CBIN_302_ValidRule,TestGuardFailureSkipsFileAndContinues; UPDATED=2026-08-30
package upgrade

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/safewrite"
	"devnw.dev/canary/pkg/sources"
)

// Change is one proposed (or, when Options.Write is true, applied) edit
// produced by a single named rule.
type Change struct {
	File string // root-relative, forward-slashed
	Line int    // 1-based, in the file state at the moment this rule ran
	Old  string
	New  string
	Rule string
}

// Options configures Run.
type Options struct {
	Root     string
	Skip     *regexp.Regexp
	Ignore   *ignore.GitIgnore
	Registry *sources.Registry
	Map      map[string]string // old REQ/BUG id -> new id
	Write    bool              // false = dry run
	Today    string            // YYYY-MM-DD for added UPDATED=; empty -> CANARY_TEST_TIMESTAMP then time.Now().UTC()
	Rules    []string          // empty = all rules, see AllRules
}

// AllRules lists every named rule in canonical execution order.
var AllRules = []string{
	"join-multiline",
	"md-heading",
	"unicode-hyphen",
	"bare-id",
	"bug-alias",
	"status-fixed",
	"pad-flatfile",
	"add-updated",
	"remap",
}

func ruleSet(rules []string) map[string]bool {
	m := make(map[string]bool, len(AllRules))
	if len(rules) == 0 {
		for _, r := range AllRules {
			m[r] = true
		}
		return m
	}
	for _, r := range rules {
		m[r] = true
	}
	return m
}

// ValidRule reports whether name is a known rule.
func ValidRule(name string) bool {
	for _, r := range AllRules {
		if r == name {
			return true
		}
	}
	return false
}

// Run walks Options.Root, applies the enabled rules to every non-skipped
// file, and returns every Change found (in file, then rule-application,
// order). When Options.Write is true, modified files are rewritten in
// place (permissions preserved); otherwise Run never touches disk.
//
// A file that fails the token-preservation guard, or whose write fails, is
// skipped and left byte-identical; the walk continues and the returned error
// joins every such failure, naming each file. The changes found are returned
// either way, so a partial write reports both what it did and what it refused.
func Run(o Options) ([]Change, error) {
	root := o.Root
	if root == "" {
		root = "."
	}
	skip := o.Skip
	if skip == nil {
		skip = canaryscan.DefaultSkipRegex()
	}
	reg := o.Registry
	if reg == nil {
		var err error
		reg, err = sources.LoadFromRoot(root)
		if err != nil {
			return nil, fmt.Errorf("load .canary/project.yaml: %w", err)
		}
	}
	today := resolveToday(o.Today)
	enabled := ruleSet(o.Rules)

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "" {
			rel = "."
		}
		if o.Ignore != nil && o.Ignore.MatchesPath(rel) {
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
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	var all []Change
	// A file this walk refuses to write is a per-file verdict, not a reason to
	// abandon the remaining files: aborting mid-walk leaves the files already
	// rewritten with no report of what stopped, and every later file unvisited.
	// Failures are collected here, the offending file is skipped unwritten, and
	// the joined error names every one of them at the end -- so the exit is
	// still nonzero and the report is complete.
	var writeErrs []error
	for _, path := range files {
		info, ierr := os.Stat(path)
		if ierr != nil {
			continue
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			continue
		}
		if !isProbablyText(b) {
			continue
		}
		content := string(b)
		newContent, changes := upgradeFile(content, filepath.Ext(path), enabled, reg, o.Map, today)
		if len(changes) == 0 {
			continue
		}
		rel, rerr2 := filepath.Rel(root, path)
		if rerr2 != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		for i := range changes {
			changes[i].File = rel
		}
		all = append(all, changes...)
		if o.Write && newContent != content {
			// A rule is a regex rewrite over other people's source: the one
			// thing it must never do is make a requirement disappear. Prove
			// every pre-existing token survives before anything is written.
			if gerr := checkTokensPreserved(content, newContent, enabled); gerr != nil {
				writeErrs = append(writeErrs, fmt.Errorf("%s: %w (file left unchanged)", rel, gerr))
				continue
			}
			if _, werr := safewrite.Write(path, []byte(newContent), info.Mode(), safewrite.Options{
				Root:   root,
				Force:  true, // --write is an explicit instruction to rewrite these files
				Backup: true,
			}); werr != nil {
				writeErrs = append(writeErrs, fmt.Errorf("write %s: %w", path, werr))
				continue
			}
		}
	}
	return all, errors.Join(writeErrs...)
}

// ---- token-preservation guard -----------------------------------------------------

// ruleTouches lists, per rule, the token field keys that rule is allowed to
// change. The guard normalizes exactly these keys away before comparing a
// file's tokens before and after a rewrite, so a rule's intended edit is never
// mistaken for a lost token — and an edit to any *other* field, or the loss of
// a whole token, still is.
var ruleTouches = map[string][]string{
	"join-multiline": nil, // folds continuation lines; changes no field's value
	"md-heading":     nil, // changes the comment syntax, not the token
	"unicode-hyphen": {"REQ", "BUG", "TASK", "PARENT"},
	"bare-id":        {"REQ"},
	"bug-alias":      {"FEATURE"},
	"status-fixed":   {"STATUS"},
	"pad-flatfile":   {"REQ", "BUG"},
	"add-updated":    {"UPDATED"},
	"remap":          {"REQ", "BUG"},
}

// htmlTokenRe matches an HTML-comment CANARY line. The scanner's grammar does
// not accept that comment style, but the md-heading rule produces it, so the
// guard normalizes it back into a parseable line: md-heading changes a token's
// comment syntax, and the guard must not read that as the token vanishing.
var htmlTokenRe = regexp.MustCompile(`^\s*<!--\s*(CANARY:.*?)\s*-->\s*$`)

// checkTokensPreserved reports an error when any CANARY token present in pre
// has no counterpart in post, once the enabled rules' own field edits are
// normalized away. Tokens are compared as a multiset, so dropping one of
// several identical tokens is caught too.
//
// A post-token counts as a counterpart when it still carries every field the
// pre-token had. Growing a token is allowed -- join-multiline folds a token's
// own continuation lines into it, and bug-alias adds an aliased field -- but
// losing a field, or losing a whole token, is not.
func checkTokensPreserved(pre, post string, enabled map[string]bool) error {
	touched := touchedKeys(enabled)
	before := tokenShapes(pre, touched)
	after := tokenShapes(post, touched)
	used := make([]bool, len(after))
	for _, want := range before {
		if !claimCounterpart(want, after, used) {
			return fmt.Errorf("token preservation: %s has no counterpart after the rewrite", want.display)
		}
	}
	return nil
}

// touchedKeys is the union of the field keys the enabled rules may change,
// upper-cased for case-insensitive comparison against token field keys.
func touchedKeys(enabled map[string]bool) map[string]bool {
	out := map[string]bool{}
	for rule, keys := range ruleTouches {
		if !enabled[rule] {
			continue
		}
		for _, k := range keys {
			out[strings.ToUpper(k)] = true
		}
	}
	return out
}

// tokenShape is one token reduced to the parts a rewrite must not lose.
type tokenShape struct {
	display string         // canonical rendering, for the error message
	exact   string         // the fields joined, for cheap equality
	fields  map[string]int // canonical "KEY=value" -> occurrences
}

// claimCounterpart marks the first unused post-token that still carries every
// field of want. An exact match is preferred over a superset so a token is
// never consumed by a larger one that a different pre-token needs.
func claimCounterpart(want tokenShape, after []tokenShape, used []bool) bool {
	for _, exactOnly := range []bool{true, false} {
		for i, got := range after {
			if used[i] {
				continue
			}
			if exactOnly && got.exact != want.exact {
				continue
			}
			if !exactOnly && !containsFields(got.fields, want.fields) {
				continue
			}
			used[i] = true
			return true
		}
	}
	return false
}

// containsFields reports whether got holds at least as many of every field as
// want does.
func containsFields(got, want map[string]int) bool {
	for f, n := range want {
		if got[f] < n {
			return false
		}
	}
	return true
}

// tokenShapes reduces every CANARY token in content to its comparable shape.
func tokenShapes(content string, touched map[string]bool) []tokenShape {
	var out []tokenShape
	for _, line := range strings.Split(content, "\n") {
		shape, ok := tokenShapeOf(line, touched)
		if !ok {
			continue
		}
		out = append(out, shape)
	}
	return out
}

// tokenShapeOf reduces one line's token to the fields no enabled rule is
// allowed to change. A line that holds no token reports ok=false. A malformed
// token cannot be normalized, so its raw text stands in as its identity --
// that keeps the guard conservative rather than blind: a malformed token that
// disappears is still reported.
func tokenShapeOf(line string, touched map[string]bool) (tokenShape, bool) {
	if m := htmlTokenRe.FindStringSubmatch(line); m != nil {
		line = "// " + m[1]
	}
	fields, ok, err := canaryscan.ParseTokenLine(line)
	if !ok {
		return tokenShape{}, false
	}
	if err != nil {
		raw := "malformed:" + strings.TrimSpace(line)
		return tokenShape{display: raw, exact: raw, fields: map[string]int{raw: 1}}, true
	}

	kept := make([]canaryscan.Field, 0, len(fields))
	counts := make(map[string]int, len(fields))
	var parts []string
	for _, f := range fields {
		if touched[strings.ToUpper(f.Key)] {
			continue
		}
		f.Value = canonicalValue(f.Value)
		kept = append(kept, f)
		canon := strings.ToUpper(f.Key) + "=" + f.Value
		counts[canon]++
		parts = append(parts, canon)
	}
	sort.Strings(parts)

	display, serr := canaryscan.SerializeToken(kept)
	if serr != nil {
		// SerializeToken refuses values the scanner would also refuse
		// (unknown enum members, malformed ids). Such a token still has to
		// be tracked, so the sorted field list stands in for its rendering.
		display = "token[" + strings.Join(parts, " ") + "]"
	}
	return tokenShape{display: display, exact: strings.Join(parts, ";"), fields: counts}, true
}

// canonicalValue strips a block-comment terminator the token grammar leaves
// glued to the last field's value ("STATUS=IMPL */"). The terminator belongs
// to the comment, not the value, and a rule that relocates it -- add-updated
// appends its field before the terminator -- must not read as a changed field.
func canonicalValue(v string) string {
	v = strings.TrimSpace(v)
	for _, suffix := range []string{"*/", "-->"} {
		if strings.HasSuffix(v, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(v, suffix))
		}
	}
	return v
}

// resolveToday implements Options.Today's documented fallback chain.
func resolveToday(today string) string {
	if today != "" {
		return today
	}
	if ts := os.Getenv("CANARY_TEST_TIMESTAMP"); ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			return t.Format("2006-01-02")
		}
		if len(ts) >= len("2006-01-02") {
			return ts[:len("2006-01-02")]
		}
		return ts
	}
	return time.Now().UTC().Format("2006-01-02")
}

// isProbablyText is a cheap NUL-byte binary guard so Run never reads huge
// binary blobs (e.g. stray database files not caught by skip patterns) as
// line-oriented text.
func isProbablyText(b []byte) bool {
	n := len(b)
	if n > 2<<20 { // 2MB: never worth scanning for legacy token shapes
		return false
	}
	if n > 8000 {
		n = 8000
	}
	for i := 0; i < n; i++ {
		if b[i] == 0 {
			return false
		}
	}
	return true
}

// ---- shared line-shape regexes -------------------------------------------------

var (
	// tokenLineRe mirrors pkg/canaryscan/parse.go's tokenLineRe: group 1 is
	// everything after "CANARY:" on the line. CP-285: prefix group also
	// accepts "--" (SQL line-comment) so the upgrade rewriter (e.g. --map)
	// sees and remaps tokens in .sql files, matching the scanner.
	tokenLineRe = regexp.MustCompile(`^[ \t]*(?://|#|/\*|--)?[ \t]*CANARY:\s*(.*)$`)
	// lineMarkerRe splits a line into leading whitespace (1), an optional
	// comment marker (2), the whitespace following it (3), and the rest (4).
	lineMarkerRe = regexp.MustCompile(`^(\s*)(//|#|/\*)?(\s*)(.*)$`)
	// legacyReqRe mirrors parse.go's bare-ID pattern.
	legacyReqRe = regexp.MustCompile(`^((?:REQ|TASK|BUG)(?:-[A-Z]+)?-?\d{1,4})$`)

	reReqPad = regexp.MustCompile(`^(REQ(?:-[A-Z]+)?-)(\d{1,3})$`)
	reTBPad  = regexp.MustCompile(`^((?:TASK|BUG)-)(\d{1,3})$`)

	unicodeHyphenReplacer = strings.NewReplacer("‑", "-", "–", "-")

	kvKeyRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// isGuardedLine reports whether line must never be touched by any rule: a
// migrate-guidance marker, or a section start/end marker. Each marker's
// literal text is split with a "+" below (and never begins a comment line
// in this doc comment) so this file's own source can never be mistaken by
// the scanner for one of the markers it describes.
func isGuardedLine(line string) bool {
	return strings.Contains(line, "CANARY:MIGRATE") ||
		strings.Contains(line, "CANARY:START") ||
		strings.Contains(line, "CANARY:END")
}

// isPlaceholderContent reports whether content looks like a template/example
// snippet (Go template syntax, angle-bracket placeholders, printf verbs)
// rather than a real token — mirrors canaryscan's parseKV escape hatch.
func isPlaceholderContent(content string) bool {
	return strings.ContainsAny(content, "<>") ||
		strings.Contains(content, "{{") ||
		strings.Contains(content, "}}") ||
		strings.Contains(content, "%s")
}

// isPureKVList reports whether s (a candidate join-multiline continuation's
// content) consists entirely of `KEY=value;`-shaped segments.
func isPureKVList(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	any := false
	for _, p := range strings.Split(s, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.Index(p, "=")
		if eq <= 0 {
			return false
		}
		if !kvKeyRe.MatchString(strings.TrimSpace(p[:eq])) {
			return false
		}
		any = true
	}
	return any
}

// padNumeric zero-pads a bare legacy ID's numeric suffix to at least 3
// digits. Its only caller (bareIDContent) only ever passes ids matched by
// bareRe's `(?:REQ|TASK|BUG)(?:-[A-Z]+)?-?\d{1,4}` shape — a bare "CBIN-N"
// segment is never a legacy shape this rule handles (CBIN ids only ever
// appear keyed, e.g. "REQ=CBIN-101", which pad-flatfile handles instead) —
// so only the REQ and TASK/BUG branches are reachable here.
func padNumeric(id string) string {
	pad := func(prefix, num string) string {
		for len(num) < 3 {
			num = "0" + num
		}
		return prefix + num
	}
	if m := reReqPad.FindStringSubmatch(id); len(m) == 3 {
		return pad(m[1], m[2])
	}
	if m := reTBPad.FindStringSubmatch(id); len(m) == 3 {
		return pad(m[1], m[2])
	}
	return id
}

// ---- fenced code block tracking (markdown only) ---------------------------------

// fenceMask returns, for each line, whether that line sits strictly inside a
// ``` or ~~~ fenced code block. Fence delimiter lines themselves are marked
// using the state *before* the toggle (they never look like CANARY content
// anyway).
func fenceMask(lines []string) []bool {
	mask := make([]bool, len(lines))
	inFence := false
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			mask[i] = inFence
			inFence = !inFence
			continue
		}
		mask[i] = inFence
	}
	return mask
}

// ---- file-level rules: join-multiline, md-heading --------------------------------

// applyJoinMultiline folds the legacy bug-create continuation shape
// (comment-prefixed `KEY=value;` lines following a CANARY: line) into a
// single logical line.
func applyJoinMultiline(lines []string, fence []bool) ([]string, []Change) {
	var out []string
	var changes []Change
	i := 0
	for i < len(lines) {
		line := lines[i]
		if isGuardedLine(line) || (fence != nil && fence[i]) || !tokenLineRe.MatchString(line) {
			out = append(out, line)
			i++
			continue
		}
		hm := lineMarkerRe.FindStringSubmatch(line)
		marker := hm[2]
		if marker == "" {
			// Only fold the historical comment-prefixed continuation shape;
			// a bare, marker-less CANARY line never has folded continuations.
			out = append(out, line)
			i++
			continue
		}

		j := i + 1
		var contParts []string
		for j < len(lines) {
			cline := lines[j]
			if isGuardedLine(cline) || (fence != nil && fence[j]) {
				break
			}
			cm := lineMarkerRe.FindStringSubmatch(cline)
			if cm[2] != marker {
				break
			}
			content := cm[4]
			if strings.Contains(content, "CANARY:") {
				break
			}
			if !isPureKVList(content) {
				break
			}
			contParts = append(contParts, strings.TrimSpace(content))
			j++
		}

		if len(contParts) == 0 {
			out = append(out, line)
			i++
			continue
		}

		head := strings.TrimRight(line, " \t")
		if !strings.HasSuffix(strings.TrimSpace(head), ";") {
			head += ";"
		}
		newLine := head + " " + strings.Join(contParts, " ")
		changes = append(changes, Change{
			Line: i + 1,
			Old:  strings.Join(lines[i:j], "\n"),
			New:  newLine,
			Rule: "join-multiline",
		})
		out = append(out, newLine)
		i = j
	}
	return out, changes
}

// applyMDHeading converts `# CANARY:` markdown headings into
// `<!-- CANARY: ... -->` HTML comments. A `# CANARY:` line inside a fenced
// code block is a documentation example showing the old shape (e.g. a
// skill's "here's what a legacy heading looked like" sample), not a live
// token, so — like every other rule — it consults the fence mask and
// leaves fenced content untouched.
//
// Converted `<!-- CANARY: ... -->` tokens stay readable in the rendered
// source (a human, or a tool that greps for the literal marker, can still
// find them), but `canary scan` does not parse HTML comments, so they are
// never counted toward status.json's requirement totals and cannot back a
// live requirement. This rule primarily serves doc examples (e.g. a skill
// or command doc illustrating a legacy token shape), not requirements that
// need scan visibility.
func applyMDHeading(lines []string, fence []bool) []Change {
	var changes []Change
	for i, line := range lines {
		if isGuardedLine(line) || (fence != nil && fence[i]) {
			continue
		}
		if !strings.HasPrefix(line, "# CANARY:") {
			continue
		}
		newLine := strings.Replace(line, "# CANARY:", "<!-- CANARY:", 1) + " -->"
		changes = append(changes, Change{Line: i + 1, Old: line, New: newLine, Rule: "md-heading"})
		lines[i] = newLine
	}
	return changes
}

// ---- per-line token-content rules -------------------------------------------------

// contentRule transforms the CANARY: token content (everything after
// "CANARY:" on the line) and reports whether it changed anything.
type contentRule func(content string) (string, bool)

func unicodeHyphenContent(content string) (string, bool) {
	segs := strings.Split(content, ";")
	changed := false
	kvRe := regexp.MustCompile(`(?i)^(\s*(?:REQ|BUG|TASK|PARENT)\s*=\s*)(.*)$`)
	for i, seg := range segs {
		if m := kvRe.FindStringSubmatch(seg); m != nil {
			nv := unicodeHyphenReplacer.Replace(m[2])
			if nv != m[2] {
				segs[i] = m[1] + nv
				changed = true
			}
			continue
		}
		trimmed := strings.TrimSpace(seg)
		if trimmed == "" || strings.Contains(trimmed, "=") {
			continue
		}
		nv := unicodeHyphenReplacer.Replace(trimmed)
		if nv != trimmed && legacyReqRe.MatchString(nv) {
			lead := seg[:len(seg)-len(strings.TrimLeft(seg, " \t"))]
			segs[i] = lead + nv
			changed = true
		}
	}
	if !changed {
		return content, false
	}
	return strings.Join(segs, ";"), true
}

func bareIDContent(content string) (string, bool) {
	segs := strings.Split(content, ";")
	bareRe := regexp.MustCompile(`^(\s*)((?:REQ|TASK|BUG)(?:-[A-Z]+)?-?\d{1,4})(\s*)$`)
	changed := false
	for i, seg := range segs {
		m := bareRe.FindStringSubmatch(seg)
		if m == nil {
			continue
		}
		newSeg := m[1] + "REQ=" + padNumeric(m[2]) + m[3]
		if newSeg != seg {
			segs[i] = newSeg
			changed = true
		}
	}
	if !changed {
		return content, false
	}
	return strings.Join(segs, ";"), true
}

// bugAliasContent fixes the historical bug-create defect: a BUG=-only token
// (no REQ=) that carries TITLE= but not FEATURE= fails the scanner's
// required-field gate. Aliasing TITLE's value into FEATURE= makes the line
// parseable while keeping BUG= as the primary key, per the rule's contract.
func bugAliasContent(content string) (string, bool) {
	segs := strings.Split(content, ";")
	hasBUG, hasREQ, hasFEATURE := false, false, false
	titleIdx := -1
	titleVal := ""
	for i, seg := range segs {
		trimmed := strings.TrimSpace(seg)
		if trimmed == "" {
			continue
		}
		eq := strings.Index(trimmed, "=")
		if eq <= 0 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(trimmed[:eq]))
		val := strings.TrimSpace(trimmed[eq+1:])
		switch key {
		case "BUG":
			hasBUG = true
		case "REQ":
			hasREQ = true
		case "FEATURE":
			hasFEATURE = true
		case "TITLE":
			titleIdx = i
			titleVal = val
		}
	}
	if !hasBUG || hasREQ || hasFEATURE || titleIdx < 0 {
		return content, false
	}
	newSegs := make([]string, 0, len(segs)+1)
	newSegs = append(newSegs, segs[:titleIdx+1]...)
	newSegs = append(newSegs, " FEATURE="+titleVal)
	newSegs = append(newSegs, segs[titleIdx+1:]...)
	return strings.Join(newSegs, ";"), true
}

func statusFixedContent(content string) (string, bool) {
	segs := strings.Split(content, ";")
	segRe := regexp.MustCompile(`^(\s*STATUS\s*=\s*)FIXED(\s*)$`)
	changed := false
	for i, seg := range segs {
		if m := segRe.FindStringSubmatch(seg); m != nil {
			segs[i] = m[1] + "REMOVED" + m[2]
			changed = true
		}
	}
	if !changed {
		return content, false
	}
	return strings.Join(segs, ";"), true
}

func padFlatfileContent(content string, reg *sources.Registry) (string, bool) {
	if reg == nil {
		return content, false
	}
	segs := strings.Split(content, ";")
	kvRe := regexp.MustCompile(`(?i)^(\s*(?:REQ|BUG)\s*=\s*)(.*)$`)
	changed := false
	for i, seg := range segs {
		m := kvRe.FindStringSubmatch(seg)
		if m == nil {
			continue
		}
		val := strings.TrimSpace(m[2])
		norm := reg.Normalize(val)
		if norm != val {
			segs[i] = m[1] + norm
			changed = true
		}
	}
	if !changed {
		return content, false
	}
	return strings.Join(segs, ";"), true
}

func addUpdatedContent(content, today string) (string, bool) {
	// A CRLF file, split on "\n" alone, leaves a trailing "\r" attached to
	// this line's content. Strip it before any suffix/whitespace analysis
	// and re-append it at the very end, or it ends up embedded mid-line
	// (e.g. "...STATUS=IMPL\r; UPDATED=...") instead of trailing the line.
	cr := ""
	body0 := content
	if strings.HasSuffix(body0, "\r") {
		cr = "\r"
		body0 = strings.TrimSuffix(body0, "\r")
	}

	kvRe := regexp.MustCompile(`(?i)^\s*UPDATED\s*=`)
	for _, seg := range strings.Split(body0, ";") {
		if kvRe.MatchString(seg) {
			return content, false
		}
	}
	if strings.TrimSpace(body0) == "" {
		return content, false
	}
	body := body0
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

func remapTokenContent(content string, idMap map[string]string) (string, bool) {
	if len(idMap) == 0 {
		return content, false
	}
	segs := strings.Split(content, ";")
	kvRe := regexp.MustCompile(`(?i)^(\s*(?:REQ|BUG)\s*=\s*)(.*)$`)
	changed := false
	for i, seg := range segs {
		m := kvRe.FindStringSubmatch(seg)
		if m == nil {
			continue
		}
		val := strings.TrimSpace(m[2])
		quote := ""
		unquoted := val
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			quote = `"`
			unquoted = val[1 : len(val)-1]
		}
		newID, ok := idMap[unquoted]
		if !ok {
			continue
		}
		segs[i] = m[1] + quote + newID + quote
		changed = true
	}
	if !changed {
		return content, false
	}
	return strings.Join(segs, ";"), true
}

// ---- remap claim lines (.md GAP_ANALYSIS-style "✅ <ID>" lines) -------------------

func applyRemapClaims(lines []string, fence []bool, idMap map[string]string) []Change {
	if len(idMap) == 0 {
		return nil
	}
	oldIDs := make([]string, 0, len(idMap))
	for k := range idMap {
		oldIDs = append(oldIDs, k)
	}
	sort.Strings(oldIDs)

	var changes []Change
	for i, line := range lines {
		if isGuardedLine(line) || (fence != nil && fence[i]) {
			continue
		}
		for _, oldID := range oldIDs {
			re := regexp.MustCompile(`^(\s*✅\s+)` + regexp.QuoteMeta(oldID) + `\b`)
			if !re.MatchString(line) {
				continue
			}
			newLine := re.ReplaceAllString(line, "${1}"+idMap[oldID])
			if newLine != line {
				changes = append(changes, Change{Line: i + 1, Old: line, New: newLine, Rule: "remap"})
				lines[i] = newLine
			}
			break
		}
	}
	return changes
}

// ---- per-file orchestration --------------------------------------------------------

// upgradeFile runs every enabled rule against content (one file's text) and
// returns the possibly-rewritten content plus every Change made.
func upgradeFile(content, ext string, enabled map[string]bool, reg *sources.Registry, idMap map[string]string, today string) (string, []Change) {
	trailingNL := strings.HasSuffix(content, "\n")
	lines := strings.Split(content, "\n")
	isMD := strings.EqualFold(ext, ".md")

	var all []Change

	if enabled["join-multiline"] {
		var fence []bool
		if isMD {
			fence = fenceMask(lines)
		}
		var changes []Change
		lines, changes = applyJoinMultiline(lines, fence)
		all = append(all, changes...)
	}

	if isMD && enabled["md-heading"] {
		changes := applyMDHeading(lines, fenceMask(lines))
		all = append(all, changes...)
	}

	var fence []bool
	if isMD {
		fence = fenceMask(lines)
	}

	rules := []struct {
		name string
		fn   contentRule
	}{
		{"unicode-hyphen", unicodeHyphenContent},
		{"bare-id", bareIDContent},
		{"bug-alias", bugAliasContent},
		{"status-fixed", statusFixedContent},
		{"pad-flatfile", func(c string) (string, bool) { return padFlatfileContent(c, reg) }},
		{"add-updated", func(c string) (string, bool) { return addUpdatedContent(c, today) }},
		{"remap", func(c string) (string, bool) { return remapTokenContent(c, idMap) }},
	}

	for i, line := range lines {
		if isGuardedLine(line) || (fence != nil && fence[i]) {
			continue
		}
		m := tokenLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		contentPart := m[1]
		if isPlaceholderContent(contentPart) {
			continue
		}
		prefix := line[:len(line)-len(contentPart)]
		cur := contentPart
		for _, r := range rules {
			if !enabled[r.name] {
				continue
			}
			next, changed := r.fn(cur)
			if !changed {
				continue
			}
			oldLine := prefix + cur
			newLine := prefix + next
			all = append(all, Change{Line: i + 1, Old: oldLine, New: newLine, Rule: r.name})
			cur = next
		}
		if cur != contentPart {
			lines[i] = prefix + cur
		}
	}

	if isMD && enabled["remap"] {
		fence = fenceMask(lines) // recompute: line count may have shifted from join-multiline only, but content-only edits above don't change line count
		changes := applyRemapClaims(lines, fence, idMap)
		all = append(all, changes...)
	}

	out := strings.Join(lines, "\n")
	if trailingNL && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out, all
}
