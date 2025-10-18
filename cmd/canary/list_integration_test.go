// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-135; FEATURE="ListIntegration"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_Integration_EndToEnd; UPDATED=2025-10-17
func TestCANARY_CBIN_135_Integration_EndToEnd(t *testing.T) {
	// Test complete workflow: database setup → insert tokens → query → verify results
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Step 1: Initialize database
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Step 2: Insert realistic project tokens
	projectTokens := []*storage.Token{
		{
			ReqID:      "CBIN-400",
			Feature:    "UserAuthentication",
			Aspect:     "Security",
			Status:     "TESTED",
			FilePath:   "internal/auth/auth.go",
			LineNumber: 15,
			Test:       "TestUserAuthentication",
			Priority:   1,
			Owner:      "security-team",
			UpdatedAt:  "2025-10-17",
		},
		{
			ReqID:      "CBIN-401",
			Feature:    "DataValidation",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   "internal/api/validation.go",
			LineNumber: 42,
			Priority:   2,
			Owner:      "api-team",
			UpdatedAt:  "2025-10-16",
		},
		{
			ReqID:      "CBIN-402",
			Feature:    "RateLimiting",
			Aspect:     "Security",
			Status:     "STUB",
			FilePath:   ".canary/specs/CBIN-402-rate-limiting/spec.md",
			LineNumber: 1,
			Priority:   1,
			Owner:      "security-team",
			UpdatedAt:  "2025-10-15",
		},
	}

	for _, token := range projectTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token: %v", err)
		}
	}

	// Step 3: Test high-priority STUB work (typical development workflow)
	filters := map[string]string{
		"status":       "STUB",
		"priority_max": "1",
	}
	stubTokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 10)
	if err != nil {
		t.Fatalf("Query for STUB work failed: %v", err)
	}

	if len(stubTokens) != 1 {
		t.Errorf("Expected 1 high-priority STUB token, got %d", len(stubTokens))
	}

	if len(stubTokens) > 0 && stubTokens[0].Feature != "RateLimiting" {
		t.Errorf("Expected RateLimiting, got %s", stubTokens[0].Feature)
	}

	// Step 4: Test security team work (team-based filtering)
	secFilters := map[string]string{
		"owner":  "security-team",
		"aspect": "Security",
	}
	secTokens, err := db.ListTokens(secFilters, "CBIN-[1-9][0-9]{2,}", "status ASC", 0)
	if err != nil {
		t.Fatalf("Query for security team work failed: %v", err)
	}

	if len(secTokens) != 2 {
		t.Errorf("Expected 2 security team tokens, got %d", len(secTokens))
	}

	// Step 5: Test work needing tests (IMPL status)
	implFilters := map[string]string{"status": "IMPL"}
	implTokens, err := db.ListTokens(implFilters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
	if err != nil {
		t.Fatalf("Query for IMPL tokens failed: %v", err)
	}

	if len(implTokens) != 1 {
		t.Errorf("Expected 1 IMPL token needing tests, got %d", len(implTokens))
	}

	// Step 6: Test completed work (TESTED status)
	testedFilters := map[string]string{"status": "TESTED"}
	testedTokens, err := db.ListTokens(testedFilters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
	if err != nil {
		t.Fatalf("Query for TESTED tokens failed: %v", err)
	}

	if len(testedTokens) != 1 {
		t.Errorf("Expected 1 TESTED token, got %d", len(testedTokens))
	}

	if len(testedTokens) > 0 && testedTokens[0].Test == "" {
		t.Error("TESTED token should have Test field populated")
	}
}

// CANARY: REQ=CBIN-135; FEATURE="ListIntegration"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_Integration_JSONOutput; UPDATED=2025-10-17
func TestCANARY_CBIN_135_Integration_JSONOutput(t *testing.T) {
	// Test JSON output format for programmatic parsing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert test token
	token := &storage.Token{
		ReqID:      "CBIN-500",
		Feature:    "JSONTestFeature",
		Aspect:     "API",
		Status:     "IMPL",
		FilePath:   "api/json.go",
		LineNumber: 10,
		Priority:   1,
		UpdatedAt:  "2025-10-17",
	}

	if err := db.UpsertToken(token); err != nil {
		t.Fatalf("Failed to insert token: %v", err)
	}

	// Query tokens
	filters := make(map[string]string)
	tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(tokens)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	// Unmarshal and verify
	var parsed []*storage.Token
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("Expected 1 token in JSON, got %d", len(parsed))
	}

	if len(parsed) > 0 {
		if parsed[0].ReqID != "CBIN-500" {
			t.Errorf("JSON ReqID: got %s, want CBIN-500", parsed[0].ReqID)
		}
		if parsed[0].Feature != "JSONTestFeature" {
			t.Errorf("JSON Feature: got %s, want JSONTestFeature", parsed[0].Feature)
		}
	}
}

// CANARY: REQ=CBIN-135; FEATURE="ListIntegration"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_Integration_AgentWorkflow; UPDATED=2025-10-17
func TestCANARY_CBIN_135_Integration_AgentWorkflow(t *testing.T) {
	// Simulate AI agent workflow: query top 3 priorities with minimal context
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert multiple priority tokens
	for i := 1; i <= 10; i++ {
		token := &storage.Token{
			ReqID:      "CBIN-" + string(rune('6'+i/10)) + string(rune('0'+i%10)) + "0",
			Feature:    "Feature" + string(rune('0'+i)),
			Aspect:     "CLI",
			Status:     "STUB",
			FilePath:   "cli/feature.go",
			LineNumber: i * 10,
			Priority:   i, // 1 = highest priority
			UpdatedAt:  "2025-10-17",
		}
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token: %v", err)
		}
	}

	// Agent query: top 3 priorities for context-constrained environment
	filters := map[string]string{"status": "STUB"}
	tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 3)
	if err != nil {
		t.Fatalf("Agent query failed: %v", err)
	}

	// Verify exactly 3 results
	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens for agent context, got %d", len(tokens))
	}

	// Verify priorities are 1, 2, 3 (highest first)
	expectedPriorities := []int{1, 2, 3}
	for i, token := range tokens {
		if token.Priority != expectedPriorities[i] {
			t.Errorf("Token %d priority: got %d, want %d", i, token.Priority, expectedPriorities[i])
		}
	}

	// Verify all are STUB (work to be done)
	for i, token := range tokens {
		if token.Status != "STUB" {
			t.Errorf("Token %d status: got %s, want STUB", i, token.Status)
		}
	}
}

// CANARY: REQ=CBIN-135; FEATURE="ListIntegration"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_Integration_MultipleFilters; UPDATED=2025-10-17
func TestCANARY_CBIN_135_Integration_MultipleFilters(t *testing.T) {
	// Test complex filtering scenarios
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert diverse set of tokens
	testData := []struct {
		reqID    string
		feature  string
		aspect   string
		status   string
		owner    string
		priority int
	}{
		{"CBIN-700", "F1", "CLI", "STUB", "alice", 1},
		{"CBIN-701", "F2", "CLI", "IMPL", "alice", 2},
		{"CBIN-702", "F3", "API", "STUB", "alice", 1},
		{"CBIN-703", "F4", "API", "IMPL", "bob", 3},
		{"CBIN-704", "F5", "Engine", "TESTED", "bob", 2},
	}

	for i, data := range testData {
		token := &storage.Token{
			ReqID:      data.reqID,
			Feature:    data.feature,
			Aspect:     data.aspect,
			Status:     data.status,
			Owner:      data.owner,
			Priority:   data.priority,
			FilePath:   "internal/feature" + string(rune('0'+i)) + ".go",
			LineNumber: 1,
			UpdatedAt:  "2025-10-17",
		}
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token: %v", err)
		}
	}

	// Test 1: alice's CLI work
	filters := map[string]string{
		"owner":  "alice",
		"aspect": "CLI",
	}
	tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 alice CLI tokens, got %d", len(tokens))
	}

	// Test 2: High-priority STUB work (priority 1, status STUB)
	filters = map[string]string{
		"status":       "STUB",
		"priority_max": "1",
	}
	tokens, err = db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 high-priority STUB tokens, got %d", len(tokens))
	}

	// Test 3: bob's completed work
	filters = map[string]string{
		"owner":  "bob",
		"status": "TESTED",
	}
	tokens, err = db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("Expected 1 bob TESTED token, got %d", len(tokens))
	}
}

// CANARY: REQ=CBIN-135; FEATURE="ListIntegration"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_Integration_Performance; UPDATED=2025-10-17
func TestCANARY_CBIN_135_Integration_Performance(t *testing.T) {
	// Test performance with realistic dataset size
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert 200 tokens (simulating medium-sized project)
	for i := 100; i < 300; i++ {
		token := &storage.Token{
			ReqID:      "CBIN-" + string(rune('0'+i/100)) + string(rune('0'+(i%100)/10)) + string(rune('0'+i%10)),
			Feature:    "Feature" + string(rune('0'+i%100)),
			Aspect:     []string{"CLI", "API", "Engine", "Storage"}[i%4],
			Status:     []string{"STUB", "IMPL", "TESTED", "BENCHED"}[i%4],
			FilePath:   "src/file.go",
			LineNumber: i,
			Priority:   (i % 5) + 1,
			Owner:      []string{"alice", "bob", "charlie"}[i%3],
			UpdatedAt:  "2025-10-17",
		}
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token %d: %v", i, err)
		}
	}

	// Test query performance with filters
	tests := []struct {
		name    string
		filters map[string]string
		limit   int
	}{
		{
			name:    "top 10 priorities",
			filters: make(map[string]string),
			limit:   10,
		},
		{
			name:    "STUB work for alice",
			filters: map[string]string{"status": "STUB", "owner": "alice"},
			limit:   20,
		},
		{
			name:    "high-priority CLI work",
			filters: map[string]string{"aspect": "CLI", "priority_max": "2"},
			limit:   15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := db.ListTokens(tt.filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", tt.limit)
			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			if len(tokens) > tt.limit {
				t.Errorf("Expected max %d tokens, got %d", tt.limit, len(tokens))
			}

			// Verify results match filters
			for _, token := range tokens {
				if statusFilter, ok := tt.filters["status"]; ok {
					if token.Status != statusFilter {
						t.Errorf("Token status %s doesn't match filter %s", token.Status, statusFilter)
					}
				}
				if ownerFilter, ok := tt.filters["owner"]; ok {
					if token.Owner != ownerFilter {
						t.Errorf("Token owner %s doesn't match filter %s", token.Owner, ownerFilter)
					}
				}
			}
		})
	}
}
