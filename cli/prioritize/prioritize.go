package prioritize

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-127; FEATURE="PrioritizeCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var PrioritizeCmd = &cobra.Command{
	Use:   "prioritize <REQ-ID> <feature> <priority>",
	Short: "Update priority of a CANARY token",
	Long: `Update the priority of a specific token (1=highest, 10=lowest).

Priority affects ordering in list and search results.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		reqID := args[0]
		feature := args[1]
		priority, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid priority: %s (must be 1-10)", args[2])
		}

		if priority < 1 || priority > 10 {
			return fmt.Errorf("priority must be between 1 (highest) and 10 (lowest)")
		}

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		if err := db.UpdatePriority(reqID, feature, priority); err != nil {
			return fmt.Errorf("update priority: %w", err)
		}

		fmt.Printf("✅ Updated priority for %s/%s to %d\n", reqID, feature, priority)
		return nil
	},
}
