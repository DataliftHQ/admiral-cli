package token

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/user/v1"
)

func newRevokeCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "revoke <token-id>",
		Short: "Revoke a personal access token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to revoke token %s", id)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.User().RevokePersonalAccessToken(cmd.Context(), &userv1.RevokePersonalAccessTokenRequest{
				TokenId: id,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				_, _ = fmt.Fprintf(w, "Token %s revoked\n", id)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm token revocation")

	return cmd
}
