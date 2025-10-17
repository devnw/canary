// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="PlanGeneration"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17

package migrate

import (
	"fmt"
	"strings"
	"time"
)

// GeneratePlan creates an implementation plan from an orphaned requirement
func GeneratePlan(orphan *OrphanedRequirement) (string, error) {
	if len(orphan.Features) == 0 {
		return "", fmt.Errorf("cannot generate plan for orphan with no features")
	}

	today := time.Now().Format("2006-01-02")
	primaryFeature := orphan.Features[0].Feature

	var plan strings.Builder

	// Header
	plan.WriteString(fmt.Sprintf("# Implementation Plan: %s\n\n", primaryFeature))
	plan.WriteString(fmt.Sprintf("**Requirement ID:** %s\n", orphan.ReqID))
	plan.WriteString(fmt.Sprintf("**Generated:** %s (Auto-generated from legacy tokens)\n", today))
	plan.WriteString(fmt.Sprintf("**Confidence:** %s\n\n", orphan.Confidence))

	if orphan.Confidence == ConfidenceLow || orphan.Confidence == ConfidenceMedium {
		plan.WriteString("> **⚠️ MIGRATION NOTICE:** This plan was generated from existing code.\n")
		plan.WriteString("> Please review and update to reflect the actual implementation approach.\n\n")
	}

	// Overview
	plan.WriteString("## Overview\n\n")
	plan.WriteString(fmt.Sprintf("This plan documents the existing implementation of %s, which was detected in the codebase with %d features.\n\n", primaryFeature, orphan.FeatureCount))

	// Current Implementation Status
	plan.WriteString("## Current Implementation Status\n\n")

	statusCounts := make(map[string]int)
	for _, token := range orphan.Features {
		statusCounts[token.Status]++
	}

	plan.WriteString("**Status Breakdown:**\n")
	for _, status := range []string{"STUB", "IMPL", "TESTED", "BENCHED"} {
		count := statusCounts[status]
		if count > 0 {
			plan.WriteString(fmt.Sprintf("- %s: %d features\n", status, count))
		}
	}
	plan.WriteString("\n")

	// Architecture
	plan.WriteString("## Architecture\n\n")
	plan.WriteString("**Components:**\n\n")

	aspectGroups := make(map[string][]string)
	aspectFiles := make(map[string]map[string]bool)
	for _, token := range orphan.Features {
		aspectGroups[token.Aspect] = append(aspectGroups[token.Aspect], token.Feature)
		if aspectFiles[token.Aspect] == nil {
			aspectFiles[token.Aspect] = make(map[string]bool)
		}
		aspectFiles[token.Aspect][token.FilePath] = true
	}

	for aspect, features := range aspectGroups {
		plan.WriteString(fmt.Sprintf("### %s Layer\n", aspect))
		plan.WriteString(fmt.Sprintf("**Features:** %d\n", len(features)))
		plan.WriteString("**Files:**\n")
		for file := range aspectFiles[aspect] {
			plan.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		plan.WriteString("\n")
	}

	// Implementation Phases
	plan.WriteString("## Implementation Phases\n\n")
	plan.WriteString("_Note: These phases reflect the current state of the codebase._\n\n")

	// Phase 1: Stubbed features
	if statusCounts["STUB"] > 0 {
		plan.WriteString("### Phase 1: Stub Features (TODO)\n\n")
		plan.WriteString("The following features are stubbed and need implementation:\n\n")
		for _, token := range orphan.Features {
			if token.Status == "STUB" {
				plan.WriteString(fmt.Sprintf("- **%s** (%s) - `%s:%d`\n", token.Feature, token.Aspect, token.FilePath, token.LineNumber))
			}
		}
		plan.WriteString("\n")
	}

	// Phase 2: Implemented features
	if statusCounts["IMPL"] > 0 {
		plan.WriteString("### Phase 2: Implemented Features (COMPLETE)\n\n")
		plan.WriteString("The following features are implemented:\n\n")
		for _, token := range orphan.Features {
			if token.Status == "IMPL" {
				plan.WriteString(fmt.Sprintf("- **%s** (%s) - `%s:%d`\n", token.Feature, token.Aspect, token.FilePath, token.LineNumber))
			}
		}
		plan.WriteString("\n")
	}

	// Phase 3: Tested features
	if statusCounts["TESTED"] > 0 {
		plan.WriteString("### Phase 3: Tested Features (COMPLETE)\n\n")
		plan.WriteString("The following features have tests:\n\n")
		for _, token := range orphan.Features {
			if token.Status == "TESTED" {
				plan.WriteString(fmt.Sprintf("- **%s** (%s) - `%s:%d`\n", token.Feature, token.Aspect, token.FilePath, token.LineNumber))
				if token.Test != "" {
					plan.WriteString(fmt.Sprintf("  - Test: `%s`\n", token.Test))
				}
			}
		}
		plan.WriteString("\n")
	}

	// Phase 4: Benchmarked features
	if statusCounts["BENCHED"] > 0 {
		plan.WriteString("### Phase 4: Benchmarked Features (COMPLETE)\n\n")
		plan.WriteString("The following features have benchmarks:\n\n")
		for _, token := range orphan.Features {
			if token.Status == "BENCHED" {
				plan.WriteString(fmt.Sprintf("- **%s** (%s) - `%s:%d`\n", token.Feature, token.Aspect, token.FilePath, token.LineNumber))
				if token.Bench != "" {
					plan.WriteString(fmt.Sprintf("  - Benchmark: `%s`\n", token.Bench))
				}
			}
		}
		plan.WriteString("\n")
	}

	// Testing Strategy
	plan.WriteString("## Testing Strategy\n\n")

	hasTests := false
	hasBenchmarks := false
	for _, token := range orphan.Features {
		if token.Test != "" {
			hasTests = true
		}
		if token.Bench != "" {
			hasBenchmarks = true
		}
	}

	if hasTests {
		plan.WriteString("**Existing Tests:**\n")
		for _, token := range orphan.Features {
			if token.Test != "" {
				plan.WriteString(fmt.Sprintf("- `%s`\n", token.Test))
			}
		}
		plan.WriteString("\n")
	} else {
		plan.WriteString("**Test Coverage:** No tests detected. Consider adding:\n")
		plan.WriteString("- Unit tests for core functionality\n")
		plan.WriteString("- Integration tests for end-to-end flows\n\n")
	}

	if hasBenchmarks {
		plan.WriteString("**Existing Benchmarks:**\n")
		for _, token := range orphan.Features {
			if token.Bench != "" {
				plan.WriteString(fmt.Sprintf("- `%s`\n", token.Bench))
			}
		}
		plan.WriteString("\n")
	}

	// Next Steps
	plan.WriteString("## Next Steps\n\n")
	plan.WriteString("1. Review this auto-generated plan for accuracy\n")
	plan.WriteString("2. Update the Overview section with implementation approach\n")
	plan.WriteString("3. Document architectural decisions\n")

	if statusCounts["STUB"] > 0 {
		plan.WriteString(fmt.Sprintf("4. Complete %d stubbed features\n", statusCounts["STUB"]))
	}

	if !hasTests {
		plan.WriteString("4. Add comprehensive test coverage\n")
	}

	plan.WriteString("5. Update CANARY tokens as implementation progresses\n")
	plan.WriteString(fmt.Sprintf("6. Run `canary status %s` to track progress\n\n", orphan.ReqID))

	// Constitutional Compliance
	plan.WriteString("## Constitutional Compliance\n\n")
	plan.WriteString("**Article I (Requirement-First):** ✅ Specification now exists\n")
	plan.WriteString("**Article II (Specification Discipline):** ⚠️  Spec needs review\n")

	if hasTests {
		plan.WriteString("**Article IV (Test-First):** ✅ Tests exist\n")
	} else {
		plan.WriteString("**Article IV (Test-First):** ❌ Tests needed\n")
	}

	plan.WriteString("**Article VII (Documentation Currency):** ✅ Tokens up to date\n\n")

	plan.WriteString("---\n\n")
	plan.WriteString("_This plan was auto-generated by `canary migrate` on " + today + "_\n")

	return plan.String(), nil
}
