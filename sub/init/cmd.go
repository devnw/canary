package init

import (
	"log/slog"
	"sync"

	"github.com/spf13/cobra"
	"go.spyder.org/gen/cli"
)

var Cmd = &cobra.Command{Use: "init [flags]", Short: "initialize a new CANARY configuration", RunE: Run}

func init() {
	if err := cli.SetArgs(Cmd.Flags(), "", Args()); err != nil {
		slog.Error("error setting arguments", "error", err, "args", Args())
		return
	}
}

func Run(cmd *cobra.Command, args []string) error { return nil }

//nolint:gochecknoglobals // this is expected to setup the cli args
var once sync.Once

//nolint:gochecknoglobals // this is expected to setup the cli args
var args cli.Args

func Args() cli.Args { once.Do(func() { args = cli.Args{} }); return args }
