package variable

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/resolve"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag     string
		globalFlag  bool
		sensitive   bool
		varType     string
		description string
		confirm     bool
	)

	cmd := &cobra.Command{
		Use:   "create [app] KEY=VALUE",
		Short: "Create a variable",
		Long: `Create a single configuration variable.

Scope is inferred automatically:
  KEY=VALUE (no app)         → GLOBAL (requires --confirm)
  --global KEY=VALUE         → GLOBAL (explicit, requires --confirm)
  <app> KEY=VALUE            → APP
  <app> -e <env> KEY=VALUE   → APP_ENV`,
		Example: `  # Create an app-scoped variable
  admiral variable create my-api IMAGE_TAG=v2.0.0

  # Create a variable for an app+environment
  admiral variable create my-api -e staging IMAGE_TAG=v2.0.0

  # Create a global variable (with confirmation)
  admiral variable create LOG_FORMAT=json --confirm

  # Create a sensitive variable with type
  admiral variable create my-api DB_PASSWORD=secret --sensitive --type string`,
		Args: cmdutil.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appArg, kvPairs, err := splitArgs(args)
			if err != nil {
				return err
			}

			if len(kvPairs) == 0 {
				return cmd.Help()
			}
			if len(kvPairs) > 1 {
				return fmt.Errorf("create accepts exactly one KEY=VALUE pair; got %d", len(kvPairs))
			}

			key, value, err := parseKV(kvPairs[0])
			if err != nil {
				return err
			}

			rs, err := resolveScopeWithHelp(cmd, globalFlag, envFlag, appArg, "")
			if err != nil {
				return err
			}

			if rs.Scope == scopeGlobal && !confirm {
				return fmt.Errorf("creating a GLOBAL variable affects all apps; use --confirm to proceed")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			appID, envID, err := resolve.ScopeIDs(cmd.Context(), c.Application(), c.Environment(), rs.App, rs.Env)
			if err != nil {
				return err
			}

			req := &variablev1.CreateVariableRequest{
				Key:       key,
				Value:     value,
				Sensitive: sensitive,
			}

			if appID != "" {
				req.ApplicationId = &appID
			}
			if envID != "" {
				req.EnvironmentId = &envID
			}

			if cmd.Flags().Changed("type") {
				vt, err := resolve.VariableType(varType)
				if err != nil {
					return err
				}
				req.Type = vt
			}

			if cmd.Flags().Changed("description") {
				req.Description = description
			}

			checkShadowWarning(cmd.Context(), cmd.ErrOrStderr(), c.Variable(), key, rs, appID, envID)

			resp, err := c.Variable().CreateVariable(cmd.Context(), req)
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				v := resp.Variable
				output.Writeln(w, "KEY\tVALUE\tTYPE\tSENSITIVE\tSCOPE\tAGE")
				output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					v.Key,
					formatValue(v),
					output.FormatEnum(v.Type.String(), "VARIABLE_TYPE_"),
					formatSensitive(v.Sensitive),
					formatScope(v),
					output.FormatAge(v.CreatedAt),
				)
			})
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "create variable at global scope")
	cmd.Flags().BoolVar(&sensitive, "sensitive", false, "mark variable as sensitive")
	cmd.Flags().StringVar(&varType, "type", "", "variable type (string, number, boolean, complex)")
	cmd.Flags().StringVar(&description, "description", "", "variable description")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm global scope operation")

	return cmd
}
