package create

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.spyder.org/gen/cli"
)

var Cmd = &cobra.Command{Use: "create", Short: "create new CANARY token in a file (prepends line)", RunE: Run}

func init() {
	if err := cli.SetArgs(Cmd.Flags(), "", Args()); err != nil {
		slog.Error("error setting arguments", "error", err, "args", Args())
		return
	}
	f := Cmd.Flags()
	f.String("file", "", "target file to prepend token (required)")
	f.String("req", "", "requirement id e.g. CBIN-101 (required)")
	f.String("feature", "", "feature name (required)")
	f.String("aspect", "API", "aspect enum")
	f.String("status", "STUB", "initial status")
	f.String("test", "", "optional test name(s) comma separated")
	f.String("bench", "", "optional bench name(s) comma separated")
	f.String("owner", "", "owner/team alias")
	f.String("updated", time.Now().UTC().Format("2006-01-02"), "UPDATED date (YYYY-MM-DD)")
}

func Run(cmd *cobra.Command, args []string) error {
	fg := cmd.Flags()
	file, _ := fg.GetString("file")
	req, _ := fg.GetString("req")
	feature, _ := fg.GetString("feature")
	aspect, _ := fg.GetString("aspect")
	status, _ := fg.GetString("status")
	test, _ := fg.GetString("test")
	bench, _ := fg.GetString("bench")
	owner, _ := fg.GetString("owner")
	updated, _ := fg.GetString("updated")
	if file == "" || req == "" || feature == "" {
		return errors.New("file, req, feature required")
	}
	line := buildTokenLine(file, req, feature, aspect, status, test, bench, owner, updated)
	if err := prependLine(file, line+"\n"); err != nil {
		return err
	}
	slog.Info("token created", "file", file, "req", req)
	return nil
}

func buildTokenLine(file, req, feature, aspect, status, test, bench, owner, updated string) string {
	// normalize req zero pad if CBIN-n
	if m := regexp.MustCompile(`^(CBIN-)(\d{1,3})$`).FindStringSubmatch(req); len(m) == 3 {
		n := m[2]
		for len(n) < 3 {
			n = "0" + n
		}
		req = m[1] + n
	}
	var parts []string
	add := func(k, v string) {
		if v != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}
	add("REQ", req)
	add("FEATURE", fmt.Sprintf("\"%s\"", feature))
	add("ASPECT", aspect)
	add("STATUS", status)
	if test != "" {
		add("TEST", test)
	}
	if bench != "" {
		add("BENCH", bench)
	}
	if owner != "" {
		add("OWNER", owner)
	}
	if updated != "" {
		add("UPDATED", updated)
	}
	return commentPrefix(file) + " CANARY: " + strings.Join(parts, "; ")
}

func commentPrefix(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".sh", ".py", ".rb", ".yml", ".yaml", ".md", ".txt":
		return "#"
	case ".sql":
		return "--"
	default:
		return "//"
	}
}

func prependLine(path, line string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// avoid duplicate if already present for same REQ
	if strings.Contains(string(b), line) {
		return fmt.Errorf("token already present")
	}
	// keep shebang if present
	if len(b) > 2 && b[0] == '#' && b[1] == '!' { // shebang line preserving
		sc := bufio.NewScanner(strings.NewReader(string(b)))
		if sc.Scan() {
			first := sc.Text()
			rest := sc.Bytes()
			content := first + "\n" + line + string(rest) + "\n"
			return os.WriteFile(path, []byte(content), 0o644)
		}
	}
	return os.WriteFile(path, append([]byte(line), b...), 0o644)
}

//nolint:gochecknoglobals // cli args singleton
var once sync.Once

//nolint:gochecknoglobals // cli args singleton
var args cli.Args

func Args() cli.Args {
	once.Do(func() {
		args = cli.Args{}
	})
	return args
}
