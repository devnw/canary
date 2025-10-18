package specs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_Valid; UPDATED=2025-10-18
func TestValidateDependencies_Valid(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.True(t, result.IsValid)
	assert.Len(t, result.Cycles, 0)
	assert.Len(t, result.MissingRequirements, 0)
	assert.Len(t, result.Errors, 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_SimpleCycle; UPDATED=2025-10-18
func TestValidateDependencies_SimpleCycle(t *testing.T) {
	graph := NewDependencyGraph()
	// Create cycle: A -> B -> A
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-101", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-101", Target: "CBIN-100", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	assert.Len(t, result.Cycles, 1)
	assert.Len(t, result.Cycles[0], 3) // CBIN-100 -> CBIN-101 -> CBIN-100

	// Verify cycle contains both requirements
	cycleStr := result.Cycles[0][0] + result.Cycles[0][1] + result.Cycles[0][2]
	assert.Contains(t, cycleStr, "CBIN-100")
	assert.Contains(t, cycleStr, "CBIN-101")
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_ComplexCycle; UPDATED=2025-10-18
func TestValidateDependencies_ComplexCycle(t *testing.T) {
	graph := NewDependencyGraph()
	// Create cycle: A -> B -> C -> A
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-101", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-101", Target: "CBIN-102", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-102", Target: "CBIN-100", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	assert.Len(t, result.Cycles, 1)
	assert.Len(t, result.Cycles[0], 4) // CBIN-100 -> CBIN-101 -> CBIN-102 -> CBIN-100
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_MultipleCycles; UPDATED=2025-10-18
func TestValidateDependencies_MultipleCycles(t *testing.T) {
	graph := NewDependencyGraph()
	// Cycle 1: A -> B -> A
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-101", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-101", Target: "CBIN-100", Type: DependencyTypeFull})
	// Cycle 2: C -> D -> C
	graph.AddDependency(Dependency{Source: "CBIN-200", Target: "CBIN-201", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-201", Target: "CBIN-200", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	// Should detect at least one cycle (may detect both)
	assert.GreaterOrEqual(t, len(result.Cycles), 1)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_SelfCycle; UPDATED=2025-10-18
func TestValidateDependencies_SelfCycle(t *testing.T) {
	graph := NewDependencyGraph()
	// Self-cycle: A -> A
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-100", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	assert.Len(t, result.Cycles, 1)
	assert.Contains(t, result.Cycles[0], "CBIN-100")
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_DAG; UPDATED=2025-10-18
func TestValidateDependencies_DAG(t *testing.T) {
	graph := NewDependencyGraph()
	// Valid DAG:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-101", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-102", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-101", Target: "CBIN-103", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-102", Target: "CBIN-103", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.True(t, result.IsValid)
	assert.Len(t, result.Cycles, 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateDependencies_EmptyGraph; UPDATED=2025-10-18
func TestValidateDependencies_EmptyGraph(t *testing.T) {
	graph := NewDependencyGraph()
	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.True(t, result.IsValid)
	assert.Len(t, result.Cycles, 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidateRequirementExists; UPDATED=2025-10-18
func TestValidateRequirementExists(t *testing.T) {
	// Mock spec finder that knows about CBIN-146 and CBIN-145
	specFinder := &MockSpecFinder{
		existingSpecs: map[string]bool{
			"CBIN-146": true,
			"CBIN-145": true,
			"CBIN-129": true,
		},
	}

	graph := NewDependencyGraph()
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-147", Target: "CBIN-999", Type: DependencyTypeFull}) // Missing

	validator := NewDependencyValidator(graph)
	validator.SetSpecFinder(specFinder)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	assert.Len(t, result.MissingRequirements, 1)
	assert.Equal(t, "CBIN-999", result.MissingRequirements[0])
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestValidationResult_FormatErrors; UPDATED=2025-10-18
func TestValidationResult_FormatErrors(t *testing.T) {
	result := ValidationResult{
		IsValid: false,
		Cycles: [][]string{
			{"CBIN-100", "CBIN-101", "CBIN-100"},
		},
		MissingRequirements: []string{"CBIN-999"},
		Errors: []string{
			"Cycle detected: CBIN-100 -> CBIN-101 -> CBIN-100",
			"Missing requirement: CBIN-999",
		},
	}

	formatted := result.FormatErrors()
	assert.Contains(t, formatted, "Cycle detected")
	assert.Contains(t, formatted, "CBIN-100")
	assert.Contains(t, formatted, "CBIN-101")
	assert.Contains(t, formatted, "Missing requirement")
	assert.Contains(t, formatted, "CBIN-999")
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyValidator"; ASPECT=Engine; STATUS=TESTED; TEST=TestDetectCycles_LongPath; UPDATED=2025-10-18
func TestDetectCycles_LongPath(t *testing.T) {
	graph := NewDependencyGraph()
	// Create long cycle: A -> B -> C -> D -> E -> A
	graph.AddDependency(Dependency{Source: "CBIN-100", Target: "CBIN-101", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-101", Target: "CBIN-102", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-102", Target: "CBIN-103", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-103", Target: "CBIN-104", Type: DependencyTypeFull})
	graph.AddDependency(Dependency{Source: "CBIN-104", Target: "CBIN-100", Type: DependencyTypeFull})

	validator := NewDependencyValidator(graph)
	result := validator.Validate()

	assert.False(t, result.IsValid)
	require.Len(t, result.Cycles, 1)
	assert.Len(t, result.Cycles[0], 6) // 5 nodes + return to start
}

// MockSpecFinder is a test helper that implements spec existence checking
type MockSpecFinder struct {
	existingSpecs map[string]bool
}

func (m *MockSpecFinder) SpecExists(reqID string) bool {
	return m.existingSpecs[reqID]
}

func (m *MockSpecFinder) FindSpecPath(reqID string) (string, error) {
	if m.existingSpecs[reqID] {
		return "/tmp/" + reqID + "/spec.md", nil
	}
	return "", &SpecNotFoundError{ReqID: reqID}
}

// SpecNotFoundError is returned when a spec file cannot be found
type SpecNotFoundError struct {
	ReqID string
}

func (e *SpecNotFoundError) Error() string {
	return "specification not found: " + e.ReqID
}
