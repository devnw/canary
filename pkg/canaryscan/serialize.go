package canaryscan

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Field is one key/value pair of a CANARY token, in source order. Values are
// held decoded: quoting and escaping belong to the wire form only.
type Field struct{ Key, Value string }

var (
	// fieldKeyRe matches a token field key. Keys are preserved verbatim by
	// ParseTokenLine (so a round trip is exact); callers that need the
	// canonical map form get them upper-cased by parseKV.
	fieldKeyRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

	// kvKeyRe locates the first "KEY=" inside a token segment. It is
	// deliberately unanchored, matching the historical parser.
	kvKeyRe = regexp.MustCompile(`([A-Za-z_]+)\s*=\s*`)

	// reqIDRe is the canonical requirement-ID shape, e.g. "CBIN-001".
	reqIDRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*-\d+$`)
	// legacyReqIDRe additionally accepts padded multi-segment legacy IDs,
	// e.g. "REQ-GQL-046" or "BUG-API-001".
	legacyReqIDRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)+-\d+$`)
)

// ParseTokenLine parses one line; ok=false when the line holds no CANARY
// token. CANARY:MIGRATE guidance lines are free text, not KV tokens, and so
// report ok=false as well. A malformed token returns ok=true with err set.
func ParseTokenLine(line string) (fields []Field, ok bool, err error) {
	m := tokenLineRe.FindStringSubmatch(line)
	if m == nil {
		return nil, false, nil
	}
	if isMigrateCapture(m[1]) {
		return nil, false, nil
	}
	fields, err = parseFields(m[1])
	if err != nil {
		return nil, true, err
	}
	return fields, true, nil
}

// SerializeToken renders canonical form: `// CANARY: K=V; K="quoted"`.
// Values that could otherwise be misread are quoted with `\`→`\\` and
// `"`→`\"` escapes; `;` is legal only inside quotes. Control characters
// (C0 and DEL), oversized values, unknown ASPECT/STATUS enum members and
// malformed REQ ids are refused.
func SerializeToken(fields []Field) (string, error) {
	var b strings.Builder
	b.WriteString("// CANARY:")
	for i, f := range fields {
		if !fieldKeyRe.MatchString(f.Key) {
			return "", fmt.Errorf("field key %q is not a valid token key", f.Key)
		}
		if len(f.Value) > MaxFieldBytes {
			return "", fmt.Errorf("field %s exceeds %d bytes", f.Key, MaxFieldBytes)
		}
		if !utf8.ValidString(f.Value) {
			return "", fmt.Errorf("field %s is not valid UTF-8", f.Key)
		}
		for _, r := range f.Value {
			if r < 0x20 || r == 0x7f {
				return "", fmt.Errorf("field %s contains control character", f.Key)
			}
		}
		if err := validateField(f); err != nil {
			return "", err
		}
		if i > 0 {
			b.WriteString(";")
		}
		b.WriteString(" " + f.Key + "=")
		if needsQuote(f.Value) {
			b.WriteString(`"` + escape(f.Value) + `"`)
		} else {
			b.WriteString(f.Value)
		}
	}
	return b.String(), nil
}

// needsQuote reports whether v must be emitted quoted. Beyond the structural
// characters (separator, quote, whitespace, escape) it also quotes anything
// that looks like a template placeholder, because unquoted placeholders are
// the parser's signal to skip a token entirely — quoting keeps a literal
// "<x>" value a value. Values whose edges the parser would trim (any Unicode
// space, not just ASCII) are quoted so those bytes survive the round trip.
func needsQuote(v string) bool {
	return v == "" || v != strings.TrimSpace(v) ||
		strings.ContainsAny(v, ";\" \\<>") ||
		strings.Contains(v, "{{") || strings.Contains(v, "}}") ||
		strings.Contains(v, "%s")
}

func escape(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	return strings.ReplaceAll(v, `"`, `\"`)
}

// validateField enforces the enum and ID constraints the scanner also
// applies, so a token can never be serialized into a form the scanner would
// then reject.
func validateField(f Field) error {
	switch strings.ToUpper(f.Key) {
	case "ASPECT":
		if _, ok := aspects[f.Value]; !ok {
			return fmt.Errorf("invalid ASPECT %q (want one of %s)", f.Value, strings.Join(mapKeys(aspects), ", "))
		}
	case "STATUS":
		if _, ok := statusSet[f.Value]; !ok {
			return fmt.Errorf("invalid STATUS %q (want one of %s)", f.Value, strings.Join(statuses, ", "))
		}
	case "REQ":
		if !reqIDRe.MatchString(f.Value) && !legacyReqIDRe.MatchString(f.Value) {
			return fmt.Errorf("invalid REQ id %q", f.Value)
		}
	}
	return nil
}

// parseFields tokenizes a CANARY token body (everything tokenLineRe captured
// after "CANARY:") into ordered fields. Quoted values may contain ";" and
// escaped quotes; unquoted values keep the historical "up to the next
// separator" behavior.
func parseFields(s string) ([]Field, error) {
	// Template placeholders outside quotes mean this is an example/doc token
	// rather than a real one; it parses to nothing rather than erroring.
	if hasTemplateMarkersOutsideQuotes(s) {
		return nil, nil
	}
	var out []Field
	for _, seg := range splitSegments(s) {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if !strings.Contains(seg, "=") && legacyBareReqRe.MatchString(seg) {
			out = append(out, Field{Key: "REQ", Value: seg})
			continue
		}
		loc := kvKeyRe.FindStringSubmatchIndex(seg)
		if loc == nil {
			return nil, fmt.Errorf("bad kv segment %q", seg)
		}
		value, err := decodeValue(seg[loc[1]:])
		if err != nil {
			return nil, fmt.Errorf("bad kv segment %q: %w", seg, err)
		}
		out = append(out, Field{Key: seg[loc[2]:loc[3]], Value: value})
	}
	return out, nil
}

// splitSegments splits a token body on ";" that are not inside a quoted
// value.
func splitSegments(s string) []string {
	var segs []string
	var cur strings.Builder
	inQuote, esc := false, false
	for _, r := range s {
		switch {
		case esc:
			cur.WriteRune(r)
			esc = false
		case inQuote && r == '\\':
			cur.WriteRune(r)
			esc = true
		case r == '"':
			inQuote = !inQuote
			cur.WriteRune(r)
		case r == ';' && !inQuote:
			segs = append(segs, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(segs, cur.String())
}

// decodeValue decodes a field value. A leading quote starts a quoted value,
// which must terminate and may hold `\\` and `\"` escapes; anything else is
// an unquoted legacy value taken verbatim.
func decodeValue(v string) (string, error) {
	t := strings.TrimSpace(v)
	if !strings.HasPrefix(t, `"`) {
		return t, nil
	}
	// Byte-wise: escapes and the delimiter are all ASCII, and multi-byte
	// sequences pass through untouched, so a value that is not valid UTF-8
	// still round-trips rather than being replaced.
	var b strings.Builder
	esc, closed := false, false
	i := 1
	for ; i < len(t); i++ {
		switch {
		case esc:
			b.WriteByte(t[i])
			esc = false
		case t[i] == '\\':
			esc = true
		case t[i] == '"':
			closed = true
		default:
			b.WriteByte(t[i])
		}
		if closed {
			i++
			break
		}
	}
	if !closed {
		return "", fmt.Errorf("unterminated quoted value")
	}
	if rest := strings.TrimSpace(t[i:]); rest != "" {
		return "", fmt.Errorf("trailing text %q after quoted value", rest)
	}
	return b.String(), nil
}

// hasTemplateMarkersOutsideQuotes reports whether s carries a template
// placeholder in an unquoted region. Markers inside quotes are literal data.
func hasTemplateMarkersOutsideQuotes(s string) bool {
	var plain strings.Builder
	inQuote, esc := false, false
	for _, r := range s {
		switch {
		case esc:
			esc = false
		case inQuote && r == '\\':
			esc = true
		case r == '"':
			inQuote = !inQuote
		case !inQuote:
			plain.WriteRune(r)
		}
	}
	p := plain.String()
	return strings.ContainsAny(p, "<>") ||
		strings.Contains(p, "{{") || strings.Contains(p, "}}") ||
		strings.Contains(p, "%s")
}
