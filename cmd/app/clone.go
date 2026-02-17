package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newCloneCmd(opts *factory.Options) *cobra.Command {
	var (
		from             string
		to               string
		includeVariables bool
		excludeVariable  []string
	)

	cmd := &cobra.Command{
		Use:   "clone [app]",
		Short: "Clone application configuration between environments",
		Long: `Clone an application's configuration from one environment to another.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.

Both --from and --to are required.`,
		Example: `  # Clone staging to production
  admiral app clone billing-api --from staging --to production

  # Clone including variables
  admiral app clone billing-api --from staging --to production --include-variables

  # Clone but exclude specific variables
  admiral app clone billing-api --from staging --to production --include-variables --exclude-variable SECRET_KEY

  # Use the active app context
  admiral use billing-api
  admiral app clone --from staging --to production`,
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

			flags := map[string]any{
				"from": from,
				"to":   to,
			}
			if includeVariables {
				flags["include-variables"] = true
			}
			if len(excludeVariable) > 0 {
				flags["exclude-variable"] = excludeVariable
			}

			stub := cmdutil.StubResult{
				Command: "app clone",
				App:     app,
				Flags:   flags,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "source environment (required)")
	cmd.Flags().StringVar(&to, "to", "", "target environment (required)")
	cmd.Flags().BoolVar(&includeVariables, "include-variables", false, "include variables in the clone")
	cmd.Flags().StringArrayVar(&excludeVariable, "exclude-variable", nil, "variable keys to exclude (repeatable)")

	return cmd
}
