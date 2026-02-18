package app

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/properties"
	applicationv1 "go.admiral.io/sdk/proto/application/v1"
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

			appName, err := resolveAppWithHelp(cmd, appArg, props.App)
			if err != nil {
				return err
			}

			if !confirm {
				return fmt.Errorf("use --confirm to delete application %s", appName)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Application().DeleteApplication(cmd.Context(), &applicationv1.DeleteApplicationRequest{
				ApplicationId: appName,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Application %s deleted\n", appName)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")

	return cmd
}
