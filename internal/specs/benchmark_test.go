package specs

import (
	"fmt"
	"testing"
)

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkDependencyParsing; UPDATED=2025-10-18

// BenchmarkDependencyParsing benchmarks parsing dependencies from spec files
func BenchmarkDependencyParsing(b *testing.B) {
	specContent := `# Test Spec

## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support)
- CBIN-145 (Legacy Migration)
- CBIN-129 (Database Migrations)

### Partial Dependencies
- CBIN-140:GapRepository,GapService (gap storage)
- CBIN-133:Engine (only Engine aspect)

## Features
- Feature1
- Feature2
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deps, err := ParseDependencies("CBIN-147", stringReader(specContent))
		if err != nil {
			b.Fatal(err)
		}
		if len(deps) != 5 {
			b.Fatalf("expected 5 dependencies, got %d", len(deps))
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkCircularDetection_100Nodes; UPDATED=2025-10-18

// BenchmarkCircularDetection_100Nodes benchmarks cycle detection with 100 nodes
func BenchmarkCircularDetection_100Nodes(b *testing.B) {
	graph := createLinearGraph(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator := NewDependencyValidator(graph)
		result := validator.Validate()
		if !result.IsValid {
			b.Fatal("expected valid graph")
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkCircularDetection_500Nodes; UPDATED=2025-10-18

// BenchmarkCircularDetection_500Nodes benchmarks cycle detection with 500 nodes
// Target: < 200ms (from spec)
func BenchmarkCircularDetection_500Nodes(b *testing.B) {
	graph := createLinearGraph(500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator := NewDependencyValidator(graph)
		result := validator.Validate()
		if !result.IsValid {
			b.Fatal("expected valid graph")
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkCircularDetection_WithCycle; UPDATED=2025-10-18

// BenchmarkCircularDetection_WithCycle benchmarks cycle detection when cycle exists
func BenchmarkCircularDetection_WithCycle(b *testing.B) {
	graph := createLinearGraph(100)
	// Add cycle: CBIN-099 -> CBIN-000 (close the loop)
	graph.AddDependency(Dependency{
		Source: "CBIN-099",
		Target: "CBIN-000",
		Type:   DependencyTypeFull,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator := NewDependencyValidator(graph)
		result := validator.Validate()
		if result.IsValid {
			b.Fatal("expected cycle to be detected")
		}
		if len(result.Cycles) == 0 {
			b.Fatal("expected at least one cycle")
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkTransitiveDependencies_Linear; UPDATED=2025-10-18

// BenchmarkTransitiveDependencies_Linear benchmarks transitive dependency resolution
func BenchmarkTransitiveDependencies_Linear(b *testing.B) {
	graph := createLinearGraph(100)
	generator := NewGraphGenerator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transitive := generator.GetTransitiveDependencies(graph, "CBIN-000")
		if len(transitive) != 99 {
			b.Fatalf("expected 99 transitive deps, got %d", len(transitive))
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkTransitiveDependencies_Diamond; UPDATED=2025-10-18

// BenchmarkTransitiveDependencies_Diamond benchmarks with diamond patterns
func BenchmarkTransitiveDependencies_Diamond(b *testing.B) {
	graph := createDiamondGraph(50) // Creates 50 diamond layers
	generator := NewGraphGenerator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transitive := generator.GetTransitiveDependencies(graph, "CBIN-000")
		// Should handle duplicates efficiently
		_ = transitive
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkStatusChecking_FullDependency; UPDATED=2025-10-18

// BenchmarkStatusChecking_FullDependency benchmarks checking full dependency status
func BenchmarkStatusChecking_FullDependency(b *testing.B) {
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "Feature1", Aspect: "API", Status: "TESTED"},
				{Feature: "Feature2", Aspect: "CLI", Status: "TESTED"},
				{Feature: "Feature3", Aspect: "Engine", Status: "TESTED"},
				{Feature: "Feature4", Aspect: "Storage", Status: "TESTED"},
				{Feature: "Feature5", Aspect: "Security", Status: "TESTED"},
			},
		},
	}

	checker := NewStatusChecker(tokenProvider)
	dep := Dependency{
		Source: "CBIN-147",
		Target: "CBIN-146",
		Type:   DependencyTypeFull,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		status := checker.CheckDependency(dep)
		if !status.IsSatisfied {
			b.Fatal("expected dependency to be satisfied")
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkStatusChecking_Batch; UPDATED=2025-10-18

// BenchmarkStatusChecking_Batch benchmarks checking multiple dependencies
func BenchmarkStatusChecking_Batch(b *testing.B) {
	// Create 100 dependencies
	deps := make([]Dependency, 100)
	tokens := make(map[string][]MockToken)

	for i := 0; i < 100; i++ {
		reqID := fmt.Sprintf("CBIN-%03d", i)
		deps[i] = Dependency{
			Source: "CBIN-147",
			Target: reqID,
			Type:   DependencyTypeFull,
		}
		tokens[reqID] = []MockToken{
			{Feature: "Feature1", Aspect: "API", Status: "TESTED"},
		}
	}

	tokenProvider := &MockTokenProvider{tokens: tokens}
	checker := NewStatusChecker(tokenProvider)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		statuses := checker.CheckAllDependencies(deps)
		if len(statuses) != 100 {
			b.Fatalf("expected 100 statuses, got %d", len(statuses))
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkGraphVisualization; UPDATED=2025-10-18

// BenchmarkGraphVisualization benchmarks ASCII tree generation
func BenchmarkGraphVisualization(b *testing.B) {
	graph := createBinaryTree(7) // Creates tree with depth 7
	generator := NewGraphGenerator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := generator.FormatASCIITree(graph, "CBIN-000")
		if len(tree) == 0 {
			b.Fatal("expected non-empty tree")
		}
	}
}

// CANARY: REQ=CBIN-147; FEATURE="PerformanceBenchmarks"; ASPECT=Engine; STATUS=BENCHED; BENCH=BenchmarkDependencyDepth; UPDATED=2025-10-18

// BenchmarkDependencyDepth benchmarks depth calculation
func BenchmarkDependencyDepth(b *testing.B) {
	graph := createLinearGraph(100)
	generator := NewGraphGenerator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		depth := generator.GetDependencyDepth(graph, "CBIN-000")
		if depth != 99 {
			b.Fatalf("expected depth 99, got %d", depth)
		}
	}
}

// Helper functions for benchmarks

// createLinearGraph creates a linear dependency chain: 0 -> 1 -> 2 -> ... -> n-1
func createLinearGraph(n int) *DependencyGraph {
	graph := NewDependencyGraph()
	for i := 0; i < n-1; i++ {
		source := fmt.Sprintf("CBIN-%03d", i)
		target := fmt.Sprintf("CBIN-%03d", i+1)
		graph.AddDependency(Dependency{
			Source: source,
			Target: target,
			Type:   DependencyTypeFull,
		})
	}
	return graph
}

// createDiamondGraph creates multiple diamond patterns
func createDiamondGraph(layers int) *DependencyGraph {
	graph := NewDependencyGraph()
	for i := 0; i < layers; i++ {
		top := fmt.Sprintf("CBIN-%03d", i*3)
		left := fmt.Sprintf("CBIN-%03d", i*3+1)
		right := fmt.Sprintf("CBIN-%03d", i*3+2)
		bottom := fmt.Sprintf("CBIN-%03d", (i+1)*3)

		// Top -> Left, Top -> Right
		graph.AddDependency(Dependency{Source: top, Target: left, Type: DependencyTypeFull})
		graph.AddDependency(Dependency{Source: top, Target: right, Type: DependencyTypeFull})

		// Left -> Bottom, Right -> Bottom (creates diamond)
		graph.AddDependency(Dependency{Source: left, Target: bottom, Type: DependencyTypeFull})
		graph.AddDependency(Dependency{Source: right, Target: bottom, Type: DependencyTypeFull})
	}
	return graph
}

// createBinaryTree creates a balanced binary tree
func createBinaryTree(depth int) *DependencyGraph {
	graph := NewDependencyGraph()
	nodeCount := 0

	var build func(int, string)
	build = func(currentDepth int, parent string) {
		if currentDepth >= depth {
			return
		}

		left := fmt.Sprintf("CBIN-%03d", nodeCount)
		nodeCount++
		right := fmt.Sprintf("CBIN-%03d", nodeCount)
		nodeCount++

		graph.AddDependency(Dependency{Source: parent, Target: left, Type: DependencyTypeFull})
		graph.AddDependency(Dependency{Source: parent, Target: right, Type: DependencyTypeFull})

		build(currentDepth+1, left)
		build(currentDepth+1, right)
	}

	root := "CBIN-000"
	nodeCount++
	build(1, root)

	return graph
}

// stringReader wraps a string as an io.Reader
type stringReader string

func (s stringReader) Read(p []byte) (n int, err error) {
	return copy(p, []byte(s)), nil
}
