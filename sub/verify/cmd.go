package verify

import (
	"log/slog"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"go.codepros.org/canary/internal/cli/exitcodes"
	"go.codepros.org/canary/internal/core"
)

// CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20
var Cmd = &cobra.Command{Use: "verify", Short: "verify GAP claims & optional staleness", RunE: Run}

func init() {
	f := Cmd.Flags()
	f.String("root", ".", "root directory to scan before verification")
	f.String("gap", "GAP_ANALYSIS.md", "gap analysis file to verify")
	f.Bool("strict", false, "enable 30 day staleness check")
	f.String("skip", core.DefaultSkipPattern, "RE2 regex of paths to skip")
}

func Run(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	gapPath, _ := cmd.Flags().GetString("gap")
	strict, _ := cmd.Flags().GetBool("strict")
	skip, _ := cmd.Flags().GetString("skip")
	re, err := regexp.Compile(skip)
	if err != nil {
		return err
	}
	rep, err := core.Scan(core.ScanOptions{Root: root, Skip: re})
	if err != nil {
		return err
	}
	gapContent, err := core.LoadFile(gapPath)
	if err != nil {
		return err
	}
	verifyDiags, _ := core.VerifyClaims(rep, gapContent)
	staleDiags := []string{}
	if strict {
		staleDiags, err = core.CheckStaleness(rep, 30*24*time.Hour)
		if err != nil {
			return err
		}
	}
	for _, d := range verifyDiags {
		slog.Error(d)
	}
	for _, d := range staleDiags {
		slog.Error(d)
	}
	if len(verifyDiags) > 0 || (strict && len(staleDiags) > 0) {
		return exitcodes.New(exitcodes.CodeVerify, "verification/staleness failure")
	}
	slog.Info("verify ok", "requirements", len(rep.Requirements))
	return nil
}

// removed external cli arg reflection dependency
