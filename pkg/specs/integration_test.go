package specs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CANARY: REQ=CBIN-147; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestIntegration_EndToEnd; UPDATED=2025-10-18
func TestIntegration_EndToEnd(t *testing.T) {
	// Test complete workflow: parse → build graph → validate → check status
	tmpDir := t.TempDir()

	// Create spec files with dependencies
	createTestSpecs(t, tmpDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Step 1: Parse dependencies from all specs
	spec147Path := filepath.Join(".canary", "specs", "CBIN-147-test", "spec.md")
	deps147, err := ParseDependenciesFromFile("CBIN-147", spec147Path)
	require.NoError(t, err)
	require.Len(t, deps147, 2) // CBIN-146 and CBIN-145

	spec146Path := filepath.Join(".canary", "specs", "CBIN-146-test", "spec.md")
	deps146, err := ParseDependenciesFromFile("CBIN-146", spec146Path)
	require.NoError(t, err)
	require.Len(t, deps146, 1) // CBIN-129

	// Step 2: Build dependency graph
	graph := NewDependencyGraph()
	for _, dep := range deps147 {
		graph.AddDependency(dep)
	}
	for _, dep := range deps146 {
		graph.AddDependency(dep)
	}

	// Step 3: Validate graph (should have no cycles)
	validator := NewDependencyValidator(graph)
	result := validator.Validate()
	assert.True(t, result.IsValid)
	assert.Len(t, result.Cycles, 0)

	// Step 4: Check transitive dependencies
	generator := NewGraphGenerator(nil)
	transitive := generator.GetTransitiveDependencies(graph, "CBIN-147")
	assert.Contains(t, transitive, "CBIN-146")
	assert.Contains(t, transitive, "CBIN-145")
	assert.Contains(t, transitive, "CBIN-129") // Transitive through CBIN-146

	// Step 5: Generate visualization
	tree := generator.FormatASCIITree(graph, "CBIN-147")
	assert.Contains(t, tree, "CBIN-147")
	assert.Contains(t, tree, "CBIN-146")
	assert.Contains(t, tree, "├──")
}

// CANARY: REQ=CBIN-147; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestIntegration_PartialDependencies; UPDATED=2025-10-18
func TestIntegration_PartialDependencies(t *testing.T) {
	// Test partial feature and aspect dependencies
	tmpDir := t.TempDir()

	// Create spec with partial dependencies
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-147-partial")
	require.NoError(t, os.MkdirAll(specDir, 0755))

	specContent := `# Test Partial Dependencies

## Dependencies

### Partial Dependencies
- CBIN-146:ProjectRegistry,ContextManagement (need specific features)
- CBIN-129:Storage (only Storage aspect)

## Features
- TestFeature
`
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Parse dependencies
	deps, err := ParseDependenciesFromFile("CBIN-147", filepath.Join(specDir, "spec.md"))
	require.NoError(t, err)
	require.Len(t, deps, 2)

	// Verify partial feature dependency
	partialFeature := deps[0]
	assert.Equal(t, DependencyTypePartialFeatures, partialFeature.Type)
	assert.Len(t, partialFeature.RequiredFeatures, 2)
	assert.Contains(t, partialFeature.RequiredFeatures, "ProjectRegistry")
	assert.Contains(t, partialFeature.RequiredFeatures, "ContextManagement")

	// Verify partial aspect dependency
	partialAspect := deps[1]
	assert.Equal(t, DependencyTypePartialAspect, partialAspect.Type)
	assert.Equal(t, "Storage", partialAspect.RequiredAspect)
}

// CANARY: REQ=CBIN-147; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestIntegration_CircularDetection; UPDATED=2025-10-18
func TestIntegration_CircularDetection(t *testing.T) {
	// Test that circular dependencies are properly detected
	tmpDir := t.TempDir()

	// Create circular dependency: A -> B -> C -> A
	createCircularSpecs(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Build graph with circular dependencies
	graph := NewDependencyGraph()

	specAPath := filepath.Join(".canary", "specs", "CBIN-100-A", "spec.md")
	depsA, err := ParseDependenciesFromFile("CBIN-100", specAPath)
	require.NoError(t, err)
	for _, dep := range depsA {
		graph.AddDependency(dep)
	}

	specBPath := filepath.Join(".canary", "specs", "CBIN-101-B", "spec.md")
	depsB, err := ParseDependenciesFromFile("CBIN-101", specBPath)
	require.NoError(t, err)
	for _, dep := range depsB {
		graph.AddDependency(dep)
	}

	specCPath := filepath.Join(".canary", "specs", "CBIN-102-C", "spec.md")
	depsC, err := ParseDependenciesFromFile("CBIN-102", specCPath)
	require.NoError(t, err)
	for _, dep := range depsC {
		graph.AddDependency(dep)
	}

	// Validate - should detect cycle
	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	assert.GreaterOrEqual(t, len(result.Cycles), 1)

	// Verify cycle contains all three requirements
	cycle := result.Cycles[0]
	cycleStr := ""
	for _, node := range cycle {
		cycleStr += node
	}
	assert.Contains(t, cycleStr, "CBIN-100")
	assert.Contains(t, cycleStr, "CBIN-101")
	assert.Contains(t, cycleStr, "CBIN-102")
}

// CANARY: REQ=CBIN-147; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestIntegration_StatusChecking; UPDATED=2025-10-18
func TestIntegration_StatusChecking(t *testing.T) {
	// Test dependency status checking with mock tokens
	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{
		Source: "CBIN-147",
		Target: "CBIN-146",
		Type:   DependencyTypeFull,
	})
	graph.AddDependency(Dependency{
		Source: "CBIN-147",
		Target: "CBIN-145",
		Type:   DependencyTypeFull,
	})

	// Create mock token provider
	tokenProvider := &MockTokenProvider{
		tokens: map[string][]MockToken{
			"CBIN-146": {
				{Feature: "ProjectRegistry", Aspect: "Storage", Status: "TESTED"},
				{Feature: "ContextManagement", Aspect: "Engine", Status: "TESTED"},
			},
			"CBIN-145": {
				{Feature: "Migration", Aspect: "Engine", Status: "IMPL"}, // Blocking
			},
		},
	}

	// Check dependencies
	checker := NewStatusChecker(tokenProvider)
	deps := graph.GetDependencies("CBIN-147")
	statuses := checker.CheckAllDependencies(deps)

	require.Len(t, statuses, 2)
	assert.True(t, statuses[0].IsSatisfied)  // CBIN-146 satisfied
	assert.False(t, statuses[1].IsSatisfied) // CBIN-145 blocking

	// Get blocking dependencies
	blocking := checker.GetBlockingDependencies(deps)
	require.Len(t, blocking, 1)
	assert.Equal(t, "CBIN-145", blocking[0].Dependency.Target)
}

// CANARY: REQ=CBIN-147; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestIntegration_ComplexGraph; UPDATED=2025-10-18
func TestIntegration_ComplexGraph(t *testing.T) {
	// Test complex dependency graph with multiple levels
	// Graph structure:
	//       147
	//      /   \
	//    146   145
	//    / \   /
	//  129 140
	//    \ /
	//    123

	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-140", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-145", Target: "CBIN-140", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-129", Target: "CBIN-123", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-140", Target: "CBIN-123", Type: DependencyTypeFull})

	// Test graph traversal
	generator := NewGraphGenerator(nil)
	transitive := generator.GetTransitiveDependencies(graph, "CBIN-147")

	// Should find all 5 dependencies (146, 145, 129, 140, 123)
	assert.Len(t, transitive, 5)
	assert.Contains(t, transitive, "CBIN-146")
	assert.Contains(t, transitive, "CBIN-145")
	assert.Contains(t, transitive, "CBIN-129")
	assert.Contains(t, transitive, "CBIN-140")
	assert.Contains(t, transitive, "CBIN-123")

	// Test depth calculation
	depth := generator.GetDependencyDepth(graph, "CBIN-147")
	assert.Equal(t, 3, depth) // 147 -> 146 -> 129 -> 123

	// Test reverse dependencies
	reverseDeps := graph.GetReverseDependencies("CBIN-123")
	assert.Len(t, reverseDeps, 2) // CBIN-129 and CBIN-140

	reverseDeps140 := graph.GetReverseDependencies("CBIN-140")
	assert.Len(t, reverseDeps140, 2) // CBIN-146 and CBIN-145
}

// CANARY: REQ=CBIN-147; FEATURE="IntegrationTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestIntegration_DiamondDependency; UPDATED=2025-10-18
func TestIntegration_DiamondDependency(t *testing.T) {
	// Test diamond pattern: A -> B, A -> C, B -> D, C -> D
	// D should appear only once in transitive dependencies

	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-145", Target: "CBIN-129", Type: DependencyTypeFull})

	generator := NewGraphGenerator(nil)
	transitive := generator.GetTransitiveDependencies(graph, "CBIN-147")

	// Should have 3 unique dependencies (no duplicates)
	assert.Len(t, transitive, 3)
	assert.Contains(t, transitive, "CBIN-146")
	assert.Contains(t, transitive, "CBIN-145")
	assert.Contains(t, transitive, "CBIN-129")
}

// Helper functions

func createTestSpecs(t *testing.T, baseDir string) {
	t.Helper()

	// CBIN-147 depends on CBIN-146 and CBIN-145
	spec147Dir := filepath.Join(baseDir, ".canary", "specs", "CBIN-147-test")
	require.NoError(t, os.MkdirAll(spec147Dir, 0755))
	spec147Content := `# Test Spec 147

## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support)
- CBIN-145 (Legacy Migration)

## Features
- Feature147
`
	require.NoError(t, os.WriteFile(filepath.Join(spec147Dir, "spec.md"), []byte(spec147Content), 0644))

	// CBIN-146 depends on CBIN-129
	spec146Dir := filepath.Join(baseDir, ".canary", "specs", "CBIN-146-test")
	require.NoError(t, os.MkdirAll(spec146Dir, 0755))
	spec146Content := `# Test Spec 146

## Dependencies

### Full Dependencies
- CBIN-129 (Database Migrations)

## Features
- Feature146
`
	require.NoError(t, os.WriteFile(filepath.Join(spec146Dir, "spec.md"), []byte(spec146Content), 0644))

	// CBIN-145 has no dependencies
	spec145Dir := filepath.Join(baseDir, ".canary", "specs", "CBIN-145-test")
	require.NoError(t, os.MkdirAll(spec145Dir, 0755))
	spec145Content := `# Test Spec 145

## Features
- Feature145
`
	require.NoError(t, os.WriteFile(filepath.Join(spec145Dir, "spec.md"), []byte(spec145Content), 0644))

	// CBIN-129 has no dependencies
	spec129Dir := filepath.Join(baseDir, ".canary", "specs", "CBIN-129-test")
	require.NoError(t, os.MkdirAll(spec129Dir, 0755))
	spec129Content := `# Test Spec 129

## Features
- Feature129
`
	require.NoError(t, os.WriteFile(filepath.Join(spec129Dir, "spec.md"), []byte(spec129Content), 0644))
}

func createCircularSpecs(t *testing.T, baseDir string) {
	t.Helper()

	// CBIN-100 -> CBIN-101
	specADir := filepath.Join(baseDir, ".canary", "specs", "CBIN-100-A")
	require.NoError(t, os.MkdirAll(specADir, 0755))
	specAContent := `# Test Spec A

## Dependencies
- CBIN-101 (Depends on B)

## Features
- FeatureA
`
	require.NoError(t, os.WriteFile(filepath.Join(specADir, "spec.md"), []byte(specAContent), 0644))

	// CBIN-101 -> CBIN-102
	specBDir := filepath.Join(baseDir, ".canary", "specs", "CBIN-101-B")
	require.NoError(t, os.MkdirAll(specBDir, 0755))
	specBContent := `# Test Spec B

## Dependencies
- CBIN-102 (Depends on C)

## Features
- FeatureB
`
	require.NoError(t, os.WriteFile(filepath.Join(specBDir, "spec.md"), []byte(specBContent), 0644))

	// CBIN-102 -> CBIN-100 (completes cycle)
	specCDir := filepath.Join(baseDir, ".canary", "specs", "CBIN-102-C")
	require.NoError(t, os.MkdirAll(specCDir, 0755))
	specCContent := `# Test Spec C

## Dependencies
- CBIN-100 (Depends on A - creates cycle)

## Features
- FeatureC
`
	require.NoError(t, os.WriteFile(filepath.Join(specCDir, "spec.md"), []byte(specCContent), 0644))
}
