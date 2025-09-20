package report

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{Use: "report", Short: "print summary from status.json", RunE: Run}

func init() {
	f := Cmd.Flags()
	f.String("input", "status.json", "status.json to summarize")
}

type statusFile struct {
	Summary any `json:"summary"`
}

func Run(cmd *cobra.Command, args []string) error {
	in := cmd.Flags().Lookup("input").Value.String()
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	var sf statusFile
	if err := json.Unmarshal(b, &sf); err != nil {
		return errors.New("invalid status.json: " + err.Error())
	}
	out, _ := json.Marshal(map[string]any{"summary": sf.Summary})
	os.Stdout.Write(out)
	os.Stdout.Write([]byte("\n"))
	return nil
}

// removed external cli arg reflection dependency
