package variable

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
		confirm    bool
	)

	cmd := &cobra.Command{
		Use:   "delete [app] KEY",
		Short: "Delete a variable",
		Long: `Delete a variable by key.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'. Use --global for global variables
(requires confirmation).`,
		Example: `  # Delete an app-scoped variable
  admiral variable delete my-api IMAGE_TAG

  # Delete a variable for a specific environment
  admiral variable delete my-api IMAGE_TAG -e staging

  # Delete a global variable (with confirmation)
  admiral variable delete --global LOG_FORMAT --confirm

  # Use the active app context
  admiral use my-api
  admiral variable delete IMAGE_TAG`,
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

			// GLOBAL scope requires confirmation.
			if rs.Scope == scopeGlobal && !confirm {
				ok, err := cmdutil.ConfirmPrompt(
					cmd.InOrStdin(), cmd.ErrOrStderr(),
					"Deleting a GLOBAL variable affects all apps. Continue?",
				)
				if err != nil {
					return err
				}
				if !ok {
					cmdutil.Writef(cmd.ErrOrStderr(), "Aborted.\n")
					return nil
				}
			}

			stub := stubResult{
				Operation:   "delete",
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
	cmd.Flags().BoolVar(&globalFlag, "global", false, "delete variable at global scope")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")

	return cmd
}
