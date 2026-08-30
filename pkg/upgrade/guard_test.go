// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package upgrade

import "testing"

// TestTokenPreservationGuard exercises the guard directly, because the abort
// path it exists for is reachable only from a rule bug -- there is (by
// design) no rule that can produce it today, so a black-box test could not
// prove the guard actually refuses anything.
// CANARY: REQ=CP-275; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestTokenPreservationGuard; UPDATED=2026-08-30
func TestTokenPreservationGuard(t *testing.T) {
	const token = `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; OWNER=team; UPDATED=2025-01-01` + "\n"

	cases := []struct {
		name    string
		pre     string
		post    string
		rules   []string
		wantErr bool
	}{
		{
			name:  "unchanged file",
			pre:   token,
			post:  token,
			rules: AllRules,
		},
		{
			name:    "token deleted",
			pre:     token + "func x() {}\n",
			post:    "func x() {}\n",
			rules:   AllRules,
			wantErr: true,
		},
		{
			name:    "one of two identical tokens deleted",
			pre:     token + token,
			post:    token,
			rules:   AllRules,
			wantErr: true,
		},
		{
			name:    "field outside the rule's remit changed",
			pre:     token,
			post:    `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; OWNER=someone-else; UPDATED=2025-01-01` + "\n",
			rules:   []string{"status-fixed"},
			wantErr: true,
		},
		{
			name:  "field the rule declares it touches changed",
			pre:   token,
			post:  `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=REMOVED; OWNER=team; UPDATED=2025-01-01` + "\n",
			rules: []string{"status-fixed"},
		},
		{
			name:  "token grown by folding its continuation lines",
			pre:   "// CANARY: BUG=BUG-API-001; TITLE=\"T\";\n//         ASPECT=API; STATUS=OPEN\n",
			post:  "// CANARY: BUG=BUG-API-001; TITLE=\"T\"; ASPECT=API; STATUS=OPEN\n",
			rules: []string{"join-multiline"},
		},
		{
			name:  "markdown heading converted to an HTML comment",
			pre:   `# CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-01-01` + "\n",
			post:  `<!-- CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-01-01 -->` + "\n",
			rules: []string{"md-heading"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkTokensPreserved(tc.pre, tc.post, ruleSet(tc.rules))
			if tc.wantErr && err == nil {
				t.Fatal("expected the guard to refuse this rewrite")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("guard refused a legitimate rewrite: %v", err)
			}
		})
	}
}
