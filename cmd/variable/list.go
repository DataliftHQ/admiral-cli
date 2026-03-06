package variable

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/resolve"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
		pageSize   int32
		pageToken  string
	)

	cmd := &cobra.Command{
		Use:   "list [app]",
		Short: "List variables",
		Long: `List variables for a given scope.

Scope is inferred from arguments:
  admiral variable list                     → GLOBAL
  admiral variable list my-api              → APP (includes global)
  admiral variable list my-api -e staging   → APP_ENV (includes global + app)`,
		Example: `  # List global variables
  admiral variable list

  # List app-scoped variables (includes global)
  admiral variable list my-api

  # List variables for a specific environment (includes global + app)
  admiral variable list my-api -e staging

  # Paginated listing
  admiral variable list my-api --page-size 10`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var appArg string
			if len(args) == 1 {
				appArg = args[0]
			}

			rs, err := resolveScopeWithHelp(cmd, globalFlag, envFlag, appArg, "")
			if err != nil {
				return err
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

			resp, err := c.Variable().ListVariables(cmd.Context(), &variablev1.ListVariablesRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
				Filter:    resolve.VariableFilter(appID, envID),
			})
			if err != nil {
				return err
			}

			if len(resp.Variables) == 0 {
				output.Writef(cmd.OutOrStdout(), "No variables found\n")
				return nil
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					output.Writeln(w, "ID\tKEY\tVALUE\tTYPE\tSENSITIVE\tSCOPE\tDESCRIPTION\tCREATED BY\tAGE")
					for _, v := range resp.Variables {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							v.Id,
							v.Key,
							formatValue(v),
							output.FormatEnum(v.Type.String(), "VARIABLE_TYPE_"),
							formatSensitive(v.Sensitive),
							formatScope(v),
							v.Description,
							v.CreatedBy,
							output.FormatAge(v.CreatedAt),
						)
					}
				} else {
					output.Writeln(w, "KEY\tVALUE\tTYPE\tSENSITIVE\tSCOPE\tAGE")
					for _, v := range resp.Variables {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
							v.Key,
							formatValue(v),
							output.FormatEnum(v.Type.String(), "VARIABLE_TYPE_"),
							formatSensitive(v.Sensitive),
							formatScope(v),
							output.FormatAge(v.CreatedAt),
						)
					}
				}
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" && opts.OutputFormat != output.FormatJSON && opts.OutputFormat != output.FormatYAML {
				output.Writef(cmd.ErrOrStderr(), "\nNEXT PAGE TOKEN: %s\n", resp.NextPageToken)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "list variables at global scope")
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "maximum number of results per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "pagination token from a previous response")

	return cmd
}
