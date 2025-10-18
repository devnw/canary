// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="PlanGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_PlanTemplate; UPDATED=2025-10-17

package migrate

import (
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// TestCANARY_CBIN_145_Engine_PlanTemplate verifies plan generation uses correct template
func TestCANARY_CBIN_145_Engine_PlanTemplate(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1200",
		Features: []*storage.Token{
			{ReqID: "CBIN-1200", Feature: "TestFeature", Aspect: "API", Status: "IMPL", FilePath: "test.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify required sections
	requiredSections := []string{
		"# Implementation Plan",
		"## Overview",
		"## Current Implementation Status",
		"## Architecture",
		"CBIN-1200",
	}

	for _, section := range requiredSections {
		if !contains(plan, section) {
			t.Errorf("plan missing required section: %s", section)
		}
	}
}

// TestCANARY_CBIN_145_Engine_PlanStatusProgression verifies status reflected in plan
func TestCANARY_CBIN_145_Engine_PlanStatusProgression(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1300",
		Features: []*storage.Token{
			{ReqID: "CBIN-1300", Feature: "Feature1", Aspect: "API", Status: "STUB", FilePath: "test.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-1300", Feature: "Feature2", Aspect: "Engine", Status: "IMPL", FilePath: "test.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-1300", Feature: "Feature3", Aspect: "Storage", Status: "TESTED", Test: "TestCANARY_CBIN_1300_Storage_Feature3", FilePath: "test.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-1300", Feature: "Feature4", Aspect: "Bench", Status: "BENCHED", Bench: "BenchmarkCANARY_CBIN_1300_Bench_Feature4", FilePath: "test.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 4,
		Confidence:   ConfidenceHigh,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify all statuses are reflected
	statuses := []string{"STUB", "IMPL", "TESTED", "BENCHED"}
	for _, status := range statuses {
		if !contains(plan, status) {
			t.Errorf("plan should reflect STATUS=%s", status)
		}
	}
}

// TestCANARY_CBIN_145_Engine_PlanPhaseMapping verifies status to phase mapping
func TestCANARY_CBIN_145_Engine_PlanPhaseMapping(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1400",
		Features: []*storage.Token{
			{ReqID: "CBIN-1400", Feature: "APIHandler", Aspect: "API", Status: "IMPL", FilePath: "api.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-1400", Feature: "APITests", Aspect: "API", Status: "TESTED", Test: "TestCANARY_CBIN_1400_API_Handler", FilePath: "api_test.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 2,
		Confidence:   ConfidenceHigh,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Plan should indicate implementation phases
	expectedPhases := []string{"Phase", "Implementation", "Testing"}
	for _, phase := range expectedPhases {
		if !contains(plan, phase) {
			t.Errorf("plan should contain phase information: %s", phase)
		}
	}
}

// TestCANARY_CBIN_145_Engine_PlanTestReferences verifies test names in plan
func TestCANARY_CBIN_145_Engine_PlanTestReferences(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1500",
		Features: []*storage.Token{
			{
				ReqID:     "CBIN-1500",
				Feature:   "Cache",
				Aspect:    "Engine",
				Status:    "TESTED",
				Test:      "TestCANARY_CBIN_1500_Engine_Cache",
				FilePath:  "cache_test.go",
				UpdatedAt: "2025-10-17",
			},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify test reference
	if !contains(plan, "TestCANARY_CBIN_1500_Engine_Cache") {
		t.Error("plan should reference existing test names")
	}
}

// TestCANARY_CBIN_145_Engine_PlanFileReferences verifies file paths in plan
func TestCANARY_CBIN_145_Engine_PlanFileReferences(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1600",
		Features: []*storage.Token{
			{ReqID: "CBIN-1600", Feature: "DBLayer", Aspect: "Storage", Status: "IMPL", FilePath: "pkg/database/db.go", LineNumber: 25, UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify file path is referenced
	if !contains(plan, "pkg/database/db.go") {
		t.Error("plan should reference implementation file paths")
	}
}

// TestCANARY_CBIN_145_Engine_PlanArchitectureSection verifies architecture details
func TestCANARY_CBIN_145_Engine_PlanArchitectureSection(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1700",
		Features: []*storage.Token{
			{ReqID: "CBIN-1700", Feature: "API", Aspect: "API", Status: "IMPL", FilePath: "api.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-1700", Feature: "Engine", Aspect: "Engine", Status: "IMPL", FilePath: "engine.go", UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-1700", Feature: "Storage", Aspect: "Storage", Status: "IMPL", FilePath: "storage.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 3,
		Confidence:   ConfidenceHigh,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify architecture section mentions all aspects
	aspects := []string{"API", "Engine", "Storage"}
	for _, aspect := range aspects {
		if !contains(plan, aspect) {
			t.Errorf("plan architecture should mention %s aspect", aspect)
		}
	}
}

// TestCANARY_CBIN_145_Engine_PlanConfidenceNote verifies confidence level noted
func TestCANARY_CBIN_145_Engine_PlanConfidenceNote(t *testing.T) {
	testCases := []struct {
		name       string
		confidence string
		shouldWarn bool
	}{
		{"High confidence - no warning", ConfidenceHigh, false},
		{"Medium confidence - mild note", ConfidenceMedium, true},
		{"Low confidence - strong warning", ConfidenceLow, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orphan := &OrphanedRequirement{
				ReqID:        "CBIN-TEST",
				Features:     []*storage.Token{{ReqID: "CBIN-TEST", Feature: "Test", Aspect: "API", Status: "IMPL", FilePath: "test.go", UpdatedAt: "2025-10-17"}},
				FeatureCount: 1,
				Confidence:   tc.confidence,
			}

			plan, err := GeneratePlan(orphan)
			if err != nil {
				t.Fatalf("GeneratePlan failed: %v", err)
			}

			hasWarning := contains(plan, "CONFIDENCE") || contains(plan, "review") || contains(plan, "manually")

			if tc.shouldWarn && !hasWarning {
				t.Error("plan should contain confidence warning")
			}
		})
	}
}

// TestCANARY_CBIN_145_Engine_PlanNextSteps verifies next steps section
func TestCANARY_CBIN_145_Engine_PlanNextSteps(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID: "CBIN-1800",
		Features: []*storage.Token{
			{ReqID: "CBIN-1800", Feature: "Logger", Aspect: "Engine", Status: "STUB", FilePath: "log.go", UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 1,
		Confidence:   ConfidenceMedium,
	}

	plan, err := GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify next steps are provided
	if !contains(plan, "Next") || !contains(plan, "Steps") {
		t.Error("plan should contain next steps section")
	}
}

// TestCANARY_CBIN_145_Engine_PlanEmptyFeatures verifies handling of edge cases
func TestCANARY_CBIN_145_Engine_PlanEmptyFeatures(t *testing.T) {
	orphan := &OrphanedRequirement{
		ReqID:        "CBIN-1900",
		Features:     []*storage.Token{},
		FeatureCount: 0,
		Confidence:   ConfidenceLow,
	}

	_, err := GeneratePlan(orphan)
	if err == nil {
		t.Error("GeneratePlan should return error for orphan with no features")
	}
}
