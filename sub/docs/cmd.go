package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"go.codepros.org/canary/internal/core"
	"go.codepros.org/canary/prompts"
	"go.spyder.org/gen/cli"
)

// Cmd docs generate command (local template hydration only; AI call optional future).
var Cmd = &cobra.Command{Use: "docs", Short: "generate documentation from embedded prompts", RunE: run}

func init() {
    if err := cli.SetArgs(Cmd.Flags(), "", Args()); err != nil { return }
    f := Cmd.Flags()
    f.String("out", "docs/generated", "output directory for generated docs")
    f.Bool("print", false, "print generated outputs to stdout instead of writing files")
    f.String("root", ".", "root to scan for status.json generation")
    f.String("status", "", "optional existing status.json path (skip scan)")
    f.String("model", "", "LLM model (optional; requires OPENAI_API_KEY)")
}

func run(cmd *cobra.Command, _ []string) error {
    outDir, _ := cmd.Flags().GetString("out")
    toStdout, _ := cmd.Flags().GetBool("print")
    root, _ := cmd.Flags().GetString("root")
    statusPath, _ := cmd.Flags().GetString("status")
    model, _ := cmd.Flags().GetString("model")
    if !toStdout { if err := os.MkdirAll(outDir, 0o755); err != nil { return err } }

    var rep core.Report
    var err error
    if statusPath != "" {
        b, rerr := os.ReadFile(statusPath)
        if rerr != nil { return rerr }
        if uerr := json.Unmarshal(b, &rep); uerr != nil { return uerr }
    } else {
        rep, err = core.Scan(core.ScanOptions{Root: root})
        if err != nil { return err }
    }

    summaries := map[string]string{}
    if model != "" {
        // Lazy summarization placeholder (model integration would go here)
        summaries["status"] = fmt.Sprintf("Status summary: %d requirements, %d tokens", rep.Summary.UniqueRequirements, rep.Summary.TotalTokens)
    }

    data := map[string]any{"Prompts": prompts.All(), "Status": rep, "Generated": time.Now().UTC().Format(time.RFC3339), "Summaries": summaries}
    indexT := `# Generated Prompt Index\n\nGenerated: {{.Generated}}\n\n{{range $k,$v := .Prompts}}## {{$k}}\n\n<details><summary>show content</summary>\n\n{{$v}}\n\n</details>\n\n{{end}}\n---\n## Status Summary\nTotal Requirements: {{.Status.Summary.UniqueRequirements}}  Tokens: {{.Status.Summary.TotalTokens}}\n`
    t, err := template.New("index").Parse(indexT)
    if err != nil { return err }
    outputs := map[string]string{}
    for name, content := range prompts.All() { outputs[name+".md"] = content }
    var buf []byte
    b := &bytesBuffer{buf: &buf}
    if err := t.Execute(b, data); err != nil { return err }
    outputs["INDEX.md"] = string(b.Bytes())
    bjson, _ := json.Marshal(rep)
    outputs["status.json"] = string(bjson)
    if toStdout {
        for k, v := range outputs { fmt.Printf("--- %s ---\n%s\n", k, v) }
        return nil
    }
    for name, content := range outputs { if err := os.WriteFile(filepath.Join(outDir, name), []byte(content), 0o644); err != nil { return err } }
    fmt.Fprintf(os.Stderr, "generated %d docs -> %s\n", len(outputs), outDir)
    return nil
}

// minimal buffer wrapper to avoid importing bytes unnecessarily (small).
type bytesBuffer struct { buf *[]byte }
func (b *bytesBuffer) Write(p []byte) (int, error) { *b.buf = append(*b.buf, p...); return len(p), nil }
func (b *bytesBuffer) Bytes() []byte { return *b.buf }

// args placeholder.
var args cli.Args
func Args() cli.Args { return args }
