package cluster

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
		labelStrs []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clusters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filter, err := cmdutil.BuildLabelFilter(labelStrs)
			if err != nil {
				return err
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().ListClusters(cmd.Context(), &clusterv1.ListClustersRequest{
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
					output.Writeln(w, "ID\tNAME\tHEALTH\tAGE\tCLUSTER-UID\tLABELS\tCREATED\tUPDATED")
					for _, cl := range resp.Clusters {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							cl.Id,
							cl.DisplayName,
							output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_"),
							output.FormatAge(cl.CreatedAt),
							cl.ClusterUid,
							output.FormatLabels(cl.Labels),
							output.FormatTimestamp(cl.CreatedAt),
							output.FormatTimestamp(cl.UpdatedAt),
						)
					}
				} else {
					output.Writeln(w, "ID\tNAME\tHEALTH\tAGE")
					for _, cl := range resp.Clusters {
						output.Writef(w, "%s\t%s\t%s\t%s\n",
							cl.Id,
							cl.DisplayName,
							output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_"),
							output.FormatAge(cl.CreatedAt),
						)
					}
				}
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" {
				output.Writef(cmd.ErrOrStderr(), "\nNext page token: %s\n", resp.NextPageToken)
			}

			return nil
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of clusters to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")
	cmdutil.AddLabelFlag(cmd, &labelStrs, "filter by label (key=value, can be repeated)")

	return cmd
}
