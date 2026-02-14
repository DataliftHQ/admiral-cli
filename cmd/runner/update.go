package runner

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		name   string
		labels []string
	)

	cmd := &cobra.Command{
		Use:   "update <runner-id>",
		Short: "Update a runner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &runnerv1.Runner{Id: args[0]}
			var paths []string

			if cmd.Flags().Changed("name") {
				runner.DisplayName = name
				paths = append(paths, "display_name")
			}

			if cmd.Flags().Changed("labels") {
				parsedLabels, err := parseLabels(labels)
				if err != nil {
					return err
				}
				runner.Labels = parsedLabels
				paths = append(paths, "labels")
			}

			if len(paths) == 0 {
				return fmt.Errorf("no fields to update: specify --name or --labels")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Runner().UpdateRunner(cmd.Context(), &runnerv1.UpdateRunnerRequest{
				Runner:     runner,
				UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				printRunnerTable(w, []*runnerv1.Runner{resp.Runner}, opts.OutputFormat)
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new display name")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "new labels in key=value format")

	return cmd
}
