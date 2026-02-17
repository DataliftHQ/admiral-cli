package app

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [app]",
		Short: "Get application details",
		Long: `Get detailed information about an application.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.`,
		Example: `  # Get app details by slug
  admiral app get billing-api

  # Use the active app context
  admiral use billing-api
  admiral app get`,
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

			app, err := resolveAppWithHelp(cmd, appArg, props.App)
			if err != nil {
				return err
			}

			stub := cmdutil.StubResult{
				Command: "app get",
				App:     app,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	return cmd
}
