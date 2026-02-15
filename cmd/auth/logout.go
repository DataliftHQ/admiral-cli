package auth

import (
	"context"

	"github.com/spf13/cobra"

	internalauth "go.admiral.io/cli/internal/auth"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
)

// NewLogoutCmd creates the logout command.
func NewLogoutCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out from Admiral",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := internalauth.Logout(context.Background(), internalauth.LogoutOptions{
				Issuer:    opts.Issuer,
				ClientID:  opts.ClientID,
				ConfigDir: opts.ConfigDir,
			})
			if err != nil {
				return err
			}

			output.Writeln(cmd.OutOrStdout(), "Successfully logged out.")
			return nil
		},
	}
}
