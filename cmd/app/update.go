package app

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/properties"
	applicationv1 "go.admiral.io/sdk/proto/application/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		labelStrs   []string
		description string
	)

	cmd := &cobra.Command{
		Use:   "update [app]",
		Short: "Update an application",
		Long: `Update an existing application.

The app can be provided as a positional argument or resolved from the
active context set via 'admiral use <app>'.`,
		Example: `  # Update labels
  admiral app update billing-api --label team=payments

  # Update description
  admiral app update billing-api --description "New description"

  # Update labels using active context
  admiral use billing-api
  admiral app update --label team=payments`,
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

			appName, err := resolveAppWithHelp(cmd, appArg, props.App)
			if err != nil {
				return err
			}

			var paths []string
			application := &applicationv1.Application{Id: appName}

			if cmd.Flags().Changed("label") {
				labels, err := cmdutil.ParseLabels(labelStrs)
				if err != nil {
					return err
				}
				application.Labels = labels
				paths = append(paths, "labels")
			}

			if cmd.Flags().Changed("description") {
				application.Description = description
				paths = append(paths, "description")
			}

			if len(paths) == 0 {
				return fmt.Errorf("at least --label or --description must be specified")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Application().UpdateApplication(cmd.Context(), &applicationv1.UpdateApplicationRequest{
				Application: application,
				UpdateMask:  &fieldmaskpb.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				app := resp.Application
				output.Writeln(w, "NAME\tDESCRIPTION\tAGE")
				output.Writef(w, "%s\t%s\t%s\n",
					app.Name,
					app.Description,
					output.FormatAge(app.CreatedAt),
				)
			})
		},
	}

	cmdutil.AddLabelFlag(cmd, &labelStrs, "label to set (key=value, repeatable)")
	cmd.Flags().StringVar(&description, "description", "", "application description")

	return cmd
}
