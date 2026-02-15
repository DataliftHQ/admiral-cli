package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/user/v1"
)

func newWhoamiCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current user information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.User().GetUser(cmd.Context(), &userv1.GetUserRequest{})
			if err != nil {
				return fmt.Errorf("failed to get user info: %w", err)
			}

			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Email", Value: resp.GetEmail()},
						{Key: "Display Name", Value: resp.GetDisplayName()},
						{Key: "Given Name", Value: resp.GetGivenName()},
						{Key: "Family Name", Value: resp.GetFamilyName()},
						{Key: "ID", Value: resp.GetId()},
						{Key: "Tenant", Value: resp.GetTenantId()},
					},
				},
			}

			return p.PrintDetail(nil, sections)
		},
	}
}
