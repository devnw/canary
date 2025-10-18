package specs

import (
	"fmt"
	"strings"
)

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestBuildGraphFromSpecs,TestGetTransitiveDependencies,TestFormatASCIITree; UPDATED=2025-10-18

// SpecLoader is an interface for loading dependencies from spec files.
type SpecLoader interface {
	// LoadDependencies loads all dependencies for a given requirement ID
	LoadDependencies(reqID string) ([]Dependency, error)
}

// StatusChecker interface for checking dependency satisfaction
type StatusCheckerInterface interface {
	// IsDependencySatisfied checks if a dependency is satisfied
	IsDependencySatisfied(dep Dependency) bool
}

// GraphGenerator builds and visualizes dependency graphs.
type GraphGenerator struct {
	specLoader    SpecLoader
	statusChecker StatusCheckerInterface
}

// NewGraphGenerator creates a new graph generator.
func NewGraphGenerator(loader SpecLoader) *GraphGenerator {
	return &GraphGenerator{
		specLoader:    loader,
		statusChecker: nil, // Optional
	}
}

// SetStatusChecker configures the generator to show dependency status in visualizations.
func (gg *GraphGenerator) SetStatusChecker(checker StatusCheckerInterface) {
	gg.statusChecker = checker
}

// BuildGraph builds a complete dependency graph from a list of requirement IDs.
// It loads dependencies for each requirement and constructs the full graph.
func (gg *GraphGenerator) BuildGraph(reqIDs []string) (*DependencyGraph, error) {
	graph := NewDependencyGraph()

	if gg.specLoader == nil {
		return graph, fmt.Errorf("spec loader not configured")
	}

	// Load dependencies for each requirement
	for _, reqID := range reqIDs {
		deps, err := gg.specLoader.LoadDependencies(reqID)
		if err != nil {
			// Skip requirements that can't be loaded (may not have deps)
			continue
		}

		for _, dep := range deps {
			graph.AddDependency(dep)
		}
	}

	return graph, nil
}

// GetTransitiveDependencies returns all transitive dependencies of a requirement.
// Uses BFS to traverse the dependency graph and collect all reachable requirements.
func (gg *GraphGenerator) GetTransitiveDependencies(graph *DependencyGraph, reqID string) []string {
	visited := make(map[string]bool)
	resultSet := make(map[string]bool) // Use set to avoid duplicates
	queue := []string{reqID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		// Get direct dependencies
		deps := graph.GetDependencies(current)
		for _, dep := range deps {
			// Add to result set (avoids duplicates)
			resultSet[dep.Target] = true

			if !visited[dep.Target] {
				queue = append(queue, dep.Target)
			}
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(resultSet))
	for reqID := range resultSet {
		result = append(result, reqID)
	}

	return result
}

// GetDependencyDepth returns the maximum depth of the dependency tree.
// Depth is the longest path from the root to any leaf node.
func (gg *GraphGenerator) GetDependencyDepth(graph *DependencyGraph, reqID string) int {
	visited := make(map[string]bool)
	return gg.getDependencyDepthRecursive(graph, reqID, visited)
}

func (gg *GraphGenerator) getDependencyDepthRecursive(graph *DependencyGraph, reqID string, visited map[string]bool) int {
	if visited[reqID] {
		return 0 // Avoid cycles
	}
	visited[reqID] = true

	deps := graph.GetDependencies(reqID)
	if len(deps) == 0 {
		return 0
	}

	maxDepth := 0
	for _, dep := range deps {
		depth := gg.getDependencyDepthRecursive(graph, dep.Target, visited)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	delete(visited, reqID) // Allow revisiting in other branches
	return maxDepth + 1
}

// FormatASCIITree generates an ASCII tree visualization of the dependency graph.
// Shows the structure with Unicode box-drawing characters and optional status indicators.
//
// Example output:
// CBIN-147 (Specification Dependencies)
// ├── CBIN-146 (Multi-Project Support) ✅
// │   └── CBIN-129 (Migrations) ❌
// └── CBIN-145 (Legacy Migration) ✅
func (gg *GraphGenerator) FormatASCIITree(graph *DependencyGraph, rootReqID string) string {
	deps := graph.GetDependencies(rootReqID)
	if len(deps) == 0 {
		return fmt.Sprintf("%s\n  No dependencies", rootReqID)
	}

	var lines []string
	lines = append(lines, rootReqID)

	visited := make(map[string]bool)
	gg.formatTreeRecursive(graph, deps, "", true, &lines, visited)

	return strings.Join(lines, "\n")
}

func (gg *GraphGenerator) formatTreeRecursive(graph *DependencyGraph, deps []Dependency, prefix string, isLast bool, lines *[]string, visited map[string]bool) {
	for i, dep := range deps {
		isLastDep := i == len(deps)-1

		// Build the tree characters
		var connector, childPrefix string
		if isLastDep {
			connector = "└── "
			childPrefix = prefix + "    "
		} else {
			connector = "├── "
			childPrefix = prefix + "│   "
		}

		// Format the dependency line
		line := prefix + connector + dep.Target

		// Add type annotation if not full
		if dep.Type == DependencyTypePartialFeatures {
			line += fmt.Sprintf(":%s", strings.Join(dep.RequiredFeatures, ","))
		} else if dep.Type == DependencyTypePartialAspect {
			line += fmt.Sprintf(":%s", dep.RequiredAspect)
		}

		// Add status indicator if checker is configured
		if gg.statusChecker != nil {
			if gg.statusChecker.IsDependencySatisfied(dep) {
				line += " ✅"
			} else {
				line += " ❌"
			}
		}

		*lines = append(*lines, line)

		// Recursively process children (avoid cycles)
		if !visited[dep.Target] {
			visited[dep.Target] = true
			childDeps := graph.GetDependencies(dep.Target)
			if len(childDeps) > 0 {
				gg.formatTreeRecursive(graph, childDeps, childPrefix, isLastDep, lines, visited)
			}
		}
	}
}

// FormatDependencyChain formats a list of requirement IDs as a chain.
// Example: "CBIN-147 → CBIN-146 → CBIN-129"
func (gg *GraphGenerator) FormatDependencyChain(reqIDs []string) string {
	return strings.Join(reqIDs, " → ")
}

// FormatCompactList formats dependencies as a compact comma-separated list.
// Example: "CBIN-146, CBIN-145, CBIN-129"
func (gg *GraphGenerator) FormatCompactList(graph *DependencyGraph, reqID string) string {
	deps := graph.GetDependencies(reqID)
	if len(deps) == 0 {
		return "none"
	}

	targets := make([]string, len(deps))
	for i, dep := range deps {
		targets[i] = dep.Target
	}

	return strings.Join(targets, ", ")
}

// FormatDependencySummary generates a multi-line summary of dependencies.
// Includes direct dependencies, transitive count, and depth.
func (gg *GraphGenerator) FormatDependencySummary(graph *DependencyGraph, reqID string) string {
	directDeps := graph.GetDependencies(reqID)
	transitiveDeps := gg.GetTransitiveDependencies(graph, reqID)
	depth := gg.GetDependencyDepth(graph, reqID)

	var lines []string
	lines = append(lines, fmt.Sprintf("Dependency Summary for %s:", reqID))
	lines = append(lines, fmt.Sprintf("  Direct dependencies: %d", len(directDeps)))
	lines = append(lines, fmt.Sprintf("  Transitive dependencies: %d", len(transitiveDeps)))
	lines = append(lines, fmt.Sprintf("  Maximum depth: %d", depth))

	if len(directDeps) > 0 {
		lines = append(lines, "\nDirect dependencies:")
		for _, dep := range directDeps {
			line := fmt.Sprintf("  - %s", dep.Target)
			if dep.Type == DependencyTypePartialFeatures {
				line += fmt.Sprintf(" (features: %s)", strings.Join(dep.RequiredFeatures, ", "))
			} else if dep.Type == DependencyTypePartialAspect {
				line += fmt.Sprintf(" (aspect: %s)", dep.RequiredAspect)
			}
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}
