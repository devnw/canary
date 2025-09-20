package scan

import (
	"log/slog"
	"regexp"
	"sync"

	"github.com/spf13/cobra"
	"go.codepros.org/canary/internal/cli/exitcodes"
	"go.codepros.org/canary/internal/core"
	"go.spyder.org/gen/cli"
)

// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20
var Cmd = &cobra.Command{Use: "scan", Short: "scan the given paths for CANARY tokens", RunE: Run}

func init() {
	if err := cli.SetArgs(Cmd.Flags(), "", Args()); err != nil {
		slog.Error("error setting arguments", "error", err, "args", Args())
		return
	}
	f := Cmd.Flags()
	f.String("root", ".", "root directory to scan")
	f.String("out", "status.json", "output status.json path")
	f.String("csv", "", "optional status.csv output path")
	f.String("skip", core.DefaultSkipPattern, "RE2 regex of paths to skip")
}

func Run(cmd *cobra.Command, args []string) error {
	root, _ := cmd.Flags().GetString("root")
	out, _ := cmd.Flags().GetString("out")
	csv, _ := cmd.Flags().GetString("csv")
	skip, _ := cmd.Flags().GetString("skip")
	re, err := regexp.Compile(skip)
	if err != nil {
		return exitcodes.New(exitcodes.CodeParseIO, err.Error())
	}
	rep, err := core.Scan(core.ScanOptions{Root: root, Skip: re})
	if err != nil {
		return exitcodes.New(exitcodes.CodeParseIO, err.Error())
	}
	if err := core.ComposeAndWrite(rep, out, csv); err != nil {
		return exitcodes.New(exitcodes.CodeParseIO, err.Error())
	}
	slog.Info("scan complete", "requirements", len(rep.Requirements), "tokens", rep.Summary.TotalTokens)
	return nil
}

//nolint:gochecknoglobals
var once sync.Once

//nolint:gochecknoglobals
var args cli.Args

func Args() cli.Args {
	once.Do(func() { args = cli.Args{} })
	return args
}
