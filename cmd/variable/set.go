package variable

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/properties"
)

func newSetCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
		sensitive  bool
		confirm    bool
	)

	cmd := &cobra.Command{
		Use:   "set [app] KEY=VALUE [KEY=VALUE ...]",
		Short: "Set one or more variables",
		Long: `Set one or more configuration variables.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'. Use --global for variables
that apply across all apps (requires confirmation).

Scope is inferred automatically:
  --global               → GLOBAL (all apps, requires --confirm or prompt)
  <app> -e <env> K=V     → APP_ENV (app + environment)
  <app> K=V              → APP (app-scoped)
  K=V (with context set) → APP (from active context)`,
		Example: `  # Set an app-scoped variable (app as argument)
  admiral variable set my-api IMAGE_TAG=v2.0.0

  # Set multiple variables for an app+environment
  admiral variable set my-api -e staging IMAGE_TAG=v2.0.0 LOG_LEVEL=debug

  # Set a global variable (with confirmation)
  admiral variable set --global LOG_FORMAT=json --confirm

  # Use the active app context instead of a positional argument
  admiral use my-api
  admiral variable set IMAGE_TAG=v2.0.0`,
		Args: cmdutil.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appArg, kvPairs, err := splitArgs(args)
			if err != nil {
				return err
			}

			if len(kvPairs) == 0 {
				return cmd.Help()
			}

			// Parse all KEY=VALUE pairs.
			vars := make(map[string]string, len(kvPairs))
			for _, kv := range kvPairs {
				k, v, err := parseKV(kv)
				if err != nil {
					return err
				}
				vars[k] = v
			}

			// Load context app.
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
					"Setting GLOBAL variables affects all apps. Continue?",
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
				Operation:   "set",
				Scope:       string(rs.Scope),
				App:         rs.App,
				Environment: rs.Env,
				Variables:   vars,
				Sensitive:   sensitive,
				Status:      stubStatus,
			}

			return printStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "set variables at global scope")
	cmd.Flags().BoolVar(&sensitive, "sensitive", false, "mark variables as sensitive")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "skip confirmation prompt")

	return cmd
}
