package cluster

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a cluster",
		Args:  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to delete cluster %s", id)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().DeleteCluster(cmd.Context(), &clusterv1.DeleteClusterRequest{
				ClusterId: id,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Cluster %s deleted\n", id)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm cluster deletion")

	return cmd
}
