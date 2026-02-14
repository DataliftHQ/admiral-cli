package serviceaccount

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <service-account-id>",
		Short: "Get a service account by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().GetServiceAccount(cmd.Context(), &serviceaccountv1.GetServiceAccountRequest{
				ServiceAccountId: args[0],
			})
			if err != nil {
				return err
			}

			sa := resp.ServiceAccount
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Name", Value: sa.DisplayName},
						{Key: "ID", Value: sa.Id},
						{Key: "Status", Value: output.FormatEnum(sa.Status.String(), "SERVICE_ACCOUNT_STATUS_")},
						{Key: "Description", Value: sa.Description},
						{Key: "Scopes", Value: output.FormatScopes(sa.Scopes)},
						{Key: "Created", Value: output.FormatTimestamp(sa.CreatedAt)},
						{Key: "Updated", Value: output.FormatTimestamp(sa.UpdatedAt)},
						{Key: "Age", Value: output.FormatAge(sa.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}
}
