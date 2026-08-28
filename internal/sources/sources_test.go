// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package sources

import "testing"

func testRegistry(t *testing.T) *Registry {
	t.Helper()
	r, err := NewRegistry([]Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"},
		{Name: "app", Type: "gitlab", Key: "GL", URL: "https://gitlab.com/devnw/app/-/issues/{num}"},
		{Name: "oss", Type: "github", Key: "GH", URL: "https://github.com/devnw/app/issues/{num}"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return r
}

func TestCANARY_CBIN_201_RegistryPattern(t *testing.T) {
	r := testRegistry(t)
	for _, id := range []string{"CBIN-105", "PLAT-4521", "GL-88", "GH-7"} {
		if !r.Pattern().MatchString(id) {
			t.Errorf("Pattern should match %s", id)
		}
	}
	if r.Pattern().MatchString("OTHER-123") {
		t.Error("Pattern should not match unconfigured prefix OTHER-123")
	}
}

func TestCANARY_CBIN_201_RegistryResolve(t *testing.T) {
	r := testRegistry(t)
	src, ok := r.Resolve("PLAT-4521")
	if !ok || src.Type != "jira" {
		t.Errorf("Resolve(PLAT-4521) = %+v, %v; want jira source", src, ok)
	}
	if _, ok := r.Resolve("OTHER-1"); ok {
		t.Error("Resolve should fail for unconfigured prefix")
	}
}

func TestCANARY_CBIN_201_TicketURL(t *testing.T) {
	r := testRegistry(t)
	cases := map[string]string{
		"PLAT-4521": "https://company.atlassian.net/browse/PLAT-4521",
		"GL-88":     "https://gitlab.com/devnw/app/-/issues/88",
		"GH-7":      "https://github.com/devnw/app/issues/7",
		"CBIN-105":  "",
		"OTHER-9":   "",
	}
	for id, want := range cases {
		if got := r.TicketURL(id); got != want {
			t.Errorf("TicketURL(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestCANARY_CBIN_201_Normalize(t *testing.T) {
	r := testRegistry(t)
	cases := map[string]string{
		"CBIN-42":   "CBIN-042", // flatfile: padded
		"CBIN-105":  "CBIN-105",
		"PLAT-42":   "PLAT-42", // jira: verbatim, never padded
		"GL-8":      "GL-8",
		"OTHER-7":   "OTHER-7", // unknown: verbatim
	}
	for id, want := range cases {
		if got := r.Normalize(id); got != want {
			t.Errorf("Normalize(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestCANARY_CBIN_201_ClaimPattern(t *testing.T) {
	r := testRegistry(t)
	gap := "✅ CBIN-105\n  ✅ PLAT-4521\n- [ ] GL-88\n✅ OTHER-1\n"
	got := map[string]bool{}
	for _, m := range r.ClaimPattern().FindAllStringSubmatch(gap, -1) {
		got[m[1]] = true
	}
	if !got["CBIN-105"] || !got["PLAT-4521"] {
		t.Errorf("claims missing: %v", got)
	}
	if got["GL-88"] || got["OTHER-1"] {
		t.Errorf("false claims matched: %v", got)
	}
}

func TestCANARY_CBIN_201_DefaultRegistry(t *testing.T) {
	r := Default()
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("Default registry must match CBIN IDs")
	}
	if got := r.Normalize("CBIN-42"); got != "CBIN-042" {
		t.Errorf("Default Normalize(CBIN-42) = %q, want CBIN-042", got)
	}
}

func TestCANARY_CBIN_201_NewRegistryValidation(t *testing.T) {
	if _, err := NewRegistry([]Source{{Name: "a", Type: "jira", Key: "bad-key"}}); err == nil {
		t.Error("lowercase/hyphen key must be rejected")
	}
	if _, err := NewRegistry([]Source{
		{Name: "a", Type: "flatfile", Key: "X"},
		{Name: "b", Type: "jira", Key: "X"},
	}); err == nil {
		t.Error("duplicate keys must be rejected")
	}
	if _, err := NewRegistry([]Source{{Name: "a", Type: "svn", Key: "X"}}); err == nil {
		t.Error("unknown type must be rejected")
	}
	if _, err := NewRegistry(nil); err == nil {
		t.Error("empty registry must be rejected")
	}
}
