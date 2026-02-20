package cluster

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var labelStrs []string

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a cluster",
		Args:  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			var paths []string
			cluster := &clusterv1.Cluster{Id: name}

			if cmd.Flags().Changed("label") {
				labels, err := cmdutil.ParseLabels(labelStrs)
				if err != nil {
					return err
				}
				cluster.Labels = labels
				paths = append(paths, "labels")
			}

			if len(paths) == 0 {
				return fmt.Errorf("at least --label must be specified")
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
				output.Writeln(w, "NAME\tHEALTH\tAGE")
				output.Writef(w, "%s\t%s\t%s\n",
					cl.Name,
					output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_"),
					output.FormatAge(cl.CreatedAt),
				)
			})
		},
	}

	cmdutil.AddLabelFlag(cmd, &labelStrs, "set a label (key=value, can be repeated)")

	return cmd
}
