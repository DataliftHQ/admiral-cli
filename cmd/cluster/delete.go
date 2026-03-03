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
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a cluster",
		Args:  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to delete cluster %s", name)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			clusterID, err := cmdutil.ResolveClusterID(cmd.Context(), c.Cluster(), name)
			if err != nil {
				return err
			}

			resp, err := c.Cluster().DeleteCluster(cmd.Context(), &clusterv1.DeleteClusterRequest{
				ClusterId: clusterID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Cluster %s deleted\n", name)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm cluster deletion")

	return cmd
}
