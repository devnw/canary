package specs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_FullSatisfied; UPDATED=2025-10-18
func TestCheckDependencyStatus_FullSatisfied(t *testing.T) {
	// Mock token provider with all features TESTED
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
				{Feature: "ContextManagement", Aspect: "Engine", Status: "TESTED"},
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source: "CBIN-147",
		Target: "CBIN-146",
		Type:   DependencyTypeFull,
	}

	status := checker.CheckDependency(dep)

	assert.True(t, status.IsSatisfied)
	assert.False(t, status.Blocking)
	assert.Contains(t, status.Message, "satisfied")
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_FullBlocked; UPDATED=2025-10-18
func TestCheckDependencyStatus_FullBlocked(t *testing.T) {
	// Mock token provider with one feature only IMPL
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
				{Feature: "ContextManagement", Aspect: "Engine", Status: "IMPL"}, // Blocking
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source: "CBIN-147",
		Target: "CBIN-146",
		Type:   DependencyTypeFull,
	}

	status := checker.CheckDependency(dep)

	assert.False(t, status.IsSatisfied)
	assert.True(t, status.Blocking)
	assert.Contains(t, status.Message, "not yet TESTED/BENCHED")
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_PartialFeaturesSatisfied; UPDATED=2025-10-18
func TestCheckDependencyStatus_PartialFeaturesSatisfied(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
				{Feature: "ContextManagement", Aspect: "Engine", Status: "TESTED"},
				{Feature: "TokenNamespacing", Aspect: "Storage", Status: "IMPL"}, // Not required
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source:           "CBIN-147",
		Target:           "CBIN-146",
		Type:             DependencyTypePartialFeatures,
		RequiredFeatures: []string{"ProjectRegistry", "ContextManagement"},
	}

	status := checker.CheckDependency(dep)

	assert.True(t, status.IsSatisfied)
	assert.False(t, status.Blocking)
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_PartialFeaturesBlocked; UPDATED=2025-10-18
func TestCheckDependencyStatus_PartialFeaturesBlocked(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
				{Feature: "ContextManagement", Aspect: "Engine", Status: "IMPL"}, // Blocking
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source:           "CBIN-147",
		Target:           "CBIN-146",
		Type:             DependencyTypePartialFeatures,
		RequiredFeatures: []string{"ProjectRegistry", "ContextManagement"},
	}

	status := checker.CheckDependency(dep)

	assert.False(t, status.IsSatisfied)
	assert.True(t, status.Blocking)
	assert.Len(t, status.MissingFeatures, 1)
	assert.Contains(t, status.MissingFeatures, "ContextManagement")
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_PartialAspectSatisfied; UPDATED=2025-10-18
func TestCheckDependencyStatus_PartialAspectSatisfied(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-129": {
				{Feature: "DatabaseMigrations", Aspect: "Storage", Status: "TESTED"},
				{Feature: "MigrationRollback", Aspect: "Storage", Status: "BENCHED"},
				{Feature: "MigrationCLI", Aspect: "CLI", Status: "IMPL"}, // Different aspect
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source:         "CBIN-147",
		Target:         "CBIN-129",
		Type:           DependencyTypePartialAspect,
		RequiredAspect: "Storage",
	}

	status := checker.CheckDependency(dep)

	assert.True(t, status.IsSatisfied)
	assert.False(t, status.Blocking)
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_PartialAspectBlocked; UPDATED=2025-10-18
func TestCheckDependencyStatus_PartialAspectBlocked(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-129": {
				{Feature: "DatabaseMigrations", Aspect: "Storage", Status: "TESTED"},
				{Feature: "MigrationRollback", Aspect: "Storage", Status: "IMPL"}, // Blocking
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source:         "CBIN-147",
		Target:         "CBIN-129",
		Type:           DependencyTypePartialAspect,
		RequiredAspect: "Storage",
	}

	status := checker.CheckDependency(dep)

	assert.False(t, status.IsSatisfied)
	assert.True(t, status.Blocking)
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_MissingRequirement; UPDATED=2025-10-18
func TestCheckDependencyStatus_MissingRequirement(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{}, // No tokens for CBIN-999
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source: "CBIN-147",
		Target: "CBIN-999",
		Type:   DependencyTypeFull,
	}

	status := checker.CheckDependency(dep)

	assert.False(t, status.IsSatisfied)
	assert.True(t, status.Blocking)
	assert.Contains(t, status.Message, "not found")
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_BENCHEDAccepted; UPDATED=2025-10-18
func TestCheckDependencyStatus_BENCHEDAccepted(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "BENCHED"},
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source: "CBIN-147",
		Target: "CBIN-146",
		Type:   DependencyTypeFull,
	}

	status := checker.CheckDependency(dep)

	assert.True(t, status.IsSatisfied)
	assert.False(t, status.Blocking)
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckAllDependencies; UPDATED=2025-10-18
func TestCheckAllDependencies(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
			},
			"CBIN-145": {
				{Feature: "Migration", Aspect: "Engine", Status: "IMPL"}, // Blocking
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	deps := []Dependency{
		{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull},
		{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull},
	}

	statuses := checker.CheckAllDependencies(deps)

	require.Len(t, statuses, 2)
	assert.True(t, statuses[0].IsSatisfied)
	assert.False(t, statuses[1].IsSatisfied)
}

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestGetBlockingDependencies; UPDATED=2025-10-18
func TestGetBlockingDependencies(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
			},
			"CBIN-145": {
				{Feature: "Migration", Aspect: "Engine", Status: "IMPL"},
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	deps := []Dependency{
		{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull},
		{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull},
	}

	blocking := checker.GetBlockingDependencies(deps)

	require.Len(t, blocking, 1)
	assert.Equal(t, "CBIN-145", blocking[0].Dependency.Target)
}

// MockTokenProvider simulates reading CANARY tokens from storage
type MockTokenProvider struct {
	tokens map[string][]MockToken
}

type MockToken struct {
	Feature string
	Aspect  string
	Status  string
}

func (m *MockTokenProvider) GetTokensByReqID(reqID string) []TokenInfo {
	mockTokens, exists := m.tokens[reqID]
	if !exists {
		return []TokenInfo{}
	}

	var tokens []TokenInfo
	for _, mt := range mockTokens {
		tokens = append(tokens, TokenInfo{
			ReqID:   reqID,
			Feature: mt.Feature,
			Aspect:  mt.Aspect,
			Status:  mt.Status,
		})
	}
	return tokens
}
