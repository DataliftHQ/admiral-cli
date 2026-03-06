package variable

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/resolve"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
		varID      string
		confirm    bool
	)

	cmd := &cobra.Command{
		Use:   "delete [app] KEY",
		Short: "Delete a variable",
		Long: `Delete a variable by key.

The variable is resolved by key and scope. Use --id to look up by UUID directly.
Requires --confirm to prevent accidental deletion.`,
		Example: `  # Delete an app-scoped variable
  admiral variable delete my-api IMAGE_TAG --confirm

  # Delete a variable for a specific environment
  admiral variable delete my-api IMAGE_TAG -e staging --confirm

  # Delete a global variable
  admiral variable delete LOG_FORMAT --confirm

  # Delete by UUID
  admiral variable delete --id 550e8400-e29b-41d4-a716-446655440000 --confirm

  # Use the active app context
  admiral use my-api
  admiral variable delete IMAGE_TAG --confirm`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id, display string

			if varID != "" {
				id = varID
				display = varID
			} else {
				if len(args) == 0 {
					return fmt.Errorf("variable key or --id is required")
				}

				appArg, key, err := splitArgsWithKey(args)
				if err != nil {
					return err
				}
				display = key

				rs, err := resolveScopeWithHelp(cmd, globalFlag, envFlag, appArg, "")
				if err != nil {
					return err
				}

				if !confirm {
					return fmt.Errorf("use --confirm to delete variable %s", display)
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

				id, err = resolve.VariableByKey(cmd.Context(), c.Variable(), key, appID, envID)
				if err != nil {
					return err
				}

				resp, err := c.Variable().DeleteVariable(cmd.Context(), &variablev1.DeleteVariableRequest{
					VariableId: id,
				})
				if err != nil {
					return err
				}

				p := output.NewPrinter(opts.OutputFormat)
				return p.PrintResource(resp, func(w *tabwriter.Writer) {
					output.Writef(w, "Variable %s deleted\n", display)
				})
			}

			if !confirm {
				return fmt.Errorf("use --confirm to delete variable %s", display)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Variable().DeleteVariable(cmd.Context(), &variablev1.DeleteVariableRequest{
				VariableId: id,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Variable %s deleted\n", display)
			})
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "delete variable at global scope")
	cmd.Flags().StringVar(&varID, "id", "", "variable ID (UUID)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")

	return cmd
}
