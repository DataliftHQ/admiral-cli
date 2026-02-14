package cluster

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newTokenListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list <cluster-id>",
		Short: "List cluster tokens",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID := args[0]

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().ListClusterTokens(cmd.Context(), &clusterv1.ListClusterTokensRequest{
				ClusterId: clusterID,
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				_, _ = fmt.Fprintln(w, "ID\tNAME\tPREFIX\tSTATUS\tCREATED")
				for _, t := range resp.AccessTokens {
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
						t.Id,
						t.DisplayName,
						t.TokenPrefix,
						output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
						output.FormatTimestamp(t.CreatedAt),
					)
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

	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of tokens to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")

	return cmd
}
