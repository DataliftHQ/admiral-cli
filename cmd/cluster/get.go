package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	var clusterID string

	cmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get a cluster",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" && clusterID == "" {
				return fmt.Errorf("provide a cluster name or use --id")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resolvedID, err := cmdutil.ResolveClusterID(cmd.Context(), c.Cluster(), name, clusterID)
			if err != nil {
				return err
			}

			resp, err := c.Cluster().GetCluster(cmd.Context(), &clusterv1.GetClusterRequest{
				ClusterId: resolvedID,
			})
			if err != nil {
				return err
			}

			cl := resp.Cluster
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "ID", Value: cl.Id},
						{Key: "Name", Value: cl.Name},
						{Key: "Description", Value: cl.Description},
						{Key: "Health", Value: output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_")},
						{Key: "Cluster UID", Value: cl.ClusterUid},
						{Key: "Labels", Value: output.FormatLabels(cl.Labels)},
						{Key: "Created", Value: output.FormatTimestamp(cl.CreatedAt)},
						{Key: "Created By", Value: cl.CreatedBy},
						{Key: "Updated", Value: output.FormatTimestamp(cl.UpdatedAt)},
						{Key: "Updated By", Value: cl.UpdatedBy},
						{Key: "Age", Value: output.FormatAge(cl.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}

	cmd.Flags().StringVar(&clusterID, "id", "", "cluster UUID (bypasses name resolution)")

	return cmd
}
