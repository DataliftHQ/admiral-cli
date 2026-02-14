package agent

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	agentv1 "go.admiral.io/sdk/proto/agent/v1"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
		filter    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Agent().ListAgents(cmd.Context(), &agentv1.ListAgentsRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
				Filter:    filter,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tVERSION\tCLUSTER\tRUNNER\tLAST HEARTBEAT\tAGE")
					for _, a := range resp.Agents {
						_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							a.Id,
							a.DisplayName,
							output.FormatEnum(a.Status.String(), "AGENT_STATUS_"),
							a.Version,
							a.ClusterId,
							a.RunnerId,
							output.FormatAge(a.LastHeartbeatAt),
							output.FormatAge(a.CreatedAt),
						)
					}
				} else {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tVERSION\tAGE")
					for _, a := range resp.Agents {
						_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
							a.Id,
							a.DisplayName,
							output.FormatEnum(a.Status.String(), "AGENT_STATUS_"),
							a.Version,
							output.FormatAge(a.CreatedAt),
						)
					}
				}
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nNext page token: %s\n", resp.NextPageToken)
			}

			return nil
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of agents to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")
	cmd.Flags().StringVar(&filter, "filter", "", "filter expression (e.g. 'cluster_id = \"uuid\"')")

	return cmd
}
