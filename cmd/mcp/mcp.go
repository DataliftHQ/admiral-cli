package mcp

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// MCPCmd is the parent command for MCP server operations.
type MCPCmd struct {
	Cmd *cobra.Command
}

// NewMCPCmd creates the mcp command tree.
func NewMCPCmd(opts *factory.Options) *MCPCmd {
	root := &MCPCmd{}

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol server for LLM integration",
		Long: `Start an MCP server that exposes Admiral operations as tools for LLMs.

The MCP server allows AI assistants to interact with Admiral programmatically —
querying deployments, managing variables, and triggering operations through
structured tool calls.

The server communicates over stdio using the Model Context Protocol (JSON-RPC).`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newServeCmd(opts),
	)

	root.Cmd = cmd
	return root
}
