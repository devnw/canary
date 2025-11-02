// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-140; FEATURE="GapCLI"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
package gap

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/gap"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-140; FEATURE="GapMarkCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var GapCmd = &cobra.Command{
	Use:   "gap <subcommand>",
	Short: "Manage gap analysis entries for implementation mistakes",
	Long: `Track and query implementation gaps to improve future development.

Gap analysis helps agents learn from past mistakes by:
- Recording what went wrong in implementations
- Tracking corrective actions taken
- Ranking gaps by helpfulness
- Automatically injecting relevant gaps into planning prompts

Subcommands:
  mark      Record a new gap analysis entry
  query     Query gaps with filters
  report    Generate gap analysis report
  helpful   Mark a gap as helpful
  unhelpful Mark a gap as unhelpful
  config    View or update gap analysis configuration`,
}

// CANARY: REQ=CBIN-140; FEATURE="GapMarkCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapMarkCmd = &cobra.Command{
	Use:   "mark <req-id> <feature> --category <category> --description <description>",
	Short: "Record a new gap analysis entry",
	Long: `Record a new gap analysis entry for an implementation mistake.

Categories:
  logic_error     - Incorrect business logic or algorithm
  test_failure    - Tests incorrectly written or missing cases
  performance     - Performance issues or inefficient implementation
  security        - Security vulnerabilities or insecure practices
  edge_case       - Unhandled edge cases or boundary conditions
  integration     - Integration issues with existing systems
  documentation   - Incorrect or misleading documentation
  other           - Other types of implementation gaps

Examples:
  # Record a logic error
  canary gap mark CBIN-140 GapTracking \
    --category logic_error \
    --description "Incorrect query ordering in GetFilesByReqID" \
    --action "Added ORDER BY clause to sort by file path"

  # Record a test failure
  canary gap mark CBIN-141 PromptFlag \
    --category test_failure \
    --description "Missing edge case test for empty prompt" \
    --aspect CLI`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		reqID := args[0]
		feature := args[1]
		aspect, _ := cmd.Flags().GetString("aspect")
		category, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")
		action, _ := cmd.Flags().GetString("action")
		createdBy, _ := cmd.Flags().GetString("created-by")

		// Validate required fields
		if category == "" {
			return fmt.Errorf("--category is required")
		}
		if description == "" {
			return fmt.Errorf("--description is required")
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Mark gap
		gapID, err := service.MarkGap(reqID, feature, aspect, category, description, action, createdBy)
		if err != nil {
			return fmt.Errorf("mark gap: %w", err)
		}

		fmt.Printf("✅ Created gap analysis entry: %s\n", gapID)
		fmt.Printf("\nRequirement: %s\n", reqID)
		fmt.Printf("Feature: %s\n", feature)
		fmt.Printf("Category: %s\n", category)
		fmt.Printf("Description: %s\n", description)
		if action != "" {
			fmt.Printf("Corrective Action: %s\n", action)
		}
		fmt.Printf("\nThis gap will be included in future planning for %s.\n", reqID)
		fmt.Printf("Mark as helpful with: canary gap helpful %s\n", gapID)

		return nil
	},
}

// CANARY: REQ=CBIN-140; FEATURE="GapQueryCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapQueryCmd = &cobra.Command{
	Use:   "query [flags]",
	Short: "Query gap analysis entries with filters",
	Long: `Query gap entries with optional filters.

Examples:
  # Query all gaps for a requirement
  canary gap query --req-id CBIN-140

  # Query by category
  canary gap query --category logic_error

  # Query by feature
  canary gap query --feature GapTracking

  # Query with limit
  canary gap query --req-id CBIN-140 --limit 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		reqID, _ := cmd.Flags().GetString("req-id")
		feature, _ := cmd.Flags().GetString("feature")
		aspect, _ := cmd.Flags().GetString("aspect")
		category, _ := cmd.Flags().GetString("category")
		limit, _ := cmd.Flags().GetInt("limit")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Query gaps
		gaps, err := service.QueryGaps(reqID, feature, aspect, category, limit)
		if err != nil {
			return fmt.Errorf("query gaps: %w", err)
		}

		if len(gaps) == 0 {
			fmt.Println("No gaps found matching the criteria")
			return nil
		}

		fmt.Printf("Found %d gap(s):\n\n", len(gaps))
		for i, g := range gaps {
			fmt.Printf("%d. %s - %s\n", i+1, g.GapID, g.Feature)
			fmt.Printf("   Requirement: %s\n", g.ReqID)
			fmt.Printf("   Category: %s\n", g.Category)
			if g.Aspect != "" {
				fmt.Printf("   Aspect: %s\n", g.Aspect)
			}
			fmt.Printf("   Description: %s\n", g.Description)
			if g.CorrectiveAction != "" {
				fmt.Printf("   Corrective Action: %s\n", g.CorrectiveAction)
			}
			fmt.Printf("   Helpful: %d | Unhelpful: %d\n", g.HelpfulCount, g.UnhelpfulCount)
			fmt.Printf("   Created: %s by %s\n", g.CreatedAt.Format("2006-01-02"), g.CreatedBy)
			fmt.Println()
		}

		return nil
	},
}

// CANARY: REQ=CBIN-140; FEATURE="GapReportCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapReportCmd = &cobra.Command{
	Use:   "report <req-id>",
	Short: "Generate gap analysis report for a requirement",
	Long: `Generate a comprehensive gap analysis report for a requirement.

The report includes:
- Total number of gaps
- Gaps grouped by category
- Descriptions and corrective actions
- Helpfulness ratings

Example:
  canary gap report CBIN-140`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		reqID := args[0]

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Generate report
		report, err := service.GenerateReport(reqID)
		if err != nil {
			return fmt.Errorf("generate report: %w", err)
		}

		fmt.Println(report)
		return nil
	},
}

// CANARY: REQ=CBIN-140; FEATURE="GapHelpfulCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapHelpfulCmd = &cobra.Command{
	Use:   "helpful <gap-id>",
	Short: "Mark a gap entry as helpful",
	Long: `Mark a gap analysis entry as helpful.

Helpful gaps are prioritized when injecting into planning prompts.

Example:
  canary gap helpful GAP-CBIN-140-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		gapID := args[0]

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Mark helpful
		if err := service.MarkHelpful(gapID); err != nil {
			return fmt.Errorf("mark helpful: %w", err)
		}

		fmt.Printf("✅ Marked %s as helpful\n", gapID)
		fmt.Println("\nThis gap will be prioritized in future planning prompts.")
		return nil
	},
}

// CANARY: REQ=CBIN-140; FEATURE="GapUnhelpfulCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapUnhelpfulCmd = &cobra.Command{
	Use:   "unhelpful <gap-id>",
	Short: "Mark a gap entry as unhelpful",
	Long: `Mark a gap analysis entry as unhelpful.

Unhelpful gaps are deprioritized when injecting into planning prompts.

Example:
  canary gap unhelpful GAP-CBIN-140-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		gapID := args[0]

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Mark unhelpful
		if err := service.MarkUnhelpful(gapID); err != nil {
			return fmt.Errorf("mark unhelpful: %w", err)
		}

		fmt.Printf("✅ Marked %s as unhelpful\n", gapID)
		return nil
	},
}

// CANARY: REQ=CBIN-140; FEATURE="GapConfigCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapConfigCmd = &cobra.Command{
	Use:   "config [flags]",
	Short: "View or update gap analysis configuration",
	Long: `View or update gap analysis configuration settings.

Configuration settings:
  --max-gaps           Maximum gaps to inject into planning (default: 10)
  --min-helpful        Minimum helpful count to include (default: 1)
  --ranking            Ranking strategy: helpful_desc, recency_desc, weighted (default: helpful_desc)

Examples:
  # View current configuration
  canary gap config

  # Update configuration
  canary gap config --max-gaps 20 --min-helpful 2 --ranking weighted`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		maxGaps, _ := cmd.Flags().GetInt("max-gaps")
		minHelpful, _ := cmd.Flags().GetInt("min-helpful")
		ranking, _ := cmd.Flags().GetString("ranking")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Check if update flags provided
		updateMode := cmd.Flags().Changed("max-gaps") ||
			cmd.Flags().Changed("min-helpful") ||
			cmd.Flags().Changed("ranking")

		if updateMode {
			// Get current config to preserve unchanged values
			currentConfig, err := service.GetConfig()
			if err != nil {
				return fmt.Errorf("get current config: %w", err)
			}

			// Use current values if flags not provided
			if !cmd.Flags().Changed("max-gaps") {
				maxGaps = currentConfig.MaxGapInjection
			}
			if !cmd.Flags().Changed("min-helpful") {
				minHelpful = currentConfig.MinHelpfulThreshold
			}
			if !cmd.Flags().Changed("ranking") {
				ranking = currentConfig.RankingStrategy
			}

			// Update configuration
			if err := service.UpdateConfig(maxGaps, minHelpful, ranking); err != nil {
				return fmt.Errorf("update config: %w", err)
			}

			fmt.Println("✅ Configuration updated:")
		} else {
			// Just display current config
			fmt.Println("Current gap analysis configuration:")
		}

		// Display configuration
		config, err := service.GetConfig()
		if err != nil {
			return fmt.Errorf("get config: %w", err)
		}

		fmt.Printf("\n  Max Gaps to Inject: %d\n", config.MaxGapInjection)
		fmt.Printf("  Min Helpful Threshold: %d\n", config.MinHelpfulThreshold)
		fmt.Printf("  Ranking Strategy: %s\n", config.RankingStrategy)
		fmt.Printf("  Last Updated: %s\n", config.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

// CANARY: REQ=CBIN-140; FEATURE="GapCategoriesCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var gapCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List available gap categories",
	Long: `List all available gap analysis categories.

Categories help classify the type of implementation mistake.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Create service
		repo := storage.NewGapRepository(db)
		service := gap.NewService(repo)

		// Get categories
		categories, err := service.GetCategories()
		if err != nil {
			return fmt.Errorf("get categories: %w", err)
		}

		fmt.Println("Available gap categories:")
		for _, cat := range categories {
			fmt.Printf("  %-20s %s\n", cat.Name, cat.Description)
		}

		return nil
	},
}

func init() {
	// Add gap subcommands
	GapCmd.AddCommand(gapMarkCmd)
	GapCmd.AddCommand(gapQueryCmd)
	GapCmd.AddCommand(gapReportCmd)
	GapCmd.AddCommand(gapHelpfulCmd)
	GapCmd.AddCommand(gapUnhelpfulCmd)
	GapCmd.AddCommand(gapConfigCmd)
	GapCmd.AddCommand(gapCategoriesCmd)

	// Global flags for all gap commands
	GapCmd.PersistentFlags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	GapCmd.PersistentFlags().String("db", ".canary/canary.db", "path to database file")

	// gapMarkCmd flags
	gapMarkCmd.Flags().String("aspect", "", "implementation aspect (API, CLI, Engine, etc.)")
	gapMarkCmd.Flags().String("category", "", "gap category (required)")
	gapMarkCmd.Flags().String("description", "", "what went wrong (required)")
	gapMarkCmd.Flags().String("action", "", "corrective action taken")
	gapMarkCmd.Flags().String("created-by", "agent", "who identified the gap")

	// gapQueryCmd flags
	gapQueryCmd.Flags().String("req-id", "", "filter by requirement ID")
	gapQueryCmd.Flags().String("feature", "", "filter by feature name")
	gapQueryCmd.Flags().String("aspect", "", "filter by aspect")
	gapQueryCmd.Flags().String("category", "", "filter by category")
	gapQueryCmd.Flags().Int("limit", 0, "maximum number of results (0 = no limit)")

	// gapConfigCmd flags
	gapConfigCmd.Flags().Int("max-gaps", 10, "maximum gaps to inject into planning")
	gapConfigCmd.Flags().Int("min-helpful", 1, "minimum helpful count to include")
	gapConfigCmd.Flags().String("ranking", "helpful_desc", "ranking strategy (helpful_desc, recency_desc, weighted)")
}
