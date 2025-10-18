package specs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_FullDependency,TestParseDependencies_PartialFeatures,TestParseDependencies_PartialAspect,TestParseDependencies_MixedTypes; UPDATED=2025-10-18

var (
	// Regex patterns for parsing dependency lines
	// Format: "- CBIN-123 (Description)" for full dependencies
	// Format: "- CBIN-123:Feature1,Feature2 (Description)" for partial feature dependencies
	// Format: "- CBIN-123:AspectName (Description)" for partial aspect dependencies
	fullDependencyPattern = regexp.MustCompile(`^-\s+(CBIN-\d+)\s*(?:\(([^)]+)\))?`)
	partialDependencyPattern = regexp.MustCompile(`^-\s+(CBIN-\d+):([^(\s]+)\s*(?:\(([^)]+)\))?`)
)

// ParseDependenciesFromFile reads a spec.md file and extracts all dependencies.
// Returns a slice of Dependency objects or an error if the file cannot be read.
func ParseDependenciesFromFile(sourceReqID, specPath string) ([]Dependency, error) {
	file, err := os.Open(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open spec file %s: %w", specPath, err)
	}
	defer file.Close()

	return ParseDependencies(sourceReqID, file)
}

// ParseDependencies parses dependency declarations from a spec.md file reader.
// It looks for the "## Dependencies" section and extracts all dependency declarations.
//
// Supported formats:
// - Full: "- CBIN-123 (Description)"
// - Partial Features: "- CBIN-123:Feature1,Feature2 (Description)"
// - Partial Aspect: "- CBIN-123:AspectName (Description)"
//
// Returns a slice of Dependency objects. Returns empty slice if no dependencies found.
func ParseDependencies(sourceReqID string, reader io.Reader) ([]Dependency, error) {
	var dependencies []Dependency
	scanner := bufio.NewScanner(reader)

	inDependencySection := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check if we've entered the Dependencies section
		if strings.HasPrefix(trimmedLine, "## Dependencies") {
			inDependencySection = true
			continue
		}

		// Exit Dependencies section if we hit another ## section
		if inDependencySection && strings.HasPrefix(trimmedLine, "## ") {
			break
		}

		// Track subsections (### Full Dependencies, ### Partial Dependencies)
		// These are informational headers, we can skip them
		if inDependencySection && strings.HasPrefix(trimmedLine, "###") {
			continue
		}

		// Skip empty lines and non-list items
		if !inDependencySection || trimmedLine == "" || !strings.HasPrefix(trimmedLine, "-") {
			continue
		}

		// Try to parse as partial dependency first (has colon)
		if matches := partialDependencyPattern.FindStringSubmatch(trimmedLine); matches != nil {
			reqID := matches[1]
			featuresOrAspect := matches[2]
			description := ""
			if len(matches) > 3 {
				description = strings.TrimSpace(matches[3])
			}

			dep := parseDependency(sourceReqID, reqID, featuresOrAspect, description)
			dependencies = append(dependencies, dep)
			continue
		}

		// Try to parse as full dependency (no colon)
		if matches := fullDependencyPattern.FindStringSubmatch(trimmedLine); matches != nil {
			reqID := matches[1]
			description := ""
			if len(matches) > 2 {
				description = strings.TrimSpace(matches[2])
			}

			dependencies = append(dependencies, Dependency{
				Source:      sourceReqID,
				Target:      reqID,
				Type:        DependencyTypeFull,
				Description: description,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading spec file: %w", err)
	}

	return dependencies, nil
}

// parseDependency determines whether a partial dependency is for specific features
// or an entire aspect based on the content after the colon.
//
// Logic:
// - If contains comma: PartialFeatures (e.g., "Feature1,Feature2")
// - If single capitalized word matching known aspect: PartialAspect (e.g., "Storage", "Engine")
// - Otherwise: PartialFeatures with single feature
func parseDependency(sourceReqID, targetReqID, featuresOrAspect, description string) Dependency {
	// Check if it contains multiple features (comma-separated)
	if strings.Contains(featuresOrAspect, ",") {
		features := strings.Split(featuresOrAspect, ",")
		for i := range features {
			features[i] = strings.TrimSpace(features[i])
		}
		return Dependency{
			Source:           sourceReqID,
			Target:           targetReqID,
			Type:             DependencyTypePartialFeatures,
			RequiredFeatures: features,
			Description:      description,
		}
	}

	// Check if it's a known aspect
	featuresOrAspect = strings.TrimSpace(featuresOrAspect)
	if isKnownAspect(featuresOrAspect) {
		return Dependency{
			Source:         sourceReqID,
			Target:         targetReqID,
			Type:           DependencyTypePartialAspect,
			RequiredAspect: featuresOrAspect,
			Description:    description,
		}
	}

	// Default to single feature dependency
	return Dependency{
		Source:           sourceReqID,
		Target:           targetReqID,
		Type:             DependencyTypePartialFeatures,
		RequiredFeatures: []string{featuresOrAspect},
		Description:      description,
	}
}

// isKnownAspect checks if a string matches a known CANARY aspect name.
// Valid aspects from spec: API, CLI, Engine, Storage, Security, Docs, Wire,
// Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist
func isKnownAspect(s string) bool {
	knownAspects := map[string]bool{
		"API":       true,
		"CLI":       true,
		"Engine":    true,
		"Storage":   true,
		"Security":  true,
		"Docs":      true,
		"Wire":      true,
		"Planner":   true,
		"Decode":    true,
		"Encode":    true,
		"RoundTrip": true,
		"Bench":     true,
		"FrontEnd":  true,
		"Dist":      true,
	}
	return knownAspects[s]
}
