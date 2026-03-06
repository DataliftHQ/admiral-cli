package cluster

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newTokenListCmd(opts *factory.Options) *cobra.Command {
	var (
		clusterID string
		pageSize  int32
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list [cluster]",
		Short: "List cluster tokens",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterName := ""
			if len(args) > 0 {
				clusterName = args[0]
			}
			if clusterName == "" && clusterID == "" {
				return fmt.Errorf("provide a cluster name or use --cluster-id")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resolvedClusterID, err := cmdutil.ResolveClusterID(cmd.Context(), c.Cluster(), clusterName, clusterID)
			if err != nil {
				return err
			}

			resp, err := c.Cluster().ListClusterTokens(cmd.Context(), &clusterv1.ListClusterTokensRequest{
				ClusterId: resolvedClusterID,
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					output.Writeln(w, "ID\tNAME\tSTATUS\tSCOPES\tCREATED\tAGE")
					for _, t := range resp.AccessTokens {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
							t.Id,
							t.Name,
							output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
							output.FormatScopes(t.Scopes),
							output.FormatTimestamp(t.CreatedAt),
							output.FormatAge(t.CreatedAt),
						)
					}
				} else {
					output.Writeln(w, "NAME\tSTATUS\tAGE")
					for _, t := range resp.AccessTokens {
						output.Writef(w, "%s\t%s\t%s\n",
							t.Name,
							output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
							output.FormatAge(t.CreatedAt),
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

	cmd.Flags().StringVar(&clusterID, "cluster-id", "", "cluster UUID (bypasses name resolution)")
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of tokens to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")

	return cmd
}
