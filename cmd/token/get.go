package token

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/user/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <token-id>",
		Short: "Get a personal access token by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.User().GetPersonalAccessToken(cmd.Context(), &userv1.GetPersonalAccessTokenRequest{
				TokenId: id,
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
