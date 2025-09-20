package update

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.spyder.org/gen/cli"
)

var Cmd = &cobra.Command{Use: "update", Short: "update first matching CANARY token fields", RunE: Run}

func init() {
	if err := cli.SetArgs(Cmd.Flags(), "", Args()); err != nil {
		slog.Error("error setting arguments", "error", err, "args", Args())
		return
	}
	f := Cmd.Flags()
	f.String("file", "", "file containing token (required)")
	f.String("req", "", "requirement id to match (required)")
	f.String("status", "", "new status (optional)")
	f.String("updated", time.Now().UTC().Format("2006-01-02"), "UPDATED date (default today)")
	f.String("addtest", "", "append test name (optional)")
	f.String("addbench", "", "append benchmark name (optional)")
}

var tokenLineRe = regexp.MustCompile(`(?m)^(?P<prefix>\s*(?://|#|--)\s*CANARY:)(?P<body>.*)$`)

func Run(cmd *cobra.Command, args []string) error {
	fg := cmd.Flags()
	file, _ := fg.GetString("file")
	req, _ := fg.GetString("req")
	newStatus, _ := fg.GetString("status")
	updated, _ := fg.GetString("updated")
	addTest, _ := fg.GetString("addtest")
	addBench, _ := fg.GetString("addbench")
	if file == "" || req == "" {
		return errors.New("file, req required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	changed := false
	for i, l := range lines {
		if !strings.Contains(l, "CANARY:") {
			continue
		}
		if !strings.Contains(l, req) {
			continue
		}
		m := tokenLineRe.FindStringSubmatch(l)
		if len(m) == 0 {
			continue
		}
		body := m[2]
		// mutate key/value segments
		segs := strings.Split(body, ";")
		kv := map[string]string{}
		order := []string{}
		for _, seg := range segs {
			seg = strings.TrimSpace(seg)
			if seg == "" {
				continue
			}
			parts := strings.SplitN(seg, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.ToUpper(strings.TrimSpace(parts[0]))
			v := strings.TrimSpace(parts[1])
			kv[k] = v
			order = append(order, k)
		}
		if newStatus != "" {
			kv["STATUS"] = newStatus
		}
		if updated != "" {
			kv["UPDATED"] = updated
		}
		appendList := func(field, val string) {
			if val == "" {
				return
			}
			if existing, ok := kv[field]; ok && existing != "" {
				kv[field] = existing + "," + val
			} else {
				kv[field] = val
			}
		}
		appendList("TEST", addTest)
		appendList("BENCH", addBench)
		// rebuild preserving order + ensure critical fields at end if newly added
		ensure := []string{"REQ", "FEATURE", "ASPECT", "STATUS", "TEST", "BENCH", "OWNER", "UPDATED"}
		seen := map[string]bool{}
		var rebuilt []string
		for _, k := range order {
			if v, ok := kv[k]; ok {
				rebuilt = append(rebuilt, fmt.Sprintf("%s=%s", k, v))
				seen[k] = true
			}
		}
		for _, k := range ensure {
			if !seen[k] {
				if v, ok := kv[k]; ok && v != "" {
					rebuilt = append(rebuilt, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
		lines[i] = m[1] + " " + strings.Join(rebuilt, "; ")
		changed = true
		break
	}
	if !changed {
		return fmt.Errorf("no matching token for %s in %s", req, file)
	}
	out := strings.Join(lines, "\n")
	if err := os.WriteFile(file, []byte(out), 0o644); err != nil {
		return err
	}
	slog.Info("token updated", "file", file, "req", req, "status", newStatus)
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
