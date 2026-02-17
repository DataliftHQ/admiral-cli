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

func newTokenRevokeCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "revoke <cluster-id> <token-id>",
		Short: "Revoke a cluster token",
		Args:  cmdutil.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID := args[0]
			tokenID := args[1]

			if !confirm {
				return fmt.Errorf("use --confirm to revoke token %s", tokenID)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().RevokeClusterToken(cmd.Context(), &clusterv1.RevokeClusterTokenRequest{
				ClusterId: clusterID,
				TokenId:   tokenID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Token %s revoked\n", tokenID)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm token revocation")

	return cmd
}
