package list

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-125; FEATURE="ListCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var ListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "List CANARY tokens with filtering and ordering",
	Long: `List tokens from the database with priority ordering and filtering.

Supports filtering by status, aspect, phase, owner, and spec status.
Results are ordered by priority (1=highest) and updated date by default.

By default, hides requirements from:
- Test files (*_test.go, /tests/, /test/)
- Template directories (.canary/templates/, /base/, /embedded/)
- Documentation examples (IMPLEMENTATION_SUMMARY, FINAL_SUMMARY, etc.)
- AI agent directories (.claude/, .cursor/, .github/prompts/, etc.)

Use --include-hidden to show all requirements including hidden ones.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		filterStatus, _ := cmd.Flags().GetString("status")
		filterAspect, _ := cmd.Flags().GetString("aspect")
		filterPhase, _ := cmd.Flags().GetString("phase")
		filterOwner, _ := cmd.Flags().GetString("owner")
		filterSpecStatus, _ := cmd.Flags().GetString("spec-status")
		priorityMin, _ := cmd.Flags().GetInt("priority-min")
		priorityMax, _ := cmd.Flags().GetInt("priority-max")
		orderBy, _ := cmd.Flags().GetString("order-by")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		includeHidden, _ := cmd.Flags().GetBool("include-hidden")

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		// Load project config to get ID pattern
		cfg, err := utils.LoadProjectConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load project config: %v\n", err)
		}
		idPattern := ""
		if cfg != nil && cfg.Requirements.IDPattern != "" {
			idPattern = cfg.Requirements.IDPattern
		}

		// Build filters
		filters := make(map[string]string)
		if filterStatus != "" {
			filters["status"] = filterStatus
		}
		if filterAspect != "" {
			filters["aspect"] = filterAspect
		}
		if filterPhase != "" {
			filters["phase"] = filterPhase
		}
		if filterOwner != "" {
			filters["owner"] = filterOwner
		}
		if filterSpecStatus != "" {
			filters["spec_status"] = filterSpecStatus
		}
		if priorityMin > 0 {
			filters["priority_min"] = strconv.Itoa(priorityMin)
		}
		if priorityMax > 0 {
			filters["priority_max"] = strconv.Itoa(priorityMax)
		}
		if includeHidden {
			filters["include_hidden"] = "true"
		}

		tokens, err := db.ListTokens(filters, idPattern, orderBy, limit)
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Println("No tokens found")
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tokens)
		}

		// Display as table
		fmt.Printf("Found %d tokens:\n\n", len(tokens))
		for _, token := range tokens {
			fmt.Printf("📌 %s - %s\n", token.ReqID, token.Feature)
			fmt.Printf("   Status: %s | Aspect: %s | Priority: %d", token.Status, token.Aspect, token.Priority)
			if token.Phase != "" {
				fmt.Printf(" | Phase: %s", token.Phase)
			}
			fmt.Println()
			fmt.Printf("   Location: %s:%d\n", token.FilePath, token.LineNumber)
			if token.Test != "" {
				fmt.Printf("   Test: %s\n", token.Test)
			}
			if token.Owner != "" {
				fmt.Printf("   Owner: %s\n", token.Owner)
			}
			if token.Keywords != "" {
				fmt.Printf("   Keywords: %s\n", token.Keywords)
			}
			fmt.Println()
		}

		return nil
	},
}
