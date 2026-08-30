// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"context"
	"errors"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// runMCP runs `canary mcp ...` in root with a hard deadline, so a command
// that is supposed to refuse to start can never hang this suite waiting for
// a server that did start.
func runMCP(t *testing.T, root, bin string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, append([]string{"mcp"}, args...)...)
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(cmd.Environ(), "HOME="+home, "USERPROFILE="+home)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("canary mcp %s did not exit within the deadline; out=%s", strings.Join(args, " "), out)
	}
	return string(out), err
}

// TestAuditF09 proves the MCP server fails at the bind rather than after it.
//
// The old command printed a success banner, then called ListenAndServe inside
// a goroutine and blocked on a context that nothing ever cancelled. A port
// that was already taken therefore produced a "Server listening on ..."
// banner, an error line nobody was watching for, and a process that sat there
// forever -- eventually exiting 0. Every one of those is a lie about whether
// the server is reachable.
func TestAuditF09(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	// A port already held by this test: the server cannot have it.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	port := ln.Addr().(*net.TCPAddr).Port

	out, err := runMCP(t, root, bin, "--port", strconv.Itoa(port))
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("duplicate bind must exit nonzero; err=%v out=%s", err, out)
	}
	if strings.Contains(out, "Server listening") {
		t.Fatalf("a failed bind must not announce a listening server; out=%s", out)
	}

	// Non-loopback exposure without TLS and a token is refused before any
	// listener exists: an unauthenticated MCP server on 0.0.0.0 hands every
	// tool on this box to the network.
	out, err = runMCP(t, root, bin, "--host", "0.0.0.0", "--port", "0")
	if err == nil {
		t.Fatalf("non-loopback without TLS+token must refuse to start; out=%s", out)
	}
	if !strings.Contains(out, "--tls-cert") || !strings.Contains(out, "CANARY_MCP_TOKEN") {
		t.Fatalf("refusal must name what is missing (--tls-cert/--tls-key, CANARY_MCP_TOKEN); out=%s", out)
	}
	if strings.Contains(out, "Server listening") {
		t.Fatalf("a refused start must not announce a listening server; out=%s", out)
	}

	// The default bind is loopback, not every interface.
	help, err := runMCP(t, root, bin, "--help")
	if err != nil {
		t.Fatalf("canary mcp --help: %v\n%s", err, help)
	}
	if !strings.Contains(help, "127.0.0.1") {
		t.Fatalf("--host must default to 127.0.0.1; help=%s", help)
	}
	for _, flag := range []string{"--tls-cert", "--tls-key", "--root", "--print-tools"} {
		if !strings.Contains(help, flag) {
			t.Errorf("mcp command is missing %s", flag)
		}
	}
}

// TestAuditF09PrintToolsExitsZero proves the documented regeneration command
// is a real command: it prints the tool documentation and exits, without
// binding anything.
func TestAuditF09PrintTools(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	out, err := runMCP(t, root, bin, "--print-tools")
	if err != nil {
		t.Fatalf("canary mcp --print-tools: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Canary MCP Tools") {
		t.Fatalf("--print-tools produced no tool documentation; out=%s", out)
	}
	if strings.Contains(out, "Server listening") {
		t.Fatalf("--print-tools must not start a server; out=%s", out)
	}
}
