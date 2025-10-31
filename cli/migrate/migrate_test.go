// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="MigrationUnitTests"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17

package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/migrate"
	"go.devnw.com/canary/internal/storage"
)

// TestCANARY_CBIN_145_CLI_OrphanDetection verifies detection of orphaned requirements
func TestCANARY_CBIN_145_CLI_OrphanDetection(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and migrate database
	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create specs directory
	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	// Create spec for CBIN-100 (this one has a spec, so NOT orphaned)
	specDir100 := filepath.Join(specsDir, "CBIN-100-existing")
	if err := os.MkdirAll(specDir100, 0755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}
	specFile100 := filepath.Join(specDir100, "spec.md")
	if err := os.WriteFile(specFile100, []byte("# Spec for CBIN-100"), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Insert tokens
	tokens := []*storage.Token{
		// CBIN-100: Has spec (NOT orphaned)
		{ReqID: "CBIN-100", Feature: "ExistingFeature", Aspect: "API", Status: "IMPL", FilePath: "test.go", LineNumber: 10, UpdatedAt: "2025-10-17"},
		{ReqID: "CBIN-100", Feature: "AnotherFeature", Aspect: "CLI", Status: "TESTED", FilePath: "test.go", LineNumber: 20, UpdatedAt: "2025-10-17"},

		// CBIN-999: No spec (ORPHANED)
		{ReqID: "CBIN-999", Feature: "OrphanedFeature", Aspect: "Engine", Status: "IMPL", FilePath: "orphan.go", LineNumber: 5, UpdatedAt: "2025-10-17"},
		{ReqID: "CBIN-999", Feature: "AnotherOrphan", Aspect: "Storage", Status: "STUB", FilePath: "orphan.go", LineNumber: 15, UpdatedAt: "2025-10-17"},

		// CBIN-105: No spec, but in docs/ (excluded by path filter)
		{ReqID: "CBIN-105", Feature: "SearchExample", Aspect: "API", Status: "STUB", FilePath: "docs/user/getting-started.md", LineNumber: 50, UpdatedAt: "2025-10-17"},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	// Execute: Detect orphans with path filtering
	excludePaths := []string{"/docs/", "/.claude/", "/.cursor/"}
	orphans, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	// Verify: Should find exactly 1 orphan (CBIN-999)
	if len(orphans) != 1 {
		t.Errorf("expected 1 orphan, got %d", len(orphans))
	}

	if len(orphans) > 0 && orphans[0].ReqID != "CBIN-999" {
		t.Errorf("expected CBIN-999, got %s", orphans[0].ReqID)
	}

	// Verify orphan has correct feature count
	if len(orphans) > 0 && len(orphans[0].Features) != 2 {
		t.Errorf("expected 2 features for CBIN-999, got %d", len(orphans[0].Features))
	}
}

// TestCANARY_CBIN_145_CLI_PathFiltering verifies path exclusion logic
func TestCANARY_CBIN_145_CLI_PathFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Insert tokens with various paths
	tokens := []*storage.Token{
		// Should be included
		{ReqID: "CBIN-200", Feature: "RealFeature", Aspect: "API", Status: "IMPL", FilePath: "pkg/api/handler.go", LineNumber: 10, UpdatedAt: "2025-10-17"},

		// Should be excluded (in /docs/)
		{ReqID: "CBIN-201", Feature: "DocExample", Aspect: "API", Status: "STUB", FilePath: "docs/user/examples.md", LineNumber: 20, UpdatedAt: "2025-10-17"},

		// Should be excluded (in /.claude/)
		{ReqID: "CBIN-202", Feature: "ClaudeExample", Aspect: "CLI", Status: "STUB", FilePath: ".claude/commands/example.md", LineNumber: 5, UpdatedAt: "2025-10-17"},

		// Should be excluded (in .canary/specs/)
		{ReqID: "CBIN-203", Feature: "SpecToken", Aspect: "Engine", Status: "STUB", FilePath: ".canary/specs/CBIN-203-test/spec.md", LineNumber: 100, UpdatedAt: "2025-10-17"},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	// Execute with path filtering
	excludePaths := []string{"/docs/", "/.claude/", "/.canary/specs/"}
	orphans, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	// Verify: Should find only CBIN-200
	if len(orphans) != 1 {
		t.Errorf("expected 1 orphan after path filtering, got %d", len(orphans))
	}

	if len(orphans) > 0 && orphans[0].ReqID != "CBIN-200" {
		t.Errorf("expected CBIN-200, got %s", orphans[0].ReqID)
	}
}

// TestCANARY_CBIN_145_CLI_SpecGeneration verifies spec.md generation from tokens
func TestCANARY_CBIN_145_CLI_SpecGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create orphaned requirement structure
	orphan := &migrate.OrphanedRequirement{
		ReqID: "CBIN-300",
		Features: []*storage.Token{
			{ReqID: "CBIN-300", Feature: "UserAuth", Aspect: "API", Status: "IMPL", FilePath: "pkg/auth/auth.go", LineNumber: 10, UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-300", Feature: "UserAuth", Aspect: "CLI", Status: "TESTED", Test: "TestCANARY_CBIN_300_CLI_Auth", FilePath: "cmd/auth_test.go", LineNumber: 50, UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-300", Feature: "UserAuthDocs", Aspect: "Docs", Status: "IMPL", FilePath: "docs/auth.md", LineNumber: 5, UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 3,
		Confidence:   migrate.ConfidenceHigh,
	}

	// Execute: Generate spec
	specContent, err := migrate.GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify: Spec contains required sections
	requiredSections := []string{
		"# Requirement Specification",
		"## Overview",
		"## User Stories",
		"## Functional Requirements",
		"## Implementation Checklist",
		"REQ=CBIN-300",
		"FEATURE=\"UserAuth\"",
		"ASPECT=API",
		"STATUS=IMPL",
	}

	for _, section := range requiredSections {
		if !contains(specContent, section) {
			t.Errorf("spec missing required section: %s", section)
		}
	}

	// Verify: Spec includes all features
	if !contains(specContent, "UserAuth") {
		t.Error("spec should contain UserAuth feature")
	}

	// Write and verify file creation
	specsDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-300-userauth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	specPath := filepath.Join(specsDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("spec file was not created")
	}
}

// TestCANARY_CBIN_145_CLI_PlanGeneration verifies plan.md generation
func TestCANARY_CBIN_145_CLI_PlanGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create orphaned requirement
	orphan := &migrate.OrphanedRequirement{
		ReqID: "CBIN-400",
		Features: []*storage.Token{
			{ReqID: "CBIN-400", Feature: "DataExport", Aspect: "Engine", Status: "IMPL", FilePath: "pkg/export/export.go", LineNumber: 15, UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-400", Feature: "ExportTests", Aspect: "Engine", Status: "TESTED", Test: "TestCANARY_CBIN_400_Engine_Export", FilePath: "pkg/export/export_test.go", LineNumber: 25, UpdatedAt: "2025-10-17"},
			{ReqID: "CBIN-400", Feature: "ExportBench", Aspect: "Bench", Status: "BENCHED", Bench: "BenchmarkCANARY_CBIN_400_Bench_LargeExport", FilePath: "pkg/export/export_bench_test.go", LineNumber: 100, UpdatedAt: "2025-10-17"},
		},
		FeatureCount: 3,
		Confidence:   migrate.ConfidenceHigh,
	}

	// Execute: Generate plan
	planContent, err := migrate.GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify: Plan contains required sections
	requiredSections := []string{
		"# Implementation Plan",
		"## Overview",
		"## Current Implementation Status",
		"## Architecture",
		"## Implementation Phases",
		"CBIN-400",
		"STATUS=IMPL",
		"STATUS=TESTED",
		"STATUS=BENCHED",
	}

	for _, section := range requiredSections {
		if !contains(planContent, section) {
			t.Errorf("plan missing required section: %s", section)
		}
	}

	// Verify: Plan reflects actual status progression
	if !contains(planContent, "DataExport") {
		t.Error("plan should contain DataExport feature")
	}

	// Write and verify
	specsDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-400-dataexport")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	planPath := filepath.Join(specsDir, "plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatalf("failed to write plan file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Error("plan file was not created")
	}
}

// TestCANARY_CBIN_145_CLI_DryRun verifies dry-run mode
func TestCANARY_CBIN_145_CLI_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Insert orphaned tokens
	tokens := []*storage.Token{
		{ReqID: "CBIN-500", Feature: "DryRunFeature", Aspect: "API", Status: "IMPL", FilePath: "pkg/dryrun/api.go", LineNumber: 10, UpdatedAt: "2025-10-17", IndexedAt: "2025-10-17T00:00:00Z"},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	// Execute dry-run
	excludePaths := []string{"/docs/"}
	plan, err := migrate.DryRun(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("DryRun failed: %v", err)
	}

	// Verify: Plan should list orphans without creating files
	if len(plan.Orphans) != 1 {
		t.Errorf("expected 1 orphan in dry-run plan, got %d", len(plan.Orphans))
	}

	// Verify: No files were created
	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to read specs directory: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("dry-run should not create files, but found %d entries", len(entries))
	}
}

// TestCANARY_CBIN_145_CLI_ConfidenceScoring verifies confidence calculation
func TestCANARY_CBIN_145_CLI_ConfidenceScoring(t *testing.T) {
	testCases := []struct {
		name          string
		featureCount  int
		hasTests      bool
		hasBenchmarks bool
		expectedLevel string
	}{
		{"Low confidence - 1 feature, no tests", 1, false, false, migrate.ConfidenceLow},
		{"Low confidence - 2 features, no tests", 2, false, false, migrate.ConfidenceLow},
		{"Medium confidence - 3 features, no tests", 3, false, false, migrate.ConfidenceMedium},
		{"Medium confidence - 2 features, has tests", 2, true, false, migrate.ConfidenceMedium},
		{"High confidence - 5 features, has tests", 5, true, false, migrate.ConfidenceHigh},
		{"High confidence - 3 features, tests + benchmarks", 3, true, true, migrate.ConfidenceHigh},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build orphan based on test case
			orphan := &migrate.OrphanedRequirement{
				ReqID:        "CBIN-TEST",
				FeatureCount: tc.featureCount,
			}

			// Add features
			for i := 0; i < tc.featureCount; i++ {
				token := &storage.Token{
					ReqID:    "CBIN-TEST",
					Feature:  "Feature",
					Aspect:   "API",
					Status:   "IMPL",
					FilePath: "test.go",
				}

				if tc.hasTests && i == 0 {
					token.Test = "TestCANARY_CBIN_TEST_API_Feature"
					token.Status = "TESTED"
				}

				if tc.hasBenchmarks && i == 1 {
					token.Bench = "BenchmarkCANARY_CBIN_TEST_Bench_Feature"
					token.Status = "BENCHED"
				}

				orphan.Features = append(orphan.Features, token)
			}

			// Calculate confidence
			confidence := migrate.CalculateConfidence(orphan)

			if confidence != tc.expectedLevel {
				t.Errorf("expected confidence %s, got %s", tc.expectedLevel, confidence)
			}
		})
	}
}

// contains function is defined in next_test.go and shared across test files
