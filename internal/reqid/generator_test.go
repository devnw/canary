// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-139; FEATURE="GeneratorTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectScopedIDGen; UPDATED=2025-10-16
package reqid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCBIN139_AspectScopedIDGen(t *testing.T) {
	// Create temporary .canary/specs directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to tmpDir: %v", err)
	}

	specsDir := filepath.Join(".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("Failed to create specs dir: %v", err)
	}

	// Create some existing specs
	os.MkdirAll(filepath.Join(specsDir, "CBIN-CLI-001-feature1"), 0755)
	os.MkdirAll(filepath.Join(specsDir, "CBIN-CLI-003-feature2"), 0755) // Gap at 002
	os.MkdirAll(filepath.Join(specsDir, "CBIN-API-001-feature3"), 0755)

	tests := []struct {
		aspect string
		want   string
	}{
		{"CLI", "CBIN-CLI-004"},       // Next after 003 (ignores gap)
		{"API", "CBIN-API-002"},       // Next after 001
		{"Engine", "CBIN-Engine-001"}, // First for this aspect
	}

	for _, tt := range tests {
		t.Run(tt.aspect, func(t *testing.T) {
			got, err := GenerateNextID("CBIN", tt.aspect)
			if err != nil {
				t.Fatalf("GenerateNextID() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GenerateNextID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCBIN139_AspectScopedIDGen_NoExistingSpecs(t *testing.T) {
	// Create temporary .canary/specs directory with no existing specs
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to tmpDir: %v", err)
	}

	specsDir := filepath.Join(".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("Failed to create specs dir: %v", err)
	}

	// Generate first ID for an aspect
	got, err := GenerateNextID("CBIN", "CLI")
	if err != nil {
		t.Fatalf("GenerateNextID() error = %v", err)
	}

	want := "CBIN-CLI-001"
	if got != want {
		t.Errorf("GenerateNextID() = %q, want %q", got, want)
	}
}

func TestCBIN139_AspectScopedIDGen_InvalidAspect(t *testing.T) {
	// Try to generate ID with invalid aspect
	_, err := GenerateNextID("CBIN", "InvalidAspect")
	if err == nil {
		t.Error("GenerateNextID() with invalid aspect should return error")
	}
}

func TestCBIN139_AspectScopedIDGen_CaseInsensitive(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir to tmpDir: %v", err)
	}

	specsDir := filepath.Join(".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("Failed to create specs dir: %v", err)
	}

	// Generate ID with lowercase aspect (should normalize to proper casing)
	got, err := GenerateNextID("CBIN", "cli")
	if err != nil {
		t.Fatalf("GenerateNextID() error = %v", err)
	}

	// Should normalize to CLI
	want := "CBIN-CLI-001"
	if got != want {
		t.Errorf("GenerateNextID() = %q, want %q", got, want)
	}
}
