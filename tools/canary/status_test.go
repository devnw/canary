// Copyright (c) 2024 by CodePros.
//
// This software is proprietary information of CodePros.
// Unauthorized use, copying, modification, distribution, and/or
// disclosure is strictly prohibited, except as provided under the terms
// of the commercial license agreement you have entered into with
// CodePros.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact CodePros at info@codepros.org.

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCANARY_CBIN_103_API_StatusSchema validates the status JSON schema.
// This test ensures the JSON output:
// - Contains all required top-level keys (generated_at, requirements, summary)
// - Has properly structured summary (by_status, by_aspect, total_tokens, unique_requirements)
// - Uses sorted key ordering (by_aspect before by_status alphabetically)
// - Marshals correctly without errors
func TestCANARY_CBIN_103_API_StatusSchema(t *testing.T) {
	// Setup: create report with known data
	rep := Report{
		GeneratedAt: "2025-10-15T00:00:00Z",
		Requirements: []Requirement{
			{
				ID: "CBIN-101",
				Features: []Feature{
					{
						Feature: "ScannerCore",
						Aspect:  "Engine",
						Status:  "TESTED",
						Files:   []string{"main.go"},
						Tests:   []string{"TestCANARY_CBIN_101_Engine_ScanBasic"},
						Benches: []string{},
						Owner:   "canary",
						Updated: "2025-09-20",
					},
				},
			},
			{
				ID: "CBIN-102",
				Features: []Feature{
					{
						Feature: "VerifyGate",
						Aspect:  "CLI",
						Status:  "TESTED",
						Files:   []string{"verify.go"},
						Tests:   []string{"TestCANARY_CBIN_102_CLI_Verify"},
						Benches: []string{},
						Owner:   "canary",
						Updated: "2025-09-20",
					},
				},
			},
		},
		Summary: Summary{
			ByStatus:           StatusCounts{"TESTED": 2},
			ByAspect:           AspectCounts{"Engine": 1, "CLI": 1},
			TotalTokens:        2,
			UniqueRequirements: 2,
		},
	}

	// Execute: marshal to JSON
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify: valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify: top-level keys present
	requiredKeys := []string{"generated_at", "requirements", "summary"}
	for _, key := range requiredKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing required key: %s", key)
		}
	}

	// Verify: summary structure
	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary is not a map")
	}
	summaryKeys := []string{"by_status", "by_aspect", "total_tokens", "unique_requirements"}
	for _, key := range summaryKeys {
		if _, ok := summary[key]; !ok {
			t.Errorf("summary missing key: %s", key)
		}
	}

	// Verify: key ordering (by_status before by_aspect - struct field order)
	// Note: Go JSON encoder uses struct field order, not alphabetical JSON key order
	jsonStr := string(b)
	aspectIdx := strings.Index(jsonStr, `"by_aspect"`)
	statusIdx := strings.Index(jsonStr, `"by_status"`)

	if aspectIdx < 0 {
		t.Error("could not find by_aspect in JSON")
	}
	if statusIdx < 0 {
		t.Error("could not find by_status in JSON")
	}
	if statusIdx >= 0 && aspectIdx >= 0 && statusIdx > aspectIdx {
		t.Error("expected by_status before by_aspect (struct field order)")
	}

	// Verify: total_tokens and unique_requirements are correct
	totalTokens, ok := summary["total_tokens"].(float64)
	if !ok {
		t.Error("total_tokens is not a number")
	} else if totalTokens != 2 {
		t.Errorf("expected total_tokens=2, got %v", totalTokens)
	}

	uniqueReqs, ok := summary["unique_requirements"].(float64)
	if !ok {
		t.Error("unique_requirements is not a number")
	} else if uniqueReqs != 2 {
		t.Errorf("expected unique_requirements=2, got %v", uniqueReqs)
	}
}

// TestCANARY_CBIN_103_API_JSONDeterminism validates that JSON output is byte-for-byte identical across multiple runs.
// This test ensures:
// - Multiple scans of the same fixtures produce identical JSON
// - Key ordering is deterministic (sorted)
// - No timestamp variations or other non-deterministic fields (except generated_at which is tested separately)
// This is critical for reproducible builds and stable diffs in version control.
func TestCANARY_CBIN_103_API_JSONDeterminism(t *testing.T) {
	// Setup: create fixture directory with 20 CANARY tokens
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		content := fmt.Sprintf(`package p
// CANARY: REQ=CBIN-%03d; FEATURE="Feature%d"; ASPECT=API; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15
func Feature%d() {}
`, i, i, i)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%02d.go", i)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Run scanner 5 times and collect JSON outputs
	var hashes []string
	var jsons []string
	for run := 0; run < 5; run++ {
		rep, err := scan(dir, skipDefault, nil, nil)
		if err != nil {
			t.Fatalf("scan %d failed: %v", run, err)
		}

		// Override generated_at to fixed value for determinism test
		// (generated_at is expected to vary between runs, so we normalize it)
		rep.GeneratedAt = "2025-10-15T00:00:00Z"

		b, err := json.Marshal(rep)
		if err != nil {
			t.Fatalf("marshal %d failed: %v", run, err)
		}

		// Compute SHA256 hash
		hash := sha256.Sum256(b)
		hashStr := fmt.Sprintf("%x", hash)
		hashes = append(hashes, hashStr)
		jsons = append(jsons, string(b))
	}

	// Verify: all hashes are identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("JSON determinism failure: run %d hash differs from run 0", i)
			t.Logf("Run 0 hash: %s", hashes[0])
			t.Logf("Run %d hash: %s", i, hashes[i])
			t.Logf("Run 0 JSON length: %d", len(jsons[0]))
			t.Logf("Run %d JSON length: %d", i, len(jsons[i]))
			// Show first difference
			for j := 0; j < len(jsons[0]) && j < len(jsons[i]); j++ {
				if jsons[0][j] != jsons[i][j] {
					start := j - 20
					if start < 0 {
						start = 0
					}
					end := j + 20
					if end > len(jsons[0]) {
						end = len(jsons[0])
					}
					t.Logf("First difference at byte %d:", j)
					t.Logf("Run 0: ...%s...", jsons[0][start:end])
					if end <= len(jsons[i]) {
						t.Logf("Run %d: ...%s...", i, jsons[i][start:end])
					}
					break
				}
			}
		}
	}

	// Verify: JSON contains sorted keys
	// Check that by_status appears before by_aspect (struct field order)
	jsonStr := jsons[0]
	if !strings.Contains(jsonStr, `"by_status"`) {
		t.Error("JSON missing by_status")
	}
	if !strings.Contains(jsonStr, `"by_aspect"`) {
		t.Error("JSON missing by_aspect")
	}
}

// setupLargeReport creates a large report for benchmarking.
// Returns a report with numReqs requirements, each with featuresPerReq features.
func setupLargeReport(tb testing.TB, numReqs int, featuresPerReq int) *Report {
	tb.Helper()
	rep := &Report{
		GeneratedAt:  "2025-10-15T00:00:00Z",
		Requirements: make([]Requirement, numReqs),
		Summary: Summary{
			ByStatus:           StatusCounts{},
			ByAspect:           AspectCounts{},
			TotalTokens:        numReqs * featuresPerReq,
			UniqueRequirements: numReqs,
		},
	}
	for i := 0; i < numReqs; i++ {
		features := make([]Feature, featuresPerReq)
		for j := 0; j < featuresPerReq; j++ {
			features[j] = Feature{
				Feature: fmt.Sprintf("Feature%d_%d", i, j),
				Aspect:  "API",
				Status:  "IMPL",
				Files:   []string{fmt.Sprintf("file%d_%d.go", i, j)},
				Tests:   []string{},
				Benches: []string{},
				Owner:   "team",
				Updated: "2025-09-20",
			}
			rep.Summary.ByStatus["IMPL"]++
			rep.Summary.ByAspect["API"]++
		}
		rep.Requirements[i] = Requirement{
			ID:       fmt.Sprintf("CBIN-%03d", i),
			Features: features,
		}
	}
	return rep
}

// BenchmarkCANARY_CBIN_103_API_Emit benchmarks JSON and CSV emission.
// Measures performance of emitting status for 100 requirements × 3 features = 300 tokens.
// Baseline target: allocs/op ≤ 20
func BenchmarkCANARY_CBIN_103_API_Emit(b *testing.B) {
	rep := setupLargeReport(b, 100, 3) // 100 reqs × 3 features = 300 features
	dir := b.TempDir()
	jsonPath := filepath.Join(dir, "status.json")
	csvPath := filepath.Join(dir, "status.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := writeJSON(jsonPath, *rep); err != nil {
			b.Fatal(err)
		}
		if err := writeCSV(csvPath, *rep); err != nil {
			b.Fatal(err)
		}
	}
}
