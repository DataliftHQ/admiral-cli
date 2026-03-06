package variable

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/resolve"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag       string
		globalFlag    bool
		varID         string
		value         string
		sensitiveFlag bool
		noSensitive   bool
		varType       string
		description   string
	)

	cmd := &cobra.Command{
		Use:   "update [app] KEY",
		Short: "Update a variable",
		Long: `Update an existing variable.

The variable is resolved by key and scope. Use --id to look up by UUID directly.
At least one field (--value, --sensitive, --no-sensitive, --type, --description) must be specified.`,
		Example: `  # Update the value of an app-scoped variable
  admiral variable update my-api IMAGE_TAG --value v3.0.0

  # Mark a variable as sensitive
  admiral variable update my-api DB_PASSWORD --sensitive

  # Update multiple fields
  admiral variable update my-api CONFIG --value '{"key":"val"}' --type complex --description "App config"

  # Update by UUID
  admiral variable update --id 550e8400-e29b-41d4-a716-446655440000 --value new-value

  # Unmark a variable as sensitive
  admiral variable update my-api LOG_LEVEL --no-sensitive`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sensitiveFlag && noSensitive {
				return fmt.Errorf("--sensitive and --no-sensitive are mutually exclusive")
			}

			var paths []string
			variable := &variablev1.Variable{}

			if cmd.Flags().Changed("value") {
				variable.Value = value
				paths = append(paths, "value")
			}
			if sensitiveFlag {
				variable.Sensitive = true
				paths = append(paths, "sensitive")
			}
			if noSensitive {
				variable.Sensitive = false
				paths = append(paths, "sensitive")
			}
			if cmd.Flags().Changed("type") {
				vt, err := resolve.VariableType(varType)
				if err != nil {
					return err
				}
				variable.Type = vt
				paths = append(paths, "type")
			}
			if cmd.Flags().Changed("description") {
				variable.Description = description
				paths = append(paths, "description")
			}

			if len(paths) == 0 {
				return fmt.Errorf("at least one field must be specified for update (--value, --sensitive, --no-sensitive, --type, --description)")
			}

			var id string

			if varID != "" {
				id = varID
			} else {
				if len(args) == 0 {
					return fmt.Errorf("variable key or --id is required")
				}

				appArg, key, err := splitArgsWithKey(args)
				if err != nil {
					return err
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

				id, err = resolve.VariableByKey(cmd.Context(), c.Variable(), key, appID, envID)
				if err != nil {
					return err
				}

				variable.Id = id

				resp, err := c.Variable().UpdateVariable(cmd.Context(), &variablev1.UpdateVariableRequest{
					Variable:   variable,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
				})
				if err != nil {
					return err
				}

				return printUpdateResult(opts, resp)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			variable.Id = id

			resp, err := c.Variable().UpdateVariable(cmd.Context(), &variablev1.UpdateVariableRequest{
				Variable:   variable,
				UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return printUpdateResult(opts, resp)
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "update variable at global scope")
	cmd.Flags().StringVar(&varID, "id", "", "variable ID (UUID)")
	cmd.Flags().StringVar(&value, "value", "", "new variable value")
	cmd.Flags().BoolVar(&sensitiveFlag, "sensitive", false, "mark variable as sensitive")
	cmd.Flags().BoolVar(&noSensitive, "no-sensitive", false, "unmark variable as sensitive")
	cmd.Flags().StringVar(&varType, "type", "", "variable type (string, number, boolean, complex)")
	cmd.Flags().StringVar(&description, "description", "", "variable description")
	cmd.MarkFlagsMutuallyExclusive("sensitive", "no-sensitive")

	return cmd
}

func printUpdateResult(opts *factory.Options, resp *variablev1.UpdateVariableResponse) error {
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
}
