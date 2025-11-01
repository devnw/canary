// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// New returns the MCP subcommand for Canary
func New() *cobra.Command {
	var (
		port int
		host string
	)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start a Model Context Protocol (MCP) server for Canary requirement tracking",
		Long: `Start a Model Context Protocol (MCP) server that provides AI assistants
with access to Canary requirement tracking through a standardized interface.

The MCP server exposes tools for:
- Scanning for CANARY tokens
- Creating new requirements
- Listing and searching tokens
- Viewing requirement status
- Managing specifications
- Generating implementation plans

The server runs as an HTTP endpoint that AI assistants can interact with.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := slog.Default()

			// Create the MCP server
			server := mcp.NewServer(&mcp.Implementation{
				Name:    "canary-server",
				Version: "1.0.0",
			}, nil)

			// Add all Canary tools to the server
			mcp.AddTool(server, &mcp.Tool{
				Name:        "list",
				Description: "List CANARY tokens with optional filtering",
			}, handleList)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "show",
				Description: "Display all CANARY tokens for a specific requirement ID",
			}, handleShow)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "create",
				Description: "Generate a new CANARY token template",
			}, handleCreate)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "status",
				Description: "Show implementation progress for a requirement",
			}, handleStatus)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "search",
				Description: "Search CANARY tokens by keywords",
			}, handleSearch)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "next",
				Description: "Identify next highest priority unimplemented requirement",
			}, handleNext)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "scan",
				Description: "Scan codebase for CANARY tokens",
			}, handleScan)

			// Create HTTP handler for MCP
			handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
				return server
			}, nil)

			// Set up routes
			mux := http.NewServeMux()
			mux.Handle("/mcp", handler)

			// Health check endpoint
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status": "healthy",
					"time":   time.Now().Format(time.RFC3339),
				})
			})

			// Create HTTP server
			addr := fmt.Sprintf("%s:%d", host, port)
			httpServer := &http.Server{
				Addr:    addr,
				Handler: mux,
			}

			// Start server in a goroutine
			go func() {
				logger.Info("Starting Canary MCP Server", "addr", addr)
				printServerInfo(addr)

				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server error", "error", err)
				}
			}()

			// Wait for context cancellation (SIGINT/SIGTERM)
			<-ctx.Done()
			fmt.Println("\nShutting down MCP server...")

			// Graceful shutdown with timeout
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("Server shutdown error", "error", err)
				return err
			}

			logger.Info("Server stopped gracefully")
			return nil
		},
	}

	// Add flags
	cmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	cmd.Flags().StringVar(&host, "host", "localhost", "Host to bind to")

	return cmd
}

func printServerInfo(addr string) {
	fmt.Println()
	fmt.Println("Canary MCP Server")
	fmt.Println("=================")
	fmt.Printf("Server listening on http://%s\n", addr)
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health         - Health check")
	fmt.Println("  POST /mcp            - MCP endpoint")
	fmt.Println()
	fmt.Println("Available MCP Tools:")
	fmt.Println("  - list               - List CANARY tokens with filtering")
	fmt.Println("  - show               - Show details for a specific requirement")
	fmt.Println("  - create             - Create a new CANARY token")
	fmt.Println("  - status             - Show implementation status")
	fmt.Println("  - search             - Search tokens by keywords")
	fmt.Println("  - next               - Get next priority requirement")
	fmt.Println("  - scan               - Scan codebase for CANARY tokens")
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Printf("  curl -H 'Content-Type: application/json' \\\n")
	fmt.Printf("       -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"list\",\"arguments\":{\"status\":\"IMPL\"}}}' \\\n")
	fmt.Printf("       http://%s/mcp\n", addr)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
}
