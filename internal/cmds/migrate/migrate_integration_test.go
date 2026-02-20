// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="MigrationIntegrationTests"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_EndToEnd; UPDATED=2025-10-17

package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/migrate"
	"go.devnw.com/canary/internal/storage"
)

// TestCANARY_CBIN_145_CLI_EndToEnd verifies complete migration workflow
func TestCANARY_CBIN_145_CLI_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Setup: Create database with orphaned tokens
	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create .canary directory structure
	canaryDir := filepath.Join(tmpDir, ".canary")
	specsDir := filepath.Join(canaryDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create .canary/specs: %v", err)
	}

	// Insert orphaned tokens for CBIN-600
	timestamp := "2025-10-17T00:00:00Z"
	tokens := []*storage.Token{
		{
			ReqID:      "CBIN-600",
			Feature:    "PaymentAPI",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   "pkg/payment/api.go",
			LineNumber: 25,
			UpdatedAt:  "2025-10-17",
			IndexedAt:  timestamp,
		},
		{
			ReqID:      "CBIN-600",
			Feature:    "PaymentCLI",
			Aspect:     "CLI",
			Status:     "TESTED",
			Test:       "TestCANARY_CBIN_600_CLI_Payment",
			FilePath:   "cmd/payment/cli.go",
			LineNumber: 50,
			UpdatedAt:  "2025-10-17",
			IndexedAt:  timestamp,
		},
		{
			ReqID:      "CBIN-600",
			Feature:    "PaymentDocs",
			Aspect:     "Docs",
			Status:     "IMPL",
			FilePath:   "pkg/payment/README.md",
			LineNumber: 10,
			UpdatedAt:  "2025-10-17",
			IndexedAt:  timestamp,
		},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	// Step 1: Detect orphans
	excludePaths := []string{"/docs/", "/.claude/", "/.cursor/"}
	orphans, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphans))
	}

	orphan := orphans[0]
	if orphan.ReqID != "CBIN-600" {
		t.Errorf("expected CBIN-600, got %s", orphan.ReqID)
	}

	if len(orphan.Features) != 3 {
		t.Errorf("expected 3 features, got %d", len(orphan.Features))
	}

	// Step 2: Generate spec
	specContent, err := migrate.GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify spec content
	if !contains(specContent, "CBIN-600") {
		t.Error("spec should contain CBIN-600")
	}
	if !contains(specContent, "PaymentAPI") {
		t.Error("spec should contain PaymentAPI")
	}

	// Step 3: Write spec to file
	specDir := filepath.Join(specsDir, "CBIN-600-paymentapi")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Verify spec file exists and is readable
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("spec file was not created")
	}

	specData, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec file: %v", err)
	}

	if len(specData) == 0 {
		t.Error("spec file is empty")
	}

	// Step 4: Generate plan
	planContent, err := migrate.GeneratePlan(orphan)
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	// Verify plan content
	if !contains(planContent, "CBIN-600") {
		t.Error("plan should contain CBIN-600")
	}
	if !contains(planContent, "Implementation Plan") {
		t.Error("plan should have Implementation Plan header")
	}

	// Step 5: Write plan to file
	planPath := filepath.Join(specDir, "plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatalf("failed to write plan file: %v", err)
	}

	// Verify plan file exists and is readable
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		t.Error("plan file was not created")
	}

	planData, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("failed to read plan file: %v", err)
	}

	if len(planData) == 0 {
		t.Error("plan file is empty")
	}

	// Step 6: Verify orphan is no longer detected
	orphansAfter, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("second DetectOrphans failed: %v", err)
	}

	if len(orphansAfter) != 0 {
		t.Errorf("expected 0 orphans after migration, got %d", len(orphansAfter))
	}

	// Step 7: Verify directory structure
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		t.Fatalf("failed to read specs directory: %v", err)
	}

	foundCBIN600 := false
	for _, entry := range entries {
		if entry.Name() == "CBIN-600-paymentapi" {
			foundCBIN600 = true
		}
	}

	if !foundCBIN600 {
		t.Error("CBIN-600-paymentapi directory not found in specs/")
	}
}

// TestCANARY_CBIN_145_CLI_BatchMigration verifies migrating multiple orphans
func TestCANARY_CBIN_145_CLI_BatchMigration(t *testing.T) {
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

	// Create .canary/specs
	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	// Insert tokens for 3 different orphaned requirements
	timestamp := "2025-10-17T00:00:00Z"
	tokens := []*storage.Token{
		// CBIN-700: Database feature
		{ReqID: "CBIN-700", Feature: "DatabaseAPI", Aspect: "API", Status: "IMPL", FilePath: "pkg/db/api.go", LineNumber: 10, UpdatedAt: "2025-10-17", IndexedAt: timestamp},
		{ReqID: "CBIN-700", Feature: "DatabaseTests", Aspect: "Storage", Status: "TESTED", Test: "TestCANARY_CBIN_700_Storage_DB", FilePath: "pkg/db/storage.go", LineNumber: 20, UpdatedAt: "2025-10-17", IndexedAt: timestamp},

		// CBIN-701: Cache feature
		{ReqID: "CBIN-701", Feature: "CacheLayer", Aspect: "Engine", Status: "IMPL", FilePath: "pkg/cache/cache.go", LineNumber: 15, UpdatedAt: "2025-10-17", IndexedAt: timestamp},
		{ReqID: "CBIN-701", Feature: "CacheTests", Aspect: "Engine", Status: "TESTED", Test: "TestCANARY_CBIN_701_Engine_Cache", FilePath: "pkg/cache/engine.go", LineNumber: 25, UpdatedAt: "2025-10-17", IndexedAt: timestamp},

		// CBIN-702: Logging feature
		{ReqID: "CBIN-702", Feature: "Logger", Aspect: "Engine", Status: "IMPL", FilePath: "pkg/log/log.go", LineNumber: 5, UpdatedAt: "2025-10-17", IndexedAt: timestamp},
		{ReqID: "CBIN-702", Feature: "LoggerTests", Aspect: "Engine", Status: "TESTED", Test: "TestCANARY_CBIN_702_Engine_Logger", FilePath: "pkg/log/logger.go", LineNumber: 30, UpdatedAt: "2025-10-17", IndexedAt: timestamp},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	// Execute: Detect all orphans
	excludePaths := []string{"/docs/"}
	orphans, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 3 {
		t.Fatalf("expected 3 orphans, got %d", len(orphans))
	}

	// Migrate all orphans
	migratedCount := 0
	for _, orphan := range orphans {
		// Generate spec
		specContent, err := migrate.GenerateSpec(orphan)
		if err != nil {
			t.Errorf("GenerateSpec failed for %s: %v", orphan.ReqID, err)
			continue
		}

		// Generate plan
		planContent, err := migrate.GeneratePlan(orphan)
		if err != nil {
			t.Errorf("GeneratePlan failed for %s: %v", orphan.ReqID, err)
			continue
		}

		// Create directory
		dirName := orphan.ReqID + "-" + slugify(orphan.Features[0].Feature)
		specDir := filepath.Join(specsDir, dirName)
		if err := os.MkdirAll(specDir, 0755); err != nil {
			t.Errorf("failed to create directory for %s: %v", orphan.ReqID, err)
			continue
		}

		// Write spec
		specPath := filepath.Join(specDir, "spec.md")
		if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
			t.Errorf("failed to write spec for %s: %v", orphan.ReqID, err)
			continue
		}

		// Write plan
		planPath := filepath.Join(specDir, "plan.md")
		if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
			t.Errorf("failed to write plan for %s: %v", orphan.ReqID, err)
			continue
		}

		migratedCount++
	}

	if migratedCount != 3 {
		t.Errorf("expected to migrate 3 requirements, migrated %d", migratedCount)
	}

	// Verify all orphans now have specs
	orphansAfter, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("second DetectOrphans failed: %v", err)
	}

	if len(orphansAfter) != 0 {
		t.Errorf("expected 0 orphans after batch migration, got %d", len(orphansAfter))
	}

	// Verify directory structure
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		t.Fatalf("failed to read specs directory: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 spec directories, found %d", len(entries))
	}
}

// TestCANARY_CBIN_145_CLI_IdempotentMigration verifies safe re-running
func TestCANARY_CBIN_145_CLI_IdempotentMigration(t *testing.T) {
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

	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	// Insert orphaned tokens
	tokens := []*storage.Token{
		{ReqID: "CBIN-800", Feature: "Feature800", Aspect: "API", Status: "IMPL", FilePath: "pkg/feature800/api.go", LineNumber: 10, UpdatedAt: "2025-10-17", IndexedAt: "2025-10-17T00:00:00Z"},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	excludePaths := []string{"/docs/"}

	// First migration
	orphans1, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("first DetectOrphans failed: %v", err)
	}

	if len(orphans1) != 1 {
		t.Fatalf("expected 1 orphan before migration, got %d", len(orphans1))
	}

	orphan := orphans1[0]
	specContent, _ := migrate.GenerateSpec(orphan)
	planContent, _ := migrate.GeneratePlan(orphan)

	specDir := filepath.Join(specsDir, "CBIN-800-feature800")
	os.MkdirAll(specDir, 0755)

	specPath := filepath.Join(specDir, "spec.md")
	planPath := filepath.Join(specDir, "plan.md")

	os.WriteFile(specPath, []byte(specContent), 0644)
	os.WriteFile(planPath, []byte(planContent), 0644)

	// Second migration (should be idempotent)
	orphans2, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("second DetectOrphans failed: %v", err)
	}

	if len(orphans2) != 0 {
		t.Errorf("expected 0 orphans after migration, got %d", len(orphans2))
	}

	// Third migration (verify still idempotent)
	orphans3, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("third DetectOrphans failed: %v", err)
	}

	if len(orphans3) != 0 {
		t.Errorf("expected 0 orphans on third check, got %d", len(orphans3))
	}

	// Verify files weren't duplicated or corrupted
	specData, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec after multiple runs: %v", err)
	}

	if len(specData) == 0 {
		t.Error("spec file became empty after multiple runs")
	}
}

// TestCANARY_CBIN_145_CLI_LowConfidenceWarning verifies low-confidence flagging
func TestCANARY_CBIN_145_CLI_LowConfidenceWarning(t *testing.T) {
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

	// Insert tokens with minimal features (low confidence scenario)
	tokens := []*storage.Token{
		{ReqID: "CBIN-900", Feature: "SingleFeature", Aspect: "API", Status: "STUB", FilePath: "pkg/single/api.go", LineNumber: 10, UpdatedAt: "2025-10-17", IndexedAt: "2025-10-17T00:00:00Z"},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to upsert token: %v", err)
		}
	}

	// Detect orphans
	excludePaths := []string{"/docs/"}
	orphans, err := migrate.DetectOrphans(db, tmpDir, excludePaths)
	if err != nil {
		t.Fatalf("DetectOrphans failed: %v", err)
	}

	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphans))
	}

	orphan := orphans[0]

	// Calculate confidence
	confidence := migrate.CalculateConfidence(orphan)

	// Verify low confidence
	if confidence != migrate.ConfidenceLow {
		t.Errorf("expected low confidence for single STUB feature, got %s", confidence)
	}

	// Generate spec should still work but include warning
	specContent, err := migrate.GenerateSpec(orphan)
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify warning in spec
	if !contains(specContent, "CONFIDENCE") || !contains(specContent, "LOW") {
		t.Error("spec should contain low confidence warning")
	}
}

// slugify function is defined in migrate.go and shared
// contains function is defined in next_test.go and shared across test files
