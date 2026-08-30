// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitAgentFilesAreSubstituted covers the F-01 regression: the unguarded
// bulk copy of base/ used to land the raw agent templates in .canary/agents/
// before the guarded writer could author the template-substituted ones, so the
// guarded writer saw a differing file, refused it, and a fresh `canary init`
// shipped agent definitions still carrying `{{ .AgentPrefix }}` -- plus a
// "kept existing" notice for a file the user had never seen.
func TestInitAgentFilesAreSubstituted(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	out := run(t, root, bin, "init", "--key", "CBIN")

	agentsDir := filepath.Join(root, ".canary", "agents")
	seen := 0
	err := filepath.WalkDir(agentsDir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		seen++
		b, rerr := os.ReadFile(path) //nolint:gosec // test-controlled path
		if rerr != nil {
			return rerr
		}
		if bytes.Contains(b, []byte("{{")) {
			t.Errorf("%s still holds an unsubstituted template placeholder", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", agentsDir, err)
	}
	if seen == 0 {
		t.Fatalf("no agent files were written to %s", agentsDir)
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "kept existing") && strings.Contains(line, "agents") {
			t.Errorf("fresh init reported a kept-existing notice for a file it just created: %s", line)
		}
	}
}

// TestInitAgentContextPreserved covers the F-02 regression: the guarded write
// of .canary/AGENT_CONTEXT.md was vacuous because the unguarded bulk copy had
// already clobbered the user's customized copy earlier in the same run.
func TestInitAgentContextPreserved(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	run(t, root, bin, "init", "--key", "CBIN")

	ctx := filepath.Join(root, ".canary", "AGENT_CONTEXT.md")
	custom := []byte("# my customized agent context\n")
	if err := os.WriteFile(ctx, custom, 0o644); err != nil {
		t.Fatalf("write custom AGENT_CONTEXT.md: %v", err)
	}

	out := run(t, root, bin, "init", "--key", "CBIN")

	got, err := os.ReadFile(ctx) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read AGENT_CONTEXT.md: %v", err)
	}
	if !bytes.Equal(got, custom) {
		t.Fatalf("re-init clobbered the customized AGENT_CONTEXT.md without --force:\n%s", got)
	}
	if !strings.Contains(out, "kept existing") || !strings.Contains(out, "AGENT_CONTEXT.md") {
		t.Fatalf("re-init kept AGENT_CONTEXT.md silently; output:\n%s", out)
	}
}

// TestInitAgentContextForce proves --force is still the escape hatch for
// .canary/AGENT_CONTEXT.md, and that it keeps the prior bytes in a .bak.
func TestInitAgentContextForce(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	run(t, root, bin, "init", "--key", "CBIN")

	ctx := filepath.Join(root, ".canary", "AGENT_CONTEXT.md")
	custom := []byte("# my customized agent context\n")
	if err := os.WriteFile(ctx, custom, 0o644); err != nil {
		t.Fatalf("write custom AGENT_CONTEXT.md: %v", err)
	}

	run(t, root, bin, "init", "--key", "CBIN", "--force")

	got, err := os.ReadFile(ctx) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read AGENT_CONTEXT.md: %v", err)
	}
	if bytes.Equal(got, custom) {
		t.Fatal("--force did not overwrite the customized AGENT_CONTEXT.md")
	}
	bak, err := os.ReadFile(ctx + ".bak") //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read AGENT_CONTEXT.md.bak: %v", err)
	}
	if !bytes.Equal(bak, custom) {
		t.Fatalf("backup = %q, want the pre-overwrite bytes", bak)
	}
}
