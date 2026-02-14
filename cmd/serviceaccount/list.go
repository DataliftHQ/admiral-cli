package serviceaccount

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().ListServiceAccounts(cmd.Context(), &serviceaccountv1.ListServiceAccountsRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tAGE\tDESCRIPTION\tSCOPES\tCREATED\tUPDATED")
					for _, sa := range resp.ServiceAccounts {
						_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							sa.Id,
							sa.DisplayName,
							output.FormatEnum(sa.Status.String(), "SERVICE_ACCOUNT_STATUS_"),
							output.FormatAge(sa.CreatedAt),
							sa.Description,
							strings.Join(sa.Scopes, ","),
							output.FormatTimestamp(sa.CreatedAt),
							output.FormatTimestamp(sa.UpdatedAt),
						)
					}
				} else {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tAGE")
					for _, sa := range resp.ServiceAccounts {
						_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
							sa.Id,
							sa.DisplayName,
							output.FormatEnum(sa.Status.String(), "SERVICE_ACCOUNT_STATUS_"),
							output.FormatAge(sa.CreatedAt),
						)
					}
				}
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nNext page token: %s\n", resp.NextPageToken)
			}

			return nil
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "number of service accounts to return per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "token for the next page of results")

	return cmd
}
