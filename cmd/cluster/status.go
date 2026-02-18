package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newStatusCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:                   "status <name>",
		Short:                 "Get cluster status and telemetry",
		DisableFlagsInUseLine: true,
		Args:                  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().GetClusterStatus(cmd.Context(), &clusterv1.GetClusterStatusRequest{
				ClusterId: args[0],
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Health", Value: output.FormatEnum(resp.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_")},
						{Key: "Reported At", Value: output.FormatTimestamp(resp.ReportedAt)},
					},
				},
			}

			s := resp.Status
			if s != nil {
				sections = append(sections,
					output.Section{
						Name: "Kubernetes",
						Details: []output.Detail{
							{Key: "Version", Value: s.K8SVersion},
						},
					},
					output.Section{
						Name: "Nodes",
						Details: []output.Detail{
							{Key: "Total", Value: fmt.Sprintf("%d", s.NodeCount)},
							{Key: "Ready", Value: fmt.Sprintf("%d", s.NodesReady)},
						},
					},
					output.Section{
						Name: "Pods",
						Details: []output.Detail{
							{Key: "Capacity", Value: fmt.Sprintf("%d", s.PodCapacity)},
							{Key: "Total", Value: fmt.Sprintf("%d", s.PodCount)},
							{Key: "Running", Value: fmt.Sprintf("%d", s.PodsRunning)},
							{Key: "Pending", Value: fmt.Sprintf("%d", s.PodsPending)},
							{Key: "Failed", Value: fmt.Sprintf("%d", s.PodsFailed)},
						},
					},
					output.Section{
						Name: "Resources",
						Details: []output.Detail{
							{Key: "CPU (millicores)", Value: fmt.Sprintf("%d / %d", s.CpuUsedMillicores, s.CpuCapacityMillicores)},
							{Key: "Memory (bytes)", Value: fmt.Sprintf("%d / %d", s.MemoryUsedBytes, s.MemoryCapacityBytes)},
						},
					},
					output.Section{
						Name: "Workloads",
						Details: []output.Detail{
							{Key: "Total", Value: fmt.Sprintf("%d", s.WorkloadsTotal)},
							{Key: "Healthy", Value: fmt.Sprintf("%d", s.WorkloadsHealthy)},
							{Key: "Degraded", Value: fmt.Sprintf("%d", s.WorkloadsDegraded)},
							{Key: "Error", Value: fmt.Sprintf("%d", s.WorkloadsError)},
						},
					},
				)
			}

			return p.PrintDetail(resp, sections)
		},
	}
}
