package specs

import (
	"fmt"
	"strings"
)

// CANARY: REQ=CBIN-147; FEATURE="StatusChecker"; ASPECT=Engine; STATUS=TESTED; TEST=TestCheckDependencyStatus_FullSatisfied,TestCheckDependencyStatus_FullBlocked,TestCheckDependencyStatus_PartialFeaturesSatisfied,TestCheckDependencyStatus_PartialFeaturesBlocked; UPDATED=2025-10-18

// TokenProvider is an interface for retrieving CANARY tokens from storage.
// This allows the status checker to query token status without tight coupling to storage.
type TokenProvider interface {
	// GetTokensByReqID returns all CANARY tokens for a given requirement ID
	GetTokensByReqID(reqID string) []TokenInfo
}

// StatusChecker checks whether dependencies are satisfied based on CANARY token status.
type StatusChecker struct {
	tokenProvider TokenProvider
}

// NewStatusChecker creates a new status checker with the given token provider.
func NewStatusChecker(provider TokenProvider) *StatusChecker {
	return &StatusChecker{
		tokenProvider: provider,
	}
}

// CheckDependency checks whether a single dependency is satisfied.
// Satisfaction rules:
// - Full: All features of target requirement must be TESTED or BENCHED
// - PartialFeatures: All RequiredFeatures must be TESTED or BENCHED
// - PartialAspect: All features of RequiredAspect must be TESTED or BENCHED
//
// IMPL status is NOT sufficient - dependencies require tests.
func (sc *StatusChecker) CheckDependency(dep Dependency) DependencyStatus {
	// Get all tokens for target requirement
	tokens := sc.tokenProvider.GetTokensByReqID(dep.Target)

	// Check if requirement exists
	if len(tokens) == 0 {
		return DependencyStatus{
			Dependency:    dep,
			IsSatisfied:   false,
			Blocking:      true,
			Message:       fmt.Sprintf("Requirement %s not found or has no tokens", dep.Target),
			CurrentStatus: "MISSING",
		}
	}

	switch dep.Type {
	case DependencyTypeFull:
		return sc.checkFullDependency(dep, tokens)
	case DependencyTypePartialFeatures:
		return sc.checkPartialFeaturesDependency(dep, tokens)
	case DependencyTypePartialAspect:
		return sc.checkPartialAspectDependency(dep, tokens)
	default:
		return DependencyStatus{
			Dependency:  dep,
			IsSatisfied: false,
			Blocking:    true,
			Message:     fmt.Sprintf("Unknown dependency type: %v", dep.Type),
		}
	}
}

// checkFullDependency verifies all features are TESTED or BENCHED.
func (sc *StatusChecker) checkFullDependency(dep Dependency, tokens []TokenInfo) DependencyStatus {
	var missingFeatures []string
	allStatuses := make(map[string]string) // feature -> status

	for _, token := range tokens {
		allStatuses[token.Feature] = token.Status
		if !isStatusSatisfied(token.Status) {
			missingFeatures = append(missingFeatures, token.Feature)
		}
	}

	if len(missingFeatures) == 0 {
		return DependencyStatus{
			Dependency:    dep,
			IsSatisfied:   true,
			Blocking:      false,
			Message:       fmt.Sprintf("All features of %s are satisfied", dep.Target),
			CurrentStatus: "SATISFIED",
		}
	}

	return DependencyStatus{
		Dependency:      dep,
		IsSatisfied:     false,
		Blocking:        true,
		Message:         fmt.Sprintf("%s has %d feature(s) not yet TESTED/BENCHED", dep.Target, len(missingFeatures)),
		MissingFeatures: missingFeatures,
		CurrentStatus:   "PARTIAL",
	}
}

// checkPartialFeaturesDependency verifies only the required features are TESTED or BENCHED.
func (sc *StatusChecker) checkPartialFeaturesDependency(dep Dependency, tokens []TokenInfo) DependencyStatus {
	// Build map of feature -> status
	featureStatus := make(map[string]string)
	for _, token := range tokens {
		featureStatus[token.Feature] = token.Status
	}

	var missingFeatures []string
	for _, requiredFeature := range dep.RequiredFeatures {
		status, exists := featureStatus[requiredFeature]
		if !exists {
			missingFeatures = append(missingFeatures, requiredFeature)
			continue
		}
		if !isStatusSatisfied(status) {
			missingFeatures = append(missingFeatures, requiredFeature)
		}
	}

	if len(missingFeatures) == 0 {
		return DependencyStatus{
			Dependency:    dep,
			IsSatisfied:   true,
			Blocking:      false,
			Message:       fmt.Sprintf("Required features %v of %s are satisfied", dep.RequiredFeatures, dep.Target),
			CurrentStatus: "SATISFIED",
		}
	}

	return DependencyStatus{
		Dependency:      dep,
		IsSatisfied:     false,
		Blocking:        true,
		Message:         fmt.Sprintf("%s missing features: %v", dep.Target, missingFeatures),
		MissingFeatures: missingFeatures,
		CurrentStatus:   "PARTIAL",
	}
}

// checkPartialAspectDependency verifies all features of the required aspect are TESTED or BENCHED.
func (sc *StatusChecker) checkPartialAspectDependency(dep Dependency, tokens []TokenInfo) DependencyStatus {
	var aspectTokens []TokenInfo
	for _, token := range tokens {
		if token.Aspect == dep.RequiredAspect {
			aspectTokens = append(aspectTokens, token)
		}
	}

	if len(aspectTokens) == 0 {
		return DependencyStatus{
			Dependency:    dep,
			IsSatisfied:   false,
			Blocking:      true,
			Message:       fmt.Sprintf("No features found for aspect %s in %s", dep.RequiredAspect, dep.Target),
			CurrentStatus: "MISSING",
		}
	}

	var missingFeatures []string
	for _, token := range aspectTokens {
		if !isStatusSatisfied(token.Status) {
			missingFeatures = append(missingFeatures, token.Feature)
		}
	}

	if len(missingFeatures) == 0 {
		return DependencyStatus{
			Dependency:    dep,
			IsSatisfied:   true,
			Blocking:      false,
			Message:       fmt.Sprintf("All features of aspect %s in %s are satisfied", dep.RequiredAspect, dep.Target),
			CurrentStatus: "SATISFIED",
		}
	}

	return DependencyStatus{
		Dependency:      dep,
		IsSatisfied:     false,
		Blocking:        true,
		Message:         fmt.Sprintf("%s aspect %s has unsatisfied features: %v", dep.Target, dep.RequiredAspect, missingFeatures),
		MissingFeatures: missingFeatures,
		CurrentStatus:   "PARTIAL",
	}
}

// CheckAllDependencies checks all dependencies and returns their statuses.
func (sc *StatusChecker) CheckAllDependencies(deps []Dependency) []DependencyStatus {
	statuses := make([]DependencyStatus, len(deps))
	for i, dep := range deps {
		statuses[i] = sc.CheckDependency(dep)
	}
	return statuses
}

// GetBlockingDependencies returns only the dependencies that are blocking (not satisfied).
func (sc *StatusChecker) GetBlockingDependencies(deps []Dependency) []DependencyStatus {
	var blocking []DependencyStatus
	for _, dep := range deps {
		status := sc.CheckDependency(dep)
		if status.Blocking {
			blocking = append(blocking, status)
		}
	}
	return blocking
}

// FormatBlockingReport generates a human-readable report of blocking dependencies.
func (sc *StatusChecker) FormatBlockingReport(deps []Dependency) string {
	blocking := sc.GetBlockingDependencies(deps)
	if len(blocking) == 0 {
		return "All dependencies are satisfied ✅"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("⚠️  %d blocking dependencies:", len(blocking)))
	for i, status := range blocking {
		lines = append(lines, fmt.Sprintf("\n%d. %s -> %s (%s)",
			i+1, status.Dependency.Source, status.Dependency.Target, status.Dependency.Type))
		lines = append(lines, fmt.Sprintf("   Status: %s", status.Message))
		if len(status.MissingFeatures) > 0 {
			lines = append(lines, fmt.Sprintf("   Missing: %s", strings.Join(status.MissingFeatures, ", ")))
		}
	}

	return strings.Join(lines, "\n")
}

// isStatusSatisfied returns true if the status satisfies dependency requirements.
// Only TESTED and BENCHED satisfy dependencies. IMPL is insufficient.
func isStatusSatisfied(status string) bool {
	return status == "TESTED" || status == "BENCHED"
}
