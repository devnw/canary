package list

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=ENG-4308; FEATURE="ListCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
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
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
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
		effLimit := utils.EffectiveLimit(limit, defaultListLimit)
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

		tokens, err := db.ListTokens(filters, idPattern, orderBy, effLimit)
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Println("No tokens found")
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
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

		if effLimit > 0 && len(tokens) == effLimit {
			fmt.Printf("(showing %d; use --limit -1 for all)\n", effLimit)
		}

		return nil
	},
}

// CANARY: REQ=ENG-4323; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_205_ListDefaultLimitIsSmall; UPDATED=2026-08-28
// defaultListLimit caps list output to protect agent context. Deliberately
// small; pass --limit -1 to explicitly request everything.
const defaultListLimit = 20

func init() {
	ListCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	ListCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	ListCmd.Flags().String("status", "", "filter by status (STUB, IMPL, TESTED, BENCHED)")
	ListCmd.Flags().String("aspect", "", "filter by aspect (API, CLI, Engine, etc.)")
	ListCmd.Flags().String("phase", "", "filter by phase (Phase0, Phase1, Phase2, Phase3)")
	ListCmd.Flags().String("owner", "", "filter by owner")
	ListCmd.Flags().String("spec-status", "", "filter by spec status (draft, approved, in-progress, completed, archived)")
	ListCmd.Flags().Int("priority-min", 0, "filter by minimum priority (0 = no minimum)")
	ListCmd.Flags().Int("priority-max", 0, "filter by maximum priority (0 = no maximum)")
	ListCmd.Flags().String("order-by", "", "custom ORDER BY clause (default: priority ASC, updated_at DESC)")
	ListCmd.Flags().Int("limit", defaultListLimit, "maximum number of results (default 20 to protect agent context; -1 = unlimited)")
	ListCmd.Flags().Bool("json", false, "output as JSON")
	ListCmd.Flags().Bool("include-hidden", false, "include hidden requirements (test files, templates, examples)")
}
