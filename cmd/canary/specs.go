// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// CANARY: REQ=CBIN-145; FEATURE="SpecsCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_SpecsCmd; UPDATED=2025-10-17
var specsCmd = &cobra.Command{
	Use:   "specs [flags]",
	Short: "List all requirement specification directories",
	Long: `Specs lists all requirement specification directories in .canary/specs/.

Shows requirement ID, feature name (extracted from directory name), and paths
to spec.md and plan.md files if they exist.

Examples:
  canary specs
  canary specs --path .canary/specs
  canary specs --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		specsPath, _ := cmd.Flags().GetString("path")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// Check if specs directory exists
		if _, err := os.Stat(specsPath); os.IsNotExist(err) {
			return fmt.Errorf("specs directory not found: %s", specsPath)
		}

		// Read specs directory
		entries, err := os.ReadDir(specsPath)
		if err != nil {
			return fmt.Errorf("read specs directory: %w", err)
		}

		// Collect spec information
		type SpecInfo struct {
			ReqID       string `json:"req_id"`
			FeatureName string `json:"feature_name"`
			Directory   string `json:"directory"`
			HasSpec     bool   `json:"has_spec"`
			HasPlan     bool   `json:"has_plan"`
		}

		var specs []SpecInfo

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirName := entry.Name()
			dirPath := filepath.Join(specsPath, dirName)

			// Parse directory name (format: CBIN-XXX-feature-name)
			parts := strings.SplitN(dirName, "-", 3)
			reqID := dirName
			featureName := ""
			if len(parts) >= 3 {
				reqID = parts[0] + "-" + parts[1]
				featureName = strings.ReplaceAll(parts[2], "-", " ")
				featureName = strings.Title(featureName)
			}

			// Check for spec.md and plan.md
			specPath := filepath.Join(dirPath, "spec.md")
			planPath := filepath.Join(dirPath, "plan.md")

			hasSpec := false
			hasPlan := false

			if _, err := os.Stat(specPath); err == nil {
				hasSpec = true
			}
			if _, err := os.Stat(planPath); err == nil {
				hasPlan = true
			}

			specs = append(specs, SpecInfo{
				ReqID:       reqID,
				FeatureName: featureName,
				Directory:   dirPath,
				HasSpec:     hasSpec,
				HasPlan:     hasPlan,
			})
		}

		// Sort by requirement ID
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].ReqID < specs[j].ReqID
		})

		if len(specs) == 0 {
			fmt.Printf("No specification directories found in %s\n", specsPath)
			return nil
		}

		// Output
		if jsonOutput {
			// JSON output
			fmt.Println("[")
			for i, spec := range specs {
				comma := ","
				if i == len(specs)-1 {
					comma = ""
				}
				fmt.Printf("  {\"req_id\": \"%s\", \"feature_name\": \"%s\", \"directory\": \"%s\", \"has_spec\": %t, \"has_plan\": %t}%s\n",
					spec.ReqID, spec.FeatureName, spec.Directory, spec.HasSpec, spec.HasPlan, comma)
			}
			fmt.Println("]")
		} else {
			// Human-readable output
			fmt.Printf("Found %d specification directories:\n\n", len(specs))

			for _, spec := range specs {
				fmt.Printf("📁 %s", spec.ReqID)
				if spec.FeatureName != "" {
					fmt.Printf(" - %s", spec.FeatureName)
				}
				fmt.Println()
				fmt.Printf("   %s\n", spec.Directory)

				files := []string{}
				if spec.HasSpec {
					files = append(files, "spec.md")
				}
				if spec.HasPlan {
					files = append(files, "plan.md")
				}
				if len(files) > 0 {
					fmt.Printf("   Files: %s\n", strings.Join(files, ", "))
				} else {
					fmt.Printf("   (no spec or plan files)\n")
				}
				fmt.Println()
			}

			fmt.Printf("Total: %d specifications\n", len(specs))
		}

		return nil
	},
}

func init() {
	specsCmd.Flags().String("path", ".canary/specs", "Path to specs directory")
	specsCmd.Flags().Bool("json", false, "Output as JSON")
}
