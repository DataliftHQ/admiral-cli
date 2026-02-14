package runner

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <runner-id>",
		Short: "Get a runner by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Runner().GetRunner(cmd.Context(), &runnerv1.GetRunnerRequest{
				RunnerId: args[0],
			})
			if err != nil {
				return err
			}

			r := resp.Runner
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Name", Value: r.DisplayName},
						{Key: "ID", Value: r.Id},
						{Key: "Kind", Value: output.FormatEnum(r.Kind.String(), "RUNNER_KIND_")},
						{Key: "Labels", Value: output.FormatLabels(r.Labels)},
						{Key: "Created", Value: output.FormatTimestamp(r.CreatedAt)},
						{Key: "Updated", Value: output.FormatTimestamp(r.UpdatedAt)},
						{Key: "Age", Value: output.FormatAge(r.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}
}
