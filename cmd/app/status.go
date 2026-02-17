package app

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newStatusCmd(opts *factory.Options) *cobra.Command {
	var envFlag string

	cmd := &cobra.Command{
		Use:   "status [app]",
		Short: "Show application deployment status",
		Long: `Show the deployment status of an application across environments.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.

Use -e/--env to filter to a specific environment.`,
		Example: `  # Show status across all environments
  admiral app status billing-api

  # Show status for a specific environment
  admiral app status billing-api -e staging

  # Use the active app context
  admiral use billing-api
  admiral app status -e production`,
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

			flags := map[string]any{}
			if envFlag != "" {
				flags["env"] = envFlag
			}

			stub := cmdutil.StubResult{
				Command:     "app status",
				App:         app,
				Environment: envFlag,
				Flags:       flags,
				Status:      cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "filter to a specific environment")

	return cmd
}
