package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [app]",
		Short: "Delete an application",
		Long: `Delete an application.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.

Requires --confirm to prevent accidental deletion.`,
		Example: `  # Delete an application
  admiral app delete billing-api --confirm

  # Delete using active context
  admiral use billing-api
  admiral app delete --confirm`,
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

			if !confirm {
				return fmt.Errorf("--confirm is required to delete an application")
			}

			stub := cmdutil.StubResult{
				Command: "app delete",
				App:     app,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")

	return cmd
}
