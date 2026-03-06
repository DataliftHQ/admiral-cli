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

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var (
		clusterID string
		confirm   bool
	)

	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a cluster",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" && clusterID == "" {
				return fmt.Errorf("provide a cluster name or use --id")
			}

			displayName := name
			if displayName == "" {
				displayName = clusterID
			}

			if !confirm {
				return fmt.Errorf("use --confirm to delete cluster %s", displayName)
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

			resp, err := c.Cluster().DeleteCluster(cmd.Context(), &clusterv1.DeleteClusterRequest{
				ClusterId: resolvedID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Cluster %s deleted\n", displayName)
			})
		},
	}

	cmd.Flags().StringVar(&clusterID, "id", "", "cluster UUID (bypasses name resolution)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm cluster deletion")

	return cmd
}
