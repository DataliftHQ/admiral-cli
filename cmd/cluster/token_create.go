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

func newTokenCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		clusterID string
		name      string
	)

	cmd := &cobra.Command{
		Use:   "create [cluster]",
		Short: "Create a cluster token",
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

			resp, err := c.Cluster().CreateClusterToken(cmd.Context(), &clusterv1.CreateClusterTokenRequest{
				ClusterId: resolvedClusterID,
				Name:      name,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				t := resp.AccessToken
				output.Writeln(w, "ID\tNAME\tSTATUS\tCREATED")
				output.Writef(w, "%s\t%s\t%s\t%s\n",
					t.Id,
					t.Name,
					output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
					output.FormatTimestamp(t.CreatedAt),
				)
			}); err != nil {
				return err
			}

			output.PrintToken(cmd.ErrOrStderr(), resp.PlainTextToken)

			return nil
		},
	}

	cmd.Flags().StringVar(&clusterID, "cluster-id", "", "cluster UUID (bypasses name resolution)")
	cmd.Flags().StringVar(&name, "name", "", "display name for the token (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
