package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSlashCommandsCodexUsesPromptsDir(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if err := copyCanaryStructure(projectDir); err != nil {
		t.Fatalf("copyCanaryStructure failed: %v", err)
	}

	notes, err := installSlashCommands(projectDir, []string{"codex"}, false, false)
	if err != nil {
		t.Fatalf("installSlashCommands failed: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected no notes for global Codex install, got %v", notes)
	}

	promptPath := filepath.Join(homeDir, ".codex", "prompts", "canary.scan.md")
	if _, err := os.Stat(promptPath); err != nil {
		t.Fatalf("expected Codex prompt at %s: %v", promptPath, err)
	}

	legacyDir := filepath.Join(homeDir, ".codex", "commands")
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Fatalf("expected legacy Codex commands dir to remain unused, got err=%v", err)
	}
}

func TestInstallSlashCommandsCodexLocalModeFallsBackToGlobalPrompts(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if err := copyCanaryStructure(projectDir); err != nil {
		t.Fatalf("copyCanaryStructure failed: %v", err)
	}

	notes, err := installSlashCommands(projectDir, []string{"codex"}, false, true)
	if err != nil {
		t.Fatalf("installSlashCommands failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected one Codex note, got %v", notes)
	}
	if !strings.Contains(notes[0], filepath.Join(homeDir, ".codex", "prompts")) {
		t.Fatalf("expected note to mention global prompts dir, got %q", notes[0])
	}

	promptPath := filepath.Join(homeDir, ".codex", "prompts", "canary.scan.md")
	if _, err := os.Stat(promptPath); err != nil {
		t.Fatalf("expected Codex prompt at %s: %v", promptPath, err)
	}

	localPromptPath := filepath.Join(projectDir, ".codex", "prompts", "canary.scan.md")
	if _, err := os.Stat(localPromptPath); !os.IsNotExist(err) {
		t.Fatalf("expected no project-local Codex prompt at %s, got err=%v", localPromptPath, err)
	}
}

func TestUpdateAgentContextFilesCreatesAGENTS(t *testing.T) {
	projectDir := t.TempDir()

	if err := copyCanaryStructure(projectDir); err != nil {
		t.Fatalf("copyCanaryStructure failed: %v", err)
	}

	if err := updateAgentContextFiles(projectDir); err != nil {
		t.Fatalf("updateAgentContextFiles failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md failed: %v", err)
	}

	body := string(content)
	if !strings.Contains(body, "Repository Guidelines") {
		t.Fatalf("expected Codex AGENTS title, got %q", body)
	}
	if !strings.Contains(body, "/canary.scan") {
		t.Fatalf("expected Codex slash command guidance, got %q", body)
	}
}
