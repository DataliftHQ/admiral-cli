package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// EnvCmd is the parent command for environment operations.
type EnvCmd struct {
	Cmd *cobra.Command
}

// NewEnvCmd creates the env command tree.
func NewEnvCmd(opts *factory.Options) *EnvCmd {
	root := &EnvCmd{}

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environments",
		Long: `Manage environments within an application.

Environments are app-scoped deployment targets (e.g. dev, staging, prod).
Each environment is bound to a runtime and has its own configuration,
labels, and deployment lifecycle.

The parent application is specified with --app.`,
		Aliases:       []string{"environment"},
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newListCmd(opts),
		newCreateCmd(opts),
		newGetCmd(opts),
		newUpdateCmd(opts),
		newDeleteCmd(opts),
	)

	root.Cmd = cmd
	return root
}

// addAppFlag registers the --app flag on a subcommand so it appears
// under the command's own Flags section instead of Global Flags.
func addAppFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVar(dest, "app", "", "parent application name")
}

// resolveAppForEnv returns the --app flag value or an error if empty.
func resolveAppForEnv(appFlag string) (string, error) {
	if appFlag != "" {
		return appFlag, nil
	}
	return "", fmt.Errorf("no app specified; use --app to provide the parent application")
}
