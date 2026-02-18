package app

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/properties"
	applicationv1 "go.admiral.io/sdk/proto/application/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [app]",
		Short: "Get application details",
		Long: `Get detailed information about an application.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.`,
		Example: `  # Get app details by name
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

			appName, err := resolveAppWithHelp(cmd, appArg, props.App)
			if err != nil {
				return err
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Application().GetApplication(cmd.Context(), &applicationv1.GetApplicationRequest{
				ApplicationId: appName,
			})
			if err != nil {
				return err
			}

			app := resp.Application
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Name", Value: app.Name},
						{Key: "Description", Value: app.Description},
						{Key: "Labels", Value: output.FormatLabels(app.Labels)},
						{Key: "Created", Value: output.FormatTimestamp(app.CreatedAt)},
						{Key: "Updated", Value: output.FormatTimestamp(app.UpdatedAt)},
						{Key: "Age", Value: output.FormatAge(app.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}

	return cmd
}
