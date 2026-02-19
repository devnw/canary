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

			// Core token management
			mcp.AddTool(server, &mcp.Tool{
				Name:        "list",
				Description: "List CANARY tokens with optional filtering (default limit 25, max 50 to reduce context)",
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

			// Workflow tools
			mcp.AddTool(server, &mcp.Tool{
				Name:        "scan",
				Description: "Scan codebase for CANARY tokens",
			}, handleScan)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "specify",
				Description: "Create a requirement specification",
			}, handleSpecify)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "plan",
				Description: "Generate implementation plan for a requirement",
			}, handlePlan)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "implement",
				Description: "Get implementation guidance for a requirement",
			}, handleImplement)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "index",
				Description: "Index codebase tokens into database",
			}, handleIndex)

			// Query and navigation
			mcp.AddTool(server, &mcp.Tool{
				Name:        "files",
				Description: "Find files containing tokens for a requirement",
			}, handleFiles)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "grep",
				Description: "Search tokens by pattern in specific fields",
			}, handleGrep)

			// Management
			mcp.AddTool(server, &mcp.Tool{
				Name:        "prioritize",
				Description: "Set priority level for a requirement",
			}, handlePrioritize)

			// Bug tracking
			mcp.AddTool(server, &mcp.Tool{
				Name:        "bug-list",
				Description: "List bug tracking tokens",
			}, handleBugList)

			mcp.AddTool(server, &mcp.Tool{
				Name:        "bug-create",
				Description: "Create a new bug tracking token",
			}, handleBugCreate)

			// Gap analysis
			mcp.AddTool(server, &mcp.Tool{
				Name:        "gap-mark",
				Description: "Mark gap analysis claim as helpful or unhelpful",
			}, handleGapMark)

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
	fmt.Println("Available MCP Tools (18 total):")
	fmt.Println()
	fmt.Println("Core Token Management:")
	fmt.Println("  ? list               - List CANARY tokens with filtering")
	fmt.Println("  ? show               - Show details for a specific requirement")
	fmt.Println("  ? create             - Create a new CANARY token")
	fmt.Println("  ? status             - Show implementation status")
	fmt.Println("  ? search             - Search tokens by keywords")
	fmt.Println("  ? next               - Get next priority requirement")
	fmt.Println()
	fmt.Println("Workflow & Development:")
	fmt.Println("  ? scan               - Scan codebase for CANARY tokens")
	fmt.Println("  ? specify            - Create requirement specification")
	fmt.Println("  ? plan               - Generate implementation plan")
	fmt.Println("  ? implement          - Get implementation guidance")
	fmt.Println("  ? index              - Index codebase tokens into database")
	fmt.Println()
	fmt.Println("Query & Navigation:")
	fmt.Println("  ? files              - Find files containing requirement tokens")
	fmt.Println("  ? grep               - Search tokens by pattern")
	fmt.Println()
	fmt.Println("Management:")
	fmt.Println("  ? prioritize         - Set requirement priority")
	fmt.Println()
	fmt.Println("Bug Tracking:")
	fmt.Println("  ? bug-list           - List bug tracking tokens")
	fmt.Println("  ? bug-create         - Create new bug token")
	fmt.Println()
	fmt.Println("Gap Analysis:")
	fmt.Println("  ? gap-mark           - Mark gap claims as helpful/unhelpful")
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Printf("  curl -H 'Content-Type: application/json' \\\n")
	fmt.Printf("       -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"list\",\"arguments\":{\"status\":\"IMPL\"}}}' \\\n")
	fmt.Printf("       http://%s/mcp\n", addr)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
}
