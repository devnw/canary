package canaryscan

import "testing"

// FuzzSerializeRoundTrip asserts the canonical law: anything SerializeToken
// accepts, ParseTokenLine reads back byte-for-byte.
func FuzzSerializeRoundTrip(f *testing.F) {
	f.Add("Name", "value with ; and \" and \\ and ünïcode")
	f.Fuzz(func(t *testing.T, k, v string) {
		fields := []Field{{Key: "FEATURE", Value: v}}
		line, err := SerializeToken(fields)
		if err != nil {
			return // rejected inputs are fine
		}
		got, ok, perr := ParseTokenLine(line)
		if perr != nil || !ok {
			t.Fatalf("round-trip parse failed: %v", perr)
		}
		if got[0].Value != v {
			t.Fatalf("round-trip mismatch %q != %q", got[0].Value, v)
		}
	})
}
