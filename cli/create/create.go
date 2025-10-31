package create

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/reqid"
)

// CreateCmd generates a new CANARY token template
var CreateCmd = &cobra.Command{
	Use:   "create <req-id> <feature-name>",
	Short: "Generate a new CANARY token template",
	Long: `Create a properly formatted CANARY token for a new requirement.

Example:
  canary create CBIN-CLI-105 "UserProfile" --aspect CLI --status IMPL

Outputs a ready-to-paste CANARY token comment.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		reqID := args[0]
		feature := args[1]

		aspect, _ := cmd.Flags().GetString("aspect")
		status, _ := cmd.Flags().GetString("status")
		owner, _ := cmd.Flags().GetString("owner")
		test, _ := cmd.Flags().GetString("test")
		bench, _ := cmd.Flags().GetString("bench")

		// Validate aspect
		if err := reqid.ValidateAspect(aspect); err != nil {
			return fmt.Errorf("invalid aspect: %w", err)
		}

		// Normalize aspect to canonical form
		aspect = reqid.NormalizeAspect(aspect)

		// Get today's date
		today := time.Now().UTC().Format("2006-01-02")

		// Build token
		token := fmt.Sprintf("// CANARY: REQ=%s; FEATURE=\"%s\"; ASPECT=%s; STATUS=%s",
			reqID, feature, aspect, status)

		if test != "" {
			token += fmt.Sprintf("; TEST=%s", test)
		}
		if bench != "" {
			token += fmt.Sprintf("; BENCH=%s", bench)
		}
		if owner != "" {
			token += fmt.Sprintf("; OWNER=%s", owner)
		}

		token += fmt.Sprintf("; UPDATED=%s", today)

		fmt.Println(token)
		fmt.Println("\n// Paste this above your implementation:")
		fmt.Printf("// func %s() { ... }\n", feature)

		return nil
	},
}
