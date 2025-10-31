package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/specs"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-147; FEATURE="DepsParentCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsParentCommand; UPDATED=2025-10-18

// CreateDepsCommand creates the parent deps command
func CreateDepsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Manage requirement dependencies",
		Long: `Commands for working with requirement dependencies.

Dependencies allow specifications to declare that they depend on other
specifications being complete before implementation can begin.

Available commands:
  check    - Check if dependencies are satisfied
  graph    - Show dependency tree visualization
  reverse  - Show what depends on a requirement
  validate - Validate all dependencies for cycles`,
	}

	cmd.AddCommand(createDepsCheckCommand())
	cmd.AddCommand(createDepsGraphCommand())
	cmd.AddCommand(createDepsReverseCommand())
	cmd.AddCommand(createDepsValidateCommand())

	return cmd
}

// CANARY: REQ=CBIN-147; FEATURE="DepsCheckCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsCheckCommand; UPDATED=2025-10-18

// createDepsCheckCommand creates the deps check command
func createDepsCheckCommand() *cobra.Command {
	var showSatisfied bool

	cmd := &cobra.Command{
		Use:   "check <req-id>",
		Short: "Check if dependencies are satisfied",
		Long: `Check if all dependencies for a requirement are satisfied.

This command loads the requirement's dependencies and checks their status
against the CANARY token database. Only TESTED and BENCHED status satisfy
dependencies - IMPL is insufficient.

Example:
  canary deps check CBIN-147`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reqID := args[0]

			// Find spec file
			specPath, err := findSpecFile(reqID)
			if err != nil {
				return fmt.Errorf("failed to find spec for %s: %w", reqID, err)
			}

			// Parse dependencies
			deps, err := specs.ParseDependenciesFromFile(reqID, specPath)
			if err != nil {
				return fmt.Errorf("failed to parse dependencies: %w", err)
			}

			if len(deps) == 0 {
				cmd.Println(fmt.Sprintf("%s has no dependencies", reqID))
				return nil
			}

			// Load token provider
			tokenProvider, err := createTokenProvider()
			if err != nil {
				return fmt.Errorf("failed to create token provider: %w", err)
			}

			// Check dependency status
			checker := specs.NewStatusChecker(tokenProvider)
			statuses := checker.CheckAllDependencies(deps)

			// Display results
			cmd.Println(fmt.Sprintf("Dependency status for %s:", reqID))
			cmd.Println()

			satisfiedCount := 0
			blockingCount := 0

			for _, status := range statuses {
				if status.IsSatisfied {
					satisfiedCount++
					if showSatisfied {
						cmd.Println(fmt.Sprintf("✅ %s - %s", status.Dependency.Target, status.Message))
					}
				} else {
					blockingCount++
					cmd.Println(fmt.Sprintf("❌ %s - %s", status.Dependency.Target, status.Message))
					if len(status.MissingFeatures) > 0 {
						cmd.Println(fmt.Sprintf("   Missing: %s", strings.Join(status.MissingFeatures, ", ")))
					}
				}
			}

			cmd.Println()
			cmd.Println(fmt.Sprintf("Summary: %d satisfied, %d blocking", satisfiedCount, blockingCount))

			if blockingCount > 0 {
				return fmt.Errorf("dependencies not satisfied")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showSatisfied, "show-satisfied", false, "Show satisfied dependencies")

	return cmd
}

// CANARY: REQ=CBIN-147; FEATURE="DepsGraphCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsGraphCommand; UPDATED=2025-10-18

// createDepsGraphCommand creates the deps graph command
func createDepsGraphCommand() *cobra.Command {
	var showStatus bool

	cmd := &cobra.Command{
		Use:   "graph <req-id>",
		Short: "Show dependency tree visualization",
		Long: `Display a visual tree of all dependencies for a requirement.

The tree shows both direct and transitive dependencies with Unicode
box-drawing characters. When --status is used, shows whether each
dependency is satisfied (✅) or blocking (❌).

Example:
  canary deps graph CBIN-147
  canary deps graph CBIN-147 --status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reqID := args[0]

			// Build graph from all specs
			graph, err := buildDependencyGraph()
			if err != nil {
				return fmt.Errorf("failed to build dependency graph: %w", err)
			}

			// Create generator
			generator := specs.NewGraphGenerator(nil)

			// Add status checker if requested
			if showStatus {
				tokenProvider, err := createTokenProvider()
				if err == nil {
					statusChecker := &dependencyStatusAdapter{
						checker: specs.NewStatusChecker(tokenProvider),
					}
					generator.SetStatusChecker(statusChecker)
				}
			}

			// Format and display tree
			tree := generator.FormatASCIITree(graph, reqID)
			cmd.Println(tree)

			// Show summary
			cmd.Println()
			summary := generator.FormatDependencySummary(graph, reqID)
			cmd.Println(summary)

			return nil
		},
	}

	cmd.Flags().BoolVar(&showStatus, "status", false, "Show dependency satisfaction status")

	return cmd
}

// CANARY: REQ=CBIN-147; FEATURE="DepsReverseCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsReverseCommand; UPDATED=2025-10-18

// createDepsReverseCommand creates the deps reverse command
func createDepsReverseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reverse <req-id>",
		Short: "Show what depends on a requirement",
		Long: `Display all requirements that depend on the specified requirement.

This answers the question: "What would be blocked if this requirement changes?"

Example:
  canary deps reverse CBIN-146`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reqID := args[0]

			// Build graph from all specs
			graph, err := buildDependencyGraph()
			if err != nil {
				return fmt.Errorf("failed to build dependency graph: %w", err)
			}

			// Get reverse dependencies
			reverseDeps := graph.GetReverseDependencies(reqID)

			if len(reverseDeps) == 0 {
				cmd.Println(fmt.Sprintf("No requirements depend on %s", reqID))
				return nil
			}

			cmd.Println(fmt.Sprintf("Requirements that depend on %s:", reqID))
			cmd.Println()

			for _, dep := range reverseDeps {
				typeStr := ""
				switch dep.Type {
				case specs.DependencyTypePartialFeatures:
					typeStr = fmt.Sprintf(" (features: %s)", strings.Join(dep.RequiredFeatures, ", "))
				case specs.DependencyTypePartialAspect:
					typeStr = fmt.Sprintf(" (aspect: %s)", dep.RequiredAspect)
				}

				cmd.Println(fmt.Sprintf("  %s%s", dep.Source, typeStr))
				if dep.Description != "" {
					cmd.Println(fmt.Sprintf("    %s", dep.Description))
				}
			}

			cmd.Println()
			cmd.Println(fmt.Sprintf("Total: %d requirements depend on %s", len(reverseDeps), reqID))

			return nil
		},
	}

	return cmd
}

// CANARY: REQ=CBIN-147; FEATURE="DepsValidateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsValidateCommand,TestDepsValidateCommand_DetectsCycle; UPDATED=2025-10-18

// createDepsValidateCommand creates the deps validate command
func createDepsValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all dependencies for cycles",
		Long: `Validate the entire dependency graph for issues.

Checks for:
- Circular dependencies (A depends on B, B depends on A)
- Missing requirements (dependencies on non-existent specs)
- Self-dependencies (A depends on A)

Example:
  canary deps validate`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build graph from all specs
			graph, err := buildDependencyGraph()
			if err != nil {
				return fmt.Errorf("failed to build dependency graph: %w", err)
			}

			// Create validator
			validator := specs.NewDependencyValidator(graph)

			// Add spec finder to check for missing requirements
			specFinder := &filesystemSpecFinder{}
			validator.SetSpecFinder(specFinder)

			// Validate
			result := validator.Validate()

			if result.IsValid {
				cmd.Println("✅ All dependencies are valid")
				cmd.Println(fmt.Sprintf("Validated %d requirements with %d dependencies",
					len(graph.GetAllRequirements()), countTotalDependencies(graph)))
				return nil
			}

			// Display errors
			cmd.Println("❌ Dependency validation failed:")
			cmd.Println()
			cmd.Println(result.FormatErrors())

			return fmt.Errorf("validation failed")
		},
	}

	return cmd
}

// Helper functions

// findSpecFile finds the spec.md file for a requirement ID
func findSpecFile(reqID string) (string, error) {
	// Look in .canary/specs/
	specsDir := ".canary/specs"
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name starts with the requirement ID
		if strings.HasPrefix(entry.Name(), reqID) {
			specPath := filepath.Join(specsDir, entry.Name(), "spec.md")
			if _, err := os.Stat(specPath); err == nil {
				return specPath, nil
			}
		}
	}

	return "", fmt.Errorf("spec file not found for %s", reqID)
}

// buildDependencyGraph builds the complete dependency graph from all specs
func buildDependencyGraph() (*specs.DependencyGraph, error) {
	graph := specs.NewDependencyGraph()

	// Find all spec directories
	specsDir := ".canary/specs"
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Extract requirement ID from directory name (before first hyphen)
		parts := strings.Split(entry.Name(), "-")
		if len(parts) < 2 {
			continue
		}
		reqID := parts[0] + "-" + parts[1] // CBIN-###

		// Load dependencies
		specPath := filepath.Join(specsDir, entry.Name(), "spec.md")
		deps, err := specs.ParseDependenciesFromFile(reqID, specPath)
		if err != nil {
			// Skip specs without valid dependencies
			continue
		}

		for _, dep := range deps {
			graph.AddDependency(dep)
		}
	}

	return graph, nil
}

// createTokenProvider creates a token provider from the database
func createTokenProvider() (specs.TokenProvider, error) {
	// Try to open database
	dbPath := getDatabasePath()
	db, err := storage.Open(dbPath)
	if err != nil {
		// Return empty provider if no database
		return &emptyTokenProvider{}, nil
	}

	return &dbTokenProvider{db: db}, nil
}

// getDatabasePath returns the path to the canary database
func getDatabasePath() string {
	// Check for project-local database
	if _, err := os.Stat(".canary/canary.db"); err == nil {
		return ".canary/canary.db"
	}

	// Fall back to global database
	home, err := os.UserHomeDir()
	if err != nil {
		return "canary.db"
	}
	return filepath.Join(home, ".canary", "canary.db")
}

// countTotalDependencies counts all dependencies in the graph
func countTotalDependencies(graph *specs.DependencyGraph) int {
	count := 0
	for _, deps := range graph.Nodes {
		count += len(deps)
	}
	return count
}

// Adapter types

// dependencyStatusAdapter adapts StatusChecker to StatusCheckerInterface
type dependencyStatusAdapter struct {
	checker *specs.StatusChecker
}

func (a *dependencyStatusAdapter) IsDependencySatisfied(dep specs.Dependency) bool {
	status := a.checker.CheckDependency(dep)
	return status.IsSatisfied
}

// filesystemSpecFinder implements SpecFinder using filesystem
type filesystemSpecFinder struct{}

func (f *filesystemSpecFinder) SpecExists(reqID string) bool {
	_, err := findSpecFile(reqID)
	return err == nil
}

func (f *filesystemSpecFinder) FindSpecPath(reqID string) (string, error) {
	return findSpecFile(reqID)
}

// emptyTokenProvider returns empty token lists when database is unavailable
type emptyTokenProvider struct{}

func (e *emptyTokenProvider) GetTokensByReqID(reqID string) []specs.TokenInfo {
	return []specs.TokenInfo{}
}

// dbTokenProvider fetches tokens from the database
type dbTokenProvider struct {
	db *storage.DB
}

func (d *dbTokenProvider) GetTokensByReqID(reqID string) []specs.TokenInfo {
	// Use DB method to get tokens
	dbTokens, err := d.db.GetTokensByReqID(reqID)
	if err != nil {
		return []specs.TokenInfo{}
	}

	// Convert storage.Token to specs.TokenInfo
	var tokens []specs.TokenInfo
	for _, dbToken := range dbTokens {
		tokens = append(tokens, specs.TokenInfo{
			ReqID:   dbToken.ReqID,
			Feature: dbToken.Feature,
			Aspect:  dbToken.Aspect,
			Status:  dbToken.Status,
		})
	}

	return tokens
}
