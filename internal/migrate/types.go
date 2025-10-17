// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="LegacyTokenMigration"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17

package migrate

import "go.spyder.org/canary/internal/storage"

// Confidence levels for orphan detection
const (
	ConfidenceLow    = "LOW"
	ConfidenceMedium = "MEDIUM"
	ConfidenceHigh   = "HIGH"
)

// OrphanedRequirement represents a requirement with tokens but no specification
type OrphanedRequirement struct {
	ReqID        string
	Features     []*storage.Token
	FeatureCount int
	Confidence   string
}

// OrphanPlan represents the plan for migrating orphaned requirements
type OrphanPlan struct {
	Orphans      []*OrphanedRequirement
	TotalOrphans int
	Excluded     []string
}
