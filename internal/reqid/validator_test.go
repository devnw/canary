// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-139; FEATURE="ValidatorTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectValidation; UPDATED=2025-10-16
package reqid

import (
	"strings"
	"testing"
)

func TestCBIN139_AspectValidation(t *testing.T) {
	tests := []struct {
		aspect  string
		wantErr bool
	}{
		// Valid aspects
		{"API", false},
		{"CLI", false},
		{"Engine", false},
		{"Storage", false},
		{"Security", false},
		{"Docs", false},
		{"Wire", false},
		{"Planner", false},
		{"Decode", false},
		{"Encode", false},
		{"RoundTrip", false},
		{"Bench", false},
		{"FrontEnd", false},
		{"Dist", false},
		// Case-insensitive variations
		{"api", false},
		{"cli", false},
		{"STORAGE", false},
		// Invalid aspects
		{"Frontend", true}, // Typo (should be FrontEnd)
		{"Invalid", true},
		{"XYZ", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.aspect, func(t *testing.T) {
			err := ValidateAspect(tt.aspect)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAspect(%q) error = %v, wantErr %v", tt.aspect, err, tt.wantErr)
			}
		})
	}
}

func TestCBIN139_AspectSuggestion(t *testing.T) {
	tests := []struct {
		typo        string
		wantContain string // Suggestion should contain this
	}{
		{"Frontend", "FrontEnd"},
		{"Engin", "Engine"},
		{"Storge", "Storage"},
		{"Ap", "API"}, // Close match
	}

	for _, tt := range tests {
		t.Run(tt.typo, func(t *testing.T) {
			got := SuggestAspect(tt.typo)
			if got == "" {
				t.Errorf("SuggestAspect(%q) returned empty, want suggestion containing %q", tt.typo, tt.wantContain)
			}
			if !strings.Contains(got, tt.wantContain) {
				t.Errorf("SuggestAspect(%q) = %q, want to contain %q", tt.typo, got, tt.wantContain)
			}
		})
	}
}

func TestCBIN139_NormalizeAspect(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"api", "API"},
		{"CLI", "CLI"},
		{"engine", "Engine"},
		{"STORAGE", "Storage"},
		{"frontend", "FrontEnd"}, // Normalize casing
		{"Invalid", "Invalid"},   // Not found, return as-is
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeAspect(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeAspect(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
