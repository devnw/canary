package specs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestBuildGraphFromSpecs; UPDATED=2026-08-29
func TestBuildGraphFromSpecs(t *testing.T) {
	// Mock spec loader that returns dependencies
	specLoader := &MockSpecLoader{
		dependencies: map[string][]Dependency{
			"CBIN-147": {
				{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull},
				{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull},
			},
			"CBIN-146": {
				{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull},
			},
		},
	}

	generator := NewGraphGenerator(specLoader)
	graph, err := generator.BuildGraph([]string{"CBIN-147", "CBIN-146", "CBIN-145", "CBIN-129"})

	require.NoError(t, err)
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.GetDependencies("CBIN-147"), 2)
	assert.Len(t, graph.GetDependencies("CBIN-146"), 1)
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestGetTransitiveDependencies; UPDATED=2026-08-29
func TestGetTransitiveDependencies(t *testing.T) {
	graph := NewDependencyGraph()
	// CBIN-147 -> CBIN-146 -> CBIN-129
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})

	generator := NewGraphGenerator(nil)
	transitive := generator.GetTransitiveDependencies(graph, "CBIN-147")

	// Should include both CBIN-146 (direct) and CBIN-129 (transitive)
	assert.Len(t, transitive, 2)
	assert.Contains(t, transitive, "CBIN-146")
	assert.Contains(t, transitive, "CBIN-129")
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestGetTransitiveDependencies_NoDuplicates; UPDATED=2026-08-29
func TestGetTransitiveDependencies_NoDuplicates(t *testing.T) {
	graph := NewDependencyGraph()
	// Diamond dependency:
	//     147
	//    /   \
	//  146   145
	//    \   /
	//     129
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-145", Target: "CBIN-129", Type: DependencyTypeFull})

	generator := NewGraphGenerator(nil)
	transitive := generator.GetTransitiveDependencies(graph, "CBIN-147")

	// Should include 146, 145, and 129 (only once)
	assert.Len(t, transitive, 3)
	assert.Contains(t, transitive, "CBIN-146")
	assert.Contains(t, transitive, "CBIN-145")
	assert.Contains(t, transitive, "CBIN-129")
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestFormatASCIITree; UPDATED=2026-08-29
func TestFormatASCIITree(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})

	// Mock status checker
	statusChecker := &MockStatusChecker{
		statuses: map[string]bool{
			"CBIN-146": true,
			"CBIN-145": true,
			"CBIN-129": false, // Blocking
		},
	}

	generator := NewGraphGenerator(nil)
	generator.SetStatusChecker(statusChecker)

	tree := generator.FormatASCIITree(graph, "CBIN-147")

	// Verify tree structure contains expected elements
	assert.Contains(t, tree, "CBIN-147")
	assert.Contains(t, tree, "CBIN-146")
	assert.Contains(t, tree, "CBIN-145")
	assert.Contains(t, tree, "CBIN-129")

	// Verify tree characters
	assert.Contains(t, tree, "├──")
	assert.Contains(t, tree, "└──")
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestFormatASCIITree_WithStatus; UPDATED=2026-08-29
func TestFormatASCIITree_WithStatus(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})

	statusChecker := &MockStatusChecker{
		statuses: map[string]bool{
			"CBIN-146": true, // Satisfied
		},
	}

	generator := NewGraphGenerator(nil)
	generator.SetStatusChecker(statusChecker)

	tree := generator.FormatASCIITree(graph, "CBIN-147")

	// Should include status indicator
	assert.Contains(t, tree, "✅") // Satisfied
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestFormatASCIITree_Blocking; UPDATED=2026-08-29
func TestFormatASCIITree_Blocking(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})

	statusChecker := &MockStatusChecker{
		statuses: map[string]bool{
			"CBIN-146": false, // Blocking
		},
	}

	generator := NewGraphGenerator(nil)
	generator.SetStatusChecker(statusChecker)

	tree := generator.FormatASCIITree(graph, "CBIN-147")

	// Should include blocking indicator
	assert.Contains(t, tree, "❌") // Blocking
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestFormatASCIITree_EmptyGraph; UPDATED=2026-08-29
func TestFormatASCIITree_EmptyGraph(t *testing.T) {
	graph := NewDependencyGraph()

	generator := NewGraphGenerator(nil)
	tree := generator.FormatASCIITree(graph, "CBIN-147")

	// Should show the root node with no dependencies
	assert.Contains(t, tree, "CBIN-147")
	assert.Contains(t, tree, "No dependencies")
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestFormatDependencyChain; UPDATED=2026-08-29
func TestFormatDependencyChain(t *testing.T) {
	deps := []string{"CBIN-147", "CBIN-146", "CBIN-129"}

	generator := NewGraphGenerator(nil)
	chain := generator.FormatDependencyChain(deps)

	assert.Equal(t, "CBIN-147 → CBIN-146 → CBIN-129", chain)
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestGetDependencyDepth; UPDATED=2026-08-29
func TestGetDependencyDepth(t *testing.T) {
	graph := NewDependencyGraph()
	// CBIN-147 -> CBIN-146 -> CBIN-129 (depth 2)
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})

	generator := NewGraphGenerator(nil)
	depth := generator.GetDependencyDepth(graph, "CBIN-147")

	assert.Equal(t, 2, depth) // Two levels deep
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestGetDependencyDepth_NoDepencies; UPDATED=2026-08-29
func TestGetDependencyDepth_NoDependencies(t *testing.T) {
	graph := NewDependencyGraph()

	generator := NewGraphGenerator(nil)
	depth := generator.GetDependencyDepth(graph, "CBIN-147")

	assert.Equal(t, 0, depth) // No dependencies
}

// CANARY: REQ=CBIN-147; FEATURE="GraphGenerator"; ASPECT=Engine; STATUS=TESTED; TEST=TestFormatCompactList; UPDATED=2026-08-29
func TestFormatCompactList(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull})

	generator := NewGraphGenerator(nil)
	list := generator.FormatCompactList(graph, "CBIN-147")

	// Should be a compact comma-separated list
	assert.True(t, strings.Contains(list, "CBIN-146") && strings.Contains(list, "CBIN-145"))
}

// MockSpecLoader simulates loading dependencies from spec files
type MockSpecLoader struct {
	dependencies map[string][]Dependency
}

func (m *MockSpecLoader) LoadDependencies(reqID string) ([]Dependency, error) {
	if deps, exists := m.dependencies[reqID]; exists {
		return deps, nil
	}
	return []Dependency{}, nil
}

// MockStatusChecker simulates checking dependency status
type MockStatusChecker struct {
	statuses map[string]bool
}

func (m *MockStatusChecker) IsDependencySatisfied(dep Dependency) bool {
	return m.statuses[dep.Target]
}

// CANARY: REQ=CBIN-203; FEATURE="MermaidGraph"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_203_FormatMermaid; UPDATED=2026-08-28
func TestCANARY_CBIN_203_FormatMermaid(t *testing.T) {
	// Build a two-level graph: CBIN-300 -> CBIN-200 -> CBIN-100, CBIN-300 -> PLAT-4521
	graph := &DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-300": {
				{Source: "CBIN-300", Target: "CBIN-200", Type: DependencyTypeFull, Description: "storage"},
				{Source: "CBIN-300", Target: "PLAT-4521", Type: DependencyTypeFull, Description: "upstream"},
			},
			"CBIN-200": {
				{Source: "CBIN-200", Target: "CBIN-100", Type: DependencyTypeFull},
			},
		},
	}
	gg := NewGraphGenerator(nil)
	urlFor := func(id string) string {
		if id == "PLAT-4521" {
			return "https://company.atlassian.net/browse/PLAT-4521"
		}
		return ""
	}
	out := gg.FormatMermaid(graph, "CBIN-300", urlFor)

	for _, want := range []string{
		"flowchart TD",
		`CBIN_300["CBIN-300"]`,
		"CBIN_300 --> CBIN_200",
		"CBIN_300 --> PLAT_4521",
		"CBIN_200 --> CBIN_100",
		`click PLAT_4521 "https://company.atlassian.net/browse/PLAT-4521"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, `click CBIN_300`) {
		t.Error("no click line should be emitted for empty URLs")
	}
}

// CANARY: REQ=CBIN-203; FEATURE="MermaidGraph"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_203_FormatMermaid_CycleSafe; UPDATED=2026-08-28
func TestCANARY_CBIN_203_FormatMermaid_CycleSafe(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-100": {{Source: "CBIN-100", Target: "CBIN-200", Type: DependencyTypeFull}},
			"CBIN-200": {{Source: "CBIN-200", Target: "CBIN-100", Type: DependencyTypeFull}},
		},
	}
	gg := NewGraphGenerator(nil)
	out := gg.FormatMermaid(graph, "CBIN-100", nil) // must terminate
	if c := strings.Count(out, "CBIN_100 --> CBIN_200"); c != 1 {
		t.Errorf("edge emitted %d times, want 1", c)
	}
}
