package cluster

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newTokenGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <cluster> <token-id>",
		Short: "Get a cluster token by ID",
		Args:  cmdutil.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID := args[0]
			tokenID := args[1]

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().GetClusterToken(cmd.Context(), &clusterv1.GetClusterTokenRequest{
				ClusterId: clusterID,
				TokenId:   tokenID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				t := resp.AccessToken
				output.Writeln(w, "ID\tNAME\tSTATUS\tCREATED")
				output.Writef(w, "%s\t%s\t%s\t%s\n",
					t.Id,
					t.Name,
					output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
					output.FormatTimestamp(t.CreatedAt),
				)
			})
		},
	}

	return cmd
}
