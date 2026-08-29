package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/specs"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=CP-262; FEATURE="DepsParentCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsParentCommand; UPDATED=2026-08-29

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

// CANARY: REQ=CP-262; FEATURE="DepsCheckCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsCheckCommand; UPDATED=2026-08-29
// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3960_DepsCheck_ExternalSatisfied,TestCANARY_ENG_3960_DepsCheck_ExternalUnsatisfied,TestCANARY_ENG_3960_DepsCheck_ExternalUnknown,TestCANARY_ENG_3960_DepsValidate_ExternalNotMissing,TestCANARY_ENG_3960_DepsValidate_ExternalCountsLine,TestCANARY_ENG_3960_DepsValidate_StrictExternalFails,TestCANARY_ENG_3960_DepsGraph_MermaidExternalClass,TestCANARY_ENG_3960_DepsValidate_ExternalWithLocalTokens_CountedMissing; UPDATED=2026-08-29
// CANARY: REQ=ENG-3961; FEATURE="PeerProjects"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3961_DepsCheck_PeerDetailSurfacesName; UPDATED=2026-08-29

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
			reg := sources.LoadFromRoot(".")

			// Display results
			cmd.Println(fmt.Sprintf("Dependency status for %s:", reqID))
			cmd.Println()

			satisfiedCount := 0
			blockingCount := 0

			for _, status := range statuses {
				// A dependency with zero local tokens (CurrentStatus MISSING)
				// may actually be owned by an external ticket source rather
				// than genuinely missing — consult external.Resolve before
				// falling back to the local-spec MISSING treatment.
				if status.CurrentStatus == "MISSING" {
					res := external.Resolve(status.Dependency.Target, reg, ".")
					if res.IsExternal() {
						switch res.State {
						case external.StateSatisfied:
							satisfiedCount++
							cmd.Println(fmt.Sprintf("✔ external %s (%s)", status.Dependency.Target, res.ShortDetail()))
						case external.StateUnsatisfied:
							blockingCount++
							cmd.Println(fmt.Sprintf("✖ external %s (%s)", status.Dependency.Target, res.ShortDetail()))
						default: // unknown: never blocking here (degradation is sacred)
							cmd.Println(fmt.Sprintf("? external %s (%s)", status.Dependency.Target, res.ShortDetail()))
						}
						continue
					}
				}

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

// CANARY: REQ=CP-262; FEATURE="DepsGraphCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsGraphCommand; UPDATED=2026-08-29
// CANARY: REQ=CP-269; FEATURE="MermaidGraph"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_203_FormatMermaid; UPDATED=2026-08-28

// createDepsGraphCommand creates the deps graph command
func createDepsGraphCommand() *cobra.Command {
	var showStatus bool
	var format string

	cmd := &cobra.Command{
		Use:   "graph <req-id>",
		Short: "Show dependency tree visualization",
		Long: `Display a visual tree of all dependencies for a requirement.

The tree shows both direct and transitive dependencies with Unicode
box-drawing characters. When --status is used, shows whether each
dependency is satisfied (✅) or blocking (❌).

Use --format mermaid to render the graph as a mermaid flowchart instead,
with click-through links to tickets where a source is configured.

Example:
  canary deps graph CBIN-147
  canary deps graph CBIN-147 --status
  canary deps graph CBIN-147 --format mermaid`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reqID := args[0]

			if format != "ascii" && format != "mermaid" {
				return fmt.Errorf("invalid --format %q: must be \"ascii\" or \"mermaid\"", format)
			}

			// Build graph from all specs
			graph, err := BuildGraph()
			if err != nil {
				return fmt.Errorf("failed to build dependency graph: %w", err)
			}

			// Create generator
			generator := specs.NewGraphGenerator(nil)

			if format == "mermaid" {
				reg := sources.LoadFromRoot(".")
				tokenProvider, _ := createTokenProvider()
				// Local tokens always win: a node with at least one local
				// CANARY token in the index is never styled as external,
				// even when its id also matches a configured external
				// source's key -- external.Resolve is consulted only when
				// zero local tokens exist for the node. Mirrors the same
				// rule in countExternalDeps and view.annotateExternal.
				isExternal := func(id string) bool {
					if tokenProvider != nil && len(tokenProvider.GetTokensByReqID(id)) > 0 {
						return false
					}
					return external.Resolve(id, reg, ".").IsExternal()
				}
				cmd.Println(generator.FormatMermaid(graph, reqID, reg.TicketURL, isExternal))
				return nil
			}

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
	cmd.Flags().StringVar(&format, "format", "ascii", "Output format: ascii or mermaid")

	return cmd
}

// CANARY: REQ=CP-262; FEATURE="DepsReverseCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsReverseCommand; UPDATED=2026-08-29

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
			graph, err := BuildGraph()
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

// CANARY: REQ=CP-262; FEATURE="DepsValidateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsValidateCommand,TestDepsValidateCommand_DetectsCycle; UPDATED=2026-08-29

// createDepsValidateCommand creates the deps validate command
func createDepsValidateCommand() *cobra.Command {
	var strictExternal bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all dependencies for cycles",
		Long: `Validate the entire dependency graph for issues.

Checks for:
- Circular dependencies (A depends on B, B depends on A)
- Missing requirements (dependencies on non-existent specs)
- Self-dependencies (A depends on A)

Dependencies resolving to an external ticket source (e.g. JIRA) are never
"missing spec" errors -- their satisfied/unsatisfied/unknown counts are
reported on a separate "external:" summary line. Use --strict-external to
fail validation when any external dependency is unsatisfied or has no
cached status.

Example:
  canary deps validate
  canary deps validate --strict-external`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build graph from all specs
			graph, err := BuildGraph()
			if err != nil {
				return fmt.Errorf("failed to build dependency graph: %w", err)
			}

			reg := sources.LoadFromRoot(".")

			tokenProvider, err := createTokenProvider()
			if err != nil {
				return fmt.Errorf("failed to create token provider: %w", err)
			}

			// Create validator
			validator := specs.NewDependencyValidator(graph)

			// Add spec finder to check for missing requirements; external
			// ids never count as missing (they live in a ticket source) --
			// but only when they have no local tokens. Local tokens always
			// win (mirrors countExternalDeps and view.annotateExternal): an
			// id whose prefix matches a configured external source but that
			// also has local CANARY tokens and no spec dir is genuinely
			// missing, not external.
			validator.SetSpecFinder(&externalAwareSpecFinder{
				inner:         &filesystemSpecFinder{},
				reg:           reg,
				tokenProvider: tokenProvider,
			})

			// Validate
			result := validator.Validate()

			satisfied, unsatisfied, unknown := countExternalDeps(graph, reg, tokenProvider)
			externalLine := fmt.Sprintf("external: satisfied=%d unsatisfied=%d unknown=%d", satisfied, unsatisfied, unknown)
			externalFailing := strictExternal && (unsatisfied > 0 || unknown > 0)

			// deps validate is a REPORTING gate, not a blocking gate: by
			// design, an unsatisfied or unknown external dependency never
			// fails validate in default (non-strict) mode -- it only shows
			// up in externalLine below. Blocking on external dependency
			// status is `next`'s job (hasUnresolvedDependencies) and
			// `deps check`'s job, both of which already treat unsatisfied
			// as blocking unconditionally. --strict-external is the opt-in
			// that makes validate itself fail on unsatisfied/unknown
			// externals; this asymmetry vs next/check is intentional.
			if result.IsValid && !externalFailing {
				cmd.Println("✅ All dependencies are valid")
				cmd.Println(fmt.Sprintf("Validated %d requirements with %d dependencies",
					len(graph.GetAllRequirements()), countTotalDependencies(graph)))
				cmd.Println(externalLine)
				return nil
			}

			// Display errors
			cmd.Println("❌ Dependency validation failed:")
			cmd.Println()
			if !result.IsValid {
				cmd.Println(result.FormatErrors())
			}
			cmd.Println(externalLine)
			if externalFailing {
				cmd.Println("external dependencies failing --strict-external (unsatisfied or unknown status)")
			}

			return fmt.Errorf("validation failed")
		},
	}

	cmd.Flags().BoolVar(&strictExternal, "strict-external", false, "fail validation when external dependencies are unsatisfied or have unknown (uncached) status")

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

// BuildGraph builds the complete dependency graph from all specs under
// .canary/specs/. It is exported so callers outside this package (e.g. the
// MCP deps tool) can reuse it without reconstructing the spec-walk logic.
func BuildGraph() (*specs.DependencyGraph, error) {
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

// countExternalDeps classifies every unique dependency target in graph that
// resolves to an external (ticket-source) id -- via external.Resolve -- into
// satisfied/unsatisfied/unknown counts for `deps validate`'s summary line.
// Local/flatfile/unconfigured targets (Detail "not external") are excluded;
// those are covered by the validator's own missing-requirement check.
//
// Local tokens always win: a target with at least one local CANARY token in
// tokenProvider is excluded from this count entirely -- it's a local
// requirement (tracked by the validator's own missing/satisfied logic), not
// an external one, even when its id also matches a configured external
// source's key. external.Resolve is consulted only when tokenProvider
// reports zero local tokens for the target. tokenProvider may be nil (no
// local-token gate applied) for callers without one.
func countExternalDeps(graph *specs.DependencyGraph, reg *sources.Registry, tokenProvider specs.TokenProvider) (satisfied, unsatisfied, unknown int) {
	checked := map[string]bool{}
	for _, deps := range graph.Nodes {
		for _, dep := range deps {
			target := dep.Target
			if checked[target] {
				continue
			}
			checked[target] = true

			if tokenProvider != nil && len(tokenProvider.GetTokensByReqID(target)) > 0 {
				continue
			}

			res := external.Resolve(target, reg, ".")
			if !res.IsExternal() {
				continue
			}
			switch res.State {
			case external.StateSatisfied:
				satisfied++
			case external.StateUnsatisfied:
				unsatisfied++
			default:
				unknown++
			}
		}
	}
	return satisfied, unsatisfied, unknown
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

// externalAwareSpecFinder wraps a SpecFinder so that ids resolving to an
// external ticket source (per external.Resolve) never count as "missing" --
// they live in the ticket source, not in .canary/specs/. Local tokens always
// win: an id with at least one local CANARY token (per tokenProvider) is
// never treated as external here, even when its prefix also matches a
// configured external source's key -- external.Resolve is consulted only
// when tokenProvider reports zero local tokens for the id. tokenProvider may
// be nil (no local-token gate applied) for callers without one.
type externalAwareSpecFinder struct {
	inner         specs.SpecFinder
	reg           *sources.Registry
	tokenProvider specs.TokenProvider
}

func (f *externalAwareSpecFinder) SpecExists(reqID string) bool {
	if f.inner.SpecExists(reqID) {
		return true
	}
	if f.tokenProvider != nil && len(f.tokenProvider.GetTokensByReqID(reqID)) > 0 {
		return false
	}
	return external.Resolve(reqID, f.reg, ".").IsExternal()
}

func (f *externalAwareSpecFinder) FindSpecPath(reqID string) (string, error) {
	return f.inner.FindSpecPath(reqID)
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
