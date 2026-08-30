package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildGatedBodyUnnamedMarkdown(t *testing.T) {
	snippet := BuildGatedBody("Hello World",
		WithStyle(CommentStyle{LinePrefix: "<!--", BlockEnd: "-->", Space: true}),
		WithKey("CANARY"),
		WithTokens("START", "END"),
	)
	if !strings.Contains(snippet, "<!-- CANARY:START -->") {
		t.Errorf("expected start marker, got: %s", snippet)
	}
	if !strings.Contains(snippet, "<!-- CANARY:END -->") {
		t.Errorf("expected end marker, got: %s", snippet)
	}
	if !strings.Contains(snippet, "Hello World") {
		t.Errorf("expected body content present")
	}
	// Ensure exactly one start/end marker each
	if strings.Count(snippet, "<!-- CANARY:START -->") != 1 || strings.Count(snippet, "<!-- CANARY:END -->") != 1 {
		t.Errorf("unexpected duplicate markers: %s", snippet)
	}
}

func TestBuildGatedBodyKeyMarkdown(t *testing.T) {
	snippet := BuildGatedBodyKey("intro", "Intro Body",
		WithStyle(CommentStyle{LinePrefix: "<!--", BlockEnd: "-->", Space: true}),
		WithKey("CANARY"),
		WithTokens("START", "END"),
	)
	if !strings.Contains(snippet, "<!-- CANARY:intro:START -->") {
		t.Errorf("expected keyed start marker, got: %s", snippet)
	}
	if !strings.Contains(snippet, "<!-- CANARY:intro:END -->") {
		t.Errorf("expected keyed end marker, got: %s", snippet)
	}
	if !strings.Contains(snippet, "Intro Body") {
		t.Errorf("expected body content present")
	}
}

func TestBuildGatedBodyCustomLineStyle(t *testing.T) {
	snippet := BuildGatedBodyKey("cfg", "configuration here",
		WithStyle(CommentStyle{LinePrefix: "#"}),
		WithKey("CFG"),
		WithTokens("BEGIN", "DONE"),
	)
	// Expect pattern: # CFG:cfg:BEGIN and # CFG:cfg:DONE
	if !strings.Contains(snippet, "# CFG:cfg:BEGIN") {
		t.Errorf("expected custom start marker, got: %s", snippet)
	}
	if !strings.Contains(snippet, "# CFG:cfg:DONE") {
		t.Errorf("expected custom end marker, got: %s", snippet)
	}
}

func TestStartEndMarkerHelpers(t *testing.T) {
	opts := []Option{WithStyle(CommentStyle{LinePrefix: "<!--", BlockEnd: "-->", Space: true}), WithKey("CANARY"), WithTokens("START", "END")}
	unnamedStart := StartMarker("", opts...)
	unnamedEnd := EndMarker("", opts...)
	if unnamedStart != "<!-- CANARY:START -->" {
		t.Errorf("unexpected unnamed start marker: %s", unnamedStart)
	}
	if unnamedEnd != "<!-- CANARY:END -->" {
		t.Errorf("unexpected unnamed end marker: %s", unnamedEnd)
	}
	keyedStart := StartMarker("intro", opts...)
	keyedEnd := EndMarker("intro", opts...)
	if keyedStart != "<!-- CANARY:intro:START -->" {
		t.Errorf("unexpected keyed start: %s", keyedStart)
	}
	if keyedEnd != "<!-- CANARY:intro:END -->" {
		t.Errorf("unexpected keyed end: %s", keyedEnd)
	}
}

// TestUpdateSectionRejectsUnkeyed proves a new unkeyed section can no longer
// be created: two tools writing unkeyed sections into one file cannot tell
// their sections apart.
func TestUpdateSectionRejectsUnkeyed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.md")
	if err := UpdateSection(path, "", "body"); err == nil {
		t.Fatal("expected an error for an empty section key")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("refused update still created the file: %v", err)
	}
}

// TestUpdateSectionAdoptsLegacyUnnamed proves a file written by an older
// version -- carrying an unkeyed section -- is updated in place rather than
// growing a second, keyed section beside it.
func TestUpdateSectionAdoptsLegacyUnnamed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.md")
	legacy := "# Doc\n\n<!-- CANARY:START -->\nold body\n<!-- CANARY:END -->\n\ntail\n"
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if err := UpdateSection(path, "guide", "new body"); err != nil {
		t.Fatalf("UpdateSection: %v", err)
	}
	got := readTestFile(t, path)
	want := "# Doc\n\n<!-- CANARY:START -->\nnew body\n<!-- CANARY:END -->\n\ntail\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

// TestScanMarkersStructuralErrors proves every malformed marker arrangement is
// reported instead of guessed at, and that the file is never written.
func TestScanMarkersStructuralErrors(t *testing.T) {
	cases := map[string]string{
		"unmatched start": "<!-- CANARY:a:START -->\nbody\n",
		"end before start": "body\n<!-- CANARY:a:END -->\n" +
			"<!-- CANARY:a:START -->\nx\n<!-- CANARY:a:END -->\n",
		"nested start":   "<!-- CANARY:a:START -->\n<!-- CANARY:b:START -->\nx\n<!-- CANARY:b:END -->\n",
		"mismatched end": "<!-- CANARY:a:START -->\nx\n<!-- CANARY:b:END -->\n",
		"duplicate key": "<!-- CANARY:a:START -->\nx\n<!-- CANARY:a:END -->\n" +
			"<!-- CANARY:a:START -->\ny\n<!-- CANARY:a:END -->\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "f.md")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			if err := UpdateSection(path, "a", "replacement"); err == nil {
				t.Fatal("expected a structural marker error")
			}
			if got := readTestFile(t, path); got != content {
				t.Fatalf("file was mutated:\n%q", got)
			}
		})
	}
}

// TestMarkerMatchIsExact proves a line that merely mentions a marker is not
// treated as one -- substring matching is what let a prose reference silently
// swallow a section.
func TestMarkerMatchIsExact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.md")
	content := "Use the <!-- CANARY:a:START --> marker to open a section.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := UpdateSection(path, "a", "body"); err != nil {
		t.Fatalf("UpdateSection: %v", err)
	}
	got := readTestFile(t, path)
	want := content + "\n<!-- CANARY:a:START -->\nbody\n<!-- CANARY:a:END -->\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
