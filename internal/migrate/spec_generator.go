// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17

package migrate

import (
	"fmt"
	"strings"
	"time"
)

// GenerateSpec creates a specification from an orphaned requirement
func GenerateSpec(orphan *OrphanedRequirement) (string, error) {
	if len(orphan.Features) == 0 {
		return "", fmt.Errorf("cannot generate spec for orphan with no features")
	}

	today := time.Now().Format("2006-01-02")

	// Extract feature name from first feature
	primaryFeature := orphan.Features[0].Feature
	featureName := strings.Title(primaryFeature)

	var spec strings.Builder

	// Header
	spec.WriteString(fmt.Sprintf("# Requirement Specification: %s\n\n", featureName))
	spec.WriteString(fmt.Sprintf("**Requirement ID:** %s\n", orphan.ReqID))
	spec.WriteString(fmt.Sprintf("**Generated:** %s (Auto-generated from legacy tokens)\n", today))
	spec.WriteString(fmt.Sprintf("**Confidence:** %s\n\n", orphan.Confidence))

	if orphan.Confidence == ConfidenceLow {
		spec.WriteString("> **⚠️ MIGRATION NOTICE:** This specification was automatically generated from existing code tokens.\n")
		spec.WriteString("> The confidence level is LOW. Please review and update manually to ensure accuracy.\n\n")
	}

	// Overview
	spec.WriteString("## Overview\n\n")
	spec.WriteString(fmt.Sprintf("**Purpose:** This specification was auto-generated from %d existing CANARY tokens found in the codebase.\n\n", orphan.FeatureCount))
	spec.WriteString("**Scope:**\n")
	spec.WriteString("- Included: Features already implemented in the codebase\n")
	spec.WriteString("- Excluded: New features not yet implemented\n\n")

	// User Stories (generated from features)
	spec.WriteString("## User Stories\n\n")
	spec.WriteString("### Primary User Stories\n\n")
	spec.WriteString("**US-1: Existing Implementation**\n")
	spec.WriteString("As a developer,\n")
	spec.WriteString(fmt.Sprintf("I want to have a specification for %s,\n", featureName))
	spec.WriteString("So that the existing implementation is properly documented.\n\n")
	spec.WriteString("**Acceptance Criteria:**\n")
	spec.WriteString("- [x] Implementation exists in codebase\n")
	spec.WriteString("- [ ] Specification updated with actual functionality\n")
	spec.WriteString("- [ ] Tests cover all features\n\n")

	// Functional Requirements (from features)
	spec.WriteString("## Functional Requirements\n\n")

	// Group features by aspect
	aspectGroups := make(map[string][]string)
	for _, token := range orphan.Features {
		aspectGroups[token.Aspect] = append(aspectGroups[token.Aspect], token.Feature)
	}

	frNum := 1
	for aspect, features := range aspectGroups {
		spec.WriteString(fmt.Sprintf("### FR-%d: %s Implementation\n", frNum, aspect))
		spec.WriteString("**Priority:** Medium\n")
		spec.WriteString(fmt.Sprintf("**Description:** Implements %s aspect with the following features:\n", aspect))
		for _, feature := range features {
			spec.WriteString(fmt.Sprintf("- %s\n", feature))
		}
		spec.WriteString("**Acceptance:** Existing implementation validated and tests pass\n\n")
		frNum++
	}

	// Success Criteria
	spec.WriteString("## Success Criteria\n\n")
	spec.WriteString("**Quantitative Metrics:**\n")
	spec.WriteString("- [x] All existing features have CANARY tokens\n")
	spec.WriteString("- [ ] All features have unit tests\n")
	spec.WriteString("- [ ] Code coverage > 80%\n\n")

	// Test Scenarios
	spec.WriteString("## User Scenarios & Testing\n\n")
	spec.WriteString("### Scenario 1: Feature Usage\n")
	spec.WriteString("**Given:** The system is running\n")
	spec.WriteString("**When:** User invokes the feature\n")
	spec.WriteString("**Then:** Feature operates as implemented\n\n")
	spec.WriteString("_Note: Update scenarios based on actual feature behavior_\n\n")

	// Dependencies
	spec.WriteString("## Dependencies\n\n")
	spec.WriteString("- Existing codebase implementation\n")
	spec.WriteString("- Current system architecture\n\n")

	// Implementation Checklist (from actual tokens)
	spec.WriteString("## Implementation Checklist\n\n")
	spec.WriteString("The following features were detected in the codebase:\n\n")

	for _, token := range orphan.Features {
		// Generate CANARY token
		tokenStr := fmt.Sprintf("<!-- CANARY: REQ=%s; FEATURE=\"%s\"; ASPECT=%s; STATUS=%s",
			orphan.ReqID, token.Feature, token.Aspect, token.Status)

		if token.Test != "" {
			tokenStr += fmt.Sprintf("; TEST=%s", token.Test)
		}
		if token.Bench != "" {
			tokenStr += fmt.Sprintf("; BENCH=%s", token.Bench)
		}

		tokenStr += fmt.Sprintf("; UPDATED=%s -->", today)

		spec.WriteString(tokenStr + "\n")
		spec.WriteString(fmt.Sprintf("**%s (%s)**\n", token.Feature, token.Aspect))
		spec.WriteString(fmt.Sprintf("- [%s] Status: %s\n", getCheckbox(token.Status), token.Status))
		spec.WriteString(fmt.Sprintf("- **Location:** %s:%d\n", token.FilePath, token.LineNumber))
		if token.Test != "" {
			spec.WriteString(fmt.Sprintf("- **Test:** %s\n", token.Test))
		}
		if token.Bench != "" {
			spec.WriteString(fmt.Sprintf("- **Benchmark:** %s\n", token.Bench))
		}
		spec.WriteString("\n")
	}

	// Review Checklist
	spec.WriteString("---\n\n")
	spec.WriteString("## Review & Update Checklist\n\n")
	spec.WriteString("**Post-Migration Tasks:**\n")
	spec.WriteString("- [ ] Review and update Overview section with actual purpose\n")
	spec.WriteString("- [ ] Add detailed user stories based on actual functionality\n")
	spec.WriteString("- [ ] Update functional requirements with specific details\n")
	spec.WriteString("- [ ] Define proper success criteria\n")
	spec.WriteString("- [ ] Add realistic test scenarios\n")
	spec.WriteString("- [ ] Document dependencies and constraints\n")
	spec.WriteString("- [ ] Remove this migration notice\n\n")

	spec.WriteString("---\n\n")
	spec.WriteString("_This specification was auto-generated by `canary migrate` on " + today + "_\n")

	return spec.String(), nil
}

// getCheckbox returns "x" for completed statuses, " " for incomplete
func getCheckbox(status string) string {
	if status == "IMPL" || status == "TESTED" || status == "BENCHED" {
		return "x"
	}
	return " "
}
