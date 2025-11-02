// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

var bugCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new BUG-* CANARY token",
	Long: `Create a new BUG-* CANARY token for tracking a defect.

The bug ID will be automatically generated in the format BUG-ASPECT-XXX.

Examples:
  canary bug create "Login fails on first attempt"
  canary bug create "API returns 500 error" --aspect API --severity S1 --priority P0
  canary bug create "Memory leak in storage engine" --aspect Storage --file src/storage/cache.go:42`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		title := strings.Join(args, " ")
		aspect, _ := cmd.Flags().GetString("aspect")
		severity, _ := cmd.Flags().GetString("severity")
		priority, _ := cmd.Flags().GetString("priority")
		status, _ := cmd.Flags().GetString("status")
		file, _ := cmd.Flags().GetString("file")
		owner, _ := cmd.Flags().GetString("owner")
		dbPath, _ := cmd.Flags().GetString("db")

		// Default values
		if aspect == "" {
			aspect = "API" // Default aspect
		}
		if status == "" {
			status = "OPEN"
		}
		if severity == "" {
			severity = "S3"
		}
		if priority == "" {
			priority = "P2"
		}

		// Generate bug ID
		bugID, err := generateBugID(aspect, dbPath)
		if err != nil {
			return fmt.Errorf("generate bug ID: %w", err)
		}

		// Parse file location if provided
		var filePath string
		var lineNum int
		if file != "" {
			parts := strings.Split(file, ":")
			filePath = parts[0]
			if len(parts) > 1 {
				lineNum, _ = strconv.Atoi(parts[1])
			}
		} else {
			// Default to main.go or most relevant file
			filePath = "main.go"
			lineNum = 1
		}

		// Create token
		token := &storage.Token{
			ReqID:      bugID,
			Feature:    title,
			Aspect:     aspect,
			Status:     status,
			FilePath:   filePath,
			LineNumber: lineNum,
			UpdatedAt:  time.Now().Format("2006-01-02"),
			Owner:      owner,
			Priority:   parsePriorityValue(priority),
			Keywords:   fmt.Sprintf("SEVERITY=%s;PRIORITY=%s", severity, priority),
		}

		// Save to database
		db, err := storage.Open(dbPath)
		if err != nil {
			// Create CANARY comment in file if no database
			return createBugCanaryComment(token, severity, priority)
		}
		defer db.Close()

		if err := db.UpsertToken(token); err != nil {
			return fmt.Errorf("save bug token: %w", err)
		}

		// Generate CANARY comment format
		canaryComment := fmt.Sprintf(
			"// CANARY: BUG=%s; TITLE=\"%s\";\n"+
				"//         ASPECT=%s; STATUS=%s;\n"+
				"//         SEVERITY=%s; PRIORITY=%s;\n"+
				"//         UPDATED=%s",
			bugID, title, aspect, status,
			severity, priority,
			time.Now().Format("2006-01-02"),
		)

		fmt.Printf("? Created bug token: %s\n", bugID)
		fmt.Printf("?? Title: %s\n", title)
		fmt.Printf("?? Severity: %s | Priority: %s\n", severity, priority)
		fmt.Printf("?? Location: %s:%d\n", filePath, lineNum)
		fmt.Printf("\n%s CANARY comment to add:\n", color.YellowString("?"))
		fmt.Println(canaryComment)

		return nil
	},
}

func init() {
	bugCreateCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	bugCreateCmd.Flags().String("aspect", "", "Bug aspect (API, CLI, Engine, Storage, etc.)")
	bugCreateCmd.Flags().String("severity", "S3", "Severity level (S1-Critical, S2-High, S3-Medium, S4-Low)")
	bugCreateCmd.Flags().String("priority", "P2", "Priority level (P0, P1, P2, P3)")
	bugCreateCmd.Flags().String("status", "OPEN", "Initial status")
	bugCreateCmd.Flags().String("file", "", "File and line number (e.g., src/api/handler.go:42)")
	bugCreateCmd.Flags().String("owner", "", "Bug owner/assignee")
	bugCreateCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
