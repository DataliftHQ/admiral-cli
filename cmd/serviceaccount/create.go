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

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		name        string
		description string
		scopes      []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().CreateServiceAccount(cmd.Context(), &serviceaccountv1.CreateServiceAccountRequest{
				DisplayName: name,
				Description: description,
				Scopes:      scopes,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				sa := resp.ServiceAccount
				if opts.OutputFormat == output.FormatWide {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tAGE\tDESCRIPTION\tSCOPES\tCREATED\tUPDATED")
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
				} else {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tAGE")
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						sa.Id,
						sa.DisplayName,
						output.FormatEnum(sa.Status.String(), "SERVICE_ACCOUNT_STATUS_"),
						output.FormatAge(sa.CreatedAt),
					)
				}
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the service account (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&description, "description", "", "description for the service account")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "comma-separated list of scopes")

	return cmd
}
