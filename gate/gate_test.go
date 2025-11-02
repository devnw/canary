package gate

import (
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
