package cluster

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		name      string
		labelStrs []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			labels, err := parseLabels(labelStrs)
			if err != nil {
				return err
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().CreateCluster(cmd.Context(), &clusterv1.CreateClusterRequest{
				DisplayName: name,
				Labels:      labels,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				cl := resp.Cluster
				output.Writeln(w, "ID\tNAME\tHEALTH\tAGE")
				output.Writef(w, "%s\t%s\t%s\t%s\n",
					cl.Id,
					cl.DisplayName,
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

	cmd.Flags().StringVar(&name, "name", "", "display name for the cluster (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringArrayVar(&labelStrs, "labels", nil, "labels in key=value format (repeatable)")

	return cmd
}
