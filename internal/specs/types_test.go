package specs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyCreation; UPDATED=2025-10-18
func TestDependencyCreation(t *testing.T) {
	dep := Dependency{
		Source:      "CBIN-147",
		Target:      "CBIN-146",
		Type:        DependencyTypeFull,
		Description: "Requires multi-project support",
	}
	assert.Equal(t, "CBIN-147", dep.Source)
	assert.Equal(t, "CBIN-146", dep.Target)
	assert.Equal(t, DependencyTypeFull, dep.Type)
	assert.Equal(t, "Requires multi-project support", dep.Description)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestPartialFeatureDependency; UPDATED=2025-10-18
func TestPartialFeatureDependency(t *testing.T) {
	dep := Dependency{
		Source:           "CBIN-147",
		Target:           "CBIN-146",
		Type:             DependencyTypePartialFeatures,
		RequiredFeatures: []string{"ProjectRegistry", "ContextManagement"},
		Description:      "Only needs project registry and context management",
	}
	assert.Equal(t, "CBIN-147", dep.Source)
	assert.Equal(t, "CBIN-146", dep.Target)
	assert.Equal(t, DependencyTypePartialFeatures, dep.Type)
	assert.Len(t, dep.RequiredFeatures, 2)
	assert.Contains(t, dep.RequiredFeatures, "ProjectRegistry")
	assert.Contains(t, dep.RequiredFeatures, "ContextManagement")
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestPartialAspectDependency; UPDATED=2025-10-18
func TestPartialAspectDependency(t *testing.T) {
	dep := Dependency{
		Source:         "CBIN-147",
		Target:         "CBIN-129",
		Type:           DependencyTypePartialAspect,
		RequiredAspect: "Storage",
		Description:    "Only needs Storage aspect of database migrations",
	}
	assert.Equal(t, "CBIN-147", dep.Source)
	assert.Equal(t, "CBIN-129", dep.Target)
	assert.Equal(t, DependencyTypePartialAspect, dep.Type)
	assert.Equal(t, "Storage", dep.RequiredAspect)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyStatusCreation; UPDATED=2025-10-18
func TestDependencyStatusCreation(t *testing.T) {
	status := DependencyStatus{
		Dependency: Dependency{
			Source: "CBIN-147",
			Target: "CBIN-146",
			Type:   DependencyTypeFull,
		},
		IsSatisfied: true,
		Blocking:    false,
		Message:     "All required features are TESTED",
	}
	assert.True(t, status.IsSatisfied)
	assert.False(t, status.Blocking)
	assert.Equal(t, "All required features are TESTED", status.Message)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyGraphCreation; UPDATED=2025-10-18
func TestDependencyGraphCreation(t *testing.T) {
	graph := DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-147": {
				{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull},
				{Source: "CBIN-147", Target: "CBIN-145", Type: DependencyTypeFull},
			},
			"CBIN-146": {
				{Source: "CBIN-146", Target: "CBIN-129", Type: DependencyTypePartialFeatures},
			},
		},
	}
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.Nodes["CBIN-147"], 2)
	assert.Len(t, graph.Nodes["CBIN-146"], 1)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyGraphGetDependencies; UPDATED=2025-10-18
func TestDependencyGraphGetDependencies(t *testing.T) {
	graph := DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-147": {
				{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull},
			},
		},
	}
	deps := graph.GetDependencies("CBIN-147")
	assert.Len(t, deps, 1)
	assert.Equal(t, "CBIN-146", deps[0].Target)

	// Test non-existent requirement
	deps = graph.GetDependencies("CBIN-999")
	assert.Len(t, deps, 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyGraphAddDependency; UPDATED=2025-10-18
func TestDependencyGraphAddDependency(t *testing.T) {
	graph := NewDependencyGraph()

	dep := Dependency{
		Source: "CBIN-147",
		Target: "CBIN-146",
		Type:   DependencyTypeFull,
	}

	graph.AddDependency(dep)

	assert.Len(t, graph.Nodes, 1)
	assert.Len(t, graph.Nodes["CBIN-147"], 1)
	assert.Equal(t, "CBIN-146", graph.Nodes["CBIN-147"][0].Target)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyGraphGetReverseDependencies; UPDATED=2025-10-18
func TestDependencyGraphGetReverseDependencies(t *testing.T) {
	graph := DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-147": {
				{Source: "CBIN-147", Target: "CBIN-146", Type: DependencyTypeFull},
			},
			"CBIN-148": {
				{Source: "CBIN-148", Target: "CBIN-146", Type: DependencyTypeFull},
			},
		},
	}

	// CBIN-146 is depended upon by CBIN-147 and CBIN-148
	reverseDeps := graph.GetReverseDependencies("CBIN-146")
	assert.Len(t, reverseDeps, 2)

	sources := []string{reverseDeps[0].Source, reverseDeps[1].Source}
	assert.Contains(t, sources, "CBIN-147")
	assert.Contains(t, sources, "CBIN-148")
}
