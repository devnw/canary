package checkpoint

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-128; FEATURE="CheckpointCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var CheckpointCmd = &cobra.Command{
	Use:   "checkpoint <name> [description]",
	Short: "Create a state snapshot checkpoint",
	Long: `Create a checkpoint to capture current state of all tokens.

Checkpoints include:
- Counts by status (STUB, IMPL, TESTED, BENCHED)
- Commit hash and timestamp
- Full JSON snapshot of all tokens

Useful for tracking progress over time.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		dbPath, _ := cmd.Flags().GetString("db")
		name := args[0]
		description := ""
		if len(args) > 1 {
			description = strings.Join(args[1:], " ")
		}

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		// Get current commit hash
		commitHash := ""
		if gitCmd := exec.Command("git", "rev-parse", "HEAD"); gitCmd.Dir == "" {
			if output, err := gitCmd.Output(); err == nil {
				commitHash = strings.TrimSpace(string(output))
			}
		}

		// Load project config for ID pattern filtering
		cfg, _ := utils.LoadProjectConfig()
		idPattern := ""
		if cfg != nil && cfg.Requirements.IDPattern != "" {
			idPattern = cfg.Requirements.IDPattern
		}

		// Get all tokens for snapshot
		tokens, err := db.ListTokens(nil, idPattern, "", 0)
		if err != nil {
			return fmt.Errorf("get tokens: %w", err)
		}

		snapshotJSON, err := json.Marshal(tokens)
		if err != nil {
			return fmt.Errorf("marshal snapshot: %w", err)
		}

		if err := db.CreateCheckpoint(name, description, commitHash, string(snapshotJSON)); err != nil {
			return fmt.Errorf("create checkpoint: %w", err)
		}

		fmt.Printf("✅ Created checkpoint: %s\n", name)
		if commitHash != "" {
			fmt.Printf("Commit: %s\n", commitHash[:8])
		}
		fmt.Printf("Tokens: %d\n", len(tokens))

		return nil
	},
}

func init() {
	CheckpointCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	CheckpointCmd.Flags().String("db", ".canary/canary.db", "path to database file")
}
