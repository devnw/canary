// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-136; FEATURE="DocStalenessDetection"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_StalenessDetection; UPDATED=2025-10-16

package docs_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/docs"
	"go.devnw.com/canary/internal/storage"
)

// TestCANARY_CBIN_136_Engine_StalenessDetection verifies documentation staleness checking
func TestCANARY_CBIN_136_Engine_StalenessDetection(t *testing.T) {
	tests := []struct {
		name         string
		docContent   string
		tokenDocHash string
		wantStatus   string
	}{
		{
			name:         "current documentation",
			docContent:   "# Feature\n\nUp to date.",
			tokenDocHash: "computed-from-content", // Hash will match
			wantStatus:   "DOC_CURRENT",
		},
		{
			name:         "stale documentation",
			docContent:   "# Feature\n\nModified content.",
			tokenDocHash: "old-hash-value",
			wantStatus:   "DOC_STALE",
		},
		{
			name:         "unhashed documentation",
			docContent:   "# Feature\n\nNo hash in token.",
			tokenDocHash: "", // Empty DOC_HASH field
			wantStatus:   "DOC_UNHASHED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create test documentation file
			tmpDir := t.TempDir()
			docFile := filepath.Join(tmpDir, "feature.md")
			if err := os.WriteFile(docFile, []byte(tt.docContent), 0644); err != nil {
				t.Fatalf("failed to write doc file: %v", err)
			}

			// Setup: Create token with DOC_HASH
			var expectedHash string
			if tt.tokenDocHash == "computed-from-content" {
				// Calculate actual hash for "current" case
				hash, _ := docs.CalculateHash(docFile)
				expectedHash = hash
			} else {
				expectedHash = tt.tokenDocHash
			}

			token := &storage.Token{
				ReqID:   "CBIN-TEST",
				Feature: "TestFeature",
				DocPath: docFile,
				DocHash: expectedHash,
			}

			// Execute: Check staleness
			status, err := docs.CheckStaleness(token)
			if err != nil {
				t.Fatalf("CheckStaleness failed: %v", err)
			}

			// Verify: Status matches expectation
			if status != tt.wantStatus {
				t.Errorf("got status %s, want %s", status, tt.wantStatus)
			}
		})
	}
}

// TestCANARY_CBIN_136_Engine_MissingDocumentation verifies missing file detection
func TestCANARY_CBIN_136_Engine_MissingDocumentation(t *testing.T) {
	// Setup: Token references non-existent file
	token := &storage.Token{
		ReqID:   "CBIN-TEST",
		Feature: "TestFeature",
		DocPath: "/nonexistent/path/to/doc.md",
		DocHash: "abc123",
	}

	// Execute: Check staleness
	status, err := docs.CheckStaleness(token)

	// Verify: Status is DOC_MISSING (no error, graceful handling)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if status != "DOC_MISSING" {
		t.Errorf("got status %s, want DOC_MISSING", status)
	}
}
