package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/credentials"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/properties"
	userv1 "go.admiral.io/sdk/proto/user/v1"
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

			authType := "OAuth2"
			if os.Getenv(credentials.EnvToken) != "" {
				authType = "Token (ADMIRAL_TOKEN)"
			}

			details := []output.Detail{
				{Key: "Email", Value: resp.GetEmail()},
				{Key: "Display Name", Value: resp.GetDisplayName()},
				{Key: "ID", Value: resp.GetId()},
				{Key: "Organization", Value: resp.GetTenantId()},
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

			// Show active app context if set.
			props, err := properties.Load(opts.ConfigDir)
			if err == nil && props.App != "" {
				details = append(details, output.Detail{Key: "Active App", Value: props.App})
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
