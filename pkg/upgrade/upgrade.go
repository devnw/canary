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
// CANARY: REQ=CP-275; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_AtomicWrite,TestCANARY_CBIN_302_CRLF,TestCANARY_CBIN_302_Combined,TestCANARY_CBIN_302_FenceProtection,TestCANARY_CBIN_302_Idempotent,TestCANARY_CBIN_302_MDHeadingFenceProtection,TestCANARY_CBIN_302_MigrateGuard,TestCANARY_CBIN_302_PlaceholderGuard,TestCANARY_CBIN_302_Remap,TestCANARY_CBIN_302_RemapCollision,TestCANARY_CBIN_302_Rules,TestCANARY_CBIN_302_RuleFiltering,TestCANARY_CBIN_302_UnicodeHyphenProse,TestCANARY_CBIN_302_ValidRule; UPDATED=2026-08-29
package upgrade

import (
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
		reg = sources.LoadFromRoot(root)
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
			if werr := atomicWriteFile(path, []byte(newContent), info.Mode()); werr != nil {
				return all, fmt.Errorf("write %s: %w", path, werr)
			}
		}
	}
	return all, nil
}

// atomicWriteFile replaces path's content without ever leaving it truncated:
// it writes to a temp file in the same directory (so the final rename is
// same-filesystem and atomic on POSIX), chmods the temp file to mode, closes
// it, then renames it over path. On any error the temp file is removed and
// the original at path is left untouched.
func atomicWriteFile(path string, data []byte, mode fs.FileMode) (err error) {
	dir := filepath.Dir(path)
	tmp, cerr := os.CreateTemp(dir, ".canary-upgrade-*")
	if cerr != nil {
		return cerr
	}
	tmpPath := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, werr := tmp.Write(data); werr != nil {
		return fmt.Errorf("write temp file: %w", werr)
	}
	if cherr := tmp.Chmod(mode); cherr != nil {
		return fmt.Errorf("chmod temp file: %w", cherr)
	}
	if cerr := tmp.Close(); cerr != nil {
		return fmt.Errorf("close temp file: %w", cerr)
	}
	if rerr := os.Rename(tmpPath, path); rerr != nil {
		return fmt.Errorf("rename temp file: %w", rerr)
	}
	return nil
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

// isGuardedLine reports whether line must never be touched by any rule:
// CANARY:MIGRATE, CANARY:START, CANARY:END markers.
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
