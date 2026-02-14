package cluster

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		name      string
		labelStrs []string
	)

	cmd := &cobra.Command{
		Use:   "update <cluster-id>",
		Short: "Update a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			var paths []string
			cluster := &clusterv1.Cluster{Id: id}

			if cmd.Flags().Changed("name") {
				cluster.DisplayName = name
				paths = append(paths, "display_name")
			}

			if cmd.Flags().Changed("labels") {
				labels, err := parseLabels(labelStrs)
				if err != nil {
					return err
				}
				cluster.Labels = labels
				paths = append(paths, "labels")
			}

			if len(paths) == 0 {
				return fmt.Errorf("at least one of --name or --labels must be specified")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().UpdateCluster(cmd.Context(), &clusterv1.UpdateClusterRequest{
				Cluster:    cluster,
				UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				cl := resp.Cluster
				_, _ = fmt.Fprintln(w, "ID\tNAME\tHEALTH\tAGE")
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					cl.Id,
					cl.DisplayName,
					output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_"),
					output.FormatAge(cl.CreatedAt),
				)
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new display name for the cluster")
	cmd.Flags().StringArrayVar(&labelStrs, "labels", nil, "labels in key=value format (repeatable)")

	return cmd
}
