package serviceaccount

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newTokenRevokeCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "revoke <service-account-id> <token-id>",
		Short: "Revoke a service account token",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			saID := args[0]
			tokenID := args[1]

			if !confirm {
				return fmt.Errorf("use --confirm to revoke token %s for service account %s", tokenID, saID)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().RevokeServiceAccountToken(cmd.Context(), &serviceaccountv1.RevokeServiceAccountTokenRequest{
				ServiceAccountId: saID,
				TokenId:          tokenID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				_, _ = fmt.Fprintf(w, "Token %s revoked for service account %s\n", tokenID, saID)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm token revocation")

	return cmd
}
