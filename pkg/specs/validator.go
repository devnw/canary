package specs

import (
	"fmt"
	"strings"
)

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_Valid,TestValidateDependencies_SimpleCycle,TestValidateDependencies_ComplexCycle; UPDATED=2025-10-18

// SpecFinder is an interface for finding and checking specification existence.
// This allows validation against the actual filesystem or a mock for testing.
type SpecFinder interface {
	// SpecExists checks if a specification exists for the given requirement ID
	SpecExists(reqID string) bool

	// FindSpecPath returns the path to the spec.md file for a requirement
	FindSpecPath(reqID string) (string, error)
}

// ValidationResult contains the results of dependency validation.
type ValidationResult struct {
	// IsValid is true if the dependency graph is valid (no cycles, all requirements exist)
	IsValid bool

	// Cycles contains all detected circular dependencies.
	// Each cycle is represented as a slice of requirement IDs forming the cycle.
	// Example: ["CBIN-100", "CBIN-101", "CBIN-102", "CBIN-100"]
	Cycles [][]string

	// MissingRequirements lists requirement IDs that are referenced but don't exist
	MissingRequirements []string

	// Errors contains human-readable error messages
	Errors []string
}

// FormatErrors returns a formatted string containing all validation errors.
func (vr *ValidationResult) FormatErrors() string {
	if vr.IsValid {
		return "No validation errors"
	}

	var parts []string

	if len(vr.Cycles) > 0 {
		parts = append(parts, "Circular dependencies detected:")
		for i, cycle := range vr.Cycles {
			parts = append(parts, fmt.Sprintf("  Cycle %d: %s", i+1, strings.Join(cycle, " -> ")))
		}
	}

	if len(vr.MissingRequirements) > 0 {
		parts = append(parts, "Missing requirements:")
		for _, reqID := range vr.MissingRequirements {
			parts = append(parts, fmt.Sprintf("  - %s", reqID))
		}
	}

	if len(vr.Errors) > 0 {
		parts = append(parts, "Other errors:")
		for _, err := range vr.Errors {
			parts = append(parts, fmt.Sprintf("  - %s", err))
		}
	}

	return strings.Join(parts, "\n")
}

// DependencyValidator validates dependency graphs for cycles and missing requirements.
type DependencyValidator struct {
	graph      *DependencyGraph
	specFinder SpecFinder
}

// NewDependencyValidator creates a new dependency validator for the given graph.
func NewDependencyValidator(graph *DependencyGraph) *DependencyValidator {
	return &DependencyValidator{
		graph:      graph,
		specFinder: nil, // Optional, can be set with SetSpecFinder
	}
}

// SetSpecFinder configures the validator to check for missing requirements.
func (dv *DependencyValidator) SetSpecFinder(finder SpecFinder) {
	dv.specFinder = finder
}

// Validate performs comprehensive validation of the dependency graph.
// It checks for:
// 1. Circular dependencies (using DFS with recursion stack)
// 2. Missing requirements (if SpecFinder is configured)
func (dv *DependencyValidator) Validate() ValidationResult {
	result := ValidationResult{
		IsValid:             true,
		Cycles:              [][]string{},
		MissingRequirements: []string{},
		Errors:              []string{},
	}

	// Detect circular dependencies
	cycles := dv.detectCycles()
	if len(cycles) > 0 {
		result.IsValid = false
		result.Cycles = cycles
		for _, cycle := range cycles {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Cycle detected: %s", strings.Join(cycle, " -> ")))
		}
	}

	// Check for missing requirements if SpecFinder is configured
	if dv.specFinder != nil {
		missing := dv.checkMissingRequirements()
		if len(missing) > 0 {
			result.IsValid = false
			result.MissingRequirements = missing
			for _, reqID := range missing {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Missing requirement: %s", reqID))
			}
		}
	}

	return result
}

// detectCycles uses Depth-First Search (DFS) with a recursion stack to detect cycles.
// Algorithm:
// 1. For each unvisited node, start a DFS
// 2. Track visited nodes and recursion stack
// 3. If we visit a node already in the recursion stack, we found a cycle
// 4. Extract the cycle path from the recursion stack
//
// Time Complexity: O(V + E) where V = requirements, E = dependencies
func (dv *DependencyValidator) detectCycles() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	// DFS function that detects back edges (cycles)
	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		// Explore all dependencies of this node
		for _, dep := range dv.graph.GetDependencies(node) {
			neighbor := dep.Target

			if !visited[neighbor] {
				// Continue DFS on unvisited neighbor
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Found a back edge (cycle)!
				// Extract the cycle from path
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					// Build cycle: path from cycleStart to end + neighbor
					cycle := make([]string, len(path)-cycleStart+1)
					copy(cycle, path[cycleStart:])
					cycle[len(cycle)-1] = neighbor // Close the cycle
					cycles = append(cycles, cycle)
				}
				return true
			}
		}

		// Backtrack
		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	// Run DFS from each unvisited node
	for _, node := range dv.graph.GetAllRequirements() {
		if !visited[node] {
			// Only process nodes that have outgoing edges
			if len(dv.graph.GetDependencies(node)) > 0 {
				dfs(node)
			}
		}
	}

	return cycles
}

// checkMissingRequirements verifies that all target requirements exist.
// Returns a list of requirement IDs that are referenced but don't have specs.
func (dv *DependencyValidator) checkMissingRequirements() []string {
	var missing []string
	checked := make(map[string]bool)

	// Check all targets in the graph
	for _, deps := range dv.graph.Nodes {
		for _, dep := range deps {
			target := dep.Target

			// Skip if already checked
			if checked[target] {
				continue
			}
			checked[target] = true

			// Check if spec exists
			if !dv.specFinder.SpecExists(target) {
				missing = append(missing, target)
			}
		}
	}

	return missing
}
