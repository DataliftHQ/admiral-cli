package runner

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List runners",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Runner().ListRunners(cmd.Context(), &runnerv1.ListRunnersRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				printRunnerTable(w, resp.Runners, opts.OutputFormat)
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nNext page token: %s\n", resp.NextPageToken)
			}
			return nil
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of runners to list per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")

	return cmd
}

func printRunnerTable(w *tabwriter.Writer, runners []*runnerv1.Runner, format output.Format) {
	if format == output.FormatWide {
		_, _ = fmt.Fprintln(w, "ID\tNAME\tKIND\tLABELS\tCREATED\tUPDATED\tAGE")
		for _, r := range runners {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r.Id,
				r.DisplayName,
				output.FormatEnum(r.Kind.String(), "RUNNER_KIND_"),
				output.FormatLabels(r.Labels),
				output.FormatTimestamp(r.CreatedAt),
				output.FormatTimestamp(r.UpdatedAt),
				output.FormatAge(r.CreatedAt),
			)
		}
	} else {
		_, _ = fmt.Fprintln(w, "ID\tNAME\tKIND\tAGE")
		for _, r := range runners {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				r.Id,
				r.DisplayName,
				output.FormatEnum(r.Kind.String(), "RUNNER_KIND_"),
				output.FormatAge(r.CreatedAt),
			)
		}
	}
}
