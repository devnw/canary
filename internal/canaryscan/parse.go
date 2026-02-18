package canaryscan

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	tokenLineRe = regexp.MustCompile(`(?m)^[ \t]*(?:\/\/|#|\/\*)?[ \t]*CANARY:\s*(.*)$`)
	kvRe        = regexp.MustCompile(`\s*([A-Za-z_]+)\s*=\s*([^;]+)\s*`)
	claimRe     = regexp.MustCompile(`(?m)^\s*✅\s+(CBIN-\d{3})\b`)
)

const defaultSkipPattern = `(^|/)(.git|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|canary-new)(/|$)`

var (
	aspects  = map[string]struct{}{"API": {}, "CLI": {}, "Engine": {}, "Planner": {}, "Storage": {}, "Wire": {}, "Security": {}, "Docs": {}, "Decode": {}, "Encode": {}, "RoundTrip": {}, "Bench": {}, "FrontEnd": {}, "Dist": {}}
	statuses = []string{"MISSING", "STUB", "IMPL", "TESTED", "BENCHED", "REMOVED", "FIXED", "OPEN", "IN_PROGRESS", "VERIFIED", "BLOCKED", "WONTFIX", "DUPLICATE"}
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

func parseKV(s string) (map[string]string, error) {
	out := map[string]string{}
	legacyReqRe := regexp.MustCompile(`^((?:REQ|TASK|BUG)(?:-[A-Z]+)?-?\d{1,4})$`)
	if strings.ContainsAny(s, "<>") || strings.Contains(s, "{{") || strings.Contains(s, "}}") || strings.Contains(s, "%s") {
		return map[string]string{}, nil
	}
	for _, seg := range strings.Split(s, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if !strings.Contains(seg, "=") && legacyReqRe.MatchString(seg) {
			out["REQ"] = normalizeREQ(seg)
			continue
		}
		m := kvRe.FindStringSubmatch(seg)
		if len(m) != 3 {
			return nil, fmt.Errorf("bad kv segment %q", seg)
		}
		out[strings.ToUpper(m[1])] = strings.TrimSpace(m[2])
	}
	// So BUG tokens (BUG=BUG-API-001) are included in scan: use BUG as REQ when REQ missing.
	if out["REQ"] == "" && out["BUG"] != "" {
		v := unquote(out["BUG"])
		if legacyReqRe.MatchString(v) {
			out["REQ"] = normalizeREQ(v)
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

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
		return v[1 : len(v)-1]
	}
	return v
}

func normalizeREQ(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "‑", "-")
	v = strings.ReplaceAll(v, "–", "-")
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

func promote(status string, hasTests, hasBenches bool) string {
	if status == "IMPL" && hasTests {
		status = "TESTED"
	}
	if (status == "IMPL" || status == "TESTED") && hasBenches {
		status = "BENCHED"
	}
	return status
}

func getTimestamp() string {
	if testTS := os.Getenv("CANARY_TEST_TIMESTAMP"); testTS != "" {
		return testTS
	}
	return time.Now().UTC().Format(time.RFC3339)
}
