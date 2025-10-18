// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_SpecTemplate; UPDATED=2025-10-17

package migrate

import (
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// TestCANARY_CBIN_145_Engine_SpecTemplate verifies spec generation uses correct template
func TestCANARY_CBIN_145_Engine_SpecTemplate(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-600",
		Features: []*storage.Token{
			{ReqID: "CBIN-600", Feature: "TestFeature", Aspect: "API", Status: "IMPL", FilePath: "test.go", LineNumber: 10, UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	spec, err := GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify required sections
	requiredSections := []string{
		"# Requirement Specification",
		"## Overview",
		"## User Stories",
		"## Functional Requirements",
		"## Implementation Checklist",
		"CBIN-600",
	}

	for _, section := range requiredSections {
		if !contains(spec, section) {
			t.Errorf("spec missing required section: %s", section)
		}
	}
}

// TestCANARY_CBIN_145_Engine_SpecTokens verifies CANARY tokens in spec
func TestCANARY_CBIN_145_Engine_SpecTokens(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-700",
		Features: []*storage.Token{
			{ReqID: "CBIN-700", Feature: "DatabaseAPI", Aspect: "API", Status: "IMPL", FilePath: "pkg/db/api.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-700", Feature: "DatabaseTests", Aspect: "Storage", Status: "TESTED", Test: "TestCANARY_CBIN_700_Storage_DB", FilePath: "pkg/db/db_test.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 2,
		Confidence:   ConfidenceHigh,
	}

	spec, err := GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify tokens exist for both features
	if !contains(spec, "REQ=CBIN-700") {
		t.Error("spec should contain REQ=CBIN-700")
	}

	if !contains(spec, "FEATURE=\"DatabaseAPI\"") {
		t.Error("spec should contain DatabaseAPI feature")
	}

	if !contains(spec, "FEATURE=\"DatabaseTests\"") {
		t.Error("spec should contain DatabaseTests feature")
	}

	if !contains(spec, "ASPECT=API") {
		t.Error("spec should contain ASPECT=API")
	}

	if !contains(spec, "ASPECT=Storage") {
		t.Error("spec should contain ASPECT=Storage")
	}

	if !contains(spec, "STATUS=IMPL") {
		t.Error("spec should contain STATUS=IMPL")
	}

	if !contains(spec, "STATUS=TESTED") {
		t.Error("spec should contain STATUS=TESTED")
	}
}

// TestCANARY_CBIN_145_Engine_SpecConfidenceWarning verifies low confidence warnings
func TestCANARY_CBIN_145_Engine_SpecConfidenceWarning(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-800",
		Features: []*storage.Token{
			{ReqID: "CBIN-800", Feature: "MinimalFeature", Aspect: "API", Status: "STUB", FilePath: "test.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceLow,
	}

	spec, err := GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify warning is present
	if !contains(spec, "CONFIDENCE") && !contains(spec, "LOW") {
		t.Error("spec should contain confidence warning for low confidence orphans")
	}

	// Verify note about manual review
	if !contains(spec, "review") || !contains(spec, "manually") {
		t.Error("spec should recommend manual review for low confidence migrations")
	}
}

// TestCANARY_CBIN_145_Engine_SpecFeatureNaming verifies feature name extraction
func TestCANARY_CBIN_145_Engine_SpecFeatureNaming(t *testing.T) {
	testCases := []struct {
		name           string
		featureName    string
		expectedInSpec string
	}{
		{"Simple name", "UserAuth", "UserAuth"},
		{"Camel case", "PaymentProcessor", "PaymentProcessor"},
		{"With numbers", "OAuth2", "OAuth2"},
		{"Underscore", "User_Authentication", "User_Authentication"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orphan := &OrphanedRequirement{
				ReqID: "CBIN-TEST",
				Features: []*storage.Token{
					{ReqID: "CBIN-TEST", Feature: tc.featureName, Aspect: "API", Status: "IMPL", FilePath: "test.go", UpdatedAt: "2025-10-17"},
				},
				FeatureCount: 1,
				Confidence:   ConfidenceMedium,
			}

			spec, err := GenerateSpec(orphan)
			if err != nil {
				t.Fatalf("GenerateSpec failed: %v", err)
			}

			if !contains(spec, tc.expectedInSpec) {
				t.Errorf("spec should contain feature name: %s", tc.expectedInSpec)
			}
		})
	}
}

// TestCANARY_CBIN_145_Engine_SpecAspectGrouping verifies features grouped by aspect
func TestCANARY_CBIN_145_Engine_SpecAspectGrouping(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-900",
		Features: []*storage.Token{
			{ReqID: "CBIN-900", Feature: "APIHandler", Aspect: "API", Status: "IMPL", FilePath: "api.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-900", Feature: "CLICommand", Aspect: "CLI", Status: "IMPL", FilePath: "cli.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-900", Feature: "EngineCore", Aspect: "Engine", Status: "IMPL", FilePath: "engine.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 3,
		Confidence:   ConfidenceHigh,
	}

	spec, err := GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify all aspects are present
	aspects := []string{"API", "CLI", "Engine"}
	for _, aspect := range aspects {
		if !contains(spec, "ASPECT="+aspect) {
			t.Errorf("spec should contain ASPECT=%s", aspect)
		}
	}
}

// TestCANARY_CBIN_145_Engine_SpecTestReferences verifies test names included
func TestCANARY_CBIN_145_Engine_SpecTestReferences(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1000",
		Features: []*storage.Token{
			{
				ReqID:     "CBIN-1000",
				Feature:   "Cache",
				Aspect:    "Engine",
				Status:    "TESTED",
				Test:      "TestCANARY_CBIN_1000_Engine_Cache",
				FilePath:  "cache_test.go",
				UpdatedAt: "2025-10-17",
			},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	spec, err := GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify test name is included
	if !contains(spec, "TestCANARY_CBIN_1000_Engine_Cache") {
		t.Error("spec should include test name from token")
	}

	if !contains(spec, "TEST=") {
		t.Error("spec should include TEST= field for tested features")
	}
}

// TestCANARY_CBIN_145_Engine_SpecBenchmarkReferences verifies benchmark names included
func TestCANARY_CBIN_145_Engine_SpecBenchmarkReferences(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1100",
		Features: []*storage.Token{
			{
				ReqID:     "CBIN-1100",
				Feature:   "SortAlgorithm",
				Aspect:    "Bench",
				Status:    "BENCHED",
				Bench:     "BenchmarkCANARY_CBIN_1100_Bench_QuickSort",
				FilePath:  "sort_bench_test.go",
				UpdatedAt: "2025-10-17",
			},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	spec, err := GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify benchmark name is included
	if !contains(spec, "BenchmarkCANARY_CBIN_1100_Bench_QuickSort") {
		t.Error("spec should include benchmark name from token")
	}

	if !contains(spec, "BENCH=") {
		t.Error("spec should include BENCH= field for benchmarked features")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
