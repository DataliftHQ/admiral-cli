package agent

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// AgentCmd is the parent command for agent operations.
type AgentCmd struct {
	Cmd *cobra.Command
}

// NewAgentCmd creates the agent command tree.
func NewAgentCmd(opts *factory.Options) *AgentCmd {
	root := &AgentCmd{}

	cmd := &cobra.Command{
		Use:           "agent",
		Short:         "Manage agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newListCmd(opts),
		newGetCmd(opts),
	)

	root.Cmd = cmd
	return root
}
