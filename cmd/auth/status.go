package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	userv1 "go.admiral.io/sdk/proto/user/v1"
)

func newStatusCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
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

			tokenInfo, err := c.GetTokenInfo()
			if err != nil {
				return fmt.Errorf("failed to get token info: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\n", resp.Email)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token expires in %s\n", formatExpiresIn(tokenInfo.ExpiresIn()))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Connected to %s\n", opts.ServerAddr)

			return nil
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
