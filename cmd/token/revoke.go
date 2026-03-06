package token

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/admiral/api/user/v1"
)

func newRevokeCmd(opts *factory.Options) *cobra.Command {
	var (
		tokenID string
		confirm bool
	)

	cmd := &cobra.Command{
		Use:   "revoke [token]",
		Short: "Revoke a personal access token",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenName := ""
			if len(args) > 0 {
				tokenName = args[0]
			}
			if tokenName == "" && tokenID == "" {
				return fmt.Errorf("provide a token name or use --id")
			}

			displayName := tokenName
			if displayName == "" {
				displayName = tokenID
			}

			if !confirm {
				return fmt.Errorf("use --confirm to revoke token %s", displayName)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resolvedID, err := cmdutil.ResolvePersonalAccessTokenID(cmd.Context(), c.User(), tokenName, tokenID)
			if err != nil {
				return err
			}

			resp, err := c.User().RevokePersonalAccessToken(cmd.Context(), &userv1.RevokePersonalAccessTokenRequest{
				TokenId: resolvedID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Token %s revoked\n", displayName)
			})
		},
	}

	cmd.Flags().StringVar(&tokenID, "id", "", "token UUID (bypasses name resolution)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm token revocation")

	return cmd
}
