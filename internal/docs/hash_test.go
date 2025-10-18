// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-136; FEATURE="DocHashCalculation"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_HashCalculation; UPDATED=2025-10-16

package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.devnw.com/canary/internal/docs"
)

// TestCANARY_CBIN_136_Engine_HashCalculation verifies deterministic SHA256 hash calculation
func TestCANARY_CBIN_136_Engine_HashCalculation(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantHash string // First 16 chars of SHA256
	}{
		{
			name:     "simple markdown",
			content:  "# Hello World\n\nThis is a test.",
			wantHash: "c52f45333a141720", // Expected SHA256 (abbreviated)
		},
		{
			name:     "CRLF normalized to LF",
			content:  "# Hello World\r\n\r\nThis is a test.",
			wantHash: "c52f45333a141720", // Same as LF version
		},
		{
			name:     "empty file",
			content:  "",
			wantHash: "e3b0c44298fc1c14", // SHA256 of empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Write test file
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Execute: Calculate hash
			hash, err := docs.CalculateHash(testFile)
			if err != nil {
				t.Fatalf("CalculateHash failed: %v", err)
			}

			// Verify: Hash matches expected (first 16 chars)
			if hash[:16] != tt.wantHash {
				t.Errorf("got hash %s, want %s", hash[:16], tt.wantHash)
			}
		})
	}
}

// TestCANARY_CBIN_136_Engine_HashDeterminism verifies hash stability across multiple calculations
func TestCANARY_CBIN_136_Engine_HashDeterminism(t *testing.T) {
	// Setup: Create test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "stable.md")
	content := "# Feature Documentation\n\nThis content should hash consistently."
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Execute: Calculate hash 10 times
	var hashes []string
	for i := 0; i < 10; i++ {
		hash, err := docs.CalculateHash(testFile)
		if err != nil {
			t.Fatalf("CalculateHash iteration %d failed: %v", i, err)
		}
		hashes = append(hashes, hash)
	}

	// Verify: All hashes identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("hash %d (%s) differs from hash 0 (%s)", i, hashes[i], hashes[0])
		}
	}
}

// CANARY: REQ=CBIN-136; FEATURE="DocHashCalculation"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkCANARY_CBIN_136_Engine_HashPerformance; UPDATED=2025-10-16
// BenchmarkCANARY_CBIN_136_Engine_HashPerformance measures hash calculation performance
// Target: <10ms per 1KB documentation file (from spec FR-2)
func BenchmarkCANARY_CBIN_136_Engine_HashPerformance(b *testing.B) {
	// Setup: Create test file with typical documentation size (5KB)
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.md")
	content := strings.Repeat("# Documentation\n\nThis is test content.\n", 100) // ~5KB
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
	}

	// Reset timer
	b.ResetTimer()

	// Benchmark hash calculation
	for i := 0; i < b.N; i++ {
		_, err := docs.CalculateHash(testFile)
		if err != nil {
			b.Fatalf("CalculateHash failed: %v", err)
		}
	}

	// Report performance
	b.ReportMetric(float64(len(content))/1024, "KB/op")
}
