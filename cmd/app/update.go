package app

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		name   string
		labels []string
	)

	cmd := &cobra.Command{
		Use:   "update [app]",
		Short: "Update an application",
		Long: `Update an existing application.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.`,
		Example: `  # Update display name
  admiral app update billing-api --name "Billing Service"

  # Update labels using active context
  admiral use billing-api
  admiral app update --label team=payments`,
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

			labelMap, err := cmdutil.ParseLabels(labels)
			if err != nil {
				return err
			}

			flags := map[string]any{}
			if name != "" {
				flags["name"] = name
			}

			stub := cmdutil.StubResult{
				Command: "app update",
				App:     app,
				Labels:  labelMap,
				Flags:   flags,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the application")
	cmdutil.AddLabelFlag(cmd, &labels, "label to set (key=value, repeatable)")

	return cmd
}
