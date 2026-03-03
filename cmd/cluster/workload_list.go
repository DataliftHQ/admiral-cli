package cluster

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newWorkloadListCmd(opts *factory.Options) *cobra.Command {
	var (
		clusterID    string
		pageSize     int32
		pageToken    string
		namespace    string
		kind         string
		name         string
		healthStatus string
	)

	cmd := &cobra.Command{
		Use:   "list [cluster]",
		Short: "List workloads in a cluster",
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

			filter := buildWorkloadFilter(namespace, kind, name, healthStatus)

			resp, err := c.Cluster().ListWorkloads(cmd.Context(), &clusterv1.ListWorkloadsRequest{
				ClusterId: resolvedClusterID,
				Filter:    filter,
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					output.Writeln(w, "NAMESPACE\tNAME\tKIND\tHEALTH\tREPLICAS\tSTATUS-REASON\tCPU\tMEMORY\tLABELS\tAGE")
					for _, wl := range resp.Workloads {
						output.Writef(w, "%s\t%s\t%s\t%s\t%d/%d\t%s\t%dm/%dm\t%s/%s\t%s\t%s\n",
							wl.Namespace,
							wl.Name,
							wl.Kind,
							output.FormatEnum(wl.HealthStatus.String(), "WORKLOAD_HEALTH_STATUS_"),
							wl.ReplicasReady, wl.ReplicasDesired,
							wl.StatusReason,
							wl.CpuUsedMillicores, wl.CpuRequestsMillicores,
							formatBytes(wl.MemoryUsedBytes), formatBytes(wl.MemoryRequestsBytes),
							output.FormatLabels(wl.Labels),
							output.FormatAge(wl.LastUpdatedAt),
						)
					}
				} else {
					output.Writeln(w, "NAMESPACE\tNAME\tKIND\tHEALTH\tREPLICAS\tAGE")
					for _, wl := range resp.Workloads {
						output.Writef(w, "%s\t%s\t%s\t%s\t%d/%d\t%s\n",
							wl.Namespace,
							wl.Name,
							wl.Kind,
							output.FormatEnum(wl.HealthStatus.String(), "WORKLOAD_HEALTH_STATUS_"),
							wl.ReplicasReady, wl.ReplicasDesired,
							output.FormatAge(wl.LastUpdatedAt),
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
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of workloads to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")
	cmd.Flags().StringVar(&namespace, "namespace", "", "filter by namespace")
	cmd.Flags().StringVar(&kind, "kind", "", "filter by kind (e.g. Deployment, StatefulSet, DaemonSet)")
	cmd.Flags().StringVar(&name, "name", "", "filter by workload name")
	cmd.Flags().StringVar(&healthStatus, "health-status", "", "filter by health status (HEALTHY, DEGRADED, ERROR)")

	return cmd
}

func buildWorkloadFilter(namespace, kind, name, healthStatus string) string {
	var parts []string
	if namespace != "" {
		parts = append(parts, fmt.Sprintf("field['namespace'] = '%s'", namespace))
	}
	if kind != "" {
		parts = append(parts, fmt.Sprintf("field['kind'] = '%s'", kind))
	}
	if name != "" {
		parts = append(parts, fmt.Sprintf("field['name'] = '%s'", name))
	}
	if healthStatus != "" {
		parts = append(parts, fmt.Sprintf("field['health_status'] = '%s'", healthStatus))
	}
	return strings.Join(parts, " AND ")
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fGi", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fMi", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fKi", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
