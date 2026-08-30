// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Server limits. Every one of these was absent before: the server had a
// 10-second header timeout and nothing else, so a client could hold a
// connection open indefinitely mid-body, and the only cap on a request was
// the SDK's 4MiB default.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 30 * time.Second
	maxHeaderBytes    = 16 << 10
	// maxBodyBytes caps a request body. Tool arguments are small; a megabyte
	// is already generous, and an unbounded body is a memory-exhaustion
	// channel open to anything that can reach the port.
	maxBodyBytes = 1 << 20
	// shutdownGrace is how long in-flight requests get to finish after a
	// signal.
	shutdownGrace = 5 * time.Second
)

// CANARY: REQ=CP-281; FEATURE="MCPServerCommand"; ASPECT=Wire; STATUS=TESTED; TEST=TestMCPCommandCreation,TestAuditF09,TestAuditF09PrintTools,TestHTTPServerLimits,TestTLSFlagPairIsBothOrNeither,TestRunServerRefusesHalfConfiguredTLS; UPDATED=2026-08-30

// New returns the MCP subcommand for Canary.
//
// version is the binary's version (ldflags), reported to clients as the
// server implementation version. It used to be the string "1.0.0", which
// meant every build in the field identified itself identically.
func New(version string) *cobra.Command {
	var (
		port       int
		host       string
		root       string
		tlsCert    string
		tlsKey     string
		printTools bool
	)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start a Model Context Protocol (MCP) server for Canary requirement tracking",
		Long: `Start a Model Context Protocol (MCP) server that provides AI assistants
with access to Canary requirement tracking through a standardized interface.

The server binds 127.0.0.1 by default. Binding any other interface requires
both a TLS certificate/key pair (--tls-cert/--tls-key) and a bearer token in
CANARY_MCP_TOKEN; without them the server refuses to start rather than expose
every tool on this machine to the network.

Authentication:
  CANARY_MCP_TOKEN       bearer token granting read and write
  CANARY_MCP_READ_TOKEN  bearer token granting read only

With neither set, a loopback server serves every request unauthenticated
(the developer default). With either set, a bearer token is required even on
loopback.

Run 'canary mcp --print-tools' to print the tool documentation and exit.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if printTools {
				fmt.Fprint(cmd.OutOrStdout(), RenderDocs())
				return nil
			}
			return runServer(cmd, serverConfig{
				version: version,
				host:    host,
				port:    port,
				root:    root,
				tlsCert: tlsCert,
				tlsKey:  tlsKey,
			})
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "port to listen on")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "host to bind to (non-loopback requires --tls-cert, --tls-key and CANARY_MCP_TOKEN)")
	cmd.Flags().StringVar(&root, "root", ".", "root directory the server answers for; tool paths are confined to it")
	cmd.Flags().StringVar(&tlsCert, "tls-cert", "", "TLS certificate file (required to bind a non-loopback host)")
	cmd.Flags().StringVar(&tlsKey, "tls-key", "", "TLS key file (required to bind a non-loopback host)")
	cmd.Flags().BoolVar(&printTools, "print-tools", false, "print the generated MCP tool documentation and exit")

	return cmd
}

// serverConfig is one server run's resolved inputs.
type serverConfig struct {
	version string
	host    string
	port    int
	root    string
	tlsCert string
	tlsKey  string
}

// runServer binds, serves, and shuts down.
//
// The bind is synchronous and its error is returned, so `canary mcp` on a
// taken port exits nonzero instead of printing a success banner from a
// goroutine and then blocking forever on a context nothing cancelled.
func runServer(cmd *cobra.Command, cfg serverConfig) error {
	logger := slog.Default()

	absRoot, err := filepath.Abs(cfg.root)
	if err != nil {
		return fmt.Errorf("resolve --root %s: %w", cfg.root, err)
	}
	if info, err := os.Stat(absRoot); err != nil || !info.IsDir() {
		return fmt.Errorf("--root %s is not a directory", cfg.root)
	}

	if err := checkTLSFlags(cfg); err != nil {
		return err
	}

	loopback := isLoopbackHost(cfg.host)
	auth := authFromEnv(loopback)
	if err := checkExposure(cfg, loopback, auth); err != nil {
		return err
	}

	deps := Deps{Root: absRoot}.resolve()
	server := NewServer(cfg.version, deps)

	handler := buildHandler(server, auth)

	addr := net.JoinHostPort(cfg.host, fmt.Sprintf("%d", cfg.port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	httpServer := newHTTPServer(addr, handler)

	// Only now, with a listener actually in hand, is it true that the server
	// is reachable -- so only now is it printed. The banner asks the same
	// question the serve call below does, so it cannot announce https over a
	// plaintext listener.
	printServerInfo(cmd, ln.Addr().String(), cfg.tlsEnabled(), auth.configured())
	logger.Info("Starting Canary MCP Server", "addr", ln.Addr().String(), "root", absRoot)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		if cfg.tlsEnabled() {
			serveErr <- httpServer.ServeTLS(ln, cfg.tlsCert, cfg.tlsKey)
			return
		}
		serveErr <- httpServer.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve on %s: %w", addr, err)
		}
		return nil
	case <-ctx.Done():
	}

	fmt.Fprintln(cmd.OutOrStdout(), "\nShutting down MCP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	logger.Info("Server stopped gracefully")
	return nil
}

// tlsEnabled reports whether this run will actually serve TLS. Both halves of
// the pair are required, because ServeTLS needs both: asking for only one is
// answered by checkTLSFlags before anything binds.
func (c serverConfig) tlsEnabled() bool {
	return c.tlsCert != "" && c.tlsKey != ""
}

// checkTLSFlags refuses a half-configured TLS pair, on every host including
// loopback.
//
// --tls-cert without --tls-key used to be accepted: the serve call fell back
// to plaintext because it required both, while the banner printed "https"
// because it looked at the certificate alone. An operator who typed one flag
// got a cleartext server that told them it was encrypted. Both or neither.
func checkTLSFlags(cfg serverConfig) error {
	switch {
	case cfg.tlsCert != "" && cfg.tlsKey == "":
		return fmt.Errorf("--tls-cert was given without --tls-key: TLS needs both, and one alone would serve plaintext")
	case cfg.tlsKey != "" && cfg.tlsCert == "":
		return fmt.Errorf("--tls-key was given without --tls-cert: TLS needs both, and one alone would serve plaintext")
	default:
		return nil
	}
}

// checkExposure refuses a configuration that would put the tool surface on
// the network without transport security and a credential.
//
// This runs BEFORE the listener exists. A server that binds first and
// complains second has already accepted a connection from whoever was
// waiting.
func checkExposure(cfg serverConfig, loopback bool, auth authConfig) error {
	if loopback {
		return nil
	}
	var missing []string
	if cfg.tlsCert == "" {
		missing = append(missing, "--tls-cert")
	}
	if cfg.tlsKey == "" {
		missing = append(missing, "--tls-key")
	}
	if auth.token == "" {
		missing = append(missing, EnvToken)
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf(
		"refusing to bind non-loopback host %q without %v: an MCP server reachable off this machine must use TLS (--tls-cert, --tls-key) and require a bearer token (%s)",
		cfg.host, missing, EnvToken)
}

// buildHandler assembles the served routes with their limits.
//
// The body cap wraps the auth check, which wraps the MCP handler: a caller
// that is not allowed in should not first be allowed to send a megabyte, and
// a caller that is allowed in should not be able to send more than one.
func buildHandler(server *mcp.Server, auth authConfig) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)

	mux := http.NewServeMux()
	mux.Handle("/mcp", authMiddleware(auth, mcpHandler))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	return http.MaxBytesHandler(mux, maxBodyBytes)
}

// newHTTPServer applies the transport limits.
func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}

// printServerInfo announces the running server and its tools.
//
// The tool list is RenderDocs() -- the same text docs/MCP_TOOLS.md is
// generated from -- rather than the forty hand-maintained lines it replaced,
// which had drifted far enough to describe five tools that did nothing.
func printServerInfo(cmd *cobra.Command, addr string, tls, authConfigured bool) {
	out := cmd.OutOrStdout()
	scheme := "http"
	if tls {
		scheme = "https"
	}
	authMode := "none (loopback dev mode)"
	if authConfigured {
		authMode = "bearer token required"
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Canary MCP Server")
	fmt.Fprintln(out, "=================")
	fmt.Fprintf(out, "Server listening on %s://%s\n", scheme, addr)
	fmt.Fprintf(out, "Authentication: %s\n", authMode)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Endpoints:")
	fmt.Fprintln(out, "  GET  /health         - Health check")
	fmt.Fprintln(out, "  POST /mcp            - MCP endpoint")
	fmt.Fprintln(out)
	fmt.Fprint(out, RenderDocs())
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Press Ctrl+C to stop")
}
