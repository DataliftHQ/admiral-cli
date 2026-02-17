package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
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
Each environment is bound to a cluster and has its own lifecycle, promotion
order, and configuration.

The parent application is resolved from the active context set via
'admiral use <app>'. Set the context before running env commands:

  admiral use billing-api
  admiral env list`,
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

// resolveAppForEnv loads the active app context. Returns an error if no app
// is set, directing the user to run 'admiral use <app>'.
func resolveAppForEnv(configDir string) (string, error) {
	props, err := properties.Load(configDir)
	if err != nil {
		return "", err
	}
	if props.App == "" {
		return "", fmt.Errorf("no app context set; run 'admiral use <app>' first")
	}
	return props.App, nil
}
