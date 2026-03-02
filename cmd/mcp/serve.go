package mcp

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/version"
)

func newServeCmd(opts *factory.Options) *cobra.Command {
	var (
		transport string
		host      string
		port      int
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start a long-running MCP server that exposes Admiral operations as tools.

The server authenticates using the same credentials as the CLI.
Run 'admiral auth login' first, or set the ADMIRAL_TOKEN environment
variable with an API token.

Supported transports:
  stdio   — communicate over stdin/stdout (default)
  sse     — communicate over HTTP Server-Sent Events (uses --host and --port)

By default the SSE transport binds to 127.0.0.1 (localhost only).
Use --host 0.0.0.0 to listen on all interfaces.`,
		Example: `  # Start MCP server (stdio transport)
  admiral mcp serve

  # Start with SSE transport on custom port
  admiral mcp serve --transport sse --port 9090

  # Expose SSE transport on all interfaces (use with caution)
  admiral mcp serve --transport sse --host 0.0.0.0`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			v := version.GetVersion()

			s := mcp.NewServer(
				&mcp.Implementation{
					Name:    "admiral",
					Version: v.GitVersion,
				},
				&mcp.ServerOptions{
					Logger: slog.Default(),
				},
			)

			registerTools(s, c)

			switch transport {
			case "stdio":
				slog.Info("starting MCP server", "transport", "stdio")
				return s.Run(cmd.Context(), &mcp.StdioTransport{})

			case "sse":
				addr := fmt.Sprintf("%s:%d", host, port)
				slog.Info("starting MCP server", "transport", "sse", "addr", addr)
				handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
					return s
				}, nil)
				return http.ListenAndServe(addr, handler) //nolint:gosec // bind addr is user-configured

			default:
				return fmt.Errorf("unsupported transport %q; valid values: stdio, sse", transport)
			}
		},
	}

	cmd.Flags().StringVar(&transport, "transport", "stdio", "transport protocol (stdio, sse)")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "bind address for SSE transport")
	cmd.Flags().IntVar(&port, "port", 8080, "listen port for SSE transport")

	return cmd
}
