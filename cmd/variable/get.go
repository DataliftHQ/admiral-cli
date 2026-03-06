package variable

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/resolve"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	var (
		envFlag    string
		globalFlag bool
		varID      string
	)

	cmd := &cobra.Command{
		Use:   "get [app] KEY",
		Short: "Get a variable by key",
		Long: `Get detailed information about a single variable.

The variable is resolved by key and scope. Use --id to look up by UUID directly.`,
		Example: `  # Get an app-scoped variable
  admiral variable get my-api IMAGE_TAG

  # Get a variable for a specific environment
  admiral variable get my-api IMAGE_TAG -e staging

  # Get a global variable
  admiral variable get LOG_FORMAT

  # Get by UUID
  admiral variable get --id 550e8400-e29b-41d4-a716-446655440000`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
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

				resp, err := c.Variable().GetVariable(cmd.Context(), &variablev1.GetVariableRequest{
					VariableId: id,
				})
				if err != nil {
					return err
				}

				return printVariableDetail(opts, resp)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Variable().GetVariable(cmd.Context(), &variablev1.GetVariableRequest{
				VariableId: id,
			})
			if err != nil {
				return err
			}

			return printVariableDetail(opts, resp)
		},
	}

	cmd.Flags().StringVarP(&envFlag, "env", "e", "", "target environment")
	cmd.Flags().BoolVar(&globalFlag, "global", false, "get variable at global scope")
	cmd.Flags().StringVar(&varID, "id", "", "variable ID (UUID)")

	return cmd
}

func printVariableDetail(opts *factory.Options, resp *variablev1.GetVariableResponse) error {
	v := resp.Variable
	p := output.NewPrinter(opts.OutputFormat)

	details := []output.Detail{
		{Key: "ID", Value: v.Id},
		{Key: "Key", Value: v.Key},
		{Key: "Value", Value: formatValue(v)},
		{Key: "Type", Value: output.FormatEnum(v.Type.String(), "VARIABLE_TYPE_")},
		{Key: "Sensitive", Value: formatSensitive(v.Sensitive)},
		{Key: "Scope", Value: formatScope(v)},
		{Key: "Application ID", Value: stringPtrOrNone(v.ApplicationId)},
		{Key: "Environment ID", Value: stringPtrOrNone(v.EnvironmentId)},
		{Key: "Description", Value: v.Description},
		{Key: "Created", Value: output.FormatTimestamp(v.CreatedAt)},
		{Key: "Created By", Value: v.CreatedBy},
		{Key: "Updated", Value: output.FormatTimestamp(v.UpdatedAt)},
		{Key: "Updated By", Value: v.UpdatedBy},
		{Key: "Age", Value: output.FormatAge(v.CreatedAt)},
	}

	sections := []output.Section{
		{Details: details},
	}

	return p.PrintDetail(resp, sections)
}
