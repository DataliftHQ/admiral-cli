package token

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/admiral/api/user/v1"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List personal access tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.User().ListPersonalAccessTokens(cmd.Context(), &userv1.ListPersonalAccessTokensRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					output.Writeln(w, "ID\tNAME\tSTATUS\tSCOPES\tEXPIRES\tLAST USED\tCREATED\tAGE")
					for _, t := range resp.AccessTokens {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							t.Id,
							t.Name,
							output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
							output.FormatScopes(t.Scopes),
							output.FormatTimestamp(t.ExpiresAt),
							output.FormatTimestamp(t.LastUsedAt),
							output.FormatTimestamp(t.CreatedAt),
							output.FormatAge(t.CreatedAt),
						)
					}
				} else {
					output.Writeln(w, "NAME\tSTATUS\tSCOPES\tAGE")
					for _, t := range resp.AccessTokens {
						output.Writef(w, "%s\t%s\t%s\t%s\n",
							t.Name,
							output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
							output.FormatScopes(t.Scopes),
							output.FormatAge(t.CreatedAt),
						)
					}
				}
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" && opts.OutputFormat != output.FormatJSON && opts.OutputFormat != output.FormatYAML {
				output.Writef(cmd.ErrOrStderr(), "\nNEXT PAGE TOKEN: %s\n", resp.NextPageToken)
			}

			return nil
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of tokens to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")

	return cmd
}
