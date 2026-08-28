// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package prompts

import (
	"strings"
	"testing"
)

func TestGetCommand(t *testing.T) {
	tests := []struct {
		command string
		wantErr bool
	}{
		{"scan", false},
		{"list", false},
		{"show", false},
		{"create", false},
		{"status", false},
		{"nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			content, err := GetCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommand(%q) error = %v, wantErr %v", tt.command, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(content) == 0 {
				t.Errorf("GetCommand(%q) returned empty content", tt.command)
			}
			if !tt.wantErr && !strings.Contains(content, "Purpose") {
				t.Errorf("GetCommand(%q) content missing 'Purpose' section", tt.command)
			}
		})
	}
}

func TestListCommands(t *testing.T) {
	commands, err := ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() error = %v", err)
	}

	if len(commands) == 0 {
		t.Error("ListCommands() returned no commands")
	}

	// Check for some expected commands
	expectedCommands := []string{"scan", "list", "show", "create"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListCommands() missing expected command: %s", expected)
		}
	}
}

func TestGetAllCommands(t *testing.T) {
	allCommands, err := GetAllCommands()
	if err != nil {
		t.Fatalf("GetAllCommands() error = %v", err)
	}

	if len(allCommands) == 0 {
		t.Error("GetAllCommands() returned no commands")
	}

	// Verify all returned commands have content
	for cmd, content := range allCommands {
		if len(content) == 0 {
			t.Errorf("GetAllCommands() command %q has empty content", cmd)
		}
		if !strings.Contains(content, "Purpose") {
			t.Errorf("GetAllCommands() command %q missing 'Purpose' section", cmd)
		}
	}
}

func TestParseCommandPrompt(t *testing.T) {
	prompt, err := ParseCommandPrompt("scan")
	if err != nil {
		t.Fatalf("ParseCommandPrompt(\"scan\") error = %v", err)
	}

	if prompt.Command != "scan" {
		t.Errorf("ParseCommandPrompt().Command = %q, want %q", prompt.Command, "scan")
	}

	if len(prompt.FullContent) == 0 {
		t.Error("ParseCommandPrompt().FullContent is empty")
	}
}

func TestLegacySystemPrompts(t *testing.T) {
	// Test that legacy system prompts still work
	all := All()

	expectedPrompts := []string{"init", "policy", "requirements", "evaluate"}
	for _, name := range expectedPrompts {
		if content, ok := all[name]; !ok || len(content) == 0 {
			t.Errorf("All() missing or empty system prompt: %s", name)
		}
	}

	// Test individual prompt variables
	if len(Init) == 0 {
		t.Error("Init prompt is empty")
	}
	if len(Policy) == 0 {
		t.Error("Policy prompt is empty")
	}
	if len(Requirements) == 0 {
		t.Error("Requirements prompt is empty")
	}
	if len(Evaluate) == 0 {
		t.Error("Evaluate prompt is empty")
	}
}
