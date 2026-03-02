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

Scope is inferred from arguments:
  admiral variable list                     → GLOBAL (no app, no env)
  admiral variable list my-api              → APP
  admiral variable list my-api -e staging   → APP_ENV

Use --global for explicitness. Use -e/--env to target a specific
environment within an app.`,
		Aliases:       []string{"var"},
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newCreateCmd(opts),
		newGetCmd(opts),
		newListCmd(opts),
		newUpdateCmd(opts),
		newDeleteCmd(opts),
	)

	root.Cmd = cmd
	return root
}
