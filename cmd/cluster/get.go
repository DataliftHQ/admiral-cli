package cluster

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:                   "get <cluster-id>",
		Short:                 "Get a cluster by ID",
		DisableFlagsInUseLine: true,
		Args:                  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().GetCluster(cmd.Context(), &clusterv1.GetClusterRequest{
				ClusterId: args[0],
			})
			if err != nil {
				return err
			}

			cl := resp.Cluster
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Name", Value: cl.DisplayName},
						{Key: "ID", Value: cl.Id},
						{Key: "Health", Value: output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_")},
						{Key: "Cluster UID", Value: cl.ClusterUid},
						{Key: "Labels", Value: output.FormatLabels(cl.Labels)},
						{Key: "Created", Value: output.FormatTimestamp(cl.CreatedAt)},
						{Key: "Updated", Value: output.FormatTimestamp(cl.UpdatedAt)},
						{Key: "Age", Value: output.FormatAge(cl.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}
}
