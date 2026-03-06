package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/credentials"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/admiral/api/user/v1"
)

func newWhoamiCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current user, organization, and session",
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

			user := resp.GetUser()

			authType := "OAuth2"
			if os.Getenv(credentials.EnvToken) != "" {
				authType = "Token (ADMIRAL_TOKEN)"
			}

			details := []output.Detail{
				{Key: "Email", Value: user.GetEmail()},
				{Key: "Display Name", Value: user.GetDisplayName()},
				{Key: "ID", Value: user.GetId()},
				{Key: "Organization", Value: user.GetTenantId()},
				{Key: "Auth Type", Value: authType},
				{Key: "Server", Value: opts.ServerAddr},
			}

			// Show token expiry if available.
			tokenInfo, err := c.GetTokenInfo()
			if err == nil {
				details = append(details, output.Detail{
					Key:   "Token Expires In",
					Value: formatExpiresIn(tokenInfo.ExpiresIn()),
				})
			}

			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{Details: details},
			}

			return p.PrintDetail(resp, sections)
		},
	}
}

func formatExpiresIn(d time.Duration) string {
	if d <= 0 {
		return "expired"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}
