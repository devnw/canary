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

	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/storage"
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
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
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

		projectID, err := utils.WriteProjectID(cmd, ".")
		if err != nil {
			return err
		}

		// The id and the row are allocated through the same handle, so the
		// number the caller is told about is the number that was written.
		// Mutating command: OpenRW may create and migrate.
		db, err := storage.OpenRW(dbPath)
		var bugID string
		if err == nil {
			defer db.Close()
			// ReserveID is one immediate transaction whose primary key
			// rejects a duplicate. The read-max-then-increment generator it
			// replaces had a gap between the read and the write, so two `bug
			// create` runs at the same moment computed the same id and the
			// second silently overwrote the first's row.
			bugID, err = db.ReserveID(projectID, "BUG-"+strings.ToUpper(aspect))
			if err != nil {
				return fmt.Errorf("reserve bug ID: %w", err)
			}
		} else {
			// No index to reserve from. The bug still gets a first-in-series
			// id so the CANARY comment below is complete; the next `canary
			// index` reads the comment and mints the row.
			bugID = fmt.Sprintf("BUG-%s-001", strings.ToUpper(aspect))
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

		token.ProjectID = projectID

		if db == nil {
			// No database to reserve from or write to: emit the CANARY
			// comment so the bug still exists in source, and let the next
			// `canary index` mint the row.
			return createBugCanaryComment(token, severity, priority)
		}

		if err := db.UpsertToken(token); err != nil {
			return fmt.Errorf("save bug token: %w", err)
		}

		// Generate CANARY comment format
		canaryComment := buildBugToken(bugID, title, title, aspect, status, severity, priority, time.Now().UTC().Format("2006-01-02"))

		fmt.Printf("? Created bug token: %s\n", bugID)
		fmt.Printf("?? Title: %s\n", title)
		fmt.Printf("?? Severity: %s | Priority: %s\n", severity, priority)
		fmt.Printf("?? Location: %s:%d\n", filePath, lineNum)
		fmt.Printf("\n%s CANARY comment to add:\n", color.YellowString("?"))
		fmt.Println(canaryComment)
		fmt.Println("Note: add this comment to your source file — the DB row is rebuilt from source on the next 'canary index'.")

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
	utils.AddProjectFlag(bugCreateCmd)
}
