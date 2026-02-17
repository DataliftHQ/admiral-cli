package variable

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
	)

	cmd := &cobra.Command{
		Use:   "list [app]",
		Short: "List variables",
		Long: `List variables for a given scope.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'. Use --global for global variables.`,
		Example: `  # List app-scoped variables
  admiral variable list my-api

  # List variables for a specific environment
  admiral variable list my-api -e staging

  # List global variables
  admiral variable list --global

  # Use the active app context
  admiral use my-api
  admiral variable list`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var appArg string
			if len(args) == 1 {
				appArg = args[0]
			}

			props, err := properties.Load(opts.ConfigDir)
			if err != nil {
				return err
			}

			rs, err := resolveScopeWithHelp(cmd, globalFlag, envFlag, appArg, props.App)
			if err != nil {
				return err
			}

			stub := stubResult{
				Operation:   "list",
				Scope:       string(rs.Scope),
				App:         rs.App,
				Environment: rs.Env,
				Status:      stubStatus,
			}

			return printStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "list variables at global scope")

	return cmd
}
