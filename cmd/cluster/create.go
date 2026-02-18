package cluster

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var labelStrs []string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new cluster",
		Args:  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			labels, err := cmdutil.ParseLabels(labelStrs)
			if err != nil {
				return err
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().CreateCluster(cmd.Context(), &clusterv1.CreateClusterRequest{
				Name:   args[0],
				Labels: labels,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				cl := resp.Cluster
				output.Writeln(w, "NAME\tHEALTH\tAGE")
				output.Writef(w, "%s\t%s\t%s\n",
					cl.Name,
					output.FormatEnum(cl.HealthStatus.String(), "CLUSTER_HEALTH_STATUS_"),
					output.FormatAge(cl.CreatedAt),
				)
			}); err != nil {
				return err
			}

			if resp.PlainTextToken != "" {
				output.PrintToken(cmd.ErrOrStderr(), resp.PlainTextToken)
			}

			return nil
		},
	}

	cmdutil.AddLabelFlag(cmd, &labelStrs, "set a label (key=value, can be repeated)")

	return cmd
}
