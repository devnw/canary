package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrompt_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "p.md")
	if err := os.WriteFile(f, []byte("hello from file"), 0644); err != nil {
		t.Fatal(err)
	}
	content, err := LoadPrompt(f)
	if err != nil {
		t.Fatal(err)
	}
	if content != "hello from file" {
		t.Errorf("got %q", content)
	}
}

func TestLoadPrompt_EmbeddedCommand(t *testing.T) {
	content, err := LoadPrompt("scan")
	if err != nil {
		t.Fatal(err)
	}
	if content == "" {
		t.Error("expected non-empty scan prompt")
	}
	if len(content) < 50 {
		t.Errorf("scan prompt too short: %d", len(content))
	}
}

func TestLoadPrompt_EmbeddedSystem(t *testing.T) {
	content, err := LoadPrompt("init")
	if err != nil {
		t.Fatal(err)
	}
	if content == "" {
		t.Error("expected non-empty init prompt")
	}
}

func TestLoadPrompt_NotFound(t *testing.T) {
	_, err := LoadPrompt("nonexistent-embed-name")
	if err == nil {
		t.Error("expected error for unknown embedded prompt")
	}
}

func TestLoadPrompt_Empty(t *testing.T) {
	_, err := LoadPrompt("")
	if err == nil {
		t.Error("expected error for empty prompt")
	}
}

func TestGetAvailablePrompts(t *testing.T) {
	names, err := GetAvailablePrompts()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Error("expected at least one embedded prompt name")
	}
	hasScan := false
	hasInit := false
	for _, n := range names {
		if n == "scan" {
			hasScan = true
		}
		if n == "init" {
			hasInit = true
		}
	}
	if !hasScan {
		t.Error("expected 'scan' in available prompts")
	}
	if !hasInit {
		t.Error("expected 'init' in available prompts")
	}
}

func TestValidatePromptArg_Empty(t *testing.T) {
	if err := ValidatePromptArg(""); err != nil {
		t.Errorf("empty should be valid: %v", err)
	}
}

func TestValidatePromptArg_FileNotFound(t *testing.T) {
	err := ValidatePromptArg("/nonexistent/path/to/file.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
