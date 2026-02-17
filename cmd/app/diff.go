package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newDiffCmd(opts *factory.Options) *cobra.Command {
	var (
		from string
		to   string
	)

	cmd := &cobra.Command{
		Use:   "diff [app]",
		Short: "Compare application across environments",
		Long: `Compare an application's configuration between two environments.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.

Both --from and --to are required.`,
		Example: `  # Compare staging and production
  admiral app diff billing-api --from staging --to production

  # Use the active app context
  admiral use billing-api
  admiral app diff --from staging --to production`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" || to == "" {
				_ = cmd.Help()
				_, _ = fmt.Fprintln(cmd.ErrOrStderr())
				return fmt.Errorf("both --from and --to are required")
			}

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
				Command: "app diff",
				App:     app,
				Flags: map[string]any{
					"from": from,
					"to":   to,
				},
				Status: cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "source environment (required)")
	cmd.Flags().StringVar(&to, "to", "", "target environment (required)")

	return cmd
}
