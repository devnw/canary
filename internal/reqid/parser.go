// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-139; FEATURE="AspectIDParser"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
package reqid

import (
	"fmt"
	"regexp"
)

// RequirementID represents a parsed requirement identifier
type RequirementID struct {
	Key    string // e.g., "CBIN"
	Aspect string // e.g., "CLI", "API" (empty for v1 format)
	ID     string // e.g., "001", "042"
	Format string // "v1" (CBIN-XXX) or "v2" (CBIN-<ASPECT>-XXX)
}

var (
	// Pattern for new format: CBIN-CLI-001
	v2Pattern = regexp.MustCompile(`^([A-Z]+)-([A-Za-z]+)-(\d{3})$`)
	// Pattern for old format: CBIN-001
	v1Pattern = regexp.MustCompile(`^([A-Z]+)-(\d{3})$`)
)

// ParseRequirementID parses a requirement ID string into its components
func ParseRequirementID(reqID string) (*RequirementID, error) {
	// Try new format first (v2)
	if matches := v2Pattern.FindStringSubmatch(reqID); matches != nil {
		aspect := matches[2]

		// Validate aspect
		if err := ValidateAspect(aspect); err != nil {
			return nil, fmt.Errorf("invalid aspect %q: %w", aspect, err)
		}

		return &RequirementID{
			Key:    matches[1],
			Aspect: aspect,
			ID:     matches[3],
			Format: "v2",
		}, nil
	}

	// Try old format (v1) for backward compatibility
	if matches := v1Pattern.FindStringSubmatch(reqID); matches != nil {
		return &RequirementID{
			Key:    matches[1],
			ID:     matches[2],
			Format: "v1",
		}, nil
	}

	return nil, fmt.Errorf("invalid requirement ID format: %q", reqID)
}

// String returns the string representation of the RequirementID
func (r *RequirementID) String() string {
	if r.Format == "v2" && r.Aspect != "" {
		return fmt.Sprintf("%s-%s-%s", r.Key, r.Aspect, r.ID)
	}
	return fmt.Sprintf("%s-%s", r.Key, r.ID)
}
