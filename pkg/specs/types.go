package specs

// CANARY: REQ=CBIN-147; FEATURE="DependencyModel"; ASPECT=Storage; STATUS=TESTED; TEST=TestDependencyCreation,TestPartialFeatureDependency,TestPartialAspectDependency; UPDATED=2025-10-18

// TokenInfo represents a CANARY token from storage.
// This is used by StatusChecker to query token status.
type TokenInfo struct {
	ReqID   string
	Feature string
	Aspect  string
	Status  string
}

// DependencyType represents the type of dependency relationship between requirements.
type DependencyType int

const (
	// DependencyTypeFull indicates the entire target requirement must be complete
	// (all features must be in TESTED or BENCHED status).
	DependencyTypeFull DependencyType = iota

	// DependencyTypePartialFeatures indicates only specific features of the target
	// requirement must be complete.
	DependencyTypePartialFeatures

	// DependencyTypePartialAspect indicates all features of a specific aspect
	// of the target requirement must be complete.
	DependencyTypePartialAspect
)

// String returns a human-readable string representation of the DependencyType.
func (dt DependencyType) String() string {
	switch dt {
	case DependencyTypeFull:
		return "Full"
	case DependencyTypePartialFeatures:
		return "PartialFeatures"
	case DependencyTypePartialAspect:
		return "PartialAspect"
	default:
		return "Unknown"
	}
}

// Dependency represents a dependency relationship between two requirements.
// It captures the source requirement, target requirement, type of dependency,
// and any specific features or aspects that must be satisfied.
type Dependency struct {
	// Source is the requirement ID that has the dependency (e.g., "CBIN-147")
	Source string

	// Target is the requirement ID being depended upon (e.g., "CBIN-146")
	Target string

	// Type indicates whether this is a full, partial feature, or aspect dependency
	Type DependencyType

	// RequiredFeatures lists specific features needed for PartialFeatures dependencies.
	// Only populated when Type is DependencyTypePartialFeatures.
	RequiredFeatures []string

	// RequiredAspect specifies the aspect needed for PartialAspect dependencies.
	// Only populated when Type is DependencyTypePartialAspect.
	RequiredAspect string

	// Description provides human-readable context about why this dependency exists.
	// This is optional and used for documentation purposes.
	Description string
}

// DependencyStatus represents the current satisfaction status of a dependency.
// It indicates whether the dependency is satisfied and provides contextual information.
type DependencyStatus struct {
	// Dependency is the dependency being evaluated
	Dependency Dependency

	// IsSatisfied indicates whether the dependency requirements are met.
	// For Full: All features of target are TESTED or BENCHED
	// For PartialFeatures: All RequiredFeatures are TESTED or BENCHED
	// For PartialAspect: All features of RequiredAspect are TESTED or BENCHED
	IsSatisfied bool

	// Blocking indicates whether this unsatisfied dependency blocks implementation.
	// Set to true when IsSatisfied is false.
	Blocking bool

	// Message provides human-readable explanation of the status.
	// Examples:
	// - "All required features are TESTED"
	// - "Waiting for CBIN-146:ProjectRegistry to reach TESTED status (currently IMPL)"
	// - "Target requirement CBIN-999 does not exist"
	Message string

	// MissingFeatures lists features that are not yet in TESTED/BENCHED status.
	// Only populated for PartialFeatures dependencies when IsSatisfied is false.
	MissingFeatures []string

	// CurrentStatus describes the current status of the target requirement or features.
	// Used for debugging and reporting.
	CurrentStatus string
}

// DependencyGraph represents the complete dependency graph for all requirements.
// It provides methods for querying, traversal, and cycle detection.
type DependencyGraph struct {
	// Nodes maps requirement IDs to their list of outgoing dependencies.
	// Key: Source requirement ID (e.g., "CBIN-147")
	// Value: List of dependencies where this requirement is the source
	Nodes map[string][]Dependency
}

// NewDependencyGraph creates a new empty DependencyGraph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string][]Dependency),
	}
}

// AddDependency adds a dependency to the graph.
// If the source node doesn't exist, it creates it.
func (dg *DependencyGraph) AddDependency(dep Dependency) {
	if dg.Nodes == nil {
		dg.Nodes = make(map[string][]Dependency)
	}
	dg.Nodes[dep.Source] = append(dg.Nodes[dep.Source], dep)
}

// GetDependencies returns all dependencies for a given requirement ID.
// Returns an empty slice if the requirement has no dependencies.
func (dg *DependencyGraph) GetDependencies(reqID string) []Dependency {
	if deps, exists := dg.Nodes[reqID]; exists {
		return deps
	}
	return []Dependency{}
}

// GetReverseDependencies returns all requirements that depend on the given requirement ID.
// This answers the question: "What would be blocked if this requirement changes?"
func (dg *DependencyGraph) GetReverseDependencies(reqID string) []Dependency {
	var reverseDeps []Dependency
	for _, deps := range dg.Nodes {
		for _, dep := range deps {
			if dep.Target == reqID {
				reverseDeps = append(reverseDeps, dep)
			}
		}
	}
	return reverseDeps
}

// GetAllRequirements returns all unique requirement IDs in the graph
// (both sources and targets).
func (dg *DependencyGraph) GetAllRequirements() []string {
	uniqueReqs := make(map[string]bool)

	// Add all source requirements
	for reqID := range dg.Nodes {
		uniqueReqs[reqID] = true
	}

	// Add all target requirements
	for _, deps := range dg.Nodes {
		for _, dep := range deps {
			uniqueReqs[dep.Target] = true
		}
	}

	// Convert to slice
	reqs := make([]string, 0, len(uniqueReqs))
	for reqID := range uniqueReqs {
		reqs = append(reqs, reqID)
	}

	return reqs
}
