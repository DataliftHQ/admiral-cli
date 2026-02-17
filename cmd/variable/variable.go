package variable

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// VariableCmd is the parent command for variable operations.
type VariableCmd struct {
	Cmd *cobra.Command
}

// NewVariableCmd creates the variable command tree.
func NewVariableCmd(opts *factory.Options) *VariableCmd {
	root := &VariableCmd{}

	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage configuration variables",
		Long: `Manage configuration variables scoped to an app, app+environment, or globally.

Most commands require an app context. You can provide it in two ways:
  1. As a positional argument:  admiral variable list my-api
  2. From the active context:   admiral use my-api && admiral variable list

Use --global for variables that apply across all apps.
Use -e/--env to target a specific environment within an app.`,
		Aliases:       []string{"var"},
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newSetCmd(opts),
		newGetCmd(opts),
		newListCmd(opts),
		newDeleteCmd(opts),
	)

	root.Cmd = cmd
	return root
}
