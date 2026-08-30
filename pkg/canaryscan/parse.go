package canaryscan

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"devnw.dev/canary/pkg/sources"
)

var (
	// CP-285: prefix group also accepts "--" (SQL line-comment) so tokens
	// in .sql files scan. Alternatives share no common prefix character
	// class with each other, so ordering here does not create shadowing;
	// "--" is listed last purely for readability, matching upgrade.go.
	tokenLineRe = regexp.MustCompile(`(?m)^[ \t]*(?:\/\/|#|\/\*|--)?[ \t]*CANARY:\s*(.*)$`)

	// legacyBareReqRe matches an ID-only segment (no "="), e.g. "REQ-1" or
	// "BUG-API-001", which older tokens used in place of "REQ=".
	legacyBareReqRe = regexp.MustCompile(`^((?:REQ|TASK|BUG)(?:-[A-Z]+)?-?\d{1,4})$`)
)

const defaultSkipPattern = `(^|/)(.git|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|canary-new)(/|$)`

var (
	aspects   = map[string]struct{}{"API": {}, "CLI": {}, "Engine": {}, "Planner": {}, "Storage": {}, "Wire": {}, "Security": {}, "Docs": {}, "Decode": {}, "Encode": {}, "RoundTrip": {}, "Bench": {}, "FrontEnd": {}, "Dist": {}}
	statuses  = []string{"MISSING", "STUB", "IMPL", "TESTED", "BENCHED", "REMOVED", "FIXED", "OPEN", "IN_PROGRESS", "VERIFIED", "BLOCKED", "WONTFIX", "DUPLICATE"}
	statusSet = func() map[string]struct{} {
		m := map[string]struct{}{}
		for _, s := range statuses {
			m[s] = struct{}{}
		}
		return m
	}()
)

// DefaultSkipRegex returns the default skip path regex.
func DefaultSkipRegex() *regexp.Regexp {
	r, _ := regexp.Compile(defaultSkipPattern)
	return r
}

// migrateCapturePrefix marks a tokenLineRe capture as CANARY:MIGRATE
// free-text guidance rather than a KV token.
const migrateCapturePrefix = "MIGRATE"

// isMigrateCapture reports whether m (the capture group of a tokenLineRe
// match) is a CANARY:MIGRATE guidance line. Such lines must never reach
// parseKV — that is what keeps a MIGRATE line from ever aborting a scan.
func isMigrateCapture(m string) bool {
	return strings.HasPrefix(strings.TrimSpace(m), migrateCapturePrefix)
}

// parseKV is the map view of parseFields: keys upper-cased, REQ normalized
// through reg (nil means the built-in padding rules). Later fields with the
// same key win, matching the historical behavior.
func parseKV(s string, reg *sources.Registry) (map[string]string, error) {
	fields, err := parseFields(s)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, f := range fields {
		k := strings.ToUpper(f.Key)
		if k == "REQ" && legacyBareReqRe.MatchString(f.Value) {
			out[k] = normalizeREQWithRegistry(f.Value, reg)
			continue
		}
		out[k] = f.Value
	}
	// So BUG tokens (BUG=BUG-API-001) are included in scan: use BUG as REQ when REQ missing.
	if out["REQ"] == "" && out["BUG"] != "" {
		v := unquote(out["BUG"])
		if legacyBareReqRe.MatchString(v) {
			out["REQ"] = normalizeREQWithRegistry(v, reg)
		} else {
			out["REQ"] = v
		}
	}
	return out, nil
}

func splitList(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == ',' })
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// unquote strips a matched pair of legacy SINGLE quotes ('...') from v.
// Double-quote decoding is already performed upstream by decodeValue
// (serialize.go) when a token field is parsed, so a value that legitimately
// starts and ends with literal '"' characters (e.g. a decoded FEATURE value
// of `"quoted"`) must not be stripped again here.
func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return v[1 : len(v)-1]
	}
	return v
}

// normalizeREQWithRegistry canonicalizes a requirement ID: unicode dashes are
// folded to ASCII, registry-resolved IDs follow their source's rules, and
// legacy numeric suffixes are zero-padded to three digits. A nil registry
// means "padding rules only".
func normalizeREQWithRegistry(v string, reg *sources.Registry) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "‑", "-")
	v = strings.ReplaceAll(v, "–", "-")
	if reg != nil {
		if s, ok := reg.Resolve(v); ok {
			if s.Type == "flatfile" {
				return reg.Normalize(v)
			}
			return v // external ticket IDs are verbatim, never padded
		}
	}
	pad := func(prefix, num string) string {
		for len(num) < 3 {
			num = "0" + num
		}
		return prefix + num
	}
	if m := regexp.MustCompile(`^(CBIN-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		return pad(m[1], m[2])
	}
	if m := regexp.MustCompile(`^(REQ(?:-[A-Z]+)?-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		return pad(m[1], m[2])
	}
	if m := regexp.MustCompile(`^((?:TASK|BUG)-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		return pad(m[1], m[2])
	}
	return v
}

func getTimestamp() string {
	if testTS := os.Getenv("CANARY_TEST_TIMESTAMP"); testTS != "" {
		return testTS
	}
	return time.Now().UTC().Format(time.RFC3339)
}
