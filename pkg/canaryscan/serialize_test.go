package canaryscan

import (
	"strings"
	"testing"
)

// TestSerializeRoundTrip proves the canonical law
// ParseTokenLine(SerializeToken(f)) == f for every field set the serializer
// accepts, including unicode, embedded ";" and embedded quotes.
func TestSerializeRoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		fields []Field
	}{
		{
			name: "plain token",
			fields: []Field{
				{Key: "REQ", Value: "CBIN-001"},
				{Key: "FEATURE", Value: "Scanner"},
				{Key: "ASPECT", Value: "Engine"},
				{Key: "STATUS", Value: "IMPL"},
				{Key: "UPDATED", Value: "2026-01-01"},
			},
		},
		{
			name:   "value with spaces",
			fields: []Field{{Key: "FEATURE", Value: "Streaming Reads"}},
		},
		{
			name:   "value with semicolon",
			fields: []Field{{Key: "FEATURE", Value: "a;b;c"}},
		},
		{
			name:   "value with embedded quotes",
			fields: []Field{{Key: "FEATURE", Value: `say "hi"`}},
		},
		{
			name:   "value with backslash",
			fields: []Field{{Key: "FEATURE", Value: `C:\path\to`}},
		},
		{
			name:   "value with backslash and quote",
			fields: []Field{{Key: "FEATURE", Value: `\"`}},
		},
		{
			name:   "unicode value",
			fields: []Field{{Key: "FEATURE", Value: "ünïcode — 日本語"}},
		},
		{
			name:   "empty value",
			fields: []Field{{Key: "OWNER", Value: ""}},
		},
		{
			name:   "template-looking value",
			fields: []Field{{Key: "FEATURE", Value: "<placeholder> {{x}} %s"}},
		},
		{
			name: "list value",
			fields: []Field{
				{Key: "REQ", Value: "REQ-GQL-046"},
				{Key: "TEST", Value: "TestA,TestB"},
			},
		},
		{
			name:   "max size value",
			fields: []Field{{Key: "FEATURE", Value: strings.Repeat("x", MaxFieldBytes)}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			line, err := SerializeToken(tc.fields)
			if err != nil {
				t.Fatalf("SerializeToken: %v", err)
			}
			got, ok, err := ParseTokenLine(line)
			if err != nil {
				t.Fatalf("ParseTokenLine(%q): %v", line, err)
			}
			if !ok {
				t.Fatalf("ParseTokenLine(%q): ok=false, want a token", line)
			}
			if len(got) != len(tc.fields) {
				t.Fatalf("round-trip field count = %d, want %d (line %q, got %+v)", len(got), len(tc.fields), line, got)
			}
			for i := range tc.fields {
				if got[i] != tc.fields[i] {
					t.Errorf("field %d = %+v, want %+v (line %q)", i, got[i], tc.fields[i], line)
				}
			}
		})
	}
}

// TestSerializeRejectsControls proves C0 control characters and DEL are
// refused rather than emitted into a token line.
func TestSerializeRejectsControls(t *testing.T) {
	cases := []struct {
		name string
		ctl  rune
	}{
		{name: "NUL", ctl: '\x00'},
		{name: "US", ctl: '\x1f'},
		{name: "DEL", ctl: '\x7f'},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := SerializeToken([]Field{{Key: "FEATURE", Value: "bad" + string(tc.ctl) + "value"}})
			if err == nil {
				t.Fatalf("SerializeToken accepted control character %#U", tc.ctl)
			}
		})
	}
}

// TestSerializeRejectsOversizedField proves a field value larger than
// MaxFieldBytes is refused.
func TestSerializeRejectsOversizedField(t *testing.T) {
	_, err := SerializeToken([]Field{{Key: "FEATURE", Value: strings.Repeat("x", MaxFieldBytes+1)}})
	if err == nil {
		t.Fatalf("SerializeToken accepted a %d-byte field value", MaxFieldBytes+1)
	}
}

// TestSerializeValidatesEnums proves ASPECT/STATUS enums and REQ id shape are
// enforced at serialization time.
func TestSerializeValidatesEnums(t *testing.T) {
	cases := []struct {
		name    string
		field   Field
		wantErr bool
	}{
		{name: "valid aspect", field: Field{Key: "ASPECT", Value: "Engine"}},
		{name: "invalid aspect", field: Field{Key: "ASPECT", Value: "Nonsense"}, wantErr: true},
		{name: "valid status", field: Field{Key: "STATUS", Value: "TESTED"}},
		{name: "invalid status", field: Field{Key: "STATUS", Value: "DONE"}, wantErr: true},
		{name: "canonical req", field: Field{Key: "REQ", Value: "CBIN-001"}},
		{name: "legacy padded req", field: Field{Key: "REQ", Value: "REQ-GQL-046"}},
		{name: "invalid req", field: Field{Key: "REQ", Value: "not a req"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := SerializeToken([]Field{tc.field})
			if tc.wantErr && err == nil {
				t.Fatalf("SerializeToken(%+v) = nil error, want error", tc.field)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("SerializeToken(%+v) = %v, want nil error", tc.field, err)
			}
		})
	}
}

// TestParseTokenLineNoToken proves ok=false for lines that hold no CANARY
// token, and that MIGRATE guidance lines are not treated as KV tokens.
func TestParseTokenLineNoToken(t *testing.T) {
	for _, line := range []string{
		"package canaryscan",
		"// just a comment",
		"",
		"// CANARY:MIGRATE move this to the new API",
	} {
		_, ok, err := ParseTokenLine(line)
		if err != nil {
			t.Errorf("ParseTokenLine(%q) err = %v, want nil", line, err)
		}
		if ok {
			t.Errorf("ParseTokenLine(%q) ok = true, want false", line)
		}
	}
}

// TestParseTokenLineQuotedSemicolon proves a quoted value may contain the
// field separator without terminating the field.
func TestParseTokenLineQuotedSemicolon(t *testing.T) {
	line := `// CANARY: REQ=CBIN-001; FEATURE="a; b"; ASPECT=API; STATUS=IMPL; UPDATED=2026-01-01`
	fields, ok, err := ParseTokenLine(line)
	if err != nil {
		t.Fatalf("ParseTokenLine: %v", err)
	}
	if !ok {
		t.Fatal("ParseTokenLine: ok=false")
	}
	if len(fields) != 5 {
		t.Fatalf("got %d fields, want 5: %+v", len(fields), fields)
	}
	if fields[1].Key != "FEATURE" || fields[1].Value != "a; b" {
		t.Errorf("FEATURE field = %+v, want {FEATURE a; b}", fields[1])
	}
}
