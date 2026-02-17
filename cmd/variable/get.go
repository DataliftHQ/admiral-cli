package variable

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
	)

	cmd := &cobra.Command{
		Use:   "get [app] KEY",
		Short: "Get a variable by key",
		Long: `Get a single variable by key.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'. Use --global for global variables.`,
		Example: `  # Get an app-scoped variable
  admiral variable get my-api IMAGE_TAG

  # Get a variable for a specific environment
  admiral variable get my-api IMAGE_TAG -e staging

  # Get a global variable
  admiral variable get --global LOG_FORMAT

  # Use the active app context
  admiral use my-api
  admiral variable get IMAGE_TAG`,
		Args: cmdutil.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			appArg, key, err := splitArgsWithKey(args)
			if err != nil {
				return err
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
				Operation:   "get",
				Scope:       string(rs.Scope),
				App:         rs.App,
				Environment: rs.Env,
				Key:         key,
				Status:      stubStatus,
			}

			return printStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "get variable at global scope")

	return cmd
}
