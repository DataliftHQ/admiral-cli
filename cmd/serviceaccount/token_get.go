package serviceaccount

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newTokenGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <service-account-id> <token-id>",
		Short: "Get a service account token by ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			saID := args[0]
			tokenID := args[1]

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().GetServiceAccountToken(cmd.Context(), &serviceaccountv1.GetServiceAccountTokenRequest{
				ServiceAccountId: saID,
				TokenId:          tokenID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				t := resp.AccessToken
				_, _ = fmt.Fprintln(w, "ID\tNAME\tPREFIX\tSTATUS\tCREATED")
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					t.Id,
					t.DisplayName,
					t.TokenPrefix,
					output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
					output.FormatTimestamp(t.CreatedAt),
				)
			})
		},
	}

	return cmd
}
