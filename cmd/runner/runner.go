package runner

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// RunnerCmd holds the runner parent command.
type RunnerCmd struct {
	Cmd *cobra.Command
}

// NewRunnerCmd creates the runner command tree.
func NewRunnerCmd(opts *factory.Options) *RunnerCmd {
	root := &RunnerCmd{}

	cmd := &cobra.Command{
		Use:   "runner",
		Short: "Manage runners",
		Args:  cobra.NoArgs,
	}

	tokenCmd := newTokenCmd(opts)

	cmd.AddCommand(
		newListCmd(opts),
		newGetCmd(opts),
		newCreateCmd(opts),
		newUpdateCmd(opts),
		newDeleteCmd(opts),
		tokenCmd,
	)

	root.Cmd = cmd
	return root
}
