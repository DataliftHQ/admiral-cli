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
	var clusterID string

	cmd := &cobra.Command{
		Use:   "create [cluster] <name>",
		Short: "Create a cluster token",
		Args:  cmdutil.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var clusterName, tokenName string
			switch {
			case len(args) == 2:
				clusterName = args[0]
				tokenName = args[1]
			case clusterID != "":
				tokenName = args[0]
			default:
				return fmt.Errorf("provide cluster name as first arg or use --cluster-id")
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
				Name:      tokenName,
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

	return cmd
}
