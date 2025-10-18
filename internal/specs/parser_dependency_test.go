package specs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_FullDependency; UPDATED=2025-10-18
func TestParseDependencies_FullDependency(t *testing.T) {
	specContent := `# Specification

## Dependencies

### Full Dependencies (entire requirement must be complete)
- CBIN-123 (UserAuthentication - required for session management)
- CBIN-145 (LegacyMigration - database schema must be updated)

## Features
...
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	require.Len(t, deps, 2)

	// Check first dependency
	assert.Equal(t, "CBIN-147", deps[0].Source)
	assert.Equal(t, "CBIN-123", deps[0].Target)
	assert.Equal(t, DependencyTypeFull, deps[0].Type)
	assert.Contains(t, deps[0].Description, "UserAuthentication")

	// Check second dependency
	assert.Equal(t, "CBIN-147", deps[1].Source)
	assert.Equal(t, "CBIN-145", deps[1].Target)
	assert.Equal(t, DependencyTypeFull, deps[1].Type)
	assert.Contains(t, deps[1].Description, "LegacyMigration")
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_PartialFeatures; UPDATED=2025-10-18
func TestParseDependencies_PartialFeatures(t *testing.T) {
	specContent := `# Specification

## Dependencies

### Partial Dependencies (specific features/aspects required)
- CBIN-146:ProjectRegistry,ContextManagement (need project management features)
- CBIN-140:GapRepository,GapService (gap storage and service layer required)

## Features
...
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	require.Len(t, deps, 2)

	// Check first partial dependency
	assert.Equal(t, "CBIN-147", deps[0].Source)
	assert.Equal(t, "CBIN-146", deps[0].Target)
	assert.Equal(t, DependencyTypePartialFeatures, deps[0].Type)
	assert.Len(t, deps[0].RequiredFeatures, 2)
	assert.Contains(t, deps[0].RequiredFeatures, "ProjectRegistry")
	assert.Contains(t, deps[0].RequiredFeatures, "ContextManagement")

	// Check second partial dependency
	assert.Equal(t, "CBIN-147", deps[1].Source)
	assert.Equal(t, "CBIN-140", deps[1].Target)
	assert.Equal(t, DependencyTypePartialFeatures, deps[1].Type)
	assert.Len(t, deps[1].RequiredFeatures, 2)
	assert.Contains(t, deps[1].RequiredFeatures, "GapRepository")
	assert.Contains(t, deps[1].RequiredFeatures, "GapService")
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_PartialAspect; UPDATED=2025-10-18
func TestParseDependencies_PartialAspect(t *testing.T) {
	specContent := `# Specification

## Dependencies

### Partial Dependencies (specific features/aspects required)
- CBIN-129:Storage (only Storage aspect of database migrations needed)
- CBIN-133:Engine (only Engine aspect needed)

## Features
...
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	require.Len(t, deps, 2)

	// Check first aspect dependency
	assert.Equal(t, "CBIN-147", deps[0].Source)
	assert.Equal(t, "CBIN-129", deps[0].Target)
	assert.Equal(t, DependencyTypePartialAspect, deps[0].Type)
	assert.Equal(t, "Storage", deps[0].RequiredAspect)

	// Check second aspect dependency
	assert.Equal(t, "CBIN-147", deps[1].Source)
	assert.Equal(t, "CBIN-133", deps[1].Target)
	assert.Equal(t, DependencyTypePartialAspect, deps[1].Type)
	assert.Equal(t, "Engine", deps[1].RequiredAspect)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_MixedTypes; UPDATED=2025-10-18
func TestParseDependencies_MixedTypes(t *testing.T) {
	specContent := `# Specification

## Dependencies

### Full Dependencies (entire requirement must be complete)
- CBIN-123 (UserAuthentication)

### Partial Dependencies (specific features/aspects required)
- CBIN-146:ProjectRegistry (need project registry)
- CBIN-129:Storage (only Storage aspect needed)

## Features
...
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	require.Len(t, deps, 3)

	// Check full dependency
	assert.Equal(t, "CBIN-123", deps[0].Target)
	assert.Equal(t, DependencyTypeFull, deps[0].Type)

	// Check partial feature dependency
	assert.Equal(t, "CBIN-146", deps[1].Target)
	assert.Equal(t, DependencyTypePartialFeatures, deps[1].Type)
	assert.Contains(t, deps[1].RequiredFeatures, "ProjectRegistry")

	// Check partial aspect dependency
	assert.Equal(t, "CBIN-129", deps[2].Target)
	assert.Equal(t, DependencyTypePartialAspect, deps[2].Type)
	assert.Equal(t, "Storage", deps[2].RequiredAspect)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_NoDependencies; UPDATED=2025-10-18
func TestParseDependencies_NoDependencies(t *testing.T) {
	specContent := `# Specification

## Features
- Feature1
- Feature2
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	assert.Len(t, deps, 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_EmptyDependencySection; UPDATED=2025-10-18
func TestParseDependencies_EmptyDependencySection(t *testing.T) {
	specContent := `# Specification

## Dependencies

## Features
- Feature1
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	assert.Len(t, deps, 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_InvalidFormat; UPDATED=2025-10-18
func TestParseDependencies_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "Missing colon for partial dependency",
			content: `## Dependencies
- CBIN-146ProjectRegistry (missing colon)`,
			wantErr: false, // Treated as full dependency
		},
		{
			name: "Empty requirement ID",
			content: `## Dependencies
- (no ID)`,
			wantErr: false, // Skipped
		},
		{
			name: "Multiple colons",
			content: `## Dependencies
- CBIN-146:Feature1:Feature2 (multiple colons)`,
			wantErr: false, // Treated as Feature1:Feature2 string
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps, err := ParseDependencies("CBIN-147", strings.NewReader(tc.content))
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Just verify no crash - content is malformed
				_ = deps
			}
		})
	}
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependenciesFromFile; UPDATED=2025-10-18
func TestParseDependenciesFromFile(t *testing.T) {
	// This will test the file-based parser
	// For now, just test that it handles non-existent files
	_, err := ParseDependenciesFromFile("CBIN-147", "/tmp/nonexistent-spec-file.md")
	assert.Error(t, err)
}

// CANARY: REQ=CBIN-147; FEATURE="DependencyParser"; ASPECT=Engine; STATUS=TESTED; TEST=TestParseDependencies_PreservesDescription; UPDATED=2025-10-18
func TestParseDependencies_PreservesDescription(t *testing.T) {
	specContent := `## Dependencies

### Full Dependencies (entire requirement must be complete)
- CBIN-123 (This is a detailed description with multiple words and - special chars!)
`

	deps, err := ParseDependencies("CBIN-147", strings.NewReader(specContent))
	require.NoError(t, err)
	require.Len(t, deps, 1)

	assert.Contains(t, deps[0].Description, "This is a detailed description")
	assert.Contains(t, deps[0].Description, "special chars!")
}
